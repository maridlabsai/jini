//go:build unix

package app

// Unix hard-link and filesystem/mount-boundary checks via the stat device and
// link-count fields. Build-tagged so non-Unix platforms fail closed instead of
// silently skipping the guarantee (contract §5).

import (
	"fmt"
	"os"
	"syscall"
)

// r1aCheckBoundaries rejects hard-linked targets and targets on a different
// filesystem/mount than the workspace root.
func r1aCheckBoundaries(root *os.Root, info os.FileInfo) error {
	st, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return fmt.Errorf("r1a: cannot verify file identity on this platform (fail closed)")
	}
	if st.Nlink > 1 {
		return fmt.Errorf("r1a: hard-linked target rejected (link count %d)", st.Nlink)
	}
	rootInfo, err := root.Stat(".")
	if err != nil {
		return fmt.Errorf("r1a: cannot stat workspace root: %w", err)
	}
	rst, ok := rootInfo.Sys().(*syscall.Stat_t)
	if !ok {
		return fmt.Errorf("r1a: cannot verify workspace-root identity (fail closed)")
	}
	if uint64(st.Dev) != uint64(rst.Dev) {
		return fmt.Errorf("r1a: target crosses a filesystem/mount boundary")
	}
	return nil
}
