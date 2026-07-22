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

type ModelSummary struct {
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
	TotalModelTasks       int                        `json:"total_model_tasks,omitempty"`
	BaselineSuccessRate   float64                    `json:"baseline_success_rate"`
	ScaffoldedSuccessRate float64                    `json:"scaffolded_success_rate"`
	Lift                  float64                    `json:"lift"`
	ScaffoldedScaffold    string                     `json:"scaffolded_scaffold,omitempty"`
	ByFamily              map[string]FamilySummary   `json:"by_family"`
	ByScaffold            map[string]ScaffoldSummary `json:"by_scaffold"`
	ByModel               map[string]ModelSummary    `json:"by_model,omitempty"`
	Runs                  []Run                      `json:"runs"`
	Baseline              BenchmarkRunGroup          `json:"baseline"`
	Scaffolded            BenchmarkRunGroup          `json:"scaffolded"`
	Scaffolds             []BenchmarkScaffoldReport  `json:"scaffolds,omitempty"`
}

func BuildBenchmarkReport(tasks []Task, runs []Run) BenchmarkReport {
	return buildBenchmarkReport(tasks, runs, true)
}

func buildBenchmarkReport(tasks []Task, runs []Run, includeModels bool) BenchmarkReport {
	modelIDs := reportModelIDs(runs)
	totalModelTasks := len(tasks) * len(modelIDs)
	report := BenchmarkReport{
		TotalTasks:      len(tasks),
		TotalModelTasks: totalModelTasks,
		ByFamily:        map[string]FamilySummary{},
		ByScaffold:      map[string]ScaffoldSummary{},
		Runs:            append([]Run(nil), runs...),
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
				baselinePassedByTask[runIdentity(run)] = true
			}
		case RunModeScaffolded:
			scaffoldRunsByName[run.Scaffold.Name] = append(scaffoldRunsByName[run.Scaffold.Name], run)
			if scaffoldTasksByName[run.Scaffold.Name] == nil {
				scaffoldTasksByName[run.Scaffold.Name] = map[string]struct{}{}
			}
			scaffoldTasksByName[run.Scaffold.Name][runIdentity(run)] = struct{}{}
			if run.Passed {
				if scaffoldPassedByName[run.Scaffold.Name] == nil {
					scaffoldPassedByName[run.Scaffold.Name] = map[string]struct{}{}
				}
				scaffoldPassedByName[run.Scaffold.Name][runIdentity(run)] = struct{}{}
			}
		}
	}

	report.Baseline = buildBenchmarkRunGroup(totalModelTasks, baselineRuns)
	report.BaselineSuccessRate = report.Baseline.SuccessRate

	for _, task := range tasks {
		summary := report.ByFamily[task.TaskFamily]
		summary.TotalTasks += len(modelIDs)
		for _, modelID := range modelIDs {
			if baselinePassedByTask[modelTaskIdentity(modelID, task.ID)] {
				summary.BaselineSuccessRate++
			}
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
	report.Scaffolded = buildBenchmarkRunGroup(totalModelTasks, bestScaffoldRuns)
	report.ScaffoldedSuccessRate = report.Scaffolded.SuccessRate
	report.Lift = report.ScaffoldedSuccessRate - report.BaselineSuccessRate

	for family, summary := range report.ByFamily {
		total := float64(summary.TotalTasks)
		if total > 0 {
			baselineRate := summary.BaselineSuccessRate / total
			scaffoldPassed := 0
			for _, task := range tasks {
				for _, modelID := range modelIDs {
					if task.TaskFamily == family && bestScaffoldPassedByTask[modelTaskIdentity(modelID, task.ID)] {
						scaffoldPassed++
					}
				}
			}
			scaffoldRate := float64(scaffoldPassed) / total
			summary.BaselineSuccessRate = baselineRate
			summary.ScaffoldedSuccessRate = scaffoldRate
			summary.Lift = summary.ScaffoldedSuccessRate - summary.BaselineSuccessRate
		}
		report.ByFamily[family] = summary
	}

	if includeModels {
		report.ByModel = buildModelSummaries(tasks, runs)
	}

	return report
}

func buildBenchmarkRunGroup(totalTasks int, runs []Run) BenchmarkRunGroup {
	passedTaskIDs := map[string]struct{}{}
	for _, run := range runs {
		if run.Passed {
			passedTaskIDs[runIdentity(run)] = struct{}{}
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

func buildModelSummaries(tasks []Task, runs []Run) map[string]ModelSummary {
	runsByModel := make(map[string][]Run)
	for _, run := range runs {
		if run.ModelID != "" {
			runsByModel[run.ModelID] = append(runsByModel[run.ModelID], run)
		}
	}
	if len(runsByModel) == 0 {
		return nil
	}

	summaries := make(map[string]ModelSummary, len(runsByModel))
	for modelID, modelRuns := range runsByModel {
		modelReport := buildBenchmarkReport(tasks, modelRuns, false)
		summaries[modelID] = ModelSummary{
			TotalTasks:            len(tasks),
			BaselineSuccessRate:   modelReport.BaselineSuccessRate,
			ScaffoldedSuccessRate: modelReport.ScaffoldedSuccessRate,
			Lift:                  modelReport.Lift,
		}
	}
	return summaries
}

func reportModelIDs(runs []Run) []string {
	models := map[string]struct{}{}
	for _, run := range runs {
		models[run.ModelID] = struct{}{}
	}
	if len(models) == 0 {
		return []string{""}
	}
	ids := make([]string, 0, len(models))
	for modelID := range models {
		ids = append(ids, modelID)
	}
	sort.Strings(ids)
	return ids
}

func runIdentity(run Run) string {
	return modelTaskIdentity(run.ModelID, run.TaskID)
}

func modelTaskIdentity(modelID string, taskID string) string {
	return modelID + "\x00" + taskID
}

func rate(passed int, total int) float64 {
	if total == 0 {
		return 0
	}
	return float64(passed) / float64(total)
}
