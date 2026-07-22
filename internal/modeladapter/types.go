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

type RequestMapping struct {
	MessagesField    string `yaml:"messages_field"`
	TemperatureField string `yaml:"temperature_field"`
	MaxTokensField   string `yaml:"max_tokens_field"`
	ModelField       string `yaml:"model_field"`
}

type ResponseMapping struct {
	TextPath                  string `yaml:"text_path"`
	FinishReasonPath          string `yaml:"finish_reason_path"`
	UsagePromptTokensPath     string `yaml:"usage_prompt_tokens_path"`
	UsageCompletionTokensPath string `yaml:"usage_completion_tokens_path"`
}

type Capabilities struct {
	ToolUse          bool `yaml:"tool_use"`
	FileEditing      bool `yaml:"file_editing"`
	Browser          bool `yaml:"browser"`
	Multimodal       bool `yaml:"multimodal"`
	VisualReasoning  bool `yaml:"visual_reasoning"`
	TerminalSession  bool `yaml:"terminal_session"`
	Spreadsheet      bool `yaml:"spreadsheet"`
	Notebook         bool `yaml:"notebook"`
	Conversation     bool `yaml:"conversation"`
	StructuredOutput bool `yaml:"structured_output"`
	Judge            bool `yaml:"judge"`
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
