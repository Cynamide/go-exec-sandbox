# Fixtures And Inputs Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a fixture registry and input materialization layer so task data can live outside `benchmark.yaml` while inline test cases continue to work.

**Architecture:** Add `internal/fixtures` for loading fixture files and resolving sample inputs. Manifest loading parses fixture refs and task input sources, then produces benchmark samples consumed by the service layer.

**Tech Stack:** Go, JSON/YAML decoding, path validation, table-driven tests.

## Global Constraints

- Prerequisites: none.
- Inline `test_cases` remain supported.
- Fixture paths must resolve under the repository unless explicitly allowed by config.
- Missing required environment variables fail before task execution.
- Fixture resolution must produce stable sample IDs.

---

### Task 1: Add Fixture Registry Types

**Files:**
- Create: `internal/fixtures/registry.go`
- Create: `internal/fixtures/registry_test.go`

**Interfaces:**
- Produces: `fixtures.Registry`, `fixtures.Definition`, `fixtures.Case`
- Consumes: no prior new interface

- [ ] **Step 1: Write the failing test**

```go
func TestRegistryRejectsMissingFixturePath(t *testing.T) {
	registry := fixtures.Registry{Definitions: map[string]fixtures.Definition{
		"cases": {Kind: "task_cases", Path: "missing.json", Format: "json"},
	}}
	if err := registry.Validate("/repo"); err == nil {
		t.Fatal("Validate() error = nil, want missing path error")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/fixtures`
Expected: FAIL because `internal/fixtures` does not exist.

- [ ] **Step 3: Implement registry validation**

Add the registry contract with these exact field names so manifest parsing can consume it later:

```go
type Registry struct {
	Definitions map[string]Definition
}

type Definition struct {
	ID          string
	Kind        string
	Path        string
	Format      string
	Split       string
	Environment map[string]string
}

type Case struct {
	ID             string
	Input          string
	ExpectedOutput string
	Metadata       map[string]string
}
```

Validate supported kinds: `task_cases`, `rubric`, `preference_rubric`, `repo_seed`, `browser_fixture`, `multimodal_fixture`, `attachment`, `expected_files`, and `golden_patch`.

- [ ] **Step 4: Run tests**

Run: `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/fixtures`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/fixtures
git commit -m "Add fixture registry"
```

### Task 2: Load Task Case Fixtures

**Files:**
- Modify: `internal/fixtures/registry.go`
- Create: `internal/fixtures/task_cases_test.go`

**Interfaces:**
- Consumes: `fixtures.Registry`
- Produces: `func LoadTaskCases(root string, def Definition) ([]Case, error)`

- [ ] **Step 1: Write the failing test**

```go
func TestLoadTaskCasesReadsJSONCases(t *testing.T) {
	path := writeFile(t, `{"cases":[{"id":"case-1","input":"a","expected_output":"b"}]}`)
	cases, err := fixtures.LoadTaskCases(t.TempDir(), fixtures.Definition{Kind: "task_cases", Path: path, Format: "json"})
	if err != nil {
		t.Fatal(err)
	}
	if cases[0].ID != "case-1" || cases[0].Input != "a" {
		t.Fatalf("cases = %+v", cases)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/fixtures`
Expected: FAIL because `LoadTaskCases` is undefined.

- [ ] **Step 3: Implement JSON fixture loading**

Decode `{"cases":[...]}` with unknown-field rejection, required `id`, `input`, and optional `expected_output` plus metadata.

- [ ] **Step 4: Run tests**

Run: `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/fixtures`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/fixtures
git commit -m "Load task case fixtures"
```

### Task 3: Parse Fixture And Input References From Manifest

**Files:**
- Modify: `internal/manifest/manifest.go`
- Modify: `internal/manifest/manifest_test.go`
- Modify: `internal/benchmark/model.go`

**Interfaces:**
- Consumes: `fixtures.Registry`
- Produces: task input source metadata on `benchmark.Task`

- [ ] **Step 1: Write the failing test**

```go
func TestLoadRejectsUnknownInputFixture(t *testing.T) {
	path := writeManifest(t, manifestWithInputFixture("missing_fixture"))
	if _, err := Load(path); err == nil {
		t.Fatal("Load() error = nil, want unknown fixture error")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/manifest`
Expected: FAIL because fixture refs are not parsed.

- [ ] **Step 3: Implement manifest parsing**

Parse top-level `fixtures`, `tasks[*].inputs.sources`, `repo_seed`, `attachments`, `browser_start_state`, and environment requirements. Keep inline test cases valid.

- [ ] **Step 4: Run focused tests**

Run: `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/manifest ./internal/fixtures`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/manifest internal/benchmark internal/fixtures
git commit -m "Parse fixture input references"
```

### Task 4: Materialize Samples For Benchmark Runs

**Files:**
- Create: `internal/benchmark/samples.go`
- Create: `internal/benchmark/samples_test.go`
- Modify: `internal/benchmark/service.go`

**Interfaces:**
- Consumes: `benchmark.Task`
- Produces: `benchmark.Sample{ID, Input, ExpectedOutput, FixtureID, Split}`

- [ ] **Step 1: Write the failing test**

```go
func TestSamplesFromInlineTestCasesUseStableIDs(t *testing.T) {
	task := benchmark.Task{ID: "task", TestCases: []benchmark.TestCase{{Input: "a", ExpectedOutput: "b"}}}
	samples, err := benchmark.SamplesForTask(task)
	if err != nil {
		t.Fatal(err)
	}
	if samples[0].ID != "task/case-0001" {
		t.Fatalf("sample ID = %q", samples[0].ID)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/benchmark`
Expected: FAIL because `SamplesForTask` is undefined.

- [ ] **Step 3: Implement sample materialization**

Convert inline cases and resolved fixture cases into `Sample` values. Update service execution to iterate samples rather than raw test cases.

- [ ] **Step 4: Run tests**

Run: `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/benchmark ./internal/manifest ./internal/fixtures`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/benchmark internal/manifest internal/fixtures
git commit -m "Materialize benchmark task samples"
```
