# Task Modes Spec

## Problem

The broad config design supports multiple economically useful benchmark modes such as code execution, direct artifact production, repository patching, browser agents, API agents, visual reasoning, terminal workflows, notebooks, spreadsheets, documents, conversation workflows, and text-only tasks. The current code only supports code-generation followed by sandboxed code execution.

Interactive reset/step/cleanup semantics are handled separately in `docs/superpowers/specs/2026-07-21-interactive-environment-lifecycle-spec.md`; this spec owns task-mode dispatch and runner contracts.

## Config Fields Covered

- `tasks[*].task_mode`
- `tasks[*].language`
- `tasks[*].inputs.repo_seed`
- `tasks[*].inputs.attachments`
- `tasks[*].inputs.browser_start_state`
- `tasks[*].outputs.primary.channel`
- `tasks[*].outputs.primary.format`
- `tasks[*].outputs.side_effects`
- model response mode for generative, choice-scoring, and log-likelihood style tasks
- task-mode-specific fixture kinds

## Current Code State

- `internal/benchmark/service.go` always calls the same code-generation/execution path.
- `internal/benchmark/harness.go` and the benchmark service assume generated code can be executed.
- `internal/sandbox` supports code execution in Docker.
- There are no artifact-only, repo, browser, API, visual, terminal, spreadsheet, notebook, document, conversation, choice-scoring, or log-likelihood runners.

## Required Behavior

- Add a task-mode dispatch layer.
- Keep `code_exec` as the first and default runner.
- Add `artifact_only` runner for tasks where the model returns the scored artifact directly without code execution.
- Add `repo_patch` runner that works on an isolated repo seed and grades git diff plus test commands.
- Add `browser_agent` runner that can initialize browser state and capture visible state/screenshot evidence.
- Add `api_agent` runner for tasks whose main work is HTTP/API interaction and whose primary output may be an HTTP response, trace, or structured artifact.
- Add `visual_reasoning` runner that passes image attachments to capable model adapters.
- Add `text_only` runner for direct natural-language, structured-text, or judge-scored outputs without executor-side action.
- Add `choice_scoring` runner for multiple-choice or target-scoring tasks that compare candidate choices through model probabilities or normalized scores.
- Add `loglikelihood` runner for perplexity, rolling log-likelihood, and completion-likelihood benchmarks when a model adapter supports token-level scoring.
- Add terminal/notebook/spreadsheet/document/conversation modes only after their input and output contracts are represented.
- Validate model capabilities against task mode before execution.
- Make each task mode produce a normalized `RunArtifact` for grading.

## Validation Rules

- Reject unknown task modes.
- Reject `language: none` for `code_exec`.
- Reject code-execution tasks without runnable test cases or artifact expectation.
- Reject `choice_scoring` when no choices or target scores are available.
- Reject `loglikelihood` when the selected model adapter does not support token-level scoring.
- Reject `repo_patch` without a repo seed.
- Reject `browser_agent` without browser start state.
- Reject `api_agent` without API fixture, endpoint fixture, or allowed HTTP tool.
- Reject `visual_reasoning` without at least one image attachment.
- Reject task modes requiring capabilities unavailable in selected models.

## Acceptance Criteria

- `code_exec` behavior remains compatible with the current benchmark.
- Task-mode dispatch is explicit and testable.
- Direct-generation, choice-scoring, and log-likelihood benchmark styles have explicit runner contracts instead of being forced through code execution.
- Each new task mode can be added without modifying unrelated runner internals.
- Reports identify each run's task mode and primary output channel.
