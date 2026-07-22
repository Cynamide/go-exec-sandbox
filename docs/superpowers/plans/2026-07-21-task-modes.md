# Task Modes Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add explicit task-mode dispatch so code execution, artifact-only, repo patch, browser/API/visual/text, choice scoring, and log-likelihood tasks have separate runner contracts.

**Architecture:** Add `internal/runner` with a `Runner` interface and a registry keyed by task mode. Start by moving current code execution behind `code_exec`, then add validation-only stubs for modes that depend on other component plans.

**Tech Stack:** Go, interface-based dispatch, table-driven tests.

## Global Constraints

- `code_exec` remains compatible with the current benchmark.
- Unknown task modes fail during manifest load or runner registry construction.
- Modes requiring capabilities must validate those capabilities before execution.
- Runner outputs must produce a normalized artifact for grading.

---

### Task 1: Add Runner Interface And Registry

**Files:**
- Create: `internal/runner/runner.go`
- Create: `internal/runner/runner_test.go`

**Interfaces:**
- Produces: `Runner`, `Request`, `Result`, `Registry`
- Consumes: `benchmark.TaskMode`

- [ ] **Step 1: Write the failing test**

```go
func TestRegistryRejectsUnknownTaskMode(t *testing.T) {
	registry := runner.Registry{}
	_, err := registry.ForMode("unknown")
	if err == nil {
		t.Fatal("ForMode() error = nil, want unknown mode error")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/runner`
Expected: FAIL because `internal/runner` does not exist.

- [ ] **Step 3: Implement registry**

Define runner request with task, sample, scaffold, solver, environment, scorer, and runtime config. Add registration and lookup by mode string.

- [ ] **Step 4: Run tests**

Run: `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/runner`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/runner
git commit -m "Add task runner registry"
```

### Task 2: Move Current Code Path Behind `code_exec`

**Files:**
- Create: `internal/runner/code_exec.go`
- Create: `internal/runner/code_exec_test.go`
- Modify: `internal/benchmark/service.go`

**Interfaces:**
- Consumes: `runner.Runner`
- Produces: `runner.CodeExec`

- [ ] **Step 1: Write the failing test**

```go
func TestCodeExecRejectsLanguageNone(t *testing.T) {
	r := runner.CodeExec{}
	_, err := r.Run(context.Background(), runner.Request{Task: benchmark.Task{TaskMode: "code_exec", Language: "none"}})
	if err == nil {
		t.Fatal("Run() error = nil, want language error")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/runner`
Expected: FAIL because `CodeExec` is undefined.

- [ ] **Step 3: Implement code exec runner**

Route the existing generation/execution/grading path through `runner.CodeExec`. Keep benchmark service output unchanged.

- [ ] **Step 4: Run tests**

Run: `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/runner ./internal/benchmark`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/runner internal/benchmark
git commit -m "Add code execution runner"
```

### Task 3: Add Direct Output Runners

**Files:**
- Create: `internal/runner/direct_output.go`
- Create: `internal/runner/direct_output_test.go`

**Interfaces:**
- Produces: `ArtifactOnly`, `TextOnly`
- Consumes: model adapter generate capability

- [ ] **Step 1: Write the failing test**

```go
func TestTextOnlyReturnsModelTextAsPrimaryArtifact(t *testing.T) {
	r := runner.TextOnly{Model: fakeModel{Text: "final answer"}}
	result, err := r.Run(context.Background(), runner.Request{Task: benchmark.Task{TaskMode: "text_only"}})
	if err != nil {
		t.Fatal(err)
	}
	if result.Primary.Text != "final answer" {
		t.Fatalf("text = %q", result.Primary.Text)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/runner`
Expected: FAIL because `TextOnly` is undefined.

- [ ] **Step 3: Implement direct runners**

Implement `text_only` and `artifact_only` runners. They call model generation and parse output without executing code.

- [ ] **Step 4: Run tests**

Run: `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/runner`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/runner
git commit -m "Add direct output task runners"
```

### Task 4: Add Mode Validation Stubs

**Files:**
- Create: `internal/runner/mode_validation.go`
- Create: `internal/runner/mode_validation_test.go`
- Modify: `internal/manifest/manifest.go`

**Interfaces:**
- Produces: `ValidateTaskMode(task Task, model ModelConfig) error`
- Consumes: model capabilities and task input metadata

- [ ] **Step 1: Write the failing test**

```go
func TestValidateBrowserAgentRequiresBrowserStartState(t *testing.T) {
	err := runner.ValidateTaskMode(benchmark.Task{TaskMode: "browser_agent"}, modeladapter.Config{Capabilities: map[string]bool{"browser": true}})
	if err == nil {
		t.Fatal("ValidateTaskMode() error = nil, want browser start state error")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/runner`
Expected: FAIL because `ValidateTaskMode` is undefined.

- [ ] **Step 3: Implement validation**

Validate `repo_patch`, `browser_agent`, `api_agent`, `visual_reasoning`, `choice_scoring`, `loglikelihood`, terminal, notebook, spreadsheet, document, and conversation modes. Return unsupported execution errors for runners that are validation-only.

- [ ] **Step 4: Run tests**

Run: `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/runner ./internal/manifest`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/runner internal/manifest
git commit -m "Validate task mode requirements"
```
