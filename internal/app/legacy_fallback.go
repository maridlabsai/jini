package app

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func shouldUseLegacySurface(args []string) bool {
	if len(args) == 0 {
		return false
	}
	first := normalizeCommandName(args[0])
	switch first {
	case "help":
		if len(args) == 1 {
			return false
		}
		second := normalizeCommandName(args[1])
		return second != "all" && second != "commands" && second != "admin"
	case "commands":
		return len(args) > 1
	case "admin":
		return !(len(args) == 1 || (len(args) == 2 && normalizeCommandName(args[1]) == "help"))
	case "doctor":
		return len(args) > 1
	case "provider":
		return !(len(args) == 1 || (len(args) == 2 && normalizeCommandName(args[1]) == "doctor"))
	case "status", "continue":
		return len(args) > 1
	case "open":
		if len(args) <= 1 {
			return false
		}
		if len(args) > 2 {
			return true
		}
		return strings.HasPrefix(strings.TrimSpace(args[1]), "-")
	case "run":
		return !(len(args) == 1 || (len(args) == 2 && (normalizeCommandName(args[1]) == "new" || strings.TrimSpace(args[1]) == "--new")))
	default:
		return false
	}
}

func runLegacyPython(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	sourceRoot, scriptPath, ok := resolveLegacyPythonEntrypoint()
	commandLabel := "jini"
	if len(args) > 0 {
		commandLabel = args[0]
	}
	if !ok {
		if len(args) > 0 {
			fmt.Fprintf(stderr, "Unknown command %q.\n", args[0])
			fmt.Fprintln(stderr, "This command still lives in the legacy Python surface, but no compatible source checkout was found.")
		} else {
			fmt.Fprintln(stderr, "Jini could not open the legacy front door because no compatible source checkout was found.")
		}
		fmt.Fprintln(stderr, "Set JINI_SOURCE_DIR to a Jini checkout or run the command from the repo while the Go move is still in progress.")
		return 1
	}

	commandArgs := append([]string{scriptPath}, args...)
	command := exec.Command("python3", commandArgs...)
	command.Stdout = stdout
	command.Stderr = stderr
	if stdin != nil {
		command.Stdin = stdin
	} else {
		command.Stdin = os.Stdin
	}
	command.Dir = resolveLegacyPythonWorkingDir(sourceRoot)
	command.Env = legacyPythonEnv(sourceRoot)
	if err := command.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return exitErr.ExitCode()
		}
		fmt.Fprintf(stderr, "Could not run legacy Python command %q: %v\n", commandLabel, err)
		return 1
	}
	return 0
}

func resolveLegacyPythonWorkingDir(sourceRoot string) string {
	if cwd := strings.TrimSpace(os.Getenv("JINI_CALLER_CWD")); isUsableWorkingDirectory(cwd) {
		return cwd
	}
	if cwd, err := os.Getwd(); err == nil && isUsableWorkingDirectory(cwd) {
		return cwd
	}
	return sourceRoot
}

func isUsableWorkingDirectory(path string) bool {
	if strings.TrimSpace(path) == "" {
		return false
	}
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func legacyPythonEnv(sourceRoot string) []string {
	env := overrideEnvVar(os.Environ(), "JINI_SOURCE_DIR", sourceRoot)
	if extraPythonPath := os.Getenv("JINI_LEGACY_PYTHONPATH"); extraPythonPath != "" {
		if currentPythonPath := os.Getenv("PYTHONPATH"); currentPythonPath != "" {
			extraPythonPath = extraPythonPath + string(os.PathListSeparator) + currentPythonPath
		}
		env = overrideEnvVar(env, "PYTHONPATH", extraPythonPath)
	}
	return env
}

func overrideEnvVar(env []string, key, value string) []string {
	prefix := key + "="
	filtered := make([]string, 0, len(env)+1)
	for _, entry := range env {
		if strings.HasPrefix(entry, prefix) {
			continue
		}
		filtered = append(filtered, entry)
	}
	return append(filtered, prefix+value)
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
