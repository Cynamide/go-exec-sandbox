# Benchmark Manifest Design

Date: 2026-07-17

## Goal

Define one canonical benchmark configuration file, `benchmark.yaml`, that is rich enough to describe benchmark execution end to end:

- benchmark identity and defaults
- provider and model matrix
- tool contracts
- scaffold contracts
- task registry
- allowed input sources
- allowed output channels
- outcome grading
- process grading
- trace capture
- fixture references

The file must be self-explanatory enough that if a field appears in a scaffold or task, the same file explains what that field means and where its values come from.

## Single-File Rule

There is only one config file:

- `benchmark.yaml`

Other files are permitted only when they are data payloads rather than config:

- task case fixtures
- rubric fixtures
- repo seed files
- browser snapshots
- expected file outputs

Those files are benchmark assets, not a second config system.

## Design Principles

- One source of truth for benchmark configuration.
- Fixtures live outside YAML when payload size would make the config unreadable.
- Outcome grading and process grading are first-class and separate.
- Tools are defined declaratively in the manifest, not merely referenced by ID.
- Scaffolds declare both what they inject and what they permit.
- Tasks describe required inputs, expected outputs, and grading policy explicitly.
- Every enum-like field must have documented sample values in the same document.
- The schema should represent more cases than the repo currently implements, but each forward-looking field must be justified and tied to an end-to-end benchmark need.

## Agent Use Contract

This document is written so a coding agent can configure or validate `benchmark.yaml` without having to infer hidden rules from code.

A coding agent should be able to answer these questions from this file alone:

- What top-level sections are required?
- Which fields are identifiers versus human-readable metadata?
- Which fields are enums and what values are allowed?
- Which fields reference another section by key?
- Which fields inherit defaults?
- Which fields may be `null`, empty, or omitted?
- Which fields are config versus external fixture data?
- Which combinations are invalid even if the YAML parses?

If any future manifest field cannot be explained in those terms inside this document, the schema is incomplete.

## Reading Order For Implementers

For a coding agent implementing the loader or editing the config, the recommended reading order is:

1. `Canonical Example`
2. `Field Reference`
3. `Validation Rules`
4. `Required External Fixture Shapes`
5. `Design Review Log`

That order gives the happy-path shape first, then the precise semantics, then the edge cases.

## Canonical Example

