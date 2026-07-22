package modeladapter

import (
	"context"
	"io"
	"math"
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
			Body:       io.NopCloser(strings.NewReader(`{"data":[{"id":"local-model"}]}`)),
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

func TestOpenAICompatibleAdapterRejectsOutOfRangeTemperatureParams(t *testing.T) {
	for name, temperature := range map[string]float64{
		"below minimum":     -0.1,
		"above maximum":     2.1,
		"NaN":               math.NaN(),
		"positive infinity": math.Inf(1),
		"negative infinity": math.Inf(-1),
	} {
		t.Run(name, func(t *testing.T) {
			_, err := NewOpenAICompatibleAdapter(Config{
				ID:           "local",
				ProviderKind: "openai_compatible",
				BaseURL:      "http://openai.test/v1",
				ModelName:    "local-model",
				Params: map[string]any{
					"temperature": temperature,
				},
			})
			if err == nil {
				t.Fatal("NewOpenAICompatibleAdapter() error = nil, want invalid temperature error")
			}
		})
	}
}

func TestOpenAICompatibleAdapterRejectsInvalidMaxTokensValues(t *testing.T) {
	for name, maxTokens := range map[string]any{
		"negative":          -1,
		"zero":              0,
		"non-integer float": 1.5,
		"outside int range": math.MaxFloat64,
	} {
		t.Run(name, func(t *testing.T) {
			_, err := NewOpenAICompatibleAdapter(Config{
				ID:           "local",
				ProviderKind: "openai_compatible",
				BaseURL:      "http://openai.test/v1",
				ModelName:    "local-model",
				Params: map[string]any{
					"max_tokens": maxTokens,
				},
			})
			if err == nil {
				t.Fatal("NewOpenAICompatibleAdapter() error = nil, want invalid max_tokens error")
			}
		})
	}
}

func TestOpenAICompatibleAdapterRejectsInvalidBaseURL(t *testing.T) {
	for name, baseURL := range map[string]string{
		"relative":       "/v1",
		"schemeless":     "openai.test/v1",
		"ftp":            "ftp://openai.test/v1",
		"empty hostname": "http://:8080/v1",
		"userinfo":       "https://user:secret@openai.test/v1",
		"fragment":       "https://openai.test/v1#models",
		"query secret":   "https://openai.test/v1?api_key=secret",
	} {
		t.Run(name, func(t *testing.T) {
			_, err := NewOpenAICompatibleAdapter(Config{
				ID:           "local",
				ProviderKind: "openai_compatible",
				BaseURL:      baseURL,
				ModelName:    "local-model",
			})
			if err == nil {
				t.Fatal("NewOpenAICompatibleAdapter() error = nil, want invalid base URL error")
			}
		})
	}
}

func TestOpenAICompatibleAdapterRejectsInvalidEndpointURLWithEndpointWording(t *testing.T) {
	for name, endpointURL := range map[string]string{
		"empty hostname": "http://:8080/chat/completions",
		"fragment":       "https://inference.test/chat/completions#generate",
		"query secret":   "https://inference.test/chat/completions?access_token=secret",
	} {
		t.Run(name, func(t *testing.T) {
			_, err := NewOpenAICompatibleAdapter(Config{
				ID:           "local",
				ProviderKind: "openai_compatible",
				BaseURL:      "https://models.test/v1",
				EndpointURL:  endpointURL,
				ModelName:    "local-model",
			})
			if err == nil {
				t.Fatal("NewOpenAICompatibleAdapter() error = nil, want invalid endpoint URL error")
			}
			if !strings.Contains(err.Error(), "endpoint URL") || strings.Contains(err.Error(), "base URL") {
				t.Fatalf("NewOpenAICompatibleAdapter() error = %q, want endpoint URL wording", err)
			}
		})
	}
}

func TestOpenAICompatibleAdapterSendsGenerationToConfiguredEndpointURL(t *testing.T) {
	adapter, err := NewOpenAICompatibleAdapter(Config{
		ID:           "local",
		ProviderKind: "openai_compatible",
		BaseURL:      "https://models.test/v1",
		EndpointURL:  "https://inference.test/custom/generate",
		ModelName:    "local-model",
	})
	if err != nil {
		t.Fatalf("NewOpenAICompatibleAdapter() error = %v", err)
	}

	typedAdapter := adapter.(*openAICompatibleAdapter)
	typedAdapter.httpClient = &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if r.URL.Host != "inference.test" || r.URL.Path != "/custom/generate" {
			t.Fatalf("generation URL = %s, want https://inference.test/custom/generate", r.URL)
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       io.NopCloser(strings.NewReader(`{"choices":[{"message":{"content":"answer"}}]}`)),
		}, nil
	})}

	if _, err := adapter.Generate(context.Background(), ModelRequest{}); err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
}

