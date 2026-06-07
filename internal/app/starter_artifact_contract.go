package app

import "strings"

type starterArtifactContract struct {
	RequestCohort  string
	ArtifactFamily string
	ProviderLines  []string
	Quality        draftQualityProfile
	Transforms     map[string]artifactTransformProfile
}

type artifactTransformProfile struct {
	CoreHeading  string
	GapHeading   string
	NextHeading  string
	WatchHeading string
}

func starterArtifactContractForPack(packID string) starterArtifactContract {
	return starterArtifactContractForShape(starterRequestCohort(packID), starterArtifactFamily(packID), "")
}

func starterArtifactContractForSummary(summary *workSummary) starterArtifactContract {
	if summary == nil {
		return starterArtifactContractForShape("", "", "")
	}
	return starterArtifactContractForPack(summary.PackID)
}

func starterArtifactContractForRequest(request providerGenerationRequest) starterArtifactContract {
	return starterArtifactContractForShape(
		classifyRequestCohort(request),
		classifyArtifactFamily(request),
		classifyRouteModalitySubtype(request),
	)
}

func starterArtifactContractForShape(cohort, family, modalitySubtype string) starterArtifactContract {
	cohort = strings.TrimSpace(cohort)
	family = strings.TrimSpace(family)
	base := starterArtifactContract{
		RequestCohort:  cohort,
		ArtifactFamily: family,
		ProviderLines: []string{
			"Shape the answer as the first useful artifact for this work type.",
			"When the source already includes a named reference or URL that belongs in the artifact, preserve it as one smart Markdown link on first mention.",
		},
		Quality: draftQualityProfile{
			Cohort:                 cohort,
			ArtifactFamily:         family,
			ModalitySubtype:        strings.TrimSpace(modalitySubtype),
			RequiredHeadings:       []string{"## Still to confirm"},
			RequiredHeadingWeight:  6,
			PreferredHeadingWeight: 3,
			EvidenceWeight:         0,
			UncertaintyWeight:      4,
		},
		Transforms: map[string]artifactTransformProfile{
			"shorter": {
				CoreHeading: "Short version",
				GapHeading:  "Still to confirm",
				NextHeading: "Next move",
			},
			"executive": {
				CoreHeading: "Executive summary",
				GapHeading:  "Risks or gaps",
				NextHeading: "Next move",
			},
			"checklist": {
				CoreHeading:  "Do now",
				GapHeading:   "Confirm",
				WatchHeading: "Watch",
			},
		},
	}

	switch {
	case cohort == "sendable-followup" || family == "narrative-draft":
		base.ProviderLines = []string{
			"Return a sendable follow-up note first.",
			"Use these sections exactly:",
			"- `## Send this note`",
			"- `## Decisions captured from the notes`",
			"- `## Owners and due dates to confirm`",
			"- `## Open questions to close`",
			"- `## Recommended next move`",
			"Do not invent names, dates, or commitments that are not grounded in the source.",
		}
		base.Quality.RequiredHeadings = []string{
			"## Send this note",
			"## Decisions captured from the notes",
			"## Owners and due dates to confirm",
			"## Open questions to close",
			"## Recommended next move",
		}
		base.Quality.PreferredHeadings = []string{"## Still to confirm"}
	case cohort == "build-readiness" || family == "structured-check":
		base.ProviderLines = []string{
			"Return a build-readiness artifact, not a vague summary.",
			"Use these sections exactly:",
			"- `## What looks ready now`",
			"- `## Must clear before build`",
			"- `## Recommended first slice`",
			"- `## Who needs to answer what`",
			"- `## Still to confirm`",
			"Do not reduce the answer to a binary verdict. Keep missing proof and approval gaps visible.",
		}
		base.Quality.RequiredHeadings = []string{
			"## What looks ready now",
			"## Must clear before build",
			"## Recommended first slice",
			"## Who needs to answer what",
			"## Still to confirm",
		}
		base.Quality.PreferredHeadings = []string{"## Risks", "## Approval gaps"}
		base.Transforms["shorter"] = artifactTransformProfile{
			CoreHeading: "What looks ready now",
			GapHeading:  "Must clear before build",
			NextHeading: "Recommended first slice",
		}
	case cohort == "option-compare" || family == "comparison-matrix":
		base.ProviderLines = []string{
			"Return a recommendation artifact, not a generic comparison summary.",
			"Use these sections exactly:",
			"- `## Recommendation`",
			"- `## Tradeoffs`",
			"- `## Risks`",
			"- `## Next move`",
			"- `## Still to confirm`",
		}
		base.Quality.RequiredHeadings = []string{"## Recommendation", "## Risks", "## Still to confirm"}
		base.Quality.PreferredHeadings = []string{"## Tradeoffs", "## Next move"}
		base.Quality.RequiredHeadingWeight = 7
		base.Transforms["shorter"] = artifactTransformProfile{
			CoreHeading: "Recommendation",
			GapHeading:  "Risks",
			NextHeading: "Next move",
		}
	case cohort == "trip-itinerary" || family == "itinerary-plan":
		base.ProviderLines = []string{
			"Return a trip-planning artifact, not a generic travel essay.",
			"Use these sections exactly:",
			"- `## Day by day`",
			"- `## Budget`",
			"- `## Travel logistics`",
			"- `## Still to book`",
			"- `## Still to confirm`",
			"When a key destination, museum, or landmark is clearly part of the plan, add one smart Markdown link on first mention.",
			"Prefer canonical destination links and avoid turning every bullet into a link list.",
		}
		base.Quality.RequiredHeadings = []string{"## Day by day", "## Still to confirm"}
		base.Quality.PreferredHeadings = []string{"## Budget", "## Travel logistics", "## Still to book"}
		base.Quality.RequiredHeadingWeight = 7
		base.Transforms["shorter"] = artifactTransformProfile{
			CoreHeading: "Trip at a glance",
			GapHeading:  "Still to confirm",
			NextHeading: "Booking priorities",
		}
	case family == "multimodal-extract" || cohort == "multimodal-extract":
		base.ProviderLines = multimodalArtifactGuidance(modalitySubtype)
		base.Quality.RequiredHeadings = []string{
			"## Extracted evidence",
			"## What the source shows",
			"## Still unclear",
		}
		base.Quality.PreferredHeadings = []string{"## Recommended next move", "## Confidence notes"}
		base.Quality.EvidenceSignals = []string{
			"image", "screenshot", "pdf", "audio", "recording", "document",
			"source", "evidence", "shows", "visible", "transcript", "scan",
		}
		base.Quality.RequiredHeadingWeight = 8
		base.Quality.PreferredHeadingWeight = 4
		base.Quality.EvidenceWeight = 5
		base.Quality.UncertaintyWeight = 5
		switch strings.TrimSpace(modalitySubtype) {
		case "pdf-scan":
			base.Quality.RequiredHeadings = []string{
				"## Extracted evidence",
				"## What the document shows",
				"## Still unclear",
			}
			base.Quality.PreferredHeadings = []string{"## Recommended next move", "## OCR or confidence notes"}
			base.Quality.EvidenceSignals = []string{
				"pdf", "document", "page", "scan", "ocr", "text", "field",
				"label", "signature", "table", "section",
			}
		case "image-screenshot":
			base.Quality.RequiredHeadings = []string{
				"## Extracted evidence",
				"## What is visible",
				"## Still unclear",
			}
			base.Quality.PreferredHeadings = []string{"## Recommended next move", "## Confidence notes"}
			base.Quality.EvidenceSignals = []string{
				"image", "screenshot", "visible", "screen", "button", "label",
				"panel", "photo", "diagram", "highlight",
			}
		case "audio-transcript":
			base.Quality.RequiredHeadings = []string{
				"## Extracted evidence",
				"## What the recording says",
				"## Still unclear",
			}
			base.Quality.PreferredHeadings = []string{"## Recommended next move", "## Confidence notes"}
			base.Quality.EvidenceSignals = []string{
				"audio", "recording", "voice", "transcript", "speaker",
				"said", "heard", "timecode", "quote",
			}
		}
	}

	base.Quality.Cohort = cohort
	base.Quality.ArtifactFamily = family
	base.Quality.ModalitySubtype = strings.TrimSpace(modalitySubtype)
	return base
}