```yaml
schema_version: 1

benchmark:
  id: localeval
  title: LocalEval Workflow Benchmark
  description: Benchmark harness for outcome quality and process quality across workflow tasks.
  suite_version: "2026.07"
  owners:
    - team: eval-platform
      contact: evals@example.com
  tags:
    - workflow
    - scaffold-aware
    - artifact-aware
  manifest_version: "2026-07-17"
  release_policy:
    benchmark_type: official
    contamination_control:
      enabled: true
      strategy: canary_and_private_splits
    leaderboard:
      publish_model_outputs: optional
      publish_judgments: true
      allow_private_test_answers: true

runtime_defaults:
  task_mode: code_exec
  language: python
  timeout_ms: 60000
  max_attempts: 1
  concurrency: 4
  epochs: 1
  random_seed: 17
  working_directory: repo_root
  sandbox_profile: default
  approval_policy: never
  fail_policy:
    fail_on_error: sample
    continue_on_fail: true
    score_on_error: false
  sample_limits:
    token_limit: 32000
    message_limit: 200
    time_limit_s: 600
    cost_limit_usd: 5
  capture:
    stdout: true
    stderr: true
    exit_code: true
    generated_file: true
    generated_directory: true
    git_diff: true
    trace_events: true
    model_outputs: true
  normalization:
    trim_trailing_whitespace: true
    normalize_line_endings: true

providers:
  ollama_local:
    kind: ollama
    base_url_env: OLLAMA_HOST
    model_lookup: direct
  openai_api:
    kind: openai
    api_key_env: OPENAI_API_KEY
    base_url: https://api.openai.com/v1
  local_openai_compatible:
    kind: openai_compatible
    base_url: http://localhost:8000/v1
    api_key_env: LOCAL_MODEL_API_KEY
    model_lookup: direct
  deployed_custom_endpoint:
    kind: custom_http
    base_url: https://model.example.com
    transport:
      protocol: https
      request_format: json
      response_format: json
      inference_path: /infer
      healthcheck_path: /health

models:
  qwen3_4b_local:
    provider: ollama_local
    model_name: qwen3:4b
    enabled: true
    params:
      temperature: 0
    capabilities:
      tool_use: false
      file_editing: false
      browser: false
      multimodal: false
      judge: false
  gpt_5_4_mini_medium:
    provider: openai_api
    model_name: gpt-5.4-mini
    enabled: true
    params:
      reasoning_effort: medium
      temperature: 0
    capabilities:
      tool_use: true
      file_editing: true
      browser: true
      multimodal: true
      judge: true
  local_qwen_endpoint:
    provider: local_openai_compatible
    model_name: qwen2.5-coder-32b-instruct
    endpoint_url: http://localhost:8000/v1/chat/completions
    enabled: false
    auth:
      type: bearer_env
      env: LOCAL_MODEL_API_KEY
    params:
      temperature: 0
      max_tokens: 4096
    request_mapping:
      messages_field: messages
      temperature_field: temperature
      max_tokens_field: max_tokens
      model_field: model
    response_mapping:
      text_path: choices[0].message.content
      finish_reason_path: choices[0].finish_reason
    capabilities:
      tool_use: false
      file_editing: false
      browser: false
      multimodal: false
      judge: false

default_model_roles:
  judge: gpt_5_4_mini_medium
  planner: gpt_5_4_mini_medium

tools:
  plan:
    kind: builtin
    description: Produce a short execution plan before implementation.
    implementation:
      runtime: harness_builtin
      builtin_id: plan
    inputs:
      schema:
        type: object
        required: [task_id, task_description]
        properties:
          task_id:
            type: string
          task_description:
            type: string
          repo_context:
            type: array
            items:
              type: string
    outputs:
      primary:
        channel: trace
        format: json
      emits_events:
        - tool.called
        - tool.succeeded
        - plan.created
    side_effects: none
    process_signals:
      - planning_present
      - planning_before_edit

  task_brief:
    kind: builtin
    description: Read a narrow task brief or fixture summary before acting.
    implementation:
      runtime: harness_builtin
      builtin_id: task_brief
    inputs:
      schema:
        type: object
        required: [task_id]
        properties:
          task_id:
            type: string
          fixture_subset:
            type: string
    outputs:
      primary:
        channel: trace
        format: json
      emits_events:
        - tool.called
        - tool.succeeded
        - context.loaded
    side_effects: read_only
    process_signals:
      - context_gathering

  repo_search:
    kind: builtin
    description: Search repository files for task-relevant text.
    implementation:
      runtime: harness_builtin
      builtin_id: repo_search
    inputs:
      schema:
        type: object
        required: [query]
        properties:
          query:
            type: string
          include_globs:
            type: array
            items:
              type: string
    outputs:
      primary:
        channel: trace
        format: json
      emits_events:
        - tool.called
        - tool.succeeded
        - repo.search_results
    side_effects: read_only
    process_signals:
      - repo_inspection

  browser_open:
    kind: builtin
    description: Open a page needed for browser or web workflow tasks.
    implementation:
      runtime: harness_builtin
      builtin_id: browser_open
    inputs:
      schema:
        type: object
        required: [url]
        properties:
          url:
            type: string
          wait_for:
            type: string
    outputs:
      primary:
        channel: browser_state
        format: json
      emits_events:
        - tool.called
        - tool.succeeded
        - browser.page_loaded
    side_effects: external_navigation
    process_signals:
      - browser_navigation

  shell_exec:
    kind: builtin
    description: Run an allowed shell command in the benchmark workspace.
    implementation:
      runtime: harness_builtin
      builtin_id: shell_exec
    inputs:
      schema:
        type: object
        required: [command]
        properties:
          command:
            type: string
          timeout_ms:
            type: integer
    outputs:
      primary:
        channel: exec_result
        format: json
      emits_events:
        - tool.called
        - tool.succeeded
        - command.executed
    side_effects: workspace_mutation
    process_signals:
      - command_execution

scaffolds:
  baseline:
    baseline: true
    description: Run the task with no extra instructions or tools.
    system_instructions: []
    prompt_prefix: ""
    prompt_suffix: ""
    tool_policy:
      mode: disabled
      allowed_tools: []
      required_tools: []
    input_contract:
      task_text: required
      fixture_input: optional
      repo_context: optional
      attachments: optional
      browser_start_state: optional
      environment_variables: optional
    output_contract:
      primary_channel: stdout
      additional_channels:
        - stderr
        - exit_code
        - trace
    trace_capture:
      events:
        - model.started
        - model.completed
        - execution.finished

  tool_assisted:
    baseline: false
    description: Require a compact plan and allow bounded tool usage before the final answer.
    system_instructions:
      - Think briefly before coding.
      - Use tools only when they change the probability of success.
      - Return only the final executable output when the task asks for code.
    prompt_prefix: |
      First think through the solution briefly, then write the final code.
    prompt_suffix: ""
    tool_policy:
      mode: allow_list
      allowed_tools:
        - plan
        - task_brief
        - repo_search
        - shell_exec
      required_tools:
        - plan
      max_calls_per_tool:
        plan: 1
        task_brief: 2
        repo_search: 8
        shell_exec: 4
    input_contract:
      task_text: required
      fixture_input: optional
      repo_context: required
      attachments: optional
      browser_start_state: optional
      environment_variables: optional
    output_contract:
      primary_channel: stdout
      additional_channels:
        - stderr
        - exit_code
        - trace
        - git_diff
        - generated_file
        - generated_directory
    trace_capture:
      events:
        - model.started
        - tool.called
        - tool.succeeded
        - plan.created
        - command.executed
        - file.modified
        - execution.finished

  browser_assisted:
    baseline: false
    description: Permit browser navigation for tasks that need live web or UI interaction.
    system_instructions:
      - Verify the visible state before answering.
      - Capture the minimum browser trace needed for grading.
    prompt_prefix: |
      Use the browser only when the task requires visible state or web interaction.
    prompt_suffix: ""
    tool_policy:
      mode: allow_list
      allowed_tools:
        - browser_open
        - plan
      required_tools: []
    input_contract:
      task_text: required
      fixture_input: optional
      repo_context: optional
      attachments: optional
      browser_start_state: required
      environment_variables: optional
    output_contract:
      primary_channel: browser_state
      additional_channels:
        - trace
        - screenshot
        - stdout
    trace_capture:
      events:
        - browser.page_loaded
        - browser.action_taken
        - browser.snapshot_captured

fixtures:
  release_note_cases:
    kind: task_cases
    path: data/tasks/release-note-digest.json
    format: json
    split: public_dev
  support_escalation_cases:
    kind: task_cases
    path: data/tasks/support-escalation-summary.json
    format: json
    split: public_dev
  support_escalation_process_rubric:
    kind: rubric
    path: data/rubrics/support-escalation-summary.yaml
    format: yaml
  issue_repo_seed:
    kind: repo_seed
    path: data/repos/sample-issue-task.tar.gz
    format: tar_gz
    environment:
      repo: sample/repo
      base_commit: abcdef123456
  browser_checkout_fixture:
    kind: browser_fixture
    path: data/browser/checkout-timeout.json
    format: json

tasks:
  release_note_digest:
    id: release-note-digest
    version: v1
    split: public_dev
    task_mode: code_exec
    title: Generate A Release Notes Digest
    family: engineering_workflows
    language: python
    description: Read deployment notes and produce a markdown table.
    tags:
      - report_generation
      - markdown
    models:
      include:
        - qwen3_4b_local
        - gpt_5_4_mini_medium
    scaffolds:
      include:
        - baseline
        - tool_assisted
    fewshot:
      num_examples: 0
      sampler: none
    inputs:
      sources:
        - kind: stdin_fixture
          fixture: release_note_cases
          field: cases[*].input
      repo_seed: null
      attachments: []
      browser_start_state: null
      environment:
        required: []
        optional: []
    outputs:
      primary:
        channel: stdout
        format: markdown
        media_type: text/markdown
      side_effects: []
    grading:
      outcome:
        enabled: true
        preferred_score: exact_artifact
        metrics:
          - name: exact_artifact
            aggregation: mean
            higher_is_better: true
        checks:
          - type: artifact_markdown
            fixture: release_note_cases
            actual_from: stdout
            expected_from: cases[*].expected_output
      process:
        enabled: false

  support_escalation_summary:
    id: support-escalation-summary
    version: v1
    split: public_dev
    task_mode: code_exec
    title: Summarize Support Escalations Into Markdown
    family: support_workflows
    language: python
    description: Read escalation records and produce a markdown report.
    tags:
      - report_generation
      - markdown
      - process_scored
    models:
      include:
        - gpt_5_4_mini_medium
    scaffolds:
      include:
        - baseline
        - tool_assisted
    fewshot:
      num_examples: 0
      sampler: none
    inputs:
      sources:
        - kind: stdin_fixture
          fixture: support_escalation_cases
          field: cases[*].input
      repo_seed: null
      attachments: []
      browser_start_state: null
      environment:
        required: []
        optional: []
    outputs:
      primary:
        channel: stdout
        format: markdown
        media_type: text/markdown
      side_effects: []
    grading:
      outcome:
        enabled: true
        preferred_score: exact_artifact
        metrics:
          - name: exact_artifact
            aggregation: mean
            higher_is_better: true
        checks:
          - type: artifact_markdown
            fixture: support_escalation_cases
            actual_from: stdout
            expected_from: cases[*].expected_output
      process:
        enabled: true
        judge_model_role: judge
        checks:
          - type: required_events
            must_include:
              - model.started
              - execution.finished
          - type: allowed_tool_usage
            allowed_tools:
              - plan
              - task_brief
              - repo_search
          - type: rubric
            fixture: support_escalation_process_rubric

  repo_issue_fix:
    id: repo-issue-fix
    version: v1
    split: public_dev
    task_mode: repo_patch
    title: Fix A Repository Bug And Preserve Tests
    family: engineering_workflows
    language: python
    description: Modify repository files to satisfy failing tests and produce the required patch.
    tags:
      - repo_mutation
      - patch
      - process_scored
    models:
      include:
        - gpt_5_4_mini_medium
    scaffolds:
      include:
        - tool_assisted
    fewshot:
      num_examples: 0
      sampler: none
    inputs:
      sources:
        - kind: task_text
      repo_seed:
        fixture: issue_repo_seed
      attachments: []
      browser_start_state: null
      environment:
        required:
          - TEST_COMMAND
        optional:
          - FEATURE_FLAG
    outputs:
      primary:
        channel: git_diff
        format: unified_diff
        media_type: text/x-diff
      side_effects:
        - generated_file
        - exit_code
    grading:
      outcome:
        enabled: true
        preferred_score: resolved_rate
        metrics:
          - name: resolved_rate
            aggregation: mean
            higher_is_better: true
        checks:
          - type: git_diff_applies
          - type: tests_pass
            command_from_env: TEST_COMMAND
          - type: file_contains
            path: README.md
            must_contain: "bug fix"
      process:
        enabled: true
        checks:
          - type: required_events
            must_include:
              - repo.search_results
              - file.modified
              - command.executed
          - type: event_order
            sequence:
              - repo.search_results
              - file.modified
              - command.executed

  browser_checkout_investigation:
    id: browser-checkout-investigation
    version: v1
    split: gated_test
    task_mode: browser_agent
    title: Investigate A Browser Checkout Timeout
    family: support_workflows
    language: none
    description: Use browser context to identify the cause of a checkout failure and produce a structured incident summary.
    tags:
      - browser
      - investigation
      - process_scored
    models:
      include:
        - gpt_5_4_mini_medium
    scaffolds:
      include:
        - browser_assisted
    fewshot:
      num_examples: 0
      sampler: none
    inputs:
      sources:
        - kind: browser_fixture
          fixture: browser_checkout_fixture
      repo_seed: null
      attachments:
        - kind: screenshot
          path: data/browser/checkout-timeout.png
      browser_start_state:
        url: https://checkout.example.local/cart
      environment:
        required: []
        optional: []
    outputs:
      primary:
        channel: browser_state
        format: json
        media_type: application/json
        state_projection: visible_text_and_dom
      side_effects:
        - screenshot
        - stdout
    grading:
      outcome:
        enabled: true
        preferred_score: exact_artifact
        metrics:
          - name: exact_artifact
            aggregation: mean
            higher_is_better: true
        checks:
          - type: artifact_json
            expected_inline:
              root_keys:
                - probable_cause
                - evidence
                - next_action
      process:
        enabled: true
        judge_model_role: judge
        checks:
          - type: required_events
            must_include:
              - browser.page_loaded
              - browser.snapshot_captured
          - type: minimum_event_count
            event: browser.action_taken
            at_least: 1
```

