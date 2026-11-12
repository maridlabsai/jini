//go:build unix

package app

import (
	"path/filepath"
	"syscall"
	"testing"
)

func TestR1AResolveRejectsSpecialFile(t *testing.T) {
	ws := r1aTestWorkspace(t)
	fifo := filepath.Join(ws, "pipe")
	if err := syscall.Mkfifo(fifo, 0o644); err != nil {
		t.Skipf("cannot create FIFO here: %v", err)
	}
	root, err := r1aOpenWorkspace(ws)
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()
	if _, err := r1aResolveTarget(root, "pipe"); err == nil {
		t.Fatal("special (non-regular) file must be rejected")
	}
}
