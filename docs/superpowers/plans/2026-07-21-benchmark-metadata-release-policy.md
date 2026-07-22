# Benchmark Metadata And Release Policy Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Parse benchmark identity, suite versioning, ownership, tags, contamination controls, refresh policy, and leaderboard release policy from `benchmark.yaml`.

**Architecture:** Add metadata types in `internal/benchmark/metadata.go`, parse them in `internal/manifest`, and surface them in reports and run store snapshots.

**Tech Stack:** Go, YAML parsing, table-driven validation tests.

## Global Constraints

- Existing minimal manifests remain loadable until the root `benchmark` block is made required.
- Benchmark IDs must be stable lowercase identifiers safe for filenames.
- Release policy must be available before public export.
- Contamination controls must validate private and gated split usage.

---

### Task 1: Add Benchmark Metadata Types

**Files:**
- Create: `internal/benchmark/metadata.go`
- Create: `internal/benchmark/metadata_test.go`

**Interfaces:**
- Produces: `BenchmarkMetadata`, `BenchmarkOwner`, `ReleasePolicy`, `ContaminationControl`, `LeaderboardPolicy`
- Consumes: no prior new interface

- [ ] **Step 1: Write the failing test**

```go
func TestBenchmarkMetadataRejectsInvalidID(t *testing.T) {
	meta := benchmark.BenchmarkMetadata{ID: "Local Eval"}
	if err := meta.Validate(); err == nil {
		t.Fatal("Validate() error = nil, want invalid ID error")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/benchmark`
Expected: FAIL because `BenchmarkMetadata` is undefined.

- [ ] **Step 3: Implement metadata validation**

Validate ID, suite version when release policy exists, owner contacts, lowercase unique tags, benchmark type, contamination strategy, freshness window, and holdout fraction.

- [ ] **Step 4: Run tests**

Run: `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/benchmark`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/benchmark/metadata.go internal/benchmark/metadata_test.go
git commit -m "Add benchmark metadata model"
```

### Task 2: Parse Metadata From Manifest

**Files:**
- Modify: `internal/manifest/manifest.go`
- Modify: `internal/manifest/manifest_test.go`

**Interfaces:**
- Consumes: top-level `benchmark` manifest section
- Produces: `manifest.Loaded.Metadata benchmark.BenchmarkMetadata`

- [ ] **Step 1: Write the failing test**

```go
func TestLoadParsesBenchmarkMetadata(t *testing.T) {
	loaded, err := Load(writeManifest(t, manifestWithBenchmarkMetadata("localeval")))
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Metadata.ID != "localeval" {
		t.Fatalf("metadata = %+v", loaded.Metadata)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/manifest`
Expected: FAIL because `Loaded.Metadata` is undefined.

- [ ] **Step 3: Implement parsing**

Parse ID, title, description, suite version, owners, tags, manifest version, release policy, contamination control, refresh policy, and leaderboard policy.

- [ ] **Step 4: Run tests**

Run: `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/manifest ./internal/benchmark`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/manifest internal/benchmark
git commit -m "Parse benchmark metadata"
```

### Task 3: Add Metadata To Reports

**Files:**
- Modify: `internal/benchmark/report.go`
- Modify: `internal/benchmark/report_test.go`
- Modify: `internal/benchmark/service.go`

**Interfaces:**
- Consumes: `BenchmarkMetadata`
- Produces: `BenchmarkReport.Metadata`

- [ ] **Step 1: Write the failing test**

```go
func TestReportIncludesBenchmarkMetadata(t *testing.T) {
	report := benchmark.BuildBenchmarkReportWithMetadata(benchmark.BenchmarkMetadata{ID: "localeval"}, nil, nil)
	if report.Metadata.ID != "localeval" {
		t.Fatalf("metadata = %+v", report.Metadata)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/benchmark`
Expected: FAIL because metadata report helper is undefined.

- [ ] **Step 3: Implement report metadata**

Add metadata to reports while preserving existing `BuildBenchmarkReport` behavior through a compatibility wrapper.

- [ ] **Step 4: Run tests**

Run: `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/benchmark`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/benchmark
git commit -m "Report benchmark metadata"
```

### Task 4: Enforce Release Policy During Export

**Files:**
- Modify: `internal/runstore/interchange.go`
- Modify: `internal/runstore/interchange_test.go`
- Modify: `internal/accesspolicy/policy.go`

**Interfaces:**
- Consumes: benchmark release policy and access policy
- Produces: export policy gate

- [ ] **Step 1: Write the failing test**

```go
func TestExportRejectsPrivateAnswersWhenPolicyForbids(t *testing.T) {
	err := runstore.ValidateExportPolicy(benchmark.ReleasePolicy{Leaderboard: benchmark.LeaderboardPolicy{AllowPrivateTestAnswers: false}}, runstore.ExportRequest{IncludesPrivateAnswers: true})
	if err == nil {
		t.Fatal("ValidateExportPolicy() error = nil, want private answer policy error")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/runstore`
Expected: FAIL because `ValidateExportPolicy` is undefined.

- [ ] **Step 3: Implement export policy gate**

Block exports that violate release policy for private answers, model outputs, judgments, traces, and gated artifacts.

- [ ] **Step 4: Run tests**

Run: `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/runstore ./internal/accesspolicy ./internal/benchmark`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/runstore internal/accesspolicy internal/benchmark
git commit -m "Enforce benchmark release policy"
```
