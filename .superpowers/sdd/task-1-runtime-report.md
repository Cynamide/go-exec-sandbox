# Files Changed

- `internal/benchmark/runtime.go`
- `internal/benchmark/runtime_test.go`
- `.superpowers/sdd/task-1-runtime-report.md`

# Tests Run

- `go test ./internal/benchmark -run 'TestRunTaskAppliesScaffoldPromptPrefix|TestRunTaskWithGraderUsesInjectedGrader' -v`
- `go test ./internal/benchmark -v`

# Results

- Restored the Task 1 brief-required `RunTask(task, scaffold, mode, client, exec, cfg)` signature.
- Exported the shared grading flow as `RunTaskWithGrader(...)` so later adapters can inject custom grading behavior directly while `RunTask(...)` remains the brief-required public entrypoint.
- Preserved scaffold prompt prefixing, code generation, execution, output recording, and per-test outcome recording behavior.
- Kept the prompt-prefix assertion on `RunTask` itself and updated the focused helper test so it verifies the exported helper uses the injected grader and receives the expected task, response, and test case.

# Self-Review

- Kept the write scope to the requested runtime, runtime test, and task report files only.
- Changed the runtime seam surgically so later adapters can share the same flow through the exported helper without altering the Task 1 public API or existing scaffold/execution behavior.
- Verified the public wrapper and helper seam first with targeted tests, then reran the benchmark package coverage for regression safety.

# Concerns

- No known functional concerns in the updated scope. Later adapters should call `RunTaskWithGrader(...)` when they are ready to inject non-stdout grading behavior.
