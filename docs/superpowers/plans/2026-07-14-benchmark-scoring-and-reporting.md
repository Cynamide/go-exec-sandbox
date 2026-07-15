# Benchmark Scoring and Reporting Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Turn raw run results into a benchmark report that clearly shows baseline success, scaffolded success, scaffold lift, and breakdowns by task family and scaffold.

**Architecture:** Separate scoring from execution so the runtime can stay adapter-agnostic while the report layer handles aggregation, presentation, and future comparisons. The report should preserve per-run detail and compute summaries that make the value of scaffolding obvious to researchers and applied teams.

**Tech Stack:** Go 1.24+, existing `internal/benchmark` package, `encoding/json`, `go test`.

## Global Constraints

- benchmark reporting that separates baseline performance from scaffolded performance
- scaffold comparison across the same task
- economically meaningful task evaluation
- deterministic or clearly bounded grading where possible
- toy puzzle benchmarks are out of scope
- leaderboard-first design that ignores workflow context is out of scope

---

### Task 1: Define the report shape for task-family and scaffold breakdowns

**Files:**
- Create: `internal/benchmark/report.go`
- Modify: `internal/benchmark/model.go`
- Test: `internal/benchmark/report_test.go`

**Interfaces:**
- Consumes: `Run`, `Task`, `TaskFamily`, `Scaffold`
- Produces: `BenchmarkReport`, `FamilySummary`, and `ScaffoldSummary`

- [ ] **Step 1: Write the failing test**

```go
func TestBenchmarkReportCarriesFamilyAndScaffoldBreakdowns(t *testing.T) {
	report := BenchmarkReport{
		TotalTasks: 2,
		ByFamily: map[string]FamilySummary{
			"software_engineering": {TotalTasks: 1, BaselineSuccessRate: 0.0, ScaffoldedSuccessRate: 1.0, Lift: 1.0},
		},
		ByScaffold: map[string]ScaffoldSummary{
			"tool-assisted": {TotalTasks: 2, BaselineSuccessRate: 0.5, ScaffoldedSuccessRate: 1.0, Lift: 0.5},
		},
	}

	if report.ByFamily["software_engineering"].Lift != 1.0 {
		t.Fatalf("family lift = %v, want 1.0", report.ByFamily["software_engineering"].Lift)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/benchmark -run TestBenchmarkReportCarriesFamilyAndScaffoldBreakdowns -v`

Expected: compile failure because `BenchmarkReport` does not exist yet.

- [ ] **Step 3: Write minimal implementation**

