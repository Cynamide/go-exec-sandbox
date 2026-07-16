package benchmark

// Problem is an alias for Task used by benchmark fixtures.
type Problem = Task

type TestCase struct {
	Input          string `json:"input"`
	ExpectedOutput string `json:"expected_output"`
}
