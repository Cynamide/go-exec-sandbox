# Interactive Environment Lifecycle Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a common environment lifecycle seam for browser, terminal, API, repo, notebook, spreadsheet, and simulated conversation tasks.

**Architecture:** Create `internal/environment` with reset, observe, step, reward, and cleanup operations. Start with a deterministic fake/test environment and validation model, then wire lifecycle references into task mode validation.

**Tech Stack:** Go, context cancellation, table-driven tests.

## Global Constraints

- Interactive environments must define cleanup policy.
- Secrets cannot appear in auth/session fixture values.
- Environment traces must be usable for replay and process grading.
- Existing code execution remains independent from interactive environments.

---

### Task 1: Add Environment Interface

**Files:**
- Create: `internal/environment/environment.go`
- Create: `internal/environment/environment_test.go`

**Interfaces:**
- Produces: `Environment`, `Config`, `Observation`, `Action`, `StepResult`
- Consumes: no prior new interface

- [ ] **Step 1: Write the failing test**

```go
func TestConfigRequiresCleanupPolicy(t *testing.T) {
	cfg := environment.Config{ID: "browser", Kind: "browser"}
	if err := cfg.Validate(); err == nil {
		t.Fatal("Validate() error = nil, want cleanup policy error")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/environment`
Expected: FAIL because `internal/environment` does not exist.

- [ ] **Step 3: Implement interface and config validation**

Define lifecycle methods `Reset`, `Observe`, `Step`, and `Cleanup`. Validate kind, cleanup policy, action space, observation projection, and render mode.

- [ ] **Step 4: Run tests**

Run: `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/environment`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/environment
git commit -m "Add environment lifecycle interface"
```

### Task 2: Parse Environment Config

**Files:**
- Modify: `internal/manifest/manifest.go`
- Modify: `internal/manifest/manifest_test.go`

**Interfaces:**
- Consumes: manifest `environments`
- Produces: `manifest.Loaded.Environments map[string]environment.Config`

- [ ] **Step 1: Write the failing test**

```go
func TestLoadParsesBrowserEnvironment(t *testing.T) {
	loaded, err := Load(writeManifest(t, manifestWithBrowserEnvironment()))
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Environments["checkout"].Kind != "browser" {
		t.Fatalf("environments = %+v", loaded.Environments)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/manifest`
Expected: FAIL because `Loaded.Environments` is undefined.

- [ ] **Step 3: Implement parsing and validation**

Parse environment kind, setup, reset policy, auth/session fixture refs, observation projection, action space, reward source, termination, truncation, cleanup, render mode, and seed.

- [ ] **Step 4: Run tests**

Run: `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/manifest ./internal/environment`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/manifest internal/environment
git commit -m "Parse environment lifecycle config"
```

### Task 3: Add Deterministic Test Environment

**Files:**
- Create: `internal/environment/fake.go`
- Create: `internal/environment/fake_test.go`

**Interfaces:**
- Produces: `FakeEnvironment`
- Consumes: `Environment`

- [ ] **Step 1: Write the failing test**

```go
func TestFakeEnvironmentRecordsActions(t *testing.T) {
	env := environment.NewFakeEnvironment([]environment.Observation{{Text: "start"}, {Text: "done"}})
	if err := env.Reset(context.Background()); err != nil {
		t.Fatal(err)
	}
	_, err := env.Step(context.Background(), environment.Action{Type: "click", Value: "#pay"})
	if err != nil {
		t.Fatal(err)
	}
	if len(env.Actions()) != 1 {
		t.Fatalf("actions = %+v", env.Actions())
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/environment`
Expected: FAIL because `NewFakeEnvironment` is undefined.

- [ ] **Step 3: Implement fake environment**

Provide deterministic observations, action recording, reward values, termination flag, and cleanup status.

- [ ] **Step 4: Run tests**

Run: `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/environment`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/environment
git commit -m "Add deterministic test environment"
```

### Task 4: Connect Environment Requirements To Task Modes

**Files:**
- Modify: `internal/runner/mode_validation.go`
- Modify: `internal/runner/mode_validation_test.go`
- Modify: `internal/benchmark/model.go`

**Interfaces:**
- Consumes: `environment.Config`
- Produces: task environment reference validation

- [ ] **Step 1: Write the failing test**

```go
func TestInteractiveTaskRequiresEnvironmentReference(t *testing.T) {
	err := runner.ValidateTaskEnvironment(benchmark.Task{TaskMode: "browser_agent"}, nil)
	if err == nil {
		t.Fatal("ValidateTaskEnvironment() error = nil, want missing environment error")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/runner`
Expected: FAIL because `ValidateTaskEnvironment` is undefined.

- [ ] **Step 3: Implement validation**

Require environment references for browser, terminal, API, notebook, spreadsheet, conversation, and reward-based task modes.

- [ ] **Step 4: Run tests**

Run: `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/runner ./internal/environment`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/runner internal/benchmark internal/environment
git commit -m "Validate interactive task environments"
```
