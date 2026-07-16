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

	if catalog.Tasks[0].TaskFamily == "" {
		t.Fatalf("first task missing TaskFamily")
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

	if !catalog.Scaffolds[0].Baseline {
		t.Fatalf("first scaffold should be baseline")
	}
}
