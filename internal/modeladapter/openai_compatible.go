package modeladapter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
)

type openAICompatibleAdapter struct {
	baseURL    *url.URL
	httpClient *http.Client
	modelName  string
	configID   string
	apiKeyEnv  string
	params     map[string]any
}

type openAICompatibleChatRequest struct {
	Model       string                    `json:"model"`
	Messages    []openAICompatibleMessage `json:"messages"`
	Temperature *float64                  `json:"temperature,omitempty"`
	MaxTokens   *int                      `json:"max_tokens,omitempty"`
}

type openAICompatibleMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAICompatibleChatResponse struct {
	Choices []struct {
		Message struct {
			Content   *string `json:"content"`
			ToolCalls []struct {
				ID       string `json:"id"`
				Function struct {
					Name      string `json:"name"`
					Arguments string `json:"arguments"`
				} `json:"function"`
			} `json:"tool_calls"`
		} `json:"message"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

func NewOpenAICompatibleAdapter(cfg Config) (Adapter, error) {
	if err := validateOpenAICompatibleConfig(cfg); err != nil {
		return nil, err
	}
	if _, err := requiredAPIKey(cfg.APIKeyEnv); err != nil {
		return nil, err
	}
	if err := validateOpenAICompatibleParams(cfg.ID, cfg.Params); err != nil {
		return nil, err
	}

	baseURL, err := url.Parse(cfg.BaseURL)
	if err != nil {
		return nil, err
	}
	if err := validateOpenAICompatibleBaseURL(cfg.ID, baseURL); err != nil {
		return nil, err
	}

	return &openAICompatibleAdapter{
		baseURL:    baseURL,
		httpClient: http.DefaultClient,
		modelName:  cfg.ModelName,
		configID:   cfg.ID,
		apiKeyEnv:  cfg.APIKeyEnv,
		params:     cloneAnyMap(cfg.Params),
	}, nil
}

func (a *openAICompatibleAdapter) Generate(ctx context.Context, req ModelRequest) (ModelResponse, error) {
	body, err := a.chatRequestBody(req)
	if err != nil {
		return ModelResponse{}, err
	}

	httpReq, err := a.newJSONRequest(ctx, http.MethodPost, "/chat/completions", body)
	if err != nil {
		return ModelResponse{}, err
	}

	resp, err := a.httpClient.Do(httpReq)
	if err != nil {
		return ModelResponse{}, fmt.Errorf("openai-compatible generate: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(resp.Body)
		return ModelResponse{}, fmt.Errorf("openai-compatible generate: unexpected status %d: %s", resp.StatusCode, string(bytes.TrimSpace(raw)))
	}

	var payload openAICompatibleChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return ModelResponse{}, fmt.Errorf("openai-compatible generate: decode response: %w", err)
	}

	if len(payload.Choices) == 0 {
		return ModelResponse{}, fmt.Errorf("openai-compatible generate: response contained no choices")
	}

	toolCalls := make([]ToolCall, 0, len(payload.Choices[0].Message.ToolCalls))
	for _, toolCall := range payload.Choices[0].Message.ToolCalls {
		toolCalls = append(toolCalls, ToolCall{
			ID:        toolCall.ID,
			Name:      toolCall.Function.Name,
			Arguments: toolCall.Function.Arguments,
		})
	}

	text := ""
	if content := payload.Choices[0].Message.Content; content != nil {
		text = *content
	}

	return ModelResponse{
		Text:      text,
		ToolCalls: toolCalls,
		Usage: Usage{
			InputTokens:  payload.Usage.PromptTokens,
			OutputTokens: payload.Usage.CompletionTokens,
			TotalTokens:  payload.Usage.TotalTokens,
		},
	}, nil
}

func (a *openAICompatibleAdapter) HealthCheck(ctx context.Context) error {
	httpReq, err := a.newJSONRequest(ctx, http.MethodGet, "/models", nil)
	if err != nil {
		return err
	}

	resp, err := a.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("openai-compatible health check: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("openai-compatible health check: unexpected status %d: %s", resp.StatusCode, string(bytes.TrimSpace(raw)))
	}

	return nil
}

func (a *openAICompatibleAdapter) chatRequestBody(req ModelRequest) ([]byte, error) {
	if err := validateOpenAICompatibleParams(a.configID, req.Params); err != nil {
		return nil, err
	}

	params := make(map[string]any, len(a.params)+len(req.Params))
	for key, value := range a.params {
		params[key] = value
	}
	for key, value := range req.Params {
		params[key] = value
	}

	payload := openAICompatibleChatRequest{
		Model:    a.modelName,
		Messages: make([]openAICompatibleMessage, 0, len(req.Messages)),
	}
	for _, message := range req.Messages {
		payload.Messages = append(payload.Messages, openAICompatibleMessage{
			Role:    message.Role,
			Content: message.Content,
		})
	}

	if temperature, ok := floatParam(params["temperature"]); ok {
		payload.Temperature = &temperature
	}
	if maxTokens, ok := intParam(params["max_tokens"]); ok {
		payload.MaxTokens = &maxTokens
	}

	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("openai-compatible generate: encode request: %w", err)
	}

	return raw, nil
}

func (a *openAICompatibleAdapter) newJSONRequest(ctx context.Context, method string, path string, body []byte) (*http.Request, error) {
	endpoint := *a.baseURL
	endpoint.Path = joinURLPath(endpoint.Path, path)

	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, endpoint.String(), bodyReader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	apiKey, err := requiredAPIKey(a.apiKeyEnv)
	if err != nil {
		return nil, err
	}
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}

	return req, nil
}

func validateOpenAICompatibleConfig(cfg Config) error {
	if cfg.ID == "" {
		return fmt.Errorf("model adapter config id is required")
	}
	if cfg.ProviderKind != "" && cfg.ProviderKind != "openai_compatible" {
		return fmt.Errorf("model adapter config %q has unsupported provider kind %q", cfg.ID, cfg.ProviderKind)
	}
	if cfg.BaseURL == "" {
		return fmt.Errorf("model adapter config %q missing base URL", cfg.ID)
	}
	if cfg.ModelName == "" {
		return fmt.Errorf("model adapter config %q missing model name", cfg.ID)
	}
	return nil
}

func validateOpenAICompatibleBaseURL(configID string, baseURL *url.URL) error {
	if baseURL.Scheme != "http" && baseURL.Scheme != "https" {
		return fmt.Errorf("model adapter config %q base URL must use http or https", configID)
	}
	if baseURL.Host == "" {
		return fmt.Errorf("model adapter config %q base URL must be absolute", configID)
	}
	return nil
}

func requiredAPIKey(envName string) (string, error) {
	if envName == "" {
		return "", nil
	}
	value := os.Getenv(envName)
	if value == "" {
		return "", fmt.Errorf("model adapter missing API key environment variable %q", envName)
	}
	return value, nil
}

func validateOpenAICompatibleParams(configID string, params map[string]any) error {
	for name, value := range params {
		switch name {
		case "temperature":
			temperature, valid := floatParam(value)
			if !valid || math.IsNaN(temperature) || math.IsInf(temperature, 0) || temperature < 0 || temperature > 2 {
				return fmt.Errorf("model adapter config %q has invalid temperature param of type %T", configID, value)
			}
		case "max_tokens":
			maxTokens, valid := intParam(value)
			if !valid || maxTokens <= 0 {
				return fmt.Errorf("model adapter config %q has invalid max_tokens param of type %T", configID, value)
			}
		default:
			return fmt.Errorf("model adapter config %q has unsupported param %q", configID, name)
		}
	}
	return nil
}

func floatParam(value any) (float64, bool) {
	switch typed := value.(type) {
	case float64:
		return typed, true
	case float32:
		return float64(typed), true
	case int:
		return float64(typed), true
	case int64:
		return float64(typed), true
	default:
		return 0, false
	}
}

func intParam(value any) (int, bool) {
	switch typed := value.(type) {
	case int:
		return typed, true
	case int64:
		if typed > int64(maxInt()) || typed < int64(minInt()) {
			return 0, false
		}
		return int(typed), true
	case float64:
		if math.IsNaN(typed) || math.IsInf(typed, 0) || math.Trunc(typed) != typed || !floatFitsInt(typed) {
			return 0, false
		}
		return int(typed), true
	default:
		return 0, false
	}
}

func maxInt() int {
	return int(^uint(0) >> 1)
}

func minInt() int {
	return -maxInt() - 1
}

func floatFitsInt(value float64) bool {
	if strconv.IntSize == 64 {
		return value >= -math.Exp2(63) && value < math.Exp2(63)
	}
	return value >= float64(minInt()) && value <= float64(maxInt())
}

func joinURLPath(basePath string, requestPath string) string {
	if basePath == "" {
		return requestPath
	}
	return path.Join(basePath, requestPath)
}

func cloneAnyMap(source map[string]any) map[string]any {
	if len(source) == 0 {
		return nil
	}

	cloned := make(map[string]any, len(source))
	for key, value := range source {
		cloned[key] = value
	}
	return cloned
}
