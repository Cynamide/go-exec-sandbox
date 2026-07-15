# Scaffold-Aware Benchmark Harness Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Turn LocalEval into a scaffold-aware benchmark harness that measures baseline performance, scaffolded performance, and scaffold lift on economically meaningful tasks.

**Architecture:** Keep the current Docker-backed execution path, but add benchmark metadata for scaffold modes and a report shape that compares the same task under multiple workflow conditions. The benchmark package remains the evaluation core, `data/problems.json` becomes the first task registry with richer metadata, and the README repositions the repo around system-level benchmarking rather than puzzle solving.

**Tech Stack:** Go 1.24+, Docker, JSON fixtures, `go test`, existing `internal/benchmark`, `internal/sandbox`, and `internal/api` packages.

## Global Constraints

- economically meaningful task evaluation
- scaffold comparison across the same task
- automatic execution where possible
- deterministic or clearly bounded grading where possible
- support for tasks that produce artifacts, not just short answers
- benchmark reporting that separates baseline performance from scaffolded performance
- toy puzzle benchmarks are out of scope
- leaderboard-first design that ignores workflow context is out of scope
- optimization around narrow interview-style code challenges is out of scope

---

### Task 1: Add scaffold-aware benchmark data models

**Files:**
- Modify: `internal/benchmark/types.go`
- Modify: `internal/benchmark/harness_test.go`

**Interfaces:**
- Consumes: existing `Problem`, `TestCase`, and `Report` concepts
- Produces: `Scaffold`, `RunMode`, `TaskRun`, and `ComparisonReport` types that later harness code can use

- [ ] **Step 1: Write the failing test**

```go
func TestComparisonReportCapturesBaselineAndScaffoldLift(t *testing.T) {
	report := ComparisonReport{
		TotalTasks: 2,
		BaselineSuccessRate: 0.50,
		ScaffoldedSuccessRate: 1.00,
		Lift: 0.50,
	}

	if report.Lift != 0.50 {
		t.Fatalf("Lift = %v, want 0.50", report.Lift)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/benchmark -run TestComparisonReportCapturesBaselineAndScaffoldLift -v`

Expected: compile failure because `ComparisonReport` does not exist yet.

- [ ] **Step 3: Write minimal implementation**

```go
type Scaffold struct {
	Name         string `json:"name"`
	Description  string `json:"description"`
	PromptPrefix string `json:"prompt_prefix"`
}

type RunMode string

const (
	RunModeBaseline  RunMode = "baseline"
	RunModeScaffolded RunMode = "scaffolded"
)

type TaskRun struct {
	ProblemID string   `json:"problem_id"`
	Mode      RunMode  `json:"mode"`
	Scaffold  Scaffold `json:"scaffold"`
	Passed    bool     `json:"passed"`
}

type ComparisonReport struct {
	TotalTasks           int       `json:"total_tasks"`
	BaselineSuccessRate  float64   `json:"baseline_success_rate"`
	ScaffoldedSuccessRate float64   `json:"scaffolded_success_rate"`
	Lift                 float64   `json:"lift"`
	Runs                 []TaskRun `json:"runs"`
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/benchmark -run TestComparisonReportCapturesBaselineAndScaffoldLift -v`

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/benchmark/types.go internal/benchmark/harness_test.go
git commit -m "Add scaffold-aware benchmark types"
```

### Task 2: Implement scaffold-aware benchmark execution and comparison reporting

**Files:**
- Modify: `internal/benchmark/harness.go`
- Modify: `internal/benchmark/harness_test.go`
- Create: `internal/benchmark/scaffold.go`

**Interfaces:**
- Consumes: `Problem`, `TestCase`, `Scaffold`, `RunMode`, `TaskRun`, `ComparisonReport`
- Produces: a harness function that can run a task under baseline and scaffolded modes and return a comparison report

- [ ] **Step 1: Write the failing test**

```go
func TestRunComparisonReportsBaselineAndScaffoldLift(t *testing.T) {
	t.Setenv("OLLAMA_MODEL", "test-model")

	originalRunner := runCodeInSandbox
	t.Cleanup(func() { runCodeInSandbox = originalRunner })

	runCodeInSandbox = func(req api.ExecutionRequest, cfg config.Config) (api.ExecutionResponse, error) {
		if strings.Contains(req.SourceCode, "ok") {
			return api.ExecutionResponse{Stdout: "ok"}, nil
		}
		return api.ExecutionResponse{Stdout: "fail"}, nil
	}

	client := scaffoldStubLLMClient{}
	report := RunComparison([]Problem{{
		ID: "task-1",
		Description: "demo task",
		Language: "python",
		TestCases: []TestCase{{Input: "ok", ExpectedOutput: "ok"}},
	}}, []Scaffold{
		{Name: "baseline", PromptPrefix: "baseline: "},
		{Name: "tool-assisted", PromptPrefix: "tool-assisted: "},
	}, client)

	if report.TotalTasks != 1 {
		t.Fatalf("TotalTasks = %d, want 1", report.TotalTasks)
	}

	if len(report.Runs) != 2 {
		t.Fatalf("len(Runs) = %d, want 2", len(report.Runs))
	}

	if report.BaselineSuccessRate >= report.ScaffoldedSuccessRate {
		t.Fatalf("scaffolded success rate = %v, want greater than baseline %v", report.ScaffoldedSuccessRate, report.BaselineSuccessRate)
	}
}

