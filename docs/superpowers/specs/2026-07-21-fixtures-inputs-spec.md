# Fixtures And Inputs Spec

## Problem

The broad config design separates task definitions from reusable fixture files, attachments, repo seeds, browser fixtures, multimodal cases, and environment requirements. The current manifest embeds task cases inline and supports only stdin-style inputs for code execution.

Dataset access, gated splits, hidden answers, and publication visibility are handled separately in `docs/superpowers/specs/2026-07-21-dataset-access-visibility-spec.md`; this spec owns fixture resolution and task input materialization.

## Config Fields Covered

- `fixtures`
- `fixtures[*].kind`
- `fixtures[*].path`
- `fixtures[*].format`
- `fixtures[*].split`
- `fixtures[*].environment`
- `tasks[*].split`
- `tasks[*].inputs.sources`
- `tasks[*].inputs.repo_seed`
- `tasks[*].inputs.attachments`
- `tasks[*].inputs.browser_start_state`
- `tasks[*].inputs.environment`

## Current Code State

- `benchmark.yaml` contains inline `test_cases`.
- `internal/manifest/manifest.go` does not parse `fixtures`.
- `internal/benchmark/Task` has no split, fixture refs, attachments, repo seed, browser state, or environment contract.
- `internal/benchmark/executor.go` executes source code with stdin only.

## Required Behavior

- Introduce a fixture registry loaded from the manifest.
- Resolve fixture references into task samples before benchmark execution.
- Support JSON and YAML fixture formats first.
- Preserve inline test cases as a convenience for simple code-execution tasks.
- Add split metadata to tasks and fixtures.
- Add attachment references with kind, path, path-from, and media type.
- Add repo seed fixture resolution for repo-patch tasks.
- Add browser start-state resolution for browser-agent tasks.
- Add required and optional environment variable validation per task.

## Validation Rules

- Reject fixture IDs not referenced by any task only as a warning, not a hard error.
- Reject task input sources that reference missing fixtures.
- Reject fixture paths outside the repository unless explicitly allowed.
- Reject unsupported fixture formats.
- Reject attachment references without `kind`.
- Reject missing required environment variables before task execution.
- Reject task samples with no usable input source.

## Required External Fixture Shapes

```json
{
  "cases": [
    {
      "id": "case-001",
      "input": "raw input",
      "expected_output": "expected output",
      "metadata": {
        "created_at": "2026-07-21T00:00:00Z"
      }
    }
  ]
}
```

## Acceptance Criteria

- The harness can load task cases from external fixture files.
- Inline test cases continue to work for the current manifest.
- Each run records the fixture ID and split it came from.
- Missing fixtures fail during manifest or fixture loading, not halfway through a run.
