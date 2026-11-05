package app

import (
	"runtime"
	"testing"
)

func TestR1AGoVersionFloor(t *testing.T) {
	cases := []struct {
		in   string
		want bool
		err  bool
	}{
		{"go1.26.5", true, false},
		{"go1.26.5 darwin/amd64", true, false},
		{"1.26.5", true, false},
		{"go1.26.6", true, false},
		{"go1.27.0", true, false},
		{"go2.0.0", true, false},
		{"go1.26.4", false, false},
		{"go1.26", false, false},
		{"go1.25.9", false, false},
		{"devel go1.27-abcdef", false, true},
		{"", false, true},
		{"gobananas", false, true},
	}
	for _, c := range cases {
		ok, err := r1aGoVersionOK(c.in)
		if c.err {
			if err == nil {
				t.Fatalf("%q should error", c.in)
			}
			continue
		}
		if err != nil {
			t.Fatalf("%q unexpected error: %v", c.in, err)
		}
		if ok != c.want {
			t.Fatalf("%q: want %v got %v", c.in, c.want, ok)
		}
	}
}

// TestR1ARuntimeMeetsFloor checks the toolchain building the tests. If it is
// below 1.26.5 this fails honestly (contract §5) — it does NOT upgrade anything.
func TestR1ARuntimeMeetsFloor(t *testing.T) {
	ok, err := r1aGoVersionOK(runtime.Version())
	if err != nil {
		t.Skipf("non-release toolchain %q; floor check skipped honestly", runtime.Version())
	}
	if !ok {
		t.Fatalf("toolchain %q is below the required Go 1.26.5 floor (GO-2026-4970)", runtime.Version())
	}
}
