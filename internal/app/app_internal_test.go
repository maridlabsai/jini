package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
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

func createRecognizedLegacySourceRoot(t *testing.T, script string) string {
	t.Helper()
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "cmd", "jini"), 0o755); err != nil {
		t.Fatalf("mkdir cmd/jini: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(root, "tools"), 0o755); err != nil {
		t.Fatalf("mkdir tools: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module github.com/maridlabsai/jini\n"), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "cmd", "jini", "main.go"), []byte("package main\n"), 0o644); err != nil {
		t.Fatalf("write main.go: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "tools", "jini_validate.py"), []byte(script), 0o755); err != nil {
		t.Fatalf("write legacy script: %v", err)
	}
	return root
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
	script := "import json, sys\nprint(json.dumps({'argv': sys.argv[1:]}))\n"
	root := createRecognizedLegacySourceRoot(t, script)

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

func TestRunLegacyPythonRejectsProbeTimeScriptMutation(t *testing.T) {
	root := createRecognizedLegacySourceRoot(t, "print('legacy-before-mutation')\n")
	pythonPath := resolvePythonExecutableForTest(t)
	scriptPath := filepath.Join(root, "tools", "jini_validate.py")
	markerPath := filepath.Join(t.TempDir(), "probe-marker")
	wrapperPath := filepath.Join(t.TempDir(), "python-wrapper")
	wrapper := "#!/bin/sh\n" +
		"printf 'used' > \"" + markerPath + "\"\n" +
		"printf 'print(\\\"legacy-after-mutation\\\")\\n' > \"" + scriptPath + "\"\n" +
		"exec \"" + pythonPath + "\" \"$@\"\n"
	if err := os.WriteFile(wrapperPath, []byte(wrapper), 0o755); err != nil {
		t.Fatalf("write wrapper: %v", err)
	}

	t.Setenv("JINI_SOURCE_DIR", root)
	t.Setenv("JINI_LEGACY_PYTHON", wrapperPath)
	t.Setenv("PATH", "")

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
	if exitCode != 1 {
		t.Fatalf("expected exit code 1, got %d\nstdout:\n%s\nstderr:\n%s", exitCode, stdout.String(), stderr.String())
	}
	if !strings.Contains(stderr.String(), `Could not run legacy Python command "publish-readiness": legacy entrypoint changed during launch`) {
		t.Fatalf("expected mutation guard message, got:\n%s", stderr.String())
	}
	if strings.Contains(stdout.String(), "legacy-before-mutation") || strings.Contains(stdout.String(), "legacy-after-mutation") {
		t.Fatalf("expected neither original nor mutated script to execute, got stdout:\n%s", stdout.String())
	}
	if _, err := os.Stat(markerPath); err != nil {
		t.Fatalf("expected probe wrapper to run and create marker: %v", err)
	}
}

func TestRunLegacyPythonPreservesCallerWorkingDirectory(t *testing.T) {
	script := "import os\nprint(os.getcwd())\n"
	root := createRecognizedLegacySourceRoot(t, script)

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
	script := "import os\nprint(os.getcwd())\n"
	root := createRecognizedLegacySourceRoot(t, script)

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
	script := "import os\nprint(os.environ['JINI_SOURCE_DIR'])\n"
	root := createRecognizedLegacySourceRoot(t, script)

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
	if err := os.MkdirAll(filepath.Join(liveRoot, "cmd", "jini"), 0o755); err != nil {
		t.Fatalf("mkdir live cmd: %v", err)
	}
	if err := os.WriteFile(filepath.Join(liveRoot, "go.mod"), []byte("module github.com/maridlabsai/jini\n"), 0o644); err != nil {
		t.Fatalf("write live go.mod: %v", err)
	}
	if err := os.WriteFile(filepath.Join(liveRoot, "cmd", "jini", "main.go"), []byte("package main\n"), 0o644); err != nil {
		t.Fatalf("write live main.go: %v", err)
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

func TestResolveLegacyPythonEntrypointPrefersConfiguredSourceOverUnrelatedCwdScript(t *testing.T) {
	envRoot := t.TempDir()
	if err := os.MkdirAll(filepath.Join(envRoot, "cmd", "jini"), 0o755); err != nil {
		t.Fatalf("mkdir env cmd/jini: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(envRoot, "tools"), 0o755); err != nil {
		t.Fatalf("mkdir env tools: %v", err)
	}
	if err := os.WriteFile(filepath.Join(envRoot, "go.mod"), []byte("module github.com/maridlabsai/jini\n"), 0o644); err != nil {
		t.Fatalf("write env go.mod: %v", err)
	}
	if err := os.WriteFile(filepath.Join(envRoot, "cmd", "jini", "main.go"), []byte("package main\n"), 0o644); err != nil {
		t.Fatalf("write env main.go: %v", err)
	}
	if err := os.WriteFile(filepath.Join(envRoot, "tools", "jini_validate.py"), []byte("print('env')\n"), 0o755); err != nil {
		t.Fatalf("write env legacy script: %v", err)
	}

	unrelatedRoot := t.TempDir()
	if err := os.MkdirAll(filepath.Join(unrelatedRoot, "tools"), 0o755); err != nil {
		t.Fatalf("mkdir unrelated tools: %v", err)
	}
	if err := os.WriteFile(filepath.Join(unrelatedRoot, "tools", "jini_validate.py"), []byte("print('unrelated')\n"), 0o755); err != nil {
		t.Fatalf("write unrelated legacy script: %v", err)
	}

	previousDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(unrelatedRoot); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	defer func() {
		_ = os.Chdir(previousDir)
	}()

	t.Setenv("JINI_SOURCE_DIR", envRoot)

	sourceRoot, scriptPath, ok := resolveLegacyPythonEntrypoint()
	if !ok {
		t.Fatalf("expected to resolve legacy Python entrypoint")
	}
	if sourceRoot != envRoot {
		t.Fatalf("expected configured source root %q to beat unrelated cwd script, got %q", envRoot, sourceRoot)
	}
	expectedScriptPath := filepath.Join(envRoot, "tools", "jini_validate.py")
	if scriptPath != expectedScriptPath {
		t.Fatalf("expected script path %q, got %q", expectedScriptPath, scriptPath)
	}
}

func TestSelectLegacyPythonEntrypointPrefersExecutableRootOverUnrelatedCwdScript(t *testing.T) {
	cwdRoot := "/tmp/unrelated"
	cwdScript := filepath.Join(cwdRoot, "tools", "jini_validate.py")
	execRoot := "/tmp/installed-jini"
	execScript := filepath.Join(execRoot, "tools", "jini_validate.py")

	sourceRoot, scriptPath, ok := selectLegacyPythonEntrypoint(
		cwdRoot, cwdScript, true,
		"", "", false,
		execRoot, execScript, true,
	)
	if !ok {
		t.Fatalf("expected to resolve legacy Python entrypoint")
	}
	if sourceRoot != execRoot {
		t.Fatalf("expected executable-root source %q to beat unrelated cwd script, got %q", execRoot, sourceRoot)
	}
	if scriptPath != execScript {
		t.Fatalf("expected executable-root script %q, got %q", execScript, scriptPath)
	}
}

func TestSelectLegacyPythonEntrypointIgnoresUnrecognizedConfiguredSource(t *testing.T) {
	envRoot := "/tmp/unrecognized-env"
	envScript := filepath.Join(envRoot, "tools", "jini_validate.py")
	execRoot := "/tmp/installed-jini"
	execScript := filepath.Join(execRoot, "tools", "jini_validate.py")

	sourceRoot, scriptPath, ok := selectLegacyPythonEntrypoint(
		"", "", false,
		envRoot, envScript, true,
		execRoot, execScript, true,
	)
	if !ok {
		t.Fatalf("expected to resolve legacy Python entrypoint")
	}
	if sourceRoot != execRoot {
		t.Fatalf("expected executable-root source %q to beat unrecognized configured source %q, got %q", execRoot, envRoot, sourceRoot)
	}
	if scriptPath != execScript {
		t.Fatalf("expected executable-root script %q, got %q", execScript, scriptPath)
	}
}

func TestIsRecognizedJiniSourceRootRejectsWrongModuleDeclaration(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "cmd", "jini"), 0o755); err != nil {
		t.Fatalf("mkdir cmd/jini: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(root, "tools"), 0o755); err != nil {
		t.Fatalf("mkdir tools: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module example.com/jini\n"), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "cmd", "jini", "main.go"), []byte("package main\n"), 0o644); err != nil {
		t.Fatalf("write main.go: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "tools", "jini_validate.py"), []byte("print('fake')\n"), 0o755); err != nil {
		t.Fatalf("write legacy script: %v", err)
	}

	if isRecognizedJiniSourceRoot(root) {
		t.Fatalf("expected wrong module declaration to be rejected")
	}
}

func TestIsRecognizedJiniSourceRootRejectsSpoofedModuleSubstring(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "cmd", "jini"), 0o755); err != nil {
		t.Fatalf("mkdir cmd/jini: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(root, "tools"), 0o755); err != nil {
		t.Fatalf("mkdir tools: %v", err)
	}
	goMod := "module example.com/not-jini\nrequire github.com/maridlabsai/jini v0.0.0\n"
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte(goMod), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "cmd", "jini", "main.go"), []byte("package main\n"), 0o644); err != nil {
		t.Fatalf("write main.go: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "tools", "jini_validate.py"), []byte("print('fake')\n"), 0o755); err != nil {
		t.Fatalf("write legacy script: %v", err)
	}

	if isRecognizedJiniSourceRoot(root) {
		t.Fatalf("expected spoofed go.mod substring to be rejected")
	}
}

func TestIsRecognizedJiniSourceRootAcceptsWhitespaceSeparatedModuleDeclaration(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "cmd", "jini"), 0o755); err != nil {
		t.Fatalf("mkdir cmd/jini: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(root, "tools"), 0o755); err != nil {
		t.Fatalf("mkdir tools: %v", err)
	}
	goMod := "\n\tmodule\tgithub.com/maridlabsai/jini\t// canonical repo\n"
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte(goMod), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "cmd", "jini", "main.go"), []byte("package main\n"), 0o644); err != nil {
		t.Fatalf("write main.go: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "tools", "jini_validate.py"), []byte("print('real')\n"), 0o755); err != nil {
		t.Fatalf("write legacy script: %v", err)
	}

	if !isRecognizedJiniSourceRoot(root) {
		t.Fatalf("expected valid whitespace-separated module declaration to be accepted")
	}
}

