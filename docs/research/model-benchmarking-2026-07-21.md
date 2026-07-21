# Model Benchmarking Landscape, 2026-07-21

Scope: primary sources only. I used benchmark papers and first-party benchmark docs where available, then mapped the findings against [the benchmark manifest spec](/Users/arjit/Documents/go-exec-sandbox/docs/superpowers/specs/2026-07-17-benchmark-config-design.md).

Note: the coverage labels below capture the gaps found during the research pass. The benchmark manifest spec was then revised using these findings, so this note should be read as the audit input rather than the final post-revision coverage state.

## Bottom Line

Inference from the source set: the benchmark center of gravity in 2026 is no longer static QA. It is executable, live-updatable, tool-using work with explicit verification.

The strongest benchmark families now emphasize:

- real deliverables instead of toy answers
- long-horizon tool use
- fresh or live data to reduce contamination
- verifiers, rubrics, and executable tests
- process traces as a complement to final-answer scoring

The manifest spec already models a lot of that, but it is still strongest on classic repository repair and artifact scoring. It is only partially prepared for the newer benchmark shapes that dominate 2026.

## Mode Map

### 1. Software Engineering

What matters in 2026:

- Repository repair is still a core benchmark shape, but the frontier has widened.
- Current systems evaluate long-horizon fixes, enterprise integration, observability/debugging, and live-updated repositories.
- The field is moving toward verified solutions that accept any correct implementation, not just the reference patch’s exact trajectory.

Representative primary sources:

