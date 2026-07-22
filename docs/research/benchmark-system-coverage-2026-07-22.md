# Benchmark System Coverage Audit, 2026-07-22

Scope: compare the component specs in `docs/superpowers/specs/2026-07-21-*.md` against current benchmark-system coverage patterns from first-party docs/source for OpenAI Evals, Inspect, EleutherAI `lm-evaluation-harness`, SWE-bench, BIG-bench, LiveBench, GAIA, WebArena, and MiniWoB++.

Question: do the current specs cover the major benchmarking modes and config surfaces that a scaffold-aware benchmark harness should plan for?

Short answer: mostly yes for scaffold-aware, executable, artifact-producing benchmark harnesses, but not yet for every major benchmark-system surface. The specs cover the core execution spine: model adapters, runtime controls, fixtures, task selection/few-shot, scaffold tool policy, task-mode dispatch, outcome/process grading, traces/reports, and release metadata. The main missing categories are logprob/perplexity/discriminative evaluation, benchmark suite/group aggregation config, prompt rendering and answer extraction, solver/agent pipelines distinct from scaffolds, environment lifecycle for interactive worlds, run resume/cache/retry controls, and data-access/answer-visibility policy.

## Source-Derived Coverage Patterns

### OpenAI Evals

First-party pattern: an eval is a dataset plus an eval class/template; registry YAML names eval IDs, split/version, metrics, and data paths; basic templates cover match/includes/fuzzy/JSON matching; model-graded templates use grading prompts; runs accept completion functions; eval sets have threading, timeouts, progress files, and JSONL event logs.

Evidence:

- Build flow, JSONL sample shape, registry YAML, split/version naming, metrics, and model-graded YAML: https://github.com/openai/evals/blob/main/docs/build-eval.md
- Basic and model-graded templates: https://github.com/openai/evals/blob/main/docs/eval-templates.md
- Completion functions, eval sets, thread/timeouts, progress resume, and JSONL event logging: https://github.com/openai/evals/blob/main/docs/run-evals.md

Spec match:

- Covered: fixtures, split/version, model adapters, task selection, grading, judge model roles, trace/reporting.
- Missing or weak: eval-set/suite grouping as a first-class config surface; run progress/resume; completion-function/solver variants as named first-class runnable units distinct from model adapters and scaffolds.

### Inspect

First-party pattern: core components are tasks, datasets, solvers, scorers, and tools; tasks are flexible recipes; runtime overrides include model, generation params, run config, solver, scorer, grader model role, metrics, sample selection, and epochs; tools include built-in bash/python/text-editing/web/computer/MCP; sandboxes are explicit local/docker/custom environments; scoring supports multiple scorers, model grading, metrics, rescore workflows, and perplexity; logs are a durable analysis artifact and can export run config for reproduction.

Evidence:

- Inspect overview and broad modes/tools/sandboxes: https://inspect.aisi.org.uk/
- First-class components: https://inspect.aisi.org.uk/extensions-components.html
- Task override surface: https://inspect.aisi.org.uk/tasks.html
- Scoring surface including model grading, metrics, multiple scorers, rescoring, and perplexity: https://inspect.aisi.org.uk/scoring.html
- Sandboxes: https://inspect.aisi.org.uk/extensions-sandboxes.html
- Logs and run-config export: https://inspect.aisi.org.uk/eval-logs.html and https://inspect.aisi.org.uk/reference/inspect_log.html

Spec match:

- Covered: tasks, tools, scaffold policy, model roles, epochs, sample IDs, grading, trace capture, artifacts, sandbox profile.
- Missing or weak: solver/agent pipeline as a separate composition layer; rescore/replay/export-config workflows; logprob/perplexity scoring; richer sandbox lifecycle/config beyond profile name.

### lm-evaluation-harness

First-party pattern: YAML task configs define dataset path/name/kwargs, train/validation/test/fewshot splits, few-shot config and sampler, prompt rendering (`doc_to_text`, `doc_to_target`, `doc_to_choice`), output type, metrics/aggregation/higher-is-better, tags, groups, aggregate metrics, model backend and model args, batch/device/parallelism, output paths, sample logging, caching, HF Hub logging, and visualizers.

