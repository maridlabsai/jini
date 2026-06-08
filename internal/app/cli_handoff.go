package app

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

type cliHandoffDescriptor struct {
	Mode              string
	Label             string
	DefaultExecutable string
	ExecutableEnv     string
	ArgsEnv           string
	DefaultArgs       []string
}

type cliHandoffCommand struct {
	Descriptor cliHandoffDescriptor
	Executable string
	Args       []string
}

func cliHandoffDescriptorForMode(mode string) (cliHandoffDescriptor, bool) {
	switch strings.TrimSpace(mode) {
	case "codex":
		return cliHandoffDescriptor{
			Mode:              "codex",
			Label:             "Codex CLI handoff",
			DefaultExecutable: "codex",
			ExecutableEnv:     "JINI_CODEX_CLI",
			ArgsEnv:           "JINI_CODEX_ARGS",
			DefaultArgs:       []string{"exec", "{{prompt}}"},
		}, true
	case "claude-code":
		return cliHandoffDescriptor{
			Mode:              "claude-code",
			Label:             "Claude Code CLI handoff",
			DefaultExecutable: "claude",
			ExecutableEnv:     "JINI_CLAUDE_CODE_CLI",
			ArgsEnv:           "JINI_CLAUDE_CODE_ARGS",
			DefaultArgs:       []string{"--print", "{{prompt}}"},
		}, true
	case "gemini-cli":
		return cliHandoffDescriptor{
			Mode:              "gemini-cli",
			Label:             "Gemini CLI handoff",
			DefaultExecutable: "gemini",
			ExecutableEnv:     "JINI_GEMINI_CLI",
			ArgsEnv:           "JINI_GEMINI_ARGS",
			DefaultArgs:       []string{"-p", "{{prompt}}"},
		}, true
	case "aider":
		return cliHandoffDescriptor{
			Mode:              "aider",
			Label:             "Aider CLI handoff",
			DefaultExecutable: "aider",
			ExecutableEnv:     "JINI_AIDER_CLI",
			ArgsEnv:           "JINI_AIDER_ARGS",
			DefaultArgs:       []string{"--message", "{{prompt}}"},
		}, true
	case "opencode":
		return cliHandoffDescriptor{
			Mode:              "opencode",
			Label:             "OpenCode CLI handoff",
			DefaultExecutable: "opencode",
			ExecutableEnv:     "JINI_OPENCODE_CLI",
			ArgsEnv:           "JINI_OPENCODE_ARGS",
			DefaultArgs:       []string{"run", "{{prompt}}"},
		}, true
	default:
		return cliHandoffDescriptor{}, false
	}
}

func cliHandoffMode(mode string) bool {
	_, ok := cliHandoffDescriptorForMode(mode)
	return ok
}

func cliHandoffLabel(mode string) string {
	if descriptor, ok := cliHandoffDescriptorForMode(mode); ok {
		return descriptor.Label
	}
	return titleCase(mode)
}

func detectCLIHandoffProvider(mode string) providerConfig {
	descriptor, ok := cliHandoffDescriptorForMode(mode)
	if !ok {
		return providerConfig{
			ID:      mode,
			Label:   titleCase(mode),
			Status:  "needs setup",
			Missing: []string{"Unknown CLI handoff route."},
		}
	}
	command, missing := resolveCLIHandoffCommand(descriptor)
	settings := []string{
		"JINI_TOOL: " + mode + " -> " + descriptor.Label,
		"CLI_EXECUTABLE: " + firstNonEmpty(command.Executable, descriptor.DefaultExecutable),
		"CLI_ARGS: " + strings.Join(command.Args, " "),
		"CLI_HANDOFF_CONTRACT: cwd, prompt, stdout, stderr, exit status, route receipt",
	}
	if configValue("JINI_CLI_HANDOFF_SKIP_TRUST_CHECK") != "" {
		settings = append(settings, "CLI_TRUST_CHECK: skipped by JINI_CLI_HANDOFF_SKIP_TRUST_CHECK")
	}
	return providerConfig{
		ID:       mode,
		Label:    descriptor.Label,
		Status:   statusFromMissing(missing),
		Missing:  missing,
		Settings: settings,
	}
}

