package benchmark

import (
	"context"
	"fmt"
	"maps"

	"gexec-sandbox/internal/config"
)

type BenchmarkServiceAPI interface {
	Run(ctx context.Context) (BenchmarkReport, error)
}

type BenchmarkService struct {
	Tasks             TaskCatalog
	Scaffolds         ScaffoldCatalog
	Models            []ModelClient
	Client            LLMClient
	Executor          Executor
	Grader            Grader
	Config            config.Config
	DefaultModelRoles map[string]string
}

type ModelClient struct {
	ID          string
	Client      LLMClient
	HealthCheck func(context.Context) error
}

func (s BenchmarkService) HealthCheckModels(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if len(s.Models) == 0 {
		return fmt.Errorf("model health checks require configured models")
	}

	for _, model := range s.Models {
		if model.HealthCheck == nil {
			return fmt.Errorf("health check for model %q is required", model.ID)
		}
		if err := model.HealthCheck(ctx); err != nil {
			return fmt.Errorf("health check model %q: %w", model.ID, err)
		}
		if err := ctx.Err(); err != nil {
			return err
		}
	}
	return nil
}

func (s BenchmarkService) Run(ctx context.Context) (BenchmarkReport, error) {
	if err := ctx.Err(); err != nil {
		return BenchmarkReport{}, err
	}

	models := s.Models
	if len(models) == 0 && s.Client != nil {
		models = []ModelClient{{Client: s.Client}}
	}
	if len(models) == 0 {
		return BenchmarkReport{}, fmt.Errorf("llm client is required")
	}
	for _, model := range models {
		if model.Client == nil {
			return BenchmarkReport{}, fmt.Errorf("llm client for model %q is required", model.ID)
		}
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
	if len(s.Tasks.Tasks) == 0 {
		return BenchmarkReport{}, fmt.Errorf("at least one task is required")
	}

	grader := s.Grader
	if grader == nil {
		grader = DefaultGrader{}
	}

	runs := make([]Run, 0, len(models)*len(s.Tasks.Tasks)*(1+len(scaffoldVariants)))
	for _, model := range models {
		for _, task := range s.Tasks.Tasks {
			if err := ctx.Err(); err != nil {
				return BenchmarkReport{}, err
			}

			baselineRun := RunTaskWithGrader(ctx, task, *baselineScaffold, RunModeBaseline, model.Client, s.Executor, grader, s.Config)
			baselineRun.ModelID = model.ID
			runs = append(runs, baselineRun)

			if err := ctx.Err(); err != nil {
				return BenchmarkReport{}, err
			}

			for _, scaffold := range scaffoldVariants {
				run := RunTaskWithGrader(ctx, task, scaffold, RunModeScaffolded, model.Client, s.Executor, grader, s.Config)
				run.ModelID = model.ID
				runs = append(runs, run)
				if err := ctx.Err(); err != nil {
					return BenchmarkReport{}, err
				}
			}
		}
	}

	report := BuildBenchmarkReport(s.Tasks.Tasks, runs)
	report.DefaultModelRoles = maps.Clone(s.DefaultModelRoles)
	return report, nil
}
