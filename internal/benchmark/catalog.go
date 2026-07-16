package benchmark

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
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
	if err := requireEOF(decoder); err != nil {
		return TaskCatalog{}, err
	}

	for _, task := range catalog.Tasks {
		if err := validateTask(task); err != nil {
			return TaskCatalog{}, err
		}
	}
	seenIDs := make(map[string]struct{}, len(catalog.Tasks))
	for _, task := range catalog.Tasks {
		if _, ok := seenIDs[task.ID]; ok {
			return TaskCatalog{}, ErrInvalidTaskCatalog
		}
		seenIDs[task.ID] = struct{}{}
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
	if err := requireEOF(decoder); err != nil {
		return ScaffoldCatalog{}, err
	}

	seenNames := make(map[string]struct{}, len(catalog.Scaffolds))
	baselineCount := 0
	for _, scaffold := range catalog.Scaffolds {
		if scaffold.Name == "" {
			return ScaffoldCatalog{}, ErrInvalidTaskCatalog
		}
		if _, ok := seenNames[scaffold.Name]; ok {
			return ScaffoldCatalog{}, ErrInvalidTaskCatalog
		}
		seenNames[scaffold.Name] = struct{}{}
		if scaffold.Baseline {
			baselineCount++
		}
	}
	if baselineCount != 1 {
		return ScaffoldCatalog{}, ErrInvalidTaskCatalog
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
	if task.ArtifactExpectation != nil {
		if task.ArtifactExpectation.Type == "" || task.ArtifactExpectation.Format == "" || task.ArtifactExpectation.Description == "" {
			return ErrInvalidTaskCatalog
		}
	}

	return nil
}

func requireEOF(decoder *json.Decoder) error {
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return ErrInvalidTaskCatalog
	}

	return nil
}