## Field Reference

### `schema_version`

Top-level manifest version. Increment when backward-incompatible parsing changes.

Sample:

```yaml
schema_version: 1
```

### `benchmark`

Human and organizational metadata for the benchmark suite.

Fields:

- `id`: stable machine ID
- `title`: human name
- `description`: high-level purpose
- `owners`: ownership metadata
- `tags`: suite-level taxonomy
- `manifest_version`: version label for the benchmark definition
- `suite_version`: benchmark release identifier
- `release_policy`: contamination, visibility, and leaderboard publication rules

### `runtime_defaults`

Defaults applied unless a task or scaffold overrides them.

Fields:

- `task_mode`: default execution mode
- `language`: default execution language
- `timeout_ms`: default execution timeout
- `max_attempts`: retry count for repeated attempts
- `concurrency`: max parallel runs
- `working_directory`: where the run executes
- `sandbox_profile`: named sandbox preset
- `approval_policy`: tool approval mode
- `fail_policy`: behavior on per-sample failures
- `sample_limits`: token, time, message, and cost limits
- `capture`: what evidence to record
- `normalization`: output cleanup rules before grading

These limits are directly motivated by agent-eval systems that need bounded autonomy rather than single-shot text scoring.

Inheritance rules:

- `runtime_defaults` apply to every task unless that task overrides the field explicitly.
- A scaffold may narrow behavior further, but should not silently widen safety limits beyond `runtime_defaults`.
- If both a task and a scaffold influence the same runtime concern, the more restrictive effective value wins.

