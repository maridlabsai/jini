package app

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCurrentDeviceProfileInvalidatesOldSchemaCache(t *testing.T) {
	t.Setenv("JINI_STATE_DIR", t.TempDir())
	t.Setenv("JINI_LOCAL_SLM_ENDPOINT", "http://127.0.0.1:11434/v1")
	t.Setenv("JINI_LOCAL_SLM_MODEL", "qwen3:8b")

	stale := deviceProfile{
		SchemaVersion:             "0.1.0",
		ContextType:               "JiniDeviceProfile",
		CapturedAt:                time.Now().UTC().Format(time.RFC3339),
		JiniVersion:               currentJiniVersion(),
		CapabilityRegistryVersion: deviceCapabilityRegistryVersion,
		ProbeFingerprint:          "stale",
		LocalEndpointSignature:    "http://127.0.0.1:11434/v1",
		OS:                        "darwin",
		OSVersion:                 "stale",
		Arch:                      "arm64",
		DeviceClass:               "tiny",
		HardwareProfileStates:     unavailableLocalProfileStates(),
		LocalProfileStates:        unavailableLocalProfileStates(),
	}
	writeTestDeviceProfile(t, stale)

	profile := currentDeviceProfile()
	if profile.SchemaVersion != deviceProfileSchemaVersion {
		t.Fatalf("expected fresh schema %q, got %q", deviceProfileSchemaVersion, profile.SchemaVersion)
	}
	if profile.ProbeFingerprint == "stale" {
		t.Fatalf("expected stale fingerprint to be replaced")
	}
}

func TestCurrentDeviceProfileInvalidatesRuntimeDrift(t *testing.T) {
	t.Setenv("JINI_STATE_DIR", t.TempDir())
	t.Setenv("JINI_LOCAL_SLM_ENDPOINT", "http://127.0.0.1:11434/v1")
	t.Setenv("JINI_LOCAL_SLM_MODEL", "qwen3:8b")

	fresh := probeDeviceProfile()
	fresh.LocalRuntimeClass = "remote-openai-compatible"
	fresh.ProbeFingerprint = "runtime-before-change"
	writeTestDeviceProfile(t, fresh)

	profile := currentDeviceProfile()
	if profile.LocalRuntimeClass == "remote-openai-compatible" {
		t.Fatalf("expected runtime drift to invalidate cached profile")
	}
}

func TestEffectiveLocalProfileStatesRequireActuallyLocalRuntime(t *testing.T) {
	t.Setenv("JINI_LOCAL_SLM_MODEL", "qwen3:8b")
	profile := deviceProfile{
		DeviceClass:           "gpu-heavy",
		LocalRuntimeClass:     "remote-openai-compatible",
		HardwareProfileStates: hardwareProfileStatesForDevice(deviceProfile{DeviceClass: "gpu-heavy", AcceleratorClass: "cuda-gpu"}),
	}
	states := effectiveLocalProfileStatesForDevice(profile)
	for key, value := range states {
		if value != "unavailable" {
			t.Fatalf("expected %s to be unavailable for remote runtime, got %q", key, value)
		}
	}
}

func writeTestDeviceProfile(t *testing.T, profile deviceProfile) {
	t.Helper()
	data, err := json.MarshalIndent(profile, "", "  ")
	if err != nil {
		t.Fatalf("marshal profile: %v", err)
	}
	path := filepath.Join(sessionStateRoot(), "device-profile.json")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir state dir: %v", err)
	}
	if err := os.WriteFile(path, append(data, '\n'), 0o600); err != nil {
		t.Fatalf("write profile: %v", err)
	}
}
