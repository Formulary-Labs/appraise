# appraise

Framework-agnostic compliance narrative generator from gemara artifacts.

## Overview

`appraise` accepts any number of gemara artifact files, auto-detects each artifact's type, and produces a compliance narrative Markdown document. Each section documents its data source. Sections requiring AI judgment carry `[DATA NEEDED]` placeholders.

`appraise` is composable: pass a single `ControlCatalog` for a minimal narrative, or pass a full set of artifacts for a richer document.

## Usage

```
appraise [flags] <artifact1.yaml> [artifact2.yaml] ...

Flags:
  --output string    Output Markdown path (default: compliance-narrative.md)
  --program string   Program slug for document header and provenance
  --dry-run          Print detected artifact types without writing output
  --version          Print version and exit
```

## Supported Artifact Types

| Type | Populates |
|---|---|
| `ControlCatalog` | Section 2 — Control Framework |
| `RiskCatalog` | Section 3 — Risk Register |
| `EvaluationLog` | Section 4 — Assessment Results |
| `AuditLog` | Section 5 — Audit Trail |
| `MappingDocument` | Section 6 — Cross-Framework Mapping |
| `GuidanceCatalog` | Section 7 — Implementation Guidance |
| `Policy` | Section 1 — Program Scope |

All other sections carry `[DATA NEEDED]` placeholders when their source artifact is not provided.

## Examples

```sh
# Minimal: single catalog
appraise data/catalogs/iso42001.yaml

# Full: multiple artifact types
appraise \
  data/catalogs/iso42001.yaml \
  data/risks/org-risks.yaml \
  data/assessments/eval.yaml \
  --program iso42001 \
  --output reports/narrative.md

# Preview detected types
appraise --dry-run iso42001.yaml eval.yaml
```

## Install

```sh
go install github.com/Formulary-Labs/appraise/cmd/appraise@latest
```

## Part of Formulary

`appraise` is part of the [Formulary](https://github.com/Formulary-Labs) compliance micro-tool ecosystem. It is composable: use it alone or combine it with `impact`, `specimen`, `exhibit`, and `formula` for a complete program artifact pipeline.