Recommended `task_mode` values:

- `code_exec`
- `artifact_only`
- `repo_patch`
- `browser_agent`
- `api_agent`
- `text_only`

`task_mode` is not optional in the effective task configuration. A loader should compute it from the task value or the runtime default before selecting executors, tools, or graders.

### `providers`

API or runtime backends that models are loaded from.

Supported `kind` values in this design:

- `ollama`
- `openai`
- `openai_compatible`
- `anthropic`
- `local_binary`
- `custom_http`

This repo currently implements only Ollama-backed generation in code.

Reference semantics:

- provider keys such as `ollama_local` are symbolic IDs
- `models[*].provider` must match one of those keys exactly
- `base_url` may point to a local server or a deployed remote endpoint
- `base_url_env` is preferred when the endpoint location changes by environment
- `transport` is required for `custom_http` and optional for provider kinds with a standardized protocol

Provider fields a loader should support:

- `kind`
- `base_url`
- `base_url_env`
- `api_key_env`
- `headers`
- `transport.protocol`
- `transport.inference_path`
- `transport.healthcheck_path`
- `transport.request_format`
- `transport.response_format`

### `models`

Named runnable model entries.

Fields:

- `provider`: key into `providers`
- `model_name`: provider-facing model identifier
- `endpoint_url`: optional direct inference URL for this model entry
- `auth`: model-specific auth override
- `enabled`: inclusion toggle
- `params`: provider-specific settings
- `request_mapping`: field mapping for custom or compatible endpoints
- `response_mapping`: extraction mapping for custom or compatible endpoints
- `capabilities`: declared abilities used for validation and task selection

Useful `capabilities` flags:

- `tool_use`
- `file_editing`
- `browser`
- `multimodal`
- `structured_output`
- `judge`

Model input forms the manifest should support:

- provider plus `model_name`
  - best for standard APIs such as Ollama, OpenAI, Anthropic
- provider plus `endpoint_url`
  - best when one deployed model has its own route
- direct local endpoint
  - same as `endpoint_url`, but points to `localhost` or another private address
- OpenAI-compatible wrapper endpoint
  - use `kind: openai_compatible`
  - useful for vLLM, LiteLLM, local gateways, and deployed compatible routers
- fully custom HTTP endpoint
  - use `kind: custom_http`
  - requires request and response mapping
- local binary runner
  - use `kind: local_binary`
  - useful for benchmarks that shell out to a CLI model wrapper

Recommended semantics:

- if `endpoint_url` is present on the model, it overrides provider-level inference path construction
- if `auth` is present on the model, it overrides provider-level auth behavior
- if `request_mapping` or `response_mapping` is present, the loader must use those mappings instead of provider defaults

Example model auth shapes:

```yaml
auth:
  type: bearer_env
  env: OPENAI_API_KEY
```

```yaml
auth:
  type: header_env
  header: X-API-Key
  env: CUSTOM_MODEL_KEY
```

```yaml
auth:
  type: none
```

Example mapping shapes:

```yaml
request_mapping:
  messages_field: messages
  temperature_field: temperature
  max_tokens_field: max_tokens
  model_field: model
```

```yaml
response_mapping:
  text_path: choices[0].message.content
  finish_reason_path: choices[0].finish_reason
  usage_prompt_tokens_path: usage.prompt_tokens
  usage_completion_tokens_path: usage.completion_tokens
```

### `default_model_roles`

Named model assignments for non-primary roles.

Useful roles:

- `judge`
- `planner`
- `retriever`

This is important when outcome or process scoring depends on a judge model rather than the evaluated model alone.

Reference semantics:

- role values must match keys under `models`
- a missing role means the harness should fall back to the evaluated model only when the task does not require that role
- `primary` should not be global because benchmark runs often evaluate many models against the same task set

### `tools`

Declarative tool contracts. This section closes the gap where a scaffold might mention a tool name without defining what it is.

Fields:

- `kind`: implementation class such as `builtin`, `http`, `shell`, `browser`, `repo`
- `description`: human explanation
- `implementation`: how the harness invokes the tool at runtime
- `inputs.schema`: declarative input contract
- `outputs.primary.channel`: primary evidence channel produced by the tool
- `outputs.primary.format`: payload format
- `outputs.emits_events`: trace events emitted by successful use
- `side_effects`: `none`, `read_only`, `workspace_mutation`, `external_navigation`, `network_call`
- `process_signals`: reusable process-grade labels

#### Tool implementation contract

This closes the gap behind the question "how is the tool written?"

Every tool must declare an `implementation` block. Supported shapes in this design:

- `runtime: harness_builtin`
  - tool logic is implemented directly in the benchmark harness
  - requires `builtin_id`
- `runtime: shell_command`
  - tool shells out to a command
  - requires `command`
  - may define `args_template`
- `runtime: http_endpoint`
  - tool calls a service
  - requires `method` and `url`
  - may define auth via env-var references
- `runtime: browser_driver`
  - tool routes through a browser controller
  - requires `action`

Secrets must never be written inline in the manifest. Secret-bearing tools use env var references or named secret handles only.

Reference semantics:

- tool keys such as `plan` or `repo_search` are stable IDs
- `scaffolds[*].tool_policy.allowed_tools`
- `scaffolds[*].tool_policy.required_tools`
- `tasks[*].grading.process.checks[*].allowed_tools`
  must all reference these keys exactly

### `scaffolds`

Prompting and tool-access strategies applied to tasks.

Fields:

- `baseline`: whether this scaffold is the required baseline
- `description`: scaffold purpose
- `system_instructions`: instruction list
- `prompt_prefix`: prepended text
- `prompt_suffix`: appended text
- `tool_policy`: tool permissions and limits
- `input_contract`: what categories of input the scaffold can consume
- `output_contract`: what channels it is expected to emit
- `trace_capture.events`: which trace events should be recorded