func buildCLIHandoffDoctorReport(mode string) providerDoctorReport {
	provider := detectCLIHandoffProvider(mode)
	settings := make([]providerDoctorField, 0, len(provider.Settings))
	for _, setting := range provider.Settings {
		name, presence, ok := strings.Cut(setting, ": ")
		if !ok {
			name = setting
			presence = "present"
		}
		settings = append(settings, providerDoctorField{Name: name, Presence: presence})
	}
	return providerDoctorReport{
		SchemaVersion: "0.1.0",
		ResultType:    "JiniProviderDoctor",
		ProviderID:    provider.ID,
		Label:         provider.Label,
		Status:        provider.Status,
		Settings:      settings,
		Secrets:       []providerDoctorField{},
		Missing:       provider.Missing,
	}
}

func resolveCLIHandoffCommand(descriptor cliHandoffDescriptor) (cliHandoffCommand, []string) {
	executable := firstNonEmpty(configValue(descriptor.ExecutableEnv), descriptor.DefaultExecutable)
	args := descriptor.DefaultArgs
	if rawArgs := strings.TrimSpace(configValue(descriptor.ArgsEnv)); rawArgs != "" {
		args = strings.Fields(rawArgs)
	}
	command := cliHandoffCommand{
		Descriptor: descriptor,
		Executable: executable,
		Args:       args,
	}
	resolved, err := exec.LookPath(executable)
	if err != nil {
		return command, []string{
			descriptor.Label + " requires an installed CLI executable: " + executable + ".",
			"Set " + descriptor.ExecutableEnv + " to the executable path if it is installed outside PATH.",
			"Jini will not use a provider API alias for " + descriptor.Label + ".",
			"Run `jini doctor` after installing or configuring the CLI.",
		}
	}
	command.Executable = resolved
	if trustIssue := cliHandoffTrustIssue(resolved); trustIssue != "" {
		return command, []string{
			trustIssue,
			"Jini will not execute " + descriptor.Label + " until the executable passes local OS trust checks.",
			"Jini will not use a provider API alias for " + descriptor.Label + ".",
			"Run `jini doctor` after reinstalling the CLI from a trusted source.",
		}
	}
	return command, nil
}

func cliHandoffTrustIssue(path string) string {
	if runtime.GOOS != "darwin" || strings.TrimSpace(configValue("JINI_CLI_HANDOFF_SKIP_TRUST_CHECK")) != "" {
		return ""
	}
	cmd := exec.Command("spctl", "-a", "-q", "-t", "execute", path)
	if err := cmd.Run(); err != nil {
		return "macOS Gatekeeper rejected CLI executable: " + path + "."
	}
	return ""
}

func runCLIHandoff(ctx context.Context, mode, prompt string) (string, error) {
	descriptor, ok := cliHandoffDescriptorForMode(mode)
	if !ok {
		return "", fmt.Errorf("unknown CLI handoff route %q", mode)
	}
	command, missing := resolveCLIHandoffCommand(descriptor)
	if len(missing) > 0 {
		return "", cliHandoffSetupError(providerConfig{Missing: missing})
	}
	args := cliHandoffArgsWithPrompt(command.Args, prompt)
	cmd := exec.CommandContext(ctx, command.Executable, args...)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		detail := strings.TrimSpace(stderr.String())
		if detail != "" {
			return "", fmt.Errorf("%s failed: %v\n%s", descriptor.Label, err, detail)
		}
		return "", fmt.Errorf("%s failed: %v", descriptor.Label, err)
	}
	return strings.TrimSpace(stdout.String()), nil
}

func cliHandoffSetupError(provider providerConfig) error {
	missing := strings.Join(provider.Missing, ", ")
	if missing == "" {
		missing = "CLI handoff configuration"
	}
	return fmt.Errorf("CLI handoff needs setup. Run `jini doctor`. Missing: %s", missing)
}

func cliHandoffArgsWithPrompt(args []string, prompt string) []string {
	out := make([]string, 0, len(args)+1)
	inserted := false
	for _, arg := range args {
		if strings.Contains(arg, "{{prompt}}") {
			out = append(out, strings.ReplaceAll(arg, "{{prompt}}", prompt))
			inserted = true
			continue
		}
		out = append(out, arg)
	}
	if !inserted {
		out = append(out, prompt)
	}
	return out
}
