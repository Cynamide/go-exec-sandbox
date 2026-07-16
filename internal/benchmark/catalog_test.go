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
		"engineering_workflows",
		"monitoring_workflows",
		"finance_workflows",
		"incident_workflows",
		"support_workflows",
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
	hasArtifactExpectation := false
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
		if task.ArtifactExpectation != nil {
			hasArtifactExpectation = true
			if task.ArtifactExpectation.Type == "" {
				t.Fatalf("task %q missing artifact type", task.ID)
			}
			if task.ArtifactExpectation.Format == "" {
				t.Fatalf("task %q missing artifact format", task.ID)
			}
			if task.ArtifactExpectation.Description == "" {
				t.Fatalf("task %q missing artifact description", task.ID)
			}
			if len(task.TestCases) == 0 && task.ArtifactExpectation.ExpectedOutput == "" {
				t.Fatalf("task %q missing artifact expected output", task.ID)
			}
			if len(task.TestCases) == 0 && task.ArtifactExpectation.Input == "" {
				t.Fatalf("task %q missing artifact input", task.ID)
			}
		}
		if len(task.TestCases) == 0 && task.ArtifactExpectation == nil {
			t.Fatalf("task %q missing test cases and artifact expectation", task.ID)
		}
		families[task.TaskFamily] = true
	}

	if len(families) < 5 {
		t.Fatalf("families = %v, want at least 5", families)
	}

	if !hasArtifactExpectation {
		t.Fatalf("catalog.Tasks does not contain any artifact expectations")
	}
}

func TestLoadTaskCatalogRejectsTaskWithoutTestCases(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tasks.json")
	if err := os.WriteFile(path, []byte(`{"tasks":[{"id":"artifact-task","title":"Artifact Task","description":"desc","task_family":"support_workflows","language":"python"}]}`), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	if _, err := LoadTaskCatalog(path); err == nil {
		t.Fatal("LoadTaskCatalog() error = nil, want invalid task catalog error")
	}
}

func TestLoadTaskCatalogAcceptsArtifactOnlyTask(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tasks.json")
	if err := os.WriteFile(path, []byte(`{"tasks":[{"id":"artifact-task","title":"Artifact Task","description":"desc","task_family":"support_workflows","language":"python","artifact_expectation":{"type":"markdown_report","format":"markdown","description":"artifact output","input":"p1|platform|open\np1|billing|open\n","expected_output":"| team | open |\n| --- | --- |\n| billing | 1 |\n| platform | 1 |"}}]}`), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	catalog, err := LoadTaskCatalog(path)
	if err != nil {
		t.Fatalf("LoadTaskCatalog() error = %v", err)
	}

	if len(catalog.Tasks) != 1 || catalog.Tasks[0].ArtifactExpectation == nil {
		t.Fatalf("catalog = %+v, want artifact-only task", catalog)
	}
}

func TestLoadTaskCatalogRejectsUnknownFields(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tasks.json")
	if err := os.WriteFile(path, []byte(`{"tasks":[{"id":"task-1","title":"Task 1","description":"desc","task_family":"support_workflows","language":"python","test_cases":[{"input":"","expected_output":"ok"}],"unexpected":"boom"}]}`), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	if _, err := LoadTaskCatalog(path); err == nil {
		t.Fatal("LoadTaskCatalog() error = nil, want unknown field error")
	}
}

func TestLoadTaskCatalogRejectsDuplicateTaskIDs(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tasks.json")
	if err := os.WriteFile(path, []byte(`{"tasks":[{"id":"task-1","title":"Task 1","description":"desc","task_family":"support_workflows","language":"python","test_cases":[{"input":"","expected_output":"ok"}]},{"id":"task-1","title":"Task 2","description":"desc","task_family":"support_workflows","language":"python","test_cases":[{"input":"","expected_output":"ok"}]}]}`), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	if _, err := LoadTaskCatalog(path); err == nil {
		t.Fatal("LoadTaskCatalog() error = nil, want duplicate task ID error")
	}
}

func TestLoadTaskCatalogRejectsMalformedArtifactMetadata(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tasks.json")
	if err := os.WriteFile(path, []byte(`{"tasks":[{"id":"task-1","title":"Task 1","description":"desc","task_family":"support_workflows","language":"python","artifact_expectation":{"type":"markdown_report","format":"markdown"},"test_cases":[{"input":"","expected_output":"ok"}]}]}`), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	if _, err := LoadTaskCatalog(path); err == nil {
		t.Fatal("LoadTaskCatalog() error = nil, want malformed artifact metadata error")
	}
}

func TestLoadTaskCatalogRejectsTrailingJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tasks.json")
	if err := os.WriteFile(path, []byte(`{"tasks":[{"id":"task-1","title":"Task 1","description":"desc","task_family":"support_workflows","language":"python","test_cases":[{"input":"","expected_output":"ok"}]}]} {"unexpected":true}`), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	if _, err := LoadTaskCatalog(path); err == nil {
		t.Fatal("LoadTaskCatalog() error = nil, want trailing JSON error")
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

func TestLoadScaffoldCatalogRejectsDuplicateNamesOrBaselines(t *testing.T) {
	path := filepath.Join(t.TempDir(), "scaffolds.json")
	if err := os.WriteFile(path, []byte(`{"scaffolds":[{"name":"baseline","baseline":true,"description":"baseline","prompt_prefix":""},{"name":"baseline","baseline":false,"description":"duplicate","prompt_prefix":"dup: "}]}`), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	if _, err := LoadScaffoldCatalog(path); err == nil {
		t.Fatal("LoadScaffoldCatalog() error = nil, want duplicate scaffold name error")
	}
}

func TestTaskCatalogFiltersByFamily(t *testing.T) {
	catalog := TaskCatalog{
		Tasks: []Task{
			{ID: "a", TaskFamily: "engineering_workflows"},
			{ID: "b", TaskFamily: "finance_workflows"},
		},
	}

	filtered := catalog.FilterByFamily("finance_workflows")
	if len(filtered.Tasks) != 1 || filtered.Tasks[0].ID != "b" {
		t.Fatalf("FilterByFamily() = %+v, want only task b", filtered.Tasks)
	}
}

func TestScaffoldCatalogFiltersByName(t *testing.T) {
	catalog := ScaffoldCatalog{
		Scaffolds: []Scaffold{
			{Baseline: true, Name: "baseline"},
			{Name: "tool-assisted"},
		},
	}

	filtered := catalog.FilterByName("tool-assisted")
	if len(filtered.Scaffolds) != 2 || filtered.Scaffolds[0].Name != "baseline" || filtered.Scaffolds[1].Name != "tool-assisted" {
		t.Fatalf("FilterByName() = %+v, want baseline plus tool-assisted", filtered.Scaffolds)
	}
}
