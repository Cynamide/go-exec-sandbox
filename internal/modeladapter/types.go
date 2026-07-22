package modeladapter

import (
	"context"
	"fmt"
)

type Message struct {
	Role    string
	Content string
}

type ModelRequest struct {
	Messages []Message
	Params   map[string]any
}

type Usage struct {
	InputTokens  int
	OutputTokens int
	TotalTokens  int
}

type ToolCall struct {
	ID        string
	Name      string
	Arguments string
}

type ModelResponse struct {
	Text      string
	ToolCalls []ToolCall
	Usage     Usage
}

type Adapter interface {
	Generate(ctx context.Context, req ModelRequest) (ModelResponse, error)
	HealthCheck(ctx context.Context) error
}

type RequestMapping struct{}

type ResponseMapping struct{}

type Capabilities struct {
	Judge      bool
	Browser    bool
	Multimodal bool
}

type Config struct {
	ID              string
	ProviderID      string
	ProviderKind    string
	ModelName       string
	BaseURL         string
	APIKeyEnv       string
	Params          map[string]any
	RequestMapping  RequestMapping
	ResponseMapping ResponseMapping
	Capabilities    Capabilities
}

func (c Config) Validate() error {
	if c.ID == "" {
		return fmt.Errorf("model adapter config id is required")
	}
	if c.ModelName == "" {
		return fmt.Errorf("model adapter config %q missing model name", c.ID)
	}
	return nil
}
