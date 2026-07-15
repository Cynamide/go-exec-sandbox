package benchmark

import (
	"context"
	"fmt"

	"gexec-sandbox/internal/config"
)

type TaskCatalog struct {
	Tasks []Task `json:"tasks"`
}

type ScaffoldCatalog struct {
	Scaffolds []Scaffold `json:"scaffolds"`
}

type BenchmarkRunGroup struct {
	Runs        []Run   `json:"runs,omitempty"`
	PassedTasks int     `json:"passed_tasks"`
	SuccessRate float64 `json:"success_rate"`
}

type BenchmarkReport struct {
	TotalTasks  int               `json:"total_tasks"`
	Baseline    BenchmarkRunGroup `json:"baseline"`
	Scaffolded  BenchmarkRunGroup `json:"scaffolded"`
	Lift        float64           `json:"lift"`
}

func BuildBenchmarkReport(tasks []Task, runs []Run) BenchmarkReport {
	baselineRuns := make([]Run, 0, len(tasks))
	scaffoldedRuns := make([]Run, 0, len(tasks))
	for _, run := range runs {
		switch run.Mode {
		case RunModeBaseline:
			baselineRuns = append(baselineRuns, run)
		case RunModeScaffolded:
			scaffoldedRuns = append(scaffoldedRuns, run)
		}
	}

	baseline := buildBenchmarkRunGroup(len(tasks), baselineRuns)
	scaffolded := buildBenchmarkRunGroup(len(tasks), scaffoldedRuns)

	return BenchmarkReport{
		TotalTasks: len(tasks),
		Baseline:   baseline,
		Scaffolded: scaffolded,
		Lift:       scaffolded.SuccessRate - baseline.SuccessRate,
	}
}

func buildBenchmarkRunGroup(totalTasks int, runs []Run) BenchmarkRunGroup {
	passedTasks := 0
	for _, run := range runs {
		if run.Passed {
			passedTasks++
		}
	}

	group := BenchmarkRunGroup{
		Runs:        runs,
		PassedTasks: passedTasks,
	}
	if totalTasks > 0 {
		group.SuccessRate = float64(passedTasks) / float64(totalTasks)
	}

	return group
}

type BenchmarkServiceAPI interface {
	Run(ctx context.Context) (BenchmarkReport, error)
}

type BenchmarkService struct {
	Tasks     TaskCatalog
	Scaffolds ScaffoldCatalog
	Client    LLMClient
	Executor  Executor
	Grader    Grader
	Config    config.Config
}

func (s BenchmarkService) Run(ctx context.Context) (BenchmarkReport, error) {
	if s.Client == nil {
		return BenchmarkReport{}, fmt.Errorf("llm client is required")
	}

	if s.Executor == nil {
		return BenchmarkReport{}, fmt.Errorf("executor is required")
	}

	if len(s.Scaffolds.Scaffolds) < 2 {
		return BenchmarkReport{}, fmt.Errorf("need at least two scaffolds")
	}

	grader := s.Grader
	if grader == nil {
		grader = DefaultGrader{}
	}

	runs := make([]Run, 0, len(s.Tasks.Tasks)*2)
	for _, task := range s.Tasks.Tasks {
		runs = append(runs,
			RunTaskWithGrader(task, s.Scaffolds.Scaffolds[0], RunModeBaseline, s.Client, s.Executor, grader, s.Config),
			RunTaskWithGrader(task, s.Scaffolds.Scaffolds[1], RunModeScaffolded, s.Client, s.Executor, grader, s.Config),
		)
	}

	return BuildBenchmarkReport(s.Tasks.Tasks, runs), nil
}
