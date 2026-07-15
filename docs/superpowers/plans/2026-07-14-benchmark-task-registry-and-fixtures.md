# Benchmark Task Registry and Fixtures Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Turn the benchmark inputs into a reusable registry of tasks and scaffold profiles so the harness can evaluate different task families without hard-coding fixture data.

**Architecture:** Keep task content and scaffold content in versioned JSON fixtures that are loaded through a small catalog layer. The runtime and report layers should consume catalog objects, not raw files, so new benchmark families can be added without changing orchestration code.

**Tech Stack:** Go 1.24+, JSON fixtures under `data/`, existing `internal/benchmark` package, `go test`.

## Global Constraints

- support for tasks that produce artifacts, not just short answers
- economically meaningful task evaluation
- scaffold comparison across the same task
- automatic execution where possible
- deterministic or clearly bounded grading where possible
- toy puzzle benchmarks are out of scope
- leaderboard-first design that ignores workflow context is out of scope

---

### Task 1: Add a catalog loader for tasks and scaffolds

**Files:**
- Create: `internal/benchmark/catalog.go`
- Create: `internal/benchmark/catalog_test.go`
- Create: `data/tasks.json`
- Create: `data/scaffolds.json`

**Interfaces:**
- Consumes: task and scaffold JSON files
- Produces: `TaskCatalog`, `ScaffoldCatalog`, and loader functions that later runtime code can call

- [ ] **Step 1: Write the failing test**

```go
func TestLoadTaskCatalogReturnsFamilies(t *testing.T) {
	catalog, err := LoadTaskCatalog("../../data/tasks.json")
	if err != nil {
		t.Fatalf("LoadTaskCatalog() error = %v", err)
	}

	if len(catalog.Tasks) == 0 {
		t.Fatalf("catalog.Tasks is empty")
	}

	if catalog.Tasks[0].TaskFamily == "" {
		t.Fatalf("first task missing TaskFamily")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/benchmark -run TestLoadTaskCatalogReturnsFamilies -v`

Expected: compile failure because `LoadTaskCatalog` does not exist yet and the new fixture files are absent.

- [ ] **Step 3: Write minimal implementation**

```go
type TaskCatalog struct {
	Tasks []Task
}

type ScaffoldCatalog struct {
	Scaffolds []Scaffold
}

func LoadTaskCatalog(path string) (TaskCatalog, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return TaskCatalog{}, err
	}

	var catalog TaskCatalog
	if err := json.Unmarshal(raw, &catalog.Tasks); err != nil {
		return TaskCatalog{}, err
	}

	return catalog, nil
}

func LoadScaffoldCatalog(path string) (ScaffoldCatalog, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return ScaffoldCatalog{}, err
	}

	var catalog ScaffoldCatalog
	if err := json.Unmarshal(raw, &catalog.Scaffolds); err != nil {
		return ScaffoldCatalog{}, err
	}

	return catalog, nil
}
```

Create `data/tasks.json` with at least one task in each of these families:

- `software_engineering`
- `browser_workflows`
- `spreadsheets`
- `terminal_workflows`

Create `data/scaffolds.json` with at least:

- `baseline`
- `tool-assisted`
- `retrieval-assisted`

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/benchmark -run TestLoadTaskCatalogReturnsFamilies -v`

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/benchmark/catalog.go internal/benchmark/catalog_test.go data/tasks.json data/scaffolds.json
git commit -m "Add benchmark catalogs"
```

### Task 2: Add filtering helpers for benchmark families and scaffolds

**Files:**
- Modify: `internal/benchmark/catalog.go`
- Modify: `internal/benchmark/catalog_test.go`

**Interfaces:**
- Consumes: `TaskCatalog` and `ScaffoldCatalog`
- Produces: filter helpers that let the benchmark surface select subsets of work by family or scaffold mode

- [ ] **Step 1: Write the failing test**

```go
func TestTaskCatalogFiltersByFamily(t *testing.T) {
	catalog := TaskCatalog{
		Tasks: []Task{
			{ID: "a", TaskFamily: "software_engineering"},
			{ID: "b", TaskFamily: "spreadsheets"},
		},
	}

	filtered := catalog.FilterByFamily("spreadsheets")
	if len(filtered.Tasks) != 1 || filtered.Tasks[0].ID != "b" {
		t.Fatalf("FilterByFamily() = %+v, want only task b", filtered.Tasks)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/benchmark -run TestTaskCatalogFiltersByFamily -v`

Expected: compile failure because `FilterByFamily` does not exist yet.

- [ ] **Step 3: Write minimal implementation**

```go
func (c TaskCatalog) FilterByFamily(family string) TaskCatalog {
	filtered := TaskCatalog{}
	for _, task := range c.Tasks {
		if task.TaskFamily == family {
			filtered.Tasks = append(filtered.Tasks, task)
		}
	}
	return filtered
}

func (c ScaffoldCatalog) FilterByName(name string) ScaffoldCatalog {
	filtered := ScaffoldCatalog{}
	for _, scaffold := range c.Scaffolds {
		if scaffold.Name == name {
			filtered.Scaffolds = append(filtered.Scaffolds, scaffold)
		}
	}
	return filtered
}
```

Add the analogous scaffold filter helper for `ScaffoldCatalog`.

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/benchmark -run TestTaskCatalogFiltersByFamily -v`

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/benchmark/catalog.go internal/benchmark/catalog_test.go
git commit -m "Add benchmark catalog filters"
```

### Task 3: Expand the fixture set to cover economically meaningful work

**Files:**
- Modify: `data/tasks.json`
- Modify: `internal/benchmark/catalog_test.go`

**Interfaces:**
- Consumes: the catalog loader
- Produces: fixtures that better reflect the product thesis than puzzle-style tasks

- [ ] **Step 1: Write the failing test**

```go
func TestTaskCatalogContainsMultipleTaskFamilies(t *testing.T) {
	catalog, err := LoadTaskCatalog("../../data/tasks.json")
	if err != nil {
		t.Fatalf("LoadTaskCatalog() error = %v", err)
	}

	families := map[string]bool{}
	for _, task := range catalog.Tasks {
		families[task.TaskFamily] = true
	}

	if len(families) < 4 {
		t.Fatalf("families = %v, want at least 4", families)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/benchmark -run TestTaskCatalogContainsMultipleTaskFamilies -v`

Expected: compile failure because the fixtures do not yet contain the required variety.

- [ ] **Step 3: Write minimal implementation**

Update `data/tasks.json` so each example task includes:

- a descriptive title
- a task family
- a realistic description
- test cases or artifact expectations

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/benchmark -run TestTaskCatalogContainsMultipleTaskFamilies -v`

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add data/tasks.json internal/benchmark/catalog_test.go
git commit -m "Expand benchmark fixture families"
```

## Verification Plan

Run:

```bash
go test ./internal/benchmark -run TestLoadTaskCatalogReturnsFamilies -v
go test ./internal/benchmark -run TestTaskCatalogFiltersByFamily -v
go test ./internal/benchmark -run TestTaskCatalogContainsMultipleTaskFamilies -v
```

Expected: all tests pass.

## Coverage Check

- task registry -> TaskCatalog
- scaffold registry -> ScaffoldCatalog
- family filtering -> FilterByFamily
- fixture breadth -> multiple economically meaningful task families
