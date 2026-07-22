# Runtime Controls Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make runtime defaults in `benchmark.yaml` control execution attempts, concurrency, epochs, fail behavior, capture, normalization, sandbox profile, and default task properties.

**Architecture:** Introduce `internal/benchmark/runtime_config.go` as the runtime model used by manifest loading and benchmark service execution. Feed sandbox-specific runtime values into `config.Config` at sandbox call sites. Apply concurrency and epochs in the service layer, and keep deterministic report ordering.

**Tech Stack:** Go, context cancellation, `sync/errgroup` or bounded worker channels, table-driven tests.

## Global Constraints

- Prerequisites: none.
- Omitted runtime controls default to sequential, single-epoch, one-attempt execution.
- Runtime values must be validated before benchmark execution.
- Concurrency must not change report ordering.
- Runtime normalization must happen before outcome grading.

---

### Task 1: Add Benchmark Runtime Config Type

**Files:**
- Create: `internal/benchmark/runtime_config.go`
- Create: `internal/benchmark/runtime_config_test.go`
- Modify: `internal/manifest/manifest.go`
- Modify: `internal/manifest/manifest_test.go`

**Interfaces:**
- Produces: `benchmark.RuntimeConfig`, `benchmark.FailPolicy`, `benchmark.CaptureConfig`, `benchmark.NormalizationConfig`, `benchmark.SampleLimits`
- Consumes: manifest `runtime_defaults`

- [ ] **Step 1: Write the failing test**

```go
func TestRuntimeConfigDefaults(t *testing.T) {
	cfg := benchmark.DefaultRuntimeConfig()
	if cfg.TimeoutMS != 60000 || cfg.Concurrency != 1 || cfg.Epochs != 1 || cfg.MaxAttempts != 1 {
		t.Fatalf("DefaultRuntimeConfig() = %+v", cfg)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/benchmark`
Expected: FAIL because `DefaultRuntimeConfig` is undefined.

- [ ] **Step 3: Implement runtime config and validation**

Add defaults for timeout, max attempts, concurrency, epochs, random seed, working directory, sandbox profile, approval policy, fail policy, sample limits, capture, and normalization.

- [ ] **Step 4: Run tests**

