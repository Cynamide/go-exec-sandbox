# Dataset Access And Visibility Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add policy for public, gated, private, and evaluator-only benchmark data so reports and exports do not leak hidden answers or restricted artifacts.

**Architecture:** Add `internal/accesspolicy` for split visibility and answer-visibility rules. Fixture loading and report export consult the policy before reading, scoring, or publishing data.

**Tech Stack:** Go, manifest validation, report sanitization tests.

## Global Constraints

- Prerequisites: fixtures-inputs for fixture loading; run-persistence-interchange for public exports.
- Hidden answers must never appear in public exports.
- Evaluator-only fixtures cannot be scored locally.
- Gated fixtures require environment references, not inline secrets.
- Canary metadata must be explicit for canary-bearing task files.

---

### Task 1: Add Access Policy Types

**Files:**
- Create: `internal/accesspolicy/policy.go`
- Create: `internal/accesspolicy/policy_test.go`

**Interfaces:**
- Produces: `Policy`, `SplitPolicy`, `AnswerVisibility`, `PublishPolicy`
- Consumes: fixture split metadata

- [ ] **Step 1: Write the failing test**

```go
func TestPolicyRejectsUnknownAnswerVisibility(t *testing.T) {
	policy := accesspolicy.Policy{Splits: map[string]accesspolicy.SplitPolicy{"private": {AnswerVisibility: "everyone"}}}
	if err := policy.Validate(); err == nil {
		t.Fatal("Validate() error = nil, want unknown visibility error")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/accesspolicy`
Expected: FAIL because `internal/accesspolicy` does not exist.

- [ ] **Step 3: Implement policy validation**

Support public, local-only, evaluator-only, and hidden answer visibility. Validate publishability for outputs, judgments, traces, and artifacts.

- [ ] **Step 4: Run tests**

Run: `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/accesspolicy`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/accesspolicy
git commit -m "Add dataset access policy"
```

### Task 2: Parse Access Policy From Manifest

**Files:**
- Modify: `internal/manifest/manifest.go`
- Modify: `internal/manifest/manifest_test.go`

**Interfaces:**
- Consumes: manifest access policy section
- Produces: `manifest.Loaded.AccessPolicy accesspolicy.Policy`

- [ ] **Step 1: Write the failing test**

```go
func TestLoadRejectsHiddenSplitWithoutPolicy(t *testing.T) {
	path := writeManifest(t, manifestWithPrivateFixtureWithoutAccessPolicy())
	if _, err := Load(path); err == nil {
		t.Fatal("Load() error = nil, want access policy error")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/manifest`
Expected: FAIL because access policy is not parsed.

- [ ] **Step 3: Implement parsing**

Parse split visibility, answer visibility, gated access environment variable names, private evaluator handoff, hidden metadata, canary metadata, and annotator visibility.

- [ ] **Step 4: Run tests**

Run: `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/manifest ./internal/accesspolicy`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/manifest internal/accesspolicy
git commit -m "Parse dataset access policy"
```

### Task 3: Enforce Policy During Fixture Loading

**Files:**
- Modify: `internal/fixtures/registry.go`
- Modify: `internal/fixtures/registry_test.go`

**Interfaces:**
- Consumes: `accesspolicy.Policy`
- Produces: fixture load authorization checks

- [ ] **Step 1: Write the failing test**

```go
func TestFixtureLoaderRejectsMissingGatedToken(t *testing.T) {
	policy := accesspolicy.Policy{Splits: map[string]accesspolicy.SplitPolicy{"gated_test": {RequiresEnv: []string{"GAIA_TOKEN"}}}}
	err := fixtures.AuthorizeLoad(fixtures.Definition{Split: "gated_test"}, policy)
	if err == nil {
		t.Fatal("AuthorizeLoad() error = nil, want missing env error")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/fixtures`
Expected: FAIL because `AuthorizeLoad` is undefined.

- [ ] **Step 3: Implement authorization**

Check gated env vars, evaluator-only restrictions, hidden answer availability, and canary metadata before fixture cases are returned.

- [ ] **Step 4: Run tests**

Run: `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/fixtures ./internal/accesspolicy`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/fixtures internal/accesspolicy
git commit -m "Enforce fixture access policy"
```

### Task 4: Sanitize Public Exports

**Files:**
- Modify: `internal/benchmark/report_sanitize.go`
- Modify: `internal/benchmark/report_sanitize_test.go`
- Modify: `internal/runstore/interchange.go`

**Interfaces:**
- Consumes: `accesspolicy.Policy`
- Produces: public-safe reports and exports

- [ ] **Step 1: Write the failing test**

```go
func TestPublicExportRemovesHiddenExpectedOutput(t *testing.T) {
	report := benchmark.BenchmarkReport{Runs: []benchmark.Run{{ExpectedOutput: "secret"}}}
	safe := benchmark.SanitizeForVisibility(report, accesspolicy.Policy{DefaultAnswerVisibility: "hidden"})
	if safe.Runs[0].ExpectedOutput != "" {
		t.Fatalf("hidden answer leaked: %+v", safe.Runs[0])
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/benchmark`
Expected: FAIL because `SanitizeForVisibility` is undefined.

- [ ] **Step 3: Implement sanitization**

Remove hidden answers, evaluator-only metadata, private traces, private attachment paths, and hidden judgments from public exports.

- [ ] **Step 4: Run tests**

Run: `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/benchmark ./internal/runstore ./internal/accesspolicy`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/benchmark internal/runstore internal/accesspolicy
git commit -m "Sanitize dataset visibility in exports"
```
