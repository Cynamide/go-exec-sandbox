# Scaffold-Aware Runtime and Adapters Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a generic task runner that can execute scaffold-aware benchmark runs while keeping code execution as one adapter instead of the architectural center.

**Architecture:** The benchmark package should orchestrate a task, a scaffold, a generator, an executor, and a grader through a small runtime API. The first adapter will wrap the existing Docker sandbox, but the runtime itself must stay execution-agnostic so browser, spreadsheet, and document adapters can be added later without changing the task model.

**Tech Stack:** Go 1.24+, existing `internal/benchmark`, `internal/sandbox`, `internal/api`, `internal/config` packages, `go test`.

## Global Constraints

- economically meaningful task evaluation
- scaffold comparison across the same task
- automatic execution where possible
- support for tasks that produce artifacts, not just short answers
- deterministic or clearly bounded grading where possible
- toy puzzle benchmarks are out of scope
- leaderboard-first design that ignores workflow context is out of scope
- optimization around narrow interview-style code challenges is out of scope

---

### Task 1: Introduce the generic task runtime

**Files:**
- Create: `internal/benchmark/runtime.go`
- Modify: `internal/benchmark/model.go`
- Test: `internal/benchmark/runtime_test.go`

**Interfaces:**
- Consumes: `Task`, `Scaffold`, `RunMode`, `Run`, `Outcome`, and the existing `LLMClient`
- Produces: `Executor`, `Grader`, and `RunTask` so later adapters can plug into one shared flow

- [ ] **Step 1: Write the failing test**

```go
func TestRunTaskAppliesScaffoldPromptPrefix(t *testing.T) {
	exec := fakeExecutor{
		resp: api.ExecutionResponse{Stdout: "ok"},
	}

	client := fakeLLMClient{
		code: "```python\nprint('ok')\n```",
	}

	task := Task{
		ID:          "task-1",
		Description: "solve this",
		Language:    "python",
		TestCases:   []TestCase{{Input: "", ExpectedOutput: "ok"}},
	}

	run := RunTask(task, Scaffold{Name: "tool-assisted", PromptPrefix: "tool-assisted: "}, RunModeScaffolded, client, exec, config.Config{})

	if run.Mode != RunModeScaffolded {
		t.Fatalf("Mode = %q, want scaffolded", run.Mode)
	}
	if run.Passed != true {
		t.Fatalf("Passed = %v, want true", run.Passed)
	}
	if client.seenPrompt != "tool-assisted: solve this" {
		t.Fatalf("seenPrompt = %q, want scaffolded prompt", client.seenPrompt)
	}
}

type fakeExecutor struct {
	resp api.ExecutionResponse
}

func (f fakeExecutor) Execute(req api.ExecutionRequest, cfg config.Config) (api.ExecutionResponse, error) {
	return f.resp, nil
}

type fakeLLMClient struct {
	code       string
	seenPrompt string
}

func (f *fakeLLMClient) GenerateCode(problem string, language string) (string, error) {
	f.seenPrompt = problem
	return f.code, nil
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/benchmark -run TestRunTaskAppliesScaffoldPromptPrefix -v`

Expected: compile failure because `RunTask` does not exist yet.

- [ ] **Step 3: Write minimal implementation**

