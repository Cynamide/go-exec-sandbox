# LocalEval: Secure Code Execution Sandbox & Benchmark Harness
A secure code execution service and benchmark harness for economically meaningful, real-world agent tasks. It runs user-submitted or AI-generated code in isolated Docker containers and is designed to evaluate deliverables, workflows, and tool-using behavior rather than puzzle solving.

## Features

- 🐳 **Secure Execution**: Code runs in isolated Docker containers with network disabled
- ⚡ **Multi-Language Support**: Python and Golang out of the box (easily extensible)
- 🤖 **LLM Integration**: Built-in Ollama client for local LLM inference
- 🛡️ **Resource Limits**: Configurable memory and CPU quotas prevent abuse
- ⏱️ **Timeout Control**: Enforce execution time limits per request
- ✅ **Robust Validation**: Input validation for language support and source code
- 📊 **Structured Response**: JSON responses with stdout, stderr, exit codes, and error details
- 📈 **Metrics Tracking**: Built-in request and error tracking for observability
- 🛑 **Graceful Shutdown**: Handles SIGINT/SIGTERM with proper cleanup of in-flight requests and containers
- 🚦 **Rate Limiting**: In-memory IP-based rate limiting (10 requests/minute by default)
- 🐋 **Docker Compose Ready**: Complete orchestration with Ollama service

## Prerequisites

