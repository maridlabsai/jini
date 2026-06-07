package app

import (
	"fmt"
	"io"
)

func maybeHandleSimpleAnswer(raw string, stdout io.Writer) bool {
	switch normalizeName(raw) {
	case "what is the capital of france", "capital of france":
		fmt.Fprintln(stdout, "Paris.")
		return true
	default:
		return false
	}
}