func TestIsRecognizedJiniSourceRootAcceptsLeadingBlockComments(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "cmd", "jini"), 0o755); err != nil {
		t.Fatalf("mkdir cmd/jini: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(root, "tools"), 0o755); err != nil {
		t.Fatalf("mkdir tools: %v", err)
	}
	goMod := "/* generated install metadata */\n/* keep this in sync */ module github.com/maridlabsai/jini\n"
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte(goMod), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "cmd", "jini", "main.go"), []byte("package main\n"), 0o644); err != nil {
		t.Fatalf("write main.go: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "tools", "jini_validate.py"), []byte("print('real')\n"), 0o755); err != nil {
		t.Fatalf("write legacy script: %v", err)
	}

	if !isRecognizedJiniSourceRoot(root) {
		t.Fatalf("expected valid module declaration after block comments to be accepted")
	}
}

func TestIsRecognizedJiniSourceRootAcceptsInlineBlockCommentAfterModulePath(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "cmd", "jini"), 0o755); err != nil {
		t.Fatalf("mkdir cmd/jini: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(root, "tools"), 0o755); err != nil {
		t.Fatalf("mkdir tools: %v", err)
	}
	goMod := "module github.com/maridlabsai/jini /* canonical repo */\n"
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte(goMod), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "cmd", "jini", "main.go"), []byte("package main\n"), 0o644); err != nil {
		t.Fatalf("write main.go: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "tools", "jini_validate.py"), []byte("print('real')\n"), 0o755); err != nil {
		t.Fatalf("write legacy script: %v", err)
	}

	if !isRecognizedJiniSourceRoot(root) {
		t.Fatalf("expected valid inline block comment after module path to be accepted")
	}
}

