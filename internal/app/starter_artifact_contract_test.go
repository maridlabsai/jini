package app

import (
	"strings"
	"testing"
)

func TestStarterArtifactContractForPackUsesSharedMeetingGuidance(t *testing.T) {
	contract := starterArtifactContractForPack("meeting-followup")

	if contract.RequestCohort != "sendable-followup" {
		t.Fatalf("expected sendable-followup cohort, got %q", contract.RequestCohort)
	}
	for _, want := range []string{
		"Return a sendable follow-up note first.",
		"`## Send this note`",
		"`## Recommended next move`",
	} {
		if !containsLine(contract.ProviderLines, want) {
			t.Fatalf("expected provider guidance to contain %q, got %#v", want, contract.ProviderLines)
		}
	}
	if got := contract.Transforms["shorter"].CoreHeading; got != "Short version" {
		t.Fatalf("expected narrative shorter heading, got %q", got)
	}
}

func TestStarterArtifactContractForPackUsesSharedBuildReadinessTransforms(t *testing.T) {
	contract := starterArtifactContractForPack("research-prd")

	if contract.ArtifactFamily != "structured-check" {
		t.Fatalf("expected structured-check family, got %q", contract.ArtifactFamily)
	}
	for _, want := range []string{
		"## What looks ready now",
		"## Must clear before build",
		"## Recommended first slice",
	} {
		if !containsString(contract.Quality.RequiredHeadings, want) {
			t.Fatalf("expected quality headings to contain %q, got %#v", want, contract.Quality.RequiredHeadings)
		}
	}
	shorter := contract.Transforms["shorter"]
	if shorter.CoreHeading != "What looks ready now" || shorter.GapHeading != "Must clear before build" || shorter.NextHeading != "Recommended first slice" {
		t.Fatalf("unexpected structured-check shorter transform: %#v", shorter)
	}
}

func TestStarterArtifactContractForShapeUsesMultimodalSubtype(t *testing.T) {
	contract := starterArtifactContractForShape("multimodal-extract", "multimodal-extract", "pdf-scan")

	if !containsLine(contract.ProviderLines, "`## What the document shows`") {
		t.Fatalf("expected pdf guidance, got %#v", contract.ProviderLines)
	}
	if !containsString(contract.Quality.RequiredHeadings, "## What the document shows") {
		t.Fatalf("expected pdf required heading, got %#v", contract.Quality.RequiredHeadings)
	}
	if !containsString(contract.Quality.EvidenceSignals, "ocr") {
		t.Fatalf("expected pdf evidence signal set to include ocr, got %#v", contract.Quality.EvidenceSignals)
	}
}

func containsLine(lines []string, want string) bool {
	for _, line := range lines {
		if strings.Contains(line, want) {
			return true
		}
	}
	return false
}

func containsString(items []string, want string) bool {
	for _, item := range items {
		if strings.TrimSpace(item) == strings.TrimSpace(want) {
			return true
		}
	}
	return false
}
