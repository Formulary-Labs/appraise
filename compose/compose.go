// Package compose implements the core computation for appraise.
//
// appraise accepts any combination of gemara artifact files, auto-detects
// their type, and produces a framework-agnostic compliance narrative Markdown
// document. Each section documents its data source. Sections that require AI
// judgment carry [DATA NEEDED] placeholders.
//
// Design contract:
//   - Deterministic: same inputs → same output, every run.
//   - Framework-agnostic: framework identity comes from artifact metadata, not
//     hardcoded logic.
//   - Composable: works with a single artifact; richer output with more.
//   - [DATA NEEDED] marks every section that Formulary cannot fill from the
//     artifact data alone — these are explicit handoffs to the AI agent layer.
package compose

import (
	"fmt"
	"strings"
	"time"

	"github.com/Formulary-Labs/substrate/artifact"
)

// Input is one resolved artifact ready for narrative composition.
type Input struct {
	Path string
	Type artifact.ArtifactType

	// Only one of the following is non-nil, matching Type.
	ControlCatalog  *artifact.ControlCatalog
	GuidanceCatalog *artifact.GuidanceCatalog
	EvaluationLog   *artifact.EvaluationLog
	AuditLog        *artifact.AuditLog
	MappingDocument *artifact.MappingDocument
	RiskCatalog     *artifact.RiskCatalog
	Policy          *artifact.Policy
}

// Options configures a narrative composition run.
type Options struct {
	Inputs      []Input
	Program     string
	GeneratedAt time.Time
}

