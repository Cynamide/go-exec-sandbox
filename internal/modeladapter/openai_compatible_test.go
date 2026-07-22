package modeladapter

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestOpenAICompatibleAdapterBuildsChatCompletionsRequest(t *testing.T) {
	adapter, err := NewOpenAICompatibleAdapter(Config{
		ID:           "local",
		ProviderKind: "openai_compatible",
		BaseURL:      "http://openai.test/v1",
		ModelName:    "local-model",
	})
	if err != nil {
		t.Fatalf("NewOpenAICompatibleAdapter() error = %v", err)
	}

	typedAdapter := adapter.(*openAICompatibleAdapter)
	typedAdapter.httpClient = &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			if r.URL.Path != "/v1/chat/completions" {
				t.Fatalf("path = %s", r.URL.Path)
			}
			if r.Method != http.MethodPost {
				t.Fatalf("method = %s, want POST", r.Method)
			}
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{"Content-Type": []string{"application/json"}},
				Body:       io.NopCloser(strings.NewReader(`{"choices":[{"message":{"content":"answer"}}]}`)),
			}, nil
	})}

	resp, err := adapter.Generate(context.Background(), ModelRequest{
		Messages: []Message{{Role: "user", Content: "hi"}},
	})
	if err != nil || resp.Text != "answer" {
		t.Fatalf("Generate() = %+v, %v", resp, err)
	}
}

func TestOpenAICompatibleAdapterUsesBearerAuthAndParams(t *testing.T) {
	t.Setenv("OPENAI_COMPATIBLE_API_KEY", "secret-key")

	adapter, err := NewOpenAICompatibleAdapter(Config{
		ID:           "local",
		ProviderKind: "openai_compatible",
		BaseURL:      "http://openai.test/v1",
		ModelName:    "local-model",
		APIKeyEnv:    "OPENAI_COMPATIBLE_API_KEY",
		Params: map[string]any{
			"temperature": 0.25,
			"max_tokens":  128,
		},
	})
	if err != nil {
		t.Fatalf("NewOpenAICompatibleAdapter() error = %v", err)
	}

	typedAdapter := adapter.(*openAICompatibleAdapter)
	typedAdapter.httpClient = &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			if got := r.Header.Get("Authorization"); got != "Bearer secret-key" {
				t.Fatalf("Authorization = %q, want bearer token", got)
			}
			if got := r.Header.Get("Content-Type"); got != "application/json" {
				t.Fatalf("Content-Type = %q, want application/json", got)
			}
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("ReadAll() error = %v", err)
			}
			gotBody := string(body)
			for _, want := range []string{`"temperature":0.25`, `"max_tokens":128`} {
				if !strings.Contains(gotBody, want) {
					t.Fatalf("request body = %s, want substring %s", gotBody, want)
				}
			}
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{"Content-Type": []string{"application/json"}},
				Body:       io.NopCloser(strings.NewReader(`{"choices":[{"message":{"content":"answer"}}],"usage":{"prompt_tokens":11,"completion_tokens":7,"total_tokens":18}}`)),
			}, nil
	})}

	resp, err := adapter.Generate(context.Background(), ModelRequest{
		Messages: []Message{{Role: "user", Content: "hi"}},
	})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	if resp.Usage != (Usage{InputTokens: 11, OutputTokens: 7, TotalTokens: 18}) {
		t.Fatalf("Usage = %+v, want prompt/completion totals", resp.Usage)
	}
}

func TestOpenAICompatibleAdapterRejectsMissingAPIKey(t *testing.T) {
	_, err := NewOpenAICompatibleAdapter(Config{
		ID:           "local",
		ProviderKind: "openai_compatible",
		BaseURL:      "http://openai.test/v1",
		ModelName:    "local-model",
		APIKeyEnv:    "MISSING_OPENAI_COMPATIBLE_API_KEY",
	})
	if err == nil {
		t.Fatal("NewOpenAICompatibleAdapter() error = nil, want missing API key error")
	}
}

func TestOpenAICompatibleAdapterAllowsUnsetAPIKeyEnv(t *testing.T) {
	adapter, err := NewOpenAICompatibleAdapter(Config{
		ID:           "local",
		ProviderKind: "openai_compatible",
		BaseURL:      "http://openai.test/v1",
		ModelName:    "local-model",
	})
	if err != nil {
		t.Fatalf("NewOpenAICompatibleAdapter() error = %v", err)
	}

	typedAdapter := adapter.(*openAICompatibleAdapter)
	typedAdapter.httpClient = &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			if got := r.Header.Get("Authorization"); got != "" {
				t.Fatalf("Authorization = %q, want empty", got)
			}
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{"Content-Type": []string{"application/json"}},
				Body:       io.NopCloser(strings.NewReader(`{"choices":[{"message":{"content":"answer"}}]}`)),
			}, nil
	})}

	if _, err := typedAdapter.Generate(context.Background(), ModelRequest{
		Messages: []Message{{Role: "user", Content: "hi"}},
	}); err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
}

func TestOpenAICompatibleAdapterHealthCheckUsesModelsEndpoint(t *testing.T) {
	adapter, err := NewOpenAICompatibleAdapter(Config{
		ID:           "local",
		ProviderKind: "openai_compatible",
		BaseURL:      "http://openai.test/v1",
		ModelName:    "local-model",
	})
	if err != nil {
		t.Fatalf("NewOpenAICompatibleAdapter() error = %v", err)
	}

	typedAdapter := adapter.(*openAICompatibleAdapter)
	typedAdapter.httpClient = &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			if r.URL.Path != "/v1/models" {
				t.Fatalf("path = %s, want /v1/models", r.URL.Path)
			}
			if r.Method != http.MethodGet {
				t.Fatalf("method = %s, want GET", r.Method)
			}
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{"Content-Type": []string{"application/json"}},
				Body:       io.NopCloser(strings.NewReader(`{"data":[]}`)),
			}, nil
	})}

	if err := typedAdapter.HealthCheck(context.Background()); err != nil {
		t.Fatalf("HealthCheck() error = %v", err)
	}
}

