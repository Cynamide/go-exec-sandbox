# Run Persistence And Interchange Spec

## Problem

Serious benchmark systems need resumable runs, cache policy, retry-failed workflows, prediction-file import, result export, replay, and rescore. The current code prints a JSON report for one run and does not persist enough state to resume or rescore.

## Config Fields Covered

- run ID
- output directory
- cache policy
- retry-failed policy
- resume policy
- prediction import format
- result export format
- log replay policy
- rescore policy
- external publishing targets

## Current Code State

- `cmd/evaluator` prints benchmark JSON to stdout for CLI mode.
- `internal/benchmark/report.go` builds in-memory reports.
- There is no result store.
- There is no prediction-file import.
- There is no cache key or resume marker.
- There is no rescore command.

## Required Behavior

- Add durable run directories with manifest snapshot, resolved run matrix, model outputs, traces, artifacts, scores, and final report.
- Add cache keys based on manifest version, model ID, task ID, scaffold ID, fixture sample ID, solver ID, and runtime settings.
- Add resume behavior that skips completed valid samples and reruns incomplete or invalid samples.
- Add retry-failed behavior distinct from max attempts.
- Add prediction import for repo-patch and text-output benchmark styles.
- Add result export formats for JSON, JSONL, and leaderboard submissions.
- Add rescore from persisted run artifacts.

## Validation Rules

- Reject resume when manifest snapshot is incompatible with cached run state.
- Reject prediction imports with missing task IDs or incompatible model IDs.
- Reject result export that would publish hidden answers against release policy.
- Reject rescore when required raw artifacts were not captured.

## Acceptance Criteria

- A long benchmark run can be interrupted and resumed safely.
- External model predictions can be scored without rerunning inference.
- Scores can be recomputed from logs when grader config changes but raw artifacts are sufficient.
