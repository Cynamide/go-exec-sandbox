# Provider/Model Adapter Final Re-review Fix Report

## Status

DONE_WITH_CONCERNS

## Implementation Commit

`714ff00` (`Fix provider adapter rereview findings`)

## Fixes

- Added `scaffolded_scaffold` to each `by_model` summary so its scaffolded rate and lift explicitly name the per-model scaffold selection they describe.
- Added a two-model regression where each model prefers a different scaffold while the top-level report selects the deterministic global winner.
- Threaded resolved `default_model_roles` through `BenchmarkService` into `BenchmarkReport` JSON metadata without inventing role-driven grader behavior.
- Added canonical auth validation shared by manifest loading and direct adapter factory construction: `none` rejects `env` and `header`; `bearer_env` requires `env` and rejects `header`.
- Verified `auth: none` suppresses a provider-level API key on generation requests.
- Tightened OpenAI-compatible base/endpoint and Ollama base URL validation to require a hostname and reject userinfo, fragments, and all query strings.
- Corrected OpenAI-compatible endpoint validation errors to identify the endpoint URL rather than the base URL.
- Verified configured OpenAI-compatible endpoint URLs receive generation requests.
- Added `BenchmarkService.HealthCheckModels`, retained each adapter health callback during service construction, and invoked model health checks during production startup under a 15-second context deadline.
- Kept service construction network-free and tested health invocation separately without requiring live endpoints.

## Files Changed

- `cmd/evaluator/main.go`
- `cmd/evaluator/main_test.go`
- `internal/benchmark/report.go`
- `internal/benchmark/report_test.go`
- `internal/benchmark/service.go`
- `internal/benchmark/service_test.go`
- `internal/manifest/manifest.go`
- `internal/manifest/manifest_test.go`
- `internal/modeladapter/ollama.go`
- `internal/modeladapter/ollama_test.go`
- `internal/modeladapter/openai_compatible.go`
- `internal/modeladapter/openai_compatible_test.go`
- `internal/modeladapter/types.go`

## Verification

- `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/modeladapter ./internal/manifest ./internal/benchmark ./cmd/evaluator`
  - PASS: all four requested packages reported `ok`.
- `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./...`
  - PASS: all repository packages reported `ok`.
- `git diff --check`
  - PASS: no whitespace errors.

## Concerns

OpenAI-compatible configurations that provide only a direct generation endpoint still have no separate health endpoint to probe; their existing adapter health check remains a no-op because probing an inference URL would invent unsafe provider semantics.
