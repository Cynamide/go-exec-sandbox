package benchmark

import (
	"context"
	"testing"

	"gexec-sandbox/internal/api"
	"gexec-sandbox/internal/config"
)

func TestBenchmarkServiceRunReturnsReport(t *testing.T) {
	svc := BenchmarkService{
		Tasks: TaskCatalog{
			Tasks: []Task{{
				ID:          "a",
				Description: "demo",
				TaskFamily:  "software_engineering",
				Language:    "python",
				TestCases: []TestCase{{
					ExpectedOutput: "scaffolded",
				}},
			}},
		},
		Scaffolds: ScaffoldCatalog{
			Scaffolds: []Scaffold{
				{Name: "baseline"},
				{Name: "tool-assisted", PromptPrefix: "tool: "},
			},
		},
		Client: benchmarkServiceLLMClient{
			codeByPrompt: map[string]string{
				"demo":       "print('baseline')",
				"tool: demo": "print('scaffolded')",
			},
		},
		Executor: benchmarkServiceExecutor{
			responseBySource: map[string]api.ExecutionResponse{
				"print('baseline')":   {Stdout: "baseline"},
				"print('scaffolded')": {Stdout: "scaffolded"},
			},
		},
		Grader: DefaultGrader{},
		Config: config.Config{DefaultTimeoutMS: 1234},
	}

	report, err := svc.Run(context.Background())
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if report.TotalTasks != 1 {
		t.Fatalf("TotalTasks = %d, want 1", report.TotalTasks)
	}

	if len(report.Baseline.Runs) != 1 {
		t.Fatalf("len(Baseline.Runs) = %d, want 1", len(report.Baseline.Runs))
	}

	if len(report.Scaffolded.Runs) != 1 {
		t.Fatalf("len(Scaffolded.Runs) = %d, want 1", len(report.Scaffolded.Runs))
	}

	if report.Baseline.Runs[0].Mode != RunModeBaseline {
		t.Fatalf("Baseline mode = %q, want %q", report.Baseline.Runs[0].Mode, RunModeBaseline)
	}

	if report.Scaffolded.Runs[0].Mode != RunModeScaffolded {
		t.Fatalf("Scaffolded mode = %q, want %q", report.Scaffolded.Runs[0].Mode, RunModeScaffolded)
	}

	if report.Baseline.SuccessRate != 0 {
		t.Fatalf("Baseline.SuccessRate = %v, want 0", report.Baseline.SuccessRate)
	}

	if report.Scaffolded.SuccessRate != 1 {
		t.Fatalf("Scaffolded.SuccessRate = %v, want 1", report.Scaffolded.SuccessRate)
	}

	if report.Lift != 1 {
		t.Fatalf("Lift = %v, want 1", report.Lift)
	}
}

type benchmarkServiceLLMClient struct {
	codeByPrompt map[string]string
}

func (f benchmarkServiceLLMClient) GenerateCode(problem string, language string) (string, error) {
	return f.codeByPrompt[problem], nil
}

type benchmarkServiceExecutor struct {
	responseBySource map[string]api.ExecutionResponse
}

func (f benchmarkServiceExecutor) Execute(req api.ExecutionRequest, cfg config.Config) (api.ExecutionResponse, error) {
	return f.responseBySource[req.SourceCode], nil
}

func TestBuildBenchmarkReportSeparatesBaselineAndScaffoldedRuns(t *testing.T) {
	report := BuildBenchmarkReport([]Task{{ID: "task-1"}}, []Run{
		{TaskID: "task-1", Mode: RunModeBaseline, Passed: false},
		{TaskID: "task-1", Mode: RunModeScaffolded, Passed: true},
	})

	if len(report.Baseline.Runs) != 1 {
		t.Fatalf("len(Baseline.Runs) = %d, want 1", len(report.Baseline.Runs))
	}

	if len(report.Scaffolded.Runs) != 1 {
		t.Fatalf("len(Scaffolded.Runs) = %d, want 1", len(report.Scaffolded.Runs))
	}

	if report.Baseline.PassedTasks != 0 {
		t.Fatalf("Baseline.PassedTasks = %d, want 0", report.Baseline.PassedTasks)
	}

	if report.Scaffolded.PassedTasks != 1 {
		t.Fatalf("Scaffolded.PassedTasks = %d, want 1", report.Scaffolded.PassedTasks)
	}

	if report.Lift != 1 {
		t.Fatalf("Lift = %v, want 1", report.Lift)
	}
}