- [SWE-bench Goes Live!](https://arxiv.org/abs/2505.23419)
- [SWE-Bench Pro](https://arxiv.org/abs/2509.16941)
- [APEX-SWE](https://arxiv.org/abs/2601.08806)
- [DeepSWE](https://arxiv.org/abs/2607.07946)

Spec coverage: **well for classic repo-patch benchmarks, partial for the broader 2026 SWE landscape**.

Why:

- The manifest directly models `task_mode: repo_patch`, `git_diff`, `tests_pass`, `repo_seed`, and process checks for repo inspection and command execution. See [repo-patch task example](/Users/arjit/Documents/go-exec-sandbox/docs/superpowers/specs/2026-07-17-benchmark-config-design.md#L630), [outcome/process checks](/Users/arjit/Documents/go-exec-sandbox/docs/superpowers/specs/2026-07-17-benchmark-config-design.md#L672), and [process evidence rules](/Users/arjit/Documents/go-exec-sandbox/docs/superpowers/specs/2026-07-17-benchmark-config-design.md#L1609).
- That is a strong fit for SWE-bench-style bug fixing.
- It is weaker for 2026-style engineering tasks that rely on telemetry, cloud primitives, multi-system integration, or more general verifier-based grading.

### 2. Browser / Web

What matters in 2026:

- Browsing benchmarks now stress persistence, live web interaction, and failure-mode robustness.
- The benchmark surface is expanding from controlled web tasks to open-web search, adversarial web environments, and step-level failure analysis.

Representative primary sources:

- [BrowserGym Ecosystem for Web Agent Research](https://arxiv.org/abs/2412.05467)
- [BrowseComp](https://arxiv.org/abs/2504.12516)
- [BrowserArena](https://arxiv.org/abs/2510.02418)
- [SecureWebArena](https://arxiv.org/abs/2510.10073)

Spec coverage: **partial**.

Why:

- The manifest does have `task_mode: browser_agent`, a browser scaffold, browser fixtures, screenshot capture, and `browser_state` outputs. See [browser tool contract](/Users/arjit/Documents/go-exec-sandbox/docs/superpowers/specs/2026-07-17-benchmark-config-design.md#L309), [browser task example](/Users/arjit/Documents/go-exec-sandbox/docs/superpowers/specs/2026-07-17-benchmark-config-design.md#L701), and [browser output validation](/Users/arjit/Documents/go-exec-sandbox/docs/superpowers/specs/2026-07-17-benchmark-config-design.md#L1710).
- That is enough for structured browser tasks and visible-state grading.
- It does not yet model the richer open-web action space, browser security evaluation, or live user-submitted arena style tasks that now matter in the browser benchmark literature.

### 3. Multimodal / VLM

What matters in 2026:

- The benchmark focus has shifted from generic multimodal QA to harder vision-centric reasoning.
- Strong recent benchmarks stress charts, tables, document understanding, visual anomalies, and cross-modal consistency.

Representative primary sources:

- [MMMU-Pro](https://arxiv.org/abs/2409.02813)
- [WikiMixQA](https://arxiv.org/abs/2506.15594)
- [VIA-Bench](https://arxiv.org/abs/2602.01816)
- [WorldBench](https://arxiv.org/abs/2606.06538)

Spec coverage: **partial**.

Why:

- The manifest does mark model capability with `multimodal: true`, and it supports attachments plus browser screenshots. See [model capability flags](/Users/arjit/Documents/go-exec-sandbox/docs/superpowers/specs/2026-07-17-benchmark-config-design.md#L171) and [browser/attachment inputs](/Users/arjit/Documents/go-exec-sandbox/docs/superpowers/specs/2026-07-17-benchmark-config-design.md#L723).
- But there is no first-class image-centric task mode, no explicit multimodal input schema, and no multimodal-specific grading model beyond generic artifact or browser-state checks.
- So the spec can host some multimodal work, but it does not yet express the full VLM benchmark shape cleanly.

### 4. Terminal / Code Execution

What matters in 2026:

- Terminal-style agents are now a distinct benchmark category, not just a coding adjunct.
- The frontier includes long-horizon shell work, general-purpose terminal use, and execution-verified code tasks.

Representative primary sources:

- [LiveCodeBench](https://arxiv.org/abs/2403.07974)
- [Terminal-Bench: Benchmarking Agents on Hard, Realistic Tasks in Command Line Interfaces](https://arxiv.org/abs/2601.11868)
- [Long-Horizon-Terminal-Bench](https://arxiv.org/abs/2607.08964)
- [TUA-Bench](https://arxiv.org/abs/2606.28480)

Spec coverage: **partial**.

Why:

- The manifest has `code_exec`, a `shell_exec` tool, `exec_result`, and `exit_code` channels. See [shell execution tool](/Users/arjit/Documents/go-exec-sandbox/docs/superpowers/specs/2026-07-17-benchmark-config-design.md#L336) and [output channels](/Users/arjit/Documents/go-exec-sandbox/docs/superpowers/specs/2026-07-17-benchmark-config-design.md#L1257).
- That covers single-run execution and shell-based work.
- It does not yet model a persistent interactive terminal session, long-horizon episode structure, or dense intermediate rewards, which are important in the newer terminal benchmark family.

### 5. Tool-Using Agents

What matters in 2026:

- Tool use is now a central benchmark axis, not a side feature.
- Current benchmarks stress multi-turn tool workflows, dynamic user interaction, domain rules, and unreliable tool environments.

Representative primary sources:

- [AgentBench](https://arxiv.org/abs/2308.03688)
- [tau-bench](https://arxiv.org/abs/2406.12045)
- [ToolBench-X](https://arxiv.org/abs/2606.25819)
- [An Executable Benchmarking Suite for Tool-Using Agents](https://arxiv.org/abs/2605.11030)

Spec coverage: **partial**.

Why:

- The manifest has a declarative tool registry, allow-lists, required-tools rules, and process-event accounting. See [tool registry](/Users/arjit/Documents/go-exec-sandbox/docs/superpowers/specs/2026-07-17-benchmark-config-design.md#L220), [tool policy](/Users/arjit/Documents/go-exec-sandbox/docs/superpowers/specs/2026-07-17-benchmark-config-design.md#L363), and [process checks](/Users/arjit/Documents/go-exec-sandbox/docs/superpowers/specs/2026-07-17-benchmark-config-design.md#L1423).
- That is a solid base for tool-using agents.
- But the spec does not yet capture tool-environment unreliability, recovery paths, or a native user-agent conversation loop the way the current tool-use benchmark literature does.

### 6. Scientific / Data Tasks

What matters in 2026:

- Scientific and data benchmarks are increasingly executable and workflow-level, not just question-answering.
- The strongest tasks look like ML engineering, data analysis, spreadsheet workflows, and stepwise scientific discovery with verifiers.

Representative primary sources:

- [MLE-bench](https://arxiv.org/abs/2410.07095)
- [ScienceAgentBench](https://arxiv.org/abs/2410.05080)
- [SciAgentArena](https://arxiv.org/abs/2606.12736)
- [SpreadsheetBench 2](https://arxiv.org/abs/2606.29955)

Spec coverage: **partial**.

Why:

- The manifest can express code execution, artifact outputs, directory outputs, CSV/JSON/markdown checks, and repo-based workflows. See [outcome grading types](/Users/arjit/Documents/go-exec-sandbox/docs/superpowers/specs/2026-07-17-benchmark-config-design.md#L1367) and [expected file artifacts](/Users/arjit/Documents/go-exec-sandbox/docs/superpowers/specs/2026-07-17-benchmark-config-design.md#L1683).
- That works for many analysis tasks.
- It still lacks spreadsheet-native task modeling, notebook-oriented execution, and scientific workflow constructs such as multi-step verification with domain-specific state transitions.

### 7. Process vs Outcome Grading

What matters in 2026:

- The field is converging on a hybrid: final deliverable scoring plus trace-based or rubric-based process scoring.
- The most informative benchmarks now separate what the agent produced from how it got there.

Representative primary sources:

- [AgentProcessBench](https://arxiv.org/abs/2603.14465)
- [Evaluating Deep Research Agents on Expert Consulting Work: A Benchmark with Verifiers, Rubrics, and Cognitive Traps](https://arxiv.org/abs/2605.17554)
- [Terminal-Bench 2.0](https://arxiv.org/abs/2601.11868)
- [SWE-Bench Pro](https://arxiv.org/abs/2509.16941)

Spec coverage: **partial**.

Why:

- The manifest does distinguish outcome and process grading, defines a process-evidence model, and supports event-order, tool-usage, rubric, and event-count checks. See [process grading types](/Users/arjit/Documents/go-exec-sandbox/docs/superpowers/specs/2026-07-17-benchmark-config-design.md#L1423) and [process evidence model](/Users/arjit/Documents/go-exec-sandbox/docs/superpowers/specs/2026-07-17-benchmark-config-design.md#L1611).
- That is much better than a pure answer-only schema.
- But it still stops short of dense step annotations, partial-credit trajectory rewards, and richer verifier-plus-rubric compositions that are becoming common in 2026.

## Spec-Level Summary

Best-covered modes:

- classic repository repair
- artifact-first outcome grading
- basic browser-state tasks
- bounded tool-use workflows

Only partially covered:

- broader software engineering beyond repo patching
- open-web browsing
- multimodal / VLM work
- terminal-use agents
- scientific and data workflows
- process-vs-outcome grading

## Why The Gaps Matter

The current benchmark market is increasingly rewarding benchmarks that can:

- verify execution, not just compare text
- measure long-horizon tool use
- support live or refreshed task streams
- score process quality separately from final output
- represent artifact types that are not plain text files

The manifest spec is already moving in that direction, but the missing pieces are now the same pieces the 2026 benchmark literature keeps stressing: richer multimodal inputs, terminal session state, spreadsheet/scientific workflows, and denser process scoring.

## Sources

Primary sources used for this note are linked inline above. They are all first-party benchmark papers or official benchmark ecosystem papers, not secondary summaries.
