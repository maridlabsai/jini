package app_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMacOSAppShellScaffoldMatchesGoSidecarContract(t *testing.T) {
	root := repoRootForMigrationTest(t)

	requiredFiles := []string{
		"apps/macos/README.md",
		"apps/macos/package.json",
		"apps/macos/index.html",
		"apps/macos/src/main.js",
		"apps/macos/src/sidecar.js",
		"apps/macos/src/styles.css",
		"apps/macos/scripts/prepare-sidecar.mjs",
		"apps/macos/src-tauri/Cargo.toml",
		"apps/macos/src-tauri/src/main.rs",
		"apps/macos/src-tauri/tauri.conf.json",
		"apps/macos/src-tauri/capabilities/default.json",
	}
	for _, rel := range requiredFiles {
		if _, err := os.Stat(filepath.Join(root, rel)); err != nil {
			t.Fatalf("macOS app shell must include %s: %v", rel, err)
		}
	}

	packageJSON := readProductPositioningFile(t, root, "apps/macos/package.json")
	for _, want := range []string{
		`"@tauri-apps/cli"`,
		`"@tauri-apps/plugin-shell"`,
		`"prepare:sidecar"`,
		`"build"`,
		`"validate"`,
	} {
		if !strings.Contains(packageJSON, want) {
			t.Fatalf("macOS package.json must preserve %q", want)
		}
	}

	config := readJSONFile(t, filepath.Join(root, "apps/macos/src-tauri/tauri.conf.json"))
	if got := stringValue(t, config["productName"]); got != "Jini" {
		t.Fatalf("expected Tauri productName Jini, got %q", got)
	}
	if got := stringValue(t, config["identifier"]); got != "ai.maridlabs.jini" {
		t.Fatalf("expected app identifier ai.maridlabs.jini, got %q", got)
	}
	bundle := objectValue(t, config["bundle"])
	externalBins := stringArrayValue(t, bundle["externalBin"])
	if !containsString(externalBins, "binaries/jini-sidecar") {
		t.Fatalf("Tauri bundle must embed the Go sidecar, got %#v", externalBins)
	}

	capabilities := readProductPositioningFile(t, root, "apps/macos/src-tauri/capabilities/default.json")
	for _, want := range []string{
		`"identifier": "shell:allow-spawn"`,
		`"identifier": "shell:allow-stdin-write"`,
		`"name": "binaries/jini-sidecar"`,
		`"sidecar": true`,
		`"app"`,
		`"serve"`,
		`"--stdio"`,
		`"--surface"`,
		`"macos"`,
	} {
		if !strings.Contains(capabilities, want) {
			t.Fatalf("Tauri capability must preserve scoped sidecar contract %q", want)
		}
	}
	for _, forbidden := range []string{
		`"shell:default"`,
		`"shell:allow-execute"`,
		`"fs:default"`,
		`"fs:allow-write"`,
		`"dialog:default"`,
	} {
		if strings.Contains(capabilities, forbidden) {
			t.Fatalf("Tauri capability must not grant broad renderer permission %q", forbidden)
		}
	}
}

func TestMacOSAppShellRendersRequiredDogfoodPanelsWithoutOwningProductLogic(t *testing.T) {
	root := repoRootForMigrationTest(t)

	mainJS := readProductPositioningFile(t, root, "apps/macos/src/main.js")
	for _, want := range []string{
		"Project",
		"Sessions",
		"Task composer",
		"Diffs",
		"Artifacts",
		"Route",
		"Approvals",
		"Diagnostics",
		"turn.submit",
		"diagnostics.export",
	} {
		if !strings.Contains(mainJS, want) {
			t.Fatalf("macOS shell must render or call %q", want)
		}
	}
	for _, forbidden := range []string{"Start/Keep", "Working Draft", "Task Snapshot", "Switch to change focus"} {
		if strings.Contains(mainJS, forbidden) {
			t.Fatalf("macOS shell must not reintroduce stale UX phrase %q", forbidden)
		}
	}

	sidecarJS := readProductPositioningFile(t, root, "apps/macos/src/sidecar.js")
	for _, want := range []string{
		`Command.sidecar("binaries/jini-sidecar"`,
		`"app"`,
		`"serve"`,
		`"--stdio"`,
		`"--surface"`,
		`"macos"`,
		`protocol_version: "macos-app-v1"`,
		`idempotency_key`,
	} {
		if !strings.Contains(sidecarJS, want) {
			t.Fatalf("macOS shell sidecar bridge must preserve %q", want)
		}
	}
	for _, forbidden := range []string{"writeTextFile", "readTextFile", "invoke("} {
		if strings.Contains(sidecarJS, forbidden) {
			t.Fatalf("renderer bridge must not bypass Go sidecar through %q", forbidden)
		}
	}
}

func readJSONFile(t *testing.T, path string) map[string]any {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("decode %s: %v", path, err)
	}
	return decoded
}

func stringValue(t *testing.T, value any) string {
	t.Helper()
	text, ok := value.(string)
	if !ok {
		t.Fatalf("expected string value, got %#v", value)
	}
	return text
}

func stringArrayValue(t *testing.T, value any) []string {
	t.Helper()
	raw, ok := value.([]any)
	if !ok {
		t.Fatalf("expected string array, got %#v", value)
	}
	values := make([]string, 0, len(raw))
	for _, item := range raw {
		values = append(values, stringValue(t, item))
	}
	return values
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