Run: `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/benchmark ./internal/manifest`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/benchmark/runtime_config.go internal/benchmark/runtime_config_test.go internal/manifest
git commit -m "Add benchmark runtime config"
```

### Task 2: Apply Defaults To Tasks

**Files:**
- Modify: `internal/manifest/manifest.go`
- Modify: `internal/manifest/manifest_test.go`
- Modify: `internal/benchmark/model.go`

**Interfaces:**
- Consumes: `benchmark.RuntimeConfig`
- Produces: effective task mode and language on loaded tasks

- [ ] **Step 1: Write the failing test**

```go
func TestLoadAppliesRuntimeLanguageDefault(t *testing.T) {
	path := writeManifest(t, manifestWithRuntimeLanguageAndTaskWithoutLanguage())
	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if loaded.Tasks.Tasks[0].Language != "python" {
		t.Fatalf("Language = %q, want python", loaded.Tasks.Tasks[0].Language)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/manifest`
Expected: FAIL because runtime language is not parsed.

- [ ] **Step 3: Implement defaulting**

Add `TaskMode` to `benchmark.Task`. Parse `runtime_defaults.language` and `runtime_defaults.task_mode`; apply them when task fields are omitted.

- [ ] **Step 4: Run tests**

Run: `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/manifest ./internal/benchmark`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/manifest internal/benchmark
git commit -m "Apply runtime defaults to tasks"
```

### Task 3: Enforce Approval, Sandbox, And Sample Limits

**Files:**
- Modify: `internal/benchmark/runtime_config.go`
- Modify: `internal/benchmark/runtime_config_test.go`
- Modify: `internal/manifest/manifest.go`
- Modify: `internal/manifest/manifest_test.go`
- Modify: `internal/config/config.go`

**Interfaces:**
- Consumes: `benchmark.RuntimeConfig`
- Produces: `RuntimeConfig.RequireApproval(action string) error`, `RuntimeConfig.ValidateSampleLimits(sample SampleBudget) error`

- [ ] **Step 1: Write the failing test**

```go
func TestRuntimeRejectsExternalActionWhenApprovalPolicyDenies(t *testing.T) {
	cfg := benchmark.RuntimeConfig{ApprovalPolicy: "never", SandboxProfile: "docker", SampleLimits: benchmark.SampleLimits{MaxToolCalls: 2}}
	if err := cfg.RequireApproval("external_network"); err == nil {
		t.Fatal("RequireApproval() error = nil, want denied external action")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/benchmark`
Expected: FAIL because runtime approval and sample-limit enforcement are undefined.

- [ ] **Step 3: Implement runtime enforcement helpers**

Validate `working_directory`, `sandbox_profile`, and `approval_policy`. Add sample limits for tool calls, messages, tokens, wall time, and estimated cost. Keep enforcement helpers side-effect free so runners and solvers can call them before external or destructive actions.

- [ ] **Step 4: Run tests**

Run: `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/benchmark ./internal/manifest ./internal/config`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/benchmark internal/manifest internal/config
git commit -m "Enforce runtime safety controls"
```

### Task 4: Execute Epochs And Attempts

**Files:**
- Modify: `internal/benchmark/model.go`
- Modify: `internal/benchmark/service.go`
- Modify: `internal/benchmark/service_test.go`
- Modify: `internal/benchmark/report.go`
- Modify: `internal/benchmark/report_test.go`

**Interfaces:**
- Consumes: `benchmark.RuntimeConfig`
- Produces: run identity fields `Epoch`, `Attempt`, `SampleID`

- [ ] **Step 1: Write the failing test**

```go
func TestServiceRunsConfiguredEpochs(t *testing.T) {
	service := benchmarkServiceWithRuntime(benchmark.RuntimeConfig{Epochs: 2, MaxAttempts: 1, Concurrency: 1})
	report, err := service.Run(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Runs) != 4 {
		t.Fatalf("runs = %d, want baseline and scaffold for two epochs", len(report.Runs))
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/benchmark`
Expected: FAIL because epochs are ignored.

- [ ] **Step 3: Implement epochs and attempts**

Add nested loops over epochs and attempts. Stop attempts for a run after first pass unless fail policy requires all attempts to be recorded.

- [ ] **Step 4: Run tests**

Run: `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/benchmark`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/benchmark
git commit -m "Add runtime epochs and attempts"
```

### Task 5: Add Bounded Concurrency

**Files:**
- Modify: `internal/benchmark/service.go`
- Modify: `internal/benchmark/service_test.go`

**Interfaces:**
- Consumes: `RuntimeConfig.Concurrency`
- Produces: deterministic ordered `[]Run`

- [ ] **Step 1: Write the failing test**

```go
func TestServiceReturnsDeterministicRunOrderWithConcurrency(t *testing.T) {
	service := benchmarkServiceWithRuntime(benchmark.RuntimeConfig{Concurrency: 4, Epochs: 1, MaxAttempts: 1})
	report, err := service.Run(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !sort.SliceIsSorted(report.Runs, func(i, j int) bool { return report.Runs[i].RunID < report.Runs[j].RunID }) {
		t.Fatalf("runs are not ordered: %+v", report.Runs)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/benchmark`
Expected: FAIL because `RunID` and concurrent scheduling are absent.

- [ ] **Step 3: Implement worker pool**

Build a run matrix, execute jobs with a bounded worker channel, store results by matrix index, and return results in matrix order.

- [ ] **Step 4: Run full benchmark tests**

Run: `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/benchmark`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/benchmark
git commit -m "Add deterministic runtime concurrency"
```

### Task 6: Apply Fail Policy Across Run Levels

**Files:**
- Modify: `internal/benchmark/service.go`
- Modify: `internal/benchmark/service_test.go`
- Modify: `internal/benchmark/report.go`

**Interfaces:**
- Consumes: `benchmark.FailPolicy`
- Produces: fail-policy decisions at sample, task, model, and run levels

- [ ] **Step 1: Write the failing test**

```go
func TestFailPolicyStopsTaskAfterFirstSampleFailure(t *testing.T) {
	service := benchmarkServiceWithRuntime(benchmark.RuntimeConfig{FailPolicy: benchmark.FailPolicy{OnSampleFailure: "stop_task"}, MaxAttempts: 1, Epochs: 1})
	report, err := service.Run(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if got := countRunsForTask(report.Runs, "task-with-first-sample-failure"); got != 1 {
		t.Fatalf("runs for failed task = %d, want 1", got)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/benchmark`
Expected: FAIL because fail policy is parsed but not enforced across run levels.

- [ ] **Step 3: Implement fail-policy decisions**

Apply configured actions for sample failure, task failure, model failure, and run failure. Record skipped work in report metadata so aggregate denominators remain explainable.

- [ ] **Step 4: Run tests**

Run: `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/benchmark`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/benchmark
git commit -m "Apply runtime fail policy"
```
