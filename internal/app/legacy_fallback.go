package app

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

func runLegacyPython(args []string, stdout, stderr io.Writer) int {
	sourceRoot, scriptPath, ok := resolveLegacyPythonEntrypoint()
	if !ok {
		fmt.Fprintf(stderr, "Unknown command %q.\n", args[0])
		fmt.Fprintln(stderr, "This command still lives in the legacy Python surface, but no compatible source checkout was found.")
		fmt.Fprintln(stderr, "Set JINI_SOURCE_DIR to a Jini checkout or run the command from the repo while the Go move is still in progress.")
		return 1
	}

	commandArgs := append([]string{scriptPath}, args...)
	command := exec.Command("python3", commandArgs...)
	command.Stdout = stdout
	command.Stderr = stderr
	command.Stdin = os.Stdin
	command.Dir = sourceRoot
	command.Env = legacyPythonEnv(sourceRoot)
	if err := command.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return exitErr.ExitCode()
		}
		fmt.Fprintf(stderr, "Could not run legacy Python command %q: %v\n", args[0], err)
		return 1
	}
	return 0
}

func legacyPythonEnv(sourceRoot string) []string {
	env := append(os.Environ(), "JINI_SOURCE_DIR="+sourceRoot)
	if extraPythonPath := os.Getenv("JINI_LEGACY_PYTHONPATH"); extraPythonPath != "" {
		if currentPythonPath := os.Getenv("PYTHONPATH"); currentPythonPath != "" {
			extraPythonPath = extraPythonPath + string(os.PathListSeparator) + currentPythonPath
		}
		env = append(env, "PYTHONPATH="+extraPythonPath)
	}
	return env
}

func resolveLegacyPythonEntrypoint() (string, string, bool) {
	candidates := []string{}
	if envRoot := os.Getenv("JINI_SOURCE_DIR"); envRoot != "" {
		candidates = append(candidates, envRoot)
	}
	if cwd, err := os.Getwd(); err == nil {
		candidates = append(candidates, cwd)
	}
	if executablePath, err := os.Executable(); err == nil {
		candidates = append(candidates, filepath.Dir(executablePath))
	}
	for _, candidate := range candidates {
		if sourceRoot, scriptPath, ok := findLegacyPythonEntrypoint(candidate); ok {
			return sourceRoot, scriptPath, true
		}
	}
	return "", "", false
}

func findLegacyPythonEntrypoint(start string) (string, string, bool) {
	current := start
	for {
		scriptPath := filepath.Join(current, "tools", "jini_validate.py")
		if info, err := os.Stat(scriptPath); err == nil && !info.IsDir() {
			return current, scriptPath, true
		}
		parent := filepath.Dir(current)
		if parent == current {
			return "", "", false
		}
		current = parent
	}
}
