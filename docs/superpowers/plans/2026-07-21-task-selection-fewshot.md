# Task Selection And Few-Shot Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Resolve tasks, models, scaffolds, groups, tags, splits, and few-shot settings into a deterministic run matrix.

**Architecture:** Add a selection module that takes loaded manifest data and returns run matrix entries. Keep selection separate from service execution so reporting can explain skipped or included entries.

**Tech Stack:** Go, deterministic sorting, table-driven tests.

## Global Constraints

- Current manifest resolves to every task, one enabled model, baseline scaffold, and every scaffold variant.
- A selection resolving to zero models, scaffolds, or tasks must fail before execution.
- Few-shot sampling must use `runtime_defaults.random_seed`.
- Lift reporting requires a baseline scaffold in the selected matrix.

---

### Task 1: Add Selection Types

**Files:**
- Create: `internal/benchmark/selection.go`
- Create: `internal/benchmark/selection_test.go`
- Modify: `internal/benchmark/model.go`

**Interfaces:**
- Produces: `Selection`, `ModelFilter`, `ScaffoldFilter`, `RunMatrixEntry`
- Consumes: `Task`, `Scaffold`

- [ ] **Step 1: Write the failing test**

```go
func TestResolveSelectionRejectsZeroScaffolds(t *testing.T) {
	_, err := benchmark.ResolveRunMatrix([]benchmark.Task{{ID: "task"}}, nil, nil)
	if err == nil {
		t.Fatal("ResolveRunMatrix() error = nil, want zero scaffold error")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/benchmark`
Expected: FAIL because `ResolveRunMatrix` is undefined.

- [ ] **Step 3: Implement selection types**

Add filters for include/exclude lists and deterministic sorting by task ID, model ID, scaffold name, sample ID, epoch, and attempt.

- [ ] **Step 4: Run tests**

Run: `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/benchmark`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/benchmark/selection.go internal/benchmark/selection_test.go internal/benchmark/model.go
git commit -m "Add benchmark selection model"
```

### Task 2: Parse Task Selection Metadata

**Files:**
- Modify: `internal/manifest/manifest.go`
- Modify: `internal/manifest/manifest_test.go`
- Modify: `internal/benchmark/model.go`

**Interfaces:**
- Consumes: manifest task YAML
- Produces: `Task.Version`, `Task.Split`, `Task.Tags`, `Task.ModelSelection`, `Task.ScaffoldSelection`

- [ ] **Step 1: Write the failing test**

```go
func TestLoadParsesTaskModelIncludes(t *testing.T) {
	loaded, err := Load(writeManifest(t, manifestWithTaskModelInclude("qwen3_4b_local")))
	if err != nil {
		t.Fatal(err)
	}
	if got := loaded.Tasks.Tasks[0].ModelSelection.Include[0]; got != "qwen3_4b_local" {
		t.Fatalf("include = %q", got)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/manifest`
Expected: FAIL because selection fields are undefined.

- [ ] **Step 3: Implement parsing and validation**

Parse version, split, tags, models include/exclude, and scaffolds include/exclude. Validate referenced model and scaffold IDs.

- [ ] **Step 4: Run tests**

Run: `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/manifest ./internal/benchmark`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/manifest internal/benchmark
git commit -m "Parse task selection metadata"
```

### Task 3: Add Group Selection

**Files:**
- Create: `internal/benchmark/groups.go`
- Create: `internal/benchmark/groups_test.go`
- Modify: `internal/manifest/manifest.go`

**Interfaces:**
- Produces: `Group`, `GroupSelector`, `ResolveGroupTasks(group Group, tasks []Task) ([]Task, error)`
- Consumes: task metadata

- [ ] **Step 1: Write the failing test**

```go
func TestResolveGroupTasksByTag(t *testing.T) {
	tasks := []benchmark.Task{{ID: "a", Tags: []string{"browser"}}, {ID: "b", Tags: []string{"finance"}}}
	got, err := benchmark.ResolveGroupTasks(benchmark.Group{ID: "browser", Tags: []string{"browser"}}, tasks)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].ID != "a" {
		t.Fatalf("tasks = %+v", got)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/benchmark`
Expected: FAIL because `Group` is undefined.

- [ ] **Step 3: Implement groups**

Resolve by explicit task IDs, tags, family, split, and task mode. Reject empty group results and aggregate metrics absent from selected tasks.

- [ ] **Step 4: Run tests**

Run: `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/benchmark ./internal/manifest`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/benchmark internal/manifest
git commit -m "Add benchmark group selection"
```

### Task 4: Add Few-Shot Sampling

**Files:**
- Create: `internal/benchmark/fewshot.go`
- Create: `internal/benchmark/fewshot_test.go`
- Modify: `internal/manifest/manifest.go`

**Interfaces:**
- Produces: `FewShotConfig`, `SelectFewShotExamples(task Task, samples []Sample, seed int64) ([]Sample, error)`
- Consumes: sample materialization from fixtures plan

- [ ] **Step 1: Write the failing test**

```go
func TestFewShotFirstNSampler(t *testing.T) {
	samples := []benchmark.Sample{{ID: "1"}, {ID: "2"}, {ID: "3"}}
	got, err := benchmark.SelectFewShotExamples(benchmark.FewShotConfig{NumExamples: 2, Sampler: "first_n"}, samples, 17)
	if err != nil {
		t.Fatal(err)
	}
	if got[0].ID != "1" || got[1].ID != "2" {
		t.Fatalf("examples = %+v", got)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/benchmark`
Expected: FAIL because `SelectFewShotExamples` is undefined.

- [ ] **Step 3: Implement samplers**

Support `none`, `first_n`, `fixed_list`, and deterministic `random`. Reject `num_examples > 0` with `sampler: none`.

- [ ] **Step 4: Run tests**

Run: `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/benchmark ./internal/manifest`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/benchmark internal/manifest
git commit -m "Add deterministic few-shot sampling"
```
