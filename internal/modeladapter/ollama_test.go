package modeladapter

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/ollama/ollama/api"
)

func TestNewOllamaAdapterRejectsInvalidHost(t *testing.T) {
	_, err := NewOllamaAdapter(Config{
		ID:           "qwen",
		ProviderKind: "ollama",
		ModelName:    "qwen3:4b",
		BaseURL:      "://bad",
	})
	if err == nil {
		t.Fatal("NewOllamaAdapter() error = nil, want invalid URL error")
	}
}

func TestNewOllamaAdapterAllowsEmptyModelName(t *testing.T) {
	adapter, err := NewOllamaAdapter(Config{
		ID:           "ollama",
		ProviderKind: "ollama",
		BaseURL:      "http://localhost:11434",
	})
	if err != nil {
		t.Fatalf("NewOllamaAdapter() error = %v, want nil", err)
	}

	if adapter == nil {
		t.Fatal("NewOllamaAdapter() returned nil adapter")
	}
}

func TestOllamaAdapterHealthCheckUsesListEndpoint(t *testing.T) {
	var (
		gotMethod string
		gotPath   string
	)

	baseURL, err := url.Parse("http://ollama.test")
	if err != nil {
		t.Fatalf("url.Parse() error = %v, want nil", err)
	}

	client := api.NewClient(baseURL, &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       io.NopCloser(strings.NewReader(`{"models":[]}`)),
		}, nil
	})})

	adapter := &ollamaAdapter{
		client: client,
	}

	if err := adapter.HealthCheck(context.Background()); err != nil {
		t.Fatalf("HealthCheck() error = %v, want nil", err)
	}

	if gotMethod != http.MethodGet {
		t.Fatalf("HealthCheck() method = %q, want %q", gotMethod, http.MethodGet)
	}

	if gotPath != "/api/tags" {
		t.Fatalf("HealthCheck() path = %q, want %q", gotPath, "/api/tags")
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return fn(r)
}
