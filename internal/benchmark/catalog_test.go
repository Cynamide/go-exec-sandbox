package benchmark

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadTaskCatalogReturnsFamilies(t *testing.T) {
	catalog, err := LoadTaskCatalog("../../data/tasks.json")
	if err != nil {
		t.Fatalf("LoadTaskCatalog() error = %v", err)
	}

	if len(catalog.Tasks) == 0 {
		t.Fatalf("catalog.Tasks is empty")
	}

	gotFamilies := make(map[string]struct{}, len(catalog.Tasks))
	for _, task := range catalog.Tasks {
		gotFamilies[task.TaskFamily] = struct{}{}
	}

	wantFamilies := []string{
		"software_engineering",
		"observability",
		"finance_reporting",
		"operations",
		"customer_support",
	}
	for _, family := range wantFamilies {
		if _, ok := gotFamilies[family]; !ok {
			t.Fatalf("catalog.Tasks missing task family %q", family)
		}
	}
}

func TestTaskCatalogContainsMultipleTaskFamilies(t *testing.T) {
	catalog, err := LoadTaskCatalog("../../data/tasks.json")
	if err != nil {
		t.Fatalf("LoadTaskCatalog() error = %v", err)
	}

	families := map[string]bool{}
	for _, task := range catalog.Tasks {
		if task.TaskFamily == "" {
			t.Fatalf("task %q missing task family", task.ID)
		}
		if task.Title == "" {
			t.Fatalf("task %q missing title", task.ID)
		}
		if task.Description == "" {
			t.Fatalf("task %q missing description", task.ID)
		}
		if len(task.TestCases) == 0 {
			t.Fatalf("task %q missing test cases", task.ID)
		}
		families[task.TaskFamily] = true
	}

	if len(families) < 5 {
		t.Fatalf("families = %v, want at least 5", families)
	}
}

func TestLoadTaskCatalogRejectsTaskWithoutTestCases(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tasks.json")
	if err := os.WriteFile(path, []byte(`{"tasks":[{"id":"artifact-task","title":"Artifact Task","description":"desc","task_family":"customer_support","language":"python"}]}`), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	if _, err := LoadTaskCatalog(path); err == nil {
		t.Fatal("LoadTaskCatalog() error = nil, want invalid task catalog error")
	}
}

func TestLoadTaskCatalogRejectsUnknownFields(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tasks.json")
	if err := os.WriteFile(path, []byte(`{"tasks":[{"id":"task-1","title":"Task 1","description":"desc","task_family":"customer_support","language":"python","test_cases":[{"input":"","expected_output":"ok"}],"artifact_expectation":{"type":"markdown_report"}}]}`), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	if _, err := LoadTaskCatalog(path); err == nil {
		t.Fatal("LoadTaskCatalog() error = nil, want unknown field error")
	}
}

func TestLoadScaffoldCatalogReturnsScaffolds(t *testing.T) {
	catalog, err := LoadScaffoldCatalog("../../data/scaffolds.json")
	if err != nil {
		t.Fatalf("LoadScaffoldCatalog() error = %v", err)
	}

	if len(catalog.Scaffolds) == 0 {
		t.Fatalf("catalog.Scaffolds is empty")
	}

	gotScaffolds := make(map[string]struct{}, len(catalog.Scaffolds))
	for _, scaffold := range catalog.Scaffolds {
		gotScaffolds[scaffold.Name] = struct{}{}
	}

	wantScaffolds := []string{
		"baseline",
		"tool-assisted",
		"retrieval-assisted",
	}
	for _, name := range wantScaffolds {
		if _, ok := gotScaffolds[name]; !ok {
			t.Fatalf("catalog.Scaffolds missing scaffold %q", name)
		}
	}

	for _, scaffold := range catalog.Scaffolds {
		if scaffold.Name == "baseline" && !scaffold.Baseline {
			t.Fatalf("baseline scaffold has Baseline = false")
		}
	}
}

func TestTaskCatalogFiltersByFamily(t *testing.T) {
	catalog := TaskCatalog{
		Tasks: []Task{
			{ID: "a", TaskFamily: "software_engineering"},
			{ID: "b", TaskFamily: "spreadsheets"},
		},
	}

	filtered := catalog.FilterByFamily("spreadsheets")
	if len(filtered.Tasks) != 1 || filtered.Tasks[0].ID != "b" {
		t.Fatalf("FilterByFamily() = %+v, want only task b", filtered.Tasks)
	}
}

func TestScaffoldCatalogFiltersByName(t *testing.T) {
	catalog := ScaffoldCatalog{
		Scaffolds: []Scaffold{
			{Name: "baseline"},
			{Name: "tool-assisted"},
		},
	}

	filtered := catalog.FilterByName("tool-assisted")
	if len(filtered.Scaffolds) != 1 || filtered.Scaffolds[0].Name != "tool-assisted" {
		t.Fatalf("FilterByName() = %+v, want only scaffold tool-assisted", filtered.Scaffolds)
	}
}
