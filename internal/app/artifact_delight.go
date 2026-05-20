package app

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type artifactVersionLedger struct {
	SchemaVersion string                 `json:"schema_version"`
	ContextType   string                 `json:"context_type"`
	Entries       []artifactVersionEntry `json:"entries"`
}

type artifactVersionEntry struct {
	RevisionID    string `json:"revision_id"`
	ArtifactPath  string `json:"artifact_path"`
	ArtifactLabel string `json:"artifact_label"`
	ActionLabel   string `json:"action_label"`
	SnapshotPath  string `json:"snapshot_path"`
}

func renderContextCapsule(w io.Writer, summary *workSummary) {
	fmt.Fprintln(w, "What Jini used")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "From you")
	inputs := []string{}
	for _, item := range summary.Thread.InputItems {
		if line := contextInputLine(item); line != "" {
			inputs = append(inputs, line)
		}
	}
	if len(inputs) == 0 {
		fmt.Fprintln(w, "- No direct input is saved for this work yet")
	} else {
		for _, line := range inputs {
			fmt.Fprintf(w, "- %s\n", line)
		}
	}
	links := sourceReferenceLinks(sourceFromInputItems(summary.Thread.InputItems))
	if len(links) > 0 {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Links and sources")
		for _, link := range links {
			label := firstNonEmpty(strings.TrimSpace(firstLabel(link)), strings.TrimSpace(link.URL))
			fmt.Fprintf(w, "- %s: %s\n", label, strings.TrimSpace(link.URL))
		}
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Kept visible")
	gaps := contextGapLines(summary)
	for _, line := range gaps {
		fmt.Fprintf(w, "- %s\n", line)
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Route and continuity")
	fmt.Fprintf(w, "- Current route: %s\n", firstNonEmpty(strings.TrimSpace(summary.Thread.CurrentRoute), "Not recorded"))
	if strings.TrimSpace(summary.Thread.RouteReason) != "" {
		fmt.Fprintf(w, "- Why this route: %s\n", summary.Thread.RouteReason)
	}
	if strings.TrimSpace(summary.Thread.ContinuityReason) != "" {
		fmt.Fprintf(w, "- Continuity: %s\n", summary.Thread.ContinuityReason)
	}
}

func contextInputLine(item inputItem) string {
	title := strings.TrimSpace(item.Title)
	preview := strings.TrimSpace(item.Preview)
	switch item.Kind {
	case "text":
		if title != "" && preview != "" {
			return title + ": " + preview
		}
	case "clarification":
		if title != "" && preview != "" {
			return title + ": " + preview
		}
	}
	return formatInputItem(item)
}

func contextGapLines(summary *workSummary) []string {
	lines := []string{}
	for _, item := range summary.Missing {
		if strings.TrimSpace(item) == "" {
			continue
		}
		lines = append(lines, item)
	}
	for _, item := range summary.Uncertain {
		if strings.TrimSpace(item) == "" {
			continue
		}
		lines = append(lines, "Not sure: "+item)
	}
	if len(lines) == 0 {
		return []string{"Nothing right now"}
	}
	return uniqueDelightStrings(lines)
}

func applyArtifactTransform(summary *workSummary, mode string) (*catalogItem, error) {
	item := currentArtifactItem(summary)
	if item == nil {
		return nil, fmt.Errorf("no ready artifact is available to revise")
	}
	content, err := os.ReadFile(item.Path)
	if err != nil {
		return nil, err
	}
	actionLabel := artifactTransformLabel(mode)
	if err := saveArtifactVersion(summary.Dir, *item, actionLabel, string(content)); err != nil {
		return nil, err
	}
	updated := transformArtifactContent(summary, *item, string(content), mode)
	if err := os.WriteFile(item.Path, []byte(updated), 0o644); err != nil {
		return nil, err
	}
	updateThreadArtifactState(summary.Dir, item.Label, actionLabel)
	return item, nil
}

func renderArtifactVersions(w io.Writer, summary *workSummary) {
	item := currentArtifactItem(summary)
	fmt.Fprintln(w, "Versions")
	if item == nil {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "No ready artifact is available yet.")
		return
	}
	entries := artifactVersionsForItem(summary.Dir, *item)
	fmt.Fprintln(w)
	if len(entries) == 0 {
		fmt.Fprintln(w, "No saved versions yet.")
		fmt.Fprintln(w, "Jini will save one before the first artifact rewrite.")
		return
	}
	for index, entry := range entries {
		fmt.Fprintf(w, "%d. Before %s\n", index+1, entry.ActionLabel)
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Type `Undo last change` to restore the latest version.")
}

func undoLastArtifactChange(summary *workSummary) (*catalogItem, error) {
	item := currentArtifactItem(summary)
	if item == nil {
		return nil, fmt.Errorf("no ready artifact is available to restore")
	}
	ledger := loadArtifactVersionLedger(summary.Dir)
	targetPath := artifactRelativePath(summary.Dir, item.Path)
	latestIndex := -1
	for index := len(ledger.Entries) - 1; index >= 0; index-- {
		if ledger.Entries[index].ArtifactPath == targetPath {
			latestIndex = index
			break
		}
	}
	if latestIndex < 0 {
		return nil, fmt.Errorf("no saved versions exist for this artifact yet")
	}
	entry := ledger.Entries[latestIndex]
	snapshotPath := filepath.Join(summary.Dir, entry.SnapshotPath)
	content, err := os.ReadFile(snapshotPath)
	if err != nil {
		return nil, err
	}
	if err := os.WriteFile(item.Path, content, 0o644); err != nil {
		return nil, err
	}
	ledger.Entries = append([]artifactVersionEntry{}, append(ledger.Entries[:latestIndex], ledger.Entries[latestIndex+1:]...)...)
	if err := saveArtifactVersionLedger(summary.Dir, ledger); err != nil {
		return nil, err
	}
	_ = os.Remove(snapshotPath)
	updateThreadArtifactState(summary.Dir, item.Label, "Undo last change")
	return item, nil
}

func currentArtifactItem(summary *workSummary) *catalogItem {
	if summary == nil {
		return nil
	}
	if state := loadThreadState(summary.Dir, summary); state.CurrentFocus != nil {
		if item := focusedArtifactItem(summary, state.CurrentFocus); item != nil {
			return item
		}
	}
	if item := firstResultItem(summary); item != nil {
		return item
	}
	if len(summary.Views) > 0 {
		return &summary.Views[0]
	}
	return nil
}

func hasArtifactVersions(summary *workSummary) bool {
	item := currentArtifactItem(summary)
	if item == nil {
		return false
	}
	return len(artifactVersionsForItem(summary.Dir, *item)) > 0
}

func artifactVersionsForItem(workDir string, item catalogItem) []artifactVersionEntry {
	ledger := loadArtifactVersionLedger(workDir)
	targetPath := artifactRelativePath(workDir, item.Path)
	entries := []artifactVersionEntry{}
	for index := len(ledger.Entries) - 1; index >= 0; index-- {
		if ledger.Entries[index].ArtifactPath == targetPath {
			entries = append(entries, ledger.Entries[index])
		}
	}
	return entries
}

func saveArtifactVersion(workDir string, item catalogItem, actionLabel, content string) error {
	if strings.TrimSpace(workDir) == "" {
		return nil
	}
	ledger := loadArtifactVersionLedger(workDir)
	versionDir := filepath.Join(workDir, "history")
	if err := os.MkdirAll(versionDir, 0o755); err != nil {
		return err
	}
	next := len(ledger.Entries) + 1
	revisionID := fmt.Sprintf("%03d", next)
	snapshotName := fmt.Sprintf("%s-%s.md", revisionID, normalizeFilename(item.Label))
	snapshotPath := filepath.Join("history", snapshotName)
	if err := os.WriteFile(filepath.Join(workDir, snapshotPath), []byte(content), 0o644); err != nil {
		return err
	}
	ledger.Entries = append(ledger.Entries, artifactVersionEntry{
		RevisionID:    revisionID,
		ArtifactPath:  artifactRelativePath(workDir, item.Path),
		ArtifactLabel: item.Label,
		ActionLabel:   actionLabel,
		SnapshotPath:  snapshotPath,
	})
	return saveArtifactVersionLedger(workDir, ledger)
}

func artifactVersionLedgerPath(workDir string) string {
	return filepath.Join(workDir, "artifact-versions.json")
}

func loadArtifactVersionLedger(workDir string) artifactVersionLedger {
	data, err := os.ReadFile(artifactVersionLedgerPath(workDir))
	if err != nil {
		return artifactVersionLedger{SchemaVersion: "0.1.0", ContextType: "JiniArtifactVersions"}
	}
	var ledger artifactVersionLedger
	if err := json.Unmarshal(data, &ledger); err != nil {
		return artifactVersionLedger{SchemaVersion: "0.1.0", ContextType: "JiniArtifactVersions"}
	}
	if ledger.SchemaVersion == "" {
		ledger.SchemaVersion = "0.1.0"
	}
	if ledger.ContextType == "" {
		ledger.ContextType = "JiniArtifactVersions"
	}
	return ledger
}

func saveArtifactVersionLedger(workDir string, ledger artifactVersionLedger) error {
	ledger.SchemaVersion = "0.1.0"
	ledger.ContextType = "JiniArtifactVersions"
	data, err := json.MarshalIndent(ledger, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(artifactVersionLedgerPath(workDir), append(data, '\n'), 0o600)
}

func artifactRelativePath(workDir, path string) string {
	rel, err := filepath.Rel(workDir, path)
	if err != nil {
		return filepath.Clean(path)
	}
	return filepath.Clean(rel)
}

func artifactTransformLabel(mode string) string {
	switch normalizeName(mode) {
	case "shorter":
		return "Make it shorter"
	case "executive":
		return "Make it executive"
	case "checklist":
		return "Turn this into a checklist"
	default:
		return "Revise this artifact"
	}
}

func transformArtifactContent(summary *workSummary, item catalogItem, content, mode string) string {
	sections := artifactSections(content)
	title := artifactDocumentTitle(item, sections)
	core := collectArtifactCoreBullets(content, 4)
	if len(core) == 0 {
		core = []string{"Keep the current artifact focused on the main outcome."}
	}
	gaps := collectArtifactGapBullets(summary, sections)
	nextMove := collectArtifactNextMove(summary, sections)
	profile := artifactTransformProfileForSummary(summary, mode)

	switch normalizeName(mode) {
	case "executive":
		return strings.Join([]string{
			"# " + title,
			"",
			"## " + profile.CoreHeading,
			bulletLines(limitStrings(core, 3)),
			"## " + profile.GapHeading,
			bulletLines(limitStrings(gaps, 3)),
			"## " + profile.NextHeading,
			bulletLines(limitStrings(nextMove, 2)),
		}, "\n")
	case "checklist":
		return strings.Join([]string{
			"# " + title,
			"",
			"## " + profile.CoreHeading,
			checkboxLines(limitStrings(append([]string{}, nextMove...), 3)),
			"## " + profile.GapHeading,
			checkboxLines(limitStrings(gaps, 4)),
			"## " + profile.WatchHeading,
			checkboxLines(limitStrings(collectArtifactWatchBullets(summary, sections), 3)),
		}, "\n")
	default:
		return strings.Join([]string{
			"# " + title,
			"",
			"## " + profile.CoreHeading,
			bulletLines(limitStrings(core, 3)),
			"## " + profile.GapHeading,
			bulletLines(limitStrings(gaps, 3)),
			"## " + profile.NextHeading,
			bulletLines(limitStrings(nextMove, 2)),
		}, "\n")
	}
}

func artifactTransformProfileForSummary(summary *workSummary, mode string) artifactTransformProfile {
	contract := starterArtifactContractForSummary(summary)
	profile, ok := contract.Transforms[normalizeName(mode)]
	if !ok {
		profile = artifactTransformProfile{
			CoreHeading:  "Short version",
			GapHeading:   "Still to confirm",
			NextHeading:  "Next move",
			WatchHeading: "Watch",
		}
	}
	if strings.TrimSpace(profile.CoreHeading) == "" {
		profile.CoreHeading = "Short version"
	}
	if strings.TrimSpace(profile.GapHeading) == "" {
		profile.GapHeading = "Still to confirm"
	}
	if strings.TrimSpace(profile.NextHeading) == "" {
		profile.NextHeading = "Next move"
	}
	if strings.TrimSpace(profile.WatchHeading) == "" {
		profile.WatchHeading = "Watch"
	}
	return profile
}

func artifactDocumentTitle(item catalogItem, sections []artifactSection) string {
	for _, section := range sections {
		if strings.TrimSpace(section.Heading) == "" || normalizeName(section.Heading) == "document" {
			continue
		}
		return strings.TrimSpace(section.Heading)
	}
	return firstNonEmpty(strings.TrimSpace(item.Label), "Artifact")
}

func collectArtifactCoreBullets(content string, limit int) []string {
	sections := artifactSections(content)
	out := []string{}
	for _, section := range sections {
		if artifactSectionRole(section.Heading) != "core" {
			continue
		}
		out = append(out, extractSnippetBullets(section.Body, 2)...)
		if len(out) >= limit {
			break
		}
	}
	if len(out) == 0 {
		out = extractSnippetBullets(content, limit)
	}
	return uniqueDelightStrings(limitStrings(out, limit))
}

func collectArtifactGapBullets(summary *workSummary, sections []artifactSection) []string {
	out := []string{}
	if summary != nil {
		out = append(out, summary.Missing...)
		out = append(out, summary.Uncertain...)
		if strings.TrimSpace(summary.Thread.Need) != "" && summary.Thread.Need != "Nothing right now" {
			out = append(out, summary.Thread.Need)
		}
	}
	for _, section := range sections {
		heading := normalizeName(section.Heading)
		if !containsAny(heading, []string{"confirm", "missing", "open questions", "blocked", "risk", "still to confirm"}) {
			continue
		}
		out = append(out, extractSnippetBullets(section.Body, 2)...)
	}
	if len(out) == 0 {
		return []string{"Nothing right now"}
	}
	return uniqueDelightStrings(out)
}

func collectArtifactWatchBullets(summary *workSummary, sections []artifactSection) []string {
	out := []string{}
	if summary != nil {
		out = append(out, summary.Uncertain...)
	}
	for _, section := range sections {
		heading := normalizeName(section.Heading)
		if !containsAny(heading, []string{"risk", "open questions", "watch", "unclear", "blocked"}) {
			continue
		}
		out = append(out, extractSnippetBullets(section.Body, 2)...)
	}
	if len(out) == 0 {
		return []string{"Nothing right now"}
	}
	return uniqueDelightStrings(out)
}

func collectArtifactNextMove(summary *workSummary, sections []artifactSection) []string {
	out := []string{}
	for _, section := range sections {
		heading := normalizeName(section.Heading)
		if !containsAny(heading, []string{"recommended next move", "next move", "next step", "recommended first slice"}) {
			continue
		}
		out = append(out, extractSnippetBullets(section.Body, 2)...)
	}
	if summary != nil && strings.TrimSpace(summary.NextStep) != "" {
		out = append(out, summary.NextStep)
	}
	if len(out) == 0 {
		return []string{"Review the current artifact before changing anything external."}
	}
	return uniqueDelightStrings(out)
}

func extractSnippetBullets(body string, limit int) []string {
	lines := strings.Split(strings.ReplaceAll(body, "\r\n", "\n"), "\n")
	out := []string{}
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* ") {
			out = append(out, strings.TrimSpace(trimmed[2:]))
		}
		if len(out) >= limit {
			return uniqueDelightStrings(out)
		}
	}
	if len(out) == 0 {
		fragments := splitSourceFragments(body)
		for _, fragment := range fragments {
			clean := strings.TrimSpace(fragment)
			if clean == "" {
				continue
			}
			out = append(out, clean)
			if len(out) >= limit {
				break
			}
		}
	}
	if len(out) == 0 {
		clean := compactPreview(body, 120)
		if clean != "" {
			out = append(out, clean)
		}
	}
	return uniqueDelightStrings(out)
}

