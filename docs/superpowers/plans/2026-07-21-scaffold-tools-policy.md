# Scaffold Tools And Policy Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Turn scaffold tool metadata into validated tool registry and executable scaffold policy.

**Architecture:** Add `internal/tools` for tool definitions and policy validation. Extend scaffolds with contracts and trace capture, but keep solver/agent composition in its own plan.

**Tech Stack:** Go, JSON schema represented as `map[string]any`, table-driven tests.

## Global Constraints

- Baseline scaffolds cannot require or call tools.
- Tool IDs must resolve to top-level tool definitions.
- Secret values are never stored in manifest files.
- Existing prompt-prefix-only scaffolds remain valid.

---

### Task 1: Add Tool Registry

**Files:**
- Create: `internal/tools/registry.go`
- Create: `internal/tools/registry_test.go`

**Interfaces:**
- Produces: `tools.Definition`, `tools.Registry`, `tools.SideEffect`, `tools.ReliabilityProfile`
- Consumes: no prior new interface

- [ ] **Step 1: Write the failing test**

```go
func TestRegistryRejectsToolWithoutInputSchema(t *testing.T) {
	registry := tools.Registry{Definitions: map[string]tools.Definition{
		"plan": {Kind: "builtin", Description: "Plan"},
	}}
	if err := registry.Validate(); err == nil {
		t.Fatal("Validate() error = nil, want missing schema error")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/tools`
Expected: FAIL because `internal/tools` does not exist.

- [ ] **Step 3: Implement registry validation**

Validate tool kind, implementation runtime, input schema, output channel, side effects, process signals, and reliability profile.

- [ ] **Step 4: Run tests**

Run: `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/tools`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/tools
git commit -m "Add tool registry validation"
```

### Task 2: Add Scaffold Policy Type

**Files:**
- Modify: `internal/benchmark/model.go`
- Create: `internal/benchmark/scaffold_policy_test.go`

**Interfaces:**
- Produces: `ToolPolicy`, `InputContract`, `OutputContract`, `TraceCapture`
- Consumes: `tools.Registry`

- [ ] **Step 1: Write the failing test**

```go
func TestScaffoldPolicyRejectsRequiredToolOnDisabledPolicy(t *testing.T) {
	scaffold := benchmark.Scaffold{Name: "baseline", Baseline: true, ToolPolicy: benchmark.ToolPolicy{Mode: "disabled", RequiredTools: []string{"plan"}}}
	if err := scaffold.ValidatePolicy(tools.Registry{}); err == nil {
		t.Fatal("ValidatePolicy() error = nil, want required tool error")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/benchmark`
Expected: FAIL because policy fields are undefined.

- [ ] **Step 3: Implement policy validation**

Support `disabled`, `allow_list`, `deny_list`, and `unrestricted`. Validate required tools, denied tools, allowed tools, and max calls.

- [ ] **Step 4: Run tests**

Run: `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/benchmark ./internal/tools`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/benchmark internal/tools
git commit -m "Add scaffold tool policy model"
```

### Task 3: Parse Tools And Scaffold Contracts From Manifest

**Files:**
- Modify: `internal/manifest/manifest.go`
- Modify: `internal/manifest/manifest_test.go`
- Modify: `internal/benchmark/model.go`
- Modify: `internal/tools/registry.go`

**Interfaces:**
- Consumes: manifest `tools` and `scaffolds[*].tool_policy`
- Produces: `manifest.Loaded.Tools tools.Registry`

- [ ] **Step 1: Write the failing test**

```go
func TestLoadRejectsScaffoldUnknownTool(t *testing.T) {
	path := writeManifest(t, manifestWithAllowedTool("missing_tool"))
	if _, err := Load(path); err == nil {
		t.Fatal("Load() error = nil, want unknown tool error")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/manifest`
Expected: FAIL because top-level tools are not parsed.

- [ ] **Step 3: Implement parsing**

Parse top-level tools, scaffold system instructions, prompt suffix, tool policy, input contract, output contract, and trace capture.

- [ ] **Step 4: Run tests**

Run: `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/manifest ./internal/benchmark ./internal/tools`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/manifest internal/benchmark internal/tools
git commit -m "Parse scaffold tool policy"
```

### Task 4: Capture Tool Policy Evidence

**Files:**
- Create: `internal/benchmark/trace.go`
- Create: `internal/benchmark/trace_test.go`
- Modify: `internal/benchmark/model.go`

**Interfaces:**
- Produces: `TraceEvent`, `TraceRecorder`
- Consumes: scaffold trace capture settings

- [ ] **Step 1: Write the failing test**

```go
func TestTraceRecorderKeepsConfiguredEvents(t *testing.T) {
	recorder := benchmark.NewTraceRecorder([]string{"tool.called"})
	recorder.Record(benchmark.TraceEvent{Name: "tool.called"})
	recorder.Record(benchmark.TraceEvent{Name: "debug"})
	if got := recorder.Events(); len(got) != 1 || got[0].Name != "tool.called" {
		t.Fatalf("events = %+v", got)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/benchmark`
Expected: FAIL because `TraceRecorder` is undefined.

- [ ] **Step 3: Implement trace recorder**

Record tool events, scaffold ID, actor, timestamp, and payload. Respect scaffold trace capture filters.

- [ ] **Step 4: Run tests**

Run: `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/benchmark`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/benchmark
git commit -m "Capture scaffold tool trace events"
```
