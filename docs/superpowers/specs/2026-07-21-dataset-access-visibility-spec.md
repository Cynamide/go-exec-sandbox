# Dataset Access And Visibility Spec

## Problem

Benchmarks with public, gated, private, or cloud-mediated splits need explicit policy for who can access data, which answers are visible, what can be logged, and what can be published. The broad release policy covers contamination at a high level; this spec owns concrete access and answer-visibility semantics.

## Config Fields Covered

- fixture split visibility
- answer visibility by split
- gated dataset access
- private evaluator handoff
- hidden metadata
- publishability of model outputs, judgments, and traces
- canary or leakage sentinel metadata
- annotator metadata visibility

## Current Code State

- `benchmark.yaml` has no split access policy.
- Fixture data is either inline or local files.
- Reports include run outputs without visibility filtering.
- There is no private evaluator mode.

## Required Behavior

- Add data access policy per fixture split.
- Represent answer visibility as public, local-only, evaluator-only, or hidden.
- Prevent hidden answers from appearing in public reports.
- Support gated fixtures whose data location requires credentials or manual provisioning.
- Support private evaluator handoff where predictions are exported but answers are not available locally.
- Track canary/leakage sentinel metadata for benchmark contamination controls.

## Validation Rules

- Reject public report export when it includes hidden answers.
- Reject local scoring of evaluator-only fixtures.
- Reject gated fixture use when required access token environment variables are absent.
- Reject canary-bearing task files without canary metadata.
- Reject private traces that include hidden answer fields.

## Acceptance Criteria

- Public-dev, gated-test, and private-holdout data can coexist in one benchmark suite.
- The harness can produce safe public reports without leaking answers.
- Private evaluation workflows can score imported predictions without exposing labels to the model runner.