func checkboxLines(items []string) string {
	if len(items) == 0 {
		return "- [ ] Nothing right now\n"
	}
	lines := make([]string, 0, len(items)+1)
	for _, item := range items {
		lines = append(lines, "- [ ] "+item)
	}
	lines = append(lines, "")
	return strings.Join(lines, "\n")
}

func uniqueDelightStrings(items []string) []string {
	seen := map[string]bool{}
	out := []string{}
	for _, item := range items {
		clean := strings.TrimSpace(item)
		if clean == "" {
			continue
		}
		key := normalizeName(clean)
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, clean)
	}
	return out
}

func limitStrings(items []string, limit int) []string {
	if limit <= 0 || len(items) <= limit {
		return items
	}
	return append([]string{}, items[:limit]...)
}

func firstLabel(link smartLink) string {
	if len(link.Labels) == 0 {
		return ""
	}
	labels := append([]string{}, link.Labels...)
	sort.SliceStable(labels, func(i, j int) bool { return len(labels[i]) > len(labels[j]) })
	return labels[0]
}

func updateThreadArtifactState(workDir, label, action string) {
	if strings.TrimSpace(workDir) == "" || strings.TrimSpace(label) == "" {
		return
	}
	data, err := os.ReadFile(threadStatePath(workDir))
	if err != nil {
		return
	}
	var state savedThreadState
	if err := json.Unmarshal(data, &state); err != nil {
		return
	}
	state.CurrentTurn.ArtifactsUpdated = dedupeStrings(append(state.CurrentTurn.ArtifactsUpdated, label))
	if state.CurrentFocus == nil || strings.TrimSpace(state.CurrentFocus.Kind) == "" {
		state.CurrentFocus = &threadFocus{Kind: "artifact", ArtifactLabel: label}
	} else {
		state.CurrentFocus.Kind = "artifact"
		state.CurrentFocus.ArtifactLabel = label
	}
	if strings.TrimSpace(action) != "" {
		state.CurrentTurn.JustFinished = dedupeStrings(append(state.CurrentTurn.JustFinished, action))
	}
	_ = saveThreadState(workDir, state)
}