- **Go 1.24+** - [Install Go](https://golang.org/dl/)
- **Docker & Docker Compose** - [Install Docker](https://docs.docker.com/get-docker/)
- Docker daemon must be running

### Dependencies

Dependencies are automatically downloaded when you run `go mod download` or `go build`. Key dependencies include:

- `github.com/docker/docker` - Docker SDK for Go
- `golang.org/x/time/rate` - Rate limiting implementation
- `github.com/ollama/ollama/api` - Ollama API client for local LLM inference

## Security

This service uses Docker containers to isolate user code execution:

- **Network Disabled**: Containers run with network access disabled to prevent unauthorized network calls
- **Resource Limits**: Memory and CPU quotas restrict resource usage (default: 256MB, 50k CPU quota)
- **Ephemeral Containers**: Containers are automatically removed after execution
- **Context Timeouts**: Execution is enforced with context timeouts to prevent hanging processes
- **Graceful Shutdown**: Server catches SIGINT/SIGTERM signals and properly cleans up all active containers
- **Rate Limiting**: In-memory IP-based rate limiting prevents abuse (10 requests/minute with 10 burst)

> ⚠️ **Important**: While Docker provides strong isolation, this service should still be run behind additional security layers (authentication, firewall, etc.) in production environments.

## Installation

```bash
# Clone the repository
git clone <repository-url>
cd gexec-sandbox

# Install dependencies
go mod download

# Copy environment file and configure
cp .env.example .env
# Edit .env to set your preferred Ollama model
```

## Environment Configuration

Create a `.env` file in the project root:

```bash
# Required: Ollama model to use (set in .env before running)
OLLAMA_MODEL=qwen3:4b

# Optional: Ollama host URL (default: http://localhost:11434)
OLLAMA_HOST=http://localhost:11434
```

## Running the Evaluator

### Using Docker Compose (Recommended)

```bash
# Start all services (Ollama + Evaluator)
docker-compose up -d

# View logs
docker-compose logs -f

# Stop services
docker-compose down
```

This will:
1. Pull and start Ollama service with your configured model
2. Build and start the evaluator service
3. Connect the evaluator to Ollama automatically

### Local Development

```bash
# Start Ollama locally first (if not using Docker)
docker run -d -p 11434:11434 ollama/ollama
docker exec -it <container_id> ollama pull qwen3:4b

# Start the evaluator
go run ./cmd/evaluator

# Run the benchmark CLI and print a JSON report to stdout
go run ./cmd/evaluator benchmark

# Or build and run
go build -o evaluator ./cmd/evaluator
./evaluator
```

The evaluator will start on `http://localhost:8080`. The `benchmark` mode uses the same shared benchmark service and prints the report as JSON instead of starting the HTTP server.

### Using Docker (Standalone)

```bash
# Build the Docker image
docker build -t gexec-sandbox .

# Run the container (requires mounting Docker socket)
docker run -p 8080:8080 \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -e OLLAMA_HOST=http://host.docker.internal:11434 \
  -e OLLAMA_MODEL=qwen3:4b \
  gexec-sandbox
```

## API Usage

### Execute Code

**Endpoint**: `POST /execute`

**Request Body**:
```json
{
  "language": "python",
  "source_code": "print('Hello, World!')",
  "timeout_ms": 5000
}
```

**Response**:
```json
{
  "stdout": "Hello, World!\n",
  "stderr": "",
  "exit_code": 0,
  "error": ""
}
```

### Health Check

**Endpoint**: `GET /ping`

**Response**:
```json
{
  "status": "ok"
}
```

### Metrics

**Endpoint**: `GET /metrics`

**Response**:
```json
{
  "total_requests": 42,
  "total_errors": 3
}
```

### Rate Limiting

The `/execute` endpoint is rate limited to **10 requests per minute per IP address** (configurable).

**Response when rate limited** (HTTP 429):
```json
{
  "error": "Too many requests"
}
```

You can adjust the rate limit in `cmd/evaluator/main.go` by modifying the `RateLimitMiddleware` parameters.

## Example Commands

### Python Example

```bash
curl -X POST http://localhost:8080/execute \
  -H "Content-Type: application/json" \
  -d '{
    "language": "python",
    "source_code": "print(\"Hello from Python!\")\nfor i in range(5):\n    print(i)"
  }'
```

**Response**:
```json
{
  "stdout": "Hello from Python!\n0\n1\n2\n3\n4\n",
  "stderr": "",
  "exit_code": 0,
  "error": ""
}
```

### Golang Example

```bash
curl -X POST http://localhost:8080/execute \
  -H "Content-Type: application/json" \
  -d '{
    "language": "go",
    "source_code": "package main\n\nimport \"fmt\"\n\nfunc main() {\n    fmt.Println(\"Hello from Go!\")\n    for i := 0; i < 3; i++ {\n        fmt.Println(i)\n    }\n}"
  }'
```

**Response**:
```json
{
  "stdout": "Hello from Go!\n0\n1\n2\n",
  "stderr": "",
  "exit_code": 0,
  "error": ""
}
```

### Error Handling Examples

**Unsupported Language** (HTTP 400):
```bash
curl -X POST http://localhost:8080/execute \
  -H "Content-Type: application/json" \
  -d '{"language": "ruby", "source_code": "puts \"hi\"}'

# Response: {"error":"unsupported language: ruby"}
```

**Empty Source Code** (HTTP 400):
```bash
curl -X POST http://localhost:8080/execute \
  -H "Content-Type: application/json" \
  -d '{"language": "python", "source_code": ""}'

# Response: {"error":"source_code cannot be empty"}
```

## Configuration

### Environment Variables

The evaluator reads configuration from environment variables:

- `OLLAMA_HOST`: Ollama server URL (default: `http://localhost:11434`)
- `OLLAMA_MODEL`: Model name to use (required, e.g., `qwen3:4b`, `codellama:latest`)

### Code Configuration

Edit `internal/config/config.go` to customize:

```go
Config{
    DefaultTimeoutMS: 60000,  // Default timeout in milliseconds
    MaxMemoryMB:      256,   // Maximum memory per container (MB)
    OLLAMAHost:       "http://localhost:11434",
    OLLAMAModel:      "qwen3:4b",
    Languages: map[string]string{
        "python": "python:3.9-slim",
        "py":     "python:3.9-slim",
        "golang": "golang:1.24-alpine",
        "go":     "golang:1.24-alpine",
    },
}
```

**Rate Limiting Configuration**

Edit `cmd/evaluator/main.go` to adjust rate limiting:

```go
// Current: 10 requests per minute with burst of 2
mux.Handle("/execute", middleware.RateLimitMiddleware(
    rate.Every(6*time.Second),  // 1 request every 6 seconds = 10/min
    2,                          // Burst allowance
)(http.HandlerFunc(executeHandler(cfg))))
```

**Graceful Shutdown Configuration**

Edit `cmd/evaluator/main.go` to adjust shutdown timeout:

```go
// Current: 30 second graceful shutdown timeout
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
```

## Adding New Languages

1. Add the Docker image to `Languages` map in `config.go`
2. Update `getCommand()` function in `docker.go` to return the correct execution command
3. Update `getExtension()` function in `docker.go` to return the correct file extension

## Project Structure

```
gexec-sandbox/
├── cmd/
│   └── evaluator/
│       └── main.go          # HTTP server, LLM integration, graceful shutdown, and handlers
├── data/
│   ├── tasks.json           # Benchmark task catalog
│   ├── scaffolds.json       # Benchmark scaffold catalog
│   └── problems.json        # Legacy toy benchmark dataset
├── internal/
│   ├── api/
│   │   └── types.go         # Request/response types
│   ├── benchmark/
│   │   ├── harness.go       # Benchmark harness for running evaluations
│   │   └── types.go         # Benchmark-specific type definitions
│   ├── config/
│   │   └── config.go        # Configuration management with env var support
│   ├── llm/
│   │   └── llm.go           # Ollama client for LLM inference and model management
│   ├── metrics/
│   │   └── metrics.go       # Request and error metrics tracking
│   ├── middleware/
│   │   └── rate_limiter.go  # IP-based rate limiting middleware
│   └── sandbox/
│       └── docker.go        # Docker container execution logic with cleanup
├── .env.example             # Environment variable template
├── .gitignore               # Git ignore patterns
├── docker-compose.yml       # Multi-service orchestration (Ollama + Evaluator)
├── Dockerfile               # Evaluator container definition
├── go.mod                   # Go module definition
├── go.sum                   # Go dependencies checksums
└── README.md                # This file
```

## Development

### Testing the Service

```bash
# Test health endpoint
curl http://localhost:8080/ping

# Test Python execution
curl -X POST http://localhost:8080/execute \
  -H "Content-Type: application/json" \
  -d '{"language": "python", "source_code": "print(2+2)"}'

# Test metrics
curl http://localhost:8080/metrics
```

### Building

```bash
# Build for current platform
go build -o evaluator ./cmd/evaluator

# Build for Linux (for Docker)
GOOS=linux GOARCH=amd64 go build -o evaluator-linux ./cmd/evaluator
```

### Graceful Shutdown Testing

The evaluator gracefully handles shutdown signals (SIGINT/SIGTERM):

```bash
# Start the evaluator
go run ./cmd/evaluator

# In another terminal, send a request
curl -X POST http://localhost:8080/execute \
  -H "Content-Type: application/json" \
  -d '{"language": "python", "source_code":"import time; time.sleep(60)"}'

# Press Ctrl+C in the evaluator terminal
# Evaluator will wait up to 30 seconds for in-flight requests to complete
# All active containers will be cleaned up automatically
```

### Rate Limiting Testing

```bash
# Send 10 rapid requests (should all succeed)
for i in {1..10}; do
  curl -X POST http://localhost:8080/execute \
    -H "Content-Type: application/json" \
    -d '{"language": "python", "source_code":"print('$i')"}'
done

# 11th request will be rate limited
curl -X POST http://localhost:8080/execute \
  -H "Content-Type: application/json" \
  -d '{"language": "python", "source_code":"print(11)"}'
# Returns: HTTP 429 Too Many Requests
```

## Project Roadmap

This project is being developed as a benchmark harness for economically meaningful, real-world tasks. Here's the current status based on the project goals:

### ✅ Completed Features

- **Benchmark Harness**
  - ✅ Architected Go-based evaluation system orchestrating local inference (Ollama)
  - ✅ Docker Compose orchestration for Ollama and evaluator services
  - ✅ Environment-based configuration for model selection and host settings
  - ✅ Connection handling and availability checking for Ollama service

- **Secure Multi-Language Execution Sandbox**
  - ✅ Engineered secure sandbox isolating untrusted code in Docker containers
  - ✅ Network restrictions on containers (network disabled by default)
  - ✅ Memory and CPU resource limits to prevent abuse
  - ✅ Support for Python and Golang with easy extensibility
  - ✅ Safe execution of AI-generated code output

- **Docker Orchestration**
  - ✅ Docker Compose setup with Ollama service
  - ✅ Automatic model pulling on Ollama startup
  - ✅ Service networking and volume management
  - ✅ Environment variable configuration for flexibility

- **Core Infrastructure**
  - ✅ stdin/stdout piping for precise test case evaluation
  - ✅ Timeout enforcement and graceful container cleanup
  - ✅ Rate limiting and request metrics
  - ✅ Structured JSON API responses
  - ✅ Graceful shutdown with container cleanup

### 🚧 In Progress / Planned Features

- **Benchmarking Pipeline**
  - ✅ Benchmark harness infrastructure (internal/benchmark/)
  - ✅ Task and scaffold catalog structure (data/tasks.json, data/scaffolds.json)
  - ✅ Integration of LLM code generation with execution
  - ✅ Dataset management system for tasks and test cases

- **Evaluation Metrics**
  - ✅ Pass@k metric calculation (k=1, k=5, k=10)
  - 🚧 Statistical analysis and reporting
  - 🚧 Performance benchmarking across multiple models

- **Benchmark Scope**
  - ✅ Code execution and validation against expected outputs
  - ✅ Support for stdin-driven workflows
  - ✅ Support for software-engineering-style issue resolution
  - 🚧 Expansion to other economically valuable workflows such as office, spreadsheet, web, and computer-use tasks

- **Enhanced Features**
  - 🚧 Batch evaluation mode for comparing multiple models
  - 🚧 Result caching and persistence
  - 🚧 Progress tracking and status reporting
  - ✅ Benchmark CLI mode for running benchmarks locally

### 🎯 Future Enhancements

- Multi-model comparison support
- Distributed evaluation setup
- Custom test case format support
- Performance profiling and optimization
- Result visualization and dashboards
- Integration with additional LLM backends

## License

MIT License - feel free to use this in your projects!
