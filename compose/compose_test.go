package compose_test

import (
	"strings"
	"testing"
	"time"

	"github.com/Formulary-Labs/appraise/compose"
	"github.com/Formulary-Labs/substrate/artifact"
)

func TestComposeEmpty(t *testing.T) {
	result := compose.Compose(compose.Options{
		GeneratedAt: time.Date(2026, 9, 21, 0, 0, 0, 0, time.UTC),
	})
	if !strings.Contains(result, "# Compliance Narrative") {
		t.Error("expected header in output")
	}
	if !strings.Contains(result, "[DATA NEEDED") {
		t.Error("expected [DATA NEEDED] placeholders when no artifacts provided")
	}
}

func TestComposeWithControlCatalog(t *testing.T) {
	cat := &artifact.ControlCatalog{
		Title: "ISO 42001 Control Catalog",
		Controls: []artifact.Control{
			{Id: "6.1", Title: "Roles and responsibilities"},
		},
		Groups: []artifact.Group{
			{Id: "org", Title: "Organizational Controls"},
		},
	}
	cat.Metadata.Description = "ISO/IEC 42001:2023"
	cat.Metadata.Id = "iso42001"

	result := compose.Compose(compose.Options{
		Inputs: []compose.Input{
			{
				Path:           "iso42001.yaml",
				Type:           artifact.ControlCatalogArtifact,
				ControlCatalog: cat,
			},
		},
		Program:     "test-program",
		GeneratedAt: time.Date(2026, 9, 21, 0, 0, 0, 0, time.UTC),
	})

	if !strings.Contains(result, "ISO 42001 Control Catalog") {
		t.Error("expected catalog title in output")
	}
	if !strings.Contains(result, "ISO/IEC 42001:2023") {
		t.Error("expected framework description in output")
	}
	if !strings.Contains(result, "test-program") {
		t.Error("expected program name in output")
	}
	// Section 2 should be populated; no DATA NEEDED for control framework.
	if strings.Contains(result, "[DATA NEEDED: control framework") {
		t.Error("control framework section should be populated when ControlCatalog provided")
	}
}

func TestComposeArtifactSourcesTable(t *testing.T) {
	result := compose.Compose(compose.Options{
		Inputs: []compose.Input{
			{
				Path:           "my-catalog.yaml",
				Type:           artifact.ControlCatalogArtifact,
				ControlCatalog: &artifact.ControlCatalog{},
			},
		},
		GeneratedAt: time.Now(),
	})
	if !strings.Contains(result, "my-catalog.yaml") {
		t.Error("expected artifact source path in sources table")
	}
	if !strings.Contains(result, "ControlCatalog") {
		t.Error("expected artifact type in sources table")
	}
}
