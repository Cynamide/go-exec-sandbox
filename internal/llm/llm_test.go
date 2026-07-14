package llm

import (
	"strings"
	"testing"

	"gexec-sandbox/internal/config"
)

func TestChatRequestUsesInjectedConfig(t *testing.T) {
	client := &Client{
		cfg: config.Config{
			OLLAMAModel: "test-model",
		},
	}

	req := client.chatRequest("sum two numbers", "go")

	if got, want := req.Model, "test-model"; got != want {
		t.Fatalf("model = %q, want %q", got, want)
	}

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
