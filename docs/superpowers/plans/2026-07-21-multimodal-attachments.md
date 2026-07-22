# Multimodal Attachments Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make attachments first-class task inputs with modality validation, fixture resolution, model capability checks, preprocessing hooks, and reportable artifact references.

**Architecture:** Add `internal/attachments` for attachment definitions and modality classification. Manifest tasks reference attachments directly or through fixture fields. Model adapter validation checks capabilities before execution.

**Tech Stack:** Go, MIME type validation, path validation, table-driven tests.

## Global Constraints

- Prerequisites: fixtures-inputs for attachment resolution; provider-model-adapters for model request payloads.
- Attachment paths must stay inside approved fixture roots.
- Hidden/private attachment visibility belongs to dataset access policy.
- Model capability mismatches fail before execution.
- Image-only visual reasoning is a subset of multimodal handling.

---

### Task 1: Add Attachment Types

**Files:**
- Create: `internal/attachments/attachments.go`
- Create: `internal/attachments/attachments_test.go`

**Interfaces:**
- Produces: `Attachment`, `Modality`, `ValidateAttachment`
- Consumes: no prior new interface

- [ ] **Step 1: Write the failing test**

```go
func TestValidateAttachmentRejectsMissingMediaType(t *testing.T) {
	err := attachments.ValidateAttachment(attachments.Attachment{Kind: "image", Path: "chart.png"})
	if err == nil {
		t.Fatal("ValidateAttachment() error = nil, want media type error")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/attachments`
Expected: FAIL because `internal/attachments` does not exist.

- [ ] **Step 3: Implement attachment validation**

Support image, PDF, document, spreadsheet, presentation, audio, video, archive, screenshot, and text file attachment kinds.

- [ ] **Step 4: Run tests**