Contract semantics:

- `output_contract` describes what the scaffold can emit and expects the harness to capture
- the task still chooses the scored primary output channel
- the effective task output must be compatible with the scaffold contract, not necessarily identical to the scaffold's preferred primary channel

#### Scaffold input types

The `input_contract` section should support these categories:

- `task_text`
- `fixture_input`
- `repo_context`
- `attachments`
- `browser_start_state`
- `environment_variables`

These are categories, not concrete payloads. Concrete payloads are provided by the task under `inputs`.

Recommended values for each category:

- `required`
- `optional`
- `forbidden`

Sample:

```yaml
input_contract:
  task_text: required
  fixture_input: optional
  repo_context: required
  attachments: forbidden
  browser_start_state: optional
  environment_variables: optional
```

Reference semantics:

- scaffold keys such as `baseline` or `tool_assisted` are stable IDs
- `tasks[*].scaffolds.include` and any future scaffold references must match those keys exactly

### `fixtures`

Registry of external data assets.

Allowed `kind` values in this design:

- `task_cases`
- `rubric`
- `repo_seed`
- `browser_fixture`
- `attachment`
- `expected_files`
- `golden_patch`

Useful fixture metadata fields:

- `split`: such as `train`, `public_dev`, `private_test`, `gated_test`
- `environment`: reproduction metadata such as repo and base commit

Allowed `format` values in this design:

- `json`
- `jsonl`
- `yaml`
- `csv`
- `markdown`
- `text`
- `tar_gz`
- `zip`
- `png`
- `html`

For generated files, `format` is not enough by itself. File-producing outputs should also declare a `media_type` such as:

- `text/markdown`
- `text/csv`
- `application/json`
- `application/pdf`
- `application/vnd.openxmlformats-officedocument.wordprocessingml.document`
- `application/vnd.openxmlformats-officedocument.spreadsheetml.sheet`
- `application/vnd.openxmlformats-officedocument.presentationml.presentation`
- `image/png`

Reference semantics:

- fixture keys such as `release_note_cases` are stable IDs
- any `fixture:` field anywhere in the manifest must reference these keys exactly

### `tasks`

Canonical task registry. Each task chooses:

- which models can run it
- which scaffolds can run it
- what inputs it needs
- what output channel matters
- how outcome grading works
- whether process grading is enabled

Recommended task identity fields:

- `id`
- `version`
- `split`
- `task_mode`
- `language`
- `family`
- `tags`

These make leaderboard submissions, cached results, and benchmark-release evolution much easier to reason about.

Reference semantics:

- task map keys like `release_note_digest` are local manifest entry names
- `tasks[*].id` is the canonical externally reported task ID
- the loader should treat `tasks[*].id` as the reporting identity and the map key as an authoring convenience

### `tasks[*].fewshot`

Few-shot policy for tasks that need exemplars.

Fields:

- `num_examples`
- `sampler`
- `split`
- `examples_fixture`

Recommended `sampler` values:

- `none`
- `random`
- `first_n`
- `fixed_list`

### `tasks[*].task_mode`

Execution mode for the task instance.

Recommended values:

- `code_exec`
  - model produces code that is executed in the sandbox
- `artifact_only`
  - model produces an artifact directly without code execution
- `repo_patch`
  - model mutates a repository or emits a patch
- `browser_agent`
  - model interacts with browser state or browser tools
- `api_agent`
  - model interacts with APIs or HTTP tools as the main task path
- `text_only`
  - model returns natural-language or structured text without executor-side action

This field should drive executor selection and validation. For example:

- `browser_agent` should not require a runnable source language
- `repo_patch` should enable diff and file-mutation capture
- `code_exec` should require a concrete executable language

### `tasks[*].language`

Execution-language selector for the task.

Recommended values:

- executable language IDs such as `python`, `go`, `javascript`
- `none` for tasks whose primary path is not code execution

Compatibility guidance:

- `code_exec` should use a real executable language
- `repo_patch` may use an executable language or `none`, depending on whether code generation or patch emission is primary
- `browser_agent`, `api_agent`, and `text_only` may use `none`

Omission semantics:

- if `fewshot` is omitted, the effective default is `num_examples: 0`
- if `num_examples: 0`, `sampler`, `split`, and `examples_fixture` are ignored

### `tasks[*].inputs.sources[*].kind`

Allowed input source kinds in this design:

- `task_text`
- `stdin_fixture`
- `json_fixture`
- `csv_fixture`
- `browser_fixture`
- `api_fixture`
- `db_fixture`

`field` uses a restricted fixture path syntax:

- `cases[*].input`
- `cases[*].expected_output`
- `records[*].payload`

The loader should support only dot access plus `[*]`. Avoid full JSONPath complexity.

Omission semantics:

- `field` is required when a fixture contains more than one logical payload slot
- `field` may be omitted only when the fixture kind has exactly one canonical payload shape

### `tasks[*].outputs.primary.channel`

Allowed primary output channels in this design:

- `stdout`
- `stderr`
- `exit_code`
- `generated_file`
- `generated_directory`
- `git_diff`
- `browser_state`
- `screenshot`
- `http_response`
- `exec_result`
- `trace`

Channel-specific companion fields:

- `generated_file`
  - `path_pattern`
  - `media_type`
- `generated_directory`
  - `root_path`
- `http_response`
  - `status_field`
  - `body_field`
- `browser_state`
  - `state_projection`

Sample channel snippets:

```yaml
outputs:
  primary:
    channel: stdout
    format: markdown
    media_type: text/markdown
```

```yaml
outputs:
  primary:
    channel: generated_file
    format: binary
    path_pattern: reports/summary.pdf
    media_type: application/pdf
```

```yaml
outputs:
  primary:
    channel: generated_directory
    format: binary
    root_path: output/site
```

```yaml
outputs:
  primary:
    channel: git_diff
    format: unified_diff
    media_type: text/x-diff
```

