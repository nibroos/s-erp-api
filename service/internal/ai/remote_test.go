package ai

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// --- FallbackProvider -------------------------------------------------------

// failingProvider always errors, recording that it was asked.
type failingProvider struct{ calls int }

func (f *failingProvider) Name() string { return "failing" }
func (f *failingProvider) SupportsImages() bool { return false }
func (f *failingProvider) Generate(system string, messages []Message) (string, error) {
	f.calls++
	return "", errors.New("down")
}

// The primary's success must short-circuit the chain — fallbacks exist for
// failures, not for load balancing.
func TestFallbackUsesPrimaryOnSuccess(t *testing.T) {
	primary := &stubProvider{replies: []string{"from primary"}}
	fallback := &failingProvider{}
	f := NewFallbackProvider(primary, fallback)

	got, err := f.Generate("sys", []Message{{Role: "user", Content: "hi"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "from primary" {
		t.Fatalf("want primary reply, got %q", got)
	}
	if fallback.calls != 0 {
		t.Fatalf("fallback was called despite primary success")
	}
}

// A failed primary hands off to the fallback, and the user still gets a reply.
func TestFallbackFallsThroughOnError(t *testing.T) {
	f := NewFallbackProvider(&failingProvider{}, &stubProvider{replies: []string{"from local"}})

	got, err := f.Generate("sys", []Message{{Role: "user", Content: "hi"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "from local" {
		t.Fatalf("want fallback reply, got %q", got)
	}
}

// When everything fails the error must survive, not be swallowed into an
// empty success.
func TestFallbackReturnsErrorWhenAllFail(t *testing.T) {
	f := NewFallbackProvider(&failingProvider{}, &failingProvider{})

	_, err := f.Generate("sys", []Message{{Role: "user", Content: "hi"}})
	if err == nil {
		t.Fatal("expected an error when every provider fails")
	}
	if !strings.Contains(err.Error(), "down") {
		t.Fatalf("underlying cause lost: %v", err)
	}
}

// The chain only claims image support when every member does — otherwise a
// fallback would silently answer about a picture it never saw.
func TestFallbackSupportsImagesRequiresWholeChain(t *testing.T) {
	if NewFallbackProvider(&stubProvider{vision: true}, &stubProvider{vision: false}).SupportsImages() {
		t.Fatal("chain claimed vision with a non-vision fallback")
	}
	if !NewFallbackProvider(&stubProvider{vision: true}, &stubProvider{vision: true}).SupportsImages() {
		t.Fatal("all-vision chain should support images")
	}
}

// --- RemoteProvider ---------------------------------------------------------

// remoteStubServer answers /chat/completions, failing the first N requests
// with 500 so the model-chain fallback can be exercised.
func remoteStubServer(t *testing.T, failFirst int) *httptest.Server {
	t.Helper()
	var calls int
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chat/completions" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		calls++
		if calls <= failFirst {
			http.Error(w, `{"error":{"message":"model overloaded"}}`, http.StatusInternalServerError)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"choices": []map[string]interface{}{
				{"message": map[string]string{"role": "assistant", "content": "remote reply"}},
			},
		})
	}))
}

func newTestRemoteProvider(t *testing.T, baseURL string) *RemoteProvider {
	t.Helper()
	t.Setenv("AI_REMOTE_BASE_URL", baseURL)
	t.Setenv("AI_REMOTE_MODEL", "ai/gpt-5.6-luna,bbai")
	t.Setenv("AI_REMOTE_API_KEY", "test-key")
	return NewRemoteProvider()
}

// The primary hosted model answers; the fallback model is never touched.
func TestRemotePrimaryModelWins(t *testing.T) {
	srv := remoteStubServer(t, 0)
	defer srv.Close()
	p := newTestRemoteProvider(t, srv.URL+"/v1")

	got, err := p.Generate("sys", []Message{{Role: "user", Content: "hi"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "remote reply" {
		t.Fatalf("want remote reply, got %q", got)
	}
}

// When the primary model fails, the next model in AI_REMOTE_MODEL is tried on
// the same endpoint.
func TestRemoteFallsToNextModelOnError(t *testing.T) {
	srv := remoteStubServer(t, 1)
	defer srv.Close()
	p := newTestRemoteProvider(t, srv.URL+"/v1")

	got, err := p.Generate("sys", []Message{{Role: "user", Content: "hi"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "remote reply" {
		t.Fatalf("want fallback model reply, got %q", got)
	}
}

// Every model failing surfaces an error naming how many were tried.
func TestRemoteAllModelsFail(t *testing.T) {
	srv := remoteStubServer(t, 99)
	defer srv.Close()
	p := newTestRemoteProvider(t, srv.URL+"/v1")

	_, err := p.Generate("sys", []Message{{Role: "user", Content: "hi"}})
	if err == nil {
		t.Fatal("expected an error when all models fail")
	}
	if !strings.Contains(err.Error(), "2 model(s) failed") {
		t.Fatalf("error should name the model count: %v", err)
	}
}

// Images are refused without a vision model rather than sent to a text model.
func TestRemoteRefusesImagesWithoutVisionModel(t *testing.T) {
	srv := remoteStubServer(t, 0)
	defer srv.Close()
	p := newTestRemoteProvider(t, srv.URL+"/v1")

	_, err := p.Generate("sys", []Message{{Role: "user", Images: []Image{{Data: []byte("x"), MimeType: "image/png"}}}})
	if err == nil || !strings.Contains(err.Error(), "AI_REMOTE_VISION_MODEL") {
		t.Fatalf("want vision-model error, got %v", err)
	}
}

// The bearer key from AI_REMOTE_API_KEY must reach the hosted endpoint.
func TestRemoteSendsAPIKey(t *testing.T) {
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"choices": []map[string]interface{}{
				{"message": map[string]string{"role": "assistant", "content": "ok"}},
			},
		})
	}))
	defer srv.Close()
	p := newTestRemoteProvider(t, srv.URL+"/v1")

	if _, err := p.Generate("sys", []Message{{Role: "user", Content: "hi"}}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotAuth != "Bearer test-key" {
		t.Fatalf("Authorization header missing or wrong: %q", gotAuth)
	}
}

// Name() should surface the endpoint and the model chain for the startup log.
func TestRemoteNameShowsModelChain(t *testing.T) {
	p := newTestRemoteProvider(t, "https://ai.nibros.space/v1")
	name := p.Name()
	if !strings.Contains(name, "ai/gpt-5.6-luna → bbai") {
		t.Fatalf("model chain missing from name: %q", name)
	}
}

// Compile-time interface checks.
var (
	_ Provider = (*RemoteProvider)(nil)
	_ Provider = (*FallbackProvider)(nil)
)
