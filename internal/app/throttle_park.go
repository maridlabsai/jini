package app

// Throttle park record — specs/auto-ask-execution-mode-design.md.
// CALLER-owned: pre-written before runWithThrottleSurvival on Ask paths,
// updated with reason/fallback on the error return, deleted only on SUCCESS
// (never on approval grant), so Ctrl-C at any stage leaves work resumable.
// One slot per state dir: concurrent processes are last-writer-wins on write,
// and one process's delete-on-success may remove another's park — accepted
// loss modes for a single-user CLI. Atomic rename prevents tearing.

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type throttlePark struct {
	SchemaVersion string `json:"schema_version"`
	ContextType   string `json:"context_type"`
	Prompt        string `json:"prompt"`
	RouteLabel    string `json:"route_label"`
	Reason        string `json:"reason"`
	FallbackHint  string `json:"fallback_hint"`
	ParkedAt      string `json:"parked_at"`
}

func throttleParkPath() string {
	return filepath.Join(sessionStateRoot(), "throttle-park.json")
}

func writeThrottlePark(prompt, routeLabel string) error {
	return persistThrottlePark(throttlePark{
		SchemaVersion: "0.1.0",
		ContextType:   "JiniThrottlePark",
		Prompt:        prompt,
		RouteLabel:    routeLabel,
		ParkedAt:      time.Now().Format(time.RFC3339),
	})
}

func updateThrottleParkError(reason, fallbackHint string) {
	park := loadThrottlePark()
	if park == nil {
		return
	}
	park.Reason = reason
	park.FallbackHint = fallbackHint
	_ = persistThrottlePark(*park)
}

func persistThrottlePark(park throttlePark) error {
	if err := os.MkdirAll(sessionStateRoot(), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(park, "", "  ")
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(sessionStateRoot(), "throttle-park-*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	if _, err := tmp.Write(append(data, '\n')); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return err
	}
	if err := tmp.Chmod(0o600); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return err
	}
	return os.Rename(tmpName, throttleParkPath())
}

func loadThrottlePark() *throttlePark {
	data, err := os.ReadFile(throttleParkPath())
	if err != nil {
		return nil
	}
	var park throttlePark
	if err := json.Unmarshal(data, &park); err != nil {
		return nil
	}
	if park.ContextType != "JiniThrottlePark" {
		return nil
	}
	return &park
}

func clearThrottlePark() {
	_ = os.Remove(throttleParkPath())
}

// throttleParkResumeLine describes what is about to be resumed. Stale parks
// disclose their age so a resume days later is never a silent surprise.
func throttleParkResumeLine(park *throttlePark) string {
	excerpt := park.Prompt
	if len(excerpt) > 60 {
		excerpt = excerpt[:57] + "..."
	}
	parkedAt, err := time.Parse(time.RFC3339, park.ParkedAt)
	if err == nil {
		if age := time.Since(parkedAt); age > time.Hour {
			days := int(age.Hours() / 24)
			if days >= 1 {
				return fmt.Sprintf("parked %d days ago: %s — resuming", days, excerpt)
			}
			return fmt.Sprintf("parked %d hours ago: %s — resuming", int(age.Hours()), excerpt)
		}
	}
	return fmt.Sprintf("Resuming: %s", excerpt)
}