Evidence:

- New task guide for dataset config, splits, few-shot config, prompt rendering, metrics, tags, and groups: https://github.com/EleutherAI/lm-evaluation-harness/blob/main/docs/new_task_guide.md
- README for model backends/API models, batch/parallelism, output path, sample logging, cache, HF Hub, W&B, and Zeno: https://github.com/EleutherAI/lm-evaluation-harness

Spec match:

- Covered: task selection, split metadata, few-shot sampling, model adapters, runtime concurrency, metrics, reports.
- Missing or weak: prompt rendering and answer extraction as explicit config; loglikelihood/multiple-choice/perplexity mode; suite groups and aggregate metric config; cache/resume/result publishing details.

### SWE-bench

First-party pattern: repo issue tasks require patch predictions in JSONL, Docker-based reproducible environments, layered base/environment/instance image caching, patch application, tests, run IDs, max workers, cache levels, cleanup, logs, per-instance results, cloud evaluation, and dataset variants including Lite, Verified, Multimodal, and Multilingual.

Evidence:

- Evaluation guide: https://www.swebench.com/SWE-bench/guides/evaluation/
- Harness reference: https://www.swebench.com/SWE-bench/reference/harness/
- Dataset variants: https://www.swebench.com/SWE-bench/guides/datasets/
- GitHub README: https://github.com/SWE-bench/SWE-bench

Spec match:

- Covered: `repo_patch`, repo seeds, git diff, test commands, Docker-like sandbox profile, max workers/concurrency, reports/logs, multimodal capability flag.
- Missing or weak: cache levels/image lifecycle, cleanup policy, cloud evaluation handoff, explicit prediction-file import/export format.

### BIG-bench

First-party pattern: benchmark has JSON tasks and programmatic tasks; JSON tasks declare metadata, metrics, preferred score, input/target or target-scores examples; programmatic tasks can interact with models over several rounds through generate/logprob APIs; task files include canary strings to reduce contamination; task testing and result submission are part of the workflow.

Evidence:

- BIG-bench README, task creation, JSON/programmatic task modes, metrics/preferred score, target scores, model logprob calls, and canaries: https://github.com/google/BIG-bench

Spec match:

- Covered: benchmark metadata, tags, metrics/preferred score, fixtures, process grading, contamination metadata.
- Missing or weak: canary/data-leak sentinels as task-file metadata; target-scores/logprob evaluation; programmatic multi-round task API independent of browser/tool modes.

### LiveBench

First-party pattern: live benchmark releases new questions over time, uses recent sources to limit contamination, requires objective ground truth, has categories/tasks, release selection, API endpoint/key config, parallel requests, resume/retry, task/category parallelism, error checking, downloaded questions, model answers, and judgments.

Evidence:

- LiveBench README: https://github.com/livebench/livebench

Spec match:

- Covered: release policy, contamination control, suite versioning, model endpoint/auth config, concurrency, max attempts, objective grading, task tags.
- Missing or weak: release selection as an execution-time selector; category/task grouping; resume/retry-failed semantics; public/private availability of latest releases as access policy.

### GAIA

First-party pattern: agentic benchmark with tooling/autonomy, levels, gated validation/test data, private test answers/metadata, attachments across modalities, unambiguous final answers, and dataset columns such as task ID, question, level, final answer, file name/path, and annotator metadata.

Evidence:

- GAIA dataset card: https://huggingface.co/datasets/gaia-benchmark/GAIA/blob/main/README.md

Spec match:

- Covered: tool/scaffold policy, task splits, attachments, private-holdout/contamination policy, metadata, grading.
- Missing or weak: dataset access/gating policy, answer visibility by split, annotator metadata, audio/document multimodal modes beyond image-oriented `visual_reasoning`.

### WebArena

First-party pattern: standalone self-hostable websites; browser env with observation type, viewport, generated per-task config files, auth cookies, environment URLs, reset between evaluations, prompt-based agents with prompt constructor and action extraction, trajectories saved as HTML, and richer browser benchmark infrastructure via BrowserGym/AgentLab.

