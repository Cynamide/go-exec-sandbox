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
	Method       string `yaml:"method"`
	Path         string `yaml:"path"`
	BodyTemplate string `yaml:"body_template"`
}

type ResponseMapping struct {
	TextPath      string `yaml:"text_path"`
	ToolCallsPath string `yaml:"tool_calls_path"`
	UsagePath     string `yaml:"usage_path"`
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
	if err := ValidateMappings(c); err != nil {
		return err
	}
	return nil
}

func ValidateMappings(cfg Config) error {
	if cfg.ProviderKind != "custom_http" {
		return nil
	}

	var missing []string
	if cfg.RequestMapping.Method == "" {
		missing = append(missing, "request_mapping.method")
	}
	if cfg.RequestMapping.Path == "" {
		missing = append(missing, "request_mapping.path")
	}
	if cfg.RequestMapping.BodyTemplate == "" {
		missing = append(missing, "request_mapping.body_template")
	}
	if cfg.ResponseMapping.TextPath == "" {
		missing = append(missing, "response_mapping.text_path")
	}
	if len(missing) > 0 {
		return fmt.Errorf("model adapter config %q custom_http provider requires %s", cfg.ID, missing)
	}

	return nil
}
