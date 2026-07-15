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
