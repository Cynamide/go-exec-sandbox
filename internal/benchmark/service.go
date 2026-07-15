package benchmark

import (
	"context"
	"fmt"
	"sort"

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

type BenchmarkScaffoldReport struct {
	Scaffold Scaffold          `json:"scaffold"`
	Group    BenchmarkRunGroup `json:"group"`
	Lift     float64           `json:"lift"`
}

type BenchmarkReport struct {
	TotalTasks int                       `json:"total_tasks"`
	Baseline   BenchmarkRunGroup         `json:"baseline"`
	Scaffolded BenchmarkRunGroup         `json:"scaffolded"`
	Scaffolds  []BenchmarkScaffoldReport `json:"scaffolds,omitempty"`
	Lift       float64                   `json:"lift"`
}

func BuildBenchmarkReport(tasks []Task, runs []Run) BenchmarkReport {
	baselineRuns := make([]Run, 0, len(tasks))
	scaffoldedRuns := make([]Run, 0, len(tasks))
	scaffoldRunsByName := make(map[string][]Run)
	for _, run := range runs {
		switch run.Mode {
		case RunModeBaseline:
			baselineRuns = append(baselineRuns, run)
		case RunModeScaffolded:
			scaffoldedRuns = append(scaffoldedRuns, run)
			scaffoldRunsByName[run.Scaffold.Name] = append(scaffoldRunsByName[run.Scaffold.Name], run)
		}
	}

	baseline := buildBenchmarkRunGroup(len(tasks), baselineRuns)
	scaffolded := buildBenchmarkRunGroup(len(tasks), scaffoldedRuns)
	scaffoldNames := make([]string, 0, len(scaffoldRunsByName))
	for name := range scaffoldRunsByName {
		scaffoldNames = append(scaffoldNames, name)
	}
	sort.Strings(scaffoldNames)

	scaffolds := make([]BenchmarkScaffoldReport, 0, len(scaffoldNames))
	for _, name := range scaffoldNames {
		group := buildBenchmarkRunGroup(len(tasks), scaffoldRunsByName[name])
		scaffolds = append(scaffolds, BenchmarkScaffoldReport{
			Scaffold: scaffoldRunsByName[name][0].Scaffold,
			Group:    group,
			Lift:     group.SuccessRate - baseline.SuccessRate,
		})
	}

	return BenchmarkReport{
		TotalTasks: len(tasks),
		Baseline:   baseline,
		Scaffolded: scaffolded,
		Scaffolds:  scaffolds,
		Lift:       scaffolded.SuccessRate - baseline.SuccessRate,
	}
}

func buildBenchmarkRunGroup(totalTasks int, runs []Run) BenchmarkRunGroup {
	passedTaskIDs := map[string]struct{}{}
	for _, run := range runs {
		if run.Passed {
			passedTaskIDs[run.TaskID] = struct{}{}
		}
	}

	group := BenchmarkRunGroup{
		Runs:        runs,
		PassedTasks: len(passedTaskIDs),
	}
	if totalTasks > 0 {
		group.SuccessRate = float64(group.PassedTasks) / float64(totalTasks)
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

		baselineRun := RunTaskWithGrader(task, *baselineScaffold, RunModeBaseline, s.Client, s.Executor, grader, s.Config)
		runs = append(runs, baselineRun)

		if err := ctx.Err(); err != nil {
			return BenchmarkReport{}, err
		}

		for _, scaffold := range scaffoldVariants {
			runs = append(runs, RunTaskWithGrader(task, scaffold, RunModeScaffolded, s.Client, s.Executor, grader, s.Config))
		}
	}

	return BuildBenchmarkReport(s.Tasks.Tasks, runs), nil
}