```yaml
outputs:
  primary:
    channel: browser_state
    format: json
    state_projection: visible_text_and_dom
```

```yaml
outputs:
  primary:
    channel: http_response
    format: json
    status_field: status
    body_field: body
```

```yaml
outputs:
  primary:
    channel: exit_code
    format: text
```

Compatibility rule:

- the task's primary output channel must be one of the scaffold's allowed output channels
- the scaffold's `primary_channel` is a preferred default, not a mandatory match

### `tasks[*].outputs.primary.format`

Allowed primary output formats in this design:

- `text`
- `markdown`
- `csv`
- `json`
- `html`
- `xml`
- `yaml`
- `unified_diff`
- `binary`

For office and document outputs, prefer:

- `channel: generated_file`
- `format: binary`
- `media_type: application/pdf` or the relevant Office media type

Null and empty semantics:

- `outputs.side_effects: []` means no scored side effects are expected
- omission of `side_effects` should be treated the same as an empty list

### `tasks[*].grading.outcome.checks[*].type`

Outcome grading types covered by this design:

- `exact_stdout`
- `normalized_stdout`
- `regex_stdout`
- `contains_all`
- `artifact_markdown`
- `artifact_csv`
- `artifact_json`
- `artifact_yaml`
- `artifact_html`
- `exit_code_equals`
- `file_exact`
- `file_contains`
- `directory_matches`
- `git_diff_applies`
- `tests_pass`
- `http_response_match`
- `browser_state_match`

Current repo code directly implements only:

- plain trimmed stdout equality
- markdown semantic comparison
- csv semantic comparison
- json semantic comparison

Recommended interpretation:

- `format` describes the comparison family
- `media_type` describes the concrete artifact encoding

### `tasks[*].grading.outcome.metrics`

Aggregate metrics recorded for the task outcome score.

Fields:

- `name`
- `aggregation`
- `higher_is_better`
- `weight_by_size`

This comes from external harnesses that separate raw sample checks from reported aggregate metrics.

Recommended `aggregation` values:

- `mean`
- `sum`
- `pass_rate`
- `weighted_mean`

If a metric depends on a judge model rather than deterministic checks alone, the outcome section should declare `judge_model_role`.

### `tasks[*].grading.process.checks[*].type`

Process grading types covered by this design:

- `required_events`
- `forbidden_events`
- `event_order`
- `minimum_event_count`
- `maximum_event_count`
- `allowed_tool_usage`
- `required_tool_usage`
- `rubric`
- `state_transition_match`
- `recovery_after_failure`

The repo does not currently implement process scoring; this section is schema design for the next feature wave.

If a process rubric or model-judged score is used, the task should declare `judge_model_role`.

### `tasks[*].inputs.environment`

This section declares environment dependencies without embedding secret values.

Fields:

- `required`: environment variable names that must exist
- `optional`: environment variable names that may exist

Never store secrets inline. The manifest may name `OPENAI_API_KEY`, but not its value.

### `benchmark.release_policy`

Fields:

- `benchmark_type`: such as `official`, `experimental`, `internal`
- `contamination_control.enabled`
- `contamination_control.strategy`
- `leaderboard.publish_model_outputs`
- `leaderboard.publish_judgments`
- `leaderboard.allow_private_test_answers`

Recommended `publish_model_outputs` values:

- `never`
- `optional`
- `required`

### `runtime_defaults.approval_policy`

Recommended values:

- `never`
- `on_demand`
- `always`

### `runtime_defaults.fail_policy`

Fields:

- `fail_on_error`: scoring scope affected by an error
- `continue_on_fail`: whether the run continues after a failed sample
- `score_on_error`: whether errored samples still receive a score

Recommended `fail_on_error` values:

- `sample`
- `task`
- `run`

### `scaffolds[*].tool_policy.mode`

Recommended values:

- `disabled`
- `allow_list`
- `deny_list`
- `unrestricted`

## Null, Empty, And Omitted Value Semantics

The loader should treat these differently:

- omitted
  - use documented default behavior
- `null`
  - intentionally absent and not applicable
- empty list `[]`
  - explicitly present but with zero members
- empty string `""`
  - explicitly present but blank; reject for identifier fields

Examples:

- `repo_seed: null`
  means the task does not require a repository seed
- `attachments: []`
  means the task supports attachments as a category but this specific task has none
- omitted `fewshot`
  means use the default no-fewshot behavior

## Include/Exclude And Selection Semantics

For any selection block, the manifest should follow these rules:

- `include` means only these IDs are eligible
- `exclude` means these IDs are removed from the eligible set after inclusion is computed
- if `include` is omitted, the default is all defined IDs of that category
- if both are present, `exclude` wins for conflicting IDs

Recommended future-compatible shape:

```yaml
models:
  include: [gpt_5_4_mini_medium]
  exclude: [qwen3_4b_local]
```

The current examples only use `include`, but the loader should be designed so `exclude` can be added without changing the surrounding contract.

## Implementation Checklist For Coding Agents

When implementing the loader or validating a new benchmark entry, use this checklist:

1. Parse top-level sections and reject unknown top-level keys.
2. Resolve all cross-references: provider, model, tool, scaffold, and fixture IDs.
3. Apply `runtime_defaults` inheritance before validation of effective task settings.
4. Compute effective model and scaffold selections for each task.
5. Validate capability compatibility for each task-model-scaffold combination.
6. Validate that every output channel has compatible outcome checks.
7. Validate that every process check can be satisfied by declared trace capture plus tool events.
8. Validate null and omission semantics instead of treating everything missing as identical.
9. Resolve fixture paths and restricted `field` selectors.
10. Only after configuration validation succeeds, hand the task to runtime execution.

## How To Add A New Entry Safely

### Add a new model

- add a provider if needed
- add the model under `models`
- declare capability flags honestly
- optionally wire it into `model_roles`

### Add a new tool

- add it under `tools`
- declare its `implementation`
- define emitted trace events
- define side effects
- update scaffold allow lists as needed

### Add a new scaffold

- add the scaffold under `scaffolds`
- set `baseline: true` only if it is the single baseline
- declare `tool_policy`
- declare `input_contract`
- declare `trace_capture.events`

