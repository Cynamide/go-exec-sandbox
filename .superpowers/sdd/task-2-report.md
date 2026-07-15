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
- Commit created: `cfa676a` (`Expose benchmark run handler`).

---

## Reviewer Fix Follow-Up

### Change Summary

- Extracted `buildMux` from [`cmd/evaluator/main.go`](/Users/arjit/Documents/go-exec-sandbox/cmd/evaluator/main.go) so the evaluator route wiring can be exercised without booting the server.
- Replaced the placeholder `/benchmark/run` service wiring with a real benchmark service assembled in `main` from:
  - tasks loaded from `data/problems.json`
  - a minimal scaffold catalog with one baseline scaffold and one scaffolded variant
  - `llm.NewClient()`
  - `benchmark.NewCodeExecutionAdapter()`
  - `benchmark.DefaultGrader{}`
- Kept the HTTP adapter thin by leaving request handling in [`internal/httpapi/benchmark_handler.go`](/Users/arjit/Documents/go-exec-sandbox/internal/httpapi/benchmark_handler.go) and moving runtime assembly into the evaluator entrypoint.
- Added an integration-style evaluator test at [`cmd/evaluator/main_test.go`](/Users/arjit/Documents/go-exec-sandbox/cmd/evaluator/main_test.go) that sends `POST /benchmark/run` through the real mux and verifies the registered route invokes the benchmark service and returns benchmark JSON.

### Exact Test Commands and Results

1. Red test:

```bash
go test ./cmd/evaluator -run TestBuildMuxRegistersBenchmarkRunRoute -v
```

Result before the fix:

```text
# gexec-sandbox/cmd/evaluator [gexec-sandbox/cmd/evaluator.test]
cmd/evaluator/main_test.go:22:2: undefined: buildMux
FAIL	gexec-sandbox/cmd/evaluator [build failed]
FAIL
```

2. Green test after wiring fix:

```bash
go test ./cmd/evaluator -run TestBuildMuxRegistersBenchmarkRunRoute -v
```

Result:

```text
=== RUN   TestBuildMuxRegistersBenchmarkRunRoute
--- PASS: TestBuildMuxRegistersBenchmarkRunRoute (0.00s)
PASS
ok  	gexec-sandbox/cmd/evaluator	0.509s
```

3. Required focused verification:

```bash
go test ./internal/httpapi ./cmd/evaluator ./internal/benchmark -v
```

Result:

```text
ok  	gexec-sandbox/internal/httpapi	(cached)
ok  	gexec-sandbox/cmd/evaluator	0.345s
ok  	gexec-sandbox/internal/benchmark	(cached)
```

## Final State

### Final Change Summary

- The evaluator now constructs a runnable benchmark service from real repository assets instead of a placeholder service value.
- `buildMux` now accepts the benchmark service as an argument, which made it straightforward to test the `/benchmark/run` registration with a fake service.
- `cmd/evaluator/main_test.go` verifies the real mux routes `POST /benchmark/run` to the benchmark handler and returns benchmark JSON.

### Final Test Commands And Results

1. Focused route test:
   - Command: `go test ./cmd/evaluator -run TestBuildMuxRegistersBenchmarkRunRoute -v`
   - Result: PASS

2. Required package verification:
   - Command: `go test ./internal/httpapi ./cmd/evaluator ./internal/benchmark -v`
   - Result: PASS

### Final Files Changed

- [`cmd/evaluator/main.go`](/Users/arjit/Documents/go-exec-sandbox/cmd/evaluator/main.go)
- [`cmd/evaluator/main_test.go`](/Users/arjit/Documents/go-exec-sandbox/cmd/evaluator/main_test.go)
- [`internal/httpapi/benchmark_handler.go`](/Users/arjit/Documents/go-exec-sandbox/internal/httpapi/benchmark_handler.go)
- [`internal/httpapi/benchmark_handler_test.go`](/Users/arjit/Documents/go-exec-sandbox/internal/httpapi/benchmark_handler_test.go)

### Final Concern Check

- `POST /benchmark/run` is now backed by a runnable benchmark service, but actual end-to-end execution still depends on Ollama and sandbox availability in the runtime environment.
- Final code commit: `384da32` (`Wire benchmark route into evaluator`).
