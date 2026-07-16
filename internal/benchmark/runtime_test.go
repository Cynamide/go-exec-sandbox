package benchmark

import (
	"context"
	"testing"

	"gexec-sandbox/internal/api"
	"gexec-sandbox/internal/config"
)

func TestRunTaskAppliesScaffoldPromptPrefix(t *testing.T) {
	exec := fakeExecutor{
		resp: api.ExecutionResponse{Stdout: "ok"},
	}

	client := &fakeLLMClient{
		code: "```python\nprint('ok')\n```",
	}

	task := Task{
		ID:          "task-1",
		Description: "solve this",
		Language:    "python",
		TestCases:   []TestCase{{Input: "", ExpectedOutput: "ok"}},
	}

	run := RunTask(
		context.Background(),
		task,
		Scaffold{Name: "tool-assisted", PromptPrefix: "tool-assisted: "},
		RunModeScaffolded,
		client,
		exec,
		config.Config{DefaultTimeoutMS: 1234},
	)

	if run.Mode != RunModeScaffolded {
		t.Fatalf("Mode = %q, want %q", run.Mode, RunModeScaffolded)
	}

	if !run.Passed {
		t.Fatalf("Passed = %v, want true", run.Passed)
	}

	if client.seenPrompt != "tool-assisted: solve this" {
		t.Fatalf("seenPrompt = %q, want %q", client.seenPrompt, "tool-assisted: solve this")
	}
}

func TestRunTaskWithGraderUsesInjectedGrader(t *testing.T) {
	exec := fakeExecutor{
		resp: api.ExecutionResponse{Stdout: "executor output"},
	}

	client := &fakeLLMClient{
		code: "```python\nprint('executor output')\n```",
	}

	task := Task{
		ID:          "task-2",
		Description: "solve this too",
		Language:    "python",
		TestCases:   []TestCase{{Input: "sample input", ExpectedOutput: "expected by grader"}},
	}

	grader := &fakeGrader{
		outcome: Outcome{Passed: false, Score: 0.25},
	}

	run := RunTaskWithGrader(
		context.Background(),
		task,
		Scaffold{Name: "tool-assisted", PromptPrefix: "tool-assisted: "},
		RunModeBaseline,
		client,
		exec,
		grader,
		config.Config{DefaultTimeoutMS: 1234},
	)

	if run.Passed {
		t.Fatalf("Passed = %v, want false from injected grader", run.Passed)
	}

	if len(run.Outcomes) != 1 {
		t.Fatalf("len(Outcomes) = %d, want 1", len(run.Outcomes))
	}

	if run.Outcomes[0] != grader.outcome {
		t.Fatalf("Outcome = %#v, want %#v", run.Outcomes[0], grader.outcome)
	}

	if grader.seenTask.ID != task.ID {
		t.Fatalf("grader task ID = %q, want %q", grader.seenTask.ID, task.ID)
	}

	if grader.seenResp.Stdout != exec.resp.Stdout {
		t.Fatalf("grader stdout = %q, want %q", grader.seenResp.Stdout, exec.resp.Stdout)
	}

	if grader.seenTestCase != task.TestCases[0] {
		t.Fatalf("grader test case = %#v, want %#v", grader.seenTestCase, task.TestCases[0])
	}
}

func TestDefaultGraderMarksMatchingStdoutAsPass(t *testing.T) {
	grader := DefaultGrader{}
	outcome := grader.Grade(Task{}, api.ExecutionResponse{Stdout: "hello\n"}, TestCase{ExpectedOutput: "hello"})

	if !outcome.Passed {
		t.Fatalf("Outcome.Passed = false, want true")
	}
}

func TestDefaultGraderAcceptsEquivalentMarkdownTables(t *testing.T) {
	grader := DefaultGrader{}
	outcome := grader.Grade(Task{}, api.ExecutionResponse{Stdout: " | team | count |\n| --- | --- |\n| api | 2 | \n| web | 1 |"}, TestCase{ExpectedOutput: "| team | count |\n| --- | --- |\n| api | 2 |\n| web | 1 |"})

	if !outcome.Passed {
		t.Fatalf("Outcome.Passed = false, want true")
	}
}

func TestDefaultGraderAcceptsEquivalentCSVTables(t *testing.T) {
	grader := DefaultGrader{}
	outcome := grader.Grade(Task{}, api.ExecutionResponse{Stdout: "month,total\n2026-05,25\n2026-06,7\n"}, TestCase{ExpectedOutput: "month,total\n2026-05,25\n2026-06,7"})

	if !outcome.Passed {
		t.Fatalf("Outcome.Passed = false, want true")
	}
}

func TestDefaultGraderAcceptsEquivalentJSON(t *testing.T) {
	grader := DefaultGrader{}
	outcome := grader.Grade(Task{}, api.ExecutionResponse{Stdout: "{\n  \"errors\": [\n    {\"message\": \"Auth failed\", \"count\": 3},\n    {\"message\": \"Timeout\", \"count\": 2}\n  ]\n}"}, TestCase{ExpectedOutput: "{\"errors\":[{\"message\":\"Auth failed\",\"count\":3},{\"message\":\"Timeout\",\"count\":2}]}"})

	if !outcome.Passed {
		t.Fatalf("Outcome.Passed = false, want true")
	}
}

func TestCodeExecutionAdapterDelegatesToSandbox(t *testing.T) {
	adapter := CodeExecutionAdapter{
		Runner: func(ctx context.Context, req api.ExecutionRequest, cfg config.Config) (api.ExecutionResponse, error) {
			if req.Language != "python" {
				t.Fatalf("Language = %q, want python", req.Language)
			}
			return api.ExecutionResponse{Stdout: "ok"}, nil
		},
	}

	resp, err := adapter.Execute(context.Background(), api.ExecutionRequest{Language: "python", SourceCode: "print('ok')"}, config.Config{})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if resp.Stdout != "ok" {
		t.Fatalf("Stdout = %q, want ok", resp.Stdout)
	}
}

type fakeExecutor struct {
	resp api.ExecutionResponse
}

func (f fakeExecutor) Execute(ctx context.Context, req api.ExecutionRequest, cfg config.Config) (api.ExecutionResponse, error) {
	return f.resp, nil
}

type fakeLLMClient struct {
	code       string
	seenPrompt string
}

func (f *fakeLLMClient) GenerateCode(ctx context.Context, problem string, language string) (string, error) {
	f.seenPrompt = problem
	return f.code, nil
}

type fakeGrader struct {
	outcome      Outcome
	seenTask     Task
	seenResp     api.ExecutionResponse
	seenTestCase TestCase
}

func (f *fakeGrader) Grade(task Task, resp api.ExecutionResponse, tc TestCase) Outcome {
	f.seenTask = task
	f.seenResp = resp
	f.seenTestCase = tc
	return f.outcome
}
