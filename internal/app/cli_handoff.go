package app

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

type cliHandoffDescriptor struct {
	Mode              string
	Label             string
	DefaultExecutable string
	ExecutableEnv     string
	ArgsEnv           string
	DefaultArgs       []string
	// Verified permission-posture args (specs/handoff-posture-design.md,
	// specs/cli-handoff-compatibility-premortem.md). Substituted for the
	// {{posture}} token in DefaultArgs (dropped when empty), and only when the
	// user has not overridden ArgsEnv. Empty Semi/Autonomous means the route is
	// not verified for that posture and degrades to plan.
	PlanArgs       []string // read-only enforcement (e.g. aider --dry-run); empty = default is already read-only
	SemiArgs       []string // acceptEdits: applies edits, no arbitrary commands
	AutonomousArgs []string // full: applies edits and runs commands
	// PostureVerified is true only when the posture args were confirmed
	// empirically (behavioral plan=read-only / semi=edits / autonomous=commands),
	// not merely doc-verified. Unverified escalation-capable routes are surfaced
	// as "experimental" so posture is never release-claimed on docs alone.
	PostureVerified bool
}

type cliHandoffCommand struct {
	Descriptor cliHandoffDescriptor
	Executable string
	Args       []string
}

type cliHandoffReceipt struct {
	ContextType  string   `json:"context_type"`
	Mode         string   `json:"mode"`
	Label        string   `json:"label"`
	Executable   string   `json:"-"`
	ArgsTemplate []string `json:"-"`
	CWD          string   `json:"cwd"`
	ExitStatus   int      `json:"exit_status"`
	DurationMS   int64    `json:"duration_ms"`
	PromptChars  int      `json:"prompt_chars"`
	StdoutChars  int      `json:"stdout_chars"`
	StderrChars  int      `json:"stderr_chars"`
	CompletedAt  string   `json:"completed_at,omitempty"`
	// SideEffects lists working-tree paths the handoff changed, captured by
	// diffing `git status` before and after the run. Paths only: never file
	// contents and never command output. Empty when the working directory is
	// not a git work tree, because then Jini cannot honestly track changes.
	SideEffects []string `json:"side_effects,omitempty"`
	// SideEffectCount is the true number of changed paths even when
	// SideEffects is truncated, so the receipt never understates the blast
	// radius.
	SideEffectCount int    `json:"side_effect_count,omitempty"`
	RollbackHint    string `json:"rollback_hint,omitempty"`
}

var cliHandoffTrustIssueForPath = defaultCLIHandoffTrustIssue

