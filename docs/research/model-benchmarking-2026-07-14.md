# Model Benchmarking Research

Date: 2026-07-14

## Direct Answer

The strongest current benchmark signal is not LeetCode-style puzzle solving. It is **economically valuable real-world task completion by agents**.

The field is converging on benchmarks that measure:

- real deliverables, not toy answers
- tool use and multi-step workflows
- freshness and contamination resistance
- human or economic baselines
- scaffold sensitivity, meaning how much a model improves when wrapped in a better workflow

Software engineering remains an important slice of this, but the broader market includes ML engineering, scientific workflows, spreadsheets, browser use, and desktop/terminal automation.

## What Labs Seem To Prefer Right Now

The current benchmark pattern is:

- **Executable tasks over static QA**
- **Real work products over toy outputs**
- **Live or regularly refreshed sources over saturated datasets**
- **Human, professional, or monetary baselines over only model-vs-model scores**
- **Tool use, browsing, terminal work, and computer use**
- **Scaffold-aware evaluation**

That is why the moat is probably not another static benchmark. It is a **benchmark harness that measures task performance under different real-world scaffolds**.

That kind of harness can answer:

- How much does a bare model improve with a prompt scaffold?
- How much lift comes from tools, retrieval, critic loops, or memory?
- Which tasks are solved by the model alone versus the full agent stack?
- Which skills survive changes in the workflow, and which fall apart?

## Practical Positioning

If you want this project to be useful to a lab or product team, the most attractive shape is:

- a live-updatable benchmark of economically meaningful tasks
- a harness that can run multiple scaffolds against the same tasks
- automatic grading where possible, expert or human grading where needed
- clear reporting of baseline model, scaffolded model, and scaffold delta

That gives you a benchmark, a harness, and a diagnostics layer in one system.

## Primary Sources

- [SWE-bench](https://arxiv.org/abs/2310.06770)
- [SWE-bench Goes Live!](https://arxiv.org/abs/2505.23419)
- [SWE-Lancer](https://arxiv.org/abs/2502.12115)
- [GDPval](https://arxiv.org/abs/2510.04374)
- [HCAST](https://arxiv.org/abs/2503.17354)
- [LiveCodeBench](https://arxiv.org/abs/2403.07974)
- [LiveBench](https://arxiv.org/abs/2406.19314)
- [BrowseComp](https://arxiv.org/abs/2504.12516)
- [τ-bench](https://arxiv.org/abs/2406.12045)
- [Chatbot Arena](https://arxiv.org/abs/2403.04132)
- [The Leaderboard Illusion](https://arxiv.org/abs/2504.20879)
- [MMMU](https://arxiv.org/abs/2311.16502)
- [MMMU-Pro](https://arxiv.org/abs/2409.02813)
- [OSWorld](https://arxiv.org/abs/2404.07972)
- [OSWorld2.0](https://arxiv.org/abs/2606.29537)
- [SpreadsheetBench](https://arxiv.org/abs/2406.14991)
- [SpreadsheetBench 2](https://arxiv.org/abs/2606.29955)
- [MLE-bench](https://arxiv.org/abs/2410.07095)
- [ScienceAgentBench](https://arxiv.org/abs/2410.05080)
- [OctoBench](https://arxiv.org/abs/2601.10343)
