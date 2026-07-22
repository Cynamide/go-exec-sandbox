package modeladapter_test

import (
	"testing"

	"gexec-sandbox/internal/modeladapter"
)

func TestNewOllamaAdapterRejectsInvalidHost(t *testing.T) {
	_, err := modeladapter.NewOllamaAdapter(modeladapter.Config{
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
	adapter, err := modeladapter.NewOllamaAdapter(modeladapter.Config{
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
