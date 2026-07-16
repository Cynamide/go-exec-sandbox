package benchmark

import "sort"

type FamilySummary struct {
	TotalTasks            int     `json:"total_tasks"`
	BaselineSuccessRate   float64 `json:"baseline_success_rate"`
	ScaffoldedSuccessRate float64 `json:"scaffolded_success_rate"`
	Lift                  float64 `json:"lift"`
}

type ScaffoldSummary struct {
	TotalTasks            int     `json:"total_tasks"`
	BaselineSuccessRate   float64 `json:"baseline_success_rate"`
	ScaffoldedSuccessRate float64 `json:"scaffolded_success_rate"`
	Lift                  float64 `json:"lift"`
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
	TotalTasks            int                        `json:"total_tasks"`
	BaselineSuccessRate   float64                    `json:"baseline_success_rate"`
	ScaffoldedSuccessRate float64                    `json:"scaffolded_success_rate"`
	Lift                  float64                    `json:"lift"`
	ScaffoldedScaffold    string                     `json:"scaffolded_scaffold,omitempty"`
	ByFamily              map[string]FamilySummary   `json:"by_family"`
	ByScaffold            map[string]ScaffoldSummary `json:"by_scaffold"`
	Runs                  []Run                      `json:"runs"`
	Baseline              BenchmarkRunGroup          `json:"baseline"`
	Scaffolded            BenchmarkRunGroup          `json:"scaffolded"`
	Scaffolds             []BenchmarkScaffoldReport  `json:"scaffolds,omitempty"`
}

func BuildBenchmarkReport(tasks []Task, runs []Run) BenchmarkReport {
	report := BenchmarkReport{
		TotalTasks: len(tasks),
		ByFamily:   map[string]FamilySummary{},
		ByScaffold: map[string]ScaffoldSummary{},
		Runs:       append([]Run(nil), runs...),
	}

	baselineRuns := make([]Run, 0, len(runs))
	scaffoldRunsByName := make(map[string][]Run)
	scaffoldTasksByName := make(map[string]map[string]struct{})
	scaffoldPassedByName := make(map[string]map[string]struct{})

	baselinePassedByTask := make(map[string]bool)
	bestScaffoldName := ""
	bestScaffoldSuccess := -1.0
	bestScaffoldPassedByTask := make(map[string]bool)
	bestScaffoldRuns := []Run(nil)

	for _, run := range runs {
		switch run.Mode {
		case RunModeBaseline:
			baselineRuns = append(baselineRuns, run)
			if run.Passed {
				baselinePassedByTask[run.TaskID] = true
			}
		case RunModeScaffolded:
			scaffoldRunsByName[run.Scaffold.Name] = append(scaffoldRunsByName[run.Scaffold.Name], run)
			if scaffoldTasksByName[run.Scaffold.Name] == nil {
				scaffoldTasksByName[run.Scaffold.Name] = map[string]struct{}{}
			}
			scaffoldTasksByName[run.Scaffold.Name][run.TaskID] = struct{}{}
			if run.Passed {
				if scaffoldPassedByName[run.Scaffold.Name] == nil {
					scaffoldPassedByName[run.Scaffold.Name] = map[string]struct{}{}
				}
				scaffoldPassedByName[run.Scaffold.Name][run.TaskID] = struct{}{}
			}
		}
	}

	report.Baseline = buildBenchmarkRunGroup(len(tasks), baselineRuns)
	report.BaselineSuccessRate = report.Baseline.SuccessRate

	for _, task := range tasks {
		summary := report.ByFamily[task.TaskFamily]
		summary.TotalTasks++
		if baselinePassedByTask[task.ID] {
			summary.BaselineSuccessRate++
		}
		report.ByFamily[task.TaskFamily] = summary
	}

	scaffoldNames := make([]string, 0, len(scaffoldTasksByName))
	for name := range scaffoldTasksByName {
		scaffoldNames = append(scaffoldNames, name)
	}
	sort.Strings(scaffoldNames)

	report.Scaffolds = make([]BenchmarkScaffoldReport, 0, len(scaffoldNames))
	for _, name := range scaffoldNames {
		totalTasks := len(scaffoldTasksByName[name])
		baselinePassed := 0
		scaffoldPassed := len(scaffoldPassedByName[name])
		for taskID := range scaffoldTasksByName[name] {
			if baselinePassedByTask[taskID] {
				baselinePassed++
			}
		}

		scaffoldGroup := BenchmarkRunGroup{
			Runs:        scaffoldRunsByName[name],
			PassedTasks: scaffoldPassed,
		}
		if totalTasks > 0 {
			scaffoldGroup.SuccessRate = float64(scaffoldPassed) / float64(totalTasks)
		}

		lift := rate(scaffoldPassed, totalTasks) - rate(baselinePassed, totalTasks)
		report.ByScaffold[name] = ScaffoldSummary{
			TotalTasks:            totalTasks,
			BaselineSuccessRate:   rate(baselinePassed, totalTasks),
			ScaffoldedSuccessRate: rate(scaffoldPassed, totalTasks),
			Lift:                  lift,
		}
		scaffoldReport := BenchmarkScaffoldReport{
			Scaffold: scaffoldRunsByName[name][0].Scaffold,
			Group:    scaffoldGroup,
			Lift:     lift,
		}
		report.Scaffolds = append(report.Scaffolds, scaffoldReport)

		if bestScaffoldName == "" || scaffoldGroup.SuccessRate > bestScaffoldSuccess || (scaffoldGroup.SuccessRate == bestScaffoldSuccess && name < bestScaffoldName) {
			bestScaffoldName = name
			bestScaffoldSuccess = scaffoldGroup.SuccessRate
			bestScaffoldPassedByTask = make(map[string]bool, len(scaffoldPassedByName[name]))
			for taskID := range scaffoldPassedByName[name] {
				bestScaffoldPassedByTask[taskID] = true
			}
			bestScaffoldRuns = scaffoldRunsByName[name]
		}
	}

	report.ScaffoldedScaffold = bestScaffoldName
	report.Scaffolded = buildBenchmarkRunGroup(len(tasks), bestScaffoldRuns)
	report.ScaffoldedSuccessRate = report.Scaffolded.SuccessRate
	report.Lift = report.ScaffoldedSuccessRate - report.BaselineSuccessRate

	for family, summary := range report.ByFamily {
		total := float64(summary.TotalTasks)
		if total > 0 {
			baselineRate := summary.BaselineSuccessRate / total
			scaffoldPassed := 0
			for _, task := range tasks {
				if task.TaskFamily == family && bestScaffoldPassedByTask[task.ID] {
					scaffoldPassed++
				}
			}
			scaffoldRate := float64(scaffoldPassed) / total
			summary.BaselineSuccessRate = baselineRate
			summary.ScaffoldedSuccessRate = scaffoldRate
			summary.Lift = summary.ScaffoldedSuccessRate - summary.BaselineSuccessRate
		}
		report.ByFamily[family] = summary
	}

	return report
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

func rate(passed int, total int) float64 {
	if total == 0 {
		return 0
	}
	return float64(passed) / float64(total)
}
