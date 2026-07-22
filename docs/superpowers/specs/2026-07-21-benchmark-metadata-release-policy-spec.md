# Benchmark Metadata And Release Policy Spec

## Problem

The broad config design defines benchmark identity, suite versioning, ownership, tags, manifest versioning, contamination controls, refresh policy, and leaderboard release policy. The current code ignores this entire component and only loads runtime/model/task/scaffold data.

## Config Fields Covered

- `benchmark.id`
- `benchmark.title`
- `benchmark.description`
- `benchmark.suite_version`
- `benchmark.owners`
- `benchmark.tags`
- `benchmark.manifest_version`
- `benchmark.release_policy`
- `benchmark.release_policy.benchmark_type`
- `benchmark.release_policy.contamination_control`
- `benchmark.release_policy.leaderboard`

## Current Code State

- `internal/manifest/manifest.go` has no `benchmark` section.
- `internal/benchmark/report.go` does not include benchmark identity or suite metadata.
- The runtime cannot distinguish official, experimental, public-dev, gated-test, or private-holdout suites.

## Required Behavior

- The manifest loader must parse and validate a top-level `benchmark` block.
- `benchmark.id` must be stable, lowercase, and safe for filenames and report keys.
- `benchmark.suite_version` must be required when release policy is configured.
- Owner entries must include at least one contactable owner.
- Tags must be normalized to unique lowercase identifiers.
- Release policy must be represented as metadata in benchmark reports.
- Contamination-control settings must be parsed and validated before any private or gated split is used.
- Leaderboard policy must be available to downstream report/export code.

## Validation Rules

- Reject missing `benchmark.id` once the section is introduced.
- Reject duplicate owners with the same `team` and `contact`.
- Reject `benchmark.release_policy.benchmark_type` values outside `official`, `experimental`, and `internal`.
- Reject contamination control if `enabled: true` and no `strategy` is supplied.
- Reject `private_holdout_fraction` outside `0.0` to `1.0`.
- Reject `freshness_window_days` less than `1`.

## Data Model Shape

```go
type BenchmarkMetadata struct {
	ID              string
	Title           string
	Description     string
	SuiteVersion    string
	ManifestVersion string
	Owners          []BenchmarkOwner
	Tags            []string
	ReleasePolicy   ReleasePolicy
}
```

## Acceptance Criteria

- `manifest.Load` returns benchmark metadata in the loaded manifest object.
- Reports include benchmark ID, suite version, manifest version, and release policy metadata.
- Existing minimal manifests continue to load until this component is explicitly made required in `benchmark.yaml`.
- Invalid release-policy fields fail during manifest load, not during benchmark execution.
