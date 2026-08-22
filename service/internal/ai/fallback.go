package ai

import (
	"fmt"
	"log"
	"strings"
)

// FallbackProvider tries providers in order and answers with the first one
// that succeeds. It exists so a hosted endpoint can front the assistant while
// the local GPU stays as the safety net: when the hosted service is down,
// rate-limited, or returns garbage, the reply still comes from Ollama instead
// of an error bubble in the chat.
type FallbackProvider struct {
	primary Provider
	fallbacks []Provider
}

// NewFallbackProvider chains providers: the first argument is primary, the
// rest are tried in order when the primary fails.
func NewFallbackProvider(primary Provider, fallbacks ...Provider) *FallbackProvider {
	return &FallbackProvider{primary: primary, fallbacks: fallbacks}
}

func (f *FallbackProvider) Name() string {
	names := make([]string, 0, len(f.fallbacks)+1)
	names = append(names, f.primary.Name())
	for _, p := range f.fallbacks {
		names = append(names, p.Name())
	}
	return strings.Join(names, " → ")
}

// SupportsImages is true only when every provider in the chain can read
// images: a chain that would drop the picture on fallback is worse than one
// that refuses it up front, because the user gets a confident text answer
// about an image nobody saw.
func (f *FallbackProvider) SupportsImages() bool {
	if !f.primary.SupportsImages() {
		return false
	}
	for _, p := range f.fallbacks {
		if !p.SupportsImages() {
			return false
		}
	}
	return true
}

// Generate answers with the primary provider, falling back on any error. The
// last error is returned when everything fails, so callers still see the real
// cause (unreachable endpoint, bad model name, …).
func (f *FallbackProvider) Generate(system string, messages []Message) (string, error) {
	reply, err := f.primary.Generate(system, messages)
	if err == nil {
		return reply, nil
	}
	log.Printf("chat AI: primary provider failed, trying fallbacks: %v", err)
	lastErr := err
	for _, p := range f.fallbacks {
		reply, ferr := p.Generate(system, messages)
		if ferr == nil {
			return reply, nil
		}
		lastErr = ferr
	}
	return "", fmt.Errorf("all providers failed, last error: %w", lastErr)
}
