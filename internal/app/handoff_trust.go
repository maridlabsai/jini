package app

// Directory trust for autonomous hand-off — specs/handoff-posture-design.md.
// GLOBAL (~/.jini/trusted-dirs.json, same home as mode.json). Explicit opt-in:
// nothing is trusted until the user runs `jini trust` and confirms. Fail-safe:
// an unreadable/corrupt store reads as empty → plan-only. A grant is per exact
// directory (symlink-resolved), carries the acknowledged consent, and is
// revocable.

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	trustLevelSemi       = "semi"
	trustLevelAutonomous = "autonomous"
	trustStoreContext    = "JiniTrustedDirs"
	trustStoreSchema     = "0.1.0"
)

type trustGrant struct {
	Level        string `json:"level"`
	GrantedAt    string `json:"granted_at"`
	Acknowledged string `json:"acknowledged"`
}

type trustStore struct {
	SchemaVersion string                `json:"schema_version"`
	ContextType   string                `json:"context_type"`
	Dirs          map[string]trustGrant `json:"dirs"`
}

func trustStorePath() (string, error) {
	home, err := executionModeHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".jini", "trusted-dirs.json"), nil
}

// resolveTrustDir makes a directory absolute and symlink-resolved so grants and
// lookups compare on the same canonical path (macOS /var → /private/var).
func resolveTrustDir(dir string) string {
	d := strings.TrimSpace(dir)
	if d == "" {
		return ""
	}
	if abs, err := filepath.Abs(d); err == nil {
		d = abs
	}
	if resolved, err := filepath.EvalSymlinks(d); err == nil {
		d = resolved
	}
	return d
}

func loadTrustStore() *trustStore {
	path, err := trustStorePath()
	if err != nil {
		return nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var store trustStore
	if err := json.Unmarshal(data, &store); err != nil {
		return nil
	}
	if store.ContextType != trustStoreContext {
		return nil
	}
	return &store
}

// trustedLevelForDir returns the granted level for a directory, if any.
func trustedLevelForDir(dir string) (string, bool) {
	store := loadTrustStore()
	if store == nil {
		return "", false
	}
	grant, ok := store.Dirs[resolveTrustDir(dir)]
	if !ok {
		return "", false
	}
	return grant.Level, true
}

func persistTrustStore(store trustStore) error {
	path, err := trustStorePath()
	if err != nil {
		return err
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(store, "", "  ")
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, "trusted-dirs-*.tmp")
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
	return os.Rename(tmpName, path)
}

func saveTrustGrant(dir, level, acknowledged string) error {
	store := loadTrustStore()
	if store == nil {
		store = &trustStore{SchemaVersion: trustStoreSchema, ContextType: trustStoreContext}
	}
	if store.Dirs == nil {
		store.Dirs = map[string]trustGrant{}
	}
	store.Dirs[resolveTrustDir(dir)] = trustGrant{
		Level:        level,
		GrantedAt:    time.Now().UTC().Format(time.RFC3339),
		Acknowledged: acknowledged,
	}
	return persistTrustStore(*store)
}

func removeTrustGrant(dir string) (bool, error) {
	store := loadTrustStore()
	if store == nil || store.Dirs == nil {
		return false, nil
	}
	key := resolveTrustDir(dir)
	if _, ok := store.Dirs[key]; !ok {
		return false, nil
	}
	delete(store.Dirs, key)
	return true, persistTrustStore(*store)
}

// trustRepercussions is the neutral, factual description shown before consent.
// Route-agnostic: a grant is per-directory and Auto may select different
// hand-off CLIs there. No caps emphasis, no threat lists.
func trustRepercussions(level string) string {
	// One sentence per line so an important consent screen stays scannable
	// (no multi-sentence run-on wall); wording is unchanged.
	if level == trustLevelAutonomous {
		return "In this directory, Auto mode lets the coding CLI it hands off to " +
			"(such as Claude Code or Codex) apply file edits and run commands without " +
			"asking first.\n" +
			"This applies only here and only in Auto mode; Ask mode and " +
			"other directories are unchanged.\n" +
			"Use it where you're comfortable with unattended edits and commands."
	}
	return "In this directory, Auto mode lets the coding CLI it hands off to " +
		"(such as Claude Code or Codex) apply file edits without asking first.\n" +
		"It won't run commands.\n" +
		"This applies only here and only in Auto mode; Ask mode " +
		"and other directories are unchanged."
}

func trustAcknowledgement(level string) string {
	if level == trustLevelAutonomous {
		return "edits+commands"
	}
	return "edits"
}

func runTrust(args []string, stdout, stderr io.Writer) int {
	if len(args) > 0 {
		switch exactCommandToken(args[0]) {
		case "list":
			return runTrustList(stdout)
		case "remove", "untrust":
			target := ""
			if len(args) > 1 {
				target = args[1]
			}
			return runTrustRemove(target, stdout, stderr)
		}
	}

	level := trustLevelSemi
	for _, arg := range args {
		switch exactCommandToken(arg) {
		case "--autonomous", "autonomous":
			level = trustLevelAutonomous
		case "--semi", "semi":
			level = trustLevelSemi
		default:
			fmt.Fprintf(stderr, "Unknown argument %q. Use `jini trust`, `jini trust --autonomous`, `jini trust list`, or `jini trust remove`.\n", arg)
			return 1
		}
	}

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(stderr, "Could not determine the current directory.")
		return 1
	}

	fmt.Fprintln(stdout, trustRepercussions(level))
	if throttlePromptIsTTY == nil || !throttlePromptIsTTY() {
		fmt.Fprintln(stderr, "Trust needs an interactive terminal; nothing changed.")
		return 1
	}
	fmt.Fprintf(stdout, "Trust this directory for %s hand-off? [y/N] ", level)
	reader := bufio.NewReader(throttlePromptInput)
	line, _ := reader.ReadString('\n')
	switch strings.ToLower(strings.TrimSpace(line)) {
	case "y", "yes":
		if err := saveTrustGrant(cwd, level, trustAcknowledgement(level)); err != nil {
			fmt.Fprintln(stderr, err.Error())
			return 1
		}
		fmt.Fprintf(stdout, "Trusted %s for %s hand-off. Remove with `jini trust remove`.\n", resolveTrustDir(cwd), level)
		return 0
	default:
		fmt.Fprintln(stdout, "Nothing changed.")
		return 0
	}
}

func runTrustList(stdout io.Writer) int {
	store := loadTrustStore()
	if store == nil || len(store.Dirs) == 0 {
		fmt.Fprintln(stdout, "No trusted directories.")
		return 0
	}
	fmt.Fprintln(stdout, "Trusted directories:")
	for dir, grant := range store.Dirs {
		fmt.Fprintf(stdout, "- %s (%s, granted %s)\n", dir, grant.Level, grant.GrantedAt)
	}
	return 0
}

func runTrustRemove(target string, stdout, stderr io.Writer) int {
	dir := target
	if strings.TrimSpace(dir) == "" {
		cwd, err := os.Getwd()
		if err != nil {
			fmt.Fprintln(stderr, "Could not determine the current directory.")
			return 1
		}
		dir = cwd
	}
	removed, err := removeTrustGrant(dir)
	if err != nil {
		fmt.Fprintln(stderr, err.Error())
		return 1
	}
	if !removed {
		fmt.Fprintf(stdout, "%s was not trusted; nothing changed.\n", resolveTrustDir(dir))
		return 0
	}
	fmt.Fprintf(stdout, "Removed trust for %s.\n", resolveTrustDir(dir))
	return 0
}
