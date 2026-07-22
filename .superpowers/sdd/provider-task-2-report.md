# Task 2 Report: Move Ollama Behind The Adapter Seam

## What you implemented

- Added `internal/modeladapter/ollama.go` with `NewOllamaAdapter(cfg Config) (Adapter, error)`.
- Implemented Ollama adapter URL parsing, client construction, `Generate`, and `HealthCheck`.
- Updated `internal/llm/llm.go` so `NewClientWithConfig` constructs the Ollama-backed adapter through the new seam.
- Updated `GenerateCode` to call the provider-neutral adapter interface instead of calling the Ollama client directly.
- Updated `WaitForOllamaWithConfig` to reuse the adapter seam for health checks.
- Added tests covering invalid Ollama host rejection and adapter-based LLM generation behavior.

## TDD Evidence

### RED

Command:

```bash
GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/modeladapter
```

Output:

```text
# gexec-sandbox/internal/modeladapter_test [gexec-sandbox/internal/modeladapter.test]
internal/modeladapter/ollama_test.go:10:25: undefined: modeladapter.NewOllamaAdapter
FAIL	gexec-sandbox/internal/modeladapter [build failed]
FAIL
```

Why it failed as expected:

- The new test referenced `modeladapter.NewOllamaAdapter` before the adapter constructor existed.
- That confirmed the test was exercising the missing seam entrypoint instead of accidentally passing on existing behavior.

### GREEN

Command:

```bash
GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/modeladapter ./internal/llm
```

Output:

```text
ok  	gexec-sandbox/internal/modeladapter	(cached)
ok  	gexec-sandbox/internal/llm	0.420s
```

## Files changed

- `internal/modeladapter/ollama.go`
- `internal/modeladapter/ollama_test.go`
- `internal/llm/llm.go`
- `internal/llm/llm_test.go`

## Self-review findings

- Kept scope aligned with the brief: only Ollama adapter creation and LLM seam wiring changed.
- Left manifest model config resolution, mapping validation, and OpenAI-compatible support untouched for later tasks.
- Preserved existing Ollama-only helper behavior for model pull/list flows; only public LLM construction and wait checks now route through the adapter seam.

## Any concerns

- None.

## Follow-up Fix: Preserve Ollama Client Construction Behavior (July 22, 2026)

### RED

Command:

```bash
GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/modeladapter ./internal/llm
```

Output:

```text
--- FAIL: TestNewOllamaAdapterAllowsEmptyModelName (0.00s)
    ollama_test.go:28: NewOllamaAdapter() error = model adapter config "ollama" missing model name, want nil
FAIL
FAIL	gexec-sandbox/internal/modeladapter	0.442s
--- FAIL: TestNewClientWithConfigAllowsEmptyModelWithExplicitHost (0.00s)
    llm_test.go:70: NewClientWithConfig() error = failed to create Ollama client: model adapter config id is required, want nil
--- FAIL: TestWaitForOllamaWithConfigAllowsEmptyModelWithExplicitHost (0.00s)
    llm_test.go:86: WaitForOllamaWithConfig() error = failed to create Ollama client: model adapter config id is required, want context canceled
FAIL
FAIL	gexec-sandbox/internal/llm	0.821s
FAIL
```

Why it failed as expected:

- `ollamaAdapterConfig` derived both adapter `ID` and `ModelName` from `OLLAMAModel`, so an empty model regressed host-only construction.
- `NewOllamaAdapter` reused generic adapter validation, which incorrectly made model selection mandatory for health-check-only flows.

### GREEN

Command:

```bash
GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/modeladapter ./internal/llm
```

Output:

```text
ok  	gexec-sandbox/internal/modeladapter	0.427s
ok  	gexec-sandbox/internal/llm	0.803s
```

What changed:

- Preserved host-only `llm.NewClientWithConfig` construction by giving the Ollama adapter a stable fallback ID when `OLLAMAModel` is empty.
- Relaxed Ollama adapter construction to validate adapter identity and provider kind without requiring a model up front.
- Kept generation behind `modeladapter.Adapter` and added an explicit runtime guard so chat generation still fails clearly if no model name is configured.

## Task 2 Minor Review Finding Follow-up

Reviewer finding addressed:

- `internal/llm/llm.go` wrapped model adapter construction failures with `failed to create Ollama client`, which was misleading after the adapter seam landed.

Change made:

- Updated only the adapter-construction wrappers in `NewClientWithConfig` and `WaitForOllamaWithConfig` to say `failed to create Ollama adapter`.
- Left the real Ollama API client error wording in `PullModel`, `CheckModelExists`, and `ollamaAPIClient` paths unchanged.

Focused regression coverage:

- Added `TestNewClientWithConfigReportsAdapterConstructionFailures`.
- Added `TestWaitForOllamaWithConfigReportsAdapterConstructionFailures`.

Test evidence:

```bash
GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/llm -run 'Test(NewClientWithConfigReportsAdapterConstructionFailures|WaitForOllamaWithConfigReportsAdapterConstructionFailures)$'
```

```text
ok  	gexec-sandbox/internal/llm	0.454s
```

```bash
GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/llm ./internal/modeladapter
```

```text
ok  	gexec-sandbox/internal/llm	0.751s
ok  	gexec-sandbox/internal/modeladapter	(cached)
```

## Task 2 Minor Review Finding Follow-up: Ollama HealthCheck Seam (July 22, 2026)

Reviewer finding addressed:

- The host-only `WaitForOllamaWithConfig` regression covered adapter construction under cancellation, but it did not directly prove the Ollama adapter's `HealthCheck` path.

Change made:

- Added `TestOllamaAdapterHealthCheckUsesListEndpoint` in `internal/modeladapter/ollama_test.go`.
- The test stays at the seam level by exercising `ollamaAdapter.HealthCheck` with a stubbed Ollama API client transport and asserting it issues `GET /api/tags`.

TDD note:

- This was a coverage-only follow-up on an already-correct code path, so there was no legitimate missing-behavior red case to drive out in production code.
- The first attempted red used `httptest.NewServer` and failed because the sandbox disallows binding a local listener.
- The final in-process seam test then went green once the package-local test harness was wired correctly.

Focused test evidence:

```bash
GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/modeladapter -run 'TestOllamaAdapterHealthCheckUsesListEndpoint$'
```

```text
ok  	gexec-sandbox/internal/modeladapter	0.364s
```

Required package verification:

```bash
GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/modeladapter ./internal/llm
```

```text
ok  	gexec-sandbox/internal/modeladapter	0.550s
ok  	gexec-sandbox/internal/llm	(cached)
```