func TestIsRecognizedJiniSourceRootRejectsExtraModuleTokensBeforeComment(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "cmd", "jini"), 0o755); err != nil {
		t.Fatalf("mkdir cmd/jini: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(root, "tools"), 0o755); err != nil {
		t.Fatalf("mkdir tools: %v", err)
	}
	goMod := "module github.com/maridlabsai/jini invalid // comment\n"
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte(goMod), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "cmd", "jini", "main.go"), []byte("package main\n"), 0o644); err != nil {
		t.Fatalf("write main.go: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "tools", "jini_validate.py"), []byte("print('real')\n"), 0o755); err != nil {
		t.Fatalf("write legacy script: %v", err)
	}

	if isRecognizedJiniSourceRoot(root) {
		t.Fatalf("expected extra module tokens before comment to be rejected")
	}
}

func TestSelectLegacyPythonEntrypointRejectsUnrecognizedCurrentWorkingDirectoryScript(t *testing.T) {
	cwdRoot := "/tmp/unrelated"
	cwdScript := filepath.Join(cwdRoot, "tools", "jini_validate.py")

	sourceRoot, scriptPath, ok := selectLegacyPythonEntrypoint(
		cwdRoot, cwdScript, true,
		"", "", false,
		"", "", false,
	)
	if ok {
		t.Fatalf("expected unrecognized cwd script to be rejected, got source=%q script=%q", sourceRoot, scriptPath)
	}
}

