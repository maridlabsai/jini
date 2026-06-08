package app

import "strings"

func reservedCLIHandoffMode(mode string) bool {
	switch strings.TrimSpace(mode) {
	case "codex", "claude-code":
		return true
	default:
		return false
	}
}

func reservedCLIHandoffLabel(mode string) string {
	switch strings.TrimSpace(mode) {
	case "codex":
		return "Codex CLI handoff"
	case "claude-code":
		return "Claude Code CLI handoff"
	default:
		return titleCase(mode)
	}
}

func reservedCLIHandoffMissing(mode string) []string {
	label := reservedCLIHandoffLabel(mode)
	return []string{
		label + " is reserved for installed-CLI subprocess handoff.",
		"This build does not hand off to that CLI yet, so Jini will not use a provider API alias for it.",
		"Use `jini route set azure-code`, `jini route set claude-api`, or `jini route auto` until CLI handoff ships.",
	}
}
