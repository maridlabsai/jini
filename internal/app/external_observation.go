package app

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"
)

type externalObservationRegistry struct {
	SchemaVersion string                    `json:"schema_version"`
	ContextType   string                    `json:"context_type"`
	Items         []externalObservationItem `json:"items"`
}

type externalObservationItem struct {
	ObservationID    string `json:"observation_id"`
	ArtifactPath     string `json:"artifact_path"`
	TargetPath       string `json:"target_path"`
	TargetLabel      string `json:"target_label,omitempty"`
	ConnectorID      string `json:"connector_id,omitempty"`
	ConnectorLabel   string `json:"connector_label,omitempty"`
	AddedAt          string `json:"added_at"`
	LastObservedAt   string `json:"last_observed_at,omitempty"`
	LastFingerprint  string `json:"last_fingerprint,omitempty"`
	SharedRecorded   bool   `json:"shared_recorded,omitempty"`
	UsedRecorded     bool   `json:"used_recorded,omitempty"`
	ReplacedRecorded bool   `json:"replaced_recorded,omitempty"`
}

func externalObservationPath(workDir string) string {
	return filepath.Join(workDir, "external-observation.json")
}

func loadExternalObservations(workDir string) externalObservationRegistry {
	data, err := os.ReadFile(externalObservationPath(workDir))
	if err != nil {
		return externalObservationRegistry{Items: []externalObservationItem{}}
	}
	var payload externalObservationRegistry
	if err := json.Unmarshal(data, &payload); err != nil {
		return externalObservationRegistry{Items: []externalObservationItem{}}
	}
	if payload.Items == nil {
		payload.Items = []externalObservationItem{}
	}
	return payload
}

