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
