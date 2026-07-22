package modeladapter

import (
	"context"
	"encoding/json"
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

func TestNewOllamaAdapterRejectsUnsafeOrNonHTTPBaseURL(t *testing.T) {
	for name, baseURL := range map[string]string{
		"relative":           "/api",
		"schemeless":         "ollama.test:11434",
		"non-http":           "ftp://ollama.test",
		"empty hostname":     "http://:8080",
		"inline credentials": "http://user:secret@ollama.test",
		"fragment":           "http://ollama.test#tags",
		"query secret":       "http://ollama.test?api_token=secret",
	} {
		t.Run(name, func(t *testing.T) {
			_, err := NewOllamaAdapter(Config{ID: "qwen", ProviderKind: "ollama", ModelName: "qwen3:4b", BaseURL: baseURL})
			if err == nil {
				t.Fatal("NewOllamaAdapter() error = nil, want invalid base URL error")
			}
		})
	}
}

func TestOllamaAdapterMergesConfigAndRequestParams(t *testing.T) {
	baseURL, err := url.Parse("http://ollama.test")
	if err != nil {
		t.Fatalf("url.Parse() error = %v", err)
	}

	var options map[string]any
	client := api.NewClient(baseURL, &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		var body struct {
			Options map[string]any `json:"options"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("Decode() error = %v", err)
		}
		options = body.Options
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"application/x-ndjson"}},
			Body:       io.NopCloser(strings.NewReader(`{"message":{"role":"assistant","content":"ok"},"done":true}` + "\n")),
		}, nil
	})})

	adapter, err := NewOllamaAdapter(Config{
		ID:           "qwen",
		ProviderKind: "ollama",
		ModelName:    "qwen3:4b",
		BaseURL:      "http://ollama.test",
		Params: map[string]any{
			"temperature": 0.2,
			"num_predict": 128,
		},
	})
	if err != nil {
		t.Fatalf("NewOllamaAdapter() error = %v", err)
	}
	typedAdapter := adapter.(*ollamaAdapter)
	typedAdapter.client = client
	_, err = adapter.Generate(context.Background(), ModelRequest{Params: map[string]any{
		"temperature": 0.7,
	}})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	if options["temperature"] != 0.7 {
		t.Fatalf("temperature = %#v, want request override 0.7", options["temperature"])
	}
	if options["num_predict"] != float64(128) {
		t.Fatalf("num_predict = %#v, want config value 128", options["num_predict"])
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
