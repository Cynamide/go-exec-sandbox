# Run Persistence And Interchange Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add durable run storage, cache keys, resume, retry-failed, prediction import, result export, replay, and rescore workflows.

**Architecture:** Add `internal/runstore` for filesystem-backed run directories. Benchmark service writes sample artifacts through the store. CLI commands expose import, export, resume, and rescore operations.

**Tech Stack:** Go, JSON/JSONL, filesystem writes, deterministic cache keys.

## Global Constraints

- Run state must include a manifest snapshot.
- Resume must reject incompatible manifest snapshots.
- Prediction import must validate task and model IDs.
- Public export must respect dataset visibility policy.

---

### Task 1: Add Run Store

**Files:**
- Create: `internal/runstore/store.go`
- Create: `internal/runstore/store_test.go`

**Interfaces:**
- Produces: `Store`, `RunManifest`, `SampleRecord`
- Consumes: benchmark reports and run artifacts

- [ ] **Step 1: Write the failing test**

```go
func TestStoreWritesManifestSnapshot(t *testing.T) {
	store := runstore.New(t.TempDir())
	if err := store.CreateRun("run-1", []byte("schema_version: 1\n")); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(store.Root(), "run-1", "manifest.yaml")); err != nil {
		t.Fatal(err)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/runstore`
Expected: FAIL because `internal/runstore` does not exist.

- [ ] **Step 3: Implement filesystem store**

Create run directories with manifest snapshot, resolved matrix, model outputs, traces, artifacts, scores, and final report paths.

- [ ] **Step 4: Run tests**

Run: `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/runstore`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/runstore
git commit -m "Add benchmark run store"
```

### Task 2: Add Cache Keys And Resume

**Files:**
- Modify: `internal/runstore/store.go`
- Create: `internal/runstore/cache_test.go`

**Interfaces:**
- Produces: `CacheKey`, `ResumePlan`
- Consumes: manifest version, model ID, task ID, scaffold ID, sample ID, solver ID, runtime settings

- [ ] **Step 1: Write the failing test**

```go
func TestCacheKeyChangesWhenModelChanges(t *testing.T) {
	a := runstore.CacheKey{ManifestVersion: "v1", ModelID: "a", TaskID: "task"}.String()
	b := runstore.CacheKey{ManifestVersion: "v1", ModelID: "b", TaskID: "task"}.String()
	if a == b {
		t.Fatal("cache key did not change with model")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/runstore`
Expected: FAIL because `CacheKey` is undefined.

- [ ] **Step 3: Implement cache and resume planning**

Hash stable cache key fields. Build resume plan that skips completed valid samples and marks incomplete samples for execution.

- [ ] **Step 4: Run tests**

Run: `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/runstore`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/runstore
git commit -m "Add run cache and resume planning"
```

### Task 3: Add Prediction Import And Result Export

**Files:**
- Create: `internal/runstore/interchange.go`
- Create: `internal/runstore/interchange_test.go`
- Modify: `cmd/evaluator/main.go`
- Modify: `cmd/evaluator/benchmark_cli_test.go`

**Interfaces:**
- Produces: `ImportPredictions(path string) ([]Prediction, error)`, `ExportResults(report BenchmarkReport, format string) ([]byte, error)`
- Consumes: benchmark reports

- [ ] **Step 1: Write the failing test**

```go
func TestImportPredictionsRejectsMissingTaskID(t *testing.T) {
	path := writeJSONL(t, `{"model_id":"m","output":"x"}`+"\n")
	if _, err := runstore.ImportPredictions(path); err == nil {
		t.Fatal("ImportPredictions() error = nil, want missing task ID error")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/runstore`
Expected: FAIL because import/export functions are undefined.

- [ ] **Step 3: Implement interchange**

Support JSONL prediction import with task ID, model ID, scaffold ID, sample ID, and output. Support JSON and JSONL result export.

- [ ] **Step 4: Run tests**

Run: `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/runstore ./cmd/evaluator`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/runstore cmd/evaluator
git commit -m "Add prediction import and result export"
```

### Task 4: Add Rescore Command

**Files:**
- Create: `internal/runstore/rescore.go`
- Create: `internal/runstore/rescore_test.go`
- Modify: `cmd/evaluator/main.go`
- Modify: `cmd/evaluator/benchmark_cli_test.go`

**Interfaces:**
- Produces: `RescoreRun(ctx context.Context, store Store, runID string, grader grading.Engine) (BenchmarkReport, error)`
- Consumes: persisted artifacts and grading config

- [ ] **Step 1: Write the failing test**

```go
func TestRescoreRejectsMissingRawArtifacts(t *testing.T) {
	_, err := runstore.RescoreRun(context.Background(), runstore.New(t.TempDir()), "run-1", grading.Engine{})
	if err == nil {
		t.Fatal("RescoreRun() error = nil, want missing artifacts error")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/runstore`
Expected: FAIL because `RescoreRun` is undefined.

- [ ] **Step 3: Implement rescore**

Load persisted artifacts, traces, and manifest snapshot. Recompute grading results without model inference when required artifacts exist.

- [ ] **Step 4: Run tests**

Run: `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/runstore ./cmd/evaluator`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/runstore cmd/evaluator
git commit -m "Add run rescore workflow"
```
