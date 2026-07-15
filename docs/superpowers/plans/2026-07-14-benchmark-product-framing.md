# Benchmark Product Framing and Docs Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Reframe the repository documentation around scaffold-aware benchmarking so the public story matches the new architecture and product thesis.

**Architecture:** Keep the implementation plans and runtime changes separate from the README and context-map updates. The documentation should explain the task/scaffold/report model in plain language and avoid any language that makes the repository sound like a puzzle benchmark or a code-interview practice set.

**Tech Stack:** Markdown docs, `README.md`, `PROJECT_CONTEXT_MAP.md`, existing research note under `docs/research/`.

## Global Constraints

- economically meaningful task evaluation
- scaffold comparison across the same task
- benchmark reporting that separates baseline performance from scaffolded performance
- toy puzzle benchmarks are out of scope
- leaderboard-first design that ignores workflow context is out of scope
- optimization around narrow interview-style code challenges is out of scope

---

### Task 1: Rework the README opening to lead with the scaffold-aware thesis

**Files:**
- Modify: `README.md`

**Interfaces:**
- Consumes: the benchmark vocabulary from the domain, runtime, and reporting plans
- Produces: a README that introduces the repo as a scaffold-aware benchmark platform instead of a sandbox-with-examples repository

- [ ] **Step 1: Write the failing documentation check**

```bash
grep -n "LeetCode-style\|puzzle\|toy" README.md
```

Expected: current README still contains language that needs to be removed or rewritten.

- [ ] **Step 2: Update the introduction**

```md
LocalEval is a scaffold-aware benchmark harness for economically meaningful agent tasks.

It measures:

- model-only performance
- scaffolded performance
- the gap between the two
```

- [ ] **Step 3: Add a short benchmark model explanation**

```md
Each benchmark run records the task, the scaffold, the run mode, and the graded outcome.
That makes it possible to see how much lift comes from tools, prompts, or workflow structure.
```

- [ ] **Step 4: Re-run the documentation check**

Run: `grep -n "LeetCode-style\|puzzle\|toy" README.md`

Expected: no matches.

- [ ] **Step 5: Commit**

```bash
git add README.md
git commit -m "Reframe README around scaffold-aware benchmarking"
```

### Task 2: Update the project context map to match the new architecture

**Files:**
- Modify: `PROJECT_CONTEXT_MAP.md`

**Interfaces:**
- Consumes: the current repository inventory
- Produces: a context map that names the task/scaffold/runtime/report split and the code execution adapter as one backend

- [ ] **Step 1: Write the failing documentation check**

```bash
grep -n "problem\|puzzle\|LeetCode" PROJECT_CONTEXT_MAP.md
```

Expected: the current map still reflects the old problem-centric framing in a few places.

- [ ] **Step 2: Update the identity section**

```md
- Repository name: `gexec-sandbox`
- Primary description in the README: a scaffold-aware benchmark harness for economically meaningful agent tasks
- Main language: Go
- Runtime model: task catalog + scaffold runtime + report aggregation + optional execution adapters
```

- [ ] **Step 3: Update the benchmark section**

```md
- The benchmark package should define the task, scaffold, run, grader, and report contracts
- Code execution should appear as the first adapter, not the core architecture
```

- [ ] **Step 4: Re-run the documentation check**

Run: `grep -n "problem\|puzzle\|LeetCode" PROJECT_CONTEXT_MAP.md`

Expected: no matches in the rewritten sections.

- [ ] **Step 5: Commit**

```bash
git add PROJECT_CONTEXT_MAP.md
git commit -m "Update project context for scaffold-aware benchmark architecture"
```

### Task 3: Align the research note with the product thesis

**Files:**
- Modify: `docs/research/model-benchmarking-2026-07-14.md`

**Interfaces:**
- Consumes: the research conclusions already in the repository
- Produces: a research note that supports the scaffold-aware direction rather than the old benchmark framing

- [ ] **Step 1: Write the failing documentation check**

```bash
grep -n "LeetCode-style\|best next additions" docs/research/model-benchmarking-2026-07-14.md
```

Expected: there should be no lingering language that pulls the product back toward puzzle benchmarks or roadmap speculation.

- [ ] **Step 2: Update the conclusion section**

```md
The strongest current benchmark signal is not isolated model quality.
It is economically meaningful task completion under different scaffolds.
```

- [ ] **Step 3: Clarify the moat hypothesis**

```md
The moat is a benchmark system that shows how much tools, prompts, and workflow structure change outcomes on real tasks.
```

- [ ] **Step 4: Re-run the documentation check**

Run: `grep -n "LeetCode-style\|best next additions" docs/research/model-benchmarking-2026-07-14.md`

Expected: no matches.

- [ ] **Step 5: Commit**

```bash
git add docs/research/model-benchmarking-2026-07-14.md
git commit -m "Align research note with scaffold-aware benchmark thesis"
```

## Verification Plan

Run:

```bash
grep -n "LeetCode-style\|puzzle\|toy" README.md
grep -n "problem\|puzzle\|LeetCode" PROJECT_CONTEXT_MAP.md
grep -n "LeetCode-style\|best next additions" docs/research/model-benchmarking-2026-07-14.md
```

Expected: all checks return no matches after the docs are updated.

## Coverage Check

- README framing -> Task 1
- context map alignment -> Task 2
- research note alignment -> Task 3
- public product story -> scaffold-aware benchmark harness, not code-interview puzzle benchmark