func TestFindExecutableLegacyPythonEntrypointUsesStagedSourceRuntime(t *testing.T) {
	installRoot := t.TempDir()
	stagedSourceRoot := filepath.Join(installRoot, "source-runtime")
	if err := os.MkdirAll(filepath.Join(stagedSourceRoot, "cmd", "jini"), 0o755); err != nil {
		t.Fatalf("mkdir staged cmd/jini: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(stagedSourceRoot, "tools"), 0o755); err != nil {
		t.Fatalf("mkdir staged tools: %v", err)
	}
	if err := os.WriteFile(filepath.Join(stagedSourceRoot, "go.mod"), []byte("module github.com/maridlabsai/jini\n"), 0o644); err != nil {
		t.Fatalf("write staged go.mod: %v", err)
	}
	if err := os.WriteFile(filepath.Join(stagedSourceRoot, "cmd", "jini", "main.go"), []byte("package main\n"), 0o644); err != nil {
		t.Fatalf("write staged main.go: %v", err)
	}
	expectedScriptPath := filepath.Join(stagedSourceRoot, "tools", "jini_validate.py")
	if err := os.WriteFile(expectedScriptPath, []byte("print('staged')\n"), 0o755); err != nil {
		t.Fatalf("write staged legacy script: %v", err)
	}

	sourceRoot, scriptPath, ok := findExecutableLegacyPythonEntrypoint(installRoot)
	if !ok {
		t.Fatalf("expected staged source runtime to resolve as legacy Python entrypoint")
	}
	if sourceRoot != stagedSourceRoot {
		t.Fatalf("expected staged source root %q, got %q", stagedSourceRoot, sourceRoot)
	}
	if scriptPath != expectedScriptPath {
		t.Fatalf("expected staged script path %q, got %q", expectedScriptPath, scriptPath)
	}
}

func TestFindExecutableLegacyPythonEntrypointRejectsAncestorScriptHijack(t *testing.T) {
	ancestorRoot := t.TempDir()
	installRoot := filepath.Join(ancestorRoot, "nested", "install")
	if err := os.MkdirAll(filepath.Join(ancestorRoot, "tools"), 0o755); err != nil {
		t.Fatalf("mkdir ancestor tools: %v", err)
	}
	if err := os.WriteFile(filepath.Join(ancestorRoot, "tools", "jini_validate.py"), []byte("print('ancestor')\n"), 0o755); err != nil {
		t.Fatalf("write ancestor legacy script: %v", err)
	}
	if err := os.MkdirAll(installRoot, 0o755); err != nil {
		t.Fatalf("mkdir install root: %v", err)
	}

	sourceRoot, scriptPath, ok := findExecutableLegacyPythonEntrypoint(installRoot)
	if ok {
		t.Fatalf("expected ancestor script hijack to be rejected, got source=%q script=%q", sourceRoot, scriptPath)
	}
}

func TestFindExecutableLegacyPythonEntrypointRejectsUnrecognizedDirectRootScript(t *testing.T) {
	installRoot := t.TempDir()
	if err := os.MkdirAll(filepath.Join(installRoot, "tools"), 0o755); err != nil {
		t.Fatalf("mkdir install tools: %v", err)
	}
	if err := os.WriteFile(filepath.Join(installRoot, "tools", "jini_validate.py"), []byte("print('direct-root')\n"), 0o755); err != nil {
		t.Fatalf("write direct-root legacy script: %v", err)
	}

	sourceRoot, scriptPath, ok := findExecutableLegacyPythonEntrypoint(installRoot)
	if ok {
		t.Fatalf("expected unrecognized direct-root script to be rejected, got source=%q script=%q", sourceRoot, scriptPath)
	}
}

func TestRunLegacyPythonZeroArgFailureDoesNotPanic(t *testing.T) {
	root := createRecognizedLegacySourceRoot(t, "print('ok')\n")

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
	script := "print('legacy-ok')\n"
	root := createRecognizedLegacySourceRoot(t, script)

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

func TestRunLegacyPythonUsesConfiguredInterpreterWhenTempDirIsInvalid(t *testing.T) {
	script := "print('legacy-tempdir-ok')\n"
	root := createRecognizedLegacySourceRoot(t, script)

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
	t.Setenv("TMPDIR", filepath.Join(t.TempDir(), "missing-temp-root"))

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := runLegacyPython(nil, nil, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d\nstdout:\n%s\nstderr:\n%s", exitCode, stdout.String(), stderr.String())
	}
	if !strings.Contains(stdout.String(), "legacy-tempdir-ok") {
		t.Fatalf("expected configured interpreter to work without temp dir, got:\n%s", stdout.String())
	}
}

func TestRunLegacyPythonFallsBackWhenConfiguredInterpreterIsNotExecutable(t *testing.T) {
	script := "print('legacy-fallback-ok')\n"
	root := createRecognizedLegacySourceRoot(t, script)

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
	script := "print('legacy-python-fallback-ok')\n"
	root := createRecognizedLegacySourceRoot(t, script)

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
	script := "print('legacy-spoof-fallback-ok')\n"
	root := createRecognizedLegacySourceRoot(t, script)

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
	script := "import os\nprint(os.environ['JINI_LEGACY_PYTHON'])\n"
	root := createRecognizedLegacySourceRoot(t, script)

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
	script := "import json, sys\nprint(json.dumps({'argv': sys.argv[1:]}))\n"
	root := createRecognizedLegacySourceRoot(t, script)

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

	for _, args := range [][]string{
		{"-h", "extra"},
		{"status", "/tmp/work"},
		{"check", "--help"},
		{"init", "extra"},
		{"memory", "extra"},
		{"new", "extra"},
		{"open", "--print-path"},
		{"permissions", "extra"},
		{"route", "extra"},
		{"help", "admin", "extra"},
		{"help", "--commands"},
		{"admin", "h"},
		{"help", "--all", "extra"},
		{"admin", "--h"},
		{"--doctor"},
		{"observe", "help"},
		{"observe", "--scan"},
		{"observe", "status", "extra"},
		{"observe", "scan", "extra"},
		{"observe", "add"},
	} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			var stdout bytes.Buffer
			var stderr bytes.Buffer
			exitCode := Run(args, &stdout, &stderr)
			if exitCode != 0 {
				t.Fatalf("expected exit code 0, got %d\nstdout:\n%s\nstderr:\n%s", exitCode, stdout.String(), stderr.String())
			}
			expected := fmt.Sprintf(`"argv": ["%s"`, args[0])
			if !strings.Contains(stdout.String(), expected) {
				t.Fatalf("expected legacy argv passthrough for %v, got:\n%s", args, stdout.String())
			}
		})
	}
}

