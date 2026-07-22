# Multimodal Attachments Spec

## Problem

The current broad design names `visual_reasoning` and generic attachments, but real assistant benchmarks include images, PDFs, office documents, audio, video, archives, browser screenshots, and mixed document collections. A scaffold-aware benchmark should validate modality support and pass attachments through model adapters and tools consistently.

## Config Fields Covered

- attachment registry
- attachment kind
- media type
- path and path-from fixture references
- document/audio/video modality capability flags
- attachment visibility by split
- attachment preprocessing
- model adapter multimodal request construction

## Current Code State

- `benchmark.yaml` has no attachments.
- `internal/manifest/manifest.go` does not parse attachment config.
- Model capabilities only exist in the broad config design, not code.
- The current Ollama generation path is text-only.

## Required Behavior

- Represent attachments as first-class task inputs.
- Validate attachment kind and media type.
- Validate selected model capabilities against every attachment modality.
- Add adapter hooks for model-native multimodal payloads.
- Add preprocessing hooks for documents, spreadsheets, audio, and video when model-native support is unavailable.
- Capture attachment references and derived artifacts in reports.

## Validation Rules

- Reject attachment paths outside approved fixture roots.
- Reject unsupported media types.
- Reject model-task pairings where required modality capability is absent.
- Reject private attachments in public reports unless release policy permits publication.
- Reject visual reasoning tasks that have no visual or multimodal input.

## Acceptance Criteria

- Image-only visual reasoning is one supported case, not the whole multimodal model.
- Document/audio/video benchmarks can be represented without inventing separate task modes for every file type.
- Attachment handling is auditable from manifest through report.
