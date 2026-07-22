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
	baseURL     *url.URL
	endpointURL *url.URL
	httpClient  *http.Client
	modelName   string
	configID    string
	apiKeyEnv   string
	params      map[string]any
}

const maxProviderErrorBody = 64 << 10

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
	normalized, err := normalizeOpenAICompatibleAuth(cfg)
	if err != nil {
		return nil, err
	}
	cfg = normalized

	if err := validateOpenAICompatibleConfig(cfg); err != nil {
		return nil, err
	}
	if _, err := requiredAPIKey(cfg.APIKeyEnv); err != nil {
		return nil, err
	}
	if err := validateOpenAICompatibleParams(cfg.ID, cfg.Params); err != nil {
		return nil, err
	}

	var baseURL *url.URL
	if cfg.BaseURL != "" {
		parsed, err := parseOpenAICompatibleURL(cfg.ID, "base URL", cfg.BaseURL)
		if err != nil {
			return nil, err
		}
		baseURL = parsed
	}

	var endpointURL *url.URL
	if cfg.EndpointURL != "" {
		parsed, err := parseOpenAICompatibleURL(cfg.ID, "endpoint URL", cfg.EndpointURL)
		if err != nil {
			return nil, err
		}
		endpointURL = parsed
	}

	return &openAICompatibleAdapter{
		baseURL:     baseURL,
		endpointURL: endpointURL,
		httpClient:  http.DefaultClient,
		modelName:   cfg.ModelName,
		configID:    cfg.ID,
		apiKeyEnv:   cfg.APIKeyEnv,
		params:      cloneAnyMap(cfg.Params),
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
		return ModelResponse{}, fmt.Errorf("openai-compatible generate: unexpected status %d: %s", resp.StatusCode, readProviderErrorBody(resp.Body))
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
	if a.baseURL == nil {
		// Direct generation endpoints do not imply a safe model-list endpoint.
		// Treat construction-time URL/auth validation as the startup check.
		return nil
	}
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
		return fmt.Errorf("openai-compatible health check: unexpected status %d: %s", resp.StatusCode, readProviderErrorBody(resp.Body))
	}

	// A 200 response is enough for unknown compatible servers. When the server
	// exposes the standard model-list shape, also verify the configured model.
	var payload struct {
		Data *[]struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, maxProviderErrorBody)).Decode(&payload); err != nil || payload.Data == nil {
		return nil
	}
	for _, model := range *payload.Data {
		if model.ID == a.modelName {
			return nil
		}
	}
	return fmt.Errorf("openai-compatible health check: configured model %q was not found", a.modelName)
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
	var endpoint url.URL
	if path == "/chat/completions" && a.endpointURL != nil {
		endpoint = *a.endpointURL
	} else {
		if a.baseURL == nil {
			return nil, fmt.Errorf("openai-compatible request: base URL is unavailable for %s", path)
		}
		endpoint = *a.baseURL
		endpoint.Path = joinURLPath(endpoint.Path, path)
	}

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
	if cfg.BaseURL == "" && cfg.EndpointURL == "" {
		return fmt.Errorf("model adapter config %q missing base URL or endpoint URL", cfg.ID)
	}
	if cfg.ModelName == "" {
		return fmt.Errorf("model adapter config %q missing model name", cfg.ID)
	}
	return nil
}

func normalizeOpenAICompatibleAuth(cfg Config) (Config, error) {
	switch cfg.Auth.Type {
	case "":
		return cfg, nil
	case "none":
		if cfg.Auth.Env != "" || cfg.Auth.Header != "" {
			return Config{}, fmt.Errorf("model adapter config %q auth type none must not include env or header", cfg.ID)
		}
		cfg.APIKeyEnv = ""
		return cfg, nil
	case "bearer_env":
		if cfg.Auth.Env == "" {
			return Config{}, fmt.Errorf("model adapter config %q bearer_env auth requires env", cfg.ID)
		}
		if cfg.Auth.Header != "" {
			return Config{}, fmt.Errorf("model adapter config %q bearer_env auth does not support custom header", cfg.ID)
		}
		cfg.APIKeyEnv = cfg.Auth.Env
		return cfg, nil
	default:
		return Config{}, fmt.Errorf("model adapter config %q auth type %q is not supported by the OpenAI-compatible runtime", cfg.ID, cfg.Auth.Type)
	}
}

func validateOpenAICompatibleURL(configID string, label string, providerURL *url.URL) error {
	if providerURL.Scheme != "http" && providerURL.Scheme != "https" {
		return fmt.Errorf("model adapter config %q %s must use http or https", configID, label)
	}
	if providerURL.Hostname() == "" {
		return fmt.Errorf("model adapter config %q %s must include a hostname", configID, label)
	}
	if providerURL.User != nil {
		return fmt.Errorf("model adapter config %q %s must not contain userinfo", configID, label)
	}
	if providerURL.Fragment != "" {
		return fmt.Errorf("model adapter config %q %s must not contain a fragment", configID, label)
	}
	if providerURL.RawQuery != "" || providerURL.ForceQuery {
		return fmt.Errorf("model adapter config %q %s must not contain query parameters", configID, label)
	}
	return nil
}

func parseOpenAICompatibleURL(configID string, label string, rawURL string) (*url.URL, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("model adapter config %q %s is invalid: %w", configID, label, err)
	}
	if err := validateOpenAICompatibleURL(configID, label, parsedURL); err != nil {
		return nil, err
	}
	return parsedURL, nil
}

func readProviderErrorBody(body io.Reader) string {
	raw, _ := io.ReadAll(io.LimitReader(body, maxProviderErrorBody+1))
	truncated := len(raw) > maxProviderErrorBody
	if truncated {
		raw = raw[:maxProviderErrorBody]
	}
	message := string(bytes.TrimSpace(raw))
	if truncated {
		message += " [truncated]"
	}
	return message
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
