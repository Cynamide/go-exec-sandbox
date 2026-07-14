# Project Context Map

## Identity

- Repository name: `gexec-sandbox`
- Primary description in the README: a secure code execution sandbox plus LLM evaluation harness
- Main language: Go
- Runtime model: HTTP service backed by Docker containers and Ollama

## What The Project Contains

### Runtime service

- The executable entrypoint is [cmd/evaluator/main.go](/Users/arjit/Documents/go-exec-sandbox/cmd/evaluator/main.go)
- The service exposes:
  - `POST /execute`
  - `GET /ping`
  - `GET /metrics`
- The service starts an HTTP server on port `8080`
- Shutdown is handled with `SIGINT` and `SIGTERM`

### Execution path

- Requests are parsed into `internal/api.ExecutionRequest`
- Empty `source_code` is rejected
- Missing `timeout_ms` is filled with the default timeout from config
- Execution is delegated to `internal/sandbox.RunCodeInSandbox`
- Responses use `internal/api.ExecutionResponse`

### Sandbox layer

- Sandbox code lives in [internal/sandbox/docker.go](/Users/arjit/Documents/go-exec-sandbox/internal/sandbox/docker.go)
- Execution happens inside Docker containers
- Containers are configured with:
  - network disabled
  - memory limit
  - CPU quota
  - automatic cleanup after use
- The current language mapping supports:
  - Python
  - Golang
- Language aliases in config include:
  - `python`
  - `py`
  - `golang`
  - `go`

### Ollama / LLM layer

- Ollama integration lives in [internal/llm/llm.go](/Users/arjit/Documents/go-exec-sandbox/internal/llm/llm.go)
- The code includes:
  - a client wrapper for generation
  - a wait loop for Ollama availability
  - a model pull helper
  - a model existence check helper
- The LLM layer uses the model name from `OLLAMA_MODEL`

### Benchmark layer

- Benchmark scaffolding lives in:
  - [internal/benchmark/types.go](/Users/arjit/Documents/go-exec-sandbox/internal/benchmark/types.go)
  - [internal/benchmark/harness.go](/Users/arjit/Documents/go-exec-sandbox/internal/benchmark/harness.go)
- The benchmark package defines:
  - `Problem`
  - `TestCase`
  - `Report`
  - `RunEvaluation`
  - `extractCode`
- A sample problem set exists in [data/problems.json](/Users/arjit/Documents/go-exec-sandbox/data/problems.json)

### Observability and middleware

- Request and error counters live in [internal/metrics/metrics.go](/Users/arjit/Documents/go-exec-sandbox/internal/metrics/metrics.go)
- Per-IP rate limiting lives in [internal/middleware/rate_limiter.go](/Users/arjit/Documents/go-exec-sandbox/internal/middleware/rate_limiter.go)

### Configuration

- Configuration loading lives in [internal/config/config.go](/Users/arjit/Documents/go-exec-sandbox/internal/config/config.go)
- Environment variables used by the project:
  - `OLLAMA_HOST`
  - `OLLAMA_MODEL`
- Defaults in code:
  - Ollama host defaults to `http://localhost:11434`
  - Default execution timeout is `60000` ms
  - Default memory limit is `256` MB

## Repository Files

- `cmd/evaluator/main.go` - service entrypoint
- `data/problems.json` - sample benchmark/problem dataset
- `internal/api/types.go` - request and response types
- `internal/benchmark/harness.go` - benchmark runner logic
- `internal/benchmark/types.go` - benchmark data models
- `internal/config/config.go` - environment-based config loading
- `internal/llm/llm.go` - Ollama client helpers
- `internal/metrics/metrics.go` - atomic request/error counters
- `internal/middleware/rate_limiter.go` - IP-based rate limiting middleware
- `internal/sandbox/docker.go` - Docker execution sandbox

## Infrastructure Files

- [Dockerfile](/Users/arjit/Documents/go-exec-sandbox/Dockerfile)
- [docker-compose.yml](/Users/arjit/Documents/go-exec-sandbox/docker-compose.yml)
- [.env.example](/Users/arjit/Documents/go-exec-sandbox/.env.example)

## Documentation Shape

- The README presents the project as a secure execution sandbox and LLM evaluation harness
- The README also contains a roadmap section that describes benchmark-oriented goals
- One README reference points to `cmd/server/main.go`, but the actual entrypoint in the repo is `cmd/evaluator/main.go`

## Current State Summary

- The repository already includes the core HTTP service, Docker-backed sandbox execution, Ollama connectivity, metrics, and rate limiting
- The repository also includes benchmark data structures and an evaluation harness scaffold
- The codebase is organized around a single executable service with supporting internal packages
