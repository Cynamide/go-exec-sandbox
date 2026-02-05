package benchmark

import (
	"regexp"
	"strings"

	"gexec-sandbox/internal/api"
	"gexec-sandbox/internal/config"
	"gexec-sandbox/internal/sandbox"
)

type LLMClient interface {
	GenerateCode(problem string, language string) (string, error)
}

type Report struct {
	TotalProblems  int     `json:"total_problems"`
	PassedProblems int     `json:"passed_problems"`
	Pass1Rate      float64 `json:"pass_1_rate"`
	PassKRate      float64 `json:"pass_k_rate"`
}

func RunEvaluation(problems []Problem, k int, client LLMClient) Report {
	cfg := config.LoadConfig()
	total := len(problems)
	passed := 0
	pass1 := 0

	for _, problem := range problems {
		passedProblem := false
		pass1Attempt := false

		for attempt := 1; attempt <= k; attempt++ {
			code, err := client.GenerateCode(problem.Description, problem.Language)
			if err != nil {
				continue
			}

			code = extractCode(code)

			allPassed := true
			for _, tc := range problem.TestCases {
				req := api.ExecutionRequest{
					Language:   problem.Language,
					SourceCode: code,
					TimeoutMS:  cfg.DefaultTimeoutMS,
				}

				resp, err := sandbox.RunCodeInSandbox(req, cfg)
				if err != nil {
					allPassed = false
					break
				}

				stdout := strings.TrimSpace(resp.Stdout)
				expected := strings.TrimSpace(tc.ExpectedOutput)

				if stdout != expected {
					allPassed = false
					break
				}
			}

			if allPassed {
				passedProblem = true
				if attempt == 1 {
					pass1Attempt = true
				}
				break
			}
		}

		if passedProblem {
			passed++
			if pass1Attempt {
				pass1++
			}
		}
	}

	return Report{
		TotalProblems:  total,
		PassedProblems: passed,
		Pass1Rate:      float64(pass1) / float64(total),
		PassKRate:      float64(passed) / float64(total),
	}
}

func extractCode(text string) string {
	codeBlockRegex := regexp.MustCompile("```(?:python|go|golang)?\n?(.*?)```")
	matches := codeBlockRegex.FindStringSubmatch(text)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}

	multiCodeBlockRegex := regexp.MustCompile("```\n(.*?)```")
	matches2 := multiCodeBlockRegex.FindStringSubmatch(text)
	if len(matches2) > 1 {
		return strings.TrimSpace(matches2[1])
	}

	return strings.TrimSpace(text)
}