### Add a new task

- add the task entry
- define `id`, `version`, `split`, `family`, `language`
- point it at fixtures
- define output channel and format
- add outcome checks
- add process checks only if the runtime can emit the evidence they require

## Output Coverage Matrix

The schema is intended to cover these benchmark result shapes:

- Plain CLI answer via `stdout`
- Structured report via `stdout` with `markdown`, `csv`, `json`, `yaml`, `html`
- Failure or warning channel via `stderr`
- Non-text correctness via `exit_code`
- Generated artifact file such as `.md`, `.csv`, `.json`, `.docx`, `.pptx`, `.xlsx`
- Generated document artifact such as `.pdf`
- Generated image artifact such as `.png`
- Generated directory tree
- Repository mutation via `git_diff`
- Browser-observed state
- Browser screenshots
- HTTP response payloads
- Execution trace and tool events

## Process Evidence Model

If process grading is enabled, the harness must capture an append-only event stream. Each event should have at least:

```yaml
event:
  name: tool.called
  timestamp: "2026-07-17T18:05:01Z"
  actor: model
  payload:
    tool: repo_search
    args:
      query: "checkout timeout"
```

Required event fields:

- `name`
- `timestamp`
- `actor`
- `payload`

Recommended actors:

- `model`
- `tool`
- `executor`
- `browser`
- `grader`

## Required External Fixture Shapes

These are not config files. They are benchmark data payloads.

### Task case fixture

```json
{
  "cases": [
    {
      "case_id": "release-note-001",
      "input": "api|feature|Added audit logging\nweb|fix|Fixed checkout redirect\n",
      "expected_output": "| team | feature changes | fix changes |\n| --- | --- | --- |\n| api | 1 | 0 |\n| web | 0 | 1 |",
      "metadata": {
        "difficulty": "easy"
      }
    }
  ]
}
```

### Rubric fixture

```yaml
criteria:
  - id: planning_present
    description: A plan exists before edits begin.
    required: true
  - id: repo_inspection
    description: The model inspects relevant repository context before modifying files.
    required: true
```

### Browser fixture

```json
{
  "start_url": "https://checkout.example.local/cart",
  "cookies": [],
  "local_storage": {},
  "expected_visible_text": ["Checkout failed", "Try again"]
}
```

### Expected files fixture

```yaml
files:
  - path: report/summary.md
    media_type: text/markdown
    match:
      type: exact
      path: golden/report/summary.md
  - path: report/dashboard.xlsx
    media_type: application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
    match:
      type: exists
```

## Validation Rules

The manifest should be rejected if any of these fail:

- exactly one scaffold has `baseline: true`
- every referenced tool exists in `tools`
- every referenced fixture exists in `fixtures`
- every referenced model exists in `models`
- every referenced scaffold exists in `scaffolds`
- every task defines at least one outcome check
- every task has an effective `task_mode`
- every `task_mode: code_exec` task uses a concrete executable `language`
- every `task_mode: browser_agent` task uses a browser-capable scaffold
- every process check type is valid when `process.enabled: true`
- every declared `judge_model_role` resolves through `default_model_roles` to an existing model with `capabilities.judge: true`
- every `fewshot.num_examples > 0` task defines enough information to source the examples
- every model either resolves through a provider contract or defines a complete direct endpoint override
- every `kind: custom_http` provider defines both request and response transport semantics
- every model-level `response_mapping` includes a `text_path` when the model is used for text generation
- every `required_tools` entry is also present in `allowed_tools` when `tool_policy.mode: allow_list`
- every `outputs.primary.channel` is compatible with the grading checks
- every task output channel is allowed by the selected scaffold contract
- every file-producing output defines `media_type`
- every tool event named in process grading can actually be emitted by the selected scaffold tools or runtime capture
- any task using `git_diff` output also enables `capture.git_diff`
- any task using browser outputs includes `browser_start_state` or a browser fixture
- any selected task-model pair satisfies declared capability requirements
- no secret value is stored inline in the manifest

## Current Repo Mapping

Fields already represented in current code or data:

- tasks with IDs, titles, descriptions, families, languages
- scaffold names, descriptions, prompt prefixes, baseline flag, tool names
- task test cases
- artifact expectations with markdown, csv, json, text semantics
- baseline versus scaffolded execution

Fields designed here but not implemented yet:

- provider registry beyond current runtime env config
- multi-model run matrix in one benchmark run
- declarative tool contracts
- process trace capture and process grading
- file, directory, git diff, browser, and HTTP outcome grading
- fixture registry
- validation of tool/event compatibility

## External Benchmark Audit

This design was checked against primary-source benchmark systems and expanded where those systems exposed schema needs that were missing here.

### EleutherAI lm-evaluation-harness

Relevant patterns:

- YAML task configuration as a shareable evaluation contract
- explicit output type selection
- few-shot configuration
- metric lists and aggregation metadata
- repeats, filters, and decontamination controls

Manifest impacts:

- keep one manifest as the reproducible contract
- add `fewshot`
- add explicit `metrics`
- add release and contamination policy metadata

Primary source:

- `docs/task_guide.md` in `EleutherAI/lm-evaluation-harness`

### OpenAI Evals

Relevant patterns:

- versioned eval IDs
- YAML registry entries pointing at external sample files
- support for tool-using and model-graded evals

Manifest impacts:

- add task `version`
- keep fixtures as data assets rather than inlining all samples
- support judge-model roles for graded checks

Primary source:

- `docs/build-eval.md` and the eval registry format in `openai/evals`

### Inspect

Relevant patterns:

- task-level sandbox and approval settings
- separate solver, scorer, and grader roles
- sample execution limits
- failure handling options
- run-config style reproducibility

Manifest impacts:

- add `approval_policy`
- add `sample_limits`
- add `fail_policy`
- add `model_roles`

Primary source:

- Inspect task documentation and standard tools documentation

### SWE-bench

Relevant patterns:

- explicit split handling
- repository reproduction metadata such as repo and base commit
- fail-to-pass tests as the main outcome signal

