# Solver And Agent Composition Spec

## Problem

Scaffolds describe behavioral support, but full benchmark systems also need solver or agent pipelines: single-shot generation, chain-of-thought prompting, tool-using agents, external coding agents, multi-agent workflows, and action parsers. The current scaffold spec mixes prompt decoration and tool policy but does not define a separate solver seam.

## Config Fields Covered

- solver registry
- solver steps
- solver-to-model binding
- solver-to-scaffold binding
- external agent invocation
- multi-agent role composition
- planner/retriever/judge model roles used during solving
- action parser binding

## Current Code State

- `internal/benchmark/service.go` calls one LLM client and one executor path.
- There is no solver abstraction.
- There is no external-agent runner.
- Tool policy is scaffold metadata only.
- Planner and retriever roles are not parsed or used.

## Required Behavior

- Add a solver interface that owns how a task sample is attempted.
- Keep scaffolds as input/tool/process policy, not the whole solver pipeline.
- Support a default single-shot code-generation solver matching current behavior.
- Add composition for planner-then-act, tool-using agent, external CLI agent, and multi-agent variants.
- Allow solvers to emit trace events and intermediate artifacts.
- Validate solvers against model capabilities, scaffold tool policy, approval policy, and task mode.

## Validation Rules

- Reject solvers referencing unknown models, model roles, tools, parsers, or scaffolds.
- Reject external agents without an explicit sandbox or approval policy.
- Reject multi-agent solvers without role names and role-to-model bindings.
- Reject solver pipelines that can emit outputs incompatible with the task primary channel.

## Acceptance Criteria

- Scaffold-aware comparison can hold the task fixed while varying solver strategy.
- External coding agents can be benchmarked without pretending they are normal chat-completion models.
- Solver traces can be graded as process evidence.
