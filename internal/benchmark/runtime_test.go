package benchmark

import (
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
