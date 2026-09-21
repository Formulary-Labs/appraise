// appraise produces a framework-agnostic compliance narrative Markdown document
// from any combination of gemara artifact files.
//
// Usage:
//
//	appraise [flags] <artifact1> [artifact2] ...
package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/Formulary-Labs/appraise/compose"
	"github.com/Formulary-Labs/substrate/artifact"
	"github.com/Formulary-Labs/substrate/exit"
	"github.com/Formulary-Labs/substrate/provenance"
)

const version = "0.1.0"

func main() {
	var (
		outputFlag  = flag.String("output", "compliance-narrative.md", "Output Markdown path (use - for stdout)")
		programFlag = flag.String("program", "", "Program slug (for provenance and document header)")
		dryRunFlag  = flag.Bool("dry-run", false, "Print detected artifact types without writing output")
		versionFlag = flag.Bool("version", false, "Print version and exit")
	)
	flag.Usage = usage
	flag.Parse()

	if *versionFlag {
		fmt.Printf("appraise version %s\n", version)
		os.Exit(exit.OK)
	}

	paths := flag.Args()
	if len(paths) == 0 {
		fmt.Fprintln(os.Stderr, `{"error": "at least one artifact file is required", "code": 2}`)
		flag.Usage()
		os.Exit(exit.ToolError)
	}

	// Resolve each artifact file.
	inputs := make([]compose.Input, 0, len(paths))
	for _, path := range paths {
		inp, err := resolveArtifact(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: skipping %s: %v\n", path, err)
			continue
		}
		inputs = append(inputs, inp)
	}

	if *dryRunFlag {
		fmt.Printf("%-40s %s\n", "File", "Detected Type")
		for _, inp := range inputs {
			fmt.Printf("%-40s %s\n", inp.Path, inp.Type.String())
		}
		os.Exit(exit.OK)
	}

	narrative := compose.Compose(compose.Options{
		Inputs:      inputs,
		Program:     *programFlag,
		GeneratedAt: time.Now().UTC(),
	})

	if *outputFlag == "-" {
		fmt.Print(narrative)
	} else {
		if err := os.WriteFile(*outputFlag, []byte(narrative), 0o644); err != nil {
			fmt.Fprintf(os.Stderr, `{"error": %q, "code": 2}`+"\n", err.Error())
			os.Exit(exit.ToolError)
		}
		fmt.Fprintf(os.Stderr, "wrote %s (%d artifacts)\n", *outputFlag, len(inputs))
	}

	_ = provenance.Write("logs/provenance.jsonl", provenance.Entry{
		Spec:        "functions/appraise-spec.md",
		Output:      *outputFlag,
		OutputType:  "other",
		Program:     *programFlag,
		Purpose:     fmt.Sprintf("appraise: compliance narrative from %d artifact(s)", len(inputs)),
		Reusability: provenance.Instance,
		QualityGate: provenance.Pass,
		Tool:        "appraise",
		ToolVersion: version,
	})
}

// resolveArtifact detects the type of a gemara artifact file and loads it into
// a compose.Input. Unknown types are returned as an error so the caller can
// warn and continue.
func resolveArtifact(path string) (compose.Input, error) {
	t, err := artifact.DetectType(path)
	if err != nil {
		return compose.Input{}, fmt.Errorf("detecting type: %w", err)
	}

	inp := compose.Input{Path: path, Type: t}

	switch t {
	case artifact.ControlCatalogArtifact:
		cat, err := artifact.LoadControlCatalog(path)
		if err != nil {
			return compose.Input{}, fmt.Errorf("loading ControlCatalog: %w", err)
		}
		inp.ControlCatalog = cat

	case artifact.GuidanceCatalogArtifact:
		gc, err := artifact.LoadGuidanceCatalog(path)
		if err != nil {
			return compose.Input{}, fmt.Errorf("loading GuidanceCatalog: %w", err)
		}
		inp.GuidanceCatalog = gc

	case artifact.EvaluationLogArtifact:
		el, err := artifact.LoadEvaluationLog(path)
		if err != nil {
			return compose.Input{}, fmt.Errorf("loading EvaluationLog: %w", err)
		}
		inp.EvaluationLog = el

	case artifact.AuditLogArtifact:
		al, err := artifact.LoadAuditLog(path)
		if err != nil {
			return compose.Input{}, fmt.Errorf("loading AuditLog: %w", err)
		}
		inp.AuditLog = al

	case artifact.MappingDocumentArtifact:
		md, err := artifact.LoadMappingDocument(path)
		if err != nil {
			return compose.Input{}, fmt.Errorf("loading MappingDocument: %w", err)
		}
		inp.MappingDocument = md

	case artifact.RiskCatalogArtifact:
		rc, err := artifact.LoadRiskCatalog(path)
		if err != nil {
			return compose.Input{}, fmt.Errorf("loading RiskCatalog: %w", err)
		}
		inp.RiskCatalog = rc

	case artifact.PolicyArtifact:
		pol, err := artifact.LoadPolicy(path)
		if err != nil {
			return compose.Input{}, fmt.Errorf("loading Policy: %w", err)
		}
		inp.Policy = pol

	default:
		return compose.Input{}, fmt.Errorf("unsupported artifact type %s — appraise supports: ControlCatalog, GuidanceCatalog, EvaluationLog, AuditLog, MappingDocument, RiskCatalog, Policy", t.String())
	}

	return inp, nil
}

func usage() {
	fmt.Fprintln(os.Stderr, `appraise — framework-agnostic compliance narrative generator

Usage:
  appraise [flags] <artifact1.yaml> [artifact2.yaml] ...

Flags:
  --output string    Output Markdown path (default: compliance-narrative.md; use - for stdout)
  --program string   Program slug for document header and provenance (optional)
  --dry-run          Print detected artifact types without writing output
  --version          Print version and exit

Supported artifact types:
  ControlCatalog · GuidanceCatalog · EvaluationLog · AuditLog
  MappingDocument · RiskCatalog · Policy

Framework identity is derived from artifact metadata — no framework logic is hardcoded.
appraise works with any combination of gemara artifacts regardless of standard.
Sections marked [DATA NEEDED] require AI agent judgment (e.g. regimen) to complete.

Examples:
  appraise data/catalogs/iso42001.yaml
  appraise data/catalogs/iso42001.yaml data/risks/org-risks.yaml data/assessments/eval.yaml
  appraise --program myprogram --output reports/narrative.md iso42001.yaml eval.yaml
  appraise --dry-run iso42001.yaml eval.yaml`)
}