Evidence:

- WebArena README: https://github.com/web-arena-x/webarena

Spec match:

- Covered: `browser_agent`, browser start state, screenshots/browser state, environment variables, tool policy, trace capture.
- Missing or weak: multi-site environment lifecycle, auth/session fixture management, reset/cleanup guarantees, observation/action-space config, prompt constructor/action parser config, trajectory artifact type.

### MiniWoB++

First-party pattern: Gymnasium/Selenium web-interaction environments with reset/step episodes, observations containing utterance, fields, screenshot, and DOM elements, configurable action spaces, reward, termination/truncation, render modes, and browser driver dependencies.

Evidence:

- MiniWoB++ repository: https://github.com/Farama-Foundation/miniwob-plusplus
- MiniWoB++ usage docs: https://opendilab.github.io/miniwob-plusplus/content/python_usage/

Spec match:

- Covered: browser mode, screenshot/DOM-ish artifacts through browser state, runtime environment requirements.
- Missing or weak: reinforcement-learning-style episode semantics: reward, termination, truncation, action-space config, render mode, deterministic reset/seed, and per-step observations.

## Coverage Matrix

| Category | Current July 21 Specs | Coverage |
| --- | --- | --- |
| Static exact/fuzzy/model-graded QA | Grading, fixtures, task selection, provider adapters | Mostly covered |
| Few-shot selection | Task selection/few-shot, runtime seed | Mostly covered |
| Repo patch / executable code | Task modes, fixtures, grading, runtime, reporting | Mostly covered |
| Browser/web agents | Task modes, scaffold tools, fixtures, reporting | Partially covered |
| Tool-using scaffolds | Scaffold tools/policy, reporting, grading | Mostly covered |
| Multimodal attachments | Fixtures, task modes, provider capabilities | Partially covered |
| Notebook/spreadsheet/document workflows | Task modes names them, reports mention artifacts | Partially covered |
| Process traces | Reporting/trace, grading | Mostly covered |
| Release/contamination metadata | Benchmark metadata/release policy | Partially covered |
| Model/provider surfaces | Provider/model adapters | Mostly covered |
| Runtime attempts/concurrency/epochs | Runtime controls | Mostly covered |
| Logprob/multiple-choice/perplexity | Provider/model adapters and grading do not model logprob APIs | Missing |
| Prompt rendering/answer extraction | Scaffolds have prompt prefix/suffix; tasks lack doc-to-prompt/parser config | Missing |
| Suite/group aggregation | Tags exist; first-class groups and aggregate metrics do not | Missing |
| Solver/agent pipeline composition | Scaffolds exist; solver chains and external agents are not first-class | Missing |
| Environment lifecycle/reset/auth | Some inputs/env fields exist; lifecycle semantics are not first-class | Missing |
| Resume/cache/retry-failed | Attempts/concurrency exist; resumability/cache are not explicit | Missing |
| Access/answer visibility policy | Release policy exists; gated data/private answers are not explicit | Missing |

## Missing Categories To Plan For

1. **Logprob, target-score, and perplexity evaluation.** BIG-bench and `lm-evaluation-harness` both support discriminative/multiple-choice scoring via target probabilities or loglikelihood, and Inspect documents perplexity scoring. The current `ModelAdapter` shape only exposes `Generate`, so this is a real category gap.

2. **Prompt rendering, response parsing, and answer extraction.** `lm-evaluation-harness` makes prompt construction (`doc_to_text`, `doc_to_target`, `doc_to_choice`) central, and WebArena makes prompt constructors/action extractors central. The specs currently spread this across task inputs, scaffold prompt prefix/suffix, and output contracts, but do not name it as a surface.

3. **Suite/group aggregation config.** OpenAI Evals has eval sets, `lm-evaluation-harness` has groups with aggregate metrics, LiveBench has categories/tasks, and BIG-bench has broad task groupings. Tags are not enough if leaderboard or suite scores need macro/micro weighting.

