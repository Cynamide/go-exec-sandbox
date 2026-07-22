## What I implemented

- Replaced the placeholder request/response mapping shells in `internal/modeladapter/types.go` with transport-oriented fields:
  - `RequestMapping.Method`
  - `RequestMapping.Path`
  - `RequestMapping.BodyTemplate`
  - `ResponseMapping.TextPath`
  - `ResponseMapping.ToolCallsPath`
  - `ResponseMapping.UsagePath`
- Added `modeladapter.ValidateMappings(cfg Config) error` and wired it into `Config.Validate()`.
- Enforced `custom_http` validation so enabled configs must declare:
  - `request_mapping.method`
  - `request_mapping.path`
  - `request_mapping.body_template`
  - `response_mapping.text_path`
- Reordered `manifest.Load` validation so model config validation runs before the legacy Ollama-only runtime selection. This makes missing mapping errors surface before the generic unsupported-provider error.
- Kept live provider support unchanged: non-`ollama` enabled providers are still rejected by `manifest.Load`.
- Updated manifest tests to assert the new mapping schema is carried through loader output.

## TDD Evidence

### RED

Command:

```bash
GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/modeladapter
```

Output:

```text
# gexec-sandbox/internal/modeladapter_test [gexec-sandbox/internal/modeladapter.test]
internal/modeladapter/types_test.go:31:25: undefined: modeladapter.ValidateMappings
FAIL	gexec-sandbox/internal/modeladapter [build failed]
FAIL
```

Why it failed as expected:

- The new test referenced `modeladapter.ValidateMappings`, which did not exist yet.
- That confirmed the test was exercising the missing Task 4 behavior rather than passing accidentally on existing code.

Additional RED check:

Command:

```bash
GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/manifest
```

Output:

```text
--- FAIL: TestLoadRejectsCustomHTTPModelWithoutMappings (0.00s)
    manifest_test.go:357: Load() error = invalid benchmark manifest: provider kind "custom_http" is not supported by the current runtime, want mapping-specific error
FAIL
FAIL	gexec-sandbox/internal/manifest	0.397s
FAIL
```

Why it failed as expected:

- `manifest.Load` was still failing at the generic non-Ollama gate.
- The new test correctly exposed that mapping validation was not happening early enough for `custom_http`.

### GREEN

Focused command from the task brief:

```bash
GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/modeladapter ./internal/manifest
```

Output:

```text
ok  	gexec-sandbox/internal/modeladapter	(cached)
ok  	gexec-sandbox/internal/manifest	(cached)
```

Broader safety check:

```bash
GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./...
```

Output:

```text
ok  	gexec-sandbox/cmd/evaluator	0.330s
ok  	gexec-sandbox/internal/api	(cached)
ok  	gexec-sandbox/internal/benchmark	(cached)
ok  	gexec-sandbox/internal/config	(cached)
ok  	gexec-sandbox/internal/httpapi	(cached)
ok  	gexec-sandbox/internal/llm	0.586s
ok  	gexec-sandbox/internal/manifest	(cached)
ok  	gexec-sandbox/internal/metrics	(cached)
ok  	gexec-sandbox/internal/middleware	(cached)
ok  	gexec-sandbox/internal/modeladapter	(cached)
ok  	gexec-sandbox/internal/sandbox	(cached)
```

## Files changed

- `internal/modeladapter/types.go`
- `internal/modeladapter/types_test.go`
- `internal/manifest/manifest.go`
- `internal/manifest/manifest_test.go`

## Self-review findings

- Kept Task 4 scoped to validation only; I did not add any live `custom_http` or OpenAI-compatible adapter behavior.
- Preserved the existing runtime restriction that enabled non-`ollama` providers are unsupported.
- Chose to require `response_mapping.text_path` for `custom_http` because the current adapter contract is text-generation-oriented; `tool_calls_path` and `usage_path` are represented but remain optional for now.
- Reordered `manifest.Load` only enough to surface model validation before legacy runtime selection. This is intentional and covered by tests.

## Any concerns

- The mapping schema is now aligned to the plan for `custom_http`, but it is not backward-compatible with the earlier placeholder field names. That is acceptable in the current branch because those fields were shells and the repo test suite is green, but Task 5 should build directly on this schema rather than reintroducing compatibility aliases.

## 2026-07-22 Task 4 Review Follow-up

### Scope

- Fixed the `custom_http` missing-mapping error text in `internal/modeladapter/types.go` so required fields are rendered as a comma-separated list.
- Added a focused regression test for that clean error text in `internal/modeladapter/types_test.go`.
- Added a manifest regression test proving a fully mapped `custom_http` model is still rejected as unsupported by the current runtime in `internal/manifest/manifest_test.go`.
- Clarified in the main manifest success test that Ollama models may carry transport mappings even though current supported Ollama execution ignores them.

### RED

Command:

```bash
GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/modeladapter ./internal/manifest
```

Output:

```text
--- FAIL: TestCustomHTTPProviderMissingMappingsErrorListsFields (0.00s)
    types_test.go:54: ValidateMappings() error = "model adapter config \"custom\" custom_http provider requires [request_mapping.path request_mapping.body_template response_mapping.text_path]", want "model adapter config \"custom\" custom_http provider requires request_mapping.path, request_mapping.body_template, response_mapping.text_path"
FAIL
FAIL	gexec-sandbox/internal/modeladapter	0.517s
ok  	gexec-sandbox/internal/manifest	0.842s
FAIL
```

Why this was the right failure:

- The new modeladapter test failed on the exact broken formatting called out in review.
- The new manifest regression already passed in RED, which is desirable here because current runtime rejection of fully mapped `custom_http` models was already correct and needed to stay that way.

### GREEN

Command:

```bash
GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/modeladapter ./internal/manifest
```

Output:

```text
ok  	gexec-sandbox/internal/modeladapter	0.319s
ok  	gexec-sandbox/internal/manifest	0.594s
```

### Files touched for this follow-up

- `internal/modeladapter/types.go`
- `internal/modeladapter/types_test.go`
- `internal/manifest/manifest_test.go`
