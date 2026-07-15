package benchmark

import (
	"context"
	"testing"
)

func TestBenchmarkServiceRunReturnsReport(t *testing.T) {
	svc := BenchmarkService{
		Tasks: TaskCatalog{
			Tasks: []Task{{
				ID:          "a",
				Description: "demo",
				TaskFamily:  "software_engineering",
				Language:    "python",
			}},
		},
		Scaffolds: ScaffoldCatalog{
			Scaffolds: []Scaffold{{Name: "baseline"}, {Name: "tool-assisted"}},
		},
		Runner: fakeBenchmarkRunner{},
	}

	report, err := svc.Run(context.Background())
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if report.TotalTasks != 1 {
		t.Fatalf("TotalTasks = %d, want 1", report.TotalTasks)
	}
}

type fakeBenchmarkRunner struct{}

func (fakeBenchmarkRunner) RunTask(ctx context.Context, task Task, scaffold Scaffold, mode RunMode) Run {
	return Run{
		TaskID:   task.ID,
		Mode:     mode,
		Scaffold: scaffold,
		Passed:   scaffold.Name == "tool-assisted",
	}
}
