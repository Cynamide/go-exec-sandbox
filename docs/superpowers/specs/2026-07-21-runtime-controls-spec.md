# Runtime Controls Spec

## Problem

The broad config design includes runtime-wide controls for task mode defaults, language defaults, attempts, concurrency, epochs, random seed, working directory, sandbox profile, approval policy, fail policy, sample limits, capture settings, and normalization. The current code only uses `runtime_defaults.timeout_ms` from the manifest and keeps language images and memory in code.

## Config Fields Covered

- `runtime_defaults.task_mode`
- `runtime_defaults.language`
- `runtime_defaults.timeout_ms`
- `runtime_defaults.max_attempts`
- `runtime_defaults.concurrency`
- `runtime_defaults.epochs`
- `runtime_defaults.random_seed`
- `runtime_defaults.working_directory`
- `runtime_defaults.sandbox_profile`
- `runtime_defaults.approval_policy`
- `runtime_defaults.fail_policy`
- `runtime_defaults.sample_limits`
- `runtime_defaults.capture`
- `runtime_defaults.normalization`

## Current Code State

- `internal/manifest/manifest.go` only maps `timeout_ms` into `config.Config.DefaultTimeoutMS`.
- `internal/config/config.go` owns language Docker image mappings and memory limits.
- `internal/benchmark/service.go` runs tasks sequentially once.
- `internal/benchmark/grader.go` has hardcoded output normalization rules.

## Required Behavior

- Parse runtime defaults into a benchmark runtime configuration object.
- Apply `language` and `task_mode` defaults only when a task omits them.
- Implement max attempts with per-attempt outcomes and final aggregation.
- Implement bounded concurrency while preserving deterministic report ordering.
- Implement epochs as repeated samples with stable sample IDs.
- Seed any randomized fixture sampling from `random_seed`.
- Move sandbox profile selection into manifest-backed config.
- Enforce approval policy before any tool or task mode can perform external or destructive actions.
- Enforce fail policy for sample, task, model, and run-level errors.
- Enforce sample limits for time, messages, tokens, and cost where the adapter can measure them.
- Apply normalization rules before grading.

## Validation Rules

- Reject `concurrency < 1`.
- Reject `epochs < 1`.
- Reject `max_attempts < 1`.
- Reject unknown `sandbox_profile` names.
- Reject `approval_policy` values outside `never`, `on_demand`, and `always`.
- Reject fail-policy combinations that cannot produce a score.
- Reject sample limits that are zero or negative when present.

## Acceptance Criteria

- A manifest can control attempts, concurrency, epochs, sandbox profile, and fail behavior without code edits.
- Sequential single-epoch behavior remains the default for the current manifest.
- Reports contain enough sample identity to distinguish attempts and epochs.
- Runtime normalization is applied consistently before all outcome checks.
