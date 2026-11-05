package app

// R1A protected artifact store (contract §7). Operation evidence lives ONLY
// beneath <JINI_HOME>/.jini/r1a/<workspace-id>/<run-id>/ — outside the
// repository, never a shared OS temp dir. Directories are owner-only (0700) and
// files owner read/write (0600); if that protection cannot be established the
// store fails closed. Retention/resume policy is deferred to R2.

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// r1aSecretPatterns masks credential-shaped substrings in any Detail string.
// Deliberately narrow (anchored prefixes / key=value) to avoid redacting
// ordinary text. The real compromised key is never referenced anywhere.
var r1aSecretPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?:sk-[A-Za-z0-9]{10,}|gsk_[A-Za-z0-9]{10,}|ghp_[A-Za-z0-9]{10,}|AKIA[0-9A-Z]{12,}|xox[baprs]-[A-Za-z0-9-]{10,})`),
	regexp.MustCompile(`(?i)\b(?:password|secret|token|api[_-]?key)\b\s*[:=]\s*\S+`),
}

// r1aWorkspaceID is a deterministic, non-secret identifier for a workspace root.
func r1aWorkspaceID(canonicalRoot string) string {
	sum := sha256.Sum256([]byte(canonicalRoot))
	return hex.EncodeToString(sum[:])[:16]
}

// r1aStoreRoot returns the protected per-operation directory path.
func r1aStoreRoot(workspaceID, runID string) (string, error) {
	home, err := executionModeHomeDir()
	if err != nil {
		return "", fmt.Errorf("r1a: cannot resolve protected home: %w", err)
	}
	if strings.TrimSpace(home) == "" {
		return "", fmt.Errorf("r1a: protected home is empty (fail closed)")
	}
	if workspaceID == "" || runID == "" {
		return "", fmt.Errorf("r1a: workspace/run identifier is empty (fail closed)")
	}
	return filepath.Join(home, ".jini", "r1a", workspaceID, runID), nil
}

// r1aEnsureProtectedStore resolves the protected per-operation directory from
// the captured canonical workspace root, validates that it resolves OUTSIDE the
// workspace BEFORE creating any protected state, creates the owner-only tree,
// and re-validates the boundary against the fully canonicalized path. It fails
// closed if the boundary cannot be established.
func r1aEnsureProtectedStore(canonicalRoot, workspaceID, runID string) (string, error) {
	dir, err := r1aStoreRoot(workspaceID, runID)
	if err != nil {
		return "", err
	}
	within, err := r1aStoreWithinWorkspace(canonicalRoot, dir)
	if err != nil {
		return "", fmt.Errorf("r1a: cannot validate protected-store boundary: %w", err)
	}
	if within {
		return "", fmt.Errorf("r1a: protected store resolves inside the workspace (fail closed)")
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", fmt.Errorf("r1a: cannot create protected store: %w", err)
	}
	snap := filepath.Join(dir, "snapshot")
	if err := os.MkdirAll(snap, 0o700); err != nil {
		return "", fmt.Errorf("r1a: cannot create snapshot area: %w", err)
	}
	for _, d := range []string{dir, snap} {
		if err := r1aEnsureOwnerOnlyDir(d); err != nil {
			return "", err
		}
	}
	// Re-validate against the now fully-existing, canonicalized store path, in
	// case a symlinked component made the real location land inside the workspace.
	within, err = r1aStoreWithinWorkspace(canonicalRoot, dir)
	if err != nil {
		return "", fmt.Errorf("r1a: cannot re-validate protected-store boundary: %w", err)
	}
	if within {
		return "", fmt.Errorf("r1a: protected store resolves inside the workspace (fail closed)")
	}
	return dir, nil
}

// r1aStoreWithinWorkspace reports whether storeDir resolves to a location at or
// inside canonicalRoot (which must already be canonical).
func r1aStoreWithinWorkspace(canonicalRoot, storeDir string) (bool, error) {
	realStore, err := r1aCanonicalizeExisting(storeDir)
	if err != nil {
		return false, err
	}
	rel, err := filepath.Rel(canonicalRoot, realStore)
	if err != nil {
		return false, err
	}
	if rel == "." {
		return true, nil
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return false, nil
	}
	return true, nil
}

// r1aCanonicalizeExisting canonicalizes the longest existing ancestor of path
// (resolving symlinks) and rejoins any not-yet-created tail, so a boundary check
// is meaningful even before the full store tree exists.
func r1aCanonicalizeExisting(path string) (string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	tail := ""
	cur := abs
	for {
		if _, err := os.Lstat(cur); err == nil {
			real, err := filepath.EvalSymlinks(cur)
			if err != nil {
				return "", err
			}
			if tail == "" {
				return real, nil
			}
			return filepath.Join(real, tail), nil
		}
		parent := filepath.Dir(cur)
		if parent == cur {
			return "", fmt.Errorf("r1a: cannot canonicalize %q", path)
		}
		tail = filepath.Join(filepath.Base(cur), tail)
		cur = parent
	}
}

// r1aEnsureOwnerOnlyDir tightens and then verifies owner-only permissions,
// failing closed if group/other bits cannot be removed.
func r1aEnsureOwnerOnlyDir(dir string) error {
	info, err := os.Stat(dir)
	if err != nil {
		return fmt.Errorf("r1a: cannot stat protected dir: %w", err)
	}
	if info.Mode().Perm()&0o077 != 0 {
		if err := os.Chmod(dir, 0o700); err != nil {
			return fmt.Errorf("r1a: cannot enforce owner-only permissions: %w", err)
		}
		if info, err = os.Stat(dir); err != nil {
			return fmt.Errorf("r1a: cannot re-stat protected dir: %w", err)
		}
	}
	if info.Mode().Perm()&0o077 != 0 {
		return fmt.Errorf("r1a: protected dir is not owner-only (mode %o); fail closed", info.Mode().Perm())
	}
	return nil
}

// r1aWriteProtected atomically writes owner-only (0600) data in the store.
func r1aWriteProtected(path string, data []byte) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".r1a-*.tmp")
	if err != nil {
		return fmt.Errorf("r1a: cannot create protected temp: %w", err)
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if err := tmp.Chmod(0o600); err != nil {
		tmp.Close()
		return err
	}
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpName, path)
}

// r1aSaveSnapshot persists the pre-image bytes and a redacted metadata record.
func r1aSaveSnapshot(storeDir string, preImage []byte, meta r1aSnapshotMeta) (string, error) {
	preImagePath := filepath.Join(storeDir, "snapshot", "preimage")
	if err := r1aWriteProtected(preImagePath, preImage); err != nil {
		return "", err
	}
	metaBytes, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return "", err
	}
	if err := r1aWriteProtected(filepath.Join(storeDir, "snapshot", "meta.json"), metaBytes); err != nil {
		return "", err
	}
	return "snapshot/preimage", nil
}

// r1aSnapshotMeta is the pre-image metadata record (no content, no secrets).
type r1aSnapshotMeta struct {
	SchemaVersion string      `json:"schema_version"`
	RelTarget     string      `json:"rel_target"`
	BeforeHash    string      `json:"before_hash"`
	Mode          os.FileMode `json:"mode"`
	Size          int64       `json:"size"`
}

// r1aReadPreImage reads the stored pre-image bytes for rollback.
func r1aReadPreImage(storeDir string) ([]byte, error) {
	return os.ReadFile(filepath.Join(storeDir, "snapshot", "preimage"))
}

// r1aPersistManifest writes the operation manifest and the ordered events as
// NDJSON. Detail fields are already redacted at emit time; this re-redacts
// defensively before persisting.
func r1aPersistManifest(storeDir string, req r1aOperationRequest, events []r1aEvent, outcome r1aOutcome) error {
	manifest := struct {
		Request r1aOperationRequest `json:"request"`
		Outcome r1aOutcome          `json:"outcome"`
	}{Request: req, Outcome: outcome}
	mb, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}
	if err := r1aWriteProtected(filepath.Join(storeDir, "operation.json"), mb); err != nil {
		return err
	}
	var sb strings.Builder
	for _, e := range events {
		e.Detail = r1aRedact(e.Detail)
		line, err := json.Marshal(e)
		if err != nil {
			return err
		}
		sb.Write(line)
		sb.WriteByte('\n')
	}
	return r1aWriteProtected(filepath.Join(storeDir, "events.ndjson"), []byte(sb.String()))
}

// r1aCleanupSuccess removes a successful operation's artifacts ONLY after
// VERIFIED completion. Contract §7 permits compacting/removing successful-run
// artifacts "only after verified completion", and §8 classifies
// completed_unverified as "not success". Therefore this deletion boundary
// refuses completed_unverified (its snapshot/preimage/evidence is retained),
// along with every failed/denied/conflict/recovery outcome. The rule is
// enforced HERE, at the actual os.RemoveAll boundary, not only in the
// CleanupSuccess wrapper. Deletion scope is unchanged: exactly this one
// protected store directory, never a recursive sweep of an uncertain path.
func r1aCleanupSuccess(storeDir string, outcome r1aOutcome) error {
	if outcome != r1aCompletedVerified {
		return fmt.Errorf("r1a: refusing to remove evidence for outcome %s; contract §7 permits removal only after verified completion", outcome)
	}
	return os.RemoveAll(storeDir)
}

// r1aSanitizeLabel makes a filesystem-safe fragment from a checkpoint label.
func r1aSanitizeLabel(label string) string {
	label = strings.TrimSpace(label)
	if label == "" {
		label = "unlabeled"
	}
	var b strings.Builder
	for _, r := range label {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '-', r == '_':
			b.WriteRune(r)
		default:
			b.WriteByte('_')
		}
	}
	out := b.String()
	if len(out) > 48 {
		out = out[:48]
	}
	return out
}
