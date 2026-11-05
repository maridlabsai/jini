//go:build !unix

package app

// Non-Unix platforms (notably Windows) cannot prove hard-link and mount-boundary
// safety with the primitives used here, so R1A fails closed rather than silently
// weakening the guarantee (contract §5). A Windows-native implementation
// (GetFileInformationByHandle for link count / volume serial) is future work and
// is intentionally NOT faked here.

import (
	"fmt"
	"os"
)

func r1aCheckBoundaries(root *os.Root, info os.FileInfo) error {
	return fmt.Errorf("r1a: hard-link and mount-boundary safety is unproven on this platform (fail closed)")
}