func TestOpenAICompatibleAdapterRejectsInvalidHost(t *testing.T) {
	_, err := NewOpenAICompatibleAdapter(Config{
		ID:           "local",
		ProviderKind: "openai_compatible",
		BaseURL:      "://bad",
		ModelName:    "local-model",
	})
	if err == nil {
		t.Fatal("NewOpenAICompatibleAdapter() error = nil, want invalid URL error")
	}
}

func TestOpenAICompatibleAdapterRejectsUnsupportedProviderKind(t *testing.T) {
	_, err := NewOpenAICompatibleAdapter(Config{
		ID:           "local",
		ProviderKind: "ollama",
		BaseURL:      "http://localhost:8080/v1",
		ModelName:    "local-model",
	})
	if err == nil {
		t.Fatal("NewOpenAICompatibleAdapter() error = nil, want unsupported provider error")
	}
}

func TestOpenAICompatibleAdapterRejectsNon200Responses(t *testing.T) {
	adapter, err := NewOpenAICompatibleAdapter(Config{
		ID:           "local",
		ProviderKind: "openai_compatible",
		BaseURL:      "http://openai.test/v1",
		ModelName:    "local-model",
	})
	if err != nil {
		t.Fatalf("NewOpenAICompatibleAdapter() error = %v", err)
	}

	typedAdapter := adapter.(*openAICompatibleAdapter)
	typedAdapter.httpClient = &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusBadRequest,
				Header:     http.Header{"Content-Type": []string{"text/plain"}},
				Body:       io.NopCloser(strings.NewReader("bad request\n")),
			}, nil
	})}

	_, err = typedAdapter.Generate(context.Background(), ModelRequest{
		Messages: []Message{{Role: "user", Content: "hi"}},
	})
	if err == nil {
		t.Fatal("Generate() error = nil, want non-200 error")
	}
}

func TestOpenAICompatibleAdapterPassesRequestBodyShape(t *testing.T) {
	adapter, err := NewOpenAICompatibleAdapter(Config{
		ID:           "local",
		ProviderKind: "openai_compatible",
		BaseURL:      "http://openai.test/v1",
		ModelName:    "local-model",
	})
	if err != nil {
		t.Fatalf("NewOpenAICompatibleAdapter() error = %v", err)
	}

	typedAdapter := adapter.(*openAICompatibleAdapter)
	typedAdapter.httpClient = &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("ReadAll() error = %v", err)
			}
			got := string(body)
			for _, want := range []string{
				`"model":"local-model"`,
				`"messages":[{"role":"system","content":"be precise"},{"role":"user","content":"hi"}]`,
			} {
				if !strings.Contains(got, want) {
					t.Fatalf("request body = %s, want substring %s", got, want)
				}
			}
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{"Content-Type": []string{"application/json"}},
				Body:       io.NopCloser(strings.NewReader(`{"choices":[{"message":{"content":"answer"}}]}`)),
			}, nil
	})}

	if _, err := typedAdapter.Generate(context.Background(), ModelRequest{
		Messages: []Message{{Role: "system", Content: "be precise"}, {Role: "user", Content: "hi"}},
	}); err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
}

func TestOpenAICompatibleAdapterJoinsPathsWithoutDoubleSlashes(t *testing.T) {
	adapter, err := NewOpenAICompatibleAdapter(Config{
		ID:           "local",
		ProviderKind: "openai_compatible",
		BaseURL:      "http://openai.test/v1/",
		ModelName:    "local-model",
	})
	if err != nil {
		t.Fatalf("NewOpenAICompatibleAdapter() error = %v", err)
	}

	typedAdapter := adapter.(*openAICompatibleAdapter)
	typedAdapter.httpClient = &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if r.URL.Path != "/v1/chat/completions" {
			t.Fatalf("path = %s, want /v1/chat/completions", r.URL.Path)
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       io.NopCloser(strings.NewReader(`{"choices":[{"message":{"content":"answer"}}]}`)),
		}, nil
	})}

	if _, err := typedAdapter.Generate(context.Background(), ModelRequest{
		Messages: []Message{{Role: "user", Content: "hi"}},
	}); err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
}

func TestOpenAICompatibleAdapterRejectsInvalidTemperatureParamType(t *testing.T) {
	_, err := NewOpenAICompatibleAdapter(Config{
		ID:           "local",
		ProviderKind: "openai_compatible",
		BaseURL:      "http://openai.test/v1",
		ModelName:    "local-model",
		Params: map[string]any{
			"temperature": "0.25",
		},
	})
	if err == nil {
		t.Fatal("NewOpenAICompatibleAdapter() error = nil, want invalid temperature error")
	}
}

func TestOpenAICompatibleAdapterRejectsInvalidMaxTokensParamType(t *testing.T) {
	_, err := NewOpenAICompatibleAdapter(Config{
		ID:           "local",
		ProviderKind: "openai_compatible",
		BaseURL:      "http://openai.test/v1",
		ModelName:    "local-model",
		Params: map[string]any{
			"max_tokens": 3.5,
		},
	})
	if err == nil {
		t.Fatal("NewOpenAICompatibleAdapter() error = nil, want invalid max_tokens error")
	}
}
