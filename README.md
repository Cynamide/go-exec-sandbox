# gexec-sandbox

A secure code execution service that runs user-submitted code in isolated Docker containers. Perfect for online code editors, interview platforms, or educational tools.

## Features

- 🐳 **Secure Execution**: Code runs in isolated Docker containers with network disabled
- ⚡ **Multi-Language Support**: Python and Golang out of the box (easily extensible)
- 🛡️ **Resource Limits**: Configurable memory and CPU quotas prevent abuse
- ⏱️ **Timeout Control**: Enforce execution time limits per request
- ✅ **Robust Validation**: Input validation for language support and source code
- 📊 **Structured Response**: JSON responses with stdout, stderr, exit codes, and error details
- 📈 **Metrics Tracking**: Built-in request and error tracking for observability
- 🛑 **Graceful Shutdown**: Handles SIGINT/SIGTERM with proper cleanup of in-flight requests and containers
- 🚦 **Rate Limiting**: In-memory IP-based rate limiting (10 requests/minute by default)

## Prerequisites

- **Go 1.21+** - [Install Go](https://golang.org/dl/)
- **Docker** - [Install Docker](https://docs.docker.com/get-docker/)
- Docker daemon must be running

### Dependencies

Dependencies are automatically downloaded when you run `go mod download` or `go build`. Key dependencies include:

- `github.com/docker/docker` - Docker SDK for Go
- `golang.org/x/time/rate` - Rate limiting implementation

## Security

This service uses Docker containers to isolate user code execution:

- **Network Disabled**: Containers run with network access disabled to prevent unauthorized network calls
- **Resource Limits**: Memory and CPU quotas restrict resource usage (default: 128MB, 50k CPU quota)
- **Ephemeral Containers**: Containers are automatically removed after execution
- **Context Timeouts**: Execution is enforced with context timeouts to prevent hanging processes
- **Graceful Shutdown**: Server catches SIGINT/SIGTERM signals and properly cleans up all active containers
- **Rate Limiting**: In-memory IP-based rate limiting prevents abuse (10 requests/minute with 2 burst)

> ⚠️ **Important**: While Docker provides strong isolation, this service should still be run behind additional security layers (authentication, firewall, etc.) in production environments.

## Installation

```bash
# Clone the repository
git clone <repository-url>
cd gexec-sandbox

# Install dependencies
go mod download
```

## Running the Server

### Local Development

```bash
# Start the server
go run ./cmd/server

# Or build and run
go build -o server ./cmd/server
./server
```

The server will start on `http://localhost:8080`

### Using Docker

```bash
# Build the Docker image
docker build -t gexec-sandbox .

# Run the container (requires mounting Docker socket)
docker run -p 8080:8080 -v /var/run/docker.sock:/var/run/docker.sock gexec-sandbox
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

You can adjust the rate limit in `cmd/server/main.go` by modifying the `RateLimitMiddleware` parameters.

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

Edit `internal/config/config.go` to customize:

```go
Config{
    DefaultTimeoutMS: 60000,  // Default timeout in milliseconds
    MaxMemoryMB:      256,   // Maximum memory per container (MB)
    Languages: map[string]string{
        "python": "python:3.9-slim",
        "py":     "python:3.9-slim",
        "golang": "golang:1.21-alpine",
        "go":     "golang:1.21-alpine",
    },
}
```

**Rate Limiting Configuration**

Edit `cmd/server/main.go` to adjust rate limiting:

```go
// Current: 10 requests per minute with burst of 2
mux.Handle("/execute", middleware.RateLimitMiddleware(
    rate.Every(6*time.Second),  // 1 request every 6 seconds = 10/min
    2,                          // Burst allowance
)(http.HandlerFunc(executeHandler(cfg))))
```

**Graceful Shutdown Configuration**

Edit `cmd/server/main.go` to adjust shutdown timeout:

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
│   └── server/
│       └── main.go          # HTTP server, graceful shutdown, and handlers
├── internal/
│   ├── api/
│   │   └── types.go         # Request/response types
│   ├── config/
│   │   └── config.go        # Configuration management
│   ├── metrics/
│   │   └── metrics.go       # Request and error metrics tracking
│   ├── middleware/
│   │   └── rate_limiter.go  # IP-based rate limiting middleware
│   └── sandbox/
│       └── docker.go        # Docker container execution logic with cleanup
├── go.mod                   # Go module definition
├── go.sum                   # Go dependencies checksums
├── Dockerfile               # Service container definition
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
```

### Building

```bash
# Build for current platform
go build -o server ./cmd/server

# Build for Linux (for Docker)
GOOS=linux GOARCH=amd64 go build -o server-linux ./cmd/server
```

### Graceful Shutdown Testing

The server gracefully handles shutdown signals (SIGINT/SIGTERM):

```bash
# Start the server
go run ./cmd/server

# In another terminal, send a request
curl -X POST http://localhost:8080/execute \
  -H "Content-Type: application/json" \
  -d '{"language": "python", "source_code":"import time; time.sleep(60)"}'

# Press Ctrl+C in the server terminal
# Server will wait up to 30 seconds for in-flight requests to complete
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

## License

MIT License - feel free to use this in your projects!
