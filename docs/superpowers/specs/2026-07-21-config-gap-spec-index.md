# Config Gap Component Spec Index

## Goal

Break the benchmark configuration surface into implementation-ready component specs by comparing the broad config design in `docs/superpowers/specs/2026-07-17-benchmark-config-design.md` with the currently implemented loader and runtime path.

## Current Implemented Slice

The current runtime supports this manifest-backed surface:

- `schema_version: 1`
- `runtime_defaults.timeout_ms`
- `providers` entries with `kind: ollama`, `base_url`, and `base_url_env`
- exactly one enabled Ollama model with `provider`, `model_name`, and `enabled`
- simple code-execution tasks with `id`, `title`, `description`, `family`, `language`, `artifact_expectation`, and inline `test_cases`
- scaffolds with `baseline`, `description`, `prompt_prefix`, and tool metadata

The implementation entry points are:

- `benchmark.yaml`
- `internal/manifest/manifest.go`
- `cmd/evaluator/main.go`
- `internal/benchmark/service.go`
- `internal/benchmark/grader.go`
- `internal/benchmark/report.go`

## Component Specs

- `docs/superpowers/specs/2026-07-21-benchmark-metadata-release-policy-spec.md`
- `docs/superpowers/specs/2026-07-21-provider-model-adapters-spec.md`
- `docs/superpowers/specs/2026-07-21-runtime-controls-spec.md`
- `docs/superpowers/specs/2026-07-21-fixtures-inputs-spec.md`
- `docs/superpowers/specs/2026-07-21-task-selection-fewshot-spec.md`
- `docs/superpowers/specs/2026-07-21-prompt-rendering-response-parsing-spec.md`
- `docs/superpowers/specs/2026-07-21-scaffold-tools-policy-spec.md`
- `docs/superpowers/specs/2026-07-21-solver-agent-composition-spec.md`
- `docs/superpowers/specs/2026-07-21-task-modes-spec.md`
- `docs/superpowers/specs/2026-07-21-interactive-environment-lifecycle-spec.md`
- `docs/superpowers/specs/2026-07-21-multimodal-attachments-spec.md`
- `docs/superpowers/specs/2026-07-21-scoring-interfaces-spec.md`
- `docs/superpowers/specs/2026-07-21-grading-spec.md`
- `docs/superpowers/specs/2026-07-21-reporting-trace-spec.md`
- `docs/superpowers/specs/2026-07-21-run-persistence-interchange-spec.md`
- `docs/superpowers/specs/2026-07-21-dataset-access-visibility-spec.md`

## Dependency Order

1. Provider and model adapters
2. Runtime controls
3. Fixtures and inputs
4. Task selection and few-shot sampling
5. Prompt rendering and response parsing
6. Scaffold tools and policy
7. Solver and agent composition
8. Task modes
9. Interactive environment lifecycle
10. Multimodal attachments
11. Scoring interfaces
12. Grading
13. Reporting and trace capture
14. Run persistence and interchange
15. Dataset access and visibility
16. Benchmark metadata and release policy

## Non-Goals For These Specs

- Do not change runtime behavior directly.
- Do not expand `benchmark.yaml` until each component has implementation support.
- Do not use these specs as a claim that the described config fields already work.
- Do not add benchmark task data here; task data belongs in manifest entries or fixture files.
