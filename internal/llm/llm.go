package llm

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"gexec-sandbox/internal/config"
	"gexec-sandbox/internal/modeladapter"
	"github.com/ollama/ollama/api"
)

const (
	systemPrompt = "You are an expert programmer. Provide only the function/class implementation. Do not include usage examples."
)

type Client struct {
	adapter modeladapter.Adapter
}

func NewClient() (*Client, error) {
	return NewClientWithConfig(config.LoadConfig())
}

func NewClientWithConfig(cfg config.Config) (*Client, error) {
	adapter, err := modeladapter.NewOllamaAdapter(ollamaAdapterConfig(cfg))
	if err != nil {
		return nil, fmt.Errorf("failed to create Ollama client: %w", err)
	}
	return &Client{adapter: adapter}, nil
}

func (c *Client) chatRequest(problem string, language string) modeladapter.ModelRequest {
	return modeladapter.ModelRequest{
		Messages: []modeladapter.Message{
			{
				Role:    "system",
				Content: systemPrompt,
			},
			{
				Role:    "user",
				Content: fmt.Sprintf("Write a %s solution for:\n%s", language, problem),
			},
		},
	}
}

func (c *Client) GenerateCode(ctx context.Context, problem string, language string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()

	req := c.chatRequest(problem, language)
	resp, err := c.adapter.Generate(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to generate code: %w", err)
	}

	return resp.Text, nil
}

func WaitForOllama(ctx context.Context) error {
	return WaitForOllamaWithConfig(ctx, config.LoadConfig())
}

func WaitForOllamaWithConfig(ctx context.Context, cfg config.Config) error {
	adapter, err := modeladapter.NewOllamaAdapter(ollamaAdapterConfig(cfg))
	if err != nil {
		return fmt.Errorf("failed to create Ollama client: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			log.Println("Checking Ollama availability...")
			if err := adapter.HealthCheck(ctx); err == nil {
				log.Println("Ollama is available")
				return nil
			}
		}
	}
}

func ollamaAdapterConfig(cfg config.Config) modeladapter.Config {
	return modeladapter.Config{
		ID:           cfg.OLLAMAModel,
		ProviderID:   "ollama",
		ProviderKind: "ollama",
		ModelName:    cfg.OLLAMAModel,
		BaseURL:      cfg.OLLAMAHost,
	}
}

func PullModel() error {
	cfg := config.LoadConfig()

	client, err := ollamaAPIClient(cfg)
	if err != nil {
		return fmt.Errorf("failed to create Ollama client: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	log.Printf("Pulling model %s...", cfg.OLLAMAModel)

	req := &api.PullRequest{
		Model:  cfg.OLLAMAModel,
		Stream: new(bool),
	}

	err = client.Pull(ctx, req, func(resp api.ProgressResponse) error {
		log.Printf("Pulling: %s", resp.Status)
		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to pull model: %w", err)
	}

	log.Printf("Model %s pulled successfully", cfg.OLLAMAModel)
	return nil
}

func CheckModelExists() (bool, error) {
	cfg := config.LoadConfig()

	client, err := ollamaAPIClient(cfg)
	if err != nil {
		return false, fmt.Errorf("failed to create Ollama client: %w", err)
	}

	ctx := context.Background()

	resp, err := client.List(ctx)
	if err != nil {
		return false, err
	}

	for _, model := range resp.Models {
		if model.Name == cfg.OLLAMAModel {
			return true, nil
		}
	}

	return false, nil
}

func ollamaAPIClient(cfg config.Config) (*api.Client, error) {
	if cfg.OLLAMAHost == "" {
		return api.ClientFromEnvironment()
	}

	baseURL, err := url.Parse(cfg.OLLAMAHost)
	if err != nil {
		return nil, err
	}

	return api.NewClient(baseURL, http.DefaultClient), nil
}
