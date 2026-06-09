package app

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

type repoReviewSnapshot struct {
	Workspace      string
	InGitRepo      bool
	ChangedFiles   []string
	UntrackedFiles []string
	Branch         string
}

func isRepoReviewDirectTask(source string) bool {
	normalized := normalizeName(source)
	return strings.Contains(normalized, "review") && containsAny(normalized, []string{"repo", "repository", "branch"})
}

func renderRepoReviewDirectTaskStarted(w io.Writer, source string, snapshot repoReviewSnapshot) {
	fmt.Fprintf(w, "Working on: %s\n", strings.TrimSpace(source))
	fmt.Fprintln(w, "Repo review snapshot")
	fmt.Fprintf(w, "- Workspace: %s\n", firstNonEmpty(snapshot.Workspace, "current directory"))
	fmt.Fprintf(w, "- Changed files: %d\n", len(snapshot.ChangedFiles))
	fmt.Fprintf(w, "- Untracked files: %d\n", len(snapshot.UntrackedFiles))
	fmt.Fprintln(w, "- Open the saved draft for review focus and next commands.")
	fmt.Fprintln(w, "Saved. Use `jini status` for full status or `jini open` for the draft.")
}

func renderRepoReviewDirect(w io.Writer, snapshot repoReviewSnapshot) {
	fmt.Fprintln(w, "Repo review")
	fmt.Fprintf(w, "- Workspace: %s\n", firstNonEmpty(snapshot.Workspace, "current directory"))
	if !snapshot.InGitRepo {
		fmt.Fprintln(w, "This folder is not a git repo.")
		return
	}
	if strings.TrimSpace(snapshot.Branch) != "" {
		fmt.Fprintf(w, "- Branch: %s\n", snapshot.Branch)
	}
	fmt.Fprintf(w, "- Changed files: %d\n", len(snapshot.ChangedFiles))
	fmt.Fprintf(w, "- Untracked files: %d\n", len(snapshot.UntrackedFiles))
	if len(snapshot.ChangedFiles) > 0 {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Changed:")
		for _, path := range firstStrings(snapshot.ChangedFiles, 8) {
			fmt.Fprintf(w, "- %s\n", path)
		}
	}
	if len(snapshot.UntrackedFiles) > 0 {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Untracked:")
		for _, path := range firstStrings(snapshot.UntrackedFiles, 8) {
			fmt.Fprintf(w, "- %s\n", path)
		}
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Next:")
	fmt.Fprintln(w, "- git status --short")
	fmt.Fprintln(w, "- git diff --stat")
}

func applyRepoReviewSnapshot(summary *workSummary, source string) (*workSummary, repoReviewSnapshot, error) {
	snapshot := buildRepoReviewSnapshot()
	if summary == nil {
		return nil, snapshot, fmt.Errorf("missing work summary")
	}
	if err := writeRepoReviewSnapshot(summary.Dir, source, snapshot); err != nil {
		return nil, snapshot, err
	}
	_ = os.Remove(filepath.Join(summary.Dir, "views", "first-useful-pass.md"))
	current := &currentWork{
		PackDir:    summary.Dir,
		PackID:     summary.PackID,
		WorkUnitID: summary.WorkUnitID,
		Title:      summary.Title,
		State:      summary.State,
		Health:     inferHealthFromState(summary.State),
	}
	updated, err := loadWorkSummary(summary.Dir, current)
	if err != nil {
		return nil, snapshot, err
	}
	if err := saveThreadState(summary.Dir, synthesizeThreadState(updated)); err != nil {
		return nil, snapshot, err
	}
	return updated, snapshot, nil
}

func buildRepoReviewSnapshot() repoReviewSnapshot {
	cwd, err := os.Getwd()
	if err != nil {
		cwd = "."
	}
	snapshot := repoReviewSnapshot{
		Workspace: filepath.Base(cwd),
	}
	if strings.TrimSpace(snapshot.Workspace) == "" || snapshot.Workspace == "." {
		snapshot.Workspace = cwd
	}
	if out, ok := runGitOutput("rev-parse", "--is-inside-work-tree"); ok && strings.TrimSpace(out) == "true" {
		snapshot.InGitRepo = true
	}
	if !snapshot.InGitRepo {
		return snapshot
	}
	if branch, ok := runGitOutput("branch", "--show-current"); ok {
		snapshot.Branch = strings.TrimSpace(branch)
	}
	status, ok := runGitOutput("status", "--porcelain")
	if !ok {
		return snapshot
	}
	for _, raw := range strings.Split(status, "\n") {
		if strings.TrimSpace(raw) == "" {
			continue
		}
		if len(raw) < 3 {
			continue
		}
		state := raw[:2]
		path := strings.TrimSpace(raw[3:])
		if path == "" {
			continue
		}
		if state == "??" {
			snapshot.UntrackedFiles = append(snapshot.UntrackedFiles, path)
			continue
		}
		snapshot.ChangedFiles = append(snapshot.ChangedFiles, path)
	}
	sort.Strings(snapshot.ChangedFiles)
	sort.Strings(snapshot.UntrackedFiles)
	return snapshot
}

func runGitOutput(args ...string) (string, bool) {
	cmd := exec.Command("git", args...)
	out, err := cmd.Output()
	if err != nil {
		return "", false
	}
	return string(out), true
}

func writeRepoReviewSnapshot(workDir, source string, snapshot repoReviewSnapshot) error {
	sections := []starterArtifactSection{
		{
			Heading: "Review focus",
			Bullets: []string{
				"Start with changed files before broad architecture commentary.",
				"Separate correctness, test coverage, security, and release-risk notes.",
				"Treat this as a model-free first pass; escalate to a provider only when deeper reasoning is worth the cost.",
			},
		},
		{
			Heading: "Working tree",
			Bullets: repoReviewWorkingTreeBullets(snapshot),
		},
		{
			Heading: "Suggested next commands",
			Bullets: []string{
				"git status --short",
				"git diff --stat",
				"jini status",
			},
		},
		{
			Heading: "Original request",
			Bullets: []string{source},
		},
	}
	doc := starterArtifactDoc{
		Path:     "repo-review.md",
		Title:    "Repo Review Snapshot",
		Sections: sections,
	}
	return os.WriteFile(filepath.Join(workDir, "views", doc.Path), []byte(renderStarterArtifactDoc(doc)), 0o644)
}

func repoReviewWorkingTreeBullets(snapshot repoReviewSnapshot) []string {
	if !snapshot.InGitRepo {
		return []string{
			"Git repository not detected from the current directory.",
			"Run this command from a repository root for changed-file and untracked-file counts.",
		}
	}
	bullets := []string{
		fmt.Sprintf("Workspace: %s", firstNonEmpty(snapshot.Workspace, "current directory")),
		fmt.Sprintf("Branch: %s", firstNonEmpty(snapshot.Branch, "detached or unnamed")),
		fmt.Sprintf("Changed files: %d", len(snapshot.ChangedFiles)),
		fmt.Sprintf("Untracked files: %d", len(snapshot.UntrackedFiles)),
	}
	for _, path := range firstStrings(snapshot.ChangedFiles, 5) {
		bullets = append(bullets, "Changed: "+path)
	}
	for _, path := range firstStrings(snapshot.UntrackedFiles, 5) {
		bullets = append(bullets, "Untracked: "+path)
	}
	return bullets
}

func firstStrings(values []string, limit int) []string {
	if limit <= 0 || len(values) == 0 {
		return nil
	}
	if len(values) < limit {
		limit = len(values)
	}
	return append([]string{}, values[:limit]...)
}
