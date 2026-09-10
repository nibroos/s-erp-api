package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// RemoteProvider talks to a self-hosted, OpenAI-compatible chat-completions
// endpoint (e.g. https://ai.nibros.space/v1). It mirrors LocalProvider but
// targets a hosted service instead of a LAN GPU: a primary model with
// same-endpoint fallback models that are tried in order when the primary
// fails, so a single down or rate-limited model does not take the assistant
// with it. Configure via AI_REMOTE_BASE_URL, AI_REMOTE_MODEL (comma-separated
// list, first is primary), and AI_REMOTE_API_KEY.
type RemoteProvider struct {
	baseURL     string
	models      []string
	visionModel string
	apiKey      string
	temperature float64
	client      *http.Client
}

func NewRemoteProvider() *RemoteProvider {
	base := strings.TrimRight(os.Getenv("AI_REMOTE_BASE_URL"), "/")
	if base == "" {
		base = "https://ai.nibros.space/v1"
	}
	// The hosted endpoint serves several models behind one URL; listing them
	// here (AI_REMOTE_MODEL="ai/gpt-5.6-luna,bbai") lets the provider walk the
	// list when the primary is unavailable.
	var models []string
	for _, m := range strings.Split(os.Getenv("AI_REMOTE_MODEL"), ",") {
		if m = strings.TrimSpace(m); m != "" {
			models = append(models, m)
		}
	}
	if len(models) == 0 {
		models = []string{"ai/gpt-5.6-luna", "bbai"}
	}
	temperature := defaultTemperature
	if v, err := strconv.ParseFloat(strings.TrimSpace(os.Getenv("AI_REMOTE_TEMPERATURE")), 64); err == nil && v >= 0 {
		temperature = v
	}
	return &RemoteProvider{
		baseURL: base,
		models:  models,
		// Same split as the local provider: image turns go to a dedicated
		// vision model when one is configured, otherwise images are refused
		// rather than silently answered by a text-only model.
		visionModel: strings.TrimSpace(os.Getenv("AI_REMOTE_VISION_MODEL")),
		apiKey:      os.Getenv("AI_REMOTE_API_KEY"),
		temperature: temperature,
		client:      &http.Client{Timeout: 120 * time.Second},
	}
}

// Name includes the resolved endpoint and model chain so the startup log shows
// exactly which hosted models will be tried, and in what order.
func (p *RemoteProvider) Name() string {
	vision := "no vision model"
	if p.visionModel != "" {
		vision = "vision=" + p.visionModel
	}
	return fmt.Sprintf("remote (%s, models=%s, %s)", p.baseURL, strings.Join(p.models, " → "), vision)
}

// SupportsImages reports whether a vision model is configured.
func (p *RemoteProvider) SupportsImages() bool { return p.visionModel != "" }

func (p *RemoteProvider) Generate(system string, messages []Message) (string, error) {
	prepared := prepareMessages(messages)

	vision := HasImages(prepared)
	if vision && p.visionModel == "" {
		// Sending images to a text-only model wastes a slow round trip and
		// returns a confident answer about an image it never saw.
		return "", fmt.Errorf("remote AI: images were sent but AI_REMOTE_VISION_MODEL is not set")
	}

	var lastErr error
	for _, model := range p.models {
		reply, err := p.generateWithModel(model, system, prepared, vision)
		if err == nil {
			return reply, nil
		}
		lastErr = err
		// A context overflow is a property of the request, not the model —
		// retrying it on the next model just repeats the same failure.
		if isContextOverflow(err) {
			return "", err
		}
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("remote AI: no models configured")
	}
	return "", fmt.Errorf("remote AI: all %d model(s) failed: %w", len(p.models), lastErr)
}

// generateWithModel runs one chat-completions round trip against a single
// model. The request/response shapes are shared with the local provider.
func (p *RemoteProvider) generateWithModel(model, system string, messages []Message, vision bool) (string, error) {
	payload := ocRequest{Model: model, Stream: false, Temperature: p.temperature}
	if system != "" {
		payload.Messages = append(payload.Messages, ocMessage{Role: "system", Content: system})
	}
	for _, m := range messages {
		payload.Messages = append(payload.Messages, ocMessage{Role: m.Role, Content: contentFor(m)})
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	if p.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+p.apiKey)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("remote AI: %s unreachable: %w", p.baseURL, err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return "", fmt.Errorf("remote AI: reading response (status %d): %w", resp.StatusCode, err)
	}
	// A wrong path or model answers with a non-JSON error page; surface the body
	// rather than a decode failure.
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return "", fmt.Errorf("remote AI: status %d from %s (model %s): %s", resp.StatusCode, p.baseURL, model, snippet(raw))
	}

	var parsed ocResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return "", fmt.Errorf("remote AI: bad response (status %d): %w: %s", resp.StatusCode, err, snippet(raw))
	}
	if parsed.Error != nil {
		return "", fmt.Errorf("remote AI error (model %s): %s", model, parsed.Error.Message)
	}
	if len(parsed.Choices) == 0 {
		return "", fmt.Errorf("remote AI: empty response (status %d, model %s): %s", resp.StatusCode, model, snippet(raw))
	}
	return strings.TrimSpace(parsed.Choices[0].Message.Content), nil
}
