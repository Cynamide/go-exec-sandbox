# Benchmark Plan Index

This index splits the scaffold-aware benchmark work into focused plans so the behavior, code, data, and public framing can be implemented and reviewed independently.

The earlier single-plan draft is superseded by this split set.

## Plan Order

1. [Domain Model and Behavior Foundation](./2026-07-14-benchmark-domain-model.md)
2. [Scaffold-Aware Runtime and Adapters](./2026-07-14-benchmark-runtime-and-adapters.md)
3. [Benchmark Surface](./2026-07-14-benchmark-surface.md)
4. [Scoring and Reporting](./2026-07-14-benchmark-scoring-and-reporting.md)
5. [Task Registry and Fixtures](./2026-07-14-benchmark-task-registry-and-fixtures.md)
6. [Product Framing and Docs](./2026-07-14-benchmark-product-framing.md)

## Coverage Notes

- Behavior comes first: tasks, scaffolds, runs, graders, and outcomes are modeled before any adapter-specific work.
- Code execution is treated as the first adapter, not the center of the architecture.
- Reporting is separated from execution so scaffold lift can be measured cleanly.
- Registry content is separated from the harness so benchmark families can expand without changing core behavior.
- README and context-map updates are separate so public positioning does not block core implementation.
