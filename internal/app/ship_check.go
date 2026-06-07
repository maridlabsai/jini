package app

import (
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type shipCheckReport struct {
	SchemaVersion    string   `json:"schema_version"`
	ResultType       string   `json:"result_type"`
	GeneratedAt      string   `json:"generated_at"`
	Status           string   `json:"status"`
	Workspace        string   `json:"workspace"`
	InGitRepo        bool     `json:"in_git_repo"`
	Branch           string   `json:"branch,omitempty"`
	Upstream         string   `json:"upstream,omitempty"`
	AheadCount       int      `json:"ahead_count"`
	BehindCount      int      `json:"behind_count"`
	DirtyFiles       int      `json:"dirty_files"`
	UntrackedFiles   int      `json:"untracked_files"`
	RequiredEvidence []string `json:"required_evidence"`
	Blockers         []string `json:"blockers,omitempty"`
	Warnings         []string `json:"warnings,omitempty"`
	Next             []string `json:"next"`
}

func runShipCheck(args []string, stdout, stderr io.Writer) int {
	format, ok := parseOptionalFormatArgs(args)
	if !ok {
		fmt.Fprintln(stderr, "Unsupported check ship format. Try `jini check ship` or `jini check ship --format json`.")
		return 1
	}
	report := buildShipCheckReport()
	if format == "json" {
		encoder := json.NewEncoder(stdout)
		encoder.SetIndent("", "  ")
		encoder.SetEscapeHTML(false)
		if err := encoder.Encode(report); err != nil {
			fmt.Fprintf(stderr, "Could not render ship check report: %v\n", err)
			return 1
		}
	} else {
		renderShipCheckText(stdout, report)
	}
	if report.Status == "ok" {
		return 0
	}
	return 1
}

func buildShipCheckReport() shipCheckReport {
	report := shipCheckReport{
		SchemaVersion: "0.1.0",
		ResultType:    "JiniShipCheck",
		GeneratedAt:   time.Now().UTC().Format(time.RFC3339),
		Status:        "ok",
		RequiredEvidence: []string{
			"bash tools/run_required_gates.sh push",
			"git worktree add",
			"write validation report before push",
		},
		Next: []string{
			"Create an isolated worktree for validation.",
			"Run the required push gates there.",
			"Push only after the validation report is clean.",
		},
	}

	if cwd, ok := runGitOutput("rev-parse", "--show-toplevel"); ok {
		report.Workspace = filepath.Base(strings.TrimSpace(cwd))
	} else {
		report.Workspace = "current directory"
	}
	if inRepo, ok := runGitOutput("rev-parse", "--is-inside-work-tree"); !ok || strings.TrimSpace(inRepo) != "true" {
		report.Status = "blocked"
		report.Blockers = append(report.Blockers, "not inside a git repository")
		return report
	}
	report.InGitRepo = true

	if branch, ok := runGitOutput("branch", "--show-current"); ok {
		report.Branch = strings.TrimSpace(branch)
	}
	if upstream, ok := runGitOutput("rev-parse", "--abbrev-ref", "--symbolic-full-name", "@{u}"); ok {
		report.Upstream = strings.TrimSpace(upstream)
		report.AheadCount, report.BehindCount = gitAheadBehind()
	} else {
		report.Warnings = append(report.Warnings, "no upstream branch configured")
	}

	report.DirtyFiles, report.UntrackedFiles = gitDirtyCounts()
	if report.DirtyFiles > 0 {
		report.Status = "blocked"
		report.Blockers = append(report.Blockers, "working tree has uncommitted changes")
	}
	return report
}

func gitAheadBehind() (int, int) {
	out, ok := runGitOutput("rev-list", "--left-right", "--count", "HEAD...@{u}")
	if !ok {
		return 0, 0
	}
	fields := strings.Fields(out)
	if len(fields) != 2 {
		return 0, 0
	}
	ahead, _ := strconv.Atoi(fields[0])
	behind, _ := strconv.Atoi(fields[1])
	return ahead, behind
}

func gitDirtyCounts() (int, int) {
	out, ok := runGitOutput("status", "--porcelain")
	if !ok {
		return 0, 0
	}
	dirty := 0
	untracked := 0
	for _, line := range strings.Split(out, "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		dirty++
		if strings.HasPrefix(line, "??") {
			untracked++
		}
	}
	return dirty, untracked
}

func renderShipCheckText(w io.Writer, report shipCheckReport) {
	fmt.Fprintf(w, "Ship check %s\n", report.Status)
	fmt.Fprintf(w, "Branch: %s\n", firstNonEmpty(report.Branch, "detached or unnamed"))
	fmt.Fprintf(w, "Workspace: %s\n", firstNonEmpty(report.Workspace, "current directory"))
	fmt.Fprintf(w, "Dirty files: %d\n", report.DirtyFiles)
	if report.AheadCount > 0 || report.BehindCount > 0 {
		fmt.Fprintf(w, "Upstream delta: ahead %d, behind %d\n", report.AheadCount, report.BehindCount)
	}
	for _, blocker := range report.Blockers {
		fmt.Fprintf(w, "Blocker: %s\n", blocker)
	}
	for _, warning := range report.Warnings {
		fmt.Fprintf(w, "Warning: %s\n", warning)
	}
	fmt.Fprintln(w, "Run before push: bash tools/run_required_gates.sh push")
	fmt.Fprintln(w, "Safe lane: create an isolated worktree, run gates, then push only after evidence is clean.")
}
