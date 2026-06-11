package app

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"
)

type competitorWatchPacket struct {
	SchemaVersion         string                          `json:"schema_version"`
	ResultType            string                          `json:"result_type"`
	GeneratedAt           string                          `json:"generated_at"`
	Status                string                          `json:"status"`
	Mode                  string                          `json:"mode"`
	SourcePaths           []string                        `json:"source_paths"`
	CompetitorChanges     []competitorWatchChange         `json:"competitor_changes"`
	ScorecardDeltas       []competitorWatchDelta          `json:"scorecard_deltas"`
	RecommendedP0         []competitorWatchRecommendation `json:"recommended_p0"`
	StaleRequirementRisks []string                        `json:"stale_requirement_risks"`
	ReleasePlanIngestion  competitorWatchIngestion        `json:"release_plan_ingestion"`
}

type competitorWatchChange struct {
	Competitor     string `json:"competitor"`
	Change         string `json:"change"`
	Source         string `json:"source"`
	Classification string `json:"classification"`
}

type competitorWatchDelta struct {
	ID      string `json:"id"`
	Actual  int    `json:"actual,omitempty"`
	Minimum int    `json:"minimum,omitempty"`
	Status  string `json:"status"`
	Summary string `json:"summary"`
}

type competitorWatchRecommendation struct {
	ID             string `json:"id"`
	Priority       string `json:"priority"`
	Classification string `json:"classification"`
	Summary        string `json:"summary"`
	Reason         string `json:"reason"`
}

type competitorWatchIngestion struct {
	DecisionRecordRequired bool     `json:"decision_record_required"`
	ActiveScopeAllowed     bool     `json:"active_scope_allowed"`
	NextActions            []string `json:"next_actions"`
	CandidateUpdates       []string `json:"candidate_updates"`
	DeletionCandidates     []string `json:"deletion_candidates"`
}

func runCompetitorWatchCheck(args []string, stdout, stderr io.Writer) int {
	format, ok := parseOptionalFormatArgs(args)
	if !ok {
		fmt.Fprintln(stderr, "Unsupported competitor-watch format. Try `jini check competitor-watch` or `jini check competitor-watch --format json`.")
		return 1
	}
	packet := buildCompetitorWatchPacket(discoverSourceRoot())
	if format == "json" {
		encoder := json.NewEncoder(stdout)
		encoder.SetIndent("", "  ")
		encoder.SetEscapeHTML(false)
		if err := encoder.Encode(packet); err != nil {
			fmt.Fprintf(stderr, "Could not render competitor watch packet: %v\n", err)
			return 1
		}
		if packet.Status == "ok" {
			return 0
		}
		return 1
	}
	renderCompetitorWatchPacketText(stdout, packet)
	if packet.Status == "ok" {
		return 0
	}
	return 1
}

func buildCompetitorWatchPacket(root string) competitorWatchPacket {
	scorecard := buildScorecardGateReport(root)
	status := scorecard.Status
	if status == "" {
		status = "ok"
	}
	packet := competitorWatchPacket{
		SchemaVersion: "0.1.0",
		ResultType:    "JiniCompetitorWatchPacket",
		GeneratedAt:   time.Now().UTC().Format(time.RFC3339),
		Status:        status,
		Mode:          "local-baseline",
		SourcePaths: []string{
			"specs/golden-competitive-benchmark.yaml",
			"specs/competitive-release-plan.md",
			"specs/number-one-platform-prd.md",
		},
		CompetitorChanges: []competitorWatchChange{{
			Competitor:     "local-baseline",
			Change:         "No live source-backed competitor deltas were supplied in this run.",
			Source:         "specs/competitive-release-plan.md",
			Classification: "watch",
		}},
		ScorecardDeltas:       buildCompetitorWatchDeltas(scorecard),
		RecommendedP0:         buildCompetitorWatchP0(scorecard),
		StaleRequirementRisks: buildCompetitorWatchStaleRisks(),
		ReleasePlanIngestion:  buildCompetitorWatchIngestion(scorecard),
	}
	return packet
}

func buildCompetitorWatchDeltas(scorecard scorecardGateReport) []competitorWatchDelta {
	deltas := make([]competitorWatchDelta, 0, len(scorecard.Checks)+2)
	for _, check := range scorecard.Checks {
		deltas = append(deltas, competitorWatchDelta{
			ID:      check.ID,
			Actual:  check.Actual,
			Minimum: check.Minimum,
			Status:  check.Status,
			Summary: fmt.Sprintf("%d/%d", check.Actual, check.Minimum),
		})
	}
	if scorecard.PRDImplementation.TotalRequirements > 0 {
		deltas = append(deltas, competitorWatchDelta{
			ID:      "p0-prd-trace",
			Actual:  scorecard.PRDImplementation.ImplementedRequirements,
			Minimum: scorecard.PRDImplementation.TotalRequirements,
			Status:  scorecard.PRDImplementation.Status,
			Summary: fmt.Sprintf("%d/%d P0 requirements, %d%% complete", scorecard.PRDImplementation.ImplementedRequirements, scorecard.PRDImplementation.TotalRequirements, scorecard.PRDImplementation.CompletionPercent),
		})
	}
	if scorecard.PRDImplementation.ResidualHardeningCount > 0 {
		deltas = append(deltas, competitorWatchDelta{
			ID:      "residual-hardening",
			Actual:  scorecard.PRDImplementation.ResidualHardeningCount,
			Minimum: 0,
			Status:  "watch",
			Summary: firstNonEmpty(strings.Join(scorecard.PRDImplementation.ResidualHardening, " "), "Residual hardening remains."),
		})
	}
	return deltas
}

