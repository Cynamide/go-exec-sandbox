package benchmark

import (
	"testing"

	"gexec-sandbox/internal/api"
	"gexec-sandbox/internal/config"
)

func TestExtractCodeHandlesMultilineFencedBlock(t *testing.T) {
	text := "Here is the solution:\n```go\npackage main\n\nfunc main() {\n\tprintln(\"hi\")\n}\n```\nThanks!"

	got := extractCode(text)
	want := "package main\n\nfunc main() {\n\tprintln(\"hi\")\n}"

	if got != want {
		t.Fatalf("extractCode() = %q, want %q", got, want)
	}
}

type stubLLMClient struct {
	code string
}

func (c stubLLMClient) GenerateCode(problem string, language string) (string, error) {
	return c.code, nil
}

func TestRunEvaluationPassesTestCaseInputToSandbox(t *testing.T) {
	t.Setenv("OLLAMA_MODEL", "test-model")

	originalRunner := runCodeInSandbox
	t.Cleanup(func() {
		runCodeInSandbox = originalRunner
	})

	var seenStdin string
	runCodeInSandbox = func(req api.ExecutionRequest, cfg config.Config) (api.ExecutionResponse, error) {
		seenStdin = req.Stdin
		return api.ExecutionResponse{Stdout: req.Stdin}, nil
	}

	report := RunEvaluation([]Problem{
		{
			Description: "read stdin and echo it",
			Language:    "go",
			TestCases: []TestCase{
				{
					Input:          "hello, world\n",
					ExpectedOutput: "hello, world",
				},
			},
		},
	}, 1, stubLLMClient{code: "```go\npackage main\n\nfunc main() {}\n```"})

	if seenStdin != "hello, world\n" {
		t.Fatalf("sandbox stdin = %q, want %q", seenStdin, "hello, world\n")
	}

	if report.PassedProblems != 1 {
		t.Fatalf("RunEvaluation() passed %d problems, want 1", report.PassedProblems)
	}
}
