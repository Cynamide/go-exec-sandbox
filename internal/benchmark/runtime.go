package benchmark

import (
	"strings"

	"gexec-sandbox/internal/api"
	"gexec-sandbox/internal/config"
)

type Executor interface {
	Execute(req api.ExecutionRequest, cfg config.Config) (api.ExecutionResponse, error)
}

type Grader interface {
	Grade(task Task, resp api.ExecutionResponse, tc TestCase) Outcome
}

type stdoutGrader struct{}

func (stdoutGrader) Grade(task Task, resp api.ExecutionResponse, tc TestCase) Outcome {
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

func RunTask(task Task, scaffold Scaffold, mode RunMode, client LLMClient, exec Executor, cfg config.Config) Run {
	return RunTaskWithGrader(task, scaffold, mode, client, exec, stdoutGrader{}, cfg)
}

func RunTaskWithGrader(task Task, scaffold Scaffold, mode RunMode, client LLMClient, exec Executor, grader Grader, cfg config.Config) Run {
	prompt := task.Description
	if mode == RunModeScaffolded {
		prompt = scaffold.ApplyPrompt(prompt)
	}

	code, err := client.GenerateCode(prompt, task.Language)
	if err != nil {
		return Run{
			TaskID:   task.ID,
			Mode:     mode,
			Scaffold: scaffold,
			Error:    err.Error(),
		}
	}

	reqTemplate := api.ExecutionRequest{
		Language:   task.Language,
		SourceCode: extractCode(code),
		TimeoutMS:  cfg.DefaultTimeoutMS,
	}

	outcomes := make([]Outcome, 0, len(task.TestCases))
	run := Run{
		TaskID:   task.ID,
		Mode:     mode,
		Scaffold: scaffold,
		Passed:   true,
	}

	for _, tc := range task.TestCases {
		req := reqTemplate
		req.Stdin = tc.Input

		resp, err := exec.Execute(req, cfg)
		if err != nil {
			run.Passed = false
			run.Error = err.Error()
			return run
		}

		run.Output = resp.Stdout

		outcome := grader.Grade(task, resp, tc)
		outcomes = append(outcomes, outcome)
		if !outcome.Passed {
			run.Passed = false
		}
	}

	run.Outcomes = outcomes
	return run
}
