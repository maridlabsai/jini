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
	pythonCommand, err := resolveLegacyPythonExecutable()
	if err != nil {
		fmt.Fprintf(stderr, "Could not run legacy Python command %q: %v\n", commandLabel, err)
		return 1
	}
	return runLegacyPythonWithExecutable(sourceRoot, pythonCommand, commandLabel, commandArgs, stdin, stdout, stderr)
}

func runLegacyPythonWithExecutable(
	sourceRoot string,
	pythonCommand string,
	commandLabel string,
	commandArgs []string,
	stdin io.Reader,
	stdout io.Writer,
	stderr io.Writer,
) int {
	command := exec.Command(pythonCommand, commandArgs...)
	command.Stdout = stdout
	command.Stderr = stderr
	if stdin != nil {
		command.Stdin = stdin
	} else {
		command.Stdin = os.Stdin
	}
	command.Dir = resolveLegacyPythonWorkingDir(sourceRoot)
	command.Env = legacyPythonEnv(sourceRoot, pythonCommand)
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

func resolveLegacyPythonExecutable() (string, error) {
	if configured := strings.TrimSpace(os.Getenv("JINI_LEGACY_PYTHON")); configured != "" {
		if resolved, err := exec.LookPath(configured); err == nil {
			if canonical, canonicalErr := canonicalPythonExecutable(resolved); canonicalErr == nil {
				return canonical, nil
			}
		}
	}
	resolved, err := exec.LookPath("python3")
	if err != nil {
		return "", err
	}
	return canonicalPythonExecutable(resolved)
}

func canonicalPythonExecutable(command string) (string, error) {
	probe := "import os, sys\nprint(os.path.realpath(sys.executable))\nprint(sys.version_info[0])\n"
	cmd := exec.Command(command, "-")
	cmd.Stdin = strings.NewReader(probe)
	output, err := cmd.Output()
	if err == nil {
		lines := strings.Split(strings.TrimSpace(string(output)), "\n")
		if len(lines) == 2 && strings.TrimSpace(lines[1]) == "3" {
			candidate := strings.TrimSpace(lines[0])
			if info, statErr := os.Stat(candidate); statErr == nil && !info.IsDir() {
				return candidate, nil
			}
		}
	}
	return "", fmt.Errorf("%q is not a usable Python interpreter", command)
}

func isUsableWorkingDirectory(path string) bool {
	if strings.TrimSpace(path) == "" {
		return false
	}
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func legacyPythonEnv(sourceRoot, pythonCommand string) []string {
	env := overrideEnvVar(os.Environ(), "JINI_SOURCE_DIR", sourceRoot)
	env = overrideEnvVar(env, "JINI_LEGACY_PYTHON", pythonCommand)
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
	var (
		cwdRoot    string
		cwdScript  string
		cwdOK      bool
		envRoot    string
		envScript  string
		envOK      bool
		execRoot   string
		execScript string
		execOK     bool
	)
	if cwd, err := os.Getwd(); err == nil {
		cwdRoot, cwdScript, cwdOK = findLegacyPythonEntrypoint(cwd)
	}
	if envCandidate := strings.TrimSpace(os.Getenv("JINI_SOURCE_DIR")); envCandidate != "" {
		envRoot, envScript, envOK = findLegacyPythonEntrypoint(envCandidate)
	}
	if executablePath, err := os.Executable(); err == nil {
		execRoot, execScript, execOK = findLegacyPythonEntrypoint(filepath.Dir(executablePath))
	}

	return selectLegacyPythonEntrypoint(
		cwdRoot, cwdScript, cwdOK,
		envRoot, envScript, envOK,
		execRoot, execScript, execOK,
	)
}

func selectLegacyPythonEntrypoint(
	cwdRoot string,
	cwdScript string,
	cwdOK bool,
	envRoot string,
	envScript string,
	envOK bool,
	execRoot string,
	execScript string,
	execOK bool,
) (string, string, bool) {
	if cwdOK && isRecognizedJiniSourceRoot(cwdRoot) {
		return cwdRoot, cwdScript, true
	}
	if envOK {
		return envRoot, envScript, true
	}
	if execOK {
		return execRoot, execScript, true
	}
	if cwdOK {
		return cwdRoot, cwdScript, true
	}
	return "", "", false
}

func isRecognizedJiniSourceRoot(root string) bool {
	requiredFiles := []string{
		filepath.Join(root, "go.mod"),
		filepath.Join(root, "cmd", "jini", "main.go"),
		filepath.Join(root, "tools", "jini_validate.py"),
	}
	for _, required := range requiredFiles {
		info, err := os.Stat(required)
		if err != nil || info.IsDir() {
			return false
		}
	}
	return true
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
