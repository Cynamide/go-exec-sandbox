package benchmark

import (
	"context"
	"errors"
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

func TestBenchmarkServiceRunReturnsErrorWhenContextCanceledBeforeWorkBegins(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	client := &countingBenchmarkServiceLLMClient{}
	executor := &countingBenchmarkServiceExecutor{}
	svc := BenchmarkService{
		Tasks: TaskCatalog{
			Tasks: []Task{{
				ID:          "a",
				Description: "demo",
				TaskFamily:  "software_engineering",
				Language:    "python",
				TestCases: []TestCase{{
					ExpectedOutput: "ok",
				}},
			}},
		},
		Scaffolds: ScaffoldCatalog{
			Scaffolds: []Scaffold{
				{Name: "baseline"},
				{Name: "tool-assisted", PromptPrefix: "tool: "},
			},
		},
		Client:   client,
		Executor: executor,
		Grader:   DefaultGrader{},
	}

	_, err := svc.Run(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Run() error = %v, want %v", err, context.Canceled)
	}

	if client.calls != 0 {
		t.Fatalf("client calls = %d, want 0", client.calls)
	}

	if executor.calls != 0 {
		t.Fatalf("executor calls = %d, want 0", executor.calls)
	}
}

func TestBenchmarkServiceRunReturnsErrorWhenContextCanceledBetweenTasks(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	client := &countingBenchmarkServiceLLMClient{
		codeByPrompt: map[string]string{
			"task-1":       "print('one')",
			"tool: task-1": "print('one')",
			"task-2":       "print('two')",
			"tool: task-2": "print('two')",
		},
		cancelAfterCalls: 2,
		cancel:           cancel,
	}
	executor := &countingBenchmarkServiceExecutor{
		responseBySource: map[string]api.ExecutionResponse{
			"print('one')": {Stdout: "ok"},
			"print('two')": {Stdout: "ok"},
		},
	}
	svc := BenchmarkService{
		Tasks: TaskCatalog{
			Tasks: []Task{
				{
					ID:          "task-1",
					Description: "task-1",
					TaskFamily:  "software_engineering",
					Language:    "python",
					TestCases: []TestCase{{
						ExpectedOutput: "ok",
					}},
				},
				{
					ID:          "task-2",
					Description: "task-2",
					TaskFamily:  "software_engineering",
					Language:    "python",
					TestCases: []TestCase{{
						ExpectedOutput: "ok",
					}},
				},
			},
		},
		Scaffolds: ScaffoldCatalog{
			Scaffolds: []Scaffold{
				{Name: "baseline"},
				{Name: "tool-assisted", PromptPrefix: "tool: "},
			},
		},
		Client:   client,
		Executor: executor,
		Grader:   DefaultGrader{},
	}

	_, err := svc.Run(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Run() error = %v, want %v", err, context.Canceled)
	}

	if client.calls != 2 {
		t.Fatalf("client calls = %d, want 2", client.calls)
	}

	if executor.calls != 2 {
		t.Fatalf("executor calls = %d, want 2", executor.calls)
	}
}

func TestBenchmarkServiceRunReturnsErrorWhenClientMissing(t *testing.T) {
	svc := BenchmarkService{
		Scaffolds: ScaffoldCatalog{
			Scaffolds: []Scaffold{{Name: "baseline"}, {Name: "tool-assisted"}},
		},
		Executor: &countingBenchmarkServiceExecutor{},
	}

	_, err := svc.Run(context.Background())
	if err == nil || err.Error() != "llm client is required" {
		t.Fatalf("Run() error = %v, want llm client is required", err)
	}
}

func TestBenchmarkServiceRunReturnsErrorWhenExecutorMissing(t *testing.T) {
	svc := BenchmarkService{
		Scaffolds: ScaffoldCatalog{
			Scaffolds: []Scaffold{{Name: "baseline"}, {Name: "tool-assisted"}},
		},
		Client: &countingBenchmarkServiceLLMClient{},
	}

	_, err := svc.Run(context.Background())
	if err == nil || err.Error() != "executor is required" {
		t.Fatalf("Run() error = %v, want executor is required", err)
	}
}

func TestBenchmarkServiceRunReturnsErrorWhenScaffoldsInsufficient(t *testing.T) {
	svc := BenchmarkService{
		Client:   &countingBenchmarkServiceLLMClient{},
		Executor: &countingBenchmarkServiceExecutor{},
		Scaffolds: ScaffoldCatalog{
			Scaffolds: []Scaffold{{Name: "baseline"}},
		},
	}

	_, err := svc.Run(context.Background())
	if err == nil || err.Error() != "need at least two scaffolds" {
		t.Fatalf("Run() error = %v, want need at least two scaffolds", err)
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

type countingBenchmarkServiceLLMClient struct {
	codeByPrompt      map[string]string
	calls             int
	cancelAfterCalls  int
	cancel            context.CancelFunc
}

func (f *countingBenchmarkServiceLLMClient) GenerateCode(problem string, language string) (string, error) {
	f.calls++
	if f.cancel != nil && f.cancelAfterCalls > 0 && f.calls == f.cancelAfterCalls {
		f.cancel()
	}

	return f.codeByPrompt[problem], nil
}

type countingBenchmarkServiceExecutor struct {
	responseBySource map[string]api.ExecutionResponse
	calls            int
}

func (f *countingBenchmarkServiceExecutor) Execute(req api.ExecutionRequest, cfg config.Config) (api.ExecutionResponse, error) {
	f.calls++
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
