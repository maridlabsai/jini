package app

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

type userContextMemoryFile struct {
	SchemaVersion string                    `json:"schema_version"`
	ContextType   string                    `json:"context_type"`
	Enabled       bool                      `json:"enabled"`
	Signals       []userContextMemorySignal `json:"signals,omitempty"`
	UpdatedAt     string                    `json:"updated_at,omitempty"`
}

type userContextMemorySignal struct {
	ID         string `json:"id"`
	Kind       string `json:"kind"`
	Label      string `json:"label"`
	Evidence   string `json:"evidence"`
	Count      int    `json:"count,omitempty"`
	LastSeenAt string `json:"last_seen_at,omitempty"`
}

type rawUserContextMemoryFile struct {
	SchemaVersion string                    `json:"schema_version"`
	ContextType   string                    `json:"context_type"`
	Enabled       *bool                     `json:"enabled"`
	Signals       []userContextMemorySignal `json:"signals,omitempty"`
	UpdatedAt     string                    `json:"updated_at,omitempty"`
}

var unsafeUserMemoryPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)(secret|token|api[_ -]?key|password|passwd|private[_ -]?key)`),
	regexp.MustCompile(`(?i)-----BEGIN [A-Z ]+PRIVATE KEY-----`),
	regexp.MustCompile(`[A-Za-z0-9_\-]{32,}`),
}

func userContextMemoryPath() string {
	return filepath.Join(sessionStateRoot(), "user-context.json")
}

func defaultUserContextMemory() userContextMemoryFile {
	return userContextMemoryFile{
		SchemaVersion: "0.1.0",
		ContextType:   "JiniUserContextMemory",
		Enabled:       true,
	}
}

func loadUserContextMemory() userContextMemoryFile {
	data, err := os.ReadFile(userContextMemoryPath())
	if err != nil {
		return defaultUserContextMemory()
	}
	var raw rawUserContextMemoryFile
	if err := json.Unmarshal(data, &raw); err != nil {
		return defaultUserContextMemory()
	}
	memory := defaultUserContextMemory()
	if strings.TrimSpace(raw.SchemaVersion) != "" {
		memory.SchemaVersion = strings.TrimSpace(raw.SchemaVersion)
	}
	if strings.TrimSpace(raw.ContextType) != "" {
		memory.ContextType = strings.TrimSpace(raw.ContextType)
	}
	if raw.Enabled != nil {
		memory.Enabled = *raw.Enabled
	}
	memory.UpdatedAt = strings.TrimSpace(raw.UpdatedAt)
	memory.Signals = sanitizeUserContextSignals(raw.Signals)
	return memory
}

func saveUserContextMemory(memory userContextMemoryFile) error {
	memory.SchemaVersion = "0.1.0"
	memory.ContextType = "JiniUserContextMemory"
	memory.Signals = sanitizeUserContextSignals(memory.Signals)
	memory.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	if err := os.MkdirAll(sessionStateRoot(), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(memory, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(userContextMemoryPath(), append(data, '\n'), 0o600)
}

func runMemory(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		renderUserContextMemorySummary(stdout, currentWorkTitleForMemory())
		return 0
	}
	switch exactCommandToken(args[0]) {
	case "inspect", "status":
		renderUserContextMemoryInspect(stdout)
		return 0
	case "off", "disable":
		memory := loadUserContextMemory()
		memory.Enabled = false
		if err := saveUserContextMemory(memory); err != nil {
			fmt.Fprintf(stderr, "Could not update memory: %v\n", err)
			return 1
		}
		fmt.Fprintln(stdout, "Memory learning off.")
		fmt.Fprintln(stdout, "Run `jini memory on` to resume.")
		return 0
	case "on", "enable":
		memory := loadUserContextMemory()
		memory.Enabled = true
		if err := saveUserContextMemory(memory); err != nil {
			fmt.Fprintf(stderr, "Could not update memory: %v\n", err)
			return 1
		}
		fmt.Fprintln(stdout, "Memory learning on.")
		fmt.Fprintln(stdout, "Run `jini memory inspect` to review saved signals.")
		return 0
	case "forget", "clear", "revoke":
		memory := loadUserContextMemory()
		memory.Signals = nil
		if err := saveUserContextMemory(memory); err != nil {
			fmt.Fprintf(stderr, "Could not clear memory: %v\n", err)
			return 1
		}
		fmt.Fprintln(stdout, "Memory forgotten.")
		fmt.Fprintf(stdout, "Learning: %s\n", onOff(memory.Enabled))
		return 0
	default:
		fmt.Fprintf(stderr, "Unknown memory command %q.\n", args[0])
		fmt.Fprintln(stderr, "Run `jini memory inspect`, `jini memory off`, or `jini memory forget`.")
		return 1
	}
}

func currentWorkTitleForMemory() string {
	current, err := loadCurrentWork()
	if err != nil || current == nil {
		return "none"
	}
	if summary, loadErr := loadWorkSummary(current.PackDir, current); loadErr == nil && strings.TrimSpace(summary.Title) != "" {
		return strings.TrimSpace(summary.Title)
	}
	if strings.TrimSpace(current.Title) != "" {
		return strings.TrimSpace(current.Title)
	}
	return "none"
}

func renderUserContextMemorySummary(w io.Writer, currentWorkTitle string) {
	memory := loadUserContextMemory()
	fmt.Fprintln(w, "Memory")
	fmt.Fprintf(w, "Learning: %s, %d signals.\n", onOff(memory.Enabled), len(memory.Signals))
	if strings.TrimSpace(currentWorkTitle) == "" {
		currentWorkTitle = "none"
	}
	fmt.Fprintf(w, "Current work: %s.\n", currentWorkTitle)
}

func renderUserContextMemoryInspect(w io.Writer) {
	memory := loadUserContextMemory()
	fmt.Fprintln(w, "Memory")
	fmt.Fprintf(w, "Learning: %s\n", onOff(memory.Enabled))
	fmt.Fprintf(w, "Signals: %d\n", len(memory.Signals))
	for _, signal := range memory.Signals {
		evidence := strings.TrimSpace(signal.Evidence)
		if evidence == "" {
			fmt.Fprintf(w, "- %s: %s\n", signal.Kind, signal.Label)
			continue
		}
		fmt.Fprintf(w, "- %s: %s (%s)\n", signal.Kind, signal.Label, evidence)
	}
	fmt.Fprintln(w, "Privacy: stores safe labels only; no raw prompts, file contents, or CLI output.")
	fmt.Fprintln(w, "Controls: `jini memory off`, `jini memory on`, `jini memory forget`.")
}

func recordUserContextSignal(kind, label, evidence string) {
	kind = normalizeMemorySignalKind(kind)
	label = safeUserMemoryText(label)
	evidence = safeUserMemoryText(evidence)
	if kind == "" || label == "" || evidence == "" {
		return
	}
	memory := loadUserContextMemory()
	if !memory.Enabled {
		return
	}
	now := time.Now().UTC().Format(time.RFC3339)
	id := normalizeMemorySignalID(kind, label)
	for i := range memory.Signals {
		if memory.Signals[i].ID != id {
			continue
		}
		memory.Signals[i].Count++
		memory.Signals[i].Evidence = evidence
		memory.Signals[i].LastSeenAt = now
		_ = saveUserContextMemory(memory)
		return
	}
	memory.Signals = append(memory.Signals, userContextMemorySignal{
		ID:         id,
		Kind:       kind,
		Label:      label,
		Evidence:   evidence,
		Count:      1,
		LastSeenAt: now,
	})
	_ = saveUserContextMemory(memory)
}

func sanitizeUserContextSignals(signals []userContextMemorySignal) []userContextMemorySignal {
	out := make([]userContextMemorySignal, 0, len(signals))
	seen := map[string]bool{}
	for _, signal := range signals {
		kind := normalizeMemorySignalKind(signal.Kind)
		label := safeUserMemoryText(signal.Label)
		evidence := safeUserMemoryText(signal.Evidence)
		if kind == "" || label == "" || evidence == "" {
			continue
		}
		id := normalizeMemorySignalID(kind, label)
		if seen[id] {
			continue
		}
		seen[id] = true
		count := signal.Count
		if count < 1 {
			count = 1
		}
		out = append(out, userContextMemorySignal{
			ID:         id,
			Kind:       kind,
			Label:      label,
			Evidence:   evidence,
			Count:      count,
			LastSeenAt: safeUserMemoryTimestamp(signal.LastSeenAt),
		})
	}
	return out
}

func normalizeMemorySignalKind(kind string) string {
	switch strings.ReplaceAll(normalizeName(kind), " ", "_") {
	case "route_preference", "task_shape", "friction", "accepted_action", "rejected_action":
		return strings.ReplaceAll(normalizeName(kind), " ", "_")
	default:
		return ""
	}
}

func normalizeMemorySignalID(kind, label string) string {
	return strings.ReplaceAll(normalizeName(kind+" "+label), " ", "-")
}

func safeUserMemoryText(value string) string {
	value = strings.Join(strings.Fields(strings.TrimSpace(value)), " ")
	if value == "" || len(value) > 96 {
		return ""
	}
	for _, pattern := range unsafeUserMemoryPatterns {
		if pattern.MatchString(value) {
			return ""
		}
	}
	return value
}

func safeUserMemoryTimestamp(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if _, err := time.Parse(time.RFC3339, value); err != nil {
		return ""
	}
	return value
}

func onOff(enabled bool) string {
	if enabled {
		return "on"
	}
	return "off"
}
