package app

// R1A safe-target resolution (contract §5). The mutation-security boundary is
// os.OpenRoot (descriptor-relative, refuses symlink escape) PLUS explicit checks
// os.Root does not provide: rejecting an in-root symlink in ANY path component,
// hard links, non-regular/special files, and filesystem/mount and nested-repo
// boundaries. confinePath (agent_tools.go) is lexical/TOCTOU-prone and is NOT
// reused here. When a required guarantee cannot be proven, we fail closed.

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// r1aResolvedTarget is a target proven safe for one supervised single-file write.
type r1aResolvedTarget struct {
	root       *os.Root    // descriptor-relative handle confined to the workspace
	rel        string      // cleaned, OS-native relative path within root
	mode       os.FileMode // regular-file PERMISSION BITS only; owner/times/ACLs/xattrs are not preserved
	beforeHash string
	size       int64
}

// r1aCanonicalRoot absolutizes and canonicalizes a workspace root (resolving
// symlinks) and requires it to be an existing directory. It gives a stable
// identity for workspace and store-boundary decisions. It does NOT by itself
// eliminate filesystem races — those are handled by per-operation revalidation
// and by failing closed — and must not be described as doing so.
func r1aCanonicalRoot(root string) (string, error) {
	if strings.TrimSpace(root) == "" {
		return "", fmt.Errorf("r1a: workspace root must not be empty")
	}
	abs, err := filepath.Abs(root)
	if err != nil {
		return "", fmt.Errorf("r1a: cannot absolutize workspace root: %w", err)
	}
	real, err := filepath.EvalSymlinks(abs)
	if err != nil {
		return "", fmt.Errorf("r1a: cannot canonicalize workspace root: %w", err)
	}
	info, err := os.Stat(real)
	if err != nil {
		return "", fmt.Errorf("r1a: cannot stat workspace root: %w", err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("r1a: workspace root is not a directory")
	}
	return real, nil
}

// r1aOpenWorkspace opens a descriptor-relative root confined to canonicalRoot.
func r1aOpenWorkspace(canonicalRoot string) (*os.Root, error) {
	if strings.TrimSpace(canonicalRoot) == "" || !filepath.IsAbs(canonicalRoot) {
		return nil, fmt.Errorf("r1a: workspace root must be an absolute path")
	}
	root, err := os.OpenRoot(canonicalRoot)
	if err != nil {
		return nil, fmt.Errorf("r1a: cannot open workspace root: %w", err)
	}
	return root, nil
}

// r1aResolveTarget validates rel as an existing regular file safe to replace.
func r1aResolveTarget(root *os.Root, rel string) (*r1aResolvedTarget, error) {
	clean := filepath.Clean(rel)
	slash := filepath.ToSlash(clean)
	if clean == "." || clean == "" || slash == ".." || strings.HasPrefix(slash, "../") || filepath.IsAbs(clean) {
		return nil, fmt.Errorf("r1a: invalid or root-equivalent target %q", rel)
	}

	// Reject a symlink in EVERY component. os.Root refuses symlinks that escape
	// the root but still FOLLOWS in-root symlinks; the contract requires
	// rejecting any symlink component for a mutation target.
	parts := strings.Split(slash, "/")
	cur := ""
	for i, p := range parts {
		if cur == "" {
			cur = p
		} else {
			cur = cur + string(filepath.Separator) + p
		}
		info, err := root.Lstat(cur)
		if err != nil {
			return nil, fmt.Errorf("r1a: target path must fully exist: cannot lstat %q: %w", cur, err)
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return nil, fmt.Errorf("r1a: symlink in path component %q rejected", cur)
		}
		if i < len(parts)-1 && !info.IsDir() {
			return nil, fmt.Errorf("r1a: non-directory in path prefix %q", cur)
		}
	}

	info, err := root.Lstat(clean)
	if err != nil {
		return nil, fmt.Errorf("r1a: target must be an existing file: %w", err)
	}
	if info.IsDir() {
		return nil, fmt.Errorf("r1a: target is a directory, not a file")
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("r1a: target is not a regular file (special files are rejected)")
	}
	if info.Size() > r1aMaxFileBytes {
		return nil, fmt.Errorf("r1a: target exceeds the single-file size bound")
	}

	// Hard-link + filesystem/mount-boundary checks (platform helper; fail-closed
	// where the platform cannot prove them).
	if err := r1aCheckBoundaries(root, info); err != nil {
		return nil, err
	}

	// Nested repo / submodule / worktree boundary.
	if err := r1aCheckNestedRepo(root, clean); err != nil {
		return nil, err
	}

	data, err := r1aReadTarget(root, clean)
	if err != nil {
		return nil, err
	}
	return &r1aResolvedTarget{
		root:       root,
		rel:        clean,
		mode:       info.Mode().Perm(),
		beforeHash: r1aHashBytes(data),
		size:       info.Size(),
	}, nil
}

// r1aReadTarget reads the target through the confined root, bounded in size.
func r1aReadTarget(root *os.Root, rel string) ([]byte, error) {
	f, err := root.OpenFile(rel, os.O_RDONLY, 0)
	if err != nil {
		return nil, fmt.Errorf("r1a: cannot read target: %w", err)
	}
	defer f.Close()
	data, err := io.ReadAll(io.LimitReader(f, r1aMaxFileBytes+1))
	if err != nil {
		return nil, fmt.Errorf("r1a: cannot read target: %w", err)
	}
	if int64(len(data)) > r1aMaxFileBytes {
		return nil, fmt.Errorf("r1a: target exceeds the single-file size bound")
	}
	return data, nil
}

// r1aCheckNestedRepo rejects a target that sits inside a nested repository,
// submodule, or linked worktree (a .git entry in any ancestor strictly below
// the workspace root). The root's own .git is never inspected.
func r1aCheckNestedRepo(root *os.Root, rel string) error {
	dir := filepath.Dir(rel)
	for dir != "." && dir != "" && dir != string(filepath.Separator) {
		if _, err := root.Lstat(filepath.Join(dir, ".git")); err == nil {
			return fmt.Errorf("r1a: target lies inside a nested repo/submodule/worktree (.git under %q)", dir)
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return nil
}
