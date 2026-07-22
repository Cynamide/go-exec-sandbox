# Prompt Rendering And Response Parsing Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add explicit prompt rendering and response parsing so benchmark tasks can render samples, scaffolds, few-shot examples, choices, and agent actions without hardcoded LLM prompts.

**Architecture:** Create `internal/prompting` for renderers and parsers. Benchmark runners call a renderer to produce model requests and a parser to normalize responses into artifacts or actions.

**Tech Stack:** Go, text templates, JSON parsing, table-driven tests.

## Global Constraints

- The current code-generation prompt remains available as the default renderer.
- Rendered prompts and parsed outputs must be reportable.
- Template references to missing sample fields must fail before execution.
- Parsers must be deterministic and named.

---

### Task 1: Add Renderer Interface And Default Code Renderer

**Files:**
- Create: `internal/prompting/render.go`
- Create: `internal/prompting/render_test.go`
- Modify: `internal/llm/llm.go`

**Interfaces:**
- Produces: `Renderer`, `RenderInput`, `RenderedPrompt`, `CodeRenderer`
- Consumes: `benchmark.Task`, `benchmark.Scaffold`, `benchmark.Sample`

- [ ] **Step 1: Write the failing test**

```go
func TestCodeRendererIncludesScaffoldPrefix(t *testing.T) {
	rendered, err := prompting.CodeRenderer{}.Render(prompting.RenderInput{
		Task: benchmark.Task{Language: "python", Description: "sum numbers"},
		Scaffold: benchmark.Scaffold{PromptPrefix: "Plan first.\n"},
		Sample: benchmark.Sample{Input: "1 2"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(rendered.User, "Plan first.") || !strings.Contains(rendered.User, "sum numbers") {
		t.Fatalf("rendered prompt = %+v", rendered)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/prompting`
Expected: FAIL because `internal/prompting` does not exist.

- [ ] **Step 3: Implement renderer**

Move prompt construction out of `llm.chatRequest` into `CodeRenderer`. Keep `llm.Client.GenerateCode` using `CodeRenderer` until runners are migrated.

- [ ] **Step 4: Run tests**

Run: `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/prompting ./internal/llm`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/prompting internal/llm/llm.go
git commit -m "Add prompt renderer interface"
```

### Task 2: Add Template-Based Task Renderer

**Files:**
- Modify: `internal/prompting/render.go`
- Create: `internal/prompting/template_renderer_test.go`

**Interfaces:**
- Consumes: `RenderInput`
- Produces: `TemplateRenderer{SystemTemplate, UserTemplate string}`

- [ ] **Step 1: Write the failing test**

```go
func TestTemplateRendererRejectsMissingField(t *testing.T) {
	renderer := prompting.TemplateRenderer{UserTemplate: "{{ .Sample.Metadata.ticket }}"}
	_, err := renderer.Render(prompting.RenderInput{Sample: benchmark.Sample{ID: "case-1"}})
	if err == nil {
		t.Fatal("Render() error = nil, want missing field error")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/prompting`
Expected: FAIL because `TemplateRenderer` is undefined.

- [ ] **Step 3: Implement template renderer**

Use `text/template` with `missingkey=error`. Expose task, sample, scaffold, few-shot examples, and model capability flags to templates.

- [ ] **Step 4: Run tests**

Run: `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/prompting`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/prompting
git commit -m "Add template prompt renderer"
```

### Task 3: Add Response Parser Registry

**Files:**
- Create: `internal/prompting/parser.go`
- Create: `internal/prompting/parser_test.go`

**Interfaces:**
- Produces: `Parser`, `ParseResult`, `Registry`
- Consumes: raw model text

- [ ] **Step 1: Write the failing test**

```go
func TestStripMarkdownFenceParser(t *testing.T) {
	result, err := prompting.StripMarkdownFenceParser{}.Parse("```json\n{\"ok\":true}\n```")
	if err != nil {
		t.Fatal(err)
	}
	if result.Text != "{\"ok\":true}" {
		t.Fatalf("Text = %q", result.Text)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/prompting`
Expected: FAIL because parser types are undefined.

- [ ] **Step 3: Implement parsers**

Add parsers for raw text, markdown fence stripping, JSON object extraction, choice extraction, and action extraction.

- [ ] **Step 4: Run tests**

Run: `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/prompting`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/prompting
git commit -m "Add response parser registry"
```

### Task 4: Parse Renderer And Parser Config From Manifest

**Files:**
- Modify: `internal/manifest/manifest.go`
- Modify: `internal/manifest/manifest_test.go`
- Modify: `internal/benchmark/model.go`

**Interfaces:**
- Consumes: manifest task renderer/parser fields
- Produces: `Task.RendererID`, `Task.ParserID`

- [ ] **Step 1: Write the failing test**

```go
func TestLoadParsesTaskRendererAndParser(t *testing.T) {
	loaded, err := Load(writeManifest(t, manifestWithRendererParser("template", "json_object")))
	if err != nil {
		t.Fatal(err)
	}
	task := loaded.Tasks.Tasks[0]
	if task.RendererID != "template" || task.ParserID != "json_object" {
		t.Fatalf("task = %+v", task)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/manifest`
Expected: FAIL because task renderer/parser fields are undefined.

- [ ] **Step 3: Implement parsing and validation**

Parse renderer and parser IDs. Validate IDs against built-in registries unless custom registry config is added in the same manifest.

- [ ] **Step 4: Run tests**

Run: `GOCACHE=$PWD/.cache/go-build /usr/local/go/bin/go test ./internal/manifest ./internal/prompting`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/manifest internal/benchmark internal/prompting
git commit -m "Parse prompt renderer config"
```
