package benchmark

import (
	"context"
	"fmt"
)

type TaskCatalog struct {
	Tasks []Task `json:"tasks"`
}

type ScaffoldCatalog struct {
	Scaffolds []Scaffold `json:"scaffolds"`
}

type BenchmarkReport struct {
	TotalTasks int   `json:"total_tasks"`
	Runs       []Run `json:"runs,omitempty"`
}

func BuildBenchmarkReport(tasks []Task, runs []Run) BenchmarkReport {
	return BenchmarkReport{
		TotalTasks: len(tasks),
		Runs:       runs,
	}
}

type BenchmarkRunner interface {
	RunTask(ctx context.Context, task Task, scaffold Scaffold, mode RunMode) Run
}

type BenchmarkServiceAPI interface {
	Run(ctx context.Context) (BenchmarkReport, error)
}

type BenchmarkService struct {
	Tasks     TaskCatalog
	Scaffolds ScaffoldCatalog
	Runner    BenchmarkRunner
}

func (s BenchmarkService) Run(ctx context.Context) (BenchmarkReport, error) {
	if s.Runner == nil {
		return BenchmarkReport{}, fmt.Errorf("benchmark runner is required")
	}

	if len(s.Scaffolds.Scaffolds) < 2 {
		return BenchmarkReport{}, fmt.Errorf("need at least two scaffolds")
	}

	runs := make([]Run, 0, len(s.Tasks.Tasks)*2)
	for _, task := range s.Tasks.Tasks {
		runs = append(runs,
			s.Runner.RunTask(ctx, task, s.Scaffolds.Scaffolds[0], RunModeBaseline),
			s.Runner.RunTask(ctx, task, s.Scaffolds.Scaffolds[1], RunModeScaffolded),
		)
	}

	return BuildBenchmarkReport(s.Tasks.Tasks, runs), nil
}
