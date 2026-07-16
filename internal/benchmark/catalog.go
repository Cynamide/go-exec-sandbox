package benchmark

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
)

var ErrInvalidTaskCatalog = errors.New("invalid task catalog")

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
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&catalog); err != nil {
		return TaskCatalog{}, err
	}

	for _, task := range catalog.Tasks {
		if err := validateTask(task); err != nil {
			return TaskCatalog{}, err
		}
	}

	return catalog, nil
}

func LoadScaffoldCatalog(path string) (ScaffoldCatalog, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return ScaffoldCatalog{}, err
	}

	var catalog ScaffoldCatalog
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&catalog); err != nil {
		return ScaffoldCatalog{}, err
	}

	return catalog, nil
}

func (c TaskCatalog) FilterByFamily(family string) TaskCatalog {
	filtered := TaskCatalog{}
	for _, task := range c.Tasks {
		if task.TaskFamily == family {
			filtered.Tasks = append(filtered.Tasks, task)
		}
	}
	return filtered
}

func (c ScaffoldCatalog) FilterByName(name string) ScaffoldCatalog {
	filtered := ScaffoldCatalog{}
	for _, scaffold := range c.Scaffolds {
		if scaffold.Name == name {
			filtered.Scaffolds = append(filtered.Scaffolds, scaffold)
		}
	}
	return filtered
}

func validateTask(task Task) error {
	if task.ID == "" {
		return ErrInvalidTaskCatalog
	}
	if task.Title == "" {
		return ErrInvalidTaskCatalog
	}
	if task.Description == "" {
		return ErrInvalidTaskCatalog
	}
	if task.TaskFamily == "" {
		return ErrInvalidTaskCatalog
	}
	if task.Language == "" {
		return ErrInvalidTaskCatalog
	}
	if len(task.TestCases) == 0 {
		return ErrInvalidTaskCatalog
	}

	return nil
}
