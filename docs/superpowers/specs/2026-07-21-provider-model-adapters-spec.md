# Provider And Model Adapter Spec

## Problem

The broad config design supports multiple provider kinds, model roles, endpoint URLs, auth, params, request mappings, response mappings, and model capabilities. The current code supports only one enabled Ollama model and builds an Ollama client directly.

## Config Fields Covered

- `providers[*].kind`
- `providers[*].api_key_env`
- `providers[*].base_url`
- `providers[*].base_url_env`
- `providers[*].model_lookup`
- `providers[*].transport`
- `models[*].provider`
- `models[*].model_name`
- `models[*].endpoint_url`
- `models[*].enabled`
- `models[*].auth`
- `models[*].params`
- `models[*].request_mapping`
- `models[*].response_mapping`
- `models[*].capabilities`
- `default_model_roles`

## Current Code State

- `internal/manifest/manifest.go` rejects provider kinds other than `ollama`.
- `internal/llm/llm.go` directly wraps the Ollama Go client.
- `internal/benchmark/service.go` assumes one client for all runs.
- `default_model_roles` is not parsed.
- Model capability metadata is not parsed or enforced.

## Required Behavior

- Introduce a provider-neutral model adapter interface that can generate task responses from normalized messages.
- Keep Ollama as one adapter rather than the whole LLM layer.
- Add OpenAI-compatible chat-completions support for local and deployed endpoints.
- Add custom HTTP endpoint support only after request/response mappings are implemented.
- Support multiple enabled models without changing task or scaffold semantics.
- Resolve default model roles such as `judge` and `planner`.
- Validate model capabilities against requested task modes and scaffold tool policies.
- Preserve deterministic params where supported, such as temperature and max tokens.

## Validation Rules

- Reject enabled models that reference missing providers.
- Reject provider kinds without an adapter.
- Reject `api_key_env` when the environment variable is required but absent at runtime startup.
- Reject `endpoint_url` when it conflicts with provider-level transport config.
- Reject model params unsupported by the selected adapter unless the adapter declares passthrough support.
- Reject judge roles pointing at models without `capabilities.judge: true`.
- Reject browser or multimodal tasks assigned only to models lacking the required capability.

## Data Model Shape

```go
type ModelRef string

type ModelAdapter interface {
	Generate(ctx context.Context, req ModelRequest) (ModelResponse, error)
	HealthCheck(ctx context.Context) error
}
```

## Acceptance Criteria

- The benchmark service can run the same task suite across more than one enabled model.
- Ollama behavior remains compatible with the current `benchmark.yaml`.
- OpenAI-compatible local endpoints can be configured by URL without code edits.
- Default model roles are resolved once during manifest load and exposed to grading/tool components.
- Capability mismatches produce manifest-load errors with actionable messages.