4. **Solver/agent pipeline composition distinct from scaffold policy.** Inspect separates solvers from tasks/tools/scorers, and WebArena separates prompt constructors/action extraction from the browser environment. The current scaffold spec is close, but a scaffold-aware harness should not force every agent strategy into `prompt_prefix` plus tool policy.

5. **Interactive environment lifecycle.** WebArena and MiniWoB++ require environment setup, reset, auth/session state, observation/action-space config, reward/termination, cleanup, and sometimes generated per-example configs. Browser start state and environment variables are necessary but insufficient.

6. **Run resumability, cache policy, and retry-failed policy.** OpenAI Evals eval sets, LiveBench, SWE-bench, and `lm-evaluation-harness` all expose some combination of progress files, cache levels, retry failures, sample logging, and reusable artifacts. The runtime spec has attempts and failure policy, but not resumability/cache as first-class config.

7. **Dataset access and answer-visibility policy.** GAIA gates data and has private test answers/metadata; LiveBench has release availability caveats; SWE-bench Multimodal test evaluation is private/cloud-mediated. The release-policy spec covers contamination in broad terms, but does not yet model gated data, answer visibility by split, or private evaluator handoff.

8. **Multimodal breadth beyond images.** GAIA includes audio, documents, images, and other files. The task-mode spec names `visual_reasoning`, while fixtures allow generic attachments. A scaffold-aware benchmark harness should plan for `multimodal_qa` or attachment-modal capability checks that include document/audio/video, not only image reasoning.

9. **Prediction-file and result interchange formats.** SWE-bench is prediction-file driven; `lm-evaluation-harness` and LiveBench write result/sample files and publish to external hubs; Inspect can export run config from logs. Current reporting covers rich internal reports, but not import/export contracts as a category.

## Specs That Should Be Merged Or Split

Recommended splits:

- Split **Task Modes Spec** into two layers: `Task Kind / Runner Spec` and `Interactive Environment Lifecycle Spec`. Browser/WebArena/MiniWoB-style lifecycle is large enough that keeping it inside a generic mode dispatcher will hide important decisions.
- Split **Scaffold Tools And Policy Spec** into `Tool Registry And Policy Spec` plus `Solver/Agent Composition Spec`. Staff-engineering pushback: overloading "scaffold" to mean prompt decoration, tool policy, solver chains, external agents, and action parsers will make the config harder to reason about.
- Split **Outcome And Process Grading Spec** to add a separate `Scoring Interfaces Spec` for generative, logprob/multiple-choice, perplexity, judge, executable-verifier, and reward/trajectory scoring modes.
- Split **Fixtures And Inputs Spec** to add `Dataset Access And Visibility Spec` if private/gated benchmarks are a target. This should own gated downloads, answer visibility, private split policy, and evaluator-only metadata.
- Split **Reporting And Trace Capture Spec** to add `Run Persistence And Interchange Spec` if resumable/replayable benchmark runs are a target. This should own cache keys, resume semantics, prediction/result import-export, log replay, and rescore workflows.

Recommended merges:

- Merge the implementation ordering content in **Config Gap Component Spec Index** with a short dependency note in each component, or keep the index strictly as an index. It currently mixes catalog, dependency order, and implementation-state summary; that is useful now, but it will drift as specs evolve.
- Do not merge **Benchmark Metadata And Release Policy Spec** into reporting. It feeds reporting, but the benchmark identity/contamination/leaderboard policy is a separate domain concept and should stay separate.

## Bottom Line

The July 21 specs are strong enough as a scaffold-aware execution harness foundation. They should not be treated as complete benchmark-system coverage yet. Before implementation locks in the manifest shape, I would add category-level specs for:

- logprob/perplexity/discriminative scoring
- prompt rendering and output parsing
- suite/group aggregation
- solver/agent composition
- interactive environment lifecycle
- run persistence/cache/resume/interchange
- dataset access and answer visibility
- multimodal attachment breadth

These are planning surfaces, not implementation details, and they recur across multiple first-party benchmark systems.