func TestShouldUseLegacySurfaceBoundaryMatrix(t *testing.T) {
	for _, tc := range loadBoundaryContractCases(t) {
		t.Run(tc.Name, func(t *testing.T) {
			got := shouldUseLegacySurface(tc.Args)
			if got != tc.LegacyFallback {
				t.Fatalf("shouldUseLegacySurface(%v) = %v, want %v", tc.Args, got, tc.LegacyFallback)
			}
		})
	}
}

type boundaryContractCase struct {
	Name           string   `json:"name"`
	Args           []string `json:"args"`
	LauncherGo     bool     `json:"launcher_go"`
	LegacyFallback bool     `json:"legacy_fallback"`
}

func loadBoundaryContractCases(t *testing.T) []boundaryContractCase {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatalf("resolve caller path")
	}
	fixturePath := filepath.Join(filepath.Dir(filename), "..", "..", "tests", "fixtures", "go_boundary_contract.json")
	data, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatalf("read boundary fixture: %v", err)
	}
	var cases []boundaryContractCase
	if err := json.Unmarshal(data, &cases); err != nil {
		t.Fatalf("decode boundary fixture: %v", err)
	}
	return cases
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
	root := createRecognizedLegacySourceRoot(t, "print('legacy')\n")

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
	root := createRecognizedLegacySourceRoot(t, "print('legacy')\n")

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
	root := createRecognizedLegacySourceRoot(t, "print('legacy-front-door')\n")

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

