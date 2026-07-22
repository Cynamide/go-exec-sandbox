package manifest

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"

	"gexec-sandbox/internal/benchmark"
	"gexec-sandbox/internal/config"
	"gexec-sandbox/internal/modeladapter"
	"gopkg.in/yaml.v3"
)

var ErrInvalidManifest = errors.New("invalid benchmark manifest")

type Loaded struct {
	Runtime           config.Config
	Models            []modeladapter.Config
	DefaultModelRoles map[string]string
	Tasks             benchmark.TaskCatalog
	Scaffolds         benchmark.ScaffoldCatalog
}

type file struct {
	SchemaVersion     int                 `yaml:"schema_version"`
	RuntimeDefaults   runtimeDefaults     `yaml:"runtime_defaults"`
	Providers         map[string]provider `yaml:"providers"`
	Models            map[string]model    `yaml:"models"`
	DefaultModelRoles map[string]string   `yaml:"default_model_roles"`
	Tasks             map[string]task     `yaml:"tasks"`
	Scaffolds         map[string]scaffold `yaml:"scaffolds"`
}

type runtimeDefaults struct {
	TimeoutMS int `yaml:"timeout_ms"`
}

type provider struct {
	Kind        string                       `yaml:"kind"`
	BaseURL     string                       `yaml:"base_url"`
	BaseURLEnv  string                       `yaml:"base_url_env"`
	APIKeyEnv   string                       `yaml:"api_key_env"`
	ModelLookup string                       `yaml:"model_lookup"`
	Transport   modeladapter.TransportConfig `yaml:"transport"`
}

type model struct {
	Provider        string                       `yaml:"provider"`
	ModelName       string                       `yaml:"model_name"`
	EndpointURL     string                       `yaml:"endpoint_url"`
	Enabled         bool                         `yaml:"enabled"`
	Auth            *modeladapter.AuthConfig     `yaml:"auth"`
	Params          map[string]any               `yaml:"params"`
	RequestMapping  modeladapter.RequestMapping  `yaml:"request_mapping"`
	ResponseMapping modeladapter.ResponseMapping `yaml:"response_mapping"`
	Capabilities    modeladapter.Capabilities    `yaml:"capabilities"`
}

type task struct {
	ID                  string                         `yaml:"id"`
	Title               string                         `yaml:"title"`
	Description         string                         `yaml:"description"`
	Family              string                         `yaml:"family"`
	Language            string                         `yaml:"language"`
	ArtifactExpectation *benchmark.ArtifactExpectation `yaml:"artifact_expectation"`
	TestCases           []benchmark.TestCase           `yaml:"test_cases"`
}

type scaffold struct {
	Baseline     bool     `yaml:"baseline"`
	Description  string   `yaml:"description"`
	PromptPrefix string   `yaml:"prompt_prefix"`
	Tools        []string `yaml:"tools"`
}

func Load(path string) (Loaded, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return Loaded{}, err
	}

	var manifest file
	decoder := yaml.NewDecoder(bytes.NewReader(raw))
	decoder.KnownFields(true)
	if err := decoder.Decode(&manifest); err != nil {
		return Loaded{}, err
	}
	if err := requireYAMLEOF(decoder); err != nil {
		return Loaded{}, err
	}

	if err := manifest.validateSchema(); err != nil {
		return Loaded{}, err
	}

	models, err := manifest.modelConfigs()
	if err != nil {
		return Loaded{}, err
	}

	runtime, err := manifest.runtimeConfig()
	if err != nil {
		return Loaded{}, err
	}
	defaultRoles, err := manifest.defaultModelRoles(models)
	if err != nil {
		return Loaded{}, err
	}

	tasks, err := manifest.taskCatalog()
	if err != nil {
		return Loaded{}, err
	}

	scaffolds, err := manifest.scaffoldCatalog()
	if err != nil {
		return Loaded{}, err
	}
	if err := validateScaffoldCapabilities(models, scaffolds); err != nil {
		return Loaded{}, err
	}

	return Loaded{
		Runtime:           runtime,
		Models:            models,
		DefaultModelRoles: defaultRoles,
		Tasks:             tasks,
		Scaffolds:         scaffolds,
	}, nil
}

func (m file) validateSchema() error {
	if m.SchemaVersion != 1 {
		return fmt.Errorf("%w: schema_version must be 1", ErrInvalidManifest)
	}
	return nil
}

func (m file) runtimeConfig() (config.Config, error) {
	ollamaModel, ollamaHost, err := m.ollamaSelection()
	if err != nil {
		return config.Config{}, err
	}

	timeoutMS := m.RuntimeDefaults.TimeoutMS
	if timeoutMS == 0 {
		timeoutMS = 60000
	}
	if timeoutMS < 0 {
		return config.Config{}, fmt.Errorf("%w: runtime_defaults.timeout_ms cannot be negative", ErrInvalidManifest)
	}

	return config.Config{
		DefaultTimeoutMS: timeoutMS,
		MaxMemoryMB:      256,
		OLLAMAHost:       ollamaHost,
		OLLAMAModel:      ollamaModel,
		Languages:        defaultLanguages(),
	}, nil
}

