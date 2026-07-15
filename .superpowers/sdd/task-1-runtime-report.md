# Files Changed

- `internal/benchmark/runtime.go`
- `internal/benchmark/runtime_test.go`
- `.superpowers/sdd/task-1-runtime-report.md`

# Tests Run

- `go test ./internal/benchmark -run 'TestRunTaskAppliesScaffoldPromptPrefix|TestRunTaskWithGraderUsesInjectedGrader' -v`
- `go test ./internal/benchmark -v`

# Results

- Restored the Task 1 brief-required `RunTask(task, scaffold, mode, client, exec, cfg)` signature.
- Moved the shared grading flow into a private `runTaskWithGrader(...)` helper so later adapters can still inject custom grading behavior.
- Preserved scaffold prompt prefixing, code generation, execution, output recording, and per-test outcome recording behavior.
- Kept the prompt-prefix assertion on `RunTask` itself and added a focused helper test that verifies the injected grader controls pass/fail and receives the expected task, response, and test case.

# Self-Review

- Kept the write scope to the requested runtime, runtime test, and task report files only.
- Changed the runtime seam surgically so later adapters can share the same flow without altering the Task 1 public API or existing scaffold/execution behavior.
- Verified the public wrapper and helper seam first with targeted tests, then reran the benchmark package coverage for regression safety.

# Concerns

- No known functional concerns in the updated scope. Later adapters will need to call the shared helper directly when they are ready to inject non-stdout grading behavior.
