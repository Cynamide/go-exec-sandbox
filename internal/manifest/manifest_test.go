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