func (m file) modelConfigs() ([]modeladapter.Config, error) {
	names := sortedKeys(m.Models)
	models := make([]modeladapter.Config, 0, len(names))
	for _, name := range names {
		candidate := m.Models[name]
		if !candidate.Enabled {
			continue
		}
		if candidate.ModelName == "" {
			return nil, fmt.Errorf("%w: model %q missing model_name", ErrInvalidManifest, name)
		}

		providerConfig, ok := m.Providers[candidate.Provider]
		if !ok {
			return nil, fmt.Errorf("%w: model %q references unknown provider %q", ErrInvalidManifest, name, candidate.Provider)
		}
		if candidate.EndpointURL != "" && providerConfig.Transport.Configured() {
			return nil, fmt.Errorf("%w: model %q endpoint_url conflicts with provider %q transport", ErrInvalidManifest, name, candidate.Provider)
		}
		if providerConfig.Transport.Configured() {
			return nil, fmt.Errorf("%w: provider %q transport is not supported by the current runtime", ErrInvalidManifest, candidate.Provider)
		}
		if providerConfig.ModelLookup != "" && providerConfig.ModelLookup != "direct" {
			return nil, fmt.Errorf("%w: provider %q model_lookup %q is not supported by the current runtime", ErrInvalidManifest, candidate.Provider, providerConfig.ModelLookup)
		}

		auth, apiKeyEnv, err := resolveAuth(name, providerConfig.APIKeyEnv, candidate.Auth)
		if err != nil {
			return nil, err
		}
		cfg := modeladapter.Config{
			ID:              name,
			ProviderID:      candidate.Provider,
			ProviderKind:    providerConfig.Kind,
			ModelName:       candidate.ModelName,
			BaseURL:         resolveBaseURL(providerConfig),
			EndpointURL:     candidate.EndpointURL,
			APIKeyEnv:       apiKeyEnv,
			ModelLookup:     providerConfig.ModelLookup,
			Auth:            auth,
			Transport:       providerConfig.Transport,
			Params:          copyAnyMap(candidate.Params),
			RequestMapping:  candidate.RequestMapping,
			ResponseMapping: candidate.ResponseMapping,
			Capabilities:    candidate.Capabilities,
		}
		if err := cfg.Validate(); err != nil {
			return nil, err
		}
		if providerConfig.Kind != "ollama" && providerConfig.Kind != "openai_compatible" {
			return nil, fmt.Errorf("%w: provider kind %q is not supported by the current runtime", ErrInvalidManifest, providerConfig.Kind)
		}
		models = append(models, cfg)
	}
	if len(models) == 0 {
		return nil, fmt.Errorf("%w: at least one enabled model is required", ErrInvalidManifest)
	}
	return models, nil
}

func resolveAuth(modelID string, providerAPIKeyEnv string, auth *modeladapter.AuthConfig) (modeladapter.AuthConfig, string, error) {
	if auth == nil {
		return modeladapter.AuthConfig{}, providerAPIKeyEnv, nil
	}

	switch auth.Type {
	case "none":
		return *auth, "", nil
	case "bearer_env":
		if auth.Env == "" {
			return modeladapter.AuthConfig{}, "", fmt.Errorf("%w: model %q auth type bearer_env requires env", ErrInvalidManifest, modelID)
		}
		return *auth, auth.Env, nil
	default:
		return modeladapter.AuthConfig{}, "", fmt.Errorf("%w: model %q auth type %q is not supported by the current runtime", ErrInvalidManifest, modelID, auth.Type)
	}
}

func (m file) defaultModelRoles(models []modeladapter.Config) (map[string]string, error) {
	if len(m.DefaultModelRoles) == 0 {
		return nil, nil
	}

	modelByID := make(map[string]modeladapter.Config, len(models))
	for _, model := range models {
		modelByID[model.ID] = model
	}

	resolved := make(map[string]string, len(m.DefaultModelRoles))
	for role, modelID := range m.DefaultModelRoles {
		model, ok := modelByID[modelID]
		if !ok {
			return nil, fmt.Errorf("%w: default_model_roles.%s references unknown enabled model %q", ErrInvalidManifest, role, modelID)
		}
		if role == "judge" && !model.Capabilities.Judge {
			return nil, fmt.Errorf("%w: default_model_roles.judge requires model %q to declare capabilities.judge", ErrInvalidManifest, modelID)
		}
		resolved[role] = modelID
	}

	return resolved, nil
}

