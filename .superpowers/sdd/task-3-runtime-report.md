# Files Changed

- `internal/benchmark/grader.go`
- `internal/benchmark/runtime.go`
- `internal/benchmark/runtime_test.go`

# Tests Run

- `go test ./internal/benchmark -run TestDefaultGraderMarksMatchingStdoutAsPass -v`
- `go test ./internal/benchmark -run TestRunTaskAppliesScaffoldPromptPrefix -v`
- `go test ./internal/benchmark -run TestCodeExecutionAdapterDelegatesToSandbox -v`

# Results

- Added an explicit `DefaultGrader` implementation that preserves the existing stdout-based pass/fail and score behavior.
- Kept `RunTaskWithGrader` as the extensibility seam and updated `RunTask` to use `DefaultGrader` without changing its public behavior.
- Added a focused test that makes the default grading contract explicit and future-adapter-friendly.

# Self-Review

- Confirmed the change keeps the runtime loop intact and does not alter the code execution adapter path.
- Confirmed the new grader contract is exercised directly in tests and the existing scaffold/runtime coverage still passes.
- Kept the write scope limited to the runtime grading files for this task.

# Concerns

- No functional concerns from this slice. The default grader still only considers trimmed stdout equality, which is intentional for the current code-task family but may be too narrow for future task families until they provide a custom grader.
