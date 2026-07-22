package manifest

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadSupportedManifestReturnsRuntimeAndCatalogs(t *testing.T) {
	path := writeManifest(t, `
schema_version: 1
runtime_defaults:
  timeout_ms: 45000
providers:
  ollama_local:
    kind: ollama
    base_url: http://localhost:11434
models:
  qwen_local:
    provider: ollama_local
    model_name: qwen3:4b
    enabled: true
tasks:
  release-note-digest:
    id: release-note-digest
    title: Generate Release Notes
    description: Produce a markdown digest.
    family: engineering_workflows
    language: python
    artifact_expectation:
      type: markdown_report
      format: markdown
      description: Markdown digest
    test_cases:
      - input: "api|feature|Audit log\n"
        expected_output: "| team | feature changes |\n| --- | --- |\n| api | 1 |"
scaffolds:
  baseline:
    baseline: true
    description: Run without extra help.
  tool_assisted:
    description: Plan before coding.
    prompt_prefix: "Plan briefly, then answer.\n\n"
    tools:
      - plan
`)

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if loaded.Runtime.DefaultTimeoutMS != 45000 {
		t.Fatalf("DefaultTimeoutMS = %d, want 45000", loaded.Runtime.DefaultTimeoutMS)
	}
	if loaded.Runtime.OLLAMAHost != "http://localhost:11434" {
		t.Fatalf("OLLAMAHost = %q, want manifest host", loaded.Runtime.OLLAMAHost)
	}
	if loaded.Runtime.OLLAMAModel != "qwen3:4b" {
		t.Fatalf("OLLAMAModel = %q, want qwen3:4b", loaded.Runtime.OLLAMAModel)
	}
	if len(loaded.Tasks.Tasks) != 1 {
		t.Fatalf("tasks = %d, want 1", len(loaded.Tasks.Tasks))
	}
	if got := loaded.Tasks.Tasks[0].TaskFamily; got != "engineering_workflows" {
		t.Fatalf("TaskFamily = %q, want engineering_workflows", got)
	}
	if len(loaded.Scaffolds.Scaffolds) != 2 {
		t.Fatalf("scaffolds = %d, want 2", len(loaded.Scaffolds.Scaffolds))
	}
	if !loaded.Scaffolds.Scaffolds[0].Baseline {
		t.Fatal("first scaffold is not baseline")
	}
}

func TestLoadUsesBaseURLEnvOverride(t *testing.T) {
	t.Setenv("OLLAMA_TEST_HOST", "http://ollama.test:11434")
	path := writeManifest(t, `
schema_version: 1
providers:
  ollama_local:
    kind: ollama
    base_url: http://localhost:11434
    base_url_env: OLLAMA_TEST_HOST
models:
  qwen_local:
    provider: ollama_local
    model_name: qwen3:4b
    enabled: true
tasks:
  task:
    id: task
    title: Task
    description: Desc
    family: support_workflows
    language: python
    test_cases:
      - input: ""
        expected_output: ok
scaffolds:
  baseline:
    baseline: true
    description: Baseline
`)

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if loaded.Runtime.OLLAMAHost != "http://ollama.test:11434" {
		t.Fatalf("OLLAMAHost = %q, want env override", loaded.Runtime.OLLAMAHost)
	}
}

