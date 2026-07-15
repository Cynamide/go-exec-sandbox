# Benchmark Surface Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Expose the scaffold-aware benchmark through a thin service layer so it can be run from the HTTP server and from a CLI entrypoint without duplicating benchmark logic.

**Architecture:** Put benchmark invocation behind a shared service that accepts a task catalog and scaffold catalog, then returns a report. The HTTP and CLI surfaces should only translate input/output and should not know how to score or execute tasks themselves.

**Tech Stack:** Go 1.24+, `net/http`, `encoding/json`, current `cmd/evaluator/main.go`, existing `internal/benchmark` package.

## Global Constraints

- benchmark reporting that separates baseline performance from scaffolded performance
- scaffold comparison across the same task
- automatic execution where possible
- deterministic or clearly bounded grading where possible
- toy puzzle benchmarks are out of scope
- leaderboard-first design that ignores workflow context is out of scope

---

### Task 1: Add a shared benchmark service

**Files:**
- Create: `internal/benchmark/service.go`
- Create: `internal/benchmark/service_test.go`

**Interfaces:**
- Consumes: `TaskCatalog`, `ScaffoldCatalog`, `Executor`, `Grader`, `LLMClient`
- Produces: `BenchmarkService` with a `Run` method that returns `BenchmarkReport`

- [ ] **Step 1: Write the failing test**

```go
func TestBenchmarkServiceRunReturnsReport(t *testing.T) {
	svc := BenchmarkService{
		Tasks: TaskCatalog{Tasks: []Task{{ID: "a", Description: "demo", TaskFamily: "software_engineering", Language: "python"}}},
		Scaffolds: ScaffoldCatalog{Scaffolds: []Scaffold{{Name: "baseline"}, {Name: "tool-assisted"}}},
		Runner: fakeBenchmarkRunner{},
	}

	report, err := svc.Run(context.Background())
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if report.TotalTasks != 1 {
		t.Fatalf("TotalTasks = %d, want 1", report.TotalTasks)
	}
}

type fakeBenchmarkRunner struct{}

func (fakeBenchmarkRunner) RunTask(ctx context.Context, task Task, scaffold Scaffold, mode RunMode) Run {
	return Run{
		TaskID:   task.ID,
		Mode:     mode,
		Scaffold: scaffold,
		Passed:   scaffold.Name == "tool-assisted",
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/benchmark -run TestBenchmarkServiceRunReturnsReport -v`

Expected: compile failure because `BenchmarkService` does not exist yet.

- [ ] **Step 3: Write minimal implementation**

```go
type BenchmarkRunner interface {
	RunTask(ctx context.Context, task Task, scaffold Scaffold, mode RunMode) Run
}

type BenchmarkServiceAPI interface {
	Run(ctx context.Context) (BenchmarkReport, error)
}

type BenchmarkService struct {
	Tasks    TaskCatalog
	Scaffolds ScaffoldCatalog
	Runner   BenchmarkRunner
}

func (s BenchmarkService) Run(ctx context.Context) (BenchmarkReport, error) {
	if len(s.Scaffolds.Scaffolds) < 2 {
		return BenchmarkReport{}, fmt.Errorf("need at least two scaffolds")
	}

	runs := make([]Run, 0, len(s.Tasks.Tasks)*2)
	for _, task := range s.Tasks.Tasks {
		runs = append(runs,
			s.Runner.RunTask(ctx, task, s.Scaffolds.Scaffolds[0], RunModeBaseline),
			s.Runner.RunTask(ctx, task, s.Scaffolds.Scaffolds[1], RunModeScaffolded),
		)
	}

	return BuildBenchmarkReport(s.Tasks.Tasks, runs), nil
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/benchmark -run TestBenchmarkServiceRunReturnsReport -v`

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/benchmark/service.go internal/benchmark/service_test.go
git commit -m "Add benchmark service layer"
```

### Task 2: Expose the benchmark service over HTTP

**Files:**
- Create: `internal/httpapi/benchmark_handler.go`
- Modify: `cmd/evaluator/main.go`
- Create: `internal/httpapi/benchmark_handler_test.go`

**Interfaces:**
- Consumes: `BenchmarkService`
- Produces: `POST /benchmark/run` JSON handler that returns `BenchmarkReport`

- [ ] **Step 1: Write the failing test**

```go
func TestBenchmarkRunHandlerReturnsJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/benchmark/run", strings.NewReader(`{}`))
	rr := httptest.NewRecorder()

	handler := BenchmarkRunHandler{
		Service: fakeBenchmarkService{report: BenchmarkReport{TotalTasks: 1}},
	}
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "\"total_tasks\":1") {
		t.Fatalf("body = %s, want benchmark report JSON", rr.Body.String())
	}
}

