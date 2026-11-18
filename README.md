# appraise

Framework-agnostic compliance narrative generator from OSCAL catalogs and gemara artifacts.

## In the ecosystem

Formulary is a **pick-and-choose** suite of small, single-purpose CLIs. Install only the tools your workflow needs; each runs standalone. Shared contracts live in [`substrate`](https://github.com/Formulary-Labs/substrate). Optional Make pipelines (or an agent such as `regimen`) compose tools into a run — plug-and-play, not all-or-nothing.

**This tool:** auditor narrative from OSCAL catalogs and gemara artifacts (may inline risk/eval).  
**Works with:** `compound` (plan, cite-only packet), `formula` / `specimen` (packet sources).  
**Pipeline stage:** assess / present.


## Overview

`appraise` accepts any number of artifact files, auto-detects OSCAL catalogs and gemara types, and produces a compliance narrative Markdown document. Each section documents its data source. Sections requiring AI judgment carry `[DATA NEEDED]` placeholders. OSCAL models other than `catalog` are skipped.

`appraise` is composable: pass a single `ControlCatalog` for a minimal narrative, or pass a full set of artifacts for a richer document.

## Usage

```
appraise [flags] <artifact1.yaml> [artifact2.yaml] ...

Flags:
  --output string            Output Markdown path (default: compliance-narrative.md; use - for stdout)
  --program string           Program slug for document header and provenance
  --warn-unrecognized bool   Warn when input files are skipped for unrecognized types (default true; set false to suppress)
  --dry-run                  Print detected artifact types without writing output
  --version                  Print version and exit
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