func TestLoadReturnsEnabledModelAdapterConfigsAndDefaultRoles(t *testing.T) {
	t.Setenv("OLLAMA_TEST_HOST", "http://ollama.test:11434")

	loaded, err := Load(writeManifest(t, `
schema_version: 1
providers:
  ollama_local:
    kind: ollama
    base_url: http://localhost:11434
    base_url_env: OLLAMA_TEST_HOST
    api_key_env: OLLAMA_API_KEY
  openai_api:
    kind: openai
    base_url: https://api.openai.com/v1
    api_key_env: OPENAI_API_KEY
models:
  qwen_local:
    provider: ollama_local
    model_name: qwen3:4b
    enabled: true
    params:
      temperature: 0
      max_tokens: 2048
    request_mapping:
      messages_field: messages
      temperature_field: temperature
      max_tokens_field: max_tokens
      model_field: model
    response_mapping:
      text_path: choices[0].message.content
      finish_reason_path: choices[0].finish_reason
      usage_prompt_tokens_path: usage.prompt_tokens
      usage_completion_tokens_path: usage.completion_tokens
    capabilities:
      tool_use: false
      file_editing: false
      browser: true
      multimodal: true
      visual_reasoning: true
      terminal_session: true
      spreadsheet: false
      notebook: false
      conversation: true
      structured_output: true
      judge: true
  disabled_remote:
    provider: openai_api
    model_name: gpt-5.4-mini
    enabled: false
default_model_roles:
  judge: qwen_local
  planner: qwen_local
tasks:
  task:
    id: task
    title: Task
    description: Desc
    family: support_workflows
    language: python
    test_cases:
      - input: ""
        expected_output: ok
scaffolds:
  baseline:
    baseline: true
    description: Baseline
`))
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if len(loaded.Models) != 1 {
		t.Fatalf("Models length = %d, want 1", len(loaded.Models))
	}

	model := loaded.Models[0]
	if model.ID != "qwen_local" {
		t.Fatalf("model.ID = %q, want qwen_local", model.ID)
	}
	if model.ProviderID != "ollama_local" {
		t.Fatalf("model.ProviderID = %q, want ollama_local", model.ProviderID)
	}
	if model.ProviderKind != "ollama" {
		t.Fatalf("model.ProviderKind = %q, want ollama", model.ProviderKind)
	}
	if model.ModelName != "qwen3:4b" {
		t.Fatalf("model.ModelName = %q, want qwen3:4b", model.ModelName)
	}
	if model.BaseURL != "http://ollama.test:11434" {
		t.Fatalf("model.BaseURL = %q, want env override", model.BaseURL)
	}
	if model.APIKeyEnv != "OLLAMA_API_KEY" {
		t.Fatalf("model.APIKeyEnv = %q, want OLLAMA_API_KEY", model.APIKeyEnv)
	}
	if got := model.Params["temperature"]; got != 0 {
		t.Fatalf("model.Params[temperature] = %#v, want 0", got)
	}
	if got := model.Params["max_tokens"]; got != 2048 {
		t.Fatalf("model.Params[max_tokens] = %#v, want 2048", got)
	}
	if model.RequestMapping.MessagesField != "messages" {
		t.Fatalf("model.RequestMapping.MessagesField = %q, want messages", model.RequestMapping.MessagesField)
	}
	if model.RequestMapping.ModelField != "model" {
		t.Fatalf("model.RequestMapping.ModelField = %q, want model", model.RequestMapping.ModelField)
	}
	if model.ResponseMapping.TextPath != "choices[0].message.content" {
		t.Fatalf("model.ResponseMapping.TextPath = %q, want choices[0].message.content", model.ResponseMapping.TextPath)
	}
	if model.ResponseMapping.UsageCompletionTokensPath != "usage.completion_tokens" {
		t.Fatalf("model.ResponseMapping.UsageCompletionTokensPath = %q, want usage.completion_tokens", model.ResponseMapping.UsageCompletionTokensPath)
	}
	if !model.Capabilities.Browser || !model.Capabilities.Multimodal || !model.Capabilities.VisualReasoning || !model.Capabilities.TerminalSession || !model.Capabilities.Conversation || !model.Capabilities.StructuredOutput || !model.Capabilities.Judge {
		t.Fatalf("model.Capabilities = %+v, want declared capabilities", model.Capabilities)
	}
	if model.Capabilities.ToolUse || model.Capabilities.FileEditing || model.Capabilities.Spreadsheet || model.Capabilities.Notebook {
		t.Fatalf("model.Capabilities = %+v, want disabled capabilities to remain false", model.Capabilities)
	}
	if got := loaded.DefaultModelRoles["judge"]; got != "qwen_local" {
		t.Fatalf("DefaultModelRoles[judge] = %q, want qwen_local", got)
	}
	if got := loaded.DefaultModelRoles["planner"]; got != "qwen_local" {
		t.Fatalf("DefaultModelRoles[planner] = %q, want qwen_local", got)
	}
}