Run: `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/attachments`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/attachments
git commit -m "Add attachment validation"
```

### Task 2: Parse Attachments From Manifest Tasks

**Files:**
- Modify: `internal/manifest/manifest.go`
- Modify: `internal/manifest/manifest_test.go`
- Modify: `internal/benchmark/model.go`

**Interfaces:**
- Consumes: task `inputs.attachments`
- Produces: `Task.Attachments []attachments.Attachment`

- [ ] **Step 1: Write the failing test**

```go
func TestLoadParsesImageAttachment(t *testing.T) {
	loaded, err := Load(writeManifest(t, manifestWithImageAttachment("data/chart.png")))
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Tasks.Tasks[0].Attachments[0].Kind != "image" {
		t.Fatalf("attachments = %+v", loaded.Tasks.Tasks[0].Attachments)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/manifest`
Expected: FAIL because `Task.Attachments` is undefined.

- [ ] **Step 3: Implement parsing**

Parse `kind`, `path`, `path_from`, `media_type`, and `description`. Validate exactly one of `path` or `path_from`.

- [ ] **Step 4: Run tests**

Run: `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/manifest ./internal/attachments`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/manifest internal/benchmark internal/attachments
git commit -m "Parse task attachments"
```

### Task 3: Validate Model Modality Capabilities

**Files:**
- Create: `internal/attachments/capability.go`
- Create: `internal/attachments/capability_test.go`
- Modify: `internal/modeladapter/types.go`

**Interfaces:**
- Consumes: `modeladapter.Config.Capabilities`
- Produces: `ValidateCapabilitiesForAttachments(model Config, attachments []Attachment) error`

- [ ] **Step 1: Write the failing test**

```go
func TestValidateCapabilitiesRejectsAudioForTextModel(t *testing.T) {
	model := modeladapter.Config{ID: "text", Capabilities: map[string]bool{"multimodal": false}}
	err := attachments.ValidateCapabilitiesForAttachments(model, []attachments.Attachment{{Kind: "audio", MediaType: "audio/wav"}})
	if err == nil {
		t.Fatal("ValidateCapabilitiesForAttachments() error = nil, want capability error")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/attachments`
Expected: FAIL because capability validation is undefined.

- [ ] **Step 3: Implement capability mapping**

Map image/screenshot to `visual_reasoning` or `multimodal`, documents to `document`, audio to `audio`, video to `video`, spreadsheets to `spreadsheet`.

- [ ] **Step 4: Run tests**

Run: `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/attachments ./internal/modeladapter`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/attachments internal/modeladapter
git commit -m "Validate attachment model capabilities"
```

### Task 4: Add Native Model Payload Hooks

**Files:**
- Create: `internal/attachments/payload.go`
- Create: `internal/attachments/payload_test.go`
- Modify: `internal/modeladapter/types.go`
- Modify: `internal/modeladapter/types_test.go`

**Interfaces:**
- Consumes: `attachments.Attachment`, `modeladapter.ModelRequest`
- Produces: `attachments.BuildModelParts(root string, atts []Attachment) ([]modeladapter.ContentPart, error)`

- [ ] **Step 1: Write the failing test**

```go
func TestBuildModelPartsIncludesImageMediaType(t *testing.T) {
	path := writePNGFixture(t)
	parts, err := attachments.BuildModelParts(filepath.Dir(path), []attachments.Attachment{{Kind: "image", Path: filepath.Base(path), MediaType: "image/png"}})
	if err != nil {
		t.Fatal(err)
	}
	if parts[0].MediaType != "image/png" {
		t.Fatalf("parts = %+v", parts)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/attachments ./internal/modeladapter`
Expected: FAIL because model request content parts are undefined.

- [ ] **Step 3: Implement native payload hooks**

Add `modeladapter.ContentPart` with text, media type, file name, and bytes or URI fields. Convert attachments into content parts for adapters that declare native modality support.

- [ ] **Step 4: Run tests**

Run: `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/attachments ./internal/modeladapter`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/attachments internal/modeladapter
git commit -m "Add multimodal model payload hooks"
```

### Task 5: Add Attachment Preprocessing Hooks

**Files:**
- Create: `internal/attachments/preprocess.go`
- Create: `internal/attachments/preprocess_test.go`
- Modify: `internal/manifest/manifest.go`
- Modify: `internal/manifest/manifest_test.go`

**Interfaces:**
- Consumes: `attachments.Attachment`
- Produces: `PreprocessorRegistry`, `PreprocessAttachment(ctx context.Context, att Attachment) (Attachment, error)`

- [ ] **Step 1: Write the failing test**

```go
func TestPreprocessorRegistryRejectsUnknownHook(t *testing.T) {
	registry := attachments.PreprocessorRegistry{}
	_, err := registry.Run(context.Background(), attachments.Attachment{Kind: "pdf", Preprocessor: "missing"})
	if err == nil {
		t.Fatal("Run() error = nil, want unknown preprocessor")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/attachments`
Expected: FAIL because preprocessing hooks are undefined.

- [ ] **Step 3: Implement preprocessing registry**

Parse explicit preprocessor names from attachment config. Add validation-only hooks for documents, spreadsheets, audio, and video, and return unsupported errors until concrete preprocessors are registered.

- [ ] **Step 4: Run tests**

Run: `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/attachments ./internal/manifest`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/attachments internal/manifest
git commit -m "Add attachment preprocessing hooks"
```

### Task 6: Add Attachment Report References

**Files:**
- Modify: `internal/benchmark/model.go`
- Modify: `internal/benchmark/report.go`
- Modify: `internal/benchmark/report_test.go`

**Interfaces:**
- Consumes: task/sample attachments
- Produces: report attachment references

- [ ] **Step 1: Write the failing test**

```go
func TestReportIncludesAttachmentReferences(t *testing.T) {
	report := benchmark.BuildBenchmarkReport([]benchmark.Task{{ID: "task", Attachments: []attachments.Attachment{{Kind: "image", Path: "data/chart.png", MediaType: "image/png"}}}}, nil)
	if len(report.Attachments) != 1 {
		t.Fatalf("attachments = %+v", report.Attachments)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/benchmark`
Expected: FAIL because report attachment references are absent.

- [ ] **Step 3: Implement report references**

Add attachment metadata without embedding file contents. Include task ID, sample ID when available, kind, media type, and sanitized path.

- [ ] **Step 4: Run tests**

Run: `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/benchmark`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/benchmark
git commit -m "Report benchmark attachment references"
```