func TestOpenAICompatibleAdapterBoundsProviderErrorBody(t *testing.T) {
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
	typedAdapter.httpClient = &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusBadRequest,
			Body:       io.NopCloser(strings.NewReader(strings.Repeat("x", 100_000) + "secret-tail")),
		}, nil
	})}

	_, err = adapter.Generate(context.Background(), ModelRequest{})
	if err == nil {
		t.Fatal("Generate() error = nil, want provider error")
	}
	if strings.Contains(err.Error(), "secret-tail") {
		t.Fatalf("Generate() error contains unbounded response tail")
	}
	if len(err.Error()) > 70_000 {
		t.Fatalf("Generate() error length = %d, want bounded error", len(err.Error()))
	}
}

func TestOpenAICompatibleHealthCheckRejectsMissingConfiguredModel(t *testing.T) {
	adapter, err := NewOpenAICompatibleAdapter(Config{
		ID:           "local",
		ProviderKind: "openai_compatible",
		BaseURL:      "http://openai.test/v1",
		ModelName:    "wanted-model",
	})
	if err != nil {
		t.Fatalf("NewOpenAICompatibleAdapter() error = %v", err)
	}

	typedAdapter := adapter.(*openAICompatibleAdapter)
	typedAdapter.httpClient = &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(`{"data":[{"id":"other-model"}]}`)),
		}, nil
	})}

	if err := adapter.HealthCheck(context.Background()); err == nil || !strings.Contains(err.Error(), "wanted-model") {
		t.Fatalf("HealthCheck() error = %v, want missing configured model error", err)
	}
}

func TestNewAdapterDispatchesByProviderKind(t *testing.T) {
	adapter, err := New(Config{
		ID:           "local",
		ProviderKind: "openai_compatible",
		BaseURL:      "http://openai.test/v1",
		ModelName:    "local-model",
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if _, ok := adapter.(*openAICompatibleAdapter); !ok {
		t.Fatalf("New() adapter = %T, want openAICompatibleAdapter", adapter)
	}
}

func TestNewAdapterValidatesBearerAuthEnvWithoutManifestResolution(t *testing.T) {
	const apiKeyEnv = "MODELADAPTER_TEST_MISSING_FACTORY_KEY"
	t.Setenv(apiKeyEnv, "")

	_, err := New(Config{
		ID:           "local",
		ProviderKind: "openai_compatible",
		BaseURL:      "http://openai.test/v1",
		ModelName:    "local-model",
		Auth: AuthConfig{
			Type: "bearer_env",
			Env:  apiKeyEnv,
		},
	})
	if err == nil || !strings.Contains(err.Error(), apiKeyEnv) {
		t.Fatalf("New() error = %v, want missing bearer auth env error", err)
	}
}

func TestNewAdapterRejectsMalformedCanonicalAuth(t *testing.T) {
	for name, auth := range map[string]AuthConfig{
		"blank type with env":    {Env: "UNUSED_API_KEY"},
		"blank type with header": {Header: "Authorization"},
		"none with env":          {Type: "none", Env: "UNUSED_API_KEY"},
		"none with header":       {Type: "none", Header: "Authorization"},
		"bearer with header":     {Type: "bearer_env", Env: "REMOTE_MODEL_API_KEY", Header: "X-API-Key"},
	} {
		t.Run(name, func(t *testing.T) {
			_, err := New(Config{
				ID:           "local",
				ProviderKind: "openai_compatible",
				BaseURL:      "http://openai.test/v1",
				ModelName:    "local-model",
				Auth:         auth,
			})
			if err == nil || !strings.Contains(err.Error(), "auth") {
				t.Fatalf("New() error = %v, want malformed auth error", err)
			}
		})
	}
}

func TestNewAdapterAuthNoneSuppressesProviderAPIKey(t *testing.T) {
	t.Setenv("PROVIDER_API_KEY", "secret-key")

	adapter, err := New(Config{
		ID:           "local",
		ProviderKind: "openai_compatible",
		BaseURL:      "http://openai.test/v1",
		ModelName:    "local-model",
		APIKeyEnv:    "PROVIDER_API_KEY",
		Auth:         AuthConfig{Type: "none"},
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	typedAdapter := adapter.(*openAICompatibleAdapter)
	typedAdapter.httpClient = &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if got := r.Header.Get("Authorization"); got != "" {
			t.Fatalf("Authorization = %q, want empty for auth none", got)
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       io.NopCloser(strings.NewReader(`{"choices":[{"message":{"content":"answer"}}]}`)),
		}, nil
	})}

	if _, err := adapter.Generate(context.Background(), ModelRequest{}); err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
}

func TestOpenAICompatibleAdapterRejectsMalformedCanonicalAuth(t *testing.T) {
	for name, auth := range map[string]AuthConfig{
		"none with env":      {Type: "none", Env: "UNUSED_API_KEY"},
		"none with header":   {Type: "none", Header: "Authorization"},
		"bearer with header": {Type: "bearer_env", Env: "REMOTE_MODEL_API_KEY", Header: "X-API-Key"},
	} {
		t.Run(name, func(t *testing.T) {
			_, err := NewOpenAICompatibleAdapter(Config{
				ID:           "local",
				ProviderKind: "openai_compatible",
				BaseURL:      "http://openai.test/v1",
				ModelName:    "local-model",
				Auth:         auth,
			})
			if err == nil || !strings.Contains(err.Error(), "auth") {
				t.Fatalf("NewOpenAICompatibleAdapter() error = %v, want malformed auth error", err)
			}
		})
	}
}

func TestOpenAICompatibleAdapterAuthNoneSuppressesProviderAPIKey(t *testing.T) {
	t.Setenv("PROVIDER_API_KEY", "secret-key")

	adapter, err := NewOpenAICompatibleAdapter(Config{
		ID:           "local",
		ProviderKind: "openai_compatible",
		BaseURL:      "http://openai.test/v1",
		ModelName:    "local-model",
		APIKeyEnv:    "PROVIDER_API_KEY",
		Auth:         AuthConfig{Type: "none"},
	})
	if err != nil {
		t.Fatalf("NewOpenAICompatibleAdapter() error = %v", err)
	}

	typedAdapter := adapter.(*openAICompatibleAdapter)
	typedAdapter.httpClient = &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if got := r.Header.Get("Authorization"); got != "" {
			t.Fatalf("Authorization = %q, want empty for auth none", got)
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       io.NopCloser(strings.NewReader(`{"choices":[{"message":{"content":"answer"}}]}`)),
		}, nil
	})}

	if _, err := adapter.Generate(context.Background(), ModelRequest{}); err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
}

