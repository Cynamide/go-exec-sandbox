# Prompt Rendering And Response Parsing Spec

## Problem

Benchmark systems often separate task data from prompt rendering and answer extraction. The current code builds one hardcoded model prompt and grades the raw generated/executed output. That is too narrow for multiple-choice, exact-answer, agent-action, and artifact benchmarks.

## Config Fields Covered

- task input-to-prompt rendering templates
- task target rendering templates
- choice rendering templates
- scaffold prompt prefix and suffix interaction with task templates
- response parser selection
- answer extraction filters
- action extraction for agent environments

## Current Code State

- `internal/llm/llm.go` uses a hardcoded system prompt and user prompt.
- `benchmark.Scaffold.ApplyPrompt` only prepends `prompt_prefix`.
- `internal/benchmark/grader.go` grades direct stdout/artifact equivalence.
- There is no prompt-template registry.
- There is no parser or action extractor interface.

## Required Behavior

- Add a prompt-rendering module that converts task sample data, scaffold instructions, few-shot examples, and model capability constraints into model requests.
- Add named renderers for direct generation, code generation, multiple-choice, repo patch, browser action, API action, and conversation turns.
- Add response parsers that convert model output into normalized artifacts, answer choices, tool calls, patches, API actions, or browser actions.
- Allow parser outputs to feed grading and trace capture.
- Make prompt rendering deterministic and inspectable in reports.

## Validation Rules

- Reject tasks that require choices but do not define choice rendering.
- Reject parser names without registered implementations.
- Reject prompt templates referencing missing fixture fields.
- Reject action parsers for task modes that do not support actions.
- Reject scaffold prompt contracts that conflict with task prompt rendering.

## Acceptance Criteria

- The same task sample can be rendered through different scaffold strategies without duplicating task data.
- Raw model output and parsed output are both available for trace/reporting.
- Multiple-choice and action-based tasks no longer have to masquerade as code-execution tasks.
