# Task 2 Report: Expose Benchmark Service Over HTTP

## What I Implemented

- Added a thin HTTP adapter at `internal/httpapi/benchmark_handler.go`.
- The handler accepts `POST /benchmark/run`, calls `benchmark.BenchmarkServiceAPI.Run`, and JSON-encodes the returned `benchmark.BenchmarkReport`.
- The handler rejects non-`POST` requests with `405 Method Not Allowed`.
- The handler returns service errors as `400 Bad Request`, preserving the existing benchmark-core behavior rather than adding new translation logic.
- Registered the route in `cmd/evaluator/main.go` with minimal structural wiring:
  - `mux.Handle("/benchmark/run", httpapi.BenchmarkRunHandler{...})`
  - The route uses `benchmark.BenchmarkService{Config: cfg}` as the concrete service shape for now.
- Added a focused handler test at `internal/httpapi/benchmark_handler_test.go` that verifies a successful `POST` returns `200 OK` and benchmark-report JSON.

## What I Tested

### Red/Green TDD Check

1. Added `TestBenchmarkRunHandlerReturnsJSON`.
2. Ran:

```bash
go test ./internal/httpapi -run TestBenchmarkRunHandlerReturnsJSON -v
```

Initial result: compile failure as expected because `BenchmarkRunHandler` did not exist yet.

3. Implemented the handler and reran the same test.

Result: PASS.

### Focused Verification

Ran:

```bash
go test ./internal/httpapi ./cmd/evaluator ./internal/benchmark -v
```

Results:
- `internal/httpapi`: PASS
- `cmd/evaluator`: builds successfully (`[no test files]`)
- `internal/benchmark`: PASS, including the existing benchmark service and adapter tests

## Files Changed

- `internal/httpapi/benchmark_handler.go`
- `internal/httpapi/benchmark_handler_test.go`
- `cmd/evaluator/main.go`

## Self-Review Findings

- The HTTP adapter is intentionally thin and does not modify benchmark-core behavior.
- The route registration is minimal and does not introduce task-registry, fixture-loading, or service-construction architecture ahead of the planned tasks.
- The handler test covers the primary success path described in the task brief.

## Issues / Concerns

- There is still no production fixture/registry loader in this task slice, so the current main wiring uses a structurally valid `benchmark.BenchmarkService{Config: cfg}` without task/scaffold/client/executor dependencies.
- As a result, `POST /benchmark/run` is registered and reachable, but it will currently return a service error such as `llm client is required` until the later task wires real benchmark inputs and dependencies into the service.
- I did not add a `cmd/evaluator` route-registration test because there is no existing mux-construction seam in `main`, and the brief asked for minimal wiring only.
