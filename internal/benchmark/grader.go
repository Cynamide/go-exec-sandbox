package benchmark

import (
	"strings"

	"gexec-sandbox/internal/api"
)

// DefaultGrader preserves the existing stdout-based scoring contract used by code tasks.
type DefaultGrader struct{}

func (DefaultGrader) Grade(task Task, resp api.ExecutionResponse, tc TestCase) Outcome {
	passed := strings.TrimSpace(resp.Stdout) == strings.TrimSpace(tc.ExpectedOutput)

	score := 0.0
	if passed {
		score = 1.0
	}

	return Outcome{
		Passed: passed,
		Score:  score,
	}
}