func saveExternalObservations(workDir string, registry externalObservationRegistry) error {
	if registry.SchemaVersion == "" {
		registry.SchemaVersion = "0.1.0"
	}
	if registry.ContextType == "" {
		registry.ContextType = "JiniExternalObservationRegistry"
	}
	if registry.Items == nil {
		registry.Items = []externalObservationItem{}
	}
	data, err := json.MarshalIndent(registry, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(externalObservationPath(workDir), append(data, '\n'), 0o600)
}

func addExternalObservation(workDir, artifactPath, targetPath, connectorID string) (externalObservationItem, error) {
	if strings.TrimSpace(workDir) == "" {
		return externalObservationItem{}, fmt.Errorf("no active work to observe")
	}
	if strings.TrimSpace(artifactPath) == "" {
		return externalObservationItem{}, fmt.Errorf("no ready artifact is available to observe")
	}
	artifactAbs, err := filepath.Abs(artifactPath)
	if err != nil {
		return externalObservationItem{}, err
	}
	targetAbs, err := filepath.Abs(strings.TrimSpace(targetPath))
	if err != nil {
		return externalObservationItem{}, err
	}
	registry := loadExternalObservations(workDir)
	for _, item := range registry.Items {
		if item.TargetPath == targetAbs && item.ArtifactPath == artifactAbs {
			return item, nil
		}
	}
	item := externalObservationItem{
		ObservationID: slugify(filepath.Base(targetAbs) + "-" + filepath.Base(artifactAbs)),
		ArtifactPath:  artifactAbs,
		TargetPath:    targetAbs,
		TargetLabel:   filepath.Base(targetAbs),
		ConnectorID:   normalizeConnectorID(firstNonEmpty(connectorID, inferConnectorIDForPath(targetAbs))),
		ConnectorLabel: connectorLabel(normalizeConnectorID(firstNonEmpty(connectorID, inferConnectorIDForPath(targetAbs)))),
		AddedAt:       time.Now().UTC().Format(time.RFC3339),
	}
	registry.Items = append(registry.Items, item)
	if err := saveExternalObservations(workDir, registry); err != nil {
		return externalObservationItem{}, err
	}
	return item, nil
}

func scanExternalObservations(workDir string) error {
	registry := loadExternalObservations(workDir)
	if len(registry.Items) == 0 {
		return nil
	}
	for index := range registry.Items {
		item := &registry.Items[index]
		if err := scanExternalObservationItem(workDir, item); err != nil {
			return err
		}
	}
	return saveExternalObservations(workDir, registry)
}

func scanExternalObservationItem(workDir string, item *externalObservationItem) error {
	if item == nil || strings.TrimSpace(item.TargetPath) == "" || strings.TrimSpace(item.ArtifactPath) == "" {
		return nil
	}
	targetData, err := os.ReadFile(item.TargetPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	targetFingerprint := fileFingerprint(targetData)
	item.LastObservedAt = time.Now().UTC().Format(time.RFC3339)
	if !item.SharedRecorded {
		if err := savePassiveArtifactOutcome(workDir, "shared-this", item.ArtifactPath, connectorAwareOutcomeReason(item.ConnectorID, "external-target-present")); err != nil {
			return err
		}
		item.SharedRecorded = true
	}
	artifactData, err := os.ReadFile(item.ArtifactPath)
	if err == nil {
		if outcome, reason := classifyExternalObservationOutcome(string(artifactData), string(targetData)); outcome != "" {
			reason = connectorAwareOutcomeReason(item.ConnectorID, reason)
			switch outcome {
			case "replaced-this":
				if !item.ReplacedRecorded {
					if err := savePassiveArtifactOutcome(workDir, outcome, item.ArtifactPath, reason); err != nil {
						return err
					}
					item.ReplacedRecorded = true
				}
			case "used-this":
				if !item.UsedRecorded {
					if err := savePassiveArtifactOutcome(workDir, outcome, item.ArtifactPath, reason); err != nil {
						return err
					}
					item.UsedRecorded = true
				}
			}
		}
	}
	item.LastFingerprint = targetFingerprint
	return nil
}

func classifyExternalObservationOutcome(artifactContent, targetContent string) (string, string) {
	if artifactContent == targetContent {
		return "", ""
	}
	if !utf8.ValidString(artifactContent) || !utf8.ValidString(targetContent) {
		return "used-this", "external-target-changed"
	}
	editClass, editScope, semanticClass := classifyArtifactEditSignal(artifactContent, targetContent)
	if semanticClass == "core-decision-change" && editClass != "none" {
		return "replaced-this", "external-target-decision-change"
	}
	if editClass == "heavy" && editScope == "core-sections" {
		return "replaced-this", "external-target-diverged"
	}
	if editClass != "none" {
		return "used-this", "external-target-edited"
	}
	return "", ""
}

func fileFingerprint(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func renderExternalObservationStatus(w io.Writer, workDir string) {
	registry := loadExternalObservations(workDir)
	fmt.Fprintln(w, "Observed external work")
	fmt.Fprintln(w)
	if len(registry.Items) == 0 {
		fmt.Fprintln(w, "- Nothing observed yet")
		return
	}
	for _, item := range registry.Items {
		fmt.Fprintf(w, "- %s\n", firstNonEmpty(strings.TrimSpace(item.TargetLabel), filepath.Base(item.TargetPath)))
		if strings.TrimSpace(item.ConnectorLabel) != "" {
			fmt.Fprintf(w, "  Connector: %s\n", item.ConnectorLabel)
		}
		fmt.Fprintf(w, "  Target: %s\n", item.TargetPath)
		fmt.Fprintf(w, "  Source: %s\n", item.ArtifactPath)
		if strings.TrimSpace(item.LastObservedAt) != "" {
			fmt.Fprintf(w, "  Last seen: %s\n", item.LastObservedAt)
		}
		signals := []string{}
		if item.SharedRecorded {
			signals = append(signals, "shared")
		}
		if item.UsedRecorded {
			signals = append(signals, "used")
		}
		if item.ReplacedRecorded {
			signals = append(signals, "replaced")
		}
		if len(signals) > 0 {
			fmt.Fprintf(w, "  Signals: %s\n", strings.Join(signals, ", "))
		}
	}
}

func inferConnectorForCatalogItem(item catalogItem) string {
	joined := normalizeName(item.ID + " " + item.Label + " " + item.Path)
	switch {
	case containsAny(joined, []string{"github issues", "github"}):
		return "github"
	case containsAny(joined, []string{"jira issues", "jira"}):
		return "jira"
	case containsAny(joined, []string{"confluence"}):
		return "confluence"
	case containsAny(joined, []string{"markdown wiki", "markdown", "wiki"}):
		return "markdown"
	default:
		return ""
	}
}

func inferConnectorIDForPath(path string) string {
	normalized := normalizeName(path)
	switch {
	case containsAny(normalized, []string{"github"}):
		return "github"
	case containsAny(normalized, []string{"jira"}):
		return "jira"
	case containsAny(normalized, []string{"confluence"}):
		return "confluence"
	case containsAny(normalized, []string{"markdown"}):
		return "markdown"
	default:
		return ""
	}
}

func normalizeConnectorID(value string) string {
	switch normalizeName(value) {
	case "github":
		return "github"
	case "jira":
		return "jira"
	case "confluence":
		return "confluence"
	case "markdown", "markdown wiki", "wiki":
		return "markdown"
	default:
		return ""
	}
}

func connectorLabel(id string) string {
	switch normalizeConnectorID(id) {
	case "github":
		return "GitHub"
	case "jira":
		return "Jira"
	case "confluence":
		return "Confluence"
	case "markdown":
		return "Markdown"
	default:
		return ""
	}
}

func connectorAwareOutcomeReason(connectorID, reason string) string {
	connectorID = normalizeConnectorID(connectorID)
	if connectorID == "" || strings.TrimSpace(reason) == "" {
		return strings.TrimSpace(reason)
	}
	return "connector-" + connectorID + "-" + strings.TrimSpace(reason)
}
