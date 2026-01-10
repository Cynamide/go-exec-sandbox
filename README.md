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

## Prerequisites

- **Go 1.21+** - [Install Go](https://golang.org/dl/)
- **Docker** - [Install Docker](https://docs.docker.com/get-docker/)
- Docker daemon must be running

## Security

This service uses Docker containers to isolate user code execution:

- **Network Disabled**: Containers run with network access disabled to prevent unauthorized network calls
- **Resource Limits**: Memory and CPU quotas restrict resource usage (default: 128MB, 50k CPU quota)
- **Ephemeral Containers**: Containers are automatically removed after execution
- **Context Timeouts**: Execution is enforced with context timeouts to prevent hanging processes

> ⚠️ **Important**: While Docker provides strong isolation, this service should still be run behind additional security layers (rate limiting, authentication, etc.) in production environments.

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

## Adding New Languages

1. Add the Docker image to `Languages` map in `config.go`
2. Update `getCommand()` function in `docker.go` to return the correct execution command
3. Update `getExtension()` function in `docker.go` to return the correct file extension

## Project Structure

```
gexec-sandbox/
├── cmd/
│   └── server/
│       └── main.go          # HTTP server and handlers
├── internal/
│   ├── api/
│   │   └── types.go         # Request/response types
│   ├── config/
│   │   └── config.go        # Configuration management
│   └── sandbox/
│       └── docker.go        # Docker container execution logic
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

## License

MIT License - feel free to use this in your projects!
