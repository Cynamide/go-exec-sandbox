# Provider And Model Adapter Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the Ollama-only LLM path with a provider-neutral model adapter layer while preserving the current Ollama manifest behavior.

**Architecture:** Add a small model adapter seam in `internal/modeladapter` and make the manifest resolve enabled model entries into adapter configs. Keep Ollama as the first adapter. Add OpenAI-compatible support after the seam exists, then wire benchmark service construction through the resolved adapter registry.

**Tech Stack:** Go, `gopkg.in/yaml.v3`, Ollama Go API, `net/http`, table-driven Go tests.

## Global Constraints

- Keep the current `benchmark.yaml` valid.
- Do not add live support for a provider kind until it has tests and an adapter.
- Provider secrets must be read from environment variables, never inline manifest values.
- Capability mismatches must fail during manifest loading or service construction.

---

### Task 1: Add Provider-Neutral Adapter Types

**Files:**
- Create: `internal/modeladapter/types.go`
- Create: `internal/modeladapter/types_test.go`

**Interfaces:**
- Produces: `type ModelRequest`, `type ModelResponse`, `type Adapter interface`, `type Config`
- Consumes: no prior new interface

- [ ] **Step 1: Write the failing test**

```go
func TestConfigRequiresModelName(t *testing.T) {
	cfg := modeladapter.Config{ID: "local"}
	if err := cfg.Validate(); err == nil {
		t.Fatal("Validate() error = nil, want missing model name error")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/modeladapter`
Expected: FAIL because `internal/modeladapter` does not exist.

- [ ] **Step 3: Implement the types and validation**

Define `Config` with `ID`, `ProviderID`, `ProviderKind`, `ModelName`, `EndpointURL`, `Params`, and `Capabilities`. Define `Adapter.Generate(ctx context.Context, req ModelRequest) (ModelResponse, error)` and `Adapter.HealthCheck(ctx context.Context) error`.

- [ ] **Step 4: Run test to verify it passes**

Run: `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/modeladapter`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/modeladapter/types.go internal/modeladapter/types_test.go
git commit -m "Add model adapter interfaces"
```

### Task 2: Move Ollama Behind The Adapter Seam

**Files:**
- Create: `internal/modeladapter/ollama.go`
- Create: `internal/modeladapter/ollama_test.go`
- Modify: `internal/llm/llm.go`

**Interfaces:**
- Consumes: `modeladapter.Config`, `modeladapter.Adapter`
- Produces: `modeladapter.NewOllamaAdapter(cfg Config) (Adapter, error)`

- [ ] **Step 1: Write the failing test**

```go
func TestNewOllamaAdapterRejectsInvalidHost(t *testing.T) {
	_, err := modeladapter.NewOllamaAdapter(modeladapter.Config{
		ID: "qwen",
		ProviderKind: "ollama",
		ModelName: "qwen3:4b",
		BaseURL: "://bad",
	})
	if err == nil {
		t.Fatal("NewOllamaAdapter() error = nil, want invalid URL error")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/modeladapter`
Expected: FAIL because `NewOllamaAdapter` is undefined.

- [ ] **Step 3: Implement adapter construction**

Move URL parsing and Ollama client creation from `internal/llm/llm.go` into `internal/modeladapter/ollama.go`. Keep `llm.NewClientWithConfig` as a compatibility wrapper around the Ollama adapter until `benchmark.Service` consumes adapters directly.

- [ ] **Step 4: Run focused tests**

Run: `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/modeladapter ./internal/llm`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/modeladapter internal/llm/llm.go
git commit -m "Move Ollama behind model adapter"
```

### Task 3: Parse Provider And Model Configs From Manifest

**Files:**
- Modify: `internal/manifest/manifest.go`
- Modify: `internal/manifest/manifest_test.go`

**Interfaces:**
- Consumes: `modeladapter.Config`
- Produces: `manifest.Loaded.Models []modeladapter.Config`

- [ ] **Step 1: Write the failing test**

```go
func TestLoadReturnsEnabledModelAdapterConfig(t *testing.T) {
	loaded, err := Load(writeManifest(t, validManifestWithOllamaModel()))
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(loaded.Models) != 1 || loaded.Models[0].ModelName != "qwen3:4b" {
		t.Fatalf("Models = %+v, want qwen3:4b", loaded.Models)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/manifest`
Expected: FAIL because `Loaded.Models` is undefined.

- [ ] **Step 3: Implement model config resolution**

Parse provider kind, base URL, base URL env, API key env, enabled model entries, params, and capabilities. Keep non-Ollama kinds rejected until their adapters are implemented.

- [ ] **Step 4: Run focused tests**

Run: `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/manifest ./internal/modeladapter`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/manifest internal/modeladapter
git commit -m "Resolve model adapter configs from manifest"
```

### Task 4: Add OpenAI-Compatible Adapter

**Files:**
- Create: `internal/modeladapter/openai_compatible.go`
- Create: `internal/modeladapter/openai_compatible_test.go`
- Modify: `internal/manifest/manifest.go`
- Modify: `internal/manifest/manifest_test.go`

**Interfaces:**
- Consumes: `modeladapter.Config`
- Produces: `modeladapter.NewOpenAICompatibleAdapter(cfg Config) (Adapter, error)`

- [ ] **Step 1: Write the failing test**

```go
func TestOpenAICompatibleAdapterBuildsChatCompletionsRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chat/completions" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"choices":[{"message":{"content":"answer"}}]}`)
	}))
	defer server.Close()
	adapter, err := modeladapter.NewOpenAICompatibleAdapter(modeladapter.Config{ID: "local", ProviderKind: "openai_compatible", BaseURL: server.URL + "/v1", ModelName: "local-model"})
	if err != nil {
		t.Fatal(err)
	}
	resp, err := adapter.Generate(context.Background(), modeladapter.ModelRequest{Messages: []modeladapter.Message{{Role: "user", Content: "hi"}}})
	if err != nil || resp.Text != "answer" {
		t.Fatalf("Generate() = %+v, %v", resp, err)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/modeladapter`
Expected: FAIL because the OpenAI-compatible adapter is undefined.

- [ ] **Step 3: Implement adapter**

Use `net/http` and JSON structs for `/chat/completions`. Support bearer auth from `api_key_env`, temperature, max tokens, and model name.

- [ ] **Step 4: Run tests**

Run: `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/modeladapter ./internal/manifest`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/modeladapter internal/manifest
git commit -m "Add OpenAI-compatible model adapter"
```
