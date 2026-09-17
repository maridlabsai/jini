package app

import "testing"

// The repo-review classifier must keep genuine read-only review requests, but
// must NOT swallow a build task that merely mentions reviewing a repo/branch —
// that would silently no-op the requested change (the routing-rubric misroute).
func TestIsRepoReviewDirectTask(t *testing.T) {
	cases := []struct {
		name   string
		source string
		want   bool
	}{
		// Genuine read-only review requests stay repo-review.
		{"plain repo review", "review this repo", true},
		{"review uncommitted", "review this repo for uncommitted changes", true},
		{"review the branch", "review the branch changes", true},
		{"review repository security", "review the repository for security issues", true},

		// Build tasks that mention reviewing a repo/branch must fall through.
		{"review and fix", "review the repo and fix the failing build", false},
		{"review and implement", "review the repository and implement the missing handler", false},
		{"review and refactor", "review the branch and refactor the router", false},
		{"review compat and fix", "review CLI compatibility across the repo and fix any issues", false},
		{"review and rewrite", "review the repo and rewrite the parser", false},
		{"review and migrate", "review the repository and migrate the schema", false},
		{"review and add", "review the repo and add tests for the router", false},
		{"review and remove", "review the branch and remove the dead code", false},
		{"review and build", "review the repository and build the missing feature", false},

		// Non-review prompts are never repo-review regardless of keywords.
		{"no review keyword", "fix the repo build", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := isRepoReviewDirectTask(tc.source); got != tc.want {
				t.Fatalf("isRepoReviewDirectTask(%q) = %v, want %v", tc.source, got, tc.want)
			}
		})
	}
}
