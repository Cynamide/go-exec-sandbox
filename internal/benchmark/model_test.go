package benchmark

import "testing"

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

func TestProblemCompatibilityAliasSupportsCurrentFixtureShape(t *testing.T) {
	problem := Problem{
		ID:          "sum-two-numbers",
		Title:       "Sum two numbers",
		Description: "Read two integers and print their sum.",
		Language:    "go",
		TestCases: []TestCase{
			{
				Input:          "1 2\n",
				ExpectedOutput: "3\n",
			},
		},
	}

	if problem.TaskFamily != "" {
		t.Fatalf("TaskFamily = %q, want empty string", problem.TaskFamily)
	}
}
