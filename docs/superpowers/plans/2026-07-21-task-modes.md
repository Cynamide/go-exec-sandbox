# Task Modes Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add explicit task-mode dispatch so code execution, artifact-only, repo patch, browser/API/visual/text, choice scoring, and log-likelihood tasks have separate runner contracts.

**Architecture:** Add `internal/runner` with a `Runner` interface and a registry keyed by task mode. Route code execution through `code_exec`, then add runner contracts for repo, browser/API, visual, choice, and log-likelihood modes. Modes whose concrete environments live in other component plans return typed unsupported-execution errors after validating their inputs.

**Tech Stack:** Go, interface-based dispatch, table-driven tests.

## Global Constraints

- Prerequisites: runtime-controls, fixtures-inputs, provider-model-adapters, prompt-rendering-response-parsing, solver-agent-composition, and scoring-interfaces.
- `code_exec` remains compatible with the repository benchmark fixtures.
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

Define runner request with `benchmark.Task`, `benchmark.Sample`, `benchmark.Scaffold`, `benchmark.RuntimeConfig`, `modeladapter.Adapter`, and parsed model output. Add registration and lookup by mode string; environment and scorer fields are added by the plans that introduce those concrete interfaces.

- [ ] **Step 4: Run tests**

Run: `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/runner`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/runner
git commit -m "Add task runner registry"
```

### Task 2: Route Code Execution Through `code_exec`

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

Route generation, execution, and grading through `runner.CodeExec`. Keep benchmark service output unchanged.

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

### Task 4: Add Repo Patch Runner Contract

**Files:**
- Create: `internal/runner/repo_patch.go`
- Create: `internal/runner/repo_patch_test.go`

**Interfaces:**
- Produces: `RepoPatch`, `RepoPatchWorkspace`, `RepoPatchResult`
- Consumes: task repo seed input metadata, solver parsed patch output, grading test-command checks

- [ ] **Step 1: Write the failing test**

```go
func TestRepoPatchRequiresRepoSeed(t *testing.T) {
	r := runner.RepoPatch{}
	_, err := r.Run(context.Background(), runner.Request{Task: benchmark.Task{TaskMode: "repo_patch"}})
	if err == nil {
		t.Fatal("Run() error = nil, want missing repo seed error")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/runner`
Expected: FAIL because `RepoPatch` is undefined.

- [ ] **Step 3: Implement repo patch runner contract**

Validate repo seed, working directory, patch parser, and side-effect output declarations. Return a `RunArtifact` with git diff and test-command artifact slots; return unsupported execution until isolated repo workspace execution is connected.

- [ ] **Step 4: Run tests**

Run: `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/runner`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/runner
git commit -m "Add repo patch runner contract"
```

### Task 5: Add Browser And API Agent Runner Contracts

**Files:**
- Create: `internal/runner/interactive.go`
- Create: `internal/runner/interactive_test.go`

**Interfaces:**
- Produces: `BrowserAgent`, `APIAgent`
- Consumes: environment lifecycle interface, browser start state, API environment config

- [ ] **Step 1: Write the failing test**

```go
func TestBrowserAgentRequiresEnvironment(t *testing.T) {
	r := runner.BrowserAgent{}
	_, err := r.Run(context.Background(), runner.Request{Task: benchmark.Task{TaskMode: "browser_agent"}})
	if err == nil {
		t.Fatal("Run() error = nil, want missing browser environment error")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/runner`
Expected: FAIL because browser/API runner contracts are undefined.

- [ ] **Step 3: Implement interactive runner contracts**

Validate browser start state, API environment config, action parser, observation capture, screenshot/HTTP response artifact declarations, and environment capability flags. Return unsupported execution when no concrete environment implementation is registered.

- [ ] **Step 4: Run tests**

Run: `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/runner`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/runner
git commit -m "Add browser and API runner contracts"
```

### Task 6: Add Visual Reasoning Runner Contract

**Files:**
- Create: `internal/runner/visual_reasoning.go`
- Create: `internal/runner/visual_reasoning_test.go`

**Interfaces:**
- Produces: `VisualReasoning`
- Consumes: attachment model payload hooks and model modality capabilities

- [ ] **Step 1: Write the failing test**

```go
func TestVisualReasoningRequiresImageAttachment(t *testing.T) {
	r := runner.VisualReasoning{}
	_, err := r.Run(context.Background(), runner.Request{Task: benchmark.Task{TaskMode: "visual_reasoning"}})
	if err == nil {
		t.Fatal("Run() error = nil, want missing visual attachment error")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/runner`
Expected: FAIL because `VisualReasoning` is undefined.

- [ ] **Step 3: Implement visual runner contract**

Validate at least one image or screenshot attachment, model visual capability, native content-part conversion, and primary output channel. Call the model adapter when native visual payload support is available.

- [ ] **Step 4: Run tests**

Run: `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/runner ./internal/attachments`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/runner internal/attachments
git commit -m "Add visual reasoning runner contract"
```

### Task 7: Add Choice And Log-Likelihood Runner Contracts

**Files:**
- Create: `internal/runner/scored_choices.go`
- Create: `internal/runner/scored_choices_test.go`

**Interfaces:**
- Produces: `ChoiceScoring`, `LogLikelihood`
- Consumes: scoring interfaces and model-side scoring adapter capabilities

- [ ] **Step 1: Write the failing test**

```go
func TestChoiceScoringRequiresChoices(t *testing.T) {
	r := runner.ChoiceScoring{}
	_, err := r.Run(context.Background(), runner.Request{Task: benchmark.Task{TaskMode: "choice_scoring"}})
	if err == nil {
		t.Fatal("Run() error = nil, want missing choices error")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/runner`
Expected: FAIL because choice and log-likelihood runners are undefined.

- [ ] **Step 3: Implement scored runner contracts**

Validate candidate choices, target strings, logprob support, rolling-window settings, and normalized score output. Call model-side scoring interfaces when available; otherwise return a capability error before execution.

- [ ] **Step 4: Run tests**

Run: `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/runner ./internal/scoring`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/runner internal/scoring
git commit -m "Add choice and log-likelihood runners"
```

### Task 8: Add Mode Validation Stubs

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
