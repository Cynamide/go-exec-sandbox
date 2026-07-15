package benchmark

import (
	"context"
	"gexec-sandbox/internal/api"
	"gexec-sandbox/internal/config"
	"gexec-sandbox/internal/sandbox"
)

type CodeExecutionAdapter struct {
	Runner func(ctx context.Context, req api.ExecutionRequest, cfg config.Config) (api.ExecutionResponse, error)
}

func NewCodeExecutionAdapter() CodeExecutionAdapter {
	return CodeExecutionAdapter{
		Runner: sandbox.RunCodeInSandbox,
	}
}

func (a CodeExecutionAdapter) Execute(ctx context.Context, req api.ExecutionRequest, cfg config.Config) (api.ExecutionResponse, error) {
	return a.Runner(ctx, req, cfg)
}
