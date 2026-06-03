package app

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func resolvePythonExecutableForTest(t *testing.T) string {
	t.Helper()
	output, err := exec.Command(
		"python3",
		"-c",
		"import os, sys; print(os.path.realpath(sys.executable))",
	).Output()
	if err != nil {
		t.Fatalf("resolve python executable: %v", err)
	}
	return strings.TrimSpace(string(output))
}

func TestSafelyRunInteractiveRecoversPanics(t *testing.T) {
	var stderr bytes.Buffer
	exitCode := safelyRunInteractive(&stderr, func() int {
		panic("boom")
	})
	if exitCode != 1 {
		t.Fatalf("expected exit code 1, got %d", exitCode)
	}
	if !strings.Contains(stderr.String(), "Jini hit an unexpected internal error and stopped safely.") {
		t.Fatalf("expected safe panic message, got:\n%s", stderr.String())
	}
}

func TestRunLegacyPythonDelegatesUnsupportedCommand(t *testing.T) {
	root := t.TempDir()
	toolsDir := filepath.Join(root, "tools")
	if err := os.MkdirAll(toolsDir, 0o755); err != nil {
		t.Fatalf("mkdir tools: %v", err)
	}
	scriptPath := filepath.Join(toolsDir, "jini_validate.py")
	script := "import json, sys\nprint(json.dumps({'argv': sys.argv[1:]}))\n"
	if err := os.WriteFile(scriptPath, []byte(script), 0o755); err != nil {
		t.Fatalf("write legacy script: %v", err)
	}

	t.Setenv("JINI_SOURCE_DIR", root)
	previousDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	neutralDir := t.TempDir()
	if err := os.Chdir(neutralDir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	defer func() {
		_ = os.Chdir(previousDir)
	}()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := runLegacyPython([]string{"publish-readiness", "--format", "json"}, nil, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d\nstdout:\n%s\nstderr:\n%s", exitCode, stdout.String(), stderr.String())
	}
	if !strings.Contains(stdout.String(), `"argv": ["publish-readiness", "--format", "json"]`) {
		t.Fatalf("expected legacy argv passthrough, got:\n%s", stdout.String())
	}
}

func TestRunLegacyPythonPreservesCallerWorkingDirectory(t *testing.T) {
	root := t.TempDir()
	toolsDir := filepath.Join(root, "tools")
	if err := os.MkdirAll(toolsDir, 0o755); err != nil {
		t.Fatalf("mkdir tools: %v", err)
	}
	scriptPath := filepath.Join(toolsDir, "jini_validate.py")
	script := "import os\nprint(os.getcwd())\n"
	if err := os.WriteFile(scriptPath, []byte(script), 0o755); err != nil {
		t.Fatalf("write legacy script: %v", err)
	}

	callerDir := t.TempDir()
	previousDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(callerDir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	defer func() {
		_ = os.Chdir(previousDir)
	}()

	t.Setenv("JINI_SOURCE_DIR", root)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := runLegacyPython(nil, nil, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d\nstdout:\n%s\nstderr:\n%s", exitCode, stdout.String(), stderr.String())
	}
	if !strings.Contains(stdout.String(), callerDir) {
		t.Fatalf("expected legacy run to preserve caller cwd %q, got:\n%s", callerDir, stdout.String())
	}
}

func TestRunLegacyPythonFallsBackWhenCallerWorkingDirectoryIsInvalid(t *testing.T) {
	root := t.TempDir()
	toolsDir := filepath.Join(root, "tools")
	if err := os.MkdirAll(toolsDir, 0o755); err != nil {
		t.Fatalf("mkdir tools: %v", err)
	}
	scriptPath := filepath.Join(toolsDir, "jini_validate.py")
	script := "import os\nprint(os.getcwd())\n"
	if err := os.WriteFile(scriptPath, []byte(script), 0o755); err != nil {
		t.Fatalf("write legacy script: %v", err)
	}

	fallbackDir := t.TempDir()
	previousDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(fallbackDir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	defer func() {
		_ = os.Chdir(previousDir)
	}()

	t.Setenv("JINI_SOURCE_DIR", root)
	t.Setenv("JINI_CALLER_CWD", filepath.Join(t.TempDir(), "missing"))

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := runLegacyPython(nil, nil, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d\nstdout:\n%s\nstderr:\n%s", exitCode, stdout.String(), stderr.String())
	}
	if !strings.Contains(stdout.String(), fallbackDir) {
		t.Fatalf("expected legacy run to fall back to working cwd %q, got:\n%s", fallbackDir, stdout.String())
	}
}

func TestRunLegacyPythonReplacesStaleSourceRootEnvForChildProcess(t *testing.T) {
	root := t.TempDir()
	toolsDir := filepath.Join(root, "tools")
	if err := os.MkdirAll(toolsDir, 0o755); err != nil {
		t.Fatalf("mkdir tools: %v", err)
	}
	scriptPath := filepath.Join(toolsDir, "jini_validate.py")
	script := "import os\nprint(os.environ['JINI_SOURCE_DIR'])\n"
	if err := os.WriteFile(scriptPath, []byte(script), 0o755); err != nil {
		t.Fatalf("write legacy script: %v", err)
	}

	staleRoot := t.TempDir()
	previousDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(root); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	defer func() {
		_ = os.Chdir(previousDir)
	}()

	t.Setenv("JINI_SOURCE_DIR", staleRoot)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := runLegacyPython(nil, nil, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d\nstdout:\n%s\nstderr:\n%s", exitCode, stdout.String(), stderr.String())
	}
	if !strings.Contains(stdout.String(), root) {
		t.Fatalf("expected child process to receive resolved source root %q, got:\n%s", root, stdout.String())
	}
	if strings.Contains(stdout.String(), staleRoot) {
		t.Fatalf("expected stale source root %q to be replaced, got:\n%s", staleRoot, stdout.String())
	}
}

func TestResolveLegacyPythonEntrypointPrefersCurrentCheckoutOverStaleEnv(t *testing.T) {
	liveRoot := t.TempDir()
	if err := os.MkdirAll(filepath.Join(liveRoot, "tools"), 0o755); err != nil {
		t.Fatalf("mkdir live tools: %v", err)
	}
	if err := os.WriteFile(filepath.Join(liveRoot, "tools", "jini_validate.py"), []byte("print('live')\n"), 0o755); err != nil {
		t.Fatalf("write live legacy script: %v", err)
	}

	staleRoot := t.TempDir()
	if err := os.MkdirAll(filepath.Join(staleRoot, "tools"), 0o755); err != nil {
		t.Fatalf("mkdir stale tools: %v", err)
	}
	if err := os.WriteFile(filepath.Join(staleRoot, "tools", "jini_validate.py"), []byte("print('stale')\n"), 0o755); err != nil {
		t.Fatalf("write stale legacy script: %v", err)
	}

	previousDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(liveRoot); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	defer func() {
		_ = os.Chdir(previousDir)
	}()

	t.Setenv("JINI_SOURCE_DIR", staleRoot)

	sourceRoot, scriptPath, ok := resolveLegacyPythonEntrypoint()
	if !ok {
		t.Fatalf("expected to resolve legacy Python entrypoint")
	}
	resolvedSourceRoot, err := filepath.EvalSymlinks(sourceRoot)
	if err != nil {
		t.Fatalf("eval source root symlinks: %v", err)
	}
	resolvedLiveRoot, err := filepath.EvalSymlinks(liveRoot)
	if err != nil {
		t.Fatalf("eval live root symlinks: %v", err)
	}
	if resolvedSourceRoot != resolvedLiveRoot {
		t.Fatalf("expected current checkout %q to win over stale env %q, got %q", resolvedLiveRoot, staleRoot, resolvedSourceRoot)
	}
	expectedScriptPath := filepath.Join(liveRoot, "tools", "jini_validate.py")
	resolvedScriptPath, err := filepath.EvalSymlinks(scriptPath)
	if err != nil {
		t.Fatalf("eval script path symlinks: %v", err)
	}
	resolvedExpectedScriptPath, err := filepath.EvalSymlinks(expectedScriptPath)
	if err != nil {
		t.Fatalf("eval expected script path symlinks: %v", err)
	}
	if resolvedScriptPath != resolvedExpectedScriptPath {
		t.Fatalf("expected script path %q, got %q", resolvedExpectedScriptPath, resolvedScriptPath)
	}
}

func TestRunLegacyPythonZeroArgFailureDoesNotPanic(t *testing.T) {
	root := t.TempDir()
	toolsDir := filepath.Join(root, "tools")
	if err := os.MkdirAll(toolsDir, 0o755); err != nil {
		t.Fatalf("mkdir tools: %v", err)
	}
	scriptPath := filepath.Join(toolsDir, "jini_validate.py")
	if err := os.WriteFile(scriptPath, []byte("print('ok')\n"), 0o755); err != nil {
		t.Fatalf("write legacy script: %v", err)
	}

	t.Setenv("JINI_SOURCE_DIR", root)
	t.Setenv("PATH", "")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := runLegacyPython(nil, nil, &stdout, &stderr)
	if exitCode != 1 {
		t.Fatalf("expected exit code 1, got %d\nstdout:\n%s\nstderr:\n%s", exitCode, stdout.String(), stderr.String())
	}
	if !strings.Contains(stderr.String(), `Could not run legacy Python command "jini":`) {
		t.Fatalf("expected zero-arg failure message, got:\n%s", stderr.String())
	}
}

func TestRunLegacyPythonUsesConfiguredInterpreterWhenPathIsMissing(t *testing.T) {
	root := t.TempDir()
	toolsDir := filepath.Join(root, "tools")
	if err := os.MkdirAll(toolsDir, 0o755); err != nil {
		t.Fatalf("mkdir tools: %v", err)
	}
	scriptPath := filepath.Join(toolsDir, "jini_validate.py")
	script := "print('legacy-ok')\n"
	if err := os.WriteFile(scriptPath, []byte(script), 0o755); err != nil {
		t.Fatalf("write legacy script: %v", err)
	}

	pythonPath := resolvePythonExecutableForTest(t)
	previousDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	neutralDir := t.TempDir()
	if err := os.Chdir(neutralDir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	defer func() {
		_ = os.Chdir(previousDir)
	}()

	t.Setenv("JINI_SOURCE_DIR", root)
	t.Setenv("JINI_LEGACY_PYTHON", pythonPath)
	t.Setenv("PATH", "")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := runLegacyPython(nil, nil, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d\nstdout:\n%s\nstderr:\n%s", exitCode, stdout.String(), stderr.String())
	}
	if !strings.Contains(stdout.String(), "legacy-ok") {
		t.Fatalf("expected configured interpreter to run legacy script, got:\n%s", stdout.String())
	}
}

func TestRunLegacyPythonFallsBackWhenConfiguredInterpreterIsNotExecutable(t *testing.T) {
	root := t.TempDir()
	toolsDir := filepath.Join(root, "tools")
	if err := os.MkdirAll(toolsDir, 0o755); err != nil {
		t.Fatalf("mkdir tools: %v", err)
	}
	scriptPath := filepath.Join(toolsDir, "jini_validate.py")
	script := "print('legacy-fallback-ok')\n"
	if err := os.WriteFile(scriptPath, []byte(script), 0o755); err != nil {
		t.Fatalf("write legacy script: %v", err)
	}

	badPython := filepath.Join(t.TempDir(), "bad-python")
	if err := os.WriteFile(badPython, []byte("not executable\n"), 0o644); err != nil {
		t.Fatalf("write bad python: %v", err)
	}

	previousDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	neutralDir := t.TempDir()
	if err := os.Chdir(neutralDir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	defer func() {
		_ = os.Chdir(previousDir)
	}()

	t.Setenv("JINI_SOURCE_DIR", root)
	t.Setenv("JINI_LEGACY_PYTHON", badPython)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := runLegacyPython(nil, nil, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d\nstdout:\n%s\nstderr:\n%s", exitCode, stdout.String(), stderr.String())
	}
	if !strings.Contains(stdout.String(), "legacy-fallback-ok") {
		t.Fatalf("expected fallback interpreter to run legacy script, got:\n%s", stdout.String())
	}
}

func TestRunLegacyPythonFallsBackWhenConfiguredInterpreterIsNotPython(t *testing.T) {
	root := t.TempDir()
	toolsDir := filepath.Join(root, "tools")
	if err := os.MkdirAll(toolsDir, 0o755); err != nil {
		t.Fatalf("mkdir tools: %v", err)
	}
	scriptPath := filepath.Join(toolsDir, "jini_validate.py")
	script := "print('legacy-python-fallback-ok')\n"
	if err := os.WriteFile(scriptPath, []byte(script), 0o755); err != nil {
		t.Fatalf("write legacy script: %v", err)
	}

	notPython := filepath.Join(t.TempDir(), "not-python")
	if err := os.WriteFile(notPython, []byte("#!/bin/sh\necho not-python\n"), 0o755); err != nil {
		t.Fatalf("write not-python: %v", err)
	}

	previousDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	neutralDir := t.TempDir()
	if err := os.Chdir(neutralDir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	defer func() {
		_ = os.Chdir(previousDir)
	}()

	t.Setenv("JINI_SOURCE_DIR", root)
	t.Setenv("JINI_LEGACY_PYTHON", notPython)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := runLegacyPython(nil, nil, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d\nstdout:\n%s\nstderr:\n%s", exitCode, stdout.String(), stderr.String())
	}
	if !strings.Contains(stdout.String(), "legacy-python-fallback-ok") {
		t.Fatalf("expected fallback interpreter to run legacy script, got:\n%s", stdout.String())
	}
}

func TestRunLegacyPythonFallsBackWhenConfiguredInterpreterSpoofsProbe(t *testing.T) {
	root := t.TempDir()
	toolsDir := filepath.Join(root, "tools")
	if err := os.MkdirAll(toolsDir, 0o755); err != nil {
		t.Fatalf("mkdir tools: %v", err)
	}
	scriptPath := filepath.Join(toolsDir, "jini_validate.py")
	script := "print('legacy-spoof-fallback-ok')\n"
	if err := os.WriteFile(scriptPath, []byte(script), 0o755); err != nil {
		t.Fatalf("write legacy script: %v", err)
	}

	spoofPython := filepath.Join(t.TempDir(), "spoof-python")
	spoofScript := "#!/bin/sh\necho /bin/sh\n"
	if err := os.WriteFile(spoofPython, []byte(spoofScript), 0o755); err != nil {
		t.Fatalf("write spoof python: %v", err)
	}

	previousDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	neutralDir := t.TempDir()
	if err := os.Chdir(neutralDir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	defer func() {
		_ = os.Chdir(previousDir)
	}()

	t.Setenv("JINI_SOURCE_DIR", root)
	t.Setenv("JINI_LEGACY_PYTHON", spoofPython)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := runLegacyPython(nil, nil, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d\nstdout:\n%s\nstderr:\n%s", exitCode, stdout.String(), stderr.String())
	}
	if !strings.Contains(stdout.String(), "legacy-spoof-fallback-ok") {
		t.Fatalf("expected real python fallback to run legacy script, got:\n%s", stdout.String())
	}
}

func TestRunLegacyPythonReplacesBrokenConfiguredInterpreterInChildEnv(t *testing.T) {
	root := t.TempDir()
	toolsDir := filepath.Join(root, "tools")
	if err := os.MkdirAll(toolsDir, 0o755); err != nil {
		t.Fatalf("mkdir tools: %v", err)
	}
	scriptPath := filepath.Join(toolsDir, "jini_validate.py")
	script := "import os\nprint(os.environ['JINI_LEGACY_PYTHON'])\n"
	if err := os.WriteFile(scriptPath, []byte(script), 0o755); err != nil {
		t.Fatalf("write legacy script: %v", err)
	}

	badPython := filepath.Join(t.TempDir(), "bad-python")
	if err := os.WriteFile(badPython, []byte("not executable\n"), 0o644); err != nil {
		t.Fatalf("write bad python: %v", err)
	}

	pythonPath := resolvePythonExecutableForTest(t)
	previousDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	neutralDir := t.TempDir()
	if err := os.Chdir(neutralDir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	defer func() {
		_ = os.Chdir(previousDir)
	}()

	t.Setenv("JINI_SOURCE_DIR", root)
	t.Setenv("JINI_LEGACY_PYTHON", badPython)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := runLegacyPython(nil, nil, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d\nstdout:\n%s\nstderr:\n%s", exitCode, stdout.String(), stderr.String())
	}
	if !strings.Contains(stdout.String(), pythonPath) {
		t.Fatalf("expected child env to carry resolved python path %q, got:\n%s", pythonPath, stdout.String())
	}
	if strings.Contains(stdout.String(), badPython) {
		t.Fatalf("expected broken configured interpreter %q to be replaced, got:\n%s", badPython, stdout.String())
	}
}

func TestRecognizedGoCommandFallsBackToLegacyForUnsupportedFlags(t *testing.T) {
	root := t.TempDir()
	toolsDir := filepath.Join(root, "tools")
	if err := os.MkdirAll(toolsDir, 0o755); err != nil {
		t.Fatalf("mkdir tools: %v", err)
	}
	scriptPath := filepath.Join(toolsDir, "jini_validate.py")
	script := "import json, sys\nprint(json.dumps({'argv': sys.argv[1:]}))\n"
	if err := os.WriteFile(scriptPath, []byte(script), 0o755); err != nil {
		t.Fatalf("write legacy script: %v", err)
	}

	t.Setenv("JINI_SOURCE_DIR", root)
	previousDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	neutralDir := t.TempDir()
	if err := os.Chdir(neutralDir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	defer func() {
		_ = os.Chdir(previousDir)
	}()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := Run([]string{"doctor", "--format", "json"}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d\nstdout:\n%s\nstderr:\n%s", exitCode, stdout.String(), stderr.String())
	}
	if !strings.Contains(stdout.String(), `"argv": ["doctor", "--format", "json"]`) {
		t.Fatalf("expected legacy doctor argv passthrough, got:\n%s", stdout.String())
	}
}

func TestLegacyPythonEnvPrependsConfiguredPythonPath(t *testing.T) {
	t.Setenv("PYTHONPATH", "/existing/site-packages")
	t.Setenv("JINI_LEGACY_PYTHONPATH", "/stable/vendor")
	t.Setenv("JINI_SOURCE_DIR", "/stale/source")
	t.Setenv("JINI_LEGACY_PYTHON", "/stale/python")

	env := legacyPythonEnv("/tmp/jini-source", "/tmp/jini-python")
	joined := strings.Join(env, "\n")

	if !strings.Contains(joined, "JINI_SOURCE_DIR=/tmp/jini-source") {
		t.Fatalf("expected source dir in env, got:\n%s", joined)
	}
	if strings.Contains(joined, "JINI_SOURCE_DIR=/stale/source") {
		t.Fatalf("expected stale source dir to be removed, got:\n%s", joined)
	}
	if !strings.Contains(joined, "JINI_LEGACY_PYTHON=/tmp/jini-python") {
		t.Fatalf("expected resolved legacy python in env, got:\n%s", joined)
	}
	if strings.Contains(joined, "JINI_LEGACY_PYTHON=/stale/python") {
		t.Fatalf("expected stale legacy python to be removed, got:\n%s", joined)
	}
	if !strings.Contains(joined, "PYTHONPATH=/stable/vendor"+string(os.PathListSeparator)+"/existing/site-packages") {
		t.Fatalf("expected prepended PYTHONPATH, got:\n%s", joined)
	}
}

func TestShouldUseLegacyFrontDoorHonorsEnv(t *testing.T) {
	t.Setenv("JINI_USE_LEGACY_FRONT_DOOR", "1")
	if !shouldUseLegacyFrontDoor() {
		t.Fatalf("expected legacy front door to be enabled")
	}

	t.Setenv("JINI_USE_LEGACY_FRONT_DOOR", "0")
	if shouldUseLegacyFrontDoor() {
		t.Fatalf("expected legacy front door to be disabled")
	}
}

func TestRunLauncherFallsBackToGoPromptWhenLegacyPythonIsUnavailable(t *testing.T) {
	root := t.TempDir()
	toolsDir := filepath.Join(root, "tools")
	if err := os.MkdirAll(toolsDir, 0o755); err != nil {
		t.Fatalf("mkdir tools: %v", err)
	}
	if err := os.WriteFile(filepath.Join(toolsDir, "jini_validate.py"), []byte("print('legacy')\n"), 0o755); err != nil {
		t.Fatalf("write legacy script: %v", err)
	}

	stateDir := t.TempDir()
	workingDir := t.TempDir()
	previousDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(workingDir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	defer func() {
		_ = os.Chdir(previousDir)
	}()

	t.Setenv("JINI_SOURCE_DIR", root)
	t.Setenv("JINI_USE_LEGACY_FRONT_DOOR", "1")
	t.Setenv("JINI_STATE_DIR", stateDir)
	t.Setenv("PATH", "")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := runLauncher(nil, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d\nstdout:\n%s\nstderr:\n%s", exitCode, stdout.String(), stderr.String())
	}
	if !strings.Contains(stdout.String(), "Paste what you want finished.") {
		t.Fatalf("expected Go prompt fallback, got:\n%s", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected no stderr output, got:\n%s", stderr.String())
	}
}

func TestRunLauncherFallsBackToGoIntakeWhenLegacyPythonIsUnavailable(t *testing.T) {
	root := t.TempDir()
	toolsDir := filepath.Join(root, "tools")
	if err := os.MkdirAll(toolsDir, 0o755); err != nil {
		t.Fatalf("mkdir tools: %v", err)
	}
	if err := os.WriteFile(filepath.Join(toolsDir, "jini_validate.py"), []byte("print('legacy')\n"), 0o755); err != nil {
		t.Fatalf("write legacy script: %v", err)
	}

	stateDir := t.TempDir()
	workingDir := t.TempDir()
	previousDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(workingDir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	defer func() {
		_ = os.Chdir(previousDir)
	}()

	t.Setenv("JINI_SOURCE_DIR", root)
	t.Setenv("JINI_USE_LEGACY_FRONT_DOOR", "1")
	t.Setenv("JINI_STATE_DIR", stateDir)
	t.Setenv("PATH", "")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := runLauncher(strings.NewReader("draft a launch checklist\n"), &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d\nstdout:\n%s\nstderr:\n%s", exitCode, stdout.String(), stderr.String())
	}
	if !strings.Contains(stdout.String(), "Paste what you want finished.") {
		t.Fatalf("expected intake prompt in fallback path, got:\n%s", stdout.String())
	}
	if strings.Contains(stderr.String(), "Could not run legacy Python command") {
		t.Fatalf("expected intake fallback instead of legacy python failure, got:\n%s", stderr.String())
	}
}

func TestRunLauncherUsesConfiguredLegacyPythonWhenPathIsMissing(t *testing.T) {
	root := t.TempDir()
	toolsDir := filepath.Join(root, "tools")
	if err := os.MkdirAll(toolsDir, 0o755); err != nil {
		t.Fatalf("mkdir tools: %v", err)
	}
	if err := os.WriteFile(filepath.Join(toolsDir, "jini_validate.py"), []byte("print('legacy-front-door')\n"), 0o755); err != nil {
		t.Fatalf("write legacy script: %v", err)
	}

	pythonPath := resolvePythonExecutableForTest(t)

	stateDir := t.TempDir()
	workingDir := t.TempDir()
	previousDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(workingDir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	defer func() {
		_ = os.Chdir(previousDir)
	}()

	t.Setenv("JINI_SOURCE_DIR", root)
	t.Setenv("JINI_USE_LEGACY_FRONT_DOOR", "1")
	t.Setenv("JINI_LEGACY_PYTHON", pythonPath)
	t.Setenv("JINI_STATE_DIR", stateDir)
	t.Setenv("PATH", "")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := runLauncher(nil, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d\nstdout:\n%s\nstderr:\n%s", exitCode, stdout.String(), stderr.String())
	}
	if !strings.Contains(stdout.String(), "legacy-front-door") {
		t.Fatalf("expected configured legacy python to keep front door alive, got:\n%s", stdout.String())
	}
}

func TestRunLauncherFallsBackToGoPromptWhenConfiguredLegacyPythonIsNotExecutable(t *testing.T) {
	root := t.TempDir()
	toolsDir := filepath.Join(root, "tools")
	if err := os.MkdirAll(toolsDir, 0o755); err != nil {
		t.Fatalf("mkdir tools: %v", err)
	}
	if err := os.WriteFile(filepath.Join(toolsDir, "jini_validate.py"), []byte("print('legacy')\n"), 0o755); err != nil {
		t.Fatalf("write legacy script: %v", err)
	}

	badPython := filepath.Join(t.TempDir(), "bad-python")
	if err := os.WriteFile(badPython, []byte("not executable\n"), 0o644); err != nil {
		t.Fatalf("write bad python: %v", err)
	}

	stateDir := t.TempDir()
	workingDir := t.TempDir()
	previousDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(workingDir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	defer func() {
		_ = os.Chdir(previousDir)
	}()

	t.Setenv("JINI_SOURCE_DIR", root)
	t.Setenv("JINI_USE_LEGACY_FRONT_DOOR", "1")
	t.Setenv("JINI_LEGACY_PYTHON", badPython)
	t.Setenv("JINI_STATE_DIR", stateDir)
	t.Setenv("PATH", "")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := runLauncher(nil, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d\nstdout:\n%s\nstderr:\n%s", exitCode, stdout.String(), stderr.String())
	}
	if !strings.Contains(stdout.String(), "Paste what you want finished.") {
		t.Fatalf("expected Go prompt fallback, got:\n%s", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected no stderr output, got:\n%s", stderr.String())
	}
}

func TestRunLauncherFallsBackToGoPromptWhenConfiguredLegacyPythonIsNotPython(t *testing.T) {
	root := t.TempDir()
	toolsDir := filepath.Join(root, "tools")
	if err := os.MkdirAll(toolsDir, 0o755); err != nil {
		t.Fatalf("mkdir tools: %v", err)
	}
	if err := os.WriteFile(filepath.Join(toolsDir, "jini_validate.py"), []byte("print('legacy')\n"), 0o755); err != nil {
		t.Fatalf("write legacy script: %v", err)
	}

	notPython := filepath.Join(t.TempDir(), "not-python")
	if err := os.WriteFile(notPython, []byte("#!/bin/sh\necho not-python\n"), 0o755); err != nil {
		t.Fatalf("write not-python: %v", err)
	}

	stateDir := t.TempDir()
	workingDir := t.TempDir()
	previousDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(workingDir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	defer func() {
		_ = os.Chdir(previousDir)
	}()

	t.Setenv("JINI_SOURCE_DIR", root)
	t.Setenv("JINI_USE_LEGACY_FRONT_DOOR", "1")
	t.Setenv("JINI_LEGACY_PYTHON", notPython)
	t.Setenv("JINI_STATE_DIR", stateDir)
	t.Setenv("PATH", "")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := runLauncher(nil, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d\nstdout:\n%s\nstderr:\n%s", exitCode, stdout.String(), stderr.String())
	}
	if !strings.Contains(stdout.String(), "Paste what you want finished.") {
		t.Fatalf("expected Go prompt fallback, got:\n%s", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected no stderr output, got:\n%s", stderr.String())
	}
}

func TestRunLauncherFallsBackToGoPromptWhenConfiguredInterpreterSpoofsPython(t *testing.T) {
	root := t.TempDir()
	toolsDir := filepath.Join(root, "tools")
	if err := os.MkdirAll(toolsDir, 0o755); err != nil {
		t.Fatalf("mkdir tools: %v", err)
	}
	if err := os.WriteFile(filepath.Join(toolsDir, "jini_validate.py"), []byte("print('legacy')\n"), 0o755); err != nil {
		t.Fatalf("write legacy script: %v", err)
	}

	spoofPython := filepath.Join(t.TempDir(), "spoof-python")
	spoofScript := "#!/bin/sh\necho /bin/sh\n"
	if err := os.WriteFile(spoofPython, []byte(spoofScript), 0o755); err != nil {
		t.Fatalf("write spoof python: %v", err)
	}

	stateDir := t.TempDir()
	workingDir := t.TempDir()
	previousDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(workingDir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	defer func() {
		_ = os.Chdir(previousDir)
	}()

	t.Setenv("JINI_SOURCE_DIR", root)
	t.Setenv("JINI_USE_LEGACY_FRONT_DOOR", "1")
	t.Setenv("JINI_LEGACY_PYTHON", spoofPython)
	t.Setenv("JINI_STATE_DIR", stateDir)
	t.Setenv("PATH", "")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := runLauncher(nil, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d\nstdout:\n%s\nstderr:\n%s", exitCode, stdout.String(), stderr.String())
	}
	if !strings.Contains(stdout.String(), "Paste what you want finished.") {
		t.Fatalf("expected Go prompt fallback, got:\n%s", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected no stderr output, got:\n%s", stderr.String())
	}
}
