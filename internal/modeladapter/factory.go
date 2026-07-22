package modeladapter

import "fmt"

func New(cfg Config) (Adapter, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	if cfg.ModelLookup != "" && cfg.ModelLookup != "direct" {
		return nil, fmt.Errorf("model adapter config %q has unsupported model lookup %q", cfg.ID, cfg.ModelLookup)
	}
	if cfg.Transport.Configured() {
		return nil, fmt.Errorf("model adapter config %q provider transport is not supported by the current runtime", cfg.ID)
	}

	switch cfg.ProviderKind {
	case "ollama":
		if cfg.EndpointURL != "" {
			return nil, fmt.Errorf("model adapter config %q endpoint URL is not supported for Ollama", cfg.ID)
		}
		if cfg.APIKeyEnv != "" || (cfg.Auth.Type != "" && cfg.Auth.Type != "none") {
			return nil, fmt.Errorf("model adapter config %q auth is not supported for Ollama", cfg.ID)
		}
		return NewOllamaAdapter(cfg)
	case "openai_compatible":
		normalized, err := normalizeOpenAICompatibleAuth(cfg)
		if err != nil {
			return nil, err
		}
		return NewOpenAICompatibleAdapter(normalized)
	default:
		return nil, fmt.Errorf("model adapter config %q has unsupported provider kind %q", cfg.ID, cfg.ProviderKind)
	}
}
