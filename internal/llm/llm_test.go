package llm

import (
	"context"
	"strings"
	"testing"

	"gexec-sandbox/internal/config"
	"gexec-sandbox/internal/modeladapter"
)

type fakeAdapter struct {
	response modeladapter.ModelResponse
	err      error
	lastReq  modeladapter.ModelRequest
}

func (f *fakeAdapter) Generate(_ context.Context, req modeladapter.ModelRequest) (modeladapter.ModelResponse, error) {
	f.lastReq = req
	return f.response, f.err
}

func (f *fakeAdapter) HealthCheck(context.Context) error {
	return nil
}

func TestChatRequestUsesSystemAndUserMessages(t *testing.T) {
	client := &Client{}

	req := client.chatRequest("sum two numbers", "go")

	if got, want := len(req.Messages), 2; got != want {
		t.Fatalf("message count = %d, want %d", got, want)
	}

	if got, want := req.Messages[0].Content, systemPrompt; got != want {
		t.Fatalf("system prompt = %q, want %q", got, want)
	}

	if got := req.Messages[1].Content; !strings.Contains(got, "Write a go solution for:\nsum two numbers") {
		t.Fatalf("user prompt = %q, want it to include the problem and language", got)
	}
}

func TestGenerateCodeUsesAdapterResponse(t *testing.T) {
	adapter := &fakeAdapter{
		response: modeladapter.ModelResponse{Text: "package main"},
	}
	client := &Client{adapter: adapter}

	got, err := client.GenerateCode(context.Background(), "sum two numbers", "go")
	if err != nil {
		t.Fatalf("GenerateCode() error = %v", err)
	}

	if got != "package main" {
		t.Fatalf("GenerateCode() = %q, want %q", got, "package main")
	}

	if got, want := len(adapter.lastReq.Messages), 2; got != want {
		t.Fatalf("request message count = %d, want %d", got, want)
	}
}

func TestNewClientWithConfigAllowsEmptyModelWithExplicitHost(t *testing.T) {
	client, err := NewClientWithConfig(config.Config{
		OLLAMAHost: "http://localhost:11434",
	})
	if err != nil {
		t.Fatalf("NewClientWithConfig() error = %v, want nil", err)
	}

	if client == nil {
		t.Fatal("NewClientWithConfig() returned nil client")
	}
}

func TestWaitForOllamaWithConfigAllowsEmptyModelWithExplicitHost(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := WaitForOllamaWithConfig(ctx, config.Config{
		OLLAMAHost: "http://localhost:11434",
	})
	if err != context.Canceled {
		t.Fatalf("WaitForOllamaWithConfig() error = %v, want %v", err, context.Canceled)
	}
}
