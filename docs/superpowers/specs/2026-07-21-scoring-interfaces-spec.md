# Scoring Interfaces Spec

## Problem

Outcome grading needs scoring interfaces that match how different benchmark families interact with models: generative text, target-score multiple choice, log-likelihood, rolling perplexity, executable verification, model-judged rubrics, pairwise preference, and trajectory reward. The current code only grades executed stdout against expected output.

## Config Fields Covered

- model response mode
- generative scoring
- target-score scoring
- log-likelihood scoring
- rolling perplexity scoring
- executable verifier scoring
- model-judge scoring
- pairwise preference scoring
- reward and trajectory scoring

## Current Code State

- `internal/benchmark/grader.go` grades execution responses.
- `internal/llm/llm.go` only exposes code generation.
- There is no model logprob interface.
- There is no scorer abstraction separate from checks.

## Required Behavior

- Add scorer interfaces for generative outputs, choice/target scoring, log-likelihood, perplexity, executable verifiers, judges, pairwise preference, and environment rewards.
- Extend model adapters only when a scoring mode requires model-side scoring.
- Keep deterministic exact/artifact checks as non-model scorers.
- Allow one task to report multiple scorer outputs.
- Allow rescoring from stored run artifacts without re-running model inference when possible.

## Validation Rules

- Reject log-likelihood scoring for adapters without logprob support.
- Reject perplexity scoring for non-tokenized model adapters.
- Reject judge scoring without judge model role.
- Reject executable verifier scoring without sandbox or environment.
- Reject reward scoring without an interactive environment lifecycle.

## Acceptance Criteria

- Multiple-choice and perplexity benchmarks are first-class, not hacked into text generation.
- Agent trajectory scoring can coexist with final-answer scoring.
- Scorer outputs can be aggregated by the grading and reporting specs.
