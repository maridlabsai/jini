package app_test

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type claudeCodexPromptBankEntry struct {
	ID                   string `json:"id"`
	SourceSignal         string `json:"source_signal"`
	Domain               string `json:"domain"`
	Prompt               string `json:"prompt"`
	AgeGroup             string `json:"age_group"`
	GenderContext        string `json:"gender_context"`
	RaceEthnicityContext string `json:"race_ethnicity_context"`
	UserContext          string `json:"user_context"`
	ExpectedBehavior     string `json:"expected_behavior"`
	ClaudeBaseline       string `json:"claude_baseline"`
	CodexBaseline        string `json:"codex_baseline"`
	SafetyRequirement    string `json:"safety_requirement"`
}

func TestClaudeCodexPromptBankCoversDiverseFirstMinuteUseCases(t *testing.T) {
	root := repoRootForMigrationTest(t)
	entries := readClaudeCodexPromptBank(t, filepath.Join(root, "specs", "claude-codex-prompt-bank.jsonl"))

	if len(entries) != 100 {
		t.Fatalf("Claude/Codex prompt bank must contain exactly 100 prompts, got %d", len(entries))
	}

	ids := map[string]bool{}
	domains := map[string]bool{}
	ageGroups := map[string]bool{}
	genderContexts := map[string]bool{}
	raceEthnicityContexts := map[string]bool{}
	behaviors := map[string]bool{}
	sourceSignals := map[string]int{}

	for index, entry := range entries {
		line := index + 1
		requirePromptBankField(t, line, "id", entry.ID)
		requirePromptBankField(t, line, "source_signal", entry.SourceSignal)
		requirePromptBankField(t, line, "domain", entry.Domain)
		requirePromptBankField(t, line, "prompt", entry.Prompt)
		requirePromptBankField(t, line, "age_group", entry.AgeGroup)
		requirePromptBankField(t, line, "gender_context", entry.GenderContext)
		requirePromptBankField(t, line, "race_ethnicity_context", entry.RaceEthnicityContext)
		requirePromptBankField(t, line, "user_context", entry.UserContext)
		requirePromptBankField(t, line, "expected_behavior", entry.ExpectedBehavior)
		requirePromptBankField(t, line, "claude_baseline", entry.ClaudeBaseline)
		requirePromptBankField(t, line, "codex_baseline", entry.CodexBaseline)
		requirePromptBankField(t, line, "safety_requirement", entry.SafetyRequirement)

		if ids[entry.ID] {
			t.Fatalf("prompt bank id %q is duplicated", entry.ID)
		}
		ids[entry.ID] = true
		if !strings.HasPrefix(entry.ID, "aryan-bank-") {
			t.Fatalf("prompt bank line %d id must preserve Aryan-derived bank prefix, got %q", line, entry.ID)
		}
		if entry.SourceSignal != "aryan-alpha-tester" {
			t.Fatalf("prompt bank line %d must cite Aryan alpha tester as the source signal, got %q", line, entry.SourceSignal)
		}
		if strings.Contains(strings.ToLower(entry.Prompt), "as a "+strings.ToLower(entry.RaceEthnicityContext)) {
			t.Fatalf("prompt bank line %d should not make identity a stereotype-driving instruction: %q", line, entry.Prompt)
		}
		for _, forbidden := range []string{"Task Snapshot", "Working Draft", "Result ready", "Saved artifact", "Next commands", "Start/Keep", "Switch"} {
			if strings.Contains(entry.ClaudeBaseline, forbidden) || strings.Contains(entry.CodexBaseline, forbidden) {
				t.Fatalf("prompt bank line %d baseline must avoid legacy scaffold phrase %q", line, forbidden)
			}
		}
		if len(entry.ClaudeBaseline) > 220 || len(entry.CodexBaseline) > 220 {
			t.Fatalf("prompt bank line %d baselines must stay compact", line)
		}

		domains[entry.Domain] = true
		ageGroups[entry.AgeGroup] = true
		genderContexts[entry.GenderContext] = true
		raceEthnicityContexts[entry.RaceEthnicityContext] = true
		behaviors[entry.ExpectedBehavior] = true
		sourceSignals[entry.SourceSignal]++
	}

	if len(domains) < 20 {
		t.Fatalf("prompt bank must cover at least 20 domains, got %d", len(domains))
	}
	for _, want := range []string{"teen", "young-adult", "adult", "older-adult", "not-specified"} {
		if !ageGroups[want] {
			t.Fatalf("prompt bank missing age group %q", want)
		}
	}
	for _, want := range []string{"woman", "man", "nonbinary", "not-specified"} {
		if !genderContexts[want] {
			t.Fatalf("prompt bank missing gender context %q", want)
		}
	}
	for _, want := range []string{"black", "latine", "south-asian", "east-asian", "middle-eastern", "indigenous", "white", "multiracial", "not-specified"} {
		if !raceEthnicityContexts[want] {
			t.Fatalf("prompt bank missing race or ethnicity context %q", want)
		}
	}
	for _, want := range []string{"direct_answer", "direct_file_edit", "ask_clarifying_question", "route_to_configured_cli", "setup_guidance", "safety_boundary", "create_artifact", "summarize_or_rewrite", "code_review"} {
		if !behaviors[want] {
			t.Fatalf("prompt bank missing expected behavior %q", want)
		}
	}
	if sourceSignals["aryan-alpha-tester"] != 100 {
		t.Fatalf("prompt bank must keep Aryan alpha tester signal on all entries, got %d", sourceSignals["aryan-alpha-tester"])
	}
}

func readClaudeCodexPromptBank(t *testing.T, path string) []claudeCodexPromptBankEntry {
	t.Helper()
	file, err := os.Open(path)
	if err != nil {
		t.Fatalf("open Claude/Codex prompt bank: %v", err)
	}
	defer file.Close()

	var entries []claudeCodexPromptBankEntry
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		var entry claudeCodexPromptBankEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			t.Fatalf("parse Claude/Codex prompt bank line %d: %v", len(entries)+1, err)
		}
		entries = append(entries, entry)
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("scan Claude/Codex prompt bank: %v", err)
	}
	return entries
}

func requirePromptBankField(t *testing.T, line int, field string, value string) {
	t.Helper()
	if strings.TrimSpace(value) == "" {
		t.Fatalf("prompt bank line %d missing %s", line, field)
	}
	if strings.Contains(value, "\n") || strings.Contains(value, "\r") {
		t.Fatalf("prompt bank line %d field %s must stay single-line", line, field)
	}
	if strings.Contains(value, "\t") {
		t.Fatalf("prompt bank line %d field %s must not contain tabs", line, field)
	}
	if field == "prompt" && len(value) < 6 {
		t.Fatalf("prompt bank line %d prompt is too short: %s", line, value)
	}
	if field == "expected_behavior" && strings.Contains(value, " ") {
		t.Fatalf("prompt bank line %d expected_behavior must be a stable token, got %s", line, value)
	}
	if field == "id" && strings.Contains(value, " ") {
		t.Fatalf("prompt bank line %d id must be a stable token, got %s", line, value)
	}
	if strings.TrimSpace(value) != value {
		t.Fatalf("prompt bank line %d field %s has leading or trailing whitespace: %s", line, field, fmt.Sprintf("%q", value))
	}
}