type fakeBenchmarkService struct {
	report BenchmarkReport
}

func (f fakeBenchmarkService) Run(ctx context.Context) (BenchmarkReport, error) {
	return f.report, nil
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/httpapi -run TestBenchmarkRunHandlerReturnsJSON -v`

Expected: compile failure because `BenchmarkRunHandler` does not exist yet.

- [ ] **Step 3: Write minimal implementation**

```go
type BenchmarkRunHandler struct {
	Service BenchmarkServiceAPI
}

func (h BenchmarkRunHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	report, err := h.Service.Run(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(report)
}
```

Register the handler in `cmd/evaluator/main.go` on `POST /benchmark/run`.

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/httpapi -run TestBenchmarkRunHandlerReturnsJSON -v`

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/httpapi/benchmark_handler.go internal/httpapi/benchmark_handler_test.go cmd/evaluator/main.go
git commit -m "Expose benchmark run handler"
```

### Task 3: Add a CLI entrypoint for local benchmark runs

**Files:**
- Modify: `cmd/evaluator/main.go`
- Create: `cmd/evaluator/benchmark_cli_test.go`

**Interfaces:**
- Consumes: the shared `BenchmarkService`
- Produces: a `benchmark` CLI mode that prints the report as JSON to stdout

- [ ] **Step 1: Write the failing test**

```go
func TestBenchmarkCLIPrintsJSON(t *testing.T) {
	out, err := runBenchmarkCLI([]string{"benchmark"}, fakeBenchmarkService{report: BenchmarkReport{TotalTasks: 1}})
	if err != nil {
		t.Fatalf("runBenchmarkCLI() error = %v", err)
	}
	if !strings.Contains(out, "\"total_tasks\"") {
		t.Fatalf("CLI output = %s, want benchmark JSON", out)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./cmd/evaluator -run TestBenchmarkCLIPrintsJSON -v`

Expected: compile failure because the CLI path does not exist yet.

- [ ] **Step 3: Write minimal implementation**

```go
func runBenchmarkCLI(args []string, service BenchmarkServiceAPI) (string, error) {
	if len(args) > 0 && args[0] == "benchmark" {
		report, err := service.Run(context.Background())
		if err != nil {
			return "", err
		}
		raw, err := json.Marshal(report)
		if err != nil {
			return "", err
		}
		return string(raw), nil
	}

	return "", fmt.Errorf("unsupported command")
}
```

In `main`, branch on `os.Args[1] == "benchmark"` and call the CLI runner before starting the server.

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./cmd/evaluator -run TestBenchmarkCLIPrintsJSON -v`

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add cmd/evaluator/main.go cmd/evaluator/benchmark_cli_test.go
git commit -m "Add benchmark CLI mode"
```

## Verification Plan

Run:

```bash
go test ./internal/benchmark -run TestBenchmarkServiceRunReturnsReport -v
go test ./internal/httpapi -run TestBenchmarkRunHandlerReturnsJSON -v
go test ./cmd/evaluator -run TestBenchmarkCLIPrintsJSON -v
```

Expected: all tests pass.

## Coverage Check

- shared benchmark service -> BenchmarkService
- HTTP surface -> /benchmark/run
- CLI surface -> benchmark mode in cmd/evaluator
- single source of truth for benchmark orchestration -> service layer only
