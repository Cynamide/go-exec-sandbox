package benchmark

import (
	"encoding/json"
	"testing"
)

func TestBenchmarkReportCarriesFamilyAndScaffoldBreakdowns(t *testing.T) {
	report := BenchmarkReport{
		TotalTasks: 2,
		ByFamily: map[string]FamilySummary{
			"software_engineering": {
				TotalTasks:            1,
				BaselineSuccessRate:   0.0,
				ScaffoldedSuccessRate: 1.0,
				Lift:                  1.0,
			},
		},
		ByScaffold: map[string]ScaffoldSummary{
			"tool-assisted": {
				TotalTasks:            2,
				BaselineSuccessRate:   0.5,
				ScaffoldedSuccessRate: 1.0,
				Lift:                  0.5,
			},
		},
		Runs: []Run{{TaskID: "task-1", Mode: RunModeScaffolded, Scaffold: Scaffold{Name: "tool-assisted"}, Passed: true}},
	}

	if report.ByFamily["software_engineering"].Lift != 1.0 {
		t.Fatalf("family lift = %v, want 1.0", report.ByFamily["software_engineering"].Lift)
	}

	if report.ByScaffold["tool-assisted"].Lift != 0.5 {
		t.Fatalf("scaffold lift = %v, want 0.5", report.ByScaffold["tool-assisted"].Lift)
	}
}

func TestBenchmarkReportJSONIncludesLegacyAndSummaryFields(t *testing.T) {
	report := BenchmarkReport{
		TotalTasks:            1,
		BaselineSuccessRate:   0.25,
		ScaffoldedSuccessRate: 1.0,
		Lift:                  0.75,
		ScaffoldedScaffold:    "tool-assisted",
		ByFamily: map[string]FamilySummary{
			"software_engineering": {
				TotalTasks:            1,
				BaselineSuccessRate:   0.25,
				ScaffoldedSuccessRate: 1.0,
				Lift:                  0.75,
			},
		},
		ByScaffold: map[string]ScaffoldSummary{
			"tool-assisted": {
				TotalTasks:            1,
				BaselineSuccessRate:   0.25,
				ScaffoldedSuccessRate: 1.0,
				Lift:                  0.75,
			},
		},
		Runs:       []Run{{TaskID: "task-1", Mode: RunModeBaseline, Scaffold: Scaffold{Name: "baseline"}, Passed: false}},
		Baseline:   BenchmarkRunGroup{Runs: []Run{{TaskID: "task-1", Mode: RunModeBaseline, Scaffold: Scaffold{Name: "baseline"}, Passed: false}}, PassedTasks: 0, SuccessRate: 0},
		Scaffolded: BenchmarkRunGroup{Runs: []Run{{TaskID: "task-1", Mode: RunModeScaffolded, Scaffold: Scaffold{Name: "tool-assisted"}, Passed: true}}, PassedTasks: 1, SuccessRate: 1},
		Scaffolds: []BenchmarkScaffoldReport{{
			Scaffold: Scaffold{Name: "tool-assisted"},
			Group:    BenchmarkRunGroup{Runs: []Run{{TaskID: "task-1", Mode: RunModeScaffolded, Scaffold: Scaffold{Name: "tool-assisted"}, Passed: true}}, PassedTasks: 1, SuccessRate: 1},
			Lift:     1.0,
		}},
	}

	raw, err := json.Marshal(report)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if got := decoded["scaffolded_scaffold"]; got != "tool-assisted" {
		t.Fatalf("scaffolded_scaffold = %v, want tool-assisted", got)
	}

	if _, ok := decoded["baseline"].(map[string]any); !ok {
		t.Fatalf("baseline JSON = %T, want object", decoded["baseline"])
	}

	if _, ok := decoded["scaffolded"].(map[string]any); !ok {
		t.Fatalf("scaffolded JSON = %T, want object", decoded["scaffolded"])
	}

	if _, ok := decoded["by_family"].(map[string]any); !ok {
		t.Fatalf("by_family JSON = %T, want object", decoded["by_family"])
	}

	if _, ok := decoded["by_scaffold"].(map[string]any); !ok {
		t.Fatalf("by_scaffold JSON = %T, want object", decoded["by_scaffold"])
	}

	if _, ok := decoded["runs"].([]any); !ok {
		t.Fatalf("runs JSON = %T, want array", decoded["runs"])
	}
}

