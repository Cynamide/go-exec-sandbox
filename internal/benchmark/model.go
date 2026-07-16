package benchmark

type ArtifactExpectation struct {
	Type           string `json:"type"`
	Format         string `json:"format,omitempty"`
	Description    string `json:"description,omitempty"`
	Input          string `json:"input,omitempty"`
	ExpectedOutput string `json:"expected_output,omitempty"`
}

type Task struct {
	ID                  string               `json:"id"`
	Title               string               `json:"title"`
	Description         string               `json:"description"`
	TaskFamily          string               `json:"task_family"`
	Language            string               `json:"language"`
	ArtifactExpectation *ArtifactExpectation `json:"artifact_expectation,omitempty"`
	TestCases           []TestCase           `json:"test_cases"`
}

type Scaffold struct {
	Baseline     bool     `json:"baseline,omitempty"`
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	PromptPrefix string   `json:"prompt_prefix"`
	Tools        []string `json:"tools,omitempty"`
}

type RunMode string

const (
	RunModeBaseline   RunMode = "baseline"
	RunModeScaffolded RunMode = "scaffolded"
)

type Run struct {
	TaskID   string    `json:"task_id"`
	Mode     RunMode   `json:"mode"`
	Scaffold Scaffold  `json:"scaffold"`
	Passed   bool      `json:"passed"`
	Outcomes []Outcome `json:"outcomes,omitempty"`
	Output   string    `json:"output,omitempty"`
	Error    string    `json:"error,omitempty"`
}

type Outcome struct {
	Passed bool    `json:"passed"`
	Score  float64 `json:"score"`
}

func (s Scaffold) ApplyPrompt(prompt string) string {
	return s.PromptPrefix + prompt
}