// Compose builds the compliance narrative Markdown from the provided inputs.
func Compose(opts Options) string {
	sb := &strings.Builder{}

	// Derive framework from the first catalog that has metadata.
	framework := inferFramework(opts.Inputs)
	program := opts.Program
	if program == "" {
		program = "[DATA NEEDED: program name]"
	}

	// --- Document header ---
	fmt.Fprintf(sb, "# Compliance Narrative\n\n")
	fmt.Fprintf(sb, "**Program:** %s  \n", program)
	if framework != "" {
		fmt.Fprintf(sb, "**Framework:** %s  \n", framework)
	}
	fmt.Fprintf(sb, "**Generated:** %s  \n", opts.GeneratedAt.Format("2006-01-02"))
	fmt.Fprintf(sb, "**Tool:** appraise (Formulary)  \n\n")
	fmt.Fprintf(sb, "> This document was generated deterministically from the gemara artifacts listed below.\n")
	fmt.Fprintf(sb, "> Sections marked `[DATA NEEDED]` require AI agent judgment to complete.\n\n")
	fmt.Fprintf(sb, "---\n\n")

	// --- Source artifacts table ---
	fmt.Fprintf(sb, "## Artifact Sources\n\n")
	if len(opts.Inputs) == 0 {
		fmt.Fprintf(sb, "[DATA NEEDED: no artifacts provided — pass gemara artifact files as arguments]\n\n")
	} else {
		fmt.Fprintf(sb, "| File | Type |\n|---|---|\n")
		for _, inp := range opts.Inputs {
			fmt.Fprintf(sb, "| `%s` | %s |\n", inp.Path, inp.Type.String())
		}
		fmt.Fprintln(sb)
	}
	fmt.Fprintf(sb, "---\n\n")

	// --- Section 1: Program Scope ---
	fmt.Fprintf(sb, "## 1. Program Scope\n\n")
	scopeWritten := false
	for _, inp := range opts.Inputs {
		if inp.Policy != nil {
			p := inp.Policy
			fmt.Fprintf(sb, "*Source: `%s` (Policy)*\n\n", inp.Path)
			fmt.Fprintf(sb, "**Policy:** %s\n\n", p.Title)
			if len(p.Scope.In.Technologies) > 0 {
				fmt.Fprintf(sb, "**In-scope technologies:** %s\n\n", strings.Join(p.Scope.In.Technologies, ", "))
			}
			if len(p.Scope.In.Geopolitical) > 0 {
				fmt.Fprintf(sb, "**Geopolitical scope:** %s\n\n", strings.Join(p.Scope.In.Geopolitical, ", "))
			}
			if len(p.Scope.Out.Technologies) > 0 {
				fmt.Fprintf(sb, "**Out-of-scope technologies:** %s\n\n", strings.Join(p.Scope.Out.Technologies, ", "))
			}
			scopeWritten = true
			break
		}
	}
	if !scopeWritten {
		fmt.Fprintf(sb, "[DATA NEEDED: scope — describe the system boundary, in-scope components, and exclusions]\n\n")
	}

	// --- Section 2: Control Framework ---
	fmt.Fprintf(sb, "## 2. Control Framework\n\n")
	controlWritten := false
	for _, inp := range opts.Inputs {
		if inp.ControlCatalog != nil {
			cat := inp.ControlCatalog
			fmt.Fprintf(sb, "*Source: `%s` (ControlCatalog)*\n\n", inp.Path)
			fmt.Fprintf(sb, "**Catalog:** %s  \n", cat.Title)
			if cat.Metadata.Description != "" {
				fmt.Fprintf(sb, "**Description:** %s  \n", cat.Metadata.Description)
			}
			fmt.Fprintf(sb, "**Controls:** %d  \n", len(cat.Controls))
			if len(cat.Groups) > 0 {
				fmt.Fprintf(sb, "**Control families:** %d  \n\n", len(cat.Groups))
				fmt.Fprintf(sb, "| Family ID | Title |\n|---|---|\n")
				for _, g := range cat.Groups {
					fmt.Fprintf(sb, "| %s | %s |\n", g.Id, g.Title)
				}
				fmt.Fprintln(sb)
			}
			controlWritten = true
			break
		}
	}
	if !controlWritten {
		fmt.Fprintf(sb, "[DATA NEEDED: control framework — provide a ControlCatalog artifact to populate this section]\n\n")
	}

	// --- Section 3: Risk Register ---
	fmt.Fprintf(sb, "## 3. Risk Register\n\n")
	riskWritten := false
	for _, inp := range opts.Inputs {
		if inp.RiskCatalog != nil {
			rc := inp.RiskCatalog
			fmt.Fprintf(sb, "*Source: `%s` (RiskCatalog)*\n\n", inp.Path)
			fmt.Fprintf(sb, "**Catalog:** %s  \n", rc.Title)
			fmt.Fprintf(sb, "**Risks:** %d  \n\n", len(rc.Risks))
			if len(rc.Risks) > 0 {
				fmt.Fprintf(sb, "| Risk ID | Title | Severity | Group |\n|---|---|---|---|\n")
				for _, r := range rc.Risks {
					fmt.Fprintf(sb, "| %s | %s | %s | %s |\n",
						r.Id, r.Title, r.Severity.String(), r.Group)
				}
				fmt.Fprintln(sb)
			}
			riskWritten = true
			break
		}
	}
	if !riskWritten {
		fmt.Fprintf(sb, "[DATA NEEDED: risk register — provide a RiskCatalog artifact or run specimen to populate]\n\n")
	}

	// --- Section 4: Assessment Results ---
	fmt.Fprintf(sb, "## 4. Assessment Results\n\n")
	evalWritten := false
	for _, inp := range opts.Inputs {
		if inp.EvaluationLog != nil {
			el := inp.EvaluationLog
			fmt.Fprintf(sb, "*Source: `%s` (EvaluationLog)*\n\n", inp.Path)
			fmt.Fprintf(sb, "**Aggregate result:** %s  \n", el.Result.String())
			fmt.Fprintf(sb, "**Evaluations:** %d  \n\n", len(el.Evaluations))
			passed, failed, needsReview, notRun := 0, 0, 0, 0
			for _, e := range el.Evaluations {
				switch e.Result.String() {
				case "Passed":
					passed++
				case "Failed":
					failed++
				case "Needs Review":
					needsReview++
				default:
					notRun++
				}
			}
			fmt.Fprintf(sb, "| Result | Count |\n|---|---|\n")
			fmt.Fprintf(sb, "| Passed | %d |\n| Failed | %d |\n| Needs Review | %d |\n| Not Run | %d |\n\n",
				passed, failed, needsReview, notRun)
			if failed > 0 || needsReview > 0 {
				fmt.Fprintf(sb, "[DATA NEEDED: remediation — describe the corrective actions planned for failed and needs-review controls]\n\n")
			}
			evalWritten = true
			break
		}
	}
	if !evalWritten {
		fmt.Fprintf(sb, "[DATA NEEDED: assessment results — run assay and provide the EvaluationLog artifact]\n\n")
	}

	// --- Section 5: Audit Trail ---
	fmt.Fprintf(sb, "## 5. Audit Trail\n\n")
	auditWritten := false
	for _, inp := range opts.Inputs {
		if inp.AuditLog != nil {
			al := inp.AuditLog
			fmt.Fprintf(sb, "*Source: `%s` (AuditLog)*\n\n", inp.Path)
			fmt.Fprintf(sb, "**Summary:** %s  \n", al.Summary)
			fmt.Fprintf(sb, "**Audit findings:** %d  \n\n", len(al.Results))
			auditWritten = true
			break
		}
	}
	if !auditWritten {
		fmt.Fprintf(sb, "[DATA NEEDED: audit trail — provide an AuditLog artifact or link to audit records]\n\n")
	}

	// --- Section 6: Control Mapping ---
	fmt.Fprintf(sb, "## 6. Cross-Framework Mapping\n\n")
	mappingWritten := false
	for _, inp := range opts.Inputs {
		if inp.MappingDocument != nil {
			md := inp.MappingDocument
			fmt.Fprintf(sb, "*Source: `%s` (MappingDocument)*\n\n", inp.Path)
			fmt.Fprintf(sb, "**Mappings:** %d relationships documented  \n\n", len(md.Mappings))
			mappingWritten = true
			break
		}
	}
	if !mappingWritten {
		fmt.Fprintf(sb, "[DATA NEEDED: cross-framework mapping — provide a MappingDocument artifact to document control equivalences]\n\n")
	}

	// --- Section 7: Guidance ---
	fmt.Fprintf(sb, "## 7. Implementation Guidance\n\n")
	guidanceWritten := false
	for _, inp := range opts.Inputs {
		if inp.GuidanceCatalog != nil {
			gc := inp.GuidanceCatalog
			fmt.Fprintf(sb, "*Source: `%s` (GuidanceCatalog)*\n\n", inp.Path)
			fmt.Fprintf(sb, "**Catalog:** %s  \n", gc.Title)
			fmt.Fprintf(sb, "**Guidelines:** %d  \n\n", len(gc.Guidelines))
			guidanceWritten = true
			break
		}
	}
	if !guidanceWritten {
		fmt.Fprintf(sb, "[DATA NEEDED: implementation guidance — describe how controls are implemented, with links to evidence]\n\n")
	}

	// --- Section 8: Management statement ---
	fmt.Fprintf(sb, "## 8. Management Statement\n\n")
	fmt.Fprintf(sb, "[DATA NEEDED: management statement — a signed statement from program leadership affirming commitment to the compliance program and accuracy of this document]\n\n")

	fmt.Fprintf(sb, "---\n\n")
	fmt.Fprintf(sb, "*Generated by [appraise](https://github.com/Formulary-Labs/appraise) — Formulary compliance micro-tool. Deterministic sections only; `[DATA NEEDED]` markers indicate AI agent handoffs.*\n")

	return sb.String()
}

// inferFramework returns a human-readable framework identifier from the first
// artifact that carries meaningful metadata description.
func inferFramework(inputs []Input) string {
	for _, inp := range inputs {
		switch {
		case inp.ControlCatalog != nil && inp.ControlCatalog.Metadata.Description != "":
			return inp.ControlCatalog.Metadata.Description
		case inp.ControlCatalog != nil && inp.ControlCatalog.Metadata.Id != "":
			return inp.ControlCatalog.Metadata.Id
		case inp.Policy != nil && inp.Policy.Metadata.Description != "":
			return inp.Policy.Metadata.Description
		case inp.RiskCatalog != nil && inp.RiskCatalog.Metadata.Description != "":
			return inp.RiskCatalog.Metadata.Description
		}
	}
	return ""
}
