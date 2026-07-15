package benchmark

// Problem remains a compatibility alias for legacy benchmark fixtures.
type Problem = Task

type TestCase struct {
	Input          string `json:"input"`
	ExpectedOutput string `json:"expected_output"`
}
