package config

import "testing"

func TestLoadConfigUsesDefaultsAndEnvOverrides(t *testing.T) {
	t.Setenv("OLLAMA_MODEL", "test-model")
	t.Setenv("OLLAMA_HOST", "http://ollama.example:11434")

	cfg := LoadConfig()

	if cfg.DefaultTimeoutMS != 60000 {
		t.Fatalf("DefaultTimeoutMS = %d, want 60000", cfg.DefaultTimeoutMS)
	}

	if cfg.MaxMemoryMB != 256 {
		t.Fatalf("MaxMemoryMB = %d, want 256", cfg.MaxMemoryMB)
	}

	if cfg.OLLAMAHost != "http://ollama.example:11434" {
		t.Fatalf("OLLAMAHost = %q, want %q", cfg.OLLAMAHost, "http://ollama.example:11434")
	}

	if cfg.OLLAMAModel != "test-model" {
		t.Fatalf("OLLAMAModel = %q, want %q", cfg.OLLAMAModel, "test-model")
	}

	if got := cfg.Languages["go"]; got != "golang:1.24-alpine" {
		t.Fatalf("Languages[go] = %q, want %q", got, "golang:1.24-alpine")
	}
}

func TestLoadConfigPanicsWithoutModel(t *testing.T) {
	t.Setenv("OLLAMA_MODEL", "")

	defer func() {
		recovered := recover()
		if recovered == nil {
			t.Fatal("LoadConfig() did not panic without OLLAMA_MODEL")
		}
		if recovered != "OLLAMA_MODEL environment variable is required" {
			t.Fatalf("panic = %v, want OLLAMA_MODEL environment variable is required", recovered)
		}
	}()

	_ = LoadConfig()
}