func (m file) ollamaSelection() (string, string, error) {
	enabledModelID := ""
	enabledModelName := ""
	enabledProvider := provider{}
	for _, name := range sortedKeys(m.Models) {
		candidate := m.Models[name]
		if !candidate.Enabled {
			continue
		}
		if candidate.ModelName == "" {
			return "", "", fmt.Errorf("%w: model %q missing model_name", ErrInvalidManifest, name)
		}
		providerConfig, ok := m.Providers[candidate.Provider]
		if !ok {
			return "", "", fmt.Errorf("%w: model %q references unknown provider %q", ErrInvalidManifest, name, candidate.Provider)
		}
		if providerConfig.Kind != "ollama" && providerConfig.Kind != "openai_compatible" {
			return "", "", fmt.Errorf("%w: provider kind %q is not supported by the current runtime", ErrInvalidManifest, providerConfig.Kind)
		}
		if providerConfig.Kind != "ollama" {
			continue
		}
		if enabledModelID == "" {
			enabledModelID = name
			enabledModelName = candidate.ModelName
			enabledProvider = providerConfig
		}
	}

	if enabledModelID == "" {
		return "", "", nil
	}

	ollamaHost := resolveBaseURL(enabledProvider)
	if ollamaHost == "" {
		ollamaHost = "http://localhost:11434"
	}

	return enabledModelName, ollamaHost, nil
}

func resolveBaseURL(p provider) string {
	baseURL := p.BaseURL
	if p.BaseURLEnv != "" {
		if value := os.Getenv(p.BaseURLEnv); value != "" {
			baseURL = value
		}
	}
	return baseURL
}

func (m file) taskCatalog() (benchmark.TaskCatalog, error) {
	names := sortedKeys(m.Tasks)
	catalog := benchmark.TaskCatalog{Tasks: make([]benchmark.Task, 0, len(names))}
	for _, name := range names {
		task := m.Tasks[name]
		if task.ID != name {
			return benchmark.TaskCatalog{}, fmt.Errorf("%w: task key %q must match id %q", ErrInvalidManifest, name, task.ID)
		}
		catalog.Tasks = append(catalog.Tasks, benchmark.Task{
			ID:                  task.ID,
			Title:               task.Title,
			Description:         task.Description,
			TaskFamily:          task.Family,
			Language:            task.Language,
			ArtifactExpectation: task.ArtifactExpectation,
			TestCases:           task.TestCases,
		})
	}
	if err := benchmark.ValidateTaskCatalog(catalog); err != nil {
		return benchmark.TaskCatalog{}, err
	}
	return catalog, nil
}

func (m file) scaffoldCatalog() (benchmark.ScaffoldCatalog, error) {
	names := sortedKeys(m.Scaffolds)
	catalog := benchmark.ScaffoldCatalog{Scaffolds: make([]benchmark.Scaffold, 0, len(names))}
	for _, name := range names {
		scaffold := m.Scaffolds[name]
		catalog.Scaffolds = append(catalog.Scaffolds, benchmark.Scaffold{
			Baseline:     scaffold.Baseline,
			Name:         name,
			Description:  scaffold.Description,
			PromptPrefix: scaffold.PromptPrefix,
			Tools:        scaffold.Tools,
		})
	}
	if err := benchmark.ValidateScaffoldCatalog(catalog); err != nil {
		return benchmark.ScaffoldCatalog{}, err
	}
	return catalog, nil
}

func validateScaffoldCapabilities(models []modeladapter.Config, scaffolds benchmark.ScaffoldCatalog) error {
	for _, scaffold := range scaffolds.Scaffolds {
		for _, tool := range scaffold.Tools {
			capability, supported := scaffoldToolCapability(tool)
			if !supported {
				continue
			}
			for _, model := range models {
				if !hasCapability(model.Capabilities, capability) {
					return fmt.Errorf("%w: scaffold %q tool %q requires enabled model %q to declare capabilities.%s", ErrInvalidManifest, scaffold.Name, tool, model.ID, capability)
				}
			}
		}
	}
	return nil
}

// Tasks do not yet declare browser or multimodal requirements. Until that
// schema exists, only capability-bearing scaffold tool names are enforceable.
// This validates declarations only; tool execution belongs to a future runtime.
func scaffoldToolCapability(tool string) (string, bool) {
	switch tool {
	case "browser":
		return "browser", true
	case "file_editing":
		return "file_editing", true
	case "multimodal":
		return "multimodal", true
	case "terminal_session":
		return "terminal_session", true
	case "spreadsheet":
		return "spreadsheet", true
	case "notebook":
		return "notebook", true
	default:
		return "", false
	}
}

func hasCapability(capabilities modeladapter.Capabilities, capability string) bool {
	switch capability {
	case "browser":
		return capabilities.Browser
	case "file_editing":
		return capabilities.FileEditing
	case "multimodal":
		return capabilities.Multimodal
	case "terminal_session":
		return capabilities.TerminalSession
	case "spreadsheet":
		return capabilities.Spreadsheet
	case "notebook":
		return capabilities.Notebook
	default:
		return false
	}
}

func requireYAMLEOF(decoder *yaml.Decoder) error {
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return ErrInvalidManifest
	}
	return nil
}

func defaultLanguages() map[string]string {
	return map[string]string{
		"python": "python:3.9-slim",
		"py":     "python:3.9-slim",
		"golang": "golang:1.24-alpine",
		"go":     "golang:1.24-alpine",
	}
}

func sortedKeys[V any](items map[string]V) []string {
	keys := make([]string, 0, len(items))
	for key := range items {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func copyAnyMap(source map[string]any) map[string]any {
	if len(source) == 0 {
		return nil
	}
	cloned := make(map[string]any, len(source))
	for key, value := range source {
		cloned[key] = value
	}
	return cloned
}
