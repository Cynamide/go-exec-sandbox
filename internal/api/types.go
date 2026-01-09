package api

type ExecutionRequest struct {
	Language   string `json:"language"`
	SourceCode string `json:"source_code"`
	TimeoutMS  int    `json:"timeout_ms"`
}

type ExecutionResponse struct {
	Stdout   string `json:"stdout"`
	Stderr   string `json:"stderr"`
	ExitCode int    `json:"exit_code"`
	Error    string `json:"error"`
}
