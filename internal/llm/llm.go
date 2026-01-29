package llm

import (
	"context"
	"fmt"
	"log"
	"time"

	"gexec-sandbox/internal/config"
	"github.com/ollama/ollama/api"
)

const (
	systemPrompt = "You are an expert programmer. Provide only the function/class implementation. Do not include usage examples."
)

var (
	cfg = config.LoadConfig()
)

type Client struct {
	client *api.Client
}

func NewClient() (*Client, error) {
	client, err := api.ClientFromEnvironment()
	if err != nil {
		return nil, fmt.Errorf("failed to create Ollama client: %w", err)
	}
	return &Client{client: client}, nil
}

func (c *Client) GenerateCode(problem string, language string) (string, error) {
	ctx := context.Background()

	messages := []api.Message{
		{
			Role:    "system",
			Content: systemPrompt,
		},
		{
			Role:    "user",
			Content: fmt.Sprintf("Write a %s solution for:\n%s", language, problem),
		},
	}

	req := &api.ChatRequest{
		Model:    cfg.OLLAMAModel,
		Messages: messages,
		Stream:   new(bool),
	}

	var response string
	err := c.client.Chat(ctx, req, func(resp api.ChatResponse) error {
		response += resp.Message.Content
		return nil
	})

	if err != nil {
		return "", fmt.Errorf("failed to generate code: %w", err)
	}

	return response, nil
}

func WaitForOllama() error {
	client, err := api.ClientFromEnvironment()
	if err != nil {
		return fmt.Errorf("failed to create Ollama client: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timed out waiting for Ollama")
		case <-ticker.C:
			log.Println("Checking Ollama availability...")
			_, err := client.List(ctx)
			if err == nil {
				log.Println("Ollama is available")
				return nil
			}
		}
	}
}

func PullModel() error {
	client, err := api.ClientFromEnvironment()
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
	client, err := api.ClientFromEnvironment()
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
