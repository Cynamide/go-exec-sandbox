# Scaffold-Aware Benchmark Harness Design

## Purpose

LocalEval is being positioned as a **scaffold-aware benchmark harness for economically meaningful agent tasks**.

The product thesis is that model quality alone is not the right unit of measurement for many real workloads. The more useful question is:

- what can the model do by itself
- what can it do when wrapped in the right workflow scaffold
- how much lift comes from tools, retrieval, prompts, or verifier loops

This design turns the repository into a benchmark system that measures **task performance under different real-world scaffolds**, rather than a static puzzle benchmark.

## Why This Product Exists

Current benchmark conversations often over-index on isolated model scores or toy tasks. That misses the business value of agent systems, where outcome quality depends on the model plus the surrounding workflow.

This repo should be useful to:

- AI labs evaluating model behavior under realistic conditions
- applied teams deciding what scaffolding actually matters
- researchers studying the difference between raw capability and system capability

The moat is not the task list alone. The moat is the ability to measure the **delta between a bare model and a scaffolded system** on the same economically relevant work.

## Product Statement

**LocalEval is a scaffold-aware benchmark harness that measures how much models improve when given the right tools, prompts, and workflow structure on economically meaningful tasks.**

## Design Principles

1. **Measure lift, not just score**
   - The important output is the improvement from scaffolded runs relative to baseline runs.
   - A benchmark that only reports pass/fail leaves the most valuable signal on the table.

2. **Favor real work over toy tasks**
   - Tasks should resemble deliverables people or companies actually care about.
   - The benchmark should not read like a coding interview practice set.

3. **Compare multiple system modes on the same task**
   - Bare model
   - Model plus prompt scaffold
   - Model plus tools
   - Model plus retrieval or memory
   - Model plus critic or verifier loop

4. **Make the harness reusable**
   - The repository should support more than one benchmark family.
   - The same evaluation pattern should apply across software, browser, spreadsheet, terminal, and document workflows.

5. **Keep the score interpretable**
   - Results should be legible to both researchers and practitioners.
   - Reports should clearly explain what improved, under which scaffold, and by how much.

## Scope

### In Scope

- economically meaningful task evaluation
- scaffold comparison across the same task
- automatic execution where possible
- deterministic or clearly bounded grading where possible
- support for tasks that produce artifacts, not just short answers
- benchmark reporting that separates baseline performance from scaffolded performance

### Out of Scope

- toy puzzle benchmarks
- leaderboard-first design that ignores workflow context
- optimization around narrow interview-style code challenges
- one-off tasks that cannot be compared across scaffold conditions

## Benchmark Model

The core unit of evaluation is a **task run**.

Each task run should capture:

- task identity
- task family
- model used
- scaffold used
- execution context
- output artifact or result
- grading result

The benchmark should support at least these comparisons:

1. **Baseline**
   - the model attempts the task with minimal assistance

2. **Scaffolded**
   - the model gets structured guidance, tools, or workflow support

3. **Delta**
   - the difference between baseline and scaffolded outcomes

The delta is the key product signal because it shows whether a scaffold is genuinely useful.

## Task Families

The repository should support economically meaningful task families such as:

- software engineering
- browser and web workflows
- spreadsheets and documents
- terminal and command-line work
- data analysis and ML engineering
- other agentic tasks that result in a deliverable

These families are broad by design. The benchmark should be able to express different kinds of useful work without becoming domain-specific in a way that limits future expansion.

## Harness Behavior

The existing sandbox and benchmark structure already supports a useful baseline:

- code execution in isolated containers
- test-based verification where possible
- report generation

The product direction is to extend that harness conceptually so it can represent different scaffold modes for the same task set. The harness should be able to answer:

- did the model complete the task?
- did the scaffold help?
- did the scaffold help enough to justify its complexity?

## Reporting

Reports should present results in a way that highlights system value, not just model rank.

The most important report dimensions are:

- baseline success rate
- scaffolded success rate
- lift from scaffold
- per-task-family breakdown
- per-scaffold breakdown

If a report cannot show what a scaffold changed, it is not aligned with this product direction.

## README Positioning

The README should lead with the scaffold-aware thesis, not with LeetCode-style framing.

Recommended positioning:

> LocalEval is a scaffold-aware benchmark harness for economically meaningful agent tasks.

The README should explain that the repo measures:

- model-only performance
- scaffolded performance
- the gap between the two

That wording makes the project feel like a diagnostic benchmark system rather than a generic sandbox.

## What Makes This a Moat

This project becomes differentiated if it can show:

- which workflows unlock performance
- which tools matter most
- where a weaker model becomes viable with better scaffolding
- how much value is created by agent design, not just model choice

That is the strategic value proposition for labs and product teams.

## Open Questions

- Which task family should be the first flagship benchmark?
- How should scaffolds be represented in configuration?
- How much of the initial reporting should be automatic versus human-reviewed?
- Which task families can be graded reliably enough to launch first?