func TestLoadRejectsEnabledNonOllamaModel(t *testing.T) {
	path := writeManifest(t, `
schema_version: 1
providers:
  openai_api:
    kind: openai
models:
  gpt:
    provider: openai_api
    model_name: gpt-5.4-mini
    enabled: true
tasks:
  task:
    id: task
    title: Task
    description: Desc
    family: support_workflows
    language: python
    test_cases:
      - input: ""
        expected_output: ok
scaffolds:
  baseline:
    baseline: true
    description: Baseline
`)

	if _, err := Load(path); err == nil {
		t.Fatal("Load() error = nil, want unsupported provider error")
	}
}

func TestLoadRejectsDefaultRoleForUnknownModel(t *testing.T) {
	path := writeManifest(t, `
schema_version: 1
providers:
  ollama_local:
    kind: ollama
models:
  qwen_local:
    provider: ollama_local
    model_name: qwen3:4b
    enabled: true
default_model_roles:
  judge: missing_model
tasks:
  task:
    id: task
    title: Task
    description: Desc
    family: support_workflows
    language: python
    test_cases:
      - input: ""
        expected_output: ok
scaffolds:
  baseline:
    baseline: true
    description: Baseline
`)

	if _, err := Load(path); err == nil {
		t.Fatal("Load() error = nil, want unknown default role model error")
	}
}

func TestLoadRejectsJudgeRoleWithoutJudgeCapability(t *testing.T) {
	path := writeManifest(t, `
schema_version: 1
providers:
  ollama_local:
    kind: ollama
models:
  qwen_local:
    provider: ollama_local
    model_name: qwen3:4b
    enabled: true
    capabilities:
      judge: false
default_model_roles:
  judge: qwen_local
tasks:
  task:
    id: task
    title: Task
    description: Desc
    family: support_workflows
    language: python
    test_cases:
      - input: ""
        expected_output: ok
scaffolds:
  baseline:
    baseline: true
    description: Baseline
`)

	if _, err := Load(path); err == nil {
		t.Fatal("Load() error = nil, want judge capability error")
	}
}

func TestLoadRejectsUnsupportedSchemaVersion(t *testing.T) {
	path := writeManifest(t, `
schema_version: 2
providers:
  ollama_local:
    kind: ollama
models:
  qwen_local:
    provider: ollama_local
    model_name: qwen3:4b
    enabled: true
tasks:
  task:
    id: task
    title: Task
    description: Desc
    family: support_workflows
    language: python
    test_cases:
      - input: ""
        expected_output: ok
scaffolds:
  baseline:
    baseline: true
    description: Baseline
`)

	if _, err := Load(path); err == nil {
		t.Fatal("Load() error = nil, want schema version error")
	}
}

func TestLoadRejectsNegativeTimeout(t *testing.T) {
	path := writeManifest(t, `
schema_version: 1
runtime_defaults:
  timeout_ms: -1
providers:
  ollama_local:
    kind: ollama
models:
  qwen_local:
    provider: ollama_local
    model_name: qwen3:4b
    enabled: true
tasks:
  task:
    id: task
    title: Task
    description: Desc
    family: support_workflows
    language: python
    test_cases:
      - input: ""
        expected_output: ok
scaffolds:
  baseline:
    baseline: true
    description: Baseline
`)

	if _, err := Load(path); err == nil {
		t.Fatal("Load() error = nil, want negative timeout error")
	}
}

func TestLoadRejectsTaskKeyIDMismatch(t *testing.T) {
	path := writeManifest(t, `
schema_version: 1
providers:
  ollama_local:
    kind: ollama
models:
  qwen_local:
    provider: ollama_local
    model_name: qwen3:4b
    enabled: true
tasks:
  task-key:
    id: task-id
    title: Task
    description: Desc
    family: support_workflows
    language: python
    test_cases:
      - input: ""
        expected_output: ok
scaffolds:
  baseline:
    baseline: true
    description: Baseline
`)

	if _, err := Load(path); err == nil {
		t.Fatal("Load() error = nil, want task key/id mismatch error")
	}
}

func writeManifest(t *testing.T, contents string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "benchmark.yaml")
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	return path
}
