# Reporting And Trace Capture Spec

## Problem

The broad config design expects reports to include benchmark metadata, selected models, task modes, sample IDs, fixture splits, traces, captured artifacts, tool events, process evidence, and richer metrics. The current report aggregates baseline/scaffold success rates and lift across tasks, families, and scaffolds.

Durable run storage, resume, cache, prediction import/export, replay, and rescore are handled separately in `docs/superpowers/specs/2026-07-21-run-persistence-interchange-spec.md`; this spec owns the report and trace data model.

## Config Fields Covered

- `runtime_defaults.capture`
- `runtime_defaults.normalization`
- `tools[*].outputs.emits_events`
- `tools[*].process_signals`
- `scaffolds[*].trace_capture`
- `tasks[*].outputs.primary`
- `tasks[*].outputs.side_effects`
- `tasks[*].grading.outcome.metrics`
- `tasks[*].grading.process.checks`
- `benchmark` metadata fields

## Current Code State

- `internal/benchmark/report.go` reports success rates and lift.
- `benchmark.Run` stores task ID, mode, scaffold, pass/fail, outcomes, output, and error.
- No trace event type exists.
- No artifact capture model exists.
- No model/sample/fixture/split metadata exists in reports.
- No report export or leaderboard policy exists.

## Required Behavior

- Add structured run identity: run ID, model ID, task ID, scaffold ID, task mode, sample ID, attempt, epoch, and fixture split.
- Add trace-event capture with timestamps, event names, source, payload, and redaction policy.
- Add captured artifact references for stdout, stderr, exit code, git diff, generated files, browser state, screenshot, notebook state, spreadsheet state, attachments, and model outputs.
- Add per-check grading results to reports.
- Add configured metric aggregation.
- Include benchmark metadata and release policy in report headers.
- Preserve current lift summaries while adding richer report sections.
- Ensure deterministic ordering of report arrays for diff-friendly outputs.

## Validation Rules

- Reject report generation when a required captured artifact is missing.
- Reject process checks whose required events are unavailable in the trace.
- Redact or omit captured fields disabled by runtime capture config.
- Reject public report export when release policy forbids publishing private test answers.

## Acceptance Criteria

- Existing JSON report fields remain available for current consumers.
- New report sections expose model, sample, fixture, task-mode, grading, and trace metadata.
- Trace data is sufficient for process grading without re-running tasks.
- Reports can distinguish baseline-vs-scaffold lift per model, not only for one selected model.