func TestRunLauncherUsesConfiguredLegacyPythonWhenTempDirIsInvalid(t *testing.T) {
	root := createRecognizedLegacySourceRoot(t, "print('legacy-front-door-tempdir')\n")

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
	t.Setenv("TMPDIR", filepath.Join(t.TempDir(), "missing-temp-root"))

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := runLauncher(nil, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d\nstdout:\n%s\nstderr:\n%s", exitCode, stdout.String(), stderr.String())
	}
	if !strings.Contains(stdout.String(), "legacy-front-door-tempdir") {
		t.Fatalf("expected configured legacy python to keep front door alive without temp dir, got:\n%s", stdout.String())
	}
}

func TestRunLauncherProbesConfiguredLegacyPythonOnlyOnceBeforeExecution(t *testing.T) {
	root := createRecognizedLegacySourceRoot(t, "print('legacy-front-door-once')\n")

	pythonPath := resolvePythonExecutableForTest(t)
	countFile := filepath.Join(t.TempDir(), "python-count.txt")
	wrapperPath := filepath.Join(t.TempDir(), "python-wrapper")
	wrapper := "#!/bin/sh\n" +
		"count_file=\"" + countFile + "\"\n" +
		"count=0\n" +
		"if [ -f \"$count_file\" ]; then count=$(cat \"$count_file\"); fi\n" +
		"count=$((count + 1))\n" +
		"printf '%s' \"$count\" > \"$count_file\"\n" +
		"exec \"" + pythonPath + "\" \"$@\"\n"
	if err := os.WriteFile(wrapperPath, []byte(wrapper), 0o755); err != nil {
		t.Fatalf("write wrapper: %v", err)
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
	t.Setenv("JINI_LEGACY_PYTHON", wrapperPath)
	t.Setenv("JINI_STATE_DIR", stateDir)
	t.Setenv("PATH", "")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := runLauncher(nil, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d\nstdout:\n%s\nstderr:\n%s", exitCode, stdout.String(), stderr.String())
	}
	if !strings.Contains(stdout.String(), "legacy-front-door-once") {
		t.Fatalf("expected legacy front door output, got:\n%s", stdout.String())
	}
	countText, err := os.ReadFile(countFile)
	if err != nil {
		t.Fatalf("read count file: %v", err)
	}
	if strings.TrimSpace(string(countText)) != "1" {
		t.Fatalf("expected only one wrapper probe before real interpreter execution, got count %q", strings.TrimSpace(string(countText)))
	}
}

func TestRunLauncherKeepsResolvedLegacyEntrypointAfterPythonProbeMutatesRoot(t *testing.T) {
	root := createRecognizedLegacySourceRoot(t, "print('legacy-front-door-still-runs')\n")
	pythonPath := resolvePythonExecutableForTest(t)
	goModPath := filepath.Join(root, "go.mod")
	wrapperPath := filepath.Join(t.TempDir(), "python-wrapper")
	wrapper := "#!/bin/sh\n" +
		"rm -f \"" + goModPath + "\"\n" +
		"exec \"" + pythonPath + "\" \"$@\"\n"
	if err := os.WriteFile(wrapperPath, []byte(wrapper), 0o755); err != nil {
		t.Fatalf("write wrapper: %v", err)
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
	t.Setenv("JINI_LEGACY_PYTHON", wrapperPath)
	t.Setenv("JINI_STATE_DIR", stateDir)
	t.Setenv("PATH", "")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := runLauncher(nil, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d\nstdout:\n%s\nstderr:\n%s", exitCode, stdout.String(), stderr.String())
	}
	if !strings.Contains(stdout.String(), "legacy-front-door-still-runs") {
		t.Fatalf("expected cached legacy front door to survive probe-time root mutation, got:\n%s", stdout.String())
	}
	if strings.Contains(stdout.String(), "Paste what you want finished.") {
		t.Fatalf("expected not to fall back to Go prompt after probe-time mutation, got:\n%s", stdout.String())
	}
}

func TestRunLauncherFallsBackToGoPromptWhenPythonProbeMutatesLegacyScript(t *testing.T) {
	root := createRecognizedLegacySourceRoot(t, "print('legacy-front-door-gone')\n")
	pythonPath := resolvePythonExecutableForTest(t)
	scriptPath := filepath.Join(root, "tools", "jini_validate.py")
	markerPath := filepath.Join(t.TempDir(), "probe-marker")
	wrapperPath := filepath.Join(t.TempDir(), "python-wrapper")
	wrapper := "#!/bin/sh\n" +
		"printf 'used' > \"" + markerPath + "\"\n" +
		"printf 'print(\\\"mutated-front-door\\\")\\n' > \"" + scriptPath + "\"\n" +
		"exec \"" + pythonPath + "\" \"$@\"\n"
	if err := os.WriteFile(wrapperPath, []byte(wrapper), 0o755); err != nil {
		t.Fatalf("write wrapper: %v", err)
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
	t.Setenv("JINI_LEGACY_PYTHON", wrapperPath)
	t.Setenv("JINI_STATE_DIR", stateDir)
	t.Setenv("PATH", "")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := runLauncher(nil, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d\nstdout:\n%s\nstderr:\n%s", exitCode, stdout.String(), stderr.String())
	}
	if !strings.Contains(stdout.String(), "Paste what you want finished.") {
		t.Fatalf("expected Go prompt fallback after probe-time script deletion, got:\n%s", stdout.String())
	}
	if strings.Contains(stdout.String(), "legacy-front-door-gone") {
		t.Fatalf("expected not to execute original legacy script after probe-time mutation, got:\n%s", stdout.String())
	}
	if strings.Contains(stdout.String(), "mutated-front-door") {
		t.Fatalf("expected not to execute mutated legacy script after probe-time mutation, got:\n%s", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected no stderr output, got:\n%s", stderr.String())
	}
	if _, err := os.Stat(markerPath); err != nil {
		t.Fatalf("expected probe wrapper to run and create marker: %v", err)
	}
	contents, err := os.ReadFile(scriptPath)
	if err != nil {
		t.Fatalf("expected mutated legacy script to remain readable for fingerprint mismatch path: %v", err)
	}
	if !strings.Contains(string(contents), "mutated-front-door") {
		t.Fatalf("expected probe wrapper to mutate legacy script contents, got:\n%s", string(contents))
	}
}

func TestRunLauncherFallsBackToGoPromptWhenConfiguredLegacyPythonIsNotExecutable(t *testing.T) {
	root := createRecognizedLegacySourceRoot(t, "print('legacy')\n")

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

func TestAdminShortHelpDoesNotRequireLegacyFallback(t *testing.T) {
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

	t.Setenv("JINI_SOURCE_DIR", "")
	t.Setenv("PATH", "")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := Run([]string{"admin", "-h"}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d\nstdout:\n%s\nstderr:\n%s", exitCode, stdout.String(), stderr.String())
	}
	if !strings.Contains(stdout.String(), "Admin and developer command inventory") {
		t.Fatalf("expected admin help inventory, got:\n%s", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected empty stderr, got:\n%s", stderr.String())
	}
}

func TestProviderSurfacesDoNotRequireLegacyFallback(t *testing.T) {
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

	t.Setenv("JINI_SOURCE_DIR", "")
	t.Setenv("PATH", "")

	cases := []struct {
		args []string
		want string
	}{
		{args: []string{"provider"}, want: "Provider"},
		{args: []string{"provider", "doctor"}, want: "Provider"},
		{args: []string{"provider", "--format", "json"}, want: `"result_type": "JiniProviderDoctor"`},
		{args: []string{"provider", "--format", "text"}, want: "Provider"},
		{args: []string{"provider", "--format=json"}, want: `"result_type": "JiniProviderDoctor"`},
		{args: []string{"provider", "--format=text"}, want: "Provider"},
		{args: []string{"provider", "doctor", "--format", "json"}, want: `"result_type": "JiniProviderDoctor"`},
		{args: []string{"provider", "doctor", "--format", "text"}, want: "Provider"},
		{args: []string{"provider", "doctor", "--format=json"}, want: `"result_type": "JiniProviderDoctor"`},
		{args: []string{"provider", "doctor", "--format=text"}, want: "Provider"},
	}

	for _, tc := range cases {
		t.Run(strings.Join(tc.args, " "), func(t *testing.T) {
			var stdout bytes.Buffer
			var stderr bytes.Buffer
			exitCode := Run(tc.args, &stdout, &stderr)
			if exitCode != 0 {
				t.Fatalf("expected exit code 0, got %d\nstdout:\n%s\nstderr:\n%s", exitCode, stdout.String(), stderr.String())
			}
			if !strings.Contains(stdout.String(), tc.want) {
				t.Fatalf("expected output to contain %q, got:\n%s", tc.want, stdout.String())
			}
			if stderr.Len() != 0 {
				t.Fatalf("expected empty stderr, got:\n%s", stderr.String())
			}
		})
	}
}

func TestDoctorSurfacesDoNotRequireLegacyFallback(t *testing.T) {
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

	t.Setenv("JINI_SOURCE_DIR", "")
	t.Setenv("PATH", "")

	cases := []struct {
		args []string
		want string
	}{
		{args: []string{"doctor"}, want: "Provider"},
		{args: []string{"doctor", "--format", "json"}, want: `"result_type": "JiniProviderDoctor"`},
		{args: []string{"doctor", "--format", "text"}, want: "Provider"},
		{args: []string{"doctor", "--format=json"}, want: `"result_type": "JiniProviderDoctor"`},
		{args: []string{"doctor", "--format=text"}, want: "Provider"},
	}

	for _, tc := range cases {
		t.Run(strings.Join(tc.args, " "), func(t *testing.T) {
			var stdout bytes.Buffer
			var stderr bytes.Buffer
			exitCode := Run(tc.args, &stdout, &stderr)
			if exitCode != 0 {
				t.Fatalf("expected exit code 0, got %d\nstdout:\n%s\nstderr:\n%s", exitCode, stdout.String(), stderr.String())
			}
			if !strings.Contains(stdout.String(), tc.want) {
				t.Fatalf("expected output to contain %q, got:\n%s", tc.want, stdout.String())
			}
			if stderr.Len() != 0 {
				t.Fatalf("expected empty stderr, got:\n%s", stderr.String())
			}
		})
	}
}

func TestRunNewSurfacesDoNotRequireLegacyFallback(t *testing.T) {
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

	t.Setenv("JINI_SOURCE_DIR", "")
	t.Setenv("PATH", "")

	for _, args := range [][]string{
		{"run", "new"},
		{"run", "--new"},
	} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			var stdout bytes.Buffer
			var stderr bytes.Buffer
			exitCode := Run(args, &stdout, &stderr)
			if exitCode != 0 {
				t.Fatalf("expected exit code 0, got %d\nstdout:\n%s\nstderr:\n%s", exitCode, stdout.String(), stderr.String())
			}
			for _, want := range []string{
				"Jini",
				"Paste what you want finished.",
				"Nothing will be sent yet.",
			} {
				if !strings.Contains(stdout.String(), want) {
					t.Fatalf("expected output to contain %q, got:\n%s", want, stdout.String())
				}
			}
			if stderr.Len() != 0 {
				t.Fatalf("expected empty stderr, got:\n%s", stderr.String())
			}
		})
	}
}

func TestRunNewFlagMatchesRunNew(t *testing.T) {
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

	t.Setenv("JINI_SOURCE_DIR", "")
	t.Setenv("PATH", "")

	var runNew bytes.Buffer
	runNewExit := Run([]string{"run", "new"}, &runNew, &runNew)
	if runNewExit != 0 {
		t.Fatalf("expected run new to succeed, got %d with output:\n%s", runNewExit, runNew.String())
	}

	var runNewFlag bytes.Buffer
	runNewFlagExit := Run([]string{"run", "--new"}, &runNewFlag, &runNewFlag)
	if runNewFlagExit != 0 {
		t.Fatalf("expected run --new to succeed, got %d with output:\n%s", runNewFlagExit, runNewFlag.String())
	}

	if runNew.String() != runNewFlag.String() {
		t.Fatalf("expected run --new to match run new.\nRUN NEW:\n%s\nRUN --NEW:\n%s", runNew.String(), runNewFlag.String())
	}
}

func TestRunLauncherFallsBackToGoPromptWhenConfiguredLegacyPythonIsNotPython(t *testing.T) {
	root := createRecognizedLegacySourceRoot(t, "print('legacy')\n")

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
	root := createRecognizedLegacySourceRoot(t, "print('legacy')\n")

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
