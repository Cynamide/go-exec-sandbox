package benchmark

import (
	"gexec-sandbox/internal/api"
	"gexec-sandbox/internal/config"
	"gexec-sandbox/internal/sandbox"
)

type CodeExecutionAdapter struct {
	Runner func(req api.ExecutionRequest, cfg config.Config) (api.ExecutionResponse, error)
}

func NewCodeExecutionAdapter() CodeExecutionAdapter {
	return CodeExecutionAdapter{
		Runner: sandbox.RunCodeInSandbox,
	}
}

func (a CodeExecutionAdapter) Execute(req api.ExecutionRequest, cfg config.Config) (api.ExecutionResponse, error) {
	return a.Runner(req, cfg)
}
