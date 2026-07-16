package benchmark

import "testing"

func TestLoadTaskCatalogReturnsFamilies(t *testing.T) {
	catalog, err := LoadTaskCatalog("../../data/tasks.json")
	if err != nil {
		t.Fatalf("LoadTaskCatalog() error = %v", err)
	}

	if len(catalog.Tasks) == 0 {
		t.Fatalf("catalog.Tasks is empty")
	}

	foundFamily := false
	for _, task := range catalog.Tasks {
		if task.TaskFamily != "" {
			foundFamily = true
			break
		}
	}
	if !foundFamily {
		t.Fatalf("catalog.Tasks does not contain any task family")
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

	foundBaseline := false
	for _, scaffold := range catalog.Scaffolds {
		if scaffold.Baseline || scaffold.Name == "baseline" {
			foundBaseline = true
			break
		}
	}
	if !foundBaseline {
		t.Fatalf("catalog.Scaffolds does not contain a baseline scaffold")
	}
}
