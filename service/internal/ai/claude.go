package ai

import (
	"context"
	"encoding/base64"
	"os"
	"strings"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

// ClaudeProvider generates replies via Anthropic's Messages API using the
// official Go SDK. Configure via ANTHROPIC_API_KEY and (optionally)
// AI_CLAUDE_MODEL (defaults to claude-opus-4-8).
type ClaudeProvider struct {
	client anthropic.Client
	model  anthropic.Model
}

func NewClaudeProvider() *ClaudeProvider {
	model := anthropic.ModelClaudeOpus4_8
	if v := strings.TrimSpace(os.Getenv("AI_CLAUDE_MODEL")); v != "" {
		model = anthropic.Model(v)
	}
	var opts []option.RequestOption
	if key := strings.TrimSpace(os.Getenv("ANTHROPIC_API_KEY")); key != "" {
		opts = append(opts, option.WithAPIKey(key))
	}
	return &ClaudeProvider{client: anthropic.NewClient(opts...), model: model}
}

func (p *ClaudeProvider) Name() string { return "claude (vision)" }

// SupportsImages is always true: every current Claude model reads images, so
// unlike the local setup there is no second model to configure.
func (p *ClaudeProvider) SupportsImages() bool { return true }

// claudeImageTypes are the media types the Messages API accepts. Anything else
// is dropped rather than sent and rejected.
var claudeImageTypes = map[string]bool{
	"image/jpeg": true, "image/png": true, "image/gif": true, "image/webp": true,
}

func (p *ClaudeProvider) Generate(system string, messages []Message) (string, error) {
	msgs := make([]anthropic.MessageParam, 0, len(messages))
	for _, m := range prepareMessages(messages) {
		if m.Role == "assistant" {
			msgs = append(msgs, anthropic.NewAssistantMessage(anthropic.NewTextBlock(m.Content)))
			continue
		}
		// Images first, then the text: Anthropic recommends that order when a
		// question refers to the attached image.
		blocks := make([]anthropic.ContentBlockParamUnion, 0, len(m.Images)+1)
		for _, img := range m.Images {
			img = img.Bounded()
			if !claudeImageTypes[img.MimeType] {
				continue
			}
			blocks = append(blocks, anthropic.NewImageBlockBase64(img.MimeType, base64.StdEncoding.EncodeToString(img.Data)))
		}
		if strings.TrimSpace(m.Content) != "" || len(blocks) == 0 {
			blocks = append(blocks, anthropic.NewTextBlock(m.Content))
		}
		msgs = append(msgs, anthropic.NewUserMessage(blocks...))
	}
	if len(msgs) == 0 {
		return "", nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	params := anthropic.MessageNewParams{
		Model:     p.model,
		MaxTokens: 2048,
		Messages:  msgs,
	}
	if system != "" {
		params.System = []anthropic.TextBlockParam{{Text: system}}
	}

	resp, err := p.client.Messages.New(ctx, params)
	if err != nil {
		return "", err
	}
	var sb strings.Builder
	for _, block := range resp.Content {
		if t, ok := block.AsAny().(anthropic.TextBlock); ok {
			sb.WriteString(t.Text)
		}
	}
	return strings.TrimSpace(sb.String()), nil
}