```go
type BenchmarkReport struct {
	TotalTasks            int                           `json:"total_tasks"`
	BaselineSuccessRate    float64                       `json:"baseline_success_rate"`
	ScaffoldedSuccessRate  float64                       `json:"scaffolded_success_rate"`
	Lift                   float64                       `json:"lift"`
	ByFamily               map[string]FamilySummary      `json:"by_family"`
	ByScaffold             map[string]ScaffoldSummary    `json:"by_scaffold"`
	Runs                   []Run                         `json:"runs"`
}

type FamilySummary struct {
	TotalTasks            int     `json:"total_tasks"`
	BaselineSuccessRate    float64 `json:"baseline_success_rate"`
	ScaffoldedSuccessRate  float64 `json:"scaffolded_success_rate"`
	Lift                   float64 `json:"lift"`
}

type ScaffoldSummary struct {
	TotalTasks            int     `json:"total_tasks"`
	BaselineSuccessRate    float64 `json:"baseline_success_rate"`
	ScaffoldedSuccessRate  float64 `json:"scaffolded_success_rate"`
	Lift                   float64 `json:"lift"`
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/benchmark -run TestBenchmarkReportCarriesFamilyAndScaffoldBreakdowns -v`

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/benchmark/report.go internal/benchmark/model.go internal/benchmark/report_test.go
git commit -m "Add benchmark report summaries"
```

### Task 2: Aggregate runs into report summaries

**Files:**
- Modify: `internal/benchmark/report.go`
- Modify: `internal/benchmark/report_test.go`

**Interfaces:**
- Consumes: `[]Run` and `[]Task`
- Produces: `BuildBenchmarkReport` with overall and grouped metrics

- [ ] **Step 1: Write the failing test**

```go
func TestBuildBenchmarkReportComputesLift(t *testing.T) {
	runs := []Run{
		{TaskID: "a", Mode: RunModeBaseline, Passed: false, Scaffold: Scaffold{Name: "baseline"}},
		{TaskID: "a", Mode: RunModeScaffolded, Passed: true, Scaffold: Scaffold{Name: "tool-assisted"}},
	}
	tasks := []Task{{ID: "a", TaskFamily: "software_engineering"}}

	report := BuildBenchmarkReport(tasks, runs)

	if report.Lift != 1.0 {
		t.Fatalf("Lift = %v, want 1.0", report.Lift)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/benchmark -run TestBuildBenchmarkReportComputesLift -v`

Expected: compile failure because `BuildBenchmarkReport` does not exist yet.

- [ ] **Step 3: Write minimal implementation**

```go
func BuildBenchmarkReport(tasks []Task, runs []Run) BenchmarkReport {
	report := BenchmarkReport{
		ByFamily:   map[string]FamilySummary{},
		ByScaffold: map[string]ScaffoldSummary{},
		Runs:       runs,
	}

	taskModes := map[string]map[RunMode]bool{}
	baselinePassed := 0
	scaffoldedPassed := 0

	for _, run := range runs {
		if taskModes[run.TaskID] == nil {
			taskModes[run.TaskID] = map[RunMode]bool{}
		}
		taskModes[run.TaskID][run.Mode] = run.Passed
		if run.Mode == RunModeBaseline && run.Passed {
			baselinePassed++
		}
		if run.Mode == RunModeScaffolded && run.Passed {
			scaffoldedPassed++
		}
	}

	total := float64(len(tasks))
	if total > 0 {
		report.TotalTasks = len(tasks)
		report.BaselineSuccessRate = float64(baselinePassed) / total
		report.ScaffoldedSuccessRate = float64(scaffoldedPassed) / total
		report.Lift = report.ScaffoldedSuccessRate - report.BaselineSuccessRate
	}

	for _, task := range tasks {
		family := task.TaskFamily
		summary := report.ByFamily[family]
		summary.TotalTasks++
		if taskModes[task.ID][RunModeBaseline] {
			summary.BaselineSuccessRate += 1
		}
		if taskModes[task.ID][RunModeScaffolded] {
			summary.ScaffoldedSuccessRate += 1
		}
		report.ByFamily[family] = summary
	}

	for family, summary := range report.ByFamily {
		totalFamily := float64(summary.TotalTasks)
		summary.BaselineSuccessRate = summary.BaselineSuccessRate / totalFamily
		summary.ScaffoldedSuccessRate = summary.ScaffoldedSuccessRate / totalFamily
		summary.Lift = summary.ScaffoldedSuccessRate - summary.BaselineSuccessRate
		report.ByFamily[family] = summary
	}

	scaffoldTotals := map[string]ScaffoldSummary{}
	for _, run := range runs {
		summary := scaffoldTotals[run.Scaffold.Name]
		summary.TotalTasks++
		if run.Mode == RunModeBaseline && run.Passed {
			summary.BaselineSuccessRate++
		}
		if run.Mode == RunModeScaffolded && run.Passed {
			summary.ScaffoldedSuccessRate++
		}
		scaffoldTotals[run.Scaffold.Name] = summary
	}

	for name, summary := range scaffoldTotals {
		totalScaffold := float64(summary.TotalTasks)
		if totalScaffold > 0 {
			summary.BaselineSuccessRate = summary.BaselineSuccessRate / totalScaffold
			summary.ScaffoldedSuccessRate = summary.ScaffoldedSuccessRate / totalScaffold
			summary.Lift = summary.ScaffoldedSuccessRate - summary.BaselineSuccessRate
		}
		report.ByScaffold[name] = summary
	}

	return report
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/benchmark -run TestBuildBenchmarkReportComputesLift -v`

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/benchmark/report.go internal/benchmark/report_test.go
git commit -m "Aggregate benchmark runs into reports"
```

### Task 3: Add JSON serialization for report output

**Files:**
- Modify: `internal/benchmark/report_test.go`

**Interfaces:**
- Consumes: `BenchmarkReport`
- Produces: stable JSON that can be returned by the service or CLI without extra mapping

- [ ] **Step 1: Write the failing test**

```go
func TestBenchmarkReportJSONIncludesLiftAndRuns(t *testing.T) {
	report := BenchmarkReport{
		TotalTasks: 1,
		Lift:       1.0,
		Runs:       []Run{{TaskID: "a", Mode: RunModeScaffolded, Passed: true}},
	}

	raw, err := json.Marshal(report)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	if !strings.Contains(string(raw), "\"lift\":1") {
		t.Fatalf("report JSON = %s, want lift field", string(raw))
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/benchmark -run TestBenchmarkReportJSONIncludesLiftAndRuns -v`

Expected: compile failure because the report fields or imports are missing.

- [ ] **Step 3: Write minimal implementation**

Make sure all report structs have explicit JSON tags and that `Runs` is preserved in the output.

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/benchmark -run TestBenchmarkReportJSONIncludesLiftAndRuns -v`

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/benchmark/report_test.go internal/benchmark/report.go
git commit -m "Make benchmark report JSON stable"
```

## Verification Plan

Run:

```bash
go test ./internal/benchmark -run TestBenchmarkReportCarriesFamilyAndScaffoldBreakdowns -v
go test ./internal/benchmark -run TestBuildBenchmarkReportComputesLift -v
go test ./internal/benchmark -run TestBenchmarkReportJSONIncludesLiftAndRuns -v
```

Expected: all tests pass.

## Coverage Check

- overall benchmark report -> BenchmarkReport
- family breakdown -> FamilySummary
- scaffold breakdown -> ScaffoldSummary
- lift calculation -> BuildBenchmarkReport
- service-ready serialization -> JSON test
