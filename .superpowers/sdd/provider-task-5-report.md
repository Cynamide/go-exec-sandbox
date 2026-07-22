# Task 5 Report: Add OpenAI-Compatible Adapter

## What I implemented

- Added `internal/modeladapter/openai_compatible.go` with `NewOpenAICompatibleAdapter(cfg Config) (Adapter, error)`.
- Implemented native OpenAI-compatible `/v1/chat/completions` request handling without using `RequestMapping` or `ResponseMapping`.
- Supported:
  - `model_name`
  - `api_key_env` via bearer auth
  - `temperature`
  - `max_tokens`
  - usage extraction from `usage.prompt_tokens`, `usage.completion_tokens`, and `usage.total_tokens`
  - health checks against `/v1/models`
- Added focused adapter tests in `internal/modeladapter/openai_compatible_test.go`.
- Updated manifest loading in `internal/manifest/manifest.go` to admit provider kind `openai_compatible`.
- Preserved current legacy runtime behavior by continuing to choose the primary Ollama runtime model from enabled Ollama models only.
- Added manifest tests covering enabled `openai_compatible` models and continued rejection of unsupported provider kinds.

## TDD Evidence

### RED

Command:

```bash
GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/modeladapter ./internal/manifest
```

Output:

```text
# gexec-sandbox/internal/modeladapter [gexec-sandbox/internal/modeladapter.test]
internal/modeladapter/openai_compatible_test.go:26:18: undefined: NewOpenAICompatibleAdapter
internal/modeladapter/openai_compatible_test.go:59:18: undefined: NewOpenAICompatibleAdapter
internal/modeladapter/openai_compatible_test.go:89:18: undefined: NewOpenAICompatibleAdapter
internal/modeladapter/openai_compatible_test.go:118:18: undefined: NewOpenAICompatibleAdapter
internal/modeladapter/openai_compatible_test.go:148:18: undefined: NewOpenAICompatibleAdapter
internal/modeladapter/openai_compatible_test.go:164:12: undefined: NewOpenAICompatibleAdapter
internal/modeladapter/openai_compatible_test.go:176:12: undefined: NewOpenAICompatibleAdapter
internal/modeladapter/openai_compatible_test.go:193:18: undefined: NewOpenAICompatibleAdapter
internal/modeladapter/openai_compatible_test.go:231:18: undefined: NewOpenAICompatibleAdapter
FAIL    gexec-sandbox/internal/modeladapter [build failed]
```

Why it failed as expected:

- The new tests referenced `NewOpenAICompatibleAdapter` before any production implementation existed, so the red phase correctly failed at compile time for the missing adapter.

### GREEN

Command:

```bash
GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/modeladapter ./internal/manifest
```

Output:

```text
ok      gexec-sandbox/internal/modeladapter   0.398s
ok      gexec-sandbox/internal/manifest       (cached)
```

Additional verification:

Command:

```bash
GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./cmd/evaluator -run TestLoadBenchmarkManifestUsesReusableFixtureFiles
```

Output:

```text
ok      gexec-sandbox/cmd/evaluator   0.320s
```

## Files changed

- `internal/modeladapter/openai_compatible.go`
- `internal/modeladapter/openai_compatible_test.go`
- `internal/manifest/manifest.go`
- `internal/manifest/manifest_test.go`

## Self-review findings

- The adapter uses the native OpenAI-compatible transport directly and does not depend on the older mapping placeholders, which matches the task brief.
- Secrets are read only from environment variables through `api_key_env`; no inline manifest secret support was added.
- `custom_http` remains validation-only and is still rejected for live runtime support.
- Manifest loading now accepts `openai_compatible` model configs while preserving the current evaluator assumption that runtime bootstrap still comes from an enabled Ollama model.

## Any concerns

- The adapter itself is fully implemented and tested, but the current evaluator and `internal/llm` path are still hard-wired to Ollama outside this task’s scoped files. That means `openai_compatible` is now loadable in manifests and constructible via `modeladapter`, but benchmark execution is not yet generalized to select among loaded model adapters at runtime.

---

## Task 5 Review Fix Pass (July 22, 2026)

### Scope fixed

- Allowed `manifest.Load` to succeed when only enabled `openai_compatible` models are present, while leaving legacy Ollama runtime fields empty unless an enabled Ollama model exists.
- Moved `api_key_env` presence validation into `NewOpenAICompatibleAdapter`, so startup fails early when the configured environment variable is missing.
- Added strict validation for `temperature` and `max_tokens` constructor params, with `max_tokens` accepting only integer values.
- Replaced brittle URL path concatenation with safe path joining for base URLs ending in `/v1/`.
- Updated request-shape tests to go through `NewOpenAICompatibleAdapter` so constructor validation is exercised.

### RED

Command:

```bash
GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/modeladapter ./internal/manifest
```

Output:

```text
--- FAIL: TestOpenAICompatibleAdapterRejectsMissingAPIKey (0.00s)
    openai_compatible_test.go:108: NewOpenAICompatibleAdapter() error = nil, want missing API key error
--- FAIL: TestOpenAICompatibleAdapterJoinsPathsWithoutDoubleSlashes (0.00s)
    openai_compatible_test.go:279: path = /v1//chat/completions, want /v1/chat/completions
--- FAIL: TestOpenAICompatibleAdapterRejectsInvalidTemperatureParamType (0.00s)
    openai_compatible_test.go:306: NewOpenAICompatibleAdapter() error = nil, want invalid temperature error
--- FAIL: TestOpenAICompatibleAdapterRejectsInvalidMaxTokensParamType (0.00s)
    openai_compatible_test.go:321: NewOpenAICompatibleAdapter() error = nil, want invalid max_tokens error
FAIL
FAIL    gexec-sandbox/internal/modeladapter  0.511s
--- FAIL: TestLoadAllowsOpenAICompatibleManifestWithoutOllamaModel (0.00s)
    manifest_test.go:421: Load() error = invalid benchmark manifest: at least one enabled ollama model is required
FAIL
FAIL    gexec-sandbox/internal/manifest      0.255s
FAIL
```

Why it failed as expected:

- The openai-compatible adapter still deferred API key validation until request time, still produced a double slash for `/v1/` base URLs, and still accepted invalid `temperature` / non-integral `max_tokens`.
- Manifest loading still rejected manifests that had no enabled Ollama model, even when an enabled `openai_compatible` model was present.

### GREEN

Focused command:

```bash
GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/modeladapter ./internal/manifest
```

Output:

```text
ok      gexec-sandbox/internal/modeladapter   0.643s
ok      gexec-sandbox/internal/manifest       0.391s
```

Full-suite command:

```bash
GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./...
```

Output:

```text
ok      gexec-sandbox/cmd/evaluator   0.336s
ok      gexec-sandbox/internal/api    (cached)
ok      gexec-sandbox/internal/benchmark      (cached)
ok      gexec-sandbox/internal/config (cached)
ok      gexec-sandbox/internal/httpapi        (cached)
ok      gexec-sandbox/internal/llm    0.588s
ok      gexec-sandbox/internal/manifest       (cached)
ok      gexec-sandbox/internal/metrics        (cached)
ok      gexec-sandbox/internal/middleware     (cached)
ok      gexec-sandbox/internal/modeladapter   (cached)
ok      gexec-sandbox/internal/sandbox        (cached)
```
