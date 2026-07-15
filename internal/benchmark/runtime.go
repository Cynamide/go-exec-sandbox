package benchmark

import (
	"context"
	"gexec-sandbox/internal/api"
	"gexec-sandbox/internal/config"
)

type Executor interface {
	Execute(ctx context.Context, req api.ExecutionRequest, cfg config.Config) (api.ExecutionResponse, error)
}

type Grader interface {
	Grade(task Task, resp api.ExecutionResponse, tc TestCase) Outcome
}

func RunTask(ctx context.Context, task Task, scaffold Scaffold, mode RunMode, client LLMClient, exec Executor, cfg config.Config) Run {
	return RunTaskWithGrader(ctx, task, scaffold, mode, client, exec, DefaultGrader{}, cfg)
}

func RunTaskWithGrader(ctx context.Context, task Task, scaffold Scaffold, mode RunMode, client LLMClient, exec Executor, grader Grader, cfg config.Config) Run {
	prompt := task.Description
	if mode == RunModeScaffolded {
		prompt = scaffold.ApplyPrompt(prompt)
	}

	if err := ctx.Err(); err != nil {
		return Run{
			TaskID:   task.ID,
			Mode:     mode,
			Scaffold: scaffold,
			Error:    err.Error(),
		}
	}

	code, err := client.GenerateCode(ctx, prompt, task.Language)
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
		if err := ctx.Err(); err != nil {
			run.Passed = false
			run.Error = err.Error()
			return run
		}

		req := reqTemplate
		req.Stdin = tc.Input

		resp, err := exec.Execute(ctx, req, cfg)
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