Manifest impacts:

- add split metadata on tasks and fixtures
- add repo-environment metadata on repo-seed fixtures
- keep `tests_pass` and patch-oriented outcome checks first-class

Primary source:

- SWE-bench dataset documentation and benchmark overview

### BIG-bench

Relevant patterns:

- task versioning
- preferred score and metrics
- multiple task data representations

Manifest impacts:

- add `preferred_score`
- add aggregate `metrics`
- keep support for both simple fixture tasks and richer programmatic tasks

Primary source:

- BIG-bench task creation documentation

### LiveBench

Relevant patterns:

- contamination-aware releases
- release cadence and benchmark versioning
- model configuration separation from question data

Manifest impacts:

- add `suite_version`
- add benchmark `release_policy`
- keep fixtures separate from model registry entries

Primary source:

- LiveBench repository documentation and datasheet

### GAIA

Relevant patterns:

- gated versus public split visibility
- tool-augmented assistant benchmark shape
- optional reasoning traces in leaderboard submissions
- attachment-aware tasks with file references

Manifest impacts:

- add `split` values such as `public_dev` and `gated_test`
- keep attachments and browser fixtures first-class
- keep process evidence and reasoning traces first-class

Primary source:

- GAIA dataset card and leaderboard submission format

## Design Review Log

### Pass 1

Issue:

- A scaffold could mention tools without defining them.

Resolution:

- Added top-level `tools` registry with declarative contracts, emitted events, and side effects.

### Pass 2

Issue:

- Process grading could reference trace events that are never emitted or captured.

Resolution:

- Added `trace_capture.events` to scaffolds and a validation rule requiring grading events to be compatible with capture plus tool contracts.

### Pass 3

Issue:

- Output coverage was too narrow if the benchmark expands to repo mutation, browser tasks, or generated files.

Resolution:

- Added output channels and grading types for `git_diff`, generated files, browser state, screenshots, and HTTP responses.

### Pass 4

Issue:

- The single-file rule was blurry because fixtures might look like secondary config.

Resolution:

- Declared fixtures to be benchmark data payloads only. The manifest remains the only config file.

### Pass 5

Issue:

- Tool names and scaffold inputs were documented, but not all categories were enumerated in one place.

Resolution:

- Added explicit lists for scaffold input categories, task input source kinds, output channels, output formats, and both outcome and process grading types.

### Pass 6

Issue:

- A model could be selected for a task even if the task implicitly needs capabilities the model does not have.

Resolution:

- Added `capabilities` to models. A future validator should reject impossible task-model pairings such as browser-required tasks on models with `browser: false`.

### Pass 7

Issue:

- Document and file outputs were named, but the schema did not distinguish comparison format from concrete file type.

Resolution:

- Added `media_type` guidance plus file-output companion fields so `.pdf`, `.docx`, `.xlsx`, `.pptx`, and image artifacts are representable without inventing one format enum per extension.

### Pass 8

Issue:

- Tool definitions described what a tool means, but not how the harness locates or invokes the implementation.

Resolution:

- Added the `implementation` block with supported runtime shapes and explicit `builtin_id` examples.

### Pass 9

Issue:

- Fixture field extraction syntax and secret handling were implicit, which would cause inconsistent loader implementations.

Resolution:

- Restricted fixture field paths to a small path grammar and added explicit rules forbidding inline secret values.

### Pass 10

Issue:

- The manifest needed stricter rules to prevent impossible scaffold and model combinations, not just valid syntax.

Resolution:

- Added validation rules for `required_tools` subset correctness and task-model capability compatibility.

### Pass 11

Issue:

- Internal review alone could still miss fields that matter for reproducibility, leaderboard operation, contamination control, or long-horizon agent evaluation.

Resolution:

- Audited the schema against lm-evaluation-harness, OpenAI Evals, Inspect, SWE-bench, BIG-bench, LiveBench, and GAIA.
- Added suite release metadata, split visibility, model roles, failure policy, sample limits, few-shot policy, judge-model wiring, and repo-environment metadata.

### Pass 12

Issue:

- After the external audit additions, the risk was not missing a field but introducing contradictions between the canonical YAML example, the field reference, and the validation rules.

Resolution:

- Re-checked naming and coverage for `split`, `fewshot`, `model_roles`, `approval_policy`, `fail_policy`, `sample_limits`, `judge_model_role`, and file-output metadata.
- No remaining schema blocker was found in this pass.

### Pass 13

Issue:

- The schema still left ambiguity around sample identity and non-executable tasks, which would make result caching and loader validation fuzzy.

Resolution:

- Added `case_id` to task fixture examples.
- Added explicit `language` semantics and validation rules for `task_mode` versus language compatibility.

### Pass 14

Issue:

- Judge-role wiring and few-shot sourcing were present in the schema but not yet enforced in validation rules.

Resolution:

- Added validation rules for `judge_model_role`, few-shot sourcing, scaffold output compatibility, and browser-scaffold compatibility.

### Pass 15

Issue:

- The model section supported provider names but still did not fully explain direct endpoint use, OpenAI-compatible wrappers, per-model auth overrides, or custom request/response mappings.

Resolution:

- Added provider and model support for direct `endpoint_url`, `openai_compatible`, `custom_http`, model-level auth overrides, and request/response mappings.
- Added validation rules so endpoint-based model entries are not underspecified.

## Remaining Open Questions

These are not schema blockers, but they affect implementation order:

- Whether `artifact_yaml` and `artifact_html` should be implemented before or after file-based grading.
- Whether rubric grading is deterministic rules only, model-judged, or both.
- Whether `generated_directory` grading needs checksum-based matching, manifest-based matching, or semantic file-by-file checks.

My recommendation:

- Implement in this order:
  1. single `benchmark.yaml` loader
  2. fixture registry
  3. model matrix
  4. tool registry validation
  5. process trace capture
  6. process grading
  7. browser and repo-mutation grading

## Recommendation

Use this design as the target manifest contract. It keeps one config file while leaving bulky task payloads in fixtures, and it is explicit enough to support both outcome correctness and agent behavior scoring without a hidden side registry.
