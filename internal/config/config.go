package config

import (
	"os"
)

type Config struct {
	DefaultTimeoutMS int
	MaxMemoryMB      int
	Languages        map[string]string
	OLLAMAHost       string
	OLLAMAModel      string
}

func LoadConfig() Config {
	ollamaHost := os.Getenv("OLLAMA_HOST")
	if ollamaHost == "" {
		ollamaHost = "http://localhost:11434"
	}

	ollamaModel := os.Getenv("OLLAMA_MODEL")
	if ollamaModel == "" {
		panic("OLLAMA_MODEL environment variable is required")
	}

	return Config{
		DefaultTimeoutMS: 60000,
		MaxMemoryMB:      256,
		OLLAMAHost:       ollamaHost,
		OLLAMAModel:      ollamaModel,
		Languages: map[string]string{
			"python": "python:3.9-slim",
			"py":     "python:3.9-slim",
			"golang": "golang:1.24-alpine",
			"go":     "golang:1.24-alpine",
		},
	}
}
