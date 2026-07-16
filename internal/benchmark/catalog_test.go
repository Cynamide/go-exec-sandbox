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

	gotFamilies := make(map[string]struct{}, len(catalog.Tasks))
	for _, task := range catalog.Tasks {
		gotFamilies[task.TaskFamily] = struct{}{}
	}

	wantFamilies := []string{
		"software_engineering",
		"browser_workflows",
		"spreadsheets",
		"terminal_workflows",
	}
	for _, family := range wantFamilies {
		if _, ok := gotFamilies[family]; !ok {
			t.Fatalf("catalog.Tasks missing task family %q", family)
		}
		delete(gotFamilies, family)
	}
	if len(gotFamilies) != 0 {
		t.Fatalf("catalog.Tasks contains unexpected task families: %v", gotFamilies)
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
		delete(gotScaffolds, name)
	}
	if len(gotScaffolds) != 0 {
		t.Fatalf("catalog.Scaffolds contains unexpected scaffold names: %v", gotScaffolds)
	}

	for _, scaffold := range catalog.Scaffolds {
		if scaffold.Name == "baseline" && !scaffold.Baseline {
			t.Fatalf("baseline scaffold has Baseline = false")
		}
	}
}
