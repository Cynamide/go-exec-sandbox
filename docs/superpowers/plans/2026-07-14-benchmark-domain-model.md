# Benchmark Domain Model Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Define the core benchmark behavior objects so the repo can describe tasks, scaffolds, run modes, and outcomes without depending on any specific execution backend.

**Architecture:** Introduce a task-centric domain model in `internal/benchmark` that captures the benchmark shape independently from code execution. The model should make scaffold-aware evaluation explicit by separating the task, the scaffold, the run, and the graded outcome into small types that later plans can reuse.

**Tech Stack:** Go 1.24+, `go test`, `encoding/json`, existing `internal/benchmark` package.

## Global Constraints

- economically meaningful task evaluation
- scaffold comparison across the same task
- support for tasks that produce artifacts, not just short answers
- deterministic or clearly bounded grading where possible
- toy puzzle benchmarks are out of scope
- leaderboard-first design that ignores workflow context is out of scope
- optimization around narrow interview-style code challenges is out of scope

---

### Task 1: Introduce task, scaffold, run, and outcome types

**Files:**
- Create: `internal/benchmark/model.go`
- Modify: `internal/benchmark/types.go`
- Test: `internal/benchmark/model_test.go`

**Interfaces:**
- Consumes: the current `Problem` and `TestCase` shape
- Produces: `Task`, `Scaffold`, `RunMode`, `Run`, and `Outcome` types for later runtime and scoring plans

- [ ] **Step 1: Write the failing test**

```go
func TestScaffoldAppliesPromptPrefix(t *testing.T) {
	scaffold := Scaffold{
		Name:         "tool-assisted",
		PromptPrefix: "Use the calculator and browser before answering.\n\n",
	}

	got := scaffold.ApplyPrompt("Solve the task.")
	want := "Use the calculator and browser before answering.\n\nSolve the task."

	if got != want {
		t.Fatalf("ApplyPrompt() = %q, want %q", got, want)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/benchmark -run TestScaffoldAppliesPromptPrefix -v`

Expected: compile failure because `Scaffold` and `ApplyPrompt` do not exist yet.

- [ ] **Step 3: Write minimal implementation**

```go
type Task struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	TaskFamily  string     `json:"task_family"`
	Language    string     `json:"language"`
	TestCases   []TestCase `json:"test_cases"`
}

type Scaffold struct {
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	PromptPrefix string   `json:"prompt_prefix"`
	Tools        []string `json:"tools,omitempty"`
}

type RunMode string

const (
	RunModeBaseline   RunMode = "baseline"
	RunModeScaffolded RunMode = "scaffolded"
)

type Run struct {
	TaskID      string  `json:"task_id"`
	Mode        RunMode `json:"mode"`
	Scaffold    Scaffold `json:"scaffold"`
	Passed      bool    `json:"passed"`
	Output      string  `json:"output,omitempty"`
	Error       string  `json:"error,omitempty"`
}

type Outcome struct {
	Passed bool    `json:"passed"`
	Score  float64 `json:"score"`
}

func (s Scaffold) ApplyPrompt(prompt string) string {
	return s.PromptPrefix + prompt
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/benchmark -run TestScaffoldAppliesPromptPrefix -v`

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/benchmark/model.go internal/benchmark/types.go internal/benchmark/model_test.go
git commit -m "Add benchmark domain model"
```

### Task 2: Keep the existing problem fixtures compatible while migrating to tasks

**Files:**
- Modify: `internal/benchmark/types.go`
- Modify: `internal/benchmark/model_test.go`

**Interfaces:**
- Consumes: the new `Task` type
- Produces: compatibility aliases so the current code keeps working while later plans migrate call sites

- [ ] **Step 1: Write the failing test**

```go
func TestTaskJSONRoundTripPreservesFamily(t *testing.T) {
	task := Task{
		ID:          "task-1",
		Title:       "Example",
		Description: "demo",
		TaskFamily:  "software_engineering",
		Language:    "python",
		TestCases:   []TestCase{{Input: "1", ExpectedOutput: "1"}},
	}

	raw, err := json.Marshal(task)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var got Task
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if got.TaskFamily != "software_engineering" {
		t.Fatalf("TaskFamily = %q, want %q", got.TaskFamily, "software_engineering")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/benchmark -run TestTaskJSONRoundTripPreservesFamily -v`

Expected: compile failure because the JSON shape and imports are not ready yet.

- [ ] **Step 3: Write minimal implementation**

```go
type Problem = Task
```

Keep `TestCase` unchanged so existing benchmark fixtures can still unmarshal.

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/benchmark -run TestTaskJSONRoundTripPreservesFamily -v`

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/benchmark/types.go internal/benchmark/model_test.go
git commit -m "Migrate benchmark fixtures to task model"
```

## Verification Plan

Run:

```bash
go test ./internal/benchmark -run TestScaffoldAppliesPromptPrefix -v
go test ./internal/benchmark -run TestTaskJSONRoundTripPreservesFamily -v
```

Expected: both tests pass.

## Coverage Check

- task-centric domain model -> Task, Scaffold, Run, Outcome
- scaffold-aware evaluation -> ApplyPrompt and RunMode
- backward compatibility -> Problem alias
- behavior-first architecture -> all later plans consume these types instead of inventing new ones