func cliHandoffDescriptorForMode(mode string) (cliHandoffDescriptor, bool) {
	switch strings.TrimSpace(mode) {
	case "codex":
		return cliHandoffDescriptor{
			Mode:              "codex",
			Label:             "Codex CLI handoff",
			DefaultExecutable: "codex",
			ExecutableEnv:     "JINI_CODEX_CLI",
			ArgsEnv:           "JINI_CODEX_ARGS",
			DefaultArgs:       []string{"exec", "{{posture}}", "{{prompt}}"},
			// doc-verified 2026-08, pending empirical --help: `codex exec`
			// defaults to a read-only sandbox; bypass gives edits+commands.
			AutonomousArgs: []string{"--dangerously-bypass-approvals-and-sandbox"},
		}, true
	case "claude-code":
		return cliHandoffDescriptor{
			Mode:              "claude-code",
			Label:             "Claude Code CLI handoff",
			DefaultExecutable: "claude",
			ExecutableEnv:     "JINI_CLAUDE_CODE_CLI",
			ArgsEnv:           "JINI_CLAUDE_CODE_ARGS",
			DefaultArgs:       []string{"--print", "{{posture}}", "{{prompt}}"},
			// Verified empirically 2026-07-31: acceptEdits applies edits with
			// no command execution; --dangerously-skip-permissions applies
			// edits and runs commands.
			SemiArgs:        []string{"--permission-mode", "acceptEdits"},
			AutonomousArgs:  []string{"--dangerously-skip-permissions"},
			PostureVerified: true, // behavioral proof 2026-07-31 + real dogfood
		}, true
	case "gemini-cli":
		return cliHandoffDescriptor{
			Mode:              "gemini-cli",
			Label:             "Gemini CLI handoff",
			DefaultExecutable: "gemini",
			ExecutableEnv:     "JINI_GEMINI_CLI",
			ArgsEnv:           "JINI_GEMINI_ARGS",
			DefaultArgs:       []string{"{{posture}}", "-p", "{{prompt}}"},
			// doc-verified 2026-08, pending empirical --help: auto_edit
			// auto-approves edit tools only; --yolo auto-approves everything.
			SemiArgs:       []string{"--approval-mode", "auto_edit"},
			AutonomousArgs: []string{"--yolo"},
		}, true
	case "aider":
		return cliHandoffDescriptor{
			Mode:              "aider",
			Label:             "Aider CLI handoff",
			DefaultExecutable: "aider",
			ExecutableEnv:     "JINI_AIDER_CLI",
			ArgsEnv:           "JINI_AIDER_ARGS",
			DefaultArgs:       []string{"{{posture}}", "--message", "{{prompt}}"},
			// doc-verified 2026-08, pending empirical --help: aider --message
			// APPLIES EDITS + AUTO-COMMITS by default, so plan must force
			// --dry-run (read-only). --yes-always auto-confirms; aider runs no
			// arbitrary shell, so semi and autonomous are the same for it.
			PlanArgs:       []string{"--dry-run"},
			SemiArgs:       []string{"--yes-always"},
			AutonomousArgs: []string{"--yes-always"},
		}, true
	case "opencode":
		return cliHandoffDescriptor{
			Mode:              "opencode",
			Label:             "OpenCode CLI handoff",
			DefaultExecutable: "opencode",
			ExecutableEnv:     "JINI_OPENCODE_CLI",
			ArgsEnv:           "JINI_OPENCODE_ARGS",
			DefaultArgs:       []string{"run", "{{posture}}", "{{prompt}}"},
			// doc-verified 2026-08, pending empirical --help: `run` defaults to
			// ask-on-edit (read-only non-interactively); --auto auto-approves.
			AutonomousArgs: []string{"--auto"},
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
		"CLI_ARGS: " + formatCLIHandoffArgs(command.Args),
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
		parsedArgs, err := parseCLIHandoffArgs(rawArgs)
		if err != nil {
			command := cliHandoffCommand{
				Descriptor: descriptor,
				Executable: executable,
			}
			return command, []string{
				descriptor.ArgsEnv + " has invalid quoting: " + err.Error() + ".",
				"Use shell-like quoted args, for example: --model \"Claude Sonnet\" {{prompt}}.",
				"Jini will not execute " + descriptor.Label + " until the custom args parse cleanly.",
				"Run `jini doctor` after fixing " + descriptor.ArgsEnv + ".",
			}
		}
		args = parsedArgs
	} else {
		// Default-args path only: fold in the resolved permission posture.
		// An explicit ArgsEnv override (handled above) means the user owns the
		// args and posture is a no-op.
		args = applyPostureArgs(descriptor, resolveHandoffPosture(descriptor))
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

func parseCLIHandoffArgs(raw string) ([]string, error) {
	var args []string
	var current strings.Builder
	var quote rune
	escaping := false
	tokenStarted := false

	for _, char := range raw {
		if escaping {
			current.WriteRune(char)
			escaping = false
			tokenStarted = true
			continue
		}
		if quote != 0 {
			if char == quote {
				quote = 0
				tokenStarted = true
				continue
			}
			if quote == '"' && char == '\\' {
				escaping = true
				tokenStarted = true
				continue
			}
			current.WriteRune(char)
			tokenStarted = true
			continue
		}
		switch {
		case char == '\\':
			escaping = true
			tokenStarted = true
		case char == '\'' || char == '"':
			quote = char
			tokenStarted = true
		case unicode.IsSpace(char):
			if tokenStarted {
				args = append(args, current.String())
				current.Reset()
				tokenStarted = false
			}
		default:
			current.WriteRune(char)
			tokenStarted = true
		}
	}
	if escaping {
		return nil, fmt.Errorf("trailing escape")
	}
	if quote != 0 {
		if quote == '\'' {
			return nil, fmt.Errorf("unterminated single quote")
		}
		return nil, fmt.Errorf("unterminated double quote")
	}
	if tokenStarted {
		args = append(args, current.String())
	}
	return args, nil
}

func cliHandoffTrustIssue(path string) string {
	return cliHandoffTrustIssueForPath(path)
}

func defaultCLIHandoffTrustIssue(path string) string {
	if runtime.GOOS != "darwin" || strings.TrimSpace(configValue("JINI_CLI_HANDOFF_SKIP_TRUST_CHECK")) != "" {
		return ""
	}
	resolved := path
	if target, err := filepath.EvalSymlinks(path); err == nil {
		resolved = target
	}
	return darwinCLIHandoffTrustIssue(
		resolved,
		func(candidate string) bool {
			return exec.Command("xattr", "-p", "com.apple.quarantine", candidate).Run() == nil
		},
		func(candidate string) bool {
			// "anchor apple generic" requires the signature to chain to an
			// Apple root, which covers both Developer ID and Apple system
			// binaries while rejecting ad-hoc signatures. A bare
			// `codesign --verify --strict` is NOT sufficient here: it accepts
			// ad-hoc signatures, and the arm64 linker ad-hoc signs every
			// native binary automatically, so any downloaded arm64 binary
			// would pass by construction.
			return exec.Command("codesign", "--verify", "--strict", "-R=anchor apple generic", candidate).Run() == nil
		},
	)
}

// darwinCLIHandoffTrustIssue decides trust for a resolved CLI executable on
// macOS. `spctl -a -t execute` is intentionally not used: it rejects every
// standalone CLI binary ("valid but does not seem to be an app"), including
// legitimately signed ones. The security intent is narrower: never silently
// execute a downstream binary that arrived through a quarantining download
// path without an identified-developer signature.
//   - Not quarantined: trusted. It did not arrive through a quarantining
//     download (locally built, or installed by a tool such as npm or a curl
//     installer). Note this is the only branch that can trust an ad-hoc
//     signature, which is what every locally built arm64 binary carries.
//   - Quarantined and identity-signed (signature chains to an Apple root —
//     Developer ID or Apple system): trusted.
//   - Quarantined otherwise (unsigned, broken, or merely ad-hoc signed):
//     untrusted, fail closed with an actionable message.
func darwinCLIHandoffTrustIssue(resolvedPath string, quarantined, identitySigned func(string) bool) string {
	if !quarantined(resolvedPath) {
		return ""
	}
	if identitySigned(resolvedPath) {
		return ""
	}
	return "macOS Gatekeeper rejected CLI executable: " + resolvedPath +
		" (quarantined download without an identified developer signature). Verify the download source, then clear it with `xattr -d com.apple.quarantine " + resolvedPath + "` or reinstall from a trusted source."
}

func runCLIHandoff(ctx context.Context, mode, prompt string) (string, *cliHandoffReceipt, error) {
	descriptor, ok := cliHandoffDescriptorForMode(mode)
	if !ok {
		return "", nil, fmt.Errorf("unknown CLI handoff route %q", mode)
	}
	command, missing := resolveCLIHandoffCommand(descriptor)
	if len(missing) > 0 {
		return "", nil, cliHandoffSetupError(providerConfig{Missing: missing})
	}
	args := cliHandoffArgsWithPrompt(command.Args, prompt)
	cmd := exec.CommandContext(ctx, command.Executable, args...)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	before, tracked := cliHandoffWorkTreeStatus()
	startedAt := time.Now()
	if err := cmd.Run(); err != nil {
		receipt := buildCLIHandoffReceipt(command, prompt, stdout.String(), stderr.String(), cmd.ProcessState, time.Since(startedAt))
		annotateCLIHandoffSideEffects(receipt, before, tracked)
		if ctxErr := ctx.Err(); ctxErr != nil {
			return "", receipt, ctxErr
		}
		sanitized := cliHandoffExecutionError(descriptor.Label, err, receipt)
		return "", receipt, classifyCLIThrottleOutput(descriptor.Label, stdout.String(), stderr.String(), sanitized)
	}
	receipt := buildCLIHandoffReceipt(command, prompt, stdout.String(), stderr.String(), cmd.ProcessState, time.Since(startedAt))
	annotateCLIHandoffSideEffects(receipt, before, tracked)
	return strings.TrimSpace(stdout.String()), receipt, nil
}

// cliHandoffSideEffectLimit caps how many paths a receipt carries. The honest
// total stays in SideEffectCount.
const cliHandoffSideEffectLimit = 12

// cliHandoffRollbackListLimit caps how many paths the rollback hint names
// inline before it falls back to a count.
const cliHandoffRollbackListLimit = 6

// cliHandoffWorkTreeStatus snapshots the working tree as path -> two-letter
// porcelain status. The second return is false when the current directory is
// not a git work tree (or git is unavailable), which means Jini cannot track
// what the handoff changed and must say nothing rather than guess.
func cliHandoffWorkTreeStatus() (map[string]string, bool) {
	if out, ok := runGitOutput("rev-parse", "--is-inside-work-tree"); !ok || strings.TrimSpace(out) != "true" {
		return nil, false
	}
	out, ok := runGitOutput("status", "--porcelain", "-uall", "-z")
	if !ok {
		return nil, false
	}
	entries := map[string]string{}
	// -z output is NUL-terminated and never quotes paths, so paths with
	// spaces or non-ASCII survive intact. Rename/copy entries emit the
	// original path as an extra field that we skip.
	fields := strings.Split(out, "\x00")
	for i := 0; i < len(fields); i++ {
		raw := fields[i]
		if len(raw) < 4 {
			continue
		}
		state := raw[:2]
		path := raw[3:]
		if state[0] == 'R' || state[0] == 'C' {
			i++
		}
		if path == "" {
			continue
		}
		entries[path] = state
	}
	return entries, true
}

// annotateCLIHandoffSideEffects records the paths whose working-tree status
// changed while the handoff ran, plus a non-destructive rollback hint.
func annotateCLIHandoffSideEffects(receipt *cliHandoffReceipt, before map[string]string, tracked bool) {
	if receipt == nil || !tracked {
		return
	}
	after, ok := cliHandoffWorkTreeStatus()
	if !ok {
		return
	}
	annotateCLIHandoffSideEffectsFor(receipt, before, after)
}

// annotateCLIHandoffSideEffectsFor is the pure half of the annotation: it takes
// the before/after status snapshots and fills in the receipt fields.
func annotateCLIHandoffSideEffectsFor(receipt *cliHandoffReceipt, before, after map[string]string) {
	changed := changedWorkTreePaths(before, after)
	if len(changed) == 0 {
		return
	}
	receipt.SideEffectCount = len(changed)
	if len(changed) > cliHandoffSideEffectLimit {
		changed = changed[:cliHandoffSideEffectLimit]
	}
	receipt.SideEffects = changed
	receipt.RollbackHint = cliHandoffRollbackHint(changed, after)
}

// changedWorkTreePaths returns the paths that are dirty after the run and were
// either clean before it or carried a different status.
func changedWorkTreePaths(before, after map[string]string) []string {
	paths := make([]string, 0, len(after))
	for path, state := range after {
		if prior, seen := before[path]; seen && prior == state {
			continue
		}
		paths = append(paths, path)
	}
	sort.Strings(paths)
	return paths
}

// cliHandoffRollbackHint describes how to undo the listed paths. It is
// advisory only: it never deletes anything and never runs a command.
func cliHandoffRollbackHint(paths []string, after map[string]string) string {
	if len(paths) == 0 {
		return ""
	}
	trackedPaths := make([]string, 0, len(paths))
	newPaths := make([]string, 0, len(paths))
	for _, path := range paths {
		if after[path] == "??" {
			newPaths = append(newPaths, path)
			continue
		}
		trackedPaths = append(trackedPaths, path)
	}
	parts := []string{"Review: git diff"}
	if len(trackedPaths) > 0 {
		if len(trackedPaths) <= cliHandoffRollbackListLimit {
			parts = append(parts, "revert tracked files with: git restore "+formatCLIHandoffArgs(trackedPaths))
		} else {
			parts = append(parts, fmt.Sprintf("revert the %d tracked files above with: git restore <files>", len(trackedPaths)))
		}
	}
	if len(newPaths) > 0 {
		if len(newPaths) <= cliHandoffRollbackListLimit {
			parts = append(parts, "new files must be removed manually: "+formatCLIHandoffArgs(newPaths))
		} else {
			parts = append(parts, fmt.Sprintf("%d new files must be removed manually", len(newPaths)))
		}
	}
	return strings.Join(parts, "; ")
}

func cliHandoffExecutionError(label string, err error, receipt *cliHandoffReceipt) error {
	if receipt == nil {
		return fmt.Errorf("%s failed: %v", label, err)
	}
	return fmt.Errorf(
		"%s failed: %v (stdout %d chars, stderr %d chars; output omitted)",
		label,
		err,
		receipt.StdoutChars,
		receipt.StderrChars,
	)
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

func formatCLIHandoffArgs(args []string) string {
	if len(args) == 0 {
		return ""
	}
	formatted := make([]string, 0, len(args))
	for _, arg := range args {
		formatted = append(formatted, quoteCLIHandoffArg(arg))
	}
	return strings.Join(formatted, " ")
}

func quoteCLIHandoffArg(arg string) string {
	if arg == "" {
		return `""`
	}
	for _, char := range arg {
		if unicode.IsSpace(char) || char == '"' || char == '\'' || char == '\\' || unicode.IsControl(char) {
			return strconv.Quote(arg)
		}
	}
	return arg
}

func buildCLIHandoffReceipt(command cliHandoffCommand, prompt, stdout, stderr string, processState *os.ProcessState, duration time.Duration) *cliHandoffReceipt {
	cwd, _ := os.Getwd()
	exitStatus := -1
	if processState != nil {
		exitStatus = processState.ExitCode()
	}
	return &cliHandoffReceipt{
		ContextType:  "JiniCLIHandoffReceipt",
		Mode:         command.Descriptor.Mode,
		Label:        command.Descriptor.Label,
		Executable:   command.Executable,
		ArgsTemplate: append([]string{}, command.Args...),
		CWD:          cwd,
		ExitStatus:   exitStatus,
		DurationMS:   duration.Milliseconds(),
		PromptChars:  utf8.RuneCountInString(prompt),
		StdoutChars:  utf8.RuneCountInString(stdout),
		StderrChars:  utf8.RuneCountInString(stderr),
		CompletedAt:  time.Now().UTC().Format(time.RFC3339),
	}
}