type scaffoldStubLLMClient struct{}

func (c scaffoldStubLLMClient) GenerateCode(problem string, language string) (string, error) {
	if strings.Contains(problem, "tool-assisted") {
		return "```python\nprint('ok')\n```", nil
	}

	return "```python\nprint('fail')\n```", nil
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/benchmark -run TestRunComparisonReportsBaselineAndScaffoldLift -v`

Expected: compile failure because `RunComparison` does not exist yet.

- [ ] **Step 3: Write minimal implementation**

```go
func RunComparison(problems []Problem, scaffolds []Scaffold, client LLMClient) ComparisonReport {
	if len(problems) == 0 {
		return ComparisonReport{}
	}

	var baselinePassed, scaffoldedPassed int
	var runs []TaskRun

	for _, problem := range problems {
		baseline := runProblem(problem, scaffolds[0], RunModeBaseline, client)
		scaffolded := runProblem(problem, scaffolds[1], RunModeScaffolded, client)

		runs = append(runs, baseline, scaffolded)
		if baseline.Passed {
			baselinePassed++
		}
		if scaffolded.Passed {
			scaffoldedPassed++
		}
	}

	total := float64(len(problems))
	return ComparisonReport{
		TotalTasks:            len(problems),
		BaselineSuccessRate:    float64(baselinePassed) / total,
		ScaffoldedSuccessRate:  float64(scaffoldedPassed) / total,
		Lift:                   float64(scaffoldedPassed-baselinePassed) / total,
		Runs:                   runs,
	}
}

func runProblem(problem Problem, scaffold Scaffold, mode RunMode, client LLMClient) TaskRun {
	prompt := scaffold.PromptPrefix + problem.Description
	code, _ := client.GenerateCode(prompt, problem.Language)
	code = extractCode(code)

	passed := true
	for _, tc := range problem.TestCases {
		req := api.ExecutionRequest{
			Language:   problem.Language,
			SourceCode: code,
			Stdin:      tc.Input,
		}

		resp, err := runCodeInSandbox(req, config.LoadConfig())
		if err != nil || strings.TrimSpace(resp.Stdout) != strings.TrimSpace(tc.ExpectedOutput) {
			passed = false
			break
		}
	}

	return TaskRun{
		ProblemID: problem.ID,
		Mode:      mode,
		Scaffold:  scaffold,
		Passed:    passed,
	}
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/benchmark -run TestRunComparisonReportsBaselineAndScaffoldLift -v`

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/benchmark/scaffold.go internal/benchmark/harness.go internal/benchmark/harness_test.go
git commit -m "Add scaffold-aware benchmark comparison"
```

### Task 3: Expand the sample problem registry for task-family metadata

**Files:**
- Modify: `data/problems.json`
- Modify: `internal/benchmark/types.go`
- Modify: `internal/benchmark/harness_test.go`

**Interfaces:**
- Consumes: the existing `Problem` JSON shape
- Produces: `Problem` entries that can carry task-family metadata without breaking current code execution

- [ ] **Step 1: Write the failing test**

```go
func TestProblemUnmarshalIncludesTaskFamily(t *testing.T) {
	raw := []byte(`{"id":"task-1","title":"Example","description":"demo","language":"python","task_family":"software_engineering","test_cases":[]}`)

	var p Problem
	if err := json.Unmarshal(raw, &p); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if p.TaskFamily != "software_engineering" {
		t.Fatalf("TaskFamily = %q, want %q", p.TaskFamily, "software_engineering")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/benchmark -run TestProblemUnmarshalIncludesTaskFamily -v`

Expected: compile failure because `TaskFamily` does not exist yet.

- [ ] **Step 3: Write minimal implementation**

```go
type Problem struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	TaskFamily  string     `json:"task_family"`
	Language    string     `json:"language"`
	TestCases   []TestCase `json:"test_cases"`
}
```

Update `data/problems.json` so each sample problem includes a `task_family` value such as `software_engineering`.

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/benchmark -run TestProblemUnmarshalIncludesTaskFamily -v`

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add data/problems.json internal/benchmark/types.go internal/benchmark/harness_test.go
git commit -m "Add task family metadata to benchmark problems"
```

### Task 4: Reposition the README around scaffold-aware benchmarking

**Files:**
- Modify: `README.md`
- Modify: `PROJECT_CONTEXT_MAP.md`

**Interfaces:**
- Consumes: the final benchmark terminology from Tasks 1 to 3
- Produces: repo-facing documentation that explains the product thesis, evaluation modes, and benchmark value

- [ ] **Step 1: Write the failing documentation check**

```bash
grep -n "LeetCode-style\|puzzle" README.md
```

Expected: current README still contains puzzle-like framing that needs to be replaced.

- [ ] **Step 2: Update the README introduction**

```md
LocalEval is a scaffold-aware benchmark harness for economically meaningful agent tasks.

It measures:

- model-only performance
- scaffolded performance
- the gap between the two
```

- [ ] **Step 3: Update the context map wording**

```md
- The repository is being positioned as a scaffold-aware benchmark harness for economically meaningful agent tasks
- The benchmark package should report baseline, scaffolded, and delta metrics
```

- [ ] **Step 4: Re-run the documentation check**

Run: `grep -n "LeetCode-style\|puzzle" README.md`

Expected: no matches.

- [ ] **Step 5: Commit**

```bash
git add README.md PROJECT_CONTEXT_MAP.md
git commit -m "Reframe repo around scaffold-aware benchmarking"
```

## Verification Plan

After all tasks land, run:

```bash
go test ./...
```

Expected: all package tests pass.

Also verify the README positioning manually by checking the first screen of the document:

```bash
sed -n '1,80p' README.md
```

Expected: the top of the file leads with scaffold-aware benchmarking, not interview-style puzzle framing.

## Coverage Check

- spec: economically meaningful task evaluation -> Tasks 2, 3, 4
- spec: scaffold comparison across the same task -> Task 2
- spec: automatic execution where possible -> Task 2
- spec: deterministic or clearly bounded grading where possible -> Task 2
- spec: support for tasks that produce artifacts -> Task 2
- spec: benchmark reporting that separates baseline and scaffolded performance -> Tasks 1, 2
- spec: README positioning -> Task 4
- spec: task-family metadata -> Task 3
