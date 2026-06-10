package app

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	stateDir, err := os.MkdirTemp("", "jini-app-test-state-")
	if err != nil {
		panic(err)
	}
	previous, hadPrevious := os.LookupEnv("JINI_STATE_DIR")
	if err := os.Setenv("JINI_STATE_DIR", stateDir); err != nil {
		panic(err)
	}
	code := m.Run()
	if hadPrevious {
		_ = os.Setenv("JINI_STATE_DIR", previous)
	} else {
		_ = os.Unsetenv("JINI_STATE_DIR")
	}
	_ = os.RemoveAll(stateDir)
	os.Exit(code)
}
