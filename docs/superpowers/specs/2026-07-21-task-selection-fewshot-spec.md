# Task Selection And Few-Shot Spec

## Problem

The broad config design lets each task describe versioning, split, tags, model inclusion, scaffold inclusion, and few-shot sampling. The current code runs every loaded task against the baseline scaffold and every non-baseline scaffold using one enabled model.

## Config Fields Covered

- `tasks[*].version`
- `tasks[*].split`
- `tasks[*].tags`
- `tasks[*].models.include`
- `tasks[*].models.exclude`
- `tasks[*].scaffolds.include`
- `tasks[*].scaffolds.exclude`
- `tasks[*].fewshot.num_examples`
- `tasks[*].fewshot.sampler`
- include/exclude selection semantics
- task groups and subsuite aggregation metadata

## Current Code State

- `internal/manifest/manifest.go` does not parse task version, split, tags, model selection, scaffold selection, or few-shot configuration.
- `internal/benchmark/service.go` runs all tasks for all scaffolds in the loaded catalog.
- There is no per-task model matrix.
- There is no sample or few-shot sampler.
- There is no group or subsuite model for running and reporting selected benchmark slices.

## Required Behavior

- Add task metadata fields for version, split, and tags.
- Add selection resolution that expands configured tasks, models, and scaffolds into a deterministic run matrix.
- Allow task-level model include/exclude filters.
- Allow task-level scaffold include/exclude filters.
- Add benchmark groups that resolve to task sets by explicit task IDs, tags, families, splits, or task modes.
- Add group-level aggregate metric declarations when all included tasks expose the named metric.
- Preserve the invariant that baseline comparisons are available whenever lift is reported.
- Add few-shot configuration to task prompts only when requested.
- Implement `sampler: none` first, then deterministic fixture-backed samplers.
- Seed few-shot sampling with `runtime_defaults.random_seed`.

## Validation Rules

- Reject task version values that are empty when versioning is enabled.
- Reject task split values outside configured or known split names.
- Reject model includes/excludes referencing unknown models.
- Reject scaffold includes/excludes referencing unknown scaffolds.
- Reject a task selection that resolves to zero models.
- Reject a task selection that resolves to zero scaffolds.
- Reject a group selection that resolves to zero tasks.
- Reject group aggregate metrics that are absent from one or more included tasks.
- Reject lift reporting for a task selection that omits the baseline scaffold.
- Reject `fewshot.num_examples < 0`.
- Reject `fewshot.num_examples > 0` when `fewshot.sampler: none`.

## Acceptance Criteria

- The run matrix can represent multiple models, multiple scaffolds, and per-task filters.
- The run matrix can be narrowed by group, tag, family, split, task mode, model, and scaffold.
- The current manifest still resolves to one enabled model, all tasks, one baseline scaffold, and all scaffold variants.
- Few-shot examples are injected deterministically when configured.
- Reports expose enough metadata to explain why each task/model/scaffold combination ran or was skipped.
