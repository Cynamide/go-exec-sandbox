# Scoring Interfaces Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add scorer interfaces for generative, choice/target-score, log-likelihood, perplexity, executable verifier, judge, pairwise preference, and environment reward scoring.

**Architecture:** Create `internal/scoring` with scorer-specific request and result types. Keep deterministic artifact checks in grading, while model-side and environment-side scoring go through explicit scorer interfaces.

**Tech Stack:** Go, model adapter interfaces, table-driven tests.

## Global Constraints

- Do not require logprob support from adapters that only generate text.
- Judge scoring requires a configured judge model role.
- Reward scoring requires an interactive environment.
- Scorer outputs must be aggregatable by the grading engine.

---

### Task 1: Add Core Scorer Interfaces

**Files:**
- Create: `internal/scoring/scoring.go`
- Create: `internal/scoring/scoring_test.go`

**Interfaces:**
- Produces: `Scorer`, `Request`, `Result`, `Score`
- Consumes: solver artifacts and model adapter responses

- [ ] **Step 1: Write the failing test**

```go
func TestScoreRequiresName(t *testing.T) {
	score := scoring.Score{Value: 1}
	if err := score.Validate(); err == nil {
		t.Fatal("Validate() error = nil, want missing name error")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/scoring`
Expected: FAIL because `internal/scoring` does not exist.

- [ ] **Step 3: Implement scorer types**

Define scorer type names, score values, metadata, higher-is-better, and error handling.

- [ ] **Step 4: Run tests**

Run: `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/scoring`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/scoring
git commit -m "Add scoring interfaces"
```

### Task 2: Add Model-Side Scoring Interfaces

**Files:**
- Modify: `internal/modeladapter/types.go`
- Modify: `internal/modeladapter/types_test.go`
- Create: `internal/scoring/model_scorers.go`
- Create: `internal/scoring/model_scorers_test.go`

**Interfaces:**
- Produces: `LogLikelihoodAdapter`, `ChoiceScorer`, `PerplexityScorer`
- Consumes: model adapter config

- [ ] **Step 1: Write the failing test**

```go
func TestLogLikelihoodScorerRejectsAdapterWithoutSupport(t *testing.T) {
	scorer := scoring.LogLikelihoodScorer{}
	_, err := scorer.Score(context.Background(), scoring.Request{Model: fakeGenerateOnlyAdapter{}})
	if err == nil {
		t.Fatal("Score() error = nil, want unsupported loglikelihood error")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/scoring ./internal/modeladapter`
Expected: FAIL because model-side scorer interfaces are undefined.

- [ ] **Step 3: Implement model-side interfaces**

Add adapter optional interfaces for `ScoreChoices`, `LogLikelihood`, and `Perplexity`. Implement scorer wrappers that reject unsupported adapters.

- [ ] **Step 4: Run tests**

Run: `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/scoring ./internal/modeladapter`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/scoring internal/modeladapter
git commit -m "Add model-side scoring interfaces"
```

### Task 3: Add Judge And Preference Scorers

**Files:**
- Create: `internal/scoring/judge.go`
- Create: `internal/scoring/judge_test.go`

**Interfaces:**
- Produces: `JudgeScorer`, `PairwisePreferenceScorer`
- Consumes: judge model adapter and rubric fixtures

- [ ] **Step 1: Write the failing test**

```go
func TestJudgeScorerRequiresJudgeModel(t *testing.T) {
	scorer := scoring.JudgeScorer{}
	_, err := scorer.Score(context.Background(), scoring.Request{})
	if err == nil {
		t.Fatal("Score() error = nil, want missing judge model error")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/scoring`
Expected: FAIL because `JudgeScorer` is undefined.

- [ ] **Step 3: Implement judge scorers**

Build judge prompts from rubric fixtures, parse structured judge responses, and record judge model ID in score metadata.

- [ ] **Step 4: Run tests**

Run: `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/scoring`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/scoring
git commit -m "Add judge scoring interfaces"
```

### Task 4: Add Executable And Reward Scorers

**Files:**
- Create: `internal/scoring/verifier.go`
- Create: `internal/scoring/verifier_test.go`

**Interfaces:**
- Produces: `ExecutableVerifierScorer`, `RewardScorer`
- Consumes: sandbox executor and environment step results

- [ ] **Step 1: Write the failing test**

```go
func TestRewardScorerRequiresEnvironmentResult(t *testing.T) {
	scorer := scoring.RewardScorer{}
	_, err := scorer.Score(context.Background(), scoring.Request{})
	if err == nil {
		t.Fatal("Score() error = nil, want missing environment result error")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/scoring`
Expected: FAIL because `RewardScorer` is undefined.

- [ ] **Step 3: Implement executable and reward scorers**

Score test-command outcomes, verifier exit codes, reward totals, termination status, and trajectory metadata.

- [ ] **Step 4: Run tests**

Run: `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/scoring`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/scoring
git commit -m "Add verifier and reward scorers"
```