func TestOpenAICompatibleAdapterDirectEndpointHealthCheckIsNoop(t *testing.T) {
	adapter, err := NewOpenAICompatibleAdapter(Config{
		ID:           "local",
		ProviderKind: "openai_compatible",
		EndpointURL:  "http://openai.test/custom/generate",
		ModelName:    "local-model",
	})
	if err != nil {
		t.Fatalf("NewOpenAICompatibleAdapter() error = %v", err)
	}

	typedAdapter := adapter.(*openAICompatibleAdapter)
	typedAdapter.httpClient = &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		t.Fatal("HealthCheck() should not probe direct endpoint-only configs")
		return nil, nil
	})}

	if err := adapter.HealthCheck(context.Background()); err != nil {
		t.Fatalf("HealthCheck() error = %v, want nil", err)
	}
}

func TestOpenAICompatibleAdapterRejectsUnsupportedConfigParam(t *testing.T) {
	_, err := NewOpenAICompatibleAdapter(Config{
		ID:           "local",
		ProviderKind: "openai_compatible",
		BaseURL:      "http://openai.test/v1",
		ModelName:    "local-model",
		Params: map[string]any{
			"top_p": 0.9,
		},
	})
	if err == nil {
		t.Fatal("NewOpenAICompatibleAdapter() error = nil, want unsupported param error")
	}
	if !strings.Contains(err.Error(), "unsupported param \"top_p\"") {
		t.Fatalf("NewOpenAICompatibleAdapter() error = %v, want unsupported top_p param error", err)
	}
}

func TestOpenAICompatibleAdapterRejectsUnsupportedRequestParam(t *testing.T) {
	adapter, err := NewOpenAICompatibleAdapter(Config{
		ID:           "local",
		ProviderKind: "openai_compatible",
		BaseURL:      "http://openai.test/v1",
		ModelName:    "local-model",
	})
	if err != nil {
		t.Fatalf("NewOpenAICompatibleAdapter() error = %v", err)
	}

	_, err = adapter.Generate(context.Background(), ModelRequest{
		Messages: []Message{{Role: "user", Content: "hi"}},
		Params: map[string]any{
			"foo": "bar",
		},
	})
	if err == nil {
		t.Fatal("Generate() error = nil, want unsupported param error")
	}
	if !strings.Contains(err.Error(), "unsupported param \"foo\"") {
		t.Fatalf("Generate() error = %v, want unsupported foo param error", err)
	}
}

func TestOpenAICompatibleAdapterParsesToolCallsWithNullContent(t *testing.T) {
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
	typedAdapter.httpClient = &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body: io.NopCloser(strings.NewReader(`{
				"choices": [{
					"message": {
						"content": null,
						"tool_calls": [{
							"id": "call_1",
							"function": {"name": "lookup", "arguments": "{\"query\":\"weather\"}"}
						}]
					}
				}]
			}`)),
		}, nil
	})}

	response, err := adapter.Generate(context.Background(), ModelRequest{
		Messages: []Message{{Role: "user", Content: "hi"}},
	})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	if response.Text != "" {
		t.Fatalf("Text = %q, want empty", response.Text)
	}
	wantToolCalls := []ToolCall{{ID: "call_1", Name: "lookup", Arguments: `{"query":"weather"}`}}
	if !equalToolCalls(response.ToolCalls, wantToolCalls) {
		t.Fatalf("ToolCalls = %#v, want %#v", response.ToolCalls, wantToolCalls)
	}
}

func equalToolCalls(got, want []ToolCall) bool {
	if len(got) != len(want) {
		return false
	}
	for index := range got {
		if got[index] != want[index] {
			return false
		}
	}
	return true
}
