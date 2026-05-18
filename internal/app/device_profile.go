package app

import (
	"encoding/json"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const (
	deviceProfileSchemaVersion      = "0.2.0"
	deviceCapabilityRegistryVersion = "2026-05-16.2"
)

type deviceProfile struct {
	SchemaVersion             string            `json:"schema_version"`
	ContextType               string            `json:"context_type"`
	CapturedAt                string            `json:"captured_at"`
	JiniVersion               string            `json:"jini_version"`
	CapabilityRegistryVersion string            `json:"capability_registry_version"`
	ProbeFingerprint          string            `json:"probe_fingerprint"`
	LocalEndpointSignature    string            `json:"local_endpoint_signature"`
	OS                        string            `json:"os"`
	OSVersion                 string            `json:"os_version"`
	Arch                      string            `json:"arch"`
	CPUCount                  int               `json:"cpu_count"`
	TotalMemoryBytes          uint64            `json:"total_memory_bytes"`
	TotalMemoryGB             int               `json:"total_memory_gb"`
	AcceleratorClass          string            `json:"accelerator_class"`
	LocalRuntimeClass         string            `json:"local_runtime_class"`
	DeviceClass               string            `json:"device_class"`
	HardwareProfileStates     map[string]string `json:"hardware_profile_states"`
	LocalProfileStates        map[string]string `json:"local_profile_states"`
}

const deviceProfileTTL = 7 * 24 * time.Hour

var runProbeCommand = func(name string, args ...string) (string, error) {
	output, err := exec.Command(name, args...).Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func deviceProfilePath() string {
	return filepath.Join(sessionStateRoot(), "device-profile.json")
}

func currentDeviceProfile() deviceProfile {
	if override := strings.TrimSpace(configValue("JINI_DEVICE_CLASS_OVERRIDE")); override != "" {
		return syntheticDeviceProfile(override)
	}
	if cached := loadDeviceProfile(); deviceProfileIsFresh(cached) {
		return cached
	}
	fresh := probeDeviceProfile()
	_ = saveDeviceProfile(fresh)
	return fresh
}

func syntheticDeviceProfile(deviceClass string) deviceProfile {
	profile := deviceProfile{
		SchemaVersion:             deviceProfileSchemaVersion,
		ContextType:               "JiniDeviceProfile",
		CapturedAt:                time.Now().UTC().Format(time.RFC3339),
		JiniVersion:               currentJiniVersion(),
		CapabilityRegistryVersion: deviceCapabilityRegistryVersion,
		LocalEndpointSignature:    normalizedLocalEndpointSignature(),
		OS:                        runtime.GOOS,
		OSVersion:                 "override",
		Arch:                      runtime.GOARCH,
		CPUCount:                  runtime.NumCPU(),
		AcceleratorClass:          defaultAcceleratorClass(runtime.GOOS, runtime.GOARCH),
		LocalRuntimeClass:         detectLocalRuntimeClass(),
		DeviceClass:               strings.TrimSpace(deviceClass),
	}
	profile.HardwareProfileStates = hardwareProfileStatesForDevice(profile)
	profile.LocalProfileStates = effectiveLocalProfileStatesForDevice(profile)
	profile.ProbeFingerprint = probeFingerprint(profile)
	return profile
}

func loadDeviceProfile() deviceProfile {
	data, err := os.ReadFile(deviceProfilePath())
	if err != nil {
		return deviceProfile{}
	}
	var payload deviceProfile
	if err := json.Unmarshal(data, &payload); err != nil {
		return deviceProfile{}
	}
	if payload.HardwareProfileStates == nil {
		payload.HardwareProfileStates = hardwareProfileStatesForDevice(payload)
	}
	if payload.LocalProfileStates == nil {
		payload.LocalProfileStates = effectiveLocalProfileStatesForDevice(payload)
	}
	return payload
}

func saveDeviceProfile(profile deviceProfile) error {
	if err := os.MkdirAll(sessionStateRoot(), 0o755); err != nil {
		return err
	}
	if profile.SchemaVersion == "" {
		profile.SchemaVersion = deviceProfileSchemaVersion
	}
	if profile.ContextType == "" {
		profile.ContextType = "JiniDeviceProfile"
	}
	if profile.CapturedAt == "" {
		profile.CapturedAt = time.Now().UTC().Format(time.RFC3339)
	}
	if profile.JiniVersion == "" {
		profile.JiniVersion = currentJiniVersion()
	}
	if profile.CapabilityRegistryVersion == "" {
		profile.CapabilityRegistryVersion = deviceCapabilityRegistryVersion
	}
	if profile.LocalEndpointSignature == "" {
		profile.LocalEndpointSignature = normalizedLocalEndpointSignature()
	}
	if profile.HardwareProfileStates == nil {
		profile.HardwareProfileStates = hardwareProfileStatesForDevice(profile)
	}
	if profile.LocalProfileStates == nil {
		profile.LocalProfileStates = effectiveLocalProfileStatesForDevice(profile)
	}
	if profile.ProbeFingerprint == "" {
		profile.ProbeFingerprint = probeFingerprint(profile)
	}
	data, err := json.MarshalIndent(profile, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(deviceProfilePath(), append(data, '\n'), 0o600)
}

func deviceProfileIsFresh(profile deviceProfile) bool {
	if profile.SchemaVersion != deviceProfileSchemaVersion {
		return false
	}
	if profile.ContextType != "JiniDeviceProfile" {
		return false
	}
	if strings.TrimSpace(profile.DeviceClass) == "" {
		return false
	}
	if profile.CapabilityRegistryVersion != deviceCapabilityRegistryVersion {
		return false
	}
	if profile.JiniVersion != currentJiniVersion() {
		return false
	}
	if profile.OS != runtime.GOOS || profile.Arch != runtime.GOARCH {
		return false
	}
	if strings.TrimSpace(profile.OSVersion) == "" || strings.TrimSpace(profile.AcceleratorClass) == "" || strings.TrimSpace(profile.LocalRuntimeClass) == "" {
		return false
	}
	if strings.TrimSpace(profile.ProbeFingerprint) == "" || profile.ProbeFingerprint != currentProbeFingerprint() {
		return false
	}
	if !hasExpectedLocalProfileSlots(profile.HardwareProfileStates) || !hasExpectedLocalProfileSlots(profile.LocalProfileStates) {
		return false
	}
	capturedAt := strings.TrimSpace(profile.CapturedAt)
	if capturedAt == "" {
		return false
	}
	stamp, err := time.Parse(time.RFC3339, capturedAt)
	if err != nil {
		return false
	}
	return time.Since(stamp) <= deviceProfileTTL
}

func probeDeviceProfile() deviceProfile {
	goos := runtime.GOOS
	arch := runtime.GOARCH
	memoryBytes := probeTotalMemoryBytes(goos)
	profile := deviceProfile{
		SchemaVersion:             deviceProfileSchemaVersion,
		ContextType:               "JiniDeviceProfile",
		CapturedAt:                time.Now().UTC().Format(time.RFC3339),
		JiniVersion:               currentJiniVersion(),
		CapabilityRegistryVersion: deviceCapabilityRegistryVersion,
		LocalEndpointSignature:    normalizedLocalEndpointSignature(),
		OS:                        goos,
		OSVersion:                 probeOSVersion(goos),
		Arch:                      arch,
		CPUCount:                  runtime.NumCPU(),
		TotalMemoryBytes:          memoryBytes,
		TotalMemoryGB:             bytesToRoundedGB(memoryBytes),
		AcceleratorClass:          detectAcceleratorClass(goos, arch),
		LocalRuntimeClass:         detectLocalRuntimeClass(),
	}
	profile.DeviceClass = classifyDeviceClass(profile)
	profile.HardwareProfileStates = hardwareProfileStatesForDevice(profile)
	profile.LocalProfileStates = effectiveLocalProfileStatesForDevice(profile)
	profile.ProbeFingerprint = probeFingerprint(profile)
	return profile
}

func currentJiniVersion() string {
	candidates := []string{}
	if executable, err := os.Executable(); err == nil {
		root := filepath.Dir(executable)
		candidates = append(candidates,
			filepath.Join(root, "VERSION"),
			filepath.Join(root, "..", "VERSION"),
			filepath.Join(root, "..", "..", "VERSION"),
		)
	}
	if cwd, err := os.Getwd(); err == nil {
		candidates = append(candidates,
			filepath.Join(cwd, "VERSION"),
			filepath.Join(cwd, "..", "VERSION"),
		)
	}
	for _, candidate := range candidates {
		data, err := os.ReadFile(candidate)
		if err == nil {
			value := strings.TrimSpace(string(data))
			if value != "" {
				return value
			}
		}
	}
	if value := strings.TrimSpace(os.Getenv("JINI_VERSION")); value != "" {
		return value
	}
	return "0.0.0"
}

func probeOSVersion(goos string) string {
	switch goos {
	case "darwin":
		if value, err := runProbeCommand("sw_vers", "-productVersion"); err == nil && value != "" {
			return value
		}
	case "linux":
		if data, err := os.ReadFile("/etc/os-release"); err == nil {
			for _, line := range strings.Split(string(data), "\n") {
				if value, ok := strings.CutPrefix(line, "VERSION_ID="); ok {
					return strings.Trim(value, "\"")
				}
			}
		}
		if value, err := runProbeCommand("uname", "-r"); err == nil && value != "" {
			return value
		}
	case "windows":
		if value, err := runProbeCommand("cmd", "/c", "ver"); err == nil && value != "" {
			return value
		}
	}
	return "unknown"
}

func probeTotalMemoryBytes(goos string) uint64 {
	switch goos {
	case "darwin":
		if value, err := runProbeCommand("sysctl", "-n", "hw.memsize"); err == nil {
			if parsed, err := strconv.ParseUint(strings.TrimSpace(value), 10, 64); err == nil {
				return parsed
			}
		}
	case "linux":
		if data, err := os.ReadFile("/proc/meminfo"); err == nil {
			for _, line := range strings.Split(string(data), "\n") {
				if strings.HasPrefix(line, "MemTotal:") {
					fields := strings.Fields(line)
					if len(fields) >= 2 {
						if parsed, err := strconv.ParseUint(fields[1], 10, 64); err == nil {
							return parsed * 1024
						}
					}
				}
			}
		}
	}
	return 0
}

func bytesToRoundedGB(value uint64) int {
	if value == 0 {
		return 0
	}
	const gb = uint64(1024 * 1024 * 1024)
	return int((value + (gb / 2)) / gb)
}

func detectAcceleratorClass(goos, arch string) string {
	switch goos {
	case "darwin":
		if arch == "arm64" {
			return "apple-gpu"
		}
		return "cpu-only"
	case "linux":
		if _, err := runProbeCommand("nvidia-smi", "-L"); err == nil {
			return "cuda-gpu"
		}
		if _, err := runProbeCommand("rocm-smi", "--showproductname"); err == nil {
			return "rocm-gpu"
		}
		return "cpu-only"
	case "windows":
		return "windows-unknown"
	default:
		return defaultAcceleratorClass(goos, arch)
	}
}

func defaultAcceleratorClass(goos, arch string) string {
	if goos == "darwin" && arch == "arm64" {
		return "apple-gpu"
	}
	return "cpu-only"
}

func detectLocalRuntimeClass() string {
	endpoint := strings.ToLower(strings.TrimSpace(configValue("JINI_LOCAL_SLM_ENDPOINT")))
	switch {
	case endpoint == "":
		return "not-configured"
	case strings.Contains(endpoint, "127.0.0.1") || strings.Contains(endpoint, "localhost"):
		if strings.Contains(endpoint, "11434") {
			return "ollama-openai-compatible"
		}
		return "local-openai-compatible"
	default:
		return "remote-openai-compatible"
	}
}

func normalizedLocalEndpointSignature() string {
	raw := strings.TrimSpace(configValue("JINI_LOCAL_SLM_ENDPOINT"))
	if raw == "" {
		return "not-configured"
	}
	if parsed, err := url.Parse(raw); err == nil {
		scheme := strings.ToLower(strings.TrimSpace(parsed.Scheme))
		host := strings.ToLower(strings.TrimSpace(parsed.Host))
		path := strings.TrimRight(strings.ToLower(strings.TrimSpace(parsed.Path)), "/")
		if scheme == "" && host == "" {
			return strings.ToLower(strings.TrimSpace(raw))
		}
		return strings.TrimSpace(scheme + "://" + host + path)
	}
	return strings.ToLower(raw)
}

func configuredLocalProfileSignature() string {
	parts := []string{}
	for _, env := range []string{
		"JINI_LOCAL_SLM_MODEL",
		"JINI_LOCAL_SLM_FAST_MODEL",
		"JINI_LOCAL_SLM_WORKHORSE_MODEL",
		"JINI_LOCAL_SLM_DEEP_MODEL",
		"JINI_LOCAL_SLM_MULTIMODAL_MODEL",
	} {
		value := strings.TrimSpace(configValue(env))
		if value == "" {
			parts = append(parts, env+"=missing")
			continue
		}
		parts = append(parts, env+"=set")
	}
	return strings.Join(parts, "|")
}

func currentProbeFingerprint() string {
	goos := runtime.GOOS
	arch := runtime.GOARCH
	profile := deviceProfile{
		OS:                     goos,
		OSVersion:              probeOSVersion(goos),
		Arch:                   arch,
		TotalMemoryGB:          bytesToRoundedGB(probeTotalMemoryBytes(goos)),
		AcceleratorClass:       detectAcceleratorClass(goos, arch),
		LocalRuntimeClass:      detectLocalRuntimeClass(),
		LocalEndpointSignature: normalizedLocalEndpointSignature(),
	}
	return probeFingerprint(profile)
}

func probeFingerprint(profile deviceProfile) string {
	return strings.Join([]string{
		strings.TrimSpace(profile.OS),
		strings.TrimSpace(profile.OSVersion),
		strings.TrimSpace(profile.Arch),
		strconv.Itoa(profile.TotalMemoryGB),
		strings.TrimSpace(profile.AcceleratorClass),
		strings.TrimSpace(profile.LocalRuntimeClass),
		strings.TrimSpace(profile.LocalEndpointSignature),
		configuredLocalProfileSignature(),
	}, "|")
}

func classifyDeviceClass(profile deviceProfile) string {
	ram := profile.TotalMemoryGB
	cpu := profile.CPUCount
	accel := profile.AcceleratorClass
	switch {
	case accel != "cpu-only" && ram >= 32:
		return "gpu-heavy"
	case ram >= 32 && cpu >= 8:
		return "workstation"
	case ram >= 16:
		return "laptop-strong"
	case ram >= 8:
		return "laptop-light"
	default:
		return "tiny"
	}
}

func hardwareProfileStatesForDevice(profile deviceProfile) map[string]string {
	states := map[string]string{
		"local-fast":       "available",
		"local-workhorse":  "available",
		"local-deep":       "limited",
		"local-multimodal": "limited",
	}
	switch profile.DeviceClass {
	case "tiny":
		states["local-workhorse"] = "limited"
		states["local-deep"] = "unavailable"
		states["local-multimodal"] = "unavailable"
	case "laptop-light":
		states["local-deep"] = "limited"
		states["local-multimodal"] = "unavailable"
	case "laptop-strong":
		states["local-deep"] = "limited"
		if profile.AcceleratorClass == "apple-gpu" {
			states["local-multimodal"] = "limited"
		}
	case "workstation":
		states["local-deep"] = "available"
		if profile.AcceleratorClass != "cpu-only" {
			states["local-multimodal"] = "available"
		}
	case "gpu-heavy":
		states["local-deep"] = "available"
		states["local-multimodal"] = "available"
	}
	return states
}

func effectiveLocalProfileStatesForDevice(profile deviceProfile) map[string]string {
	states := cloneProfileStateMap(profile.HardwareProfileStates)
	if len(states) == 0 {
		states = hardwareProfileStatesForDevice(profile)
	}
	if !isActuallyLocalRuntimeClass(profile.LocalRuntimeClass) {
		return unavailableLocalProfileStates()
	}
	defaultModelID, _ := resolveLocalSLMDefaultModel()
	if strings.TrimSpace(defaultModelID) == "" {
		return unavailableLocalProfileStates()
	}
	if strings.TrimSpace(configValue("JINI_LOCAL_SLM_MULTIMODAL_MODEL")) == "" {
		if states["local-multimodal"] == "available" {
			states["local-multimodal"] = "limited"
		}
	}
	return states
}

func cloneProfileStateMap(source map[string]string) map[string]string {
	out := map[string]string{}
	for _, key := range []string{"local-fast", "local-workhorse", "local-deep", "local-multimodal"} {
		out[key] = strings.TrimSpace(source[key])
	}
	return out
}

func unavailableLocalProfileStates() map[string]string {
	return map[string]string{
		"local-fast":       "unavailable",
		"local-workhorse":  "unavailable",
		"local-deep":       "unavailable",
		"local-multimodal": "unavailable",
	}
}

func hasExpectedLocalProfileSlots(states map[string]string) bool {
	if len(states) == 0 {
		return false
	}
	for _, key := range []string{"local-fast", "local-workhorse", "local-deep", "local-multimodal"} {
		if strings.TrimSpace(states[key]) == "" {
			return false
		}
	}
	return true
}

func isActuallyLocalRuntimeClass(runtimeClass string) bool {
	switch strings.TrimSpace(runtimeClass) {
	case "local-openai-compatible", "ollama-openai-compatible":
		return true
	default:
		return false
	}
}
