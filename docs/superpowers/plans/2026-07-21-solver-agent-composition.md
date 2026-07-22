# Solver And Agent Composition Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a solver seam so scaffold-aware comparisons can vary solver strategy independently from task data, model adapter, and scaffold policy.

**Architecture:** Add `internal/solver` with a small `Solver` interface. `single_shot_code` is the default solver. Later solvers compose planner, tool use, external CLI agents, or multi-agent roles behind the same interface.

**Tech Stack:** Go, context cancellation, dependency injection, table-driven tests.

## Global Constraints

- Prerequisites: provider-model-adapters, prompt-rendering-response-parsing, and scaffold-tools-policy.
- `single_shot_code` is the default solver.
- Solvers must emit trace events and artifacts through common result types.
- External agents require explicit sandbox and approval policy.
- Solver outputs must be compatible with the task primary channel.

---

### Task 1: Add Solver Interface And Result Types

**Files:**
- Create: `internal/solver/solver.go`
- Create: `internal/solver/solver_test.go`

**Interfaces:**
- Produces: `Solver`, `Request`, `Result`, `Artifact`
- Consumes: `benchmark.Task`, `benchmark.Sample`, `benchmark.Scaffold`

- [ ] **Step 1: Write the failing test**

```go
func TestResultRequiresPrimaryArtifact(t *testing.T) {
	result := solver.Result{}
	if err := result.Validate(); err == nil {
		t.Fatal("Validate() error = nil, want missing primary artifact error")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/solver`
Expected: FAIL because `internal/solver` does not exist.

- [ ] **Step 3: Implement solver types**

Define request fields for task, sample, scaffold, model adapter, renderer, parser, tools, environment, and runtime config. Define result fields for primary artifact, trace, model output, and errors.

- [ ] **Step 4: Run tests**

Run: `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/solver`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/solver
git commit -m "Add solver interface"
```

### Task 2: Implement Single-Shot Code Solver

**Files:**
- Create: `internal/solver/single_shot_code.go`
- Create: `internal/solver/single_shot_code_test.go`
- Modify: `internal/benchmark/service.go`

**Interfaces:**
- Consumes: `solver.Solver`
- Produces: `solver.SingleShotCode`

- [ ] **Step 1: Write the failing test**

```go
func TestSingleShotCodeSolverExecutesGeneratedCode(t *testing.T) {
	s := solver.SingleShotCode{Model: fakeModel{Text: "print('ok')"}, Executor: fakeExecutor{Stdout: "ok\n"}}
	result, err := s.Solve(context.Background(), solver.Request{Task: benchmark.Task{Language: "python"}, Sample: benchmark.Sample{Input: ""}})
	if err != nil {
		t.Fatal(err)
	}
	if result.Primary.Text != "ok\n" {
		t.Fatalf("primary text = %q", result.Primary.Text)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/solver`
Expected: FAIL because `SingleShotCode` is undefined.

- [ ] **Step 3: Implement default solver**

Put code-generation plus execution orchestration behind `SingleShotCode`. Have public benchmark helpers call `SingleShotCode` so callers get the same result shape.

- [ ] **Step 4: Run tests**

Run: `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/solver ./internal/benchmark`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/solver internal/benchmark
git commit -m "Add single-shot code solver"
```

### Task 3: Parse Solver Config

**Files:**
- Modify: `internal/manifest/manifest.go`
- Modify: `internal/manifest/manifest_test.go`
- Create: `internal/solver/config.go`

**Interfaces:**
- Produces: `solver.Config`, `solver.Registry`
- Consumes: manifest solver entries

- [ ] **Step 1: Write the failing test**

```go
func TestLoadRejectsUnknownTaskSolver(t *testing.T) {
	path := writeManifest(t, manifestWithTaskSolver("missing_solver"))
	if _, err := Load(path); err == nil {
		t.Fatal("Load() error = nil, want unknown solver error")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/manifest`
Expected: FAIL because solver config is not parsed.

- [ ] **Step 3: Implement solver registry parsing**

Support built-in `single_shot_code`, `planner_then_act`, `tool_using_agent`, `external_cli_agent`, and `multi_agent` config shapes. Only enable solver kinds that have implementations.

- [ ] **Step 4: Run tests**

Run: `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/manifest ./internal/solver`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/manifest internal/solver
git commit -m "Parse solver configuration"
```

### Task 4: Add External CLI Agent Stub With Safety Validation

**Files:**
- Create: `internal/solver/external_cli.go`
- Create: `internal/solver/external_cli_test.go`

**Interfaces:**
- Consumes: `solver.Config`
- Produces: `ExternalCLI`

- [ ] **Step 1: Write the failing test**

```go
func TestExternalCLIRequiresApprovalPolicy(t *testing.T) {
	_, err := solver.NewExternalCLI(solver.Config{ID: "codex", Kind: "external_cli_agent", Command: "codex"})
	if err == nil {
		t.Fatal("NewExternalCLI() error = nil, want missing approval policy error")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/solver`
Expected: FAIL because `NewExternalCLI` is undefined.

- [ ] **Step 3: Implement validation-only stub**

Validate command, sandbox profile, working directory, approval policy, and output parser. Return an explicit unsupported execution error until safe process execution is planned.

- [ ] **Step 4: Run tests**

Run: `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/solver`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/solver
git commit -m "Validate external CLI solver config"
```