func multimodalArtifactGuidance(modalitySubtype string) []string {
	switch strings.TrimSpace(modalitySubtype) {
	case "pdf-scan":
		return []string{
			"Return an evidence-grounded document extraction artifact, not a generic summary.",
			"Use these sections exactly:",
			"- `## Extracted evidence`",
			"- `## What the document shows`",
			"- `## Still unclear`",
			"- `## Recommended next move`",
			"Call out what is visible in the PDF or scan, what OCR may have missed, and what still needs verification.",
		}
	case "image-screenshot":
		return []string{
			"Return an evidence-grounded image or screenshot extraction artifact, not a generic summary.",
			"Use these sections exactly:",
			"- `## Extracted evidence`",
			"- `## What is visible`",
			"- `## Still unclear`",
			"- `## Recommended next move`",
			"Call out what is visually present, what may be obscured, and what still needs verification.",
		}
	case "audio-transcript":
		return []string{
			"Return an evidence-grounded audio extraction artifact, not a generic summary.",
			"Use these sections exactly:",
			"- `## Extracted evidence`",
			"- `## What the recording says`",
			"- `## Still unclear`",
			"- `## Recommended next move`",
			"Call out what is directly supported by the recording or transcript, what may have been misheard, and what still needs verification.",
		}
	default:
		return []string{
			"Return an evidence-grounded extraction artifact, not a generic summary.",
			"Use these sections exactly:",
			"- `## Extracted evidence`",
			"- `## What the source shows`",
			"- `## Still unclear`",
			"- `## Recommended next move`",
			"Call out what came from the source, what is ambiguous, and what still needs verification.",
		}
	}
}
