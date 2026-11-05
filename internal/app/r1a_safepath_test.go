package app

import (
	"os"
	"path/filepath"
	"testing"
)

func TestR1AResolveAcceptsRegularFile(t *testing.T) {
	ws := r1aTestWorkspace(t)
	r1aWriteFile(t, ws, "sub/ok.txt", []byte("hi"), 0o644)
	root, err := r1aOpenWorkspace(ws)
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()
	target, err := r1aResolveTarget(root, "sub/ok.txt")
	if err != nil {
		t.Fatalf("regular file should resolve: %v", err)
	}
	if target.beforeHash != r1aHashBytes([]byte("hi")) {
		t.Fatal("before-hash mismatch")
	}
}

func TestR1AResolveRejections(t *testing.T) {
	ws := r1aTestWorkspace(t)
	r1aWriteFile(t, ws, "real/file.txt", []byte("x"), 0o644)
	if err := os.MkdirAll(filepath.Join(ws, "adir"), 0o755); err != nil {
		t.Fatal(err)
	}
	// leaf symlink
	if err := os.Symlink(filepath.Join(ws, "real/file.txt"), filepath.Join(ws, "leaflink")); err != nil {
		t.Fatal(err)
	}
	// ancestor symlink: linkdir -> real
	if err := os.Symlink(filepath.Join(ws, "real"), filepath.Join(ws, "linkdir")); err != nil {
		t.Fatal(err)
	}

	root, err := r1aOpenWorkspace(ws)
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()

	cases := map[string]string{
		"absolute":         filepath.Join(ws, "real/file.txt"),
		"traversal":        "../escape.txt",
		"root-equiv":       ".",
		"leaf symlink":     "leaflink",
		"ancestor symlink": "linkdir/file.txt",
		"directory":        "adir",
		"missing":          "real/nope.txt",
	}
	for name, rel := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := r1aResolveTarget(root, rel); err == nil {
				t.Fatalf("%s should be rejected", name)
			}
		})
	}
}

func TestR1AResolveRejectsHardLink(t *testing.T) {
	if os.Getenv("GOOS") == "windows" {
		t.Skip("hard-link check is unix-specific")
	}
	ws := r1aTestWorkspace(t)
	a := r1aWriteFile(t, ws, "a.txt", []byte("x"), 0o644)
	if err := os.Link(a, filepath.Join(ws, "b.txt")); err != nil {
		t.Skipf("hard links unsupported here: %v", err)
	}
	root, err := r1aOpenWorkspace(ws)
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()
	if _, err := r1aResolveTarget(root, "b.txt"); err == nil {
		t.Fatal("hard-linked target must be rejected")
	}
}

func TestR1AResolveRejectsNestedRepo(t *testing.T) {
	ws := r1aTestWorkspace(t)
	r1aWriteFile(t, ws, "mod/file.txt", []byte("x"), 0o644)
	if err := os.MkdirAll(filepath.Join(ws, "mod", ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	root, err := r1aOpenWorkspace(ws)
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()
	if _, err := r1aResolveTarget(root, "mod/file.txt"); err == nil {
		t.Fatal("target inside a nested repo/submodule/worktree must be rejected")
	}
}

func TestR1AMountBoundarySkippedHonestly(t *testing.T) {
	t.Skip("cross-filesystem/mount boundary needs a second mount; device-id check is covered by the unix build helper, not runtime-tested here")
}
