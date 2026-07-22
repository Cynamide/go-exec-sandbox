# Reporting And Trace Capture Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Expand reports and traces to include benchmark metadata, run identity, model/sample/fixture/task-mode data, artifacts, per-check grades, and process evidence while preserving current lift summaries.

**Architecture:** Extend `benchmark.Run` and `BenchmarkReport` with additive fields. Keep existing JSON fields stable. Add trace/artifact structures and deterministic report ordering.

**Tech Stack:** Go, JSON structs, deterministic sorting, table-driven tests.

## Global Constraints

- Existing JSON report fields remain available.
- Report arrays must be deterministic.
- Disabled capture fields must be omitted or redacted.
- Trace data must be sufficient for process grading.

---

### Task 1: Add Structured Run Identity

**Files:**
- Modify: `internal/benchmark/model.go`
- Modify: `internal/benchmark/report.go`
- Modify: `internal/benchmark/report_test.go`

**Interfaces:**
- Produces: `RunID`, `ModelID`, `SampleID`, `FixtureID`, `Split`, `TaskMode`, `Epoch`, `Attempt`
- Consumes: run matrix entries

- [ ] **Step 1: Write the failing test**

```go
func TestReportKeepsRunIdentity(t *testing.T) {
	report := benchmark.BuildBenchmarkReport([]benchmark.Task{{ID: "task"}}, []benchmark.Run{{RunID: "run-1", TaskID: "task", ModelID: "model", SampleID: "sample"}})
	if report.Runs[0].RunID != "run-1" || report.Runs[0].ModelID != "model" {
		t.Fatalf("run = %+v", report.Runs[0])
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/benchmark`
Expected: FAIL because run identity fields are absent.

- [ ] **Step 3: Implement additive fields**

Add JSON fields with `omitempty` where possible and populate them from service execution.

- [ ] **Step 4: Run tests**

Run: `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/benchmark`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/benchmark
git commit -m "Add structured run identity"
```

### Task 2: Add Trace And Artifact Models

**Files:**
- Create: `internal/benchmark/artifact.go`
- Create: `internal/benchmark/artifact_test.go`
- Modify: `internal/benchmark/model.go`

**Interfaces:**
- Produces: `TraceEvent`, `CapturedArtifact`, `ArtifactChannel`
- Consumes: runner and grading outputs

- [ ] **Step 1: Write the failing test**

```go
func TestCapturedArtifactRequiresChannel(t *testing.T) {
	artifact := benchmark.CapturedArtifact{Text: "output"}
	if err := artifact.Validate(); err == nil {
		t.Fatal("Validate() error = nil, want missing channel error")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/benchmark`
Expected: FAIL because captured artifacts are undefined.

- [ ] **Step 3: Implement models**

Support stdout, stderr, exit code, git diff, generated files, browser state, screenshot, notebook state, spreadsheet state, attachments, model outputs, and trace events.

- [ ] **Step 4: Run tests**

Run: `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/benchmark`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/benchmark
git commit -m "Add report artifact model"
```

### Task 3: Add Grading Results To Reports

**Files:**
- Modify: `internal/benchmark/model.go`
- Modify: `internal/benchmark/report.go`
- Modify: `internal/benchmark/report_test.go`

**Interfaces:**
- Consumes: `grading.GradingResult`
- Produces: per-run check and metric results in `BenchmarkReport`

- [ ] **Step 1: Write the failing test**

```go
func TestReportIncludesPerCheckResults(t *testing.T) {
	run := benchmark.Run{TaskID: "task", Grading: benchmark.GradingResult{Checks: []benchmark.CheckResult{{Name: "exact", Passed: true}}}}
	report := benchmark.BuildBenchmarkReport([]benchmark.Task{{ID: "task"}}, []benchmark.Run{run})
	if len(report.Runs[0].Grading.Checks) != 1 {
		t.Fatalf("grading = %+v", report.Runs[0].Grading)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/benchmark`
Expected: FAIL because grading report fields are absent.

- [ ] **Step 3: Implement grading report fields**

Add per-check results, metric results, preferred score, aggregate score, and judge metadata.

- [ ] **Step 4: Run tests**

Run: `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/benchmark`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/benchmark
git commit -m "Report grading check results"
```

### Task 4: Add Capture Redaction And Ordering

**Files:**
- Create: `internal/benchmark/report_sanitize.go`
- Create: `internal/benchmark/report_sanitize_test.go`
- Modify: `internal/benchmark/report.go`

**Interfaces:**
- Consumes: runtime capture config
- Produces: sanitized deterministic report

- [ ] **Step 1: Write the failing test**

```go
func TestSanitizeReportOmitsDisabledModelOutputs(t *testing.T) {
	report := benchmark.BenchmarkReport{Runs: []benchmark.Run{{ModelOutput: "secret"}}}
	sanitized := benchmark.SanitizeReport(report, benchmark.CaptureConfig{ModelOutputs: false})
	if sanitized.Runs[0].ModelOutput != "" {
		t.Fatalf("model output leaked: %+v", sanitized.Runs[0])
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/benchmark`
Expected: FAIL because `SanitizeReport` is undefined.

- [ ] **Step 3: Implement sanitization**

Omit or redact fields disabled by capture config and sort runs, scaffolds, family summaries, metrics, and artifacts deterministically.

- [ ] **Step 4: Run tests**

Run: `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/benchmark`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/benchmark
git commit -m "Sanitize captured report fields"
```
