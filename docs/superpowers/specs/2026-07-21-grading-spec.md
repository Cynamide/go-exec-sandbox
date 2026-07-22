# Outcome And Process Grading Spec

## Problem

The broad config design separates outcome grading from process grading and supports exact stdout, artifact checks, file checks, test commands, pairwise preference, judge rubrics, required events, event order, allowed tool usage, and metrics. The current grader compares stdout-like outputs against inline expected output and supports markdown, CSV, JSON, and text equivalence.

Low-level scoring interfaces for generative, log-likelihood, perplexity, executable verifier, judge, preference, and reward scoring are handled separately in `docs/superpowers/specs/2026-07-21-scoring-interfaces-spec.md`; this spec owns how configured checks and metrics are evaluated for a task.

## Config Fields Covered

- `tasks[*].outputs`
- `tasks[*].grading.outcome.enabled`
- `tasks[*].grading.outcome.preferred_score`
- `tasks[*].grading.outcome.metrics`
- `tasks[*].grading.outcome.checks`
- `tasks[*].grading.outcome.filters`
- `tasks[*].grading.process.enabled`
- `tasks[*].grading.process.judge_model_role`
- `tasks[*].grading.process.checks`
- rubric fixtures
- preference rubric fixtures

## Current Code State

- `internal/benchmark/grader.go` implements a single `DefaultGrader`.
- `DefaultGrader` grades one execution response and one test case.
- There is no outcome/process split.
- There is no judge model role.
- There is no rubric engine.
- There is no pairwise preference grading.
- There is no configured output filtering or answer extraction pipeline.
- Process-quality evidence is not scored.

## Required Behavior

- Introduce a grading engine that evaluates a run artifact against configured outcome and process checks.
- Keep exact stdout and artifact markdown/CSV/JSON/text checks as first-class outcome checks.
- Add file-based checks for generated files and git diffs.
- Add test-command checks for repo-patch workflows.
- Add output filter pipelines before checks and metrics are evaluated.
- Add metric definitions with aggregation and `higher_is_better`.
- Add process checks for required events, event order, allowed tool usage, minimum event counts, and rubric judgment.
- Add judge-model execution through resolved model roles.
- Add pairwise preference grading for tasks configured with preference rubrics.
- Preserve deterministic non-judge grading wherever possible.

## Output Filter Pipelines

Outcome grading must be able to transform raw model or artifact outputs before scoring. Required filter functions:

- `trim`
- `normalize_whitespace`
- `regex_extract`
- `json_path`
- `take_first`
- `take_first_k`
- `majority_vote`
- `lowercase`
- `strip_markdown_fence`

Rules:

- Filters run in declaration order.
- Each named filter pipeline produces a named candidate output.
- Checks may choose the raw output or a named filtered output.
- Metrics may aggregate over different filter pipelines for the same task.
- Filter failures produce check-level failures unless `fail_policy` says otherwise.

## Validation Rules

- Reject tasks with both outcome and process grading disabled.
- Reject `preferred_score` that does not match a declared metric.
- Reject checks referencing missing fixtures.
- Reject checks referencing unknown filter pipeline names.
- Reject filter functions that are unknown or incompatible with the input format.
- Reject judge-backed checks without a configured judge model role.
- Reject pairwise preference checks without a baseline output source.
- Reject process checks requiring events that are not captured by the scaffold or runtime capture policy.

## Acceptance Criteria

- Current exact-output and artifact-equivalence behavior is represented as configured outcome checks.
- Process grading can be enabled independently of outcome grading.
- A run report includes per-check results, metric values, and aggregate score.
- Judge-backed grading is isolated behind a model adapter and can be skipped in offline/local-only runs when not configured.
