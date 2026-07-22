# Interactive Environment Lifecycle Spec

## Problem

Browser, terminal, API, game-like, and simulated-user benchmarks require environment lifecycle semantics: setup, reset, auth/session state, observations, actions, rewards, termination, truncation, cleanup, and replay. The current code has only Docker code execution and a static browser-start-state concept in the broad config design.

## Config Fields Covered

- environment registry
- environment kind
- setup commands or fixtures
- reset policy
- auth/session fixtures
- observation projection
- action space
- reward source
- termination and truncation conditions
- cleanup policy
- render mode
- deterministic seed

## Current Code State

- `internal/sandbox` runs a single code container per execution request.
- There is no persistent episode model.
- There is no action/observation loop.
- There is no browser/API/session reset contract.
- There is no reward or termination model.

## Required Behavior

- Add an environment interface for reset, step, observe, score/reward, and cleanup.
- Support browser, terminal, API, repo, notebook, spreadsheet, and simulated conversation environments behind the same lifecycle seam.
- Capture observations and actions as trace events.
- Enforce cleanup even when samples fail or time out.
- Support deterministic reset through fixture IDs and random seed.
- Make auth and session fixtures explicit and non-secret-bearing.

## Validation Rules

- Reject interactive task modes without an environment reference.
- Reject environments without cleanup policy.
- Reject action-space values unsupported by the environment kind.
- Reject auth/session fixtures that expose secret values inline.
- Reject reward-based grading without reward source or termination semantics.

## Acceptance Criteria

- Browser/WebArena-like tasks can describe environment lifecycle without embedding logic in task text.
- MiniWoB-style step/reward/termination tasks can be represented.
- Environment traces are sufficient for replay and process grading.
