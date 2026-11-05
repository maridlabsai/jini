package app

// Executable Go toolchain-floor check (contract §5). R1's descriptor-relative
// confinement assumes a PATCHED toolchain: Go >= 1.26.5 (earlier 1.26 patches
// are affected by GO-2026-4970). This is enforced by an executable check, NOT by
// a go.mod `toolchain` directive — this step does not change go.mod and does not
// activate the planned R1 promotion gate.

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	r1aGoFloorMajor = 1
	r1aGoFloorMinor = 26
	r1aGoFloorPatch = 5
)

// r1aEnforceGoFloor fails closed unless version meets the >=1.26.5 floor. It is
// called at the protected entry point (r1aPrepare) so the toolchain floor is
// enforced before any operation that depends on the protected filesystem
// primitives — not left as a tested helper. It does not install Go or modify
// go.mod, and it does not activate the planned R1 promotion gate.
func r1aEnforceGoFloor(version string) error {
	ok, err := r1aGoVersionOK(version)
	if err != nil {
		return fmt.Errorf("r1a: Go toolchain floor cannot be verified (%q): %w; refusing protected filesystem operations", version, err)
	}
	if !ok {
		return fmt.Errorf("r1a: Go toolchain %q is below the required %d.%d.%d floor (GO-2026-4970); refusing protected filesystem operations",
			version, r1aGoFloorMajor, r1aGoFloorMinor, r1aGoFloorPatch)
	}
	return nil
}

// r1aGoVersionOK reports whether a Go version string meets the >=1.26.5 floor.
// Accepts "go1.26.5", "go1.26.5 darwin/amd64", "1.26.5", etc. A development or
// unparseable version fails closed with an error.
func r1aGoVersionOK(v string) (bool, error) {
	maj, min, pat, err := r1aParseGoVersion(v)
	if err != nil {
		return false, err
	}
	switch {
	case maj != r1aGoFloorMajor:
		return maj > r1aGoFloorMajor, nil
	case min != r1aGoFloorMinor:
		return min > r1aGoFloorMinor, nil
	default:
		return pat >= r1aGoFloorPatch, nil
	}
}

func r1aParseGoVersion(v string) (int, int, int, error) {
	v = strings.TrimSpace(v)
	v = strings.TrimPrefix(v, "go")
	if i := strings.IndexByte(v, ' '); i >= 0 {
		v = v[:i]
	}
	if v == "" || strings.HasPrefix(v, "devel") {
		return 0, 0, 0, fmt.Errorf("r1a: unparseable or development Go version %q (fail closed)", v)
	}
	parts := strings.SplitN(v, ".", 3)
	nums := []int{0, 0, 0}
	for i := 0; i < 3; i++ {
		if i >= len(parts) {
			break
		}
		seg := parts[i]
		j := 0
		for j < len(seg) && seg[j] >= '0' && seg[j] <= '9' {
			j++
		}
		if j == 0 {
			return 0, 0, 0, fmt.Errorf("r1a: unparseable Go version segment %q in %q", seg, v)
		}
		n, err := strconv.Atoi(seg[:j])
		if err != nil {
			return 0, 0, 0, err
		}
		nums[i] = n
	}
	return nums[0], nums[1], nums[2], nil
}
