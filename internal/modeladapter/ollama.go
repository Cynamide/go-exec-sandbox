package modeladapter

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/ollama/ollama/api"
)

type ollamaAdapter struct {
	client    *api.Client
	modelName string
}

func NewOllamaAdapter(cfg Config) (Adapter, error) {
	if err := validateOllamaConfig(cfg); err != nil {
		return nil, err
	}

	client, err := newOllamaClient(cfg.BaseURL)
	if err != nil {
		return nil, err
	}

	return &ollamaAdapter{
		client:    client,
		modelName: cfg.ModelName,
	}, nil
}

func (a *ollamaAdapter) Generate(ctx context.Context, req ModelRequest) (ModelResponse, error) {
	if a.modelName == "" {
		return ModelResponse{}, fmt.Errorf("ollama generate: model name is required")
	}

	chatReq := &api.ChatRequest{
		Model:    a.modelName,
		Messages: make([]api.Message, 0, len(req.Messages)),
		Stream:   new(bool),
		Options:  req.Params,
	}

	for _, message := range req.Messages {
		chatReq.Messages = append(chatReq.Messages, api.Message{
			Role:    message.Role,
			Content: message.Content,
		})
	}

	var response ModelResponse
	if err := a.client.Chat(ctx, chatReq, func(resp api.ChatResponse) error {
		response.Text += resp.Message.Content
		response.ToolCalls = mapToolCalls(resp.Message.ToolCalls)
		response.Usage = Usage{
			InputTokens:  resp.PromptEvalCount,
			OutputTokens: resp.EvalCount,
			TotalTokens:  resp.PromptEvalCount + resp.EvalCount,
		}
		return nil
	}); err != nil {
		return ModelResponse{}, fmt.Errorf("ollama generate: %w", err)
	}

	return response, nil
}

func (a *ollamaAdapter) HealthCheck(ctx context.Context) error {
	_, err := a.client.List(ctx)
	return err
}

func validateOllamaConfig(cfg Config) error {
	if cfg.ID == "" {
		return fmt.Errorf("model adapter config id is required")
	}
	if cfg.ProviderKind != "" && cfg.ProviderKind != "ollama" {
		return fmt.Errorf("model adapter config %q has unsupported provider kind %q", cfg.ID, cfg.ProviderKind)
	}
	return nil
}

func newOllamaClient(baseURL string) (*api.Client, error) {
	if baseURL == "" {
		return api.ClientFromEnvironment()
	}

	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}

	return api.NewClient(parsedURL, http.DefaultClient), nil
}

func mapToolCalls(toolCalls []api.ToolCall) []ToolCall {
	if len(toolCalls) == 0 {
		return nil
	}

	mapped := make([]ToolCall, 0, len(toolCalls))
	for _, toolCall := range toolCalls {
		mapped = append(mapped, ToolCall{
			ID:        toolCall.ID,
			Name:      toolCall.Function.Name,
			Arguments: toolCall.Function.Arguments.String(),
		})
	}

	return mapped
}
