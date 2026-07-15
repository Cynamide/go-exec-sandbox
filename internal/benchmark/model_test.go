package benchmark

import (
	"encoding/json"
	"testing"
)

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
