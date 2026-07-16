package benchmark

import (
	"context"
	"fmt"

	"gexec-sandbox/internal/config"
)

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
	if err := ctx.Err(); err != nil {
		return BenchmarkReport{}, err
	}

	if s.Client == nil {
		return BenchmarkReport{}, fmt.Errorf("llm client is required")
	}

	if s.Executor == nil {
		return BenchmarkReport{}, fmt.Errorf("executor is required")
	}

	var baselineScaffold *Scaffold
	var scaffoldVariants []Scaffold
	for i := range s.Scaffolds.Scaffolds {
		scaffold := s.Scaffolds.Scaffolds[i]
		if scaffold.Baseline {
			if baselineScaffold != nil {
				return BenchmarkReport{}, fmt.Errorf("need exactly one baseline scaffold")
			}
			baselineScaffold = &s.Scaffolds.Scaffolds[i]
			continue
		}
		scaffoldVariants = append(scaffoldVariants, scaffold)
	}

	if baselineScaffold == nil {
		return BenchmarkReport{}, fmt.Errorf("baseline scaffold is required")
	}

	if len(scaffoldVariants) == 0 {
		return BenchmarkReport{}, fmt.Errorf("need at least one scaffolded scaffold")
	}

	grader := s.Grader
	if grader == nil {
		grader = DefaultGrader{}
	}

	runs := make([]Run, 0, len(s.Tasks.Tasks)*(1+len(scaffoldVariants)))
	for _, task := range s.Tasks.Tasks {
		if err := ctx.Err(); err != nil {
			return BenchmarkReport{}, err
		}

		baselineRun := RunTaskWithGrader(ctx, task, *baselineScaffold, RunModeBaseline, s.Client, s.Executor, grader, s.Config)
		runs = append(runs, baselineRun)

		if err := ctx.Err(); err != nil {
			return BenchmarkReport{}, err
		}

		for _, scaffold := range scaffoldVariants {
			runs = append(runs, RunTaskWithGrader(ctx, task, scaffold, RunModeScaffolded, s.Client, s.Executor, grader, s.Config))
			if err := ctx.Err(); err != nil {
				return BenchmarkReport{}, err
			}
		}
	}

	return BuildBenchmarkReport(s.Tasks.Tasks, runs), nil
}
