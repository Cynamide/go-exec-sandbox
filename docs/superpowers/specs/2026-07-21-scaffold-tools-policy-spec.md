# Scaffold Tools And Policy Spec

## Problem

The broad config design treats scaffolds as behavioral contracts with system instructions, prompt prefix/suffix, tool policy, input/output contracts, and trace capture. The current code only prepends `prompt_prefix` and carries `tools` as metadata in reports.

Solver and agent pipeline composition is intentionally handled separately in `docs/superpowers/specs/2026-07-21-solver-agent-composition-spec.md`; this spec owns scaffold contracts and tool policy.

## Config Fields Covered

- `tools`
- `tools[*].kind`
- `tools[*].description`
- `tools[*].implementation`
- `tools[*].inputs.schema`
- `tools[*].outputs`
- `tools[*].side_effects`
- `tools[*].process_signals`
- `tools[*].reliability_profile`
- `scaffolds[*].system_instructions`
- `scaffolds[*].prompt_prefix`
- `scaffolds[*].prompt_suffix`
- `scaffolds[*].tool_policy`
- `scaffolds[*].input_contract`
- `scaffolds[*].output_contract`
- `scaffolds[*].trace_capture`

## Current Code State

- `benchmark.Scaffold.ApplyPrompt` only concatenates `prompt_prefix`.
- No tool registry exists.
- No tool-call trace model exists.
- `tools` on a scaffold is not enforced.
- There is no distinction between allowed, required, and disabled tools.

## Required Behavior

- Load a top-level tool registry from the manifest.
- Represent scaffold policy as executable constraints, not just labels.
- Apply system instructions, prompt prefix, and prompt suffix when building model requests.
- Enforce disabled tools for baseline scaffolds.
- Enforce allow-list, deny-list, unrestricted, and required-tool rules for scaffolded runs.
- Track tool calls as trace events.
- Enforce max calls per tool.
- Validate input and output contracts before a task starts.
- Use trace capture settings to decide which events are retained for grading and reporting.

## Validation Rules

- Reject scaffolds that reference unknown tools.
- Reject baseline scaffolds with required tools.
- Reject `tool_policy.mode` values outside `disabled`, `allow_list`, `deny_list`, and `unrestricted`.
- Reject `required_tools` when `tool_policy.mode: disabled`.
- Reject required tools that are not also allowed when `tool_policy.mode: allow_list`.
- Reject tools listed in both `allowed_tools` and `denied_tools`.
- Reject `denied_tools` entries when `tool_policy.mode` is not `deny_list`.
- Reject `max_calls_per_tool` entries for unknown tools.
- Reject tool definitions without input schemas.
- Reject side-effectful tools under `approval_policy: never` unless explicitly permitted by the task mode.

## Acceptance Criteria

- A scaffold can require explicit planning evidence in trace events.
- Baseline runs cannot call tools.
- Tool-assisted runs produce tool-call evidence that process grading can consume.
- Current prompt-prefix-only scaffolds remain valid as metadata-light scaffolds.
