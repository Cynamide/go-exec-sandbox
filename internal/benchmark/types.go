package benchmark

type Problem = Task

type TestCase struct {
	Input          string `json:"input"`
	ExpectedOutput string `json:"expected_output"`
}
