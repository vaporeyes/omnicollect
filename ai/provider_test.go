// ABOUTME: Provider request MIME, cancellation, and response-budget regressions.
// ABOUTME: Uses a fake RoundTripper; no external AI requests or credentials are needed.
package ai

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

type testTransport func(*http.Request) (*http.Response, error)

func (f testTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
func TestProvidersSendActualImageMIME(t *testing.T) {
	old := providerHTTPClient
	t.Cleanup(func() { providerHTTPClient = old })
	for _, provider := range []AIProvider{&AnthropicProvider{model: "test"}, &OpenAICompatProvider{model: "test", baseURL: "https://example.invalid"}} {
		providerHTTPClient = &http.Client{Transport: testTransport(func(r *http.Request) (*http.Response, error) {
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(string(body), "image/png") || strings.Contains(string(body), "image/jpeg") {
				t.Fatalf("incorrect MIME in request: %s", body)
			}
			return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(`{"content":[{"text":"ok"}],"choices":[{"message":{"content":"ok"}}]}`)), Header: http.Header{}}, nil
		})}
		result, err := provider.AnalyzeImage(context.Background(), "aW1hZ2U=", "image/png", "prompt")
		if err != nil || result != "ok" {
			t.Fatalf("provider failed: %q %v", result, err)
		}
		if _, err := provider.AnalyzeImage(context.Background(), "", "text/html", ""); err == nil {
			t.Fatal("invalid MIME accepted")
		}
	}
}
func TestProviderResponseBudget(t *testing.T) {
	if _, err := readProviderResponse(strings.NewReader(strings.Repeat("x", (2<<20)+1))); err == nil {
		t.Fatal("unbounded provider response")
	}
}