func buildCompetitorWatchP0(scorecard scorecardGateReport) []competitorWatchRecommendation {
	recommendations := []competitorWatchRecommendation{
		{
			ID:             "source-backed-refresh",
			Priority:       "P0",
			Classification: "adopt",
			Summary:        "Add source-backed competitor refresh before changing active roadmap scope.",
			Reason:         "The local packet separates standing pressure from net-new competitor deltas.",
		},
		{
			ID:             "decision-record-ingestion",
			Priority:       "P0",
			Classification: "adopt",
			Summary:        "Route every competitor-derived candidate through a decision record before PRD or roadmap changes.",
			Reason:         "The PRD says competitor watch can nominate candidates but cannot create active scope by itself.",
		},
		{
			ID:             "real-cli-dogfood",
			Priority:       "P0",
			Classification: "integrate",
			Summary:        "Keep Wave 1 real installed CLI dogfood as release hardening.",
			Reason:         "Adapter breadth is valuable only when command templates, auth, approvals, and receipt privacy work on tester machines.",
		},
	}
	if scorecard.PRDImplementation.ResidualHardeningCount == 0 {
		recommendations = recommendations[:2]
	}
	return recommendations
}

func buildCompetitorWatchStaleRisks() []string {
	return []string{
		"Do not copy IDE, local model host, or prompt-to-app surfaces unless they improve replacement-critical Jini claims.",
		"No competitor finding becomes active scope unless the decision record changes.",
		"Reject requirements that add context bloat, new ramp-up, or vague automation without receipts.",
	}
}

func buildCompetitorWatchIngestion(scorecard scorecardGateReport) competitorWatchIngestion {
	actions := []string{
		"No competitor finding becomes active scope unless the decision record changes.",
		"Refresh public sources, classify as adopt/integrate/watch/reject/delete, then update the release plan.",
	}
	if scorecard.PRDImplementation.ResidualHardeningCount > 0 {
		actions = append(actions, "Keep residual hardening visible until real installed CLI dogfood evidence is fresh.")
	}
	return competitorWatchIngestion{
		DecisionRecordRequired: true,
		ActiveScopeAllowed:     false,
		NextActions:            actions,
		CandidateUpdates: []string{
			"source-backed competitor refresh",
			"decision-record ingestion for candidate and deletion nominations",
			"real installed CLI dogfood evidence",
		},
		DeletionCandidates: []string{
			"copycat IDE requirements",
			"local model host rebuild requirements",
			"prompt-to-app flagship scope before CLI quality wins",
		},
	}
}

func renderCompetitorWatchPacketText(w io.Writer, packet competitorWatchPacket) {
	fmt.Fprintln(w, "Competitor watch")
	fmt.Fprintf(w, "Mode: %s\n", packet.Mode)
	fmt.Fprintf(w, "Status: %s\n", packet.Status)
	fmt.Fprintln(w, "Scorecard deltas")
	for _, delta := range firstCompetitorWatchDeltas(packet.ScorecardDeltas, 2) {
		fmt.Fprintf(w, "- %s: %s (%s)\n", delta.ID, delta.Summary, delta.Status)
	}
	fmt.Fprintln(w, "Recommended P0")
	for _, recommendation := range firstCompetitorWatchRecommendations(packet.RecommendedP0, 2) {
		fmt.Fprintf(w, "- %s: %s\n", recommendation.Classification, recommendation.Summary)
	}
	fmt.Fprintln(w, "Stale risks")
	for _, risk := range firstCompetitorWatchStrings(packet.StaleRequirementRisks, 2) {
		fmt.Fprintf(w, "- %s\n", risk)
	}
	fmt.Fprintln(w, "Next: source-backed refresh, decision record, then release-plan update.")
}

func firstCompetitorWatchDeltas(items []competitorWatchDelta, limit int) []competitorWatchDelta {
	if len(items) <= limit {
		return items
	}
	return items[:limit]
}

func firstCompetitorWatchRecommendations(items []competitorWatchRecommendation, limit int) []competitorWatchRecommendation {
	if len(items) <= limit {
		return items
	}
	return items[:limit]
}

func firstCompetitorWatchStrings(items []string, limit int) []string {
	if len(items) <= limit {
		return items
	}
	return items[:limit]
}
