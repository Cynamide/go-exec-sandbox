package config

type Config struct {
	DefaultTimeoutMS int
	MaxMemoryMB      int
	Languages        map[string]string
}

func LoadConfig() Config {
	return Config{
		DefaultTimeoutMS: 5000,
		MaxMemoryMB:      128,
		Languages: map[string]string{
			"python": "python:3.9-slim",
			"py":     "python:3.9-slim",
			"golang": "golang:1.21-alpine",
			"go":     "golang:1.21-alpine",
		},
	}
}