```go
type Executor interface {
	Execute(req api.ExecutionRequest, cfg config.Config) (api.ExecutionResponse, error)
}

type Grader interface {
	Grade(task Task, resp api.ExecutionResponse, tc TestCase) Outcome
}

func RunTask(task Task, scaffold Scaffold, mode RunMode, client LLMClient, exec Executor, cfg config.Config) Run {
	prompt := scaffold.ApplyPrompt(task.Description)
	code, _ := client.GenerateCode(prompt, task.Language)
	code = extractCode(code)

	passed := true
	for _, tc := range task.TestCases {
		resp, err := exec.Execute(api.ExecutionRequest{
			Language:   task.Language,
			SourceCode: code,
			Stdin:      tc.Input,
			TimeoutMS:  cfg.DefaultTimeoutMS,
		}, cfg)
		if err != nil || strings.TrimSpace(resp.Stdout) != strings.TrimSpace(tc.ExpectedOutput) {
			passed = false
			break
		}
	}

	return Run{
		TaskID:   task.ID,
		Mode:     mode,
		Scaffold: scaffold,
		Passed:   passed,
	}
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/benchmark -run TestRunTaskAppliesScaffoldPromptPrefix -v`

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/benchmark/runtime.go internal/benchmark/model.go internal/benchmark/runtime_test.go
git commit -m "Add generic benchmark runtime"
```

### Task 2: Wrap the existing Docker sandbox as the first executor adapter

**Files:**
- Create: `internal/benchmark/code_execution_adapter.go`
- Modify: `internal/benchmark/runtime_test.go`
- Modify: `internal/sandbox/docker.go` if the adapter needs a small seam for testability

**Interfaces:**
- Consumes: `Executor`, `api.ExecutionRequest`, `api.ExecutionResponse`
- Produces: a code execution adapter that delegates to the existing sandbox without hard-coding it into the runtime

- [ ] **Step 1: Write the failing test**

```go
func TestCodeExecutionAdapterDelegatesToSandbox(t *testing.T) {
	adapter := CodeExecutionAdapter{
		Runner: func(req api.ExecutionRequest, cfg config.Config) (api.ExecutionResponse, error) {
			if req.Language != "python" {
				t.Fatalf("Language = %q, want python", req.Language)
			}
			return api.ExecutionResponse{Stdout: "ok"}, nil
		},
	}

	resp, err := adapter.Execute(api.ExecutionRequest{Language: "python", SourceCode: "print('ok')"}, config.Config{})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if resp.Stdout != "ok" {
		t.Fatalf("Stdout = %q, want ok", resp.Stdout)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/benchmark -run TestCodeExecutionAdapterDelegatesToSandbox -v`

Expected: compile failure because `CodeExecutionAdapter` does not exist yet.

- [ ] **Step 3: Write minimal implementation**

```go
type CodeExecutionAdapter struct {
	Runner func(req api.ExecutionRequest, cfg config.Config) (api.ExecutionResponse, error)
}

func (a CodeExecutionAdapter) Execute(req api.ExecutionRequest, cfg config.Config) (api.ExecutionResponse, error) {
	return a.Runner(req, cfg)
}
```

Wire the production runner to `sandbox.RunCodeInSandbox` so the runtime depends on the adapter, not on the sandbox package directly.

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/benchmark -run TestCodeExecutionAdapterDelegatesToSandbox -v`

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/benchmark/code_execution_adapter.go internal/benchmark/runtime_test.go
git commit -m "Add code execution adapter"
```

### Task 3: Add a grader seam so future adapters can share the same scoring contract

**Files:**
- Create: `internal/benchmark/grader.go`
- Modify: `internal/benchmark/runtime_test.go`

**Interfaces:**
- Consumes: `Task`, `TestCase`, `Outcome`, `Run`
- Produces: a grader implementation that can later score non-code tasks without changing the runtime loop

- [ ] **Step 1: Write the failing test**

```go
func TestDefaultGraderMarksMatchingStdoutAsPass(t *testing.T) {
	grader := DefaultGrader{}
	outcome := grader.Grade(Task{}, api.ExecutionResponse{Stdout: "hello\n"}, TestCase{ExpectedOutput: "hello"})

	if !outcome.Passed {
		t.Fatalf("Outcome.Passed = false, want true")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/benchmark -run TestDefaultGraderMarksMatchingStdoutAsPass -v`

Expected: compile failure because `DefaultGrader` does not exist yet.

- [ ] **Step 3: Write minimal implementation**

```go
type DefaultGrader struct{}

func (DefaultGrader) Grade(task Task, resp api.ExecutionResponse, tc TestCase) Outcome {
	passed := strings.TrimSpace(resp.Stdout) == strings.TrimSpace(tc.ExpectedOutput)
	score := 0.0
	if passed {
		score = 1.0
	}
	return Outcome{Passed: passed, Score: score}
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/benchmark -run TestDefaultGraderMarksMatchingStdoutAsPass -v`

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/benchmark/grader.go internal/benchmark/runtime_test.go
git commit -m "Add benchmark grader seam"
```

## Verification Plan

Run:

```bash
go test ./internal/benchmark -run TestRunTaskAppliesScaffoldPromptPrefix -v
go test ./internal/benchmark -run TestCodeExecutionAdapterDelegatesToSandbox -v
go test ./internal/benchmark -run TestDefaultGraderMarksMatchingStdoutAsPass -v
```

Expected: all three tests pass.

## Coverage Check

- generic runtime -> RunTask, Executor, Grader
- code execution as adapter -> CodeExecutionAdapter
- behavior first -> scaffolded prompt application verified before adapter specifics
- future extensibility -> grader seam exists for non-code task families