func TestBuildBenchmarkReportSelectsOneScaffoldVariant(t *testing.T) {
	tasks := []Task{
		{ID: "task-1", TaskFamily: "software_engineering"},
		{ID: "task-2", TaskFamily: "software_engineering"},
	}
	runs := []Run{
		{TaskID: "task-1", Mode: RunModeBaseline, Scaffold: Scaffold{Name: "baseline"}, Passed: false},
		{TaskID: "task-2", Mode: RunModeBaseline, Scaffold: Scaffold{Name: "baseline"}, Passed: true},
		{TaskID: "task-1", Mode: RunModeScaffolded, Scaffold: Scaffold{Name: "alpha"}, Passed: true},
		{TaskID: "task-2", Mode: RunModeScaffolded, Scaffold: Scaffold{Name: "beta"}, Passed: false},
	}

	report := BuildBenchmarkReport(tasks, runs)

	family := report.ByFamily["software_engineering"]
	if family.TotalTasks != 2 {
		t.Fatalf("family TotalTasks = %d, want 2", family.TotalTasks)
	}
	if family.BaselineSuccessRate != 0.5 {
		t.Fatalf("family BaselineSuccessRate = %v, want 0.5", family.BaselineSuccessRate)
	}
	if family.ScaffoldedSuccessRate != 0.5 {
		t.Fatalf("family ScaffoldedSuccessRate = %v, want 0.5", family.ScaffoldedSuccessRate)
	}
	if family.Lift != 0.0 {
		t.Fatalf("family Lift = %v, want 0.0", family.Lift)
	}

	alpha := report.ByScaffold["alpha"]
	if alpha.TotalTasks != 1 {
		t.Fatalf("alpha TotalTasks = %d, want 1", alpha.TotalTasks)
	}
	if alpha.BaselineSuccessRate != 0.0 {
		t.Fatalf("alpha BaselineSuccessRate = %v, want 0.0", alpha.BaselineSuccessRate)
	}
	if alpha.ScaffoldedSuccessRate != 1.0 {
		t.Fatalf("alpha ScaffoldedSuccessRate = %v, want 1.0", alpha.ScaffoldedSuccessRate)
	}
	if alpha.Lift != 1.0 {
		t.Fatalf("alpha Lift = %v, want 1.0", alpha.Lift)
	}

	if report.BaselineSuccessRate != 0.5 {
		t.Fatalf("BaselineSuccessRate = %v, want 0.5", report.BaselineSuccessRate)
	}
	if report.ScaffoldedScaffold != "alpha" {
		t.Fatalf("ScaffoldedScaffold = %q, want alpha", report.ScaffoldedScaffold)
	}
	if report.ScaffoldedSuccessRate != 0.5 {
		t.Fatalf("ScaffoldedSuccessRate = %v, want 0.5", report.ScaffoldedSuccessRate)
	}
	if report.Lift != 0.0 {
		t.Fatalf("Lift = %v, want 0.0", report.Lift)
	}
	if report.Scaffolded.PassedTasks != 1 {
		t.Fatalf("Scaffolded.PassedTasks = %d, want 1", report.Scaffolded.PassedTasks)
	}
	if len(report.Scaffolded.Runs) != 1 {
		t.Fatalf("len(Scaffolded.Runs) = %d, want 1", len(report.Scaffolded.Runs))
	}
	if len(report.Scaffolds) != 2 {
		t.Fatalf("len(Scaffolds) = %d, want 2", len(report.Scaffolds))
	}
}

func TestBuildBenchmarkReportComputesFamilyAndScaffoldSummaries(t *testing.T) {
	tasks := []Task{
		{ID: "task-1", TaskFamily: "software_engineering"},
		{ID: "task-2", TaskFamily: "software_engineering"},
	}
	runs := []Run{
		{TaskID: "task-1", Mode: RunModeBaseline, Scaffold: Scaffold{Name: "baseline"}, Passed: false},
		{TaskID: "task-2", Mode: RunModeBaseline, Scaffold: Scaffold{Name: "baseline"}, Passed: true},
		{TaskID: "task-1", Mode: RunModeScaffolded, Scaffold: Scaffold{Name: "tool-assisted"}, Passed: true},
		{TaskID: "task-2", Mode: RunModeScaffolded, Scaffold: Scaffold{Name: "tool-assisted"}, Passed: true},
	}

	report := BuildBenchmarkReport(tasks, runs)

	scaffold := report.ByScaffold["tool-assisted"]
	if scaffold.TotalTasks != 2 {
		t.Fatalf("scaffold TotalTasks = %d, want 2", scaffold.TotalTasks)
	}
	if scaffold.BaselineSuccessRate != 0.5 {
		t.Fatalf("scaffold BaselineSuccessRate = %v, want 0.5", scaffold.BaselineSuccessRate)
	}
	if scaffold.ScaffoldedSuccessRate != 1.0 {
		t.Fatalf("scaffold ScaffoldedSuccessRate = %v, want 1.0", scaffold.ScaffoldedSuccessRate)
	}
	if scaffold.Lift != 0.5 {
		t.Fatalf("scaffold Lift = %v, want 0.5", scaffold.Lift)
	}
}
