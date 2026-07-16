package benchmark

import (
	"encoding/json"
	"os"
)

type TaskCatalog struct {
	Tasks []Task `json:"tasks"`
}

type ScaffoldCatalog struct {
	Scaffolds []Scaffold `json:"scaffolds"`
}

func LoadTaskCatalog(path string) (TaskCatalog, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return TaskCatalog{}, err
	}

	var catalog TaskCatalog
	if err := json.Unmarshal(raw, &catalog); err != nil {
		return TaskCatalog{}, err
	}

	return catalog, nil
}

func LoadScaffoldCatalog(path string) (ScaffoldCatalog, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return ScaffoldCatalog{}, err
	}

	var catalog ScaffoldCatalog
	if err := json.Unmarshal(raw, &catalog); err != nil {
		return ScaffoldCatalog{}, err
	}

	return catalog, nil
}
