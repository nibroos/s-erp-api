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

// LocalProvider talks to a local, OpenAI-compatible chat-completions endpoint
// (Ollama, llama.cpp, vLLM, LM Studio, text-generation-webui, …). Configure via
// AI_LOCAL_BASE_URL (e.g. http://localhost:11434/v1), AI_LOCAL_MODEL, and an
// optional AI_LOCAL_API_KEY.
type LocalProvider struct {
	baseURL     string
	model       string
	visionModel string
	apiKey      string
	temperature float64
	client      *http.Client
}

// defaultTemperature is low on purpose. Both jobs here — turning a question
// into SQL, and restating query results — want the likeliest wording, not a
// creative one; at the model's default the SQL pass produces casts and
// arithmetic that do not parse. Override with AI_LOCAL_TEMPERATURE.
const defaultTemperature = 0.2

func NewLocalProvider() *LocalProvider {
	base := strings.TrimRight(os.Getenv("AI_LOCAL_BASE_URL"), "/")
	if base == "" {
		base = "http://localhost:11434/v1"
	}
	model := os.Getenv("AI_LOCAL_MODEL")
	if model == "" {
		model = "llama3.1"
	}
	temperature := defaultTemperature
	if v, err := strconv.ParseFloat(strings.TrimSpace(os.Getenv("AI_LOCAL_TEMPERATURE")), 64); err == nil && v >= 0 {
		temperature = v
	}
	return &LocalProvider{
		baseURL: base,
		model:   model,
		// Text and vision are separate models: the text model is tuned for the
		// SQL and chat work and cannot see images, so image turns are routed to
		// AI_LOCAL_VISION_MODEL. Empty means images are unsupported.
		visionModel: strings.TrimSpace(os.Getenv("AI_LOCAL_VISION_MODEL")),
		apiKey:      os.Getenv("AI_LOCAL_API_KEY"),
		temperature: temperature,
		client:      &http.Client{Timeout: 120 * time.Second},
	}
}

// Name includes the resolved endpoint and model so a misconfigured assistant is
// obvious in the startup log instead of only surfacing as a failed reply.
func (p *LocalProvider) Name() string {
	vision := "no vision model"
	if p.visionModel != "" {
		vision = "vision=" + p.visionModel
	}
	return fmt.Sprintf("local (%s, model=%s, %s)", p.baseURL, p.model, vision)
}

// ocMessage carries either a plain string (text-only turn) or an array of
// content parts (a turn with images). Both shapes are valid OpenAI-compatible
// input; the plain string is kept for text so nothing changes for servers that
// predate the multimodal format.
type ocMessage struct {
	Role    string      `json:"role"`
	Content interface{} `json:"content"`
}

type ocPart struct {
	Type     string      `json:"type"`
	Text     string      `json:"text,omitempty"`
	ImageURL *ocImageURL `json:"image_url,omitempty"`
}

type ocImageURL struct {
	URL string `json:"url"`
}

type ocRequest struct {
	Model       string      `json:"model"`
	Messages    []ocMessage `json:"messages"`
	Stream      bool        `json:"stream"`
	Temperature float64     `json:"temperature"`
}

// The response content is always a plain string.
type ocResponse struct {
	Choices []struct {
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

// contentFor renders one turn: a bare string normally, or text-plus-images as
// content parts. Images are inlined as data URIs because the model server runs
// in its own container and cannot fetch a presigned MinIO URL from ours.
func contentFor(m Message) interface{} {
	if len(m.Images) == 0 {
		return m.Content
	}
	parts := make([]ocPart, 0, len(m.Images)+1)
	if strings.TrimSpace(m.Content) != "" {
		parts = append(parts, ocPart{Type: "text", Text: m.Content})
	}
	for _, img := range m.Images {
		parts = append(parts, ocPart{
			Type:     "image_url",
			ImageURL: &ocImageURL{URL: img.Bounded().DataURI()},
		})
	}
	return parts
}

// SupportsImages reports whether a vision model is configured.
func (p *LocalProvider) SupportsImages() bool { return p.visionModel != "" }

func (p *LocalProvider) Generate(system string, messages []Message) (string, error) {
	prepared := prepareMessages(messages)

	model := p.model
	if HasImages(prepared) {
		if p.visionModel == "" {
			// Sending images to a text-only model wastes a slow round trip and
			// returns a confident answer about an image it never saw.
			return "", fmt.Errorf("local AI: images were sent but AI_LOCAL_VISION_MODEL is not set")
		}
		model = p.visionModel
	}

	payload := ocRequest{Model: model, Stream: false, Temperature: p.temperature}
	if system != "" {
		payload.Messages = append(payload.Messages, ocMessage{Role: "system", Content: system})
	}
	for _, m := range prepared {
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
		return "", fmt.Errorf("local AI: %s unreachable: %w", p.baseURL, err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return "", fmt.Errorf("local AI: reading response (status %d): %w", resp.StatusCode, err)
	}
	// A wrong path or model answers with a non-JSON error page; surface the body
	// rather than a decode failure.
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return "", fmt.Errorf("local AI: status %d from %s: %s", resp.StatusCode, p.baseURL, snippet(raw))
	}

	var parsed ocResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return "", fmt.Errorf("local AI: bad response (status %d): %w: %s", resp.StatusCode, err, snippet(raw))
	}
	if parsed.Error != nil {
		return "", fmt.Errorf("local AI error: %s", parsed.Error.Message)
	}
	if len(parsed.Choices) == 0 {
		return "", fmt.Errorf("local AI: empty response (status %d): %s", resp.StatusCode, snippet(raw))
	}
	return strings.TrimSpace(parsed.Choices[0].Message.Content), nil
}

// snippet trims a response body down to something loggable on one line.
func snippet(b []byte) string {
	s := strings.TrimSpace(string(b))
	if len(s) > 300 {
		s = s[:300] + "…"
	}
	return strings.Join(strings.Fields(s), " ")
}
