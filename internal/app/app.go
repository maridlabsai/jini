package app

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"
)

type currentWork struct {
	PackDir    string `json:"pack_dir"`
	PackID     string `json:"pack_id"`
	WorkUnitID string `json:"work_unit_id"`
	Title      string `json:"title"`
	State      string `json:"state"`
	Health     string `json:"health"`
}

type workSummary struct {
	Dir                string
	PackID             string
	WorkUnitID         string
	Title              string
	WorkingWith        string
	ModelLabel         string
	ModelReason        string
	ModelFeedback      string
	EffortLevel        string
	VerificationLevel  string
	VerificationReason string
	RoutePolicy        string
	AutoMode           autoModePolicy
	RouteReason        string
	ContinuityReason   string
	CLIHandoffSummary  []string
	State              string
	Views              []catalogItem
	Exports            []catalogItem
	Details            []catalogItem
	Missing            []string
	Uncertain          []string
	Using              string
	Doing              string
	Progress           string
	NextStep           string
	SafeToDo           string
	Thread             workThread
}

type catalogItem struct {
	ID      string
	Label   string
	Path    string
	Aliases []string
}

var warmLocalRuntimeCapabilities = maybeWarmLocalRuntimeCapabilitiesAsync

func Run(args []string, stdout, stderr io.Writer) int {
	return RunInteractive(args, nil, stdout, stderr)
}

func RunInteractive(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	if stdout == nil {
		stdout = io.Discard
	}
	if stderr == nil {
		stderr = io.Discard
	}

	return safelyRunInteractive(stderr, func() int {
		if len(args) == 0 {
			return runLauncher(stdin, stdout, stderr)
		}
		if isSlashCommandInput(args[0]) {
			fmt.Fprintf(stderr, "Unknown command %q.\n", args[0])
			return 1
		}
		if isHiddenAppSidecarServeCommand(args) {
			return runAppSidecarServe(stdin, stdout, stderr)
		}
		if access, ok := commercialOnlyCommandAccess(args[0]); ok {
			renderFeatureAccessDenied(stderr, access)
			return 1
		}
		if canonicalTopLevelCommand(args[0]) == "" {
			if shouldRunDirectTaskArgs(args) {
				return runDirectTaskArgsIntake(args, stdout, stderr)
			}
			fmt.Fprintf(stderr, "Unknown command %q.\n", args[0])
			fmt.Fprintln(stderr, "Run `jini commands` to see supported commands.")
			return 1
		}
		if err := validateNativeArgs(args); err != nil {
			fmt.Fprintln(stderr, err.Error())
			if shouldShowNativeCommandHint(err) {
				fmt.Fprintln(stderr, "Run `jini commands` to see supported commands.")
			}
			return 1
		}

		switch canonicalTopLevelCommand(args[0]) {
		case "help":
			if len(args) == 1 {
				return runHelp([]string{"--all"}, stdout, stderr)
			}
			return runHelp(args[1:], stdout, stderr)
		case "commands":
			return runHelp([]string{"--all"}, stdout, stderr)
		case "admin":
			return runAdmin(args[1:], stdout, stderr)
		case "check":
			return runCheck(args[1:], stdout, stderr)
		case "status":
			return runStatus(stdout, stderr)
		case "continue":
			return runContinue(stdout, stderr)
		case "doctor":
			return runProvider(args[1:], stdout, stderr)
		case "provider":
			return runProvider(args[1:], stdout, stderr)
		case "route":
			return runRoute(args[1:], stdout, stderr)
		case "memory":
			return runMemory(args[1:], stdout, stderr)
		case "permissions":
			renderSafePermissionsStatus(stdout)
			return 0
		case "init":
			renderNoInitRequired(stdout)
			fmt.Fprintln(stdout)
			renderNewWorkLauncher(stdout)
			return 0
		case "new":
			if stdin != nil {
				return runNewWorkIntake(stdin, stdout, stderr)
			}
			renderNewWorkLauncher(stdout)
			return 0
		case "observe":
			return runObserve(args[1:], stdout, stderr)
		case "open":
			return runOpen(args[1:], stdout, stderr)
		case "run":
			if len(args) > 1 && (args[1] == "--new" || args[1] == "new") {
				if stdin != nil {
					return runNewWorkIntake(stdin, stdout, stderr)
				}
				renderNewWorkLauncher(stdout)
				return 0
			}
			return runLauncher(stdin, stdout, stderr)
		case "publish-readiness":
			return runPublishReadiness(args[1:], stdout, stderr)
		case "scorecard-gate":
			return runScorecardGate(args[1:], stdout, stderr)
		default:
			fmt.Fprintf(stderr, "Unknown command %q.\n", args[0])
			fmt.Fprintln(stderr, "Run `jini commands` to see supported commands.")
			return 1
		}
	})
}

func shouldRunDirectTaskArgs(args []string) bool {
	if len(args) == 0 {
		return false
	}
	first := strings.TrimSpace(args[0])
	if first == "" || strings.HasPrefix(first, "-") {
		return false
	}
	if strings.Contains(first, "-") && !(len(args) == 1 && strings.Contains(first, " ")) {
		return false
	}
	if isRetiredCommandName(first) {
		return false
	}
	return true
}

func isRetiredCommandName(value string) bool {
	switch exactCommandToken(value) {
	case "bind-atlassian",
		"bootstrap-home",
		"capture-approval",
		"capture-output",
		"capture-publication",
		"catalog-bundles",
		"compile-pack",
		"context",
		"expand",
		"export-issues",
		"export-tasks",
		"export-wiki",
		"get-started",
		"harnesses",
		"list-routines",
		"metrics",
		"plan-install",
		"publish-issues",
		"publish-wiki",
		"recommend-execution",
		"resolve-adapter",
		"resume",
		"review-framework",
		"run-pack",
		"show",
		"show-adapters",
		"show-atlassian",
		"show-kpis",
		"stage-framework-experiment",
		"status-pack",
		"sync-tasks",
		"try-example",
		"validate-pack":
		return true
	default:
		return false
	}
}

func validateNativeArgs(args []string) error {
	if len(args) == 0 {
		return nil
	}
	first := canonicalTopLevelCommand(args[0])
	switch first {
	case "help":
		if len(args) == 1 {
			return nil
		}
		if len(args) == 2 && canonicalHelpTopic(args[1]) != "" {
			return nil
		}
	case "commands", "init", "new", "permissions", "status", "continue":
		if len(args) == 1 {
			return nil
		}
	case "memory":
		if len(args) == 1 {
			return nil
		}
		if len(args) == 2 {
			switch exactCommandToken(args[1]) {
			case "inspect", "status", "off", "disable", "on", "enable", "forget", "clear", "revoke":
				return nil
			}
		}
	case "route":
		return nil
	case "admin":
		if len(args) == 1 || (len(args) == 2 && isAdminHelpAlias(args[1])) {
			return nil
		}
	case "check":
		if len(args) > 1 && exactCommandToken(args[1]) == "ship" {
			if _, ok := parseOptionalFormatArgs(args[2:]); ok {
				return nil
			}
		}
		if len(args) > 1 && exactCommandToken(args[1]) == "competitor-watch" {
			if _, ok := parseOptionalFormatArgs(args[2:]); ok {
				return nil
			}
		}
		if len(args) == 1 || (len(args) == 2 && !strings.HasPrefix(strings.TrimSpace(args[1]), "-")) {
			return nil
		}
	case "doctor":
		if _, ok := parseOptionalFormatArgs(args[1:]); ok {
			return nil
		}
	case "provider":
		providerArgs := args[1:]
		if len(providerArgs) == 1 && isAdminHelpAlias(providerArgs[0]) {
			return nil
		}
		if len(providerArgs) > 0 && exactCommandToken(providerArgs[0]) == "doctor" {
			providerArgs = providerArgs[1:]
		}
		if _, ok := parseOptionalFormatArgs(providerArgs); ok {
			return nil
		}
	case "observe":
		if len(args) == 1 {
			return nil
		}
		second := exactCommandToken(args[1])
		if (second == "status" || second == "scan") && len(args) == 2 {
			return nil
		}
		if second == "add" {
			if _, _, err := parseObserveAddArgs(args[2:]); err == nil {
				return nil
			}
		}
	case "open":
		if len(args) <= 1 || (len(args) == 2 && !strings.HasPrefix(strings.TrimSpace(args[1]), "-")) {
			return nil
		}
	case "run":
		if len(args) == 1 || (len(args) == 2 && (exactCommandToken(args[1]) == "new" || strings.TrimSpace(args[1]) == "--new")) {
			return nil
		}
	case "publish-readiness":
		if _, ok := parseOptionalFormatArgs(args[1:]); ok {
			return nil
		}
	case "scorecard-gate":
		if _, ok := parseOptionalFormatArgs(args[1:]); ok {
			return nil
		}
	default:
		return nil
	}
	if err := catalogRequestTailError(args); err != nil {
		return err
	}
	return nativeArgError{
		message:         fmt.Sprintf("Unsupported arguments for %q.", args[0]),
		showCommandHint: true,
	}
}

type nativeArgError struct {
	message         string
	showCommandHint bool
}

func (err nativeArgError) Error() string {
	return err.message
}

func shouldShowNativeCommandHint(err error) bool {
	var nativeErr nativeArgError
	if errors.As(err, &nativeErr) {
		return nativeErr.showCommandHint
	}
	return true
}

func catalogRequestTailError(args []string) error {
	command, inventory, tail, ok := catalogRequestTail(args)
	if !ok {
		return nil
	}
	return nativeArgError{
		message: strings.Join([]string{
			fmt.Sprintf("ERROR `%s` shows the %s; it does not take a request like %q.", command, inventory, strings.Join(tail, " ")),
			"Start with `jini` to resume active work or see the start options.",
			"If you already have current work, use `jini status` once.",
		}, "\n"),
		showCommandHint: false,
	}
}

func catalogRequestTail(args []string) (command, inventory string, tail []string, ok bool) {
	if len(args) < 2 {
		return "", "", nil, false
	}

	first := exactCommandToken(args[0])
	switch first {
	case "--help", "-h":
		return "jini " + args[0], "CLI overview", args[1:], true
	case "help":
		if canonicalHelpTopic(args[1]) != "" {
			return "", "", nil, false
		}
		return "jini help", "CLI overview", args[1:], true
	case "commands":
		return "jini commands", "public command inventory", args[1:], true
	case "admin":
		if len(args) > 2 && isAdminHelpAlias(args[1]) {
			return "jini admin " + args[1], "admin command inventory", args[2:], true
		}
	case "provider":
		if len(args) > 2 && isAdminHelpAlias(args[1]) {
			return "jini provider " + args[1], "admin command inventory", args[2:], true
		}
	}
	return "", "", nil, false
}

func safelyRunInteractive(stderr io.Writer, fn func() int) (exitCode int) {
	defer func() {
		if recovered := recover(); recovered != nil {
			fmt.Fprintf(stderr, "Jini hit an unexpected internal error and stopped safely.\n")
			exitCode = 1
		}
	}()
	return fn()
}

func runLauncher(stdin io.Reader, stdout, stderr io.Writer) int {
	current, err := loadCurrentWork()
	if err != nil || current == nil {
		active, activeErr := listActiveWorkSummaries(nil)
		if activeErr != nil {
			active = nil
		}
		if stdin != nil {
			return runNoCurrentLauncher(active, stdin, stdout, stderr)
		}
		renderNewWorkPrompt(stdout)
		return 0
	}

	summary, err := loadWorkSummary(current.PackDir, current)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			_ = clearCurrentWork()
			fmt.Fprintln(stdout, "Remembered work is no longer available.")
			fmt.Fprintln(stdout)
			renderNewWorkPrompt(stdout)
			return 0
		}
		fmt.Fprintln(stderr, "Could not load current work. Run `jini` to start again or pass a valid work directory.")
		renderNewWorkPrompt(stdout)
		return 0
	}

	renderCurrentWorkPrompt(stdout, summary)
	if stdin == nil {
		return 0
	}
	_ = warmLocalRuntimeCapabilities()

	session := bufio.NewScanner(stdin)
	action, ok := readOptionalInputLine(session, stdout)
	if !ok || strings.TrimSpace(action) == "" {
		return 0
	}
	fmt.Fprintln(stdout)
	return handleCurrentWorkAction(action, summary, session, stdout, stderr)
}

func runNoCurrentLauncher(active []*workSummary, stdin io.Reader, stdout, stderr io.Writer) int {
	_ = warmLocalRuntimeCapabilities()
	renderNewWorkPrompt(stdout)
	session := bufio.NewScanner(stdin)

	for {
		action, ok := readInputLine(session, stdout)
		if !ok {
			return 0
		}
		if isHelpInput(action) {
			fmt.Fprintln(stdout)
			renderNewWorkLauncher(stdout)
			fmt.Fprintln(stdout)
			continue
		}
		if isGreetingOnly(action) {
			fmt.Fprintln(stdout)
			fmt.Fprintln(stdout, "Hi.")
			fmt.Fprintln(stdout, "Describe the task when you're ready.")
			fmt.Fprintln(stdout)
			continue
		}
		if handled, exitCode := maybeHandleNewWorkUtilityIntent(action, stdout); handled {
			if exitCode != 0 {
				return exitCode
			}
			fmt.Fprintln(stdout)
			continue
		}
		if isAcknowledgementOnly(action) || isBareContinuationIntent(action) {
			fmt.Fprintln(stdout)
			renderNewWorkNoop(stdout)
			fmt.Fprintln(stdout)
			continue
		}
		if selection, err := resolveActiveWorkSelection(action, active); err == nil {
			return switchToWorkSelection(selection, stdout, stderr)
		}
		if handled, exitCode := maybeHandleProviderSetupIntent(action, session, stdout, stderr); handled {
			if exitCode != 0 {
				fmt.Fprintln(stdout)
				renderNewWorkLauncher(stdout)
				continue
			}
			fmt.Fprintln(stdout)
			renderNewWorkLauncher(stdout)
			continue
		}
		if isSlashCommandInput(action) {
			fmt.Fprintf(stderr, "Unknown command %q.\n", strings.TrimSpace(action))
			return 1
		}
		return startNewWorkFromRawInput(action, session, stdout, stderr)
	}
}

func handleCurrentWorkAction(action string, summary *workSummary, scanner *bufio.Scanner, stdout, stderr io.Writer) int {
	if handled, exitCode := maybeHandleLocalTextFileEditIntent(action, stdout, stderr); handled {
		return exitCode
	}
	if maybeHandleSimpleAnswer(action, stdout) {
		return 0
	}
	if maybeHandleCurrentWorkQuestion(action, summary, stdout) {
		return 0
	}
	if maybeHandleStandaloneQuestion(action, stdout) {
		return 0
	}
	if access, ok := commercialOnlyCommandAccess(action); ok {
		renderFeatureAccessDenied(stdout, access)
		return 1
	}
	if resolution, resolved, err := resolveActiveAskAction(summary.Dir, summary, action); resolved {
		if err != nil {
			fmt.Fprintf(stderr, "Could not save the decision: %v\n", err)
			return 1
		}
		fmt.Fprintf(stdout, "Recorded decision: %s.\n", resolution)
		return 0
	}
	switch interactiveCommandName(action) {
	case "help":
		renderCurrentWorkHelp(stdout, summary)
	case "status":
		renderCheck(stdout, summary)
	case "resume":
		if !renderFocusedContinuation(stdout, summary) {
			renderCheck(stdout, summary)
			return 0
		}
	case "continue":
		if !renderNextContinuation(stdout, summary) {
			renderCheck(stdout, summary)
			return 0
		}
	case "open":
		return runInteractiveOpenShelf(summary, scanner, stdout, stderr)
	case "context":
		renderThreadSurface(stdout, summary, &threadFocus{Kind: "context"})
	case "missing":
		renderThreadSurface(stdout, summary, &threadFocus{Kind: "missing"})
	case "expand", "make it fuller", "make this fuller", "make it longer", "expand it", "expand this":
		if !renderNextContinuation(stdout, summary) {
			renderCheck(stdout, summary)
			return 0
		}
	case "shorter":
		item, err := applyArtifactTransform(summary, "shorter")
		if err != nil {
			fmt.Fprintf(stderr, "Could not revise the artifact: %v\n", err)
			return 1
		}
		renderItem(stdout, item)
		fmt.Fprintln(stdout)
		fmt.Fprintln(stdout, "Saved a restorable version. You can say `Versions` or `Undo`.")
	case "executive":
		item, err := applyArtifactTransform(summary, "executive")
		if err != nil {
			fmt.Fprintf(stderr, "Could not revise the artifact: %v\n", err)
			return 1
		}
		renderItem(stdout, item)
		fmt.Fprintln(stdout)
		fmt.Fprintln(stdout, "Saved a restorable version. You can say `Versions` or `Undo`.")
	case "checklist":
		item, err := applyArtifactTransform(summary, "checklist")
		if err != nil {
			fmt.Fprintf(stderr, "Could not revise the artifact: %v\n", err)
			return 1
		}
		renderItem(stdout, item)
		fmt.Fprintln(stdout)
		fmt.Fprintln(stdout, "Saved a restorable version. You can say `Versions` or `Undo`.")
	case "versions":
		renderArtifactVersions(stdout, summary)
	case "undo":
		item, err := undoLastArtifactChange(summary)
		if err != nil {
			fmt.Fprintf(stderr, "Could not restore the artifact: %v\n", err)
			return 1
		}
		fmt.Fprintln(stdout, "Restored the previous version.")
		fmt.Fprintln(stdout)
		renderItem(stdout, item)
	case "upvote":
		if err := saveModelFeedback(summary.Dir, "upvoted", ""); err != nil {
			fmt.Fprintf(stderr, "Could not save model feedback: %v\n", err)
			return 1
		}
		fmt.Fprintln(stdout, "Saved model feedback: upvoted.")
	case "accept":
		if err := saveModelFeedback(summary.Dir, "accepted-as-is", currentFeedbackArtifactPath(summary)); err != nil {
			fmt.Fprintf(stderr, "Could not save artifact feedback: %v\n", err)
			return 1
		}
		recordThreadDecision(summary.Dir, summary, "Accept")
		fmt.Fprintln(stdout, "Saved artifact feedback: accept.")
	case "edit":
		if err := saveModelFeedback(summary.Dir, "needed-light-edits", currentFeedbackArtifactPath(summary)); err != nil {
			fmt.Fprintf(stderr, "Could not save artifact feedback: %v\n", err)
			return 1
		}
		recordThreadDecision(summary.Dir, summary, "Edit")
		fmt.Fprintln(stdout, "Saved artifact feedback: edit.")
	case "use":
		if err := saveArtifactOutcome(summary.Dir, "used-this", currentFeedbackArtifactPath(summary)); err != nil {
			fmt.Fprintf(stderr, "Could not save artifact outcome: %v\n", err)
			return 1
		}
		recordThreadDecision(summary.Dir, summary, "Use")
		fmt.Fprintln(stdout, "Saved artifact outcome: use.")
	case "share":
		if err := saveArtifactOutcome(summary.Dir, "shared-this", currentFeedbackArtifactPath(summary)); err != nil {
			fmt.Fprintf(stderr, "Could not save artifact outcome: %v\n", err)
			return 1
		}
		recordThreadDecision(summary.Dir, summary, "Share")
		fmt.Fprintln(stdout, "Saved artifact outcome: share.")
	case "downvote":
		if err := saveModelFeedback(summary.Dir, "downvoted", ""); err != nil {
			fmt.Fprintf(stderr, "Could not save model feedback: %v\n", err)
			return 1
		}
		fmt.Fprintln(stdout, "Saved model feedback: downvoted.")
	case "reject":
		if err := saveModelFeedback(summary.Dir, "not-useful", currentFeedbackArtifactPath(summary)); err != nil {
			fmt.Fprintf(stderr, "Could not save artifact feedback: %v\n", err)
			return 1
		}
		recordThreadDecision(summary.Dir, summary, "Reject")
		fmt.Fprintln(stdout, "Saved artifact feedback: reject.")
	case "replace":
		if err := saveArtifactOutcome(summary.Dir, "replaced-this", currentFeedbackArtifactPath(summary)); err != nil {
			fmt.Fprintf(stderr, "Could not save artifact outcome: %v\n", err)
			return 1
		}
		recordThreadDecision(summary.Dir, summary, "Replace")
		fmt.Fprintln(stdout, "Saved artifact outcome: replace.")
	case "switch":
		return runSwitchWorkPicker(summary, scanner, stdout, stderr)
	case "doctor":
		renderInteractiveProviderDoctor(stdout)
	case "route":
		renderRouteCostStatus(stdout)
	case "route help":
		renderRouteSetupHelp(stdout)
	case "memory":
		renderUserContextMemorySummary(stdout, summary.Title)
	case "memory inspect", "memory status":
		return runMemory([]string{"inspect"}, stdout, stderr)
	case "memory off", "memory disable":
		return runMemory([]string{"off"}, stdout, stderr)
	case "memory on", "memory enable":
		return runMemory([]string{"on"}, stdout, stderr)
	case "memory forget", "memory clear", "memory revoke":
		return runMemory([]string{"forget"}, stdout, stderr)
	case "permissions":
		renderSafePermissionsStatus(stdout)
	case "init":
		renderNoInitRequired(stdout)
	case "clear":
		fmt.Fprintln(stdout, "Nothing was deleted.")
		fmt.Fprintln(stdout, "Paste a new request when ready.")
	case "plan":
		renderThreadSurface(stdout, summary, &threadFocus{Kind: "plan"})
	case "start":
		if scanner == nil {
			renderNewWorkLauncher(stdout)
			return 0
		}
		return runNewWorkIntakeWithScanner(scanner, stdout, stderr)
	default:
		if isSlashCommandInput(action) {
			fmt.Fprintf(stderr, "Unknown command %q.\n", strings.TrimSpace(action))
			return 1
		}
		if selection, err := resolveActiveWorkSelection(action, otherActiveWorkSummaries(summary)); err == nil {
			return switchToWorkSelection(selection, stdout, stderr)
		}
		if isAcknowledgementOnly(action) {
			renderCurrentWorkNoop(stdout)
			return 0
		}
		if item, ok := resolveInteractiveArtifactSelection(summary, action); ok {
			if err := renderSelectedArtifact(stdout, summary, item); err != nil {
				fmt.Fprintf(stderr, "Could not open artifact: %v\n", err)
				return 1
			}
			return 0
		}
		if scanner != nil && strings.TrimSpace(action) != "" {
			return startNewWorkFromRawInput(action, scanner, stdout, stderr)
		}
		renderCheck(stdout, summary)
	}
	return 0
}

func maybeHandleCurrentWorkQuestion(action string, summary *workSummary, stdout io.Writer) bool {
	normalized := normalizeName(action)
	kind := currentWorkQuestionKind(normalized)
	switch kind {
	case "blocked":
		fmt.Fprintln(stdout, "Blocked")
		if summary == nil || len(summary.Thread.Blocked) == 0 {
			fmt.Fprintln(stdout, "- Nothing right now")
			return true
		}
		for _, item := range summary.Thread.Blocked {
			fmt.Fprintf(stdout, "- %s\n", item)
		}
		return true
	case "missing":
		fmt.Fprintln(stdout, "Need")
		if summary == nil || strings.TrimSpace(summary.Thread.Need) == "" {
			fmt.Fprintln(stdout, "Nothing right now")
			return true
		}
		fmt.Fprintln(stdout, summary.Thread.Need)
		return true
	case "next":
		fmt.Fprintln(stdout, "Next")
		if summary == nil || strings.TrimSpace(summary.Thread.Next) == "" {
			fmt.Fprintln(stdout, "Nothing right now")
			return true
		}
		fmt.Fprintln(stdout, summary.Thread.Next)
		return true
	}
	return false
}

func currentWorkQuestionKind(normalized string) string {
	switch normalized {
	case "blocked", "blockers", "what is blocked", "whats blocked", "what are blockers", "what is blocking this", "whats blocking this", "what is blocking this work", "whats blocking this work":
		return "blocked"
	case "what is missing", "whats missing", "what do you need", "what is needed", "whats needed":
		return "missing"
	case "next", "what next", "whats next", "what is next", "what are next steps", "next steps":
		return "next"
	default:
		return ""
	}
}

type starterChoice struct {
	PackID      string
	ChoiceLabel string
	DefaultName string
	State       string
}

type workEnvelope struct {
	Choice         starterChoice
	Goal           string
	Source         string
	WorkClass      string
	RequestCohort  string
	ArtifactFamily string
}

type providerConfig struct {
	ID       string
	Label    string
	Status   string
	Missing  []string
	Settings []string
	Secrets  []string
}

func runProvider(args []string, stdout, stderr io.Writer) int {
	originalArgs := append([]string(nil), args...)
	if len(args) > 0 && isAdminHelpAlias(args[0]) {
		renderAdminCommandInventory(stdout)
		return 0
	}
	if len(args) > 0 && exactCommandToken(args[0]) == "doctor" {
		args = args[1:]
	}
	format, ok := parseOptionalFormatArgs(args)
	if !ok && len(originalArgs) > 0 && exactCommandToken(originalArgs[0]) != "doctor" {
		fmt.Fprintf(stderr, "Unknown provider command %q.\n", args[0])
		fmt.Fprintln(stderr, "Try `jini doctor`.")
		return 1
	}
	if !ok {
		fmt.Fprintln(stderr, "Unsupported doctor format. Try `jini doctor` or `jini doctor --format json`.")
		return 1
	}
	if shouldRefreshLocalBenchmarkForDoctor() && !shouldUseCachedProviderDeviceEvidence() {
		_ = currentLocalRuntimeCapabilities(context.Background())
	}
	if format == "json" {
		report := buildProviderDoctorReport()
		encoder := json.NewEncoder(stdout)
		encoder.SetIndent("", "  ")
		encoder.SetEscapeHTML(false)
		if err := encoder.Encode(report); err != nil {
			fmt.Fprintf(stderr, "Could not render provider doctor report: %v\n", err)
			return 1
		}
		if report.Status == "ok" {
			return 0
		}
		return 1
	}
	provider := detectProvider()
	renderProviderDoctor(stdout, provider)
	if provider.Status == "ok" {
		return 0
	}
	return 1
}

func runRoute(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		renderRouteCostStatus(stdout)
		return 0
	}
	switch exactCommandToken(args[0]) {
	case "list":
		renderRouteList(stdout)
		return 0
	case "help", "setup", "--help", "-h":
		renderRouteSetupHelp(stdout)
		return 0
	case "dogfood":
		return runRouteDogfood(args[1:], stdout, stderr)
	case "smoke":
		return runRouteSmoke(args[1:], stdout, stderr)
	case "validate":
		return runRouteValidate(args[1:], stdout, stderr)
	case "status":
		renderRouteCostStatus(stdout)
		return 0
	case "auto":
		return saveAutoRoute(stdout, stderr)
	case "set", "use", "lock":
		if len(args) != 2 {
			fmt.Fprintln(stderr, "Choose a route to set.")
			fmt.Fprintln(stderr, "Run `jini route list` to see available routes.")
			return 1
		}
		return saveExplicitRoute(args[1], stdout, stderr)
	default:
		fmt.Fprintf(stderr, "Unknown route command %q.\n", args[0])
		fmt.Fprintln(stderr, "Run `jini route list` to see available routes.")
		return 1
	}
}

func runRouteDogfood(args []string, stdout, stderr io.Writer) int {
	if len(args) > 0 && exactCommandToken(args[0]) == "run" {
		return runRouteDogfoodRun(args[1:], stdout, stderr)
	}
	format, ok := parseOptionalFormatArgs(args)
	if !ok {
		fmt.Fprintln(stderr, "Unsupported route dogfood format. Try `jini route dogfood`, `jini route dogfood run --claimed`, or `jini route dogfood --format json`.")
		return 1
	}
	report := buildRouteDogfoodGuide()
	if format == "json" {
		encoder := json.NewEncoder(stdout)
		encoder.SetIndent("", "  ")
		encoder.SetEscapeHTML(false)
		if err := encoder.Encode(report); err != nil {
			fmt.Fprintf(stderr, "Could not render route dogfood guide: %v\n", err)
			return 1
		}
		return 0
	}
	renderRouteDogfoodGuide(stdout, report)
	return 0
}

type routeDogfoodRunOptions struct {
	Claimed bool
	Format  string
}

type routeDogfoodRunReport struct {
	SchemaVersion string                 `json:"schema_version"`
	ResultType    string                 `json:"result_type"`
	Status        string                 `json:"status"`
	Scope         string                 `json:"scope"`
	EvidenceFile  string                 `json:"evidence_file,omitempty"`
	ConfigIssues  []string               `json:"config_issues,omitempty"`
	Routes        []routeDogfoodRunRoute `json:"routes"`
	Next          []string               `json:"next"`
}

type routeDogfoodRunRoute struct {
	RouteID       string `json:"route_id"`
	Label         string `json:"label"`
	Status        string `json:"status"`
	SetupCategory string `json:"setup_category,omitempty"`
	SmokedAt      string `json:"smoked_at,omitempty"`
	ExitStatus    int    `json:"exit_status,omitempty"`
	DurationMS    int64  `json:"duration_ms,omitempty"`
	PromptChars   int    `json:"prompt_chars,omitempty"`
	StdoutChars   int    `json:"stdout_chars,omitempty"`
	StderrChars   int    `json:"stderr_chars,omitempty"`
	Next          string `json:"next,omitempty"`
}

func runRouteDogfoodRun(args []string, stdout, stderr io.Writer) int {
	opts, message, ok := parseRouteDogfoodRunArgs(args)
	if !ok {
		fmt.Fprintln(stderr, message)
		fmt.Fprintln(stderr, "Try `jini route dogfood run --claimed`.")
		return 1
	}
	report := buildRouteDogfoodRunReport(opts)
	if opts.Format == "json" {
		encoder := json.NewEncoder(stdout)
		encoder.SetIndent("", "  ")
		encoder.SetEscapeHTML(false)
		if err := encoder.Encode(report); err != nil {
			fmt.Fprintf(stderr, "Could not render route dogfood run report: %v\n", err)
			return 1
		}
		if report.Status == "ok" {
			return 0
		}
		return 1
	}
	renderRouteDogfoodRunReport(stdout, report)
	if report.Status == "ok" {
		return 0
	}
	return 1
}

func parseRouteDogfoodRunArgs(args []string) (routeDogfoodRunOptions, string, bool) {
	var opts routeDogfoodRunOptions
	for i := 0; i < len(args); i++ {
		arg := strings.TrimSpace(args[i])
		switch {
		case arg == "--claimed":
			opts.Claimed = true
		case arg == "--format" && i+1 < len(args):
			opts.Format = normalizeName(args[i+1])
			if opts.Format != "json" && opts.Format != "text" {
				return opts, "Unsupported route dogfood run format.", false
			}
			i++
		case strings.HasPrefix(arg, "--format="):
			opts.Format = normalizeName(strings.TrimPrefix(arg, "--format="))
			if opts.Format != "json" && opts.Format != "text" {
				return opts, "Unsupported route dogfood run format.", false
			}
		default:
			return opts, "Unsupported route dogfood run option: " + arg, false
		}
	}
	if !opts.Claimed {
		return opts, "Choose `--claimed` so Jini only runs routes named in JINI_CLI_RELEASE_ROUTES.", false
	}
	return opts, "", true
}

func buildRouteDogfoodRunReport(opts routeDogfoodRunOptions) routeDogfoodRunReport {
	scope := "claimed"
	report := routeDogfoodRunReport{
		SchemaVersion: "0.1.0",
		ResultType:    "JiniRouteDogfoodRun",
		Status:        "ok",
		Scope:         scope,
		Routes:        []routeDogfoodRunRoute{},
		Next:          []string{},
	}
	claimConfig := parseCLIHandoffReleaseClaimConfig()
	if len(claimConfig.Issues) > 0 {
		report.Status = "config-error"
		report.ConfigIssues = append([]string(nil), claimConfig.Issues...)
		report.Next = []string{"Fix JINI_CLI_RELEASE_ROUTES, then rerun `jini route dogfood run --claimed`."}
		return report
	}
	if opts.Claimed && len(claimConfig.Claims) == 0 {
		report.Status = "no-claims"
		report.Next = []string{"Set JINI_CLI_RELEASE_ROUTES=codex or another CLI route, then rerun `jini route dogfood run --claimed`."}
		return report
	}
	report.EvidenceFile = ".jini/cli-smoke.json"
	for _, routeID := range cliHandoffRouteIDs() {
		if opts.Claimed && !claimConfig.Claims[routeID] {
			continue
		}
		report.Routes = append(report.Routes, runRouteDogfoodSmoke(routeID))
	}
	status := "ok"
	for _, route := range report.Routes {
		if route.Status != "smoked" {
			status = "needs-attention"
			continue
		}
		report.Next = append(report.Next, route.Next)
	}
	report.Status = status
	if len(report.Routes) == 0 {
		report.Status = "no-routes"
		report.Next = []string{"No CLI routes matched the requested scope."}
	}
	return report
}

func runRouteDogfoodSmoke(routeID string) routeDogfoodRunRoute {
	descriptor, found := cliHandoffDescriptorForMode(routeID)
	if !found {
		return routeDogfoodRunRoute{RouteID: routeID, Status: "unknown-route"}
	}
	row := routeDogfoodRunRoute{
		RouteID: routeID,
		Label:   descriptor.Label,
		Next:    "jini route validate " + routeID + " --real-cli --checks all",
	}
	if _, missing := resolveCLIHandoffCommand(descriptor); len(missing) > 0 {
		row.Status = "setup-blocked"
		row.SetupCategory = shipCLIHandoffSetupCategory(missing)
		return row
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	_, receipt, err := runCLIHandoff(ctx, routeID, routeSmokePrompt)
	if err != nil || receipt == nil {
		row.Status = "failed"
		return row
	}
	smokedAt := time.Now().UTC().Format(time.RFC3339)
	if _, err := saveCLIHandoffSmokeEvidence(routeID, receipt, smokedAt); err != nil {
		row.Status = "failed"
		return row
	}
	row.Status = "smoked"
	row.SmokedAt = smokedAt
	row.ExitStatus = receipt.ExitStatus
	row.DurationMS = receipt.DurationMS
	row.PromptChars = receipt.PromptChars
	row.StdoutChars = receipt.StdoutChars
	row.StderrChars = receipt.StderrChars
	return row
}

func renderRouteDogfoodRunReport(w io.Writer, report routeDogfoodRunReport) {
	if report.Status == "config-error" {
		fmt.Fprintln(w, "Invalid claimed CLI route configuration.")
		for _, issue := range report.ConfigIssues {
			fmt.Fprintf(w, "- %s\n", issue)
		}
		for _, next := range report.Next {
			fmt.Fprintf(w, "Next: %s\n", next)
		}
		return
	}
	if report.Status == "no-claims" {
		fmt.Fprintln(w, "No claimed CLI routes configured.")
		for _, next := range report.Next {
			fmt.Fprintf(w, "Next: %s\n", next)
		}
		return
	}
	if report.Status == "ok" {
		fmt.Fprintln(w, "CLI dogfood run complete.")
	} else {
		fmt.Fprintln(w, "CLI dogfood run needs attention.")
	}
	fmt.Fprintf(w, "- scope: %s\n", report.Scope)
	for _, route := range report.Routes {
		switch route.Status {
		case "smoked":
			fmt.Fprintf(w, "- %s: smoked; stdout %d chars, stderr %d chars\n", route.RouteID, route.StdoutChars, route.StderrChars)
		case "setup-blocked":
			fmt.Fprintf(w, "- %s: setup blocked (%s)\n", route.RouteID, firstNonEmpty(route.SetupCategory, "see doctor"))
		default:
			fmt.Fprintf(w, "- %s: %s\n", route.RouteID, route.Status)
		}
	}
	fmt.Fprintln(w, "- evidence: .jini/cli-smoke.json")
	for _, next := range report.Next {
		fmt.Fprintf(w, "Next: %s\n", next)
	}
}

type routeValidateOptions struct {
	RouteID string
	Checks  []string
	RealCLI bool
	Format  string
}

type routeValidateReport struct {
	SchemaVersion  string   `json:"schema_version"`
	ResultType     string   `json:"result_type"`
	Status         string   `json:"status"`
	RouteID        string   `json:"route_id"`
	EvidenceFile   string   `json:"evidence_file,omitempty"`
	ValidatedAt    string   `json:"validated_at,omitempty"`
	Checks         []string `json:"checks,omitempty"`
	RequiredChecks []string `json:"required_checks"`
	Next           []string `json:"next"`
}

type routeSmokeOptions struct {
	RouteID string
	Format  string
}

type routeSmokeReport struct {
	SchemaVersion string   `json:"schema_version"`
	ResultType    string   `json:"result_type"`
	Status        string   `json:"status"`
	RouteID       string   `json:"route_id"`
	EvidenceFile  string   `json:"evidence_file,omitempty"`
	SmokedAt      string   `json:"smoked_at,omitempty"`
	ExitStatus    int      `json:"exit_status"`
	DurationMS    int64    `json:"duration_ms"`
	PromptChars   int      `json:"prompt_chars"`
	StdoutChars   int      `json:"stdout_chars"`
	StderrChars   int      `json:"stderr_chars"`
	Next          []string `json:"next"`
}

func runRouteSmoke(args []string, stdout, stderr io.Writer) int {
	opts, message, ok := parseRouteSmokeArgs(args)
	if !ok {
		fmt.Fprintln(stderr, message)
		fmt.Fprintln(stderr, "Try `jini route smoke codex`.")
		return 1
	}
	descriptor, found := cliHandoffDescriptorForMode(opts.RouteID)
	if !found {
		fmt.Fprintf(stderr, "Unknown CLI handoff route %q.\n", opts.RouteID)
		fmt.Fprintf(stderr, "Valid routes: %s\n", strings.Join(cliHandoffRouteIDs(), ", "))
		return 1
	}
	if _, missing := resolveCLIHandoffCommand(descriptor); len(missing) > 0 {
		fmt.Fprintf(stderr, "Route not ready: %s (%s).\n", opts.RouteID, shipCLIHandoffSetupCategory(missing))
		fmt.Fprintln(stderr, "Run `jini route dogfood` for setup fixes.")
		return 1
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	_, receipt, err := runCLIHandoff(ctx, opts.RouteID, routeSmokePrompt)
	if err != nil {
		fmt.Fprintf(stderr, "Route smoke failed: %v\n", err)
		return 1
	}
	if receipt == nil {
		fmt.Fprintln(stderr, "Route smoke failed: missing route receipt.")
		return 1
	}
	smokedAt := time.Now().UTC().Format(time.RFC3339)
	evidenceFile, err := saveCLIHandoffSmokeEvidence(opts.RouteID, receipt, smokedAt)
	if err != nil {
		fmt.Fprintf(stderr, "Could not write CLI smoke evidence: %v\n", err)
		return 1
	}
	report := routeSmokeReport{
		SchemaVersion: "0.1.0",
		ResultType:    "JiniRouteSmoke",
		Status:        "ok",
		RouteID:       opts.RouteID,
		EvidenceFile:  evidenceFile,
		SmokedAt:      smokedAt,
		ExitStatus:    receipt.ExitStatus,
		DurationMS:    receipt.DurationMS,
		PromptChars:   receipt.PromptChars,
		StdoutChars:   receipt.StdoutChars,
		StderrChars:   receipt.StderrChars,
		Next: []string{
			"Inspect auth, approvals, output shape, and receipt privacy.",
			"Then run `jini route validate " + opts.RouteID + " --real-cli --checks all` only if the real CLI smoke passed.",
		},
	}
	if opts.Format == "json" {
		encoder := json.NewEncoder(stdout)
		encoder.SetIndent("", "  ")
		encoder.SetEscapeHTML(false)
		if err := encoder.Encode(report); err != nil {
			fmt.Fprintf(stderr, "Could not render route smoke report: %v\n", err)
			return 1
		}
		return 0
	}
	fmt.Fprintln(stdout, "Route smoke complete.")
	fmt.Fprintf(stdout, "- route: %s\n", report.RouteID)
	fmt.Fprintf(stdout, "- exit: %d\n", report.ExitStatus)
	fmt.Fprintf(stdout, "- stdout: %d chars\n", report.StdoutChars)
	fmt.Fprintf(stdout, "- stderr: %d chars\n", report.StderrChars)
	fmt.Fprintln(stdout, "- evidence: .jini/cli-smoke.json")
	fmt.Fprintln(stdout, "- receipt: prompt and output bodies omitted")
	fmt.Fprintf(stdout, "Next: jini route validate %s --real-cli --checks all\n", opts.RouteID)
	return 0
}

func parseRouteSmokeArgs(args []string) (routeSmokeOptions, string, bool) {
	var opts routeSmokeOptions
	if len(args) == 0 {
		return opts, "Choose a CLI handoff route to smoke.", false
	}
	routeID := normalizeCLIHandoffRouteID(args[0])
	if routeID == "" {
		return opts, "Choose a valid CLI handoff route to smoke.", false
	}
	opts.RouteID = routeID
	format, ok := parseOptionalFormatArgs(args[1:])
	if !ok {
		return opts, "Unsupported route smoke format.", false
	}
	opts.Format = format
	return opts, "", true
}

func runRouteValidate(args []string, stdout, stderr io.Writer) int {
	opts, message, ok := parseRouteValidateArgs(args)
	if !ok {
		fmt.Fprintln(stderr, message)
		fmt.Fprintln(stderr, "Try `jini route validate codex --real-cli --checks all`.")
		return 1
	}
	descriptor, found := cliHandoffDescriptorForMode(opts.RouteID)
	if !found {
		fmt.Fprintf(stderr, "Unknown CLI handoff route %q.\n", opts.RouteID)
		fmt.Fprintf(stderr, "Valid routes: %s\n", strings.Join(cliHandoffRouteIDs(), ", "))
		return 1
	}
	_, missing := resolveCLIHandoffCommand(descriptor)
	if len(missing) > 0 {
		fmt.Fprintf(stderr, "Route not ready: %s (%s).\n", opts.RouteID, shipCLIHandoffSetupCategory(missing))
		fmt.Fprintln(stderr, "Run `jini route dogfood` for setup fixes.")
		return 1
	}
	if !opts.RealCLI {
		fmt.Fprintln(stderr, "Refusing to write dogfood evidence without `--real-cli`.")
		fmt.Fprintln(stderr, "Use it only after the real installed CLI completed the harmless validation prompt.")
		return 1
	}
	requiredChecks := requiredCLIHandoffDogfoodChecks()
	validatedChecks, missingChecks := dogfoodCheckCoverage(requiredChecks, opts.Checks)
	if len(missingChecks) > 0 {
		fmt.Fprintf(stderr, "Missing required validation checks: %s.\n", strings.Join(missingChecks, ", "))
		fmt.Fprintln(stderr, "Use `--checks all` only after auth, approvals, output shape, and route receipt privacy were verified.")
		return 1
	}
	smokeLoad := loadCLIHandoffSmokeEvidence()
	smokeProof := smokeLoad.Routes[opts.RouteID]
	if smokeIssue := cliHandoffRecentSmokeIssue(opts.RouteID, smokeProof, time.Now().UTC()); smokeIssue != "" {
		fmt.Fprintf(stderr, "Run `jini route smoke %s` before writing dogfood evidence.\n", opts.RouteID)
		fmt.Fprintf(stderr, "Smoke prerequisite: %s.\n", smokeIssue)
		return 1
	}
	validatedAt := time.Now().UTC().Format(time.RFC3339)
	evidencePath, err := saveCLIHandoffDogfoodEvidence(opts.RouteID, validatedChecks, validatedAt)
	if err != nil {
		fmt.Fprintf(stderr, "Could not write CLI dogfood evidence: %v\n", err)
		return 1
	}
	report := routeValidateReport{
		SchemaVersion:  "0.1.0",
		ResultType:     "JiniRouteValidate",
		Status:         "validated",
		RouteID:        opts.RouteID,
		EvidenceFile:   evidencePath,
		ValidatedAt:    validatedAt,
		Checks:         validatedChecks,
		RequiredChecks: requiredChecks,
		Next:           []string{"Run `jini check ship --format json`."},
	}
	if opts.Format == "json" {
		encoder := json.NewEncoder(stdout)
		encoder.SetIndent("", "  ")
		encoder.SetEscapeHTML(false)
		if err := encoder.Encode(report); err != nil {
			fmt.Fprintf(stderr, "Could not render route validation report: %v\n", err)
			return 1
		}
		return 0
	}
	fmt.Fprintln(stdout, "Route dogfood evidence saved.")
	fmt.Fprintf(stdout, "- route: %s\n", opts.RouteID)
	fmt.Fprintf(stdout, "- checks: %s\n", strings.Join(validatedChecks, ", "))
	fmt.Fprintln(stdout, "- evidence: .jini/cli-dogfood.json")
	fmt.Fprintln(stdout, "Next: jini check ship --format json")
	return 0
}

func parseRouteValidateArgs(args []string) (routeValidateOptions, string, bool) {
	var opts routeValidateOptions
	if len(args) == 0 {
		return opts, "Choose a CLI handoff route to validate.", false
	}
	routeID := normalizeCLIHandoffRouteID(args[0])
	if routeID == "" {
		return opts, "Choose a valid CLI handoff route to validate.", false
	}
	opts.RouteID = routeID
	for i := 1; i < len(args); i++ {
		arg := strings.TrimSpace(args[i])
		switch {
		case arg == "--real-cli":
			opts.RealCLI = true
		case arg == "--format" && i+1 < len(args):
			opts.Format = normalizeName(args[i+1])
			if opts.Format != "json" && opts.Format != "text" {
				return opts, "Unsupported route validate format.", false
			}
			i++
		case strings.HasPrefix(arg, "--format="):
			opts.Format = normalizeName(strings.TrimPrefix(arg, "--format="))
			if opts.Format != "json" && opts.Format != "text" {
				return opts, "Unsupported route validate format.", false
			}
		case arg == "--checks" && i+1 < len(args):
			checks, issue := parseRouteValidateChecks(args[i+1])
			if issue != "" {
				return opts, issue, false
			}
			opts.Checks = append(opts.Checks, checks...)
			i++
		case strings.HasPrefix(arg, "--checks="):
			checks, issue := parseRouteValidateChecks(strings.TrimPrefix(arg, "--checks="))
			if issue != "" {
				return opts, issue, false
			}
			opts.Checks = append(opts.Checks, checks...)
		case arg == "--check" && i+1 < len(args):
			checks, issue := parseRouteValidateChecks(args[i+1])
			if issue != "" {
				return opts, issue, false
			}
			opts.Checks = append(opts.Checks, checks...)
			i++
		case strings.HasPrefix(arg, "--check="):
			checks, issue := parseRouteValidateChecks(strings.TrimPrefix(arg, "--check="))
			if issue != "" {
				return opts, issue, false
			}
			opts.Checks = append(opts.Checks, checks...)
		default:
			return opts, "Unsupported route validate option: " + arg, false
		}
	}
	return opts, "", true
}

func parseRouteValidateChecks(raw string) ([]string, string) {
	var checks []string
	for _, token := range strings.FieldsFunc(raw, func(r rune) bool {
		return r == ',' || r == ';' || r == '\n' || r == '\t'
	}) {
		rawToken := strings.TrimSpace(token)
		token = normalizeRouteValidateCheck(token)
		if token == "" {
			return nil, "Unknown validation check: " + rawToken + ". Valid checks: auth, approvals, output-shape, route-receipt-privacy, all."
		}
		if token == "all" {
			return requiredCLIHandoffDogfoodChecks(), ""
		}
		checks = append(checks, token)
	}
	return checks, ""
}

func normalizeRouteValidateCheck(value string) string {
	normalized := strings.ToLower(strings.TrimSpace(value))
	normalized = strings.ReplaceAll(normalized, "_", "-")
	switch normalized {
	case "all":
		return "all"
	case "auth":
		return "auth"
	case "approvals", "approval":
		return "approvals"
	case "output-shape", "output shape", "output":
		return "output shape"
	case "route-receipt-privacy", "route receipt privacy", "receipt-privacy", "privacy":
		return "route receipt privacy"
	default:
		return ""
	}
}

func saveAutoRoute(stdout, stderr io.Writer) int {
	if err := saveRouterSettings("auto"); err != nil {
		fmt.Fprintf(stderr, "Could not save auto route: %v\n", err)
		return 1
	}
	recordUserContextSignal("route_preference", "Automatic route", "auto route selection")
	fmt.Fprintln(stdout, "Auto route restored.")
	fmt.Fprintln(stdout, "Current: auto")
	fmt.Fprintln(stdout, "Jini will choose the least-expense capable route per request.")
	return 0
}

func saveExplicitRoute(raw string, stdout, stderr io.Writer) int {
	mode := normalizeToolMode(raw)
	if mode == "auto" {
		return saveAutoRoute(stdout, stderr)
	}
	if !knownRouteMode(mode) {
		fmt.Fprintf(stderr, "Unknown route %q.\n", strings.TrimSpace(raw))
		fmt.Fprintln(stderr, "Run `jini route list` to see available routes.")
		return 1
	}
	decision := detectExplicitRouteForMode(mode)
	if !decision.Active {
		fmt.Fprintf(stderr, "Unknown route %q.\n", strings.TrimSpace(raw))
		fmt.Fprintln(stderr, "Run `jini route list` to see available routes.")
		return 1
	}
	if cliHandoffMode(mode) && decision.Provider.Status != "ok" {
		fmt.Fprintf(stderr, "%s CLI handoff is not ready.\n", decision.ToolLabel)
		for _, missing := range decision.Provider.Missing {
			fmt.Fprintf(stderr, "- %s\n", missing)
		}
		return 1
	}
	if err := saveRouterSettings(mode); err != nil {
		fmt.Fprintf(stderr, "Could not save route: %v\n", err)
		return 1
	}
	recordUserContextSignal("route_preference", firstNonEmpty(decision.ToolLabel, titleCase(mode)), "explicit route selection")
	fmt.Fprintln(stdout, "Route saved.")
	fmt.Fprintf(stdout, "Current: %s\n", firstNonEmpty(decision.ToolLabel, titleCase(mode)))
	fmt.Fprintf(stdout, "Mode: %s\n", mode)
	if cliHandoffMode(mode) {
		fmt.Fprintf(stdout, "CLI handoff: %s\n", firstNonEmpty(decision.Provider.Label, decision.ToolLabel))
	}
	if strings.TrimSpace(decision.Provider.Status) != "" && decision.Provider.Status != "ok" {
		fmt.Fprintf(stdout, "Setup: %s. Run `jini doctor` to inspect missing setup.\n", decision.Provider.Status)
	}
	fmt.Fprintln(stdout, "Use `jini route auto` to restore automatic routing.")
	return 0
}

func knownRouteMode(mode string) bool {
	if mode == "auto" {
		return true
	}
	if mode == "local-slm" {
		return true
	}
	_, ok := adapterDescriptorForMode(mode)
	return ok
}

func detectExplicitRouteForMode(mode string) routeDecision {
	if mode == "local-slm" {
		return detectLocalSLMAliasRouteForRequest(providerGenerationRequest{}, false)
	}
	return detectRouteForToolMode(mode, false)
}

func renderRouteList(w io.Writer) {
	current := configuredToolMode()
	if current == "" {
		current = "auto"
	}
	fmt.Fprintln(w, "Routes")
	fmt.Fprintf(w, "Current: %s\n", current)
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Available")
	fmt.Fprintln(w, "- auto: choose the least-expense capable route per request")
	for _, target := range defaultSavedRouteTargets() {
		if !target.Enabled {
			continue
		}
		descriptor, ok := adapterDescriptorForMode(target.ID)
		cost := ""
		locality := ""
		if ok {
			cost = descriptor.CostTier
			locality = descriptor.Locality
		}
		readiness := routeTargetReadinessLabel(target)
		if detail := routeTargetReadinessDetail(target); detail != "" {
			readiness = strings.TrimSpace(readiness + ": " + detail)
		}
		details := strings.Trim(strings.Join([]string{locality, cost, readiness}, ", "), ", ")
		if details != "" {
			details = " (" + details + ")"
		}
		fmt.Fprintf(w, "- %s: %s%s\n", target.ID, target.Label, details)
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Use `jini route set codex` or `jini route set azure-code` to lock a configured route.")
	fmt.Fprintln(w, "CLI handoffs invoke installed CLIs or fail closed; they are not provider aliases.")
	fmt.Fprintln(w, "Use `jini route help` for setup guidance.")
	fmt.Fprintln(w, "Use `jini route auto` to restore automatic routing.")
}

func renderRouteSetupHelp(w io.Writer) {
	fmt.Fprintln(w, "Route setup")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "1. Run `jini route list`.")
	fmt.Fprintln(w, "2. Run `jini doctor`.")
	fmt.Fprintln(w, "3. Run `jini route set codex` or `jini route set claude-code`.")
	fmt.Fprintln(w, "4. Run `jini route dogfood`, `jini route smoke codex`, then validate after the real CLI smoke.")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Azure OpenAI API: set `AZURE_OPENAI_ENDPOINT`, `AZURE_OPENAI_API_KEY`, and `AZURE_OPENAI_DEPLOYMENT`; then `jini route set azure-openai`.")
	fmt.Fprintln(w, "Bedrock API: set `AWS_REGION` plus `AWS_PROFILE` or `AWS_ACCESS_KEY_ID`/`AWS_SECRET_ACCESS_KEY`; then `jini route set bedrock-sonnet`.")
	fmt.Fprintln(w, "Local/offline model: run a local OpenAI-compatible server, then `jini route set local-slm` or `jini route auto`.")
	fmt.Fprintln(w, "Local preview: run `jini route set local-preview` when you want deterministic no-model fallback.")
	fmt.Fprintln(w, "Auto mode: run `jini route auto`.")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Use env overrides only when auto-detect fails:")
	fmt.Fprintln(w, "- `AZURE_OPENAI_ENDPOINT=https://...openai.azure.com`")
	fmt.Fprintln(w, "- `AZURE_OPENAI_DEPLOYMENT=your-deployment`")
	fmt.Fprintln(w, "- `AWS_PROFILE=your-profile` or `AWS_ACCESS_KEY_ID=...`")
	fmt.Fprintln(w, "- Wave 1 CLI overrides: `JINI_CODEX_CLI`, `JINI_CLAUDE_CODE_CLI`, `JINI_GEMINI_CLI`, `JINI_AIDER_CLI`, `JINI_OPENCODE_CLI`.")
	fmt.Fprintln(w, "- `JINI_LOCAL_SLM_ENDPOINT=http://127.0.0.1:11434/v1`")
	fmt.Fprintln(w, "- `JINI_LOCAL_SLM_MODEL=qwen3:8b`")
	fmt.Fprintln(w, "Gemini API and Vertex AI provider routes stay planned until `gemini-cli` dogfood evidence is complete.")
}

func routeTargetReadinessLabel(target savedRouteTarget) string {
	if cliHandoffMode(target.ID) {
		return routeReadinessLabel(detectCLIHandoffProvider(target.ID))
	}
	mode := strings.TrimSpace(target.ProviderMode)
	if mode == "" {
		if descriptor, ok := adapterDescriptorForMode(target.ID); ok {
			mode = descriptor.ProviderMode
		}
	}
	if mode == "" {
		return "unknown"
	}
	return routeReadinessLabel(detectProviderForMode(mode))
}

func routeTargetReadinessDetail(target savedRouteTarget) string {
	if !cliHandoffMode(target.ID) {
		return ""
	}
	provider := detectCLIHandoffProvider(target.ID)
	if routeReadinessLabel(provider) == "ok" || len(provider.Missing) == 0 {
		return ""
	}
	return shipCLIHandoffSetupCategory(provider.Missing)
}

func parseOptionalFormatArgs(args []string) (string, bool) {
	if len(args) == 0 {
		return "", true
	}
	if len(args) == 1 && strings.HasPrefix(strings.TrimSpace(args[0]), "--format=") {
		format := normalizeName(strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(args[0]), "--format=")))
		return format, format == "json" || format == "text"
	}
	if len(args) == 2 && strings.TrimSpace(args[0]) == "--format" {
		format := normalizeName(args[1])
		return format, format == "json" || format == "text"
	}
	return "", false
}

func runStatus(stdout, stderr io.Writer) int {
	current, err := loadCurrentWork()
	if err != nil || current == nil {
		renderNoCurrentWorkStatus(stdout)
		return 0
	}
	summary, err := loadWorkSummary(current.PackDir, current)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			_ = clearCurrentWork()
			fmt.Fprintln(stdout, "Remembered work is no longer available.")
			fmt.Fprintln(stdout)
			renderNoCurrentWorkStatus(stdout)
			return 0
		}
		fmt.Fprintln(stderr, "Could not load current work. Run `jini` to start again.")
		return 1
	}
	renderCheck(stdout, summary)
	return 0
}

func runHelp(args []string, stdout, stderr io.Writer) int {
	if len(args) > 0 {
		switch canonicalHelpTopic(args[0]) {
		case "all", "commands":
			renderPublicCommandInventory(stdout)
			return 0
		case "admin":
			renderAdminCommandInventory(stdout)
			return 0
		}
	}
	current, err := loadCurrentWork()
	if err == nil && current != nil {
		summary, loadErr := loadWorkSummary(current.PackDir, current)
		if loadErr == nil {
			renderCurrentWorkHelp(stdout, summary)
			return 0
		}
		if errors.Is(loadErr, os.ErrNotExist) {
			_ = clearCurrentWork()
			fmt.Fprintln(stdout, "Remembered work is no longer available.")
			fmt.Fprintln(stdout)
			renderNewWorkLauncher(stdout)
			return 0
		}
		fmt.Fprintln(stderr, "Could not load current work. Showing start help instead.")
	}
	renderNewWorkLauncher(stdout)
	return 0
}

func runAdmin(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		renderAdminCommandInventory(stdout)
		return 0
	}
	if isAdminHelpAlias(args[0]) {
		renderAdminCommandInventory(stdout)
		return 0
	}
	fmt.Fprintf(stderr, "Unknown admin command %q.\n", args[0])
	fmt.Fprintln(stderr, "Try `jini admin help`.")
	return 1
}

func detectProvider() providerConfig {
	if route := detectRoute(); route.Active {
		return route.Provider
	}
	return detectLegacyProvider()
}

func detectLegacyProvider() providerConfig {
	switch configuredProviderMode() {
	case "local-preview":
		return detectLocalPreviewProvider()
	case "local-slm":
		return detectLocalSLMProvider()
	case "azure-openai":
		return detectAzureOpenAIProvider()
	case "bedrock":
		return detectBedrockProvider()
	case "anthropic":
		return detectAnthropicProvider()
	case "auto":
		return detectAutoProvider()
	default:
		providerID := normalizeName(configValue("JINI_PROVIDER"))
		return providerConfig{
			ID:      providerID,
			Label:   titleCase(providerID),
			Status:  "needs setup",
			Missing: []string{"Supported JINI_PROVIDER value: auto, claude, azure-openai, bedrock, local-slm, or local-preview"},
		}
	}
}

func detectLocalPreviewProvider() providerConfig {
	return providerConfig{
		ID:     "local-preview",
		Label:  "Local preview",
		Status: "ok",
		Settings: []string{
			"JINI_PROVIDER: " + firstNonEmpty(configValue("JINI_PROVIDER"), "auto") + " -> Local preview",
		},
	}
}

func detectLocalSLMProvider() providerConfig {
	defaultModelID, defaultModelLabel := resolveLocalSLMDefaultModel()
	device := currentDeviceProfileForProviderEvidence()
	missing := []string{}
	endpoint, _ := resolvedLocalSLMEndpoint()
	if strings.TrimSpace(endpoint) == "" {
		missing = append(missing, "JINI_LOCAL_SLM_ENDPOINT or local OpenAI-compatible server")
	}
	if strings.TrimSpace(defaultModelID) == "" {
		missing = append(missing, "JINI_LOCAL_SLM_MODEL or discoverable local chat model")
	}
	settings := []string{
		providerSettingLine("local-slm"),
		localSLMEndpointSettingLine(),
		"DEVICE_CLASS: " + firstNonEmpty(device.DeviceClass, "unknown"),
		"DEVICE_OS: " + strings.TrimSpace(firstNonEmpty(device.OS, "unknown")) + " " + strings.TrimSpace(firstNonEmpty(device.OSVersion, "unknown")),
		"LOCAL_ACCELERATOR: " + firstNonEmpty(device.AcceleratorClass, "unknown"),
		"LOCAL_RUNTIME_CLASS: " + firstNonEmpty(device.LocalRuntimeClass, "unknown"),
	}
	for _, slot := range []struct {
		env   string
		label string
	}{
		{"JINI_LOCAL_SLM_MODEL", "Default local model"},
		{"JINI_LOCAL_SLM_FAST_MODEL", "Fast profile"},
		{"JINI_LOCAL_SLM_WORKHORSE_MODEL", "Workhorse profile"},
		{"JINI_LOCAL_SLM_DEEP_MODEL", "Deep profile"},
		{"JINI_LOCAL_SLM_MULTIMODAL_MODEL", "Multimodal profile"},
	} {
		modelID, modelLabel := resolveLocalSLMModelForToolMode(toolModeForLocalSetting(slot.env))
		if slot.env == "JINI_LOCAL_SLM_MODEL" {
			modelID, modelLabel = defaultModelID, defaultModelLabel
		}
		if slot.env == "JINI_LOCAL_SLM_MODEL" {
			settings = append(settings, localSLMModelSettingLine(slot.env, modelID, firstNonEmpty(modelLabel, defaultModelLabel)))
			continue
		}
		state := strings.TrimSpace(device.LocalProfileStates[toolModeForLocalSetting(slot.env)])
		if state != "" {
			settings = append(settings, localSLMModelSettingLine(slot.env, modelID, modelLabel)+" ("+state+" on this device)")
			continue
		}
		settings = append(settings, localSLMModelSettingLine(slot.env, modelID, modelLabel))
	}
	if power := formatPowerProfileForRoute(); power != "" {
		settings = append(settings, "POWER_STATE: "+power)
	}
	if len(missing) == 0 {
		settings = append(settings, freshLocalBenchmarkSummaryLines()...)
		settings = append(settings, freshLocalMultimodalLearningSummaryLines()...)
	}
	return providerConfig{
		ID:       "local-slm",
		Label:    "Local SLM",
		Status:   statusFromMissing(missing),
		Missing:  missing,
		Settings: settings,
		Secrets:  []string{"JINI_LOCAL_SLM_API_KEY: " + presentOrMissing("JINI_LOCAL_SLM_API_KEY")},
	}
}

func currentDeviceProfileForProviderEvidence() deviceProfile {
	if !shouldUseCachedProviderDeviceEvidence() {
		return currentDeviceProfile()
	}
	return loadDeviceProfile()
}

func shouldUseCachedProviderDeviceEvidence() bool {
	return strings.TrimSpace(configValue("JINI_DEVICE_CLASS_OVERRIDE")) == "" && deviceProfileHasProviderEvidence(loadDeviceProfile())
}

func deviceProfileHasProviderEvidence(profile deviceProfile) bool {
	return profile.ContextType == "JiniDeviceProfile" &&
		strings.TrimSpace(profile.DeviceClass) != "" &&
		strings.TrimSpace(profile.OS) != "" &&
		strings.TrimSpace(profile.OSVersion) != "" &&
		strings.TrimSpace(profile.AcceleratorClass) != "" &&
		strings.TrimSpace(profile.LocalRuntimeClass) != "" &&
		hasExpectedLocalProfileSlots(profile.LocalProfileStates)
}

func shouldRefreshLocalBenchmarkForDoctor() bool {
	if configuredProviderMode() == "local-slm" {
		return true
	}
	switch configuredToolMode() {
	case "local-fast", "local-workhorse", "local-deep", "local-multimodal":
		return true
	default:
		return false
	}
}

func toolModeForLocalSetting(envName string) string {
	switch envName {
	case "JINI_LOCAL_SLM_FAST_MODEL":
		return "local-fast"
	case "JINI_LOCAL_SLM_WORKHORSE_MODEL":
		return "local-workhorse"
	case "JINI_LOCAL_SLM_DEEP_MODEL":
		return "local-deep"
	case "JINI_LOCAL_SLM_MULTIMODAL_MODEL":
		return "local-multimodal"
	default:
		return ""
	}
}

func detectAzureOpenAIProvider() providerConfig {
	missing := missingEnvVars([]string{
		"AZURE_OPENAI_ENDPOINT",
		"AZURE_OPENAI_API_KEY",
		"AZURE_OPENAI_DEPLOYMENT",
	})
	label := "Azure OpenAI"
	if deployment := configValue("AZURE_OPENAI_DEPLOYMENT"); deployment != "" {
		label += " / " + deployment
	}
	settings := []string{
		providerSettingLine("azure-openai"),
		"AZURE_OPENAI_ENDPOINT: " + presentOrMissing("AZURE_OPENAI_ENDPOINT"),
		"AZURE_OPENAI_DEPLOYMENT: " + presentOrMissing("AZURE_OPENAI_DEPLOYMENT"),
		"AZURE_OPENAI_API_VERSION: " + valueOrDefault("AZURE_OPENAI_API_VERSION", "2024-10-21"),
	}
	if modelLine := azureModelSettingLine(); modelLine != "" {
		settings = append(settings, modelLine)
	}
	return providerConfig{
		ID:       "azure-openai",
		Label:    label,
		Status:   statusFromMissing(missing),
		Missing:  missing,
		Settings: settings,
		Secrets:  []string{"AZURE_OPENAI_API_KEY: " + presentOrMissing("AZURE_OPENAI_API_KEY")},
	}
}

func detectBedrockProvider() providerConfig {
	missing := []string{}
	if strings.TrimSpace(resolveAWSRegion()) == "" {
		missing = append(missing, "AWS_REGION or AWS_DEFAULT_REGION")
	}
	if configValue("AWS_PROFILE") == "" && (configValue("AWS_ACCESS_KEY_ID") == "" || configValue("AWS_SECRET_ACCESS_KEY") == "") {
		missing = append(missing, "AWS_PROFILE or AWS_ACCESS_KEY_ID/AWS_SECRET_ACCESS_KEY")
	}
	modelID, modelLabel := resolveBedrockModel()
	label := "Amazon Bedrock"
	if modelLabel != "" {
		label += " / " + modelLabel
	}
	settings := []string{
		providerSettingLine("bedrock"),
		"AWS_REGION or profile region: " + presentOrMissingValue(resolveAWSRegion()),
		bedrockModelSettingLine(modelID, modelLabel),
	}
	return providerConfig{
		ID:       "bedrock",
		Label:    label,
		Status:   statusFromMissing(missing),
		Missing:  missing,
		Settings: settings,
		Secrets: []string{
			"AWS_PROFILE: " + presentOrMissing("AWS_PROFILE"),
			"AWS_ACCESS_KEY_ID: " + presentOrMissing("AWS_ACCESS_KEY_ID"),
			"AWS_SECRET_ACCESS_KEY: " + presentOrMissing("AWS_SECRET_ACCESS_KEY"),
		},
	}
}

func detectAnthropicProvider() providerConfig {
	missing := []string{}
	if configValue("ANTHROPIC_API_KEY") == "" {
		missing = append(missing, "ANTHROPIC_API_KEY")
	}
	modelID, modelLabel, modelIssue := resolveAnthropicModel()
	if modelIssue != "" {
		missing = append(missing, modelIssue)
	}
	label := "Claude API"
	if modelLabel != "" {
		label += " / " + modelLabel
	}
	settings := []string{
		providerSettingLine("anthropic"),
		anthropicModelSettingLine(modelID, modelLabel),
		"ANTHROPIC_BASE_URL: " + presentOrMissing("ANTHROPIC_BASE_URL") + " (default https://api.anthropic.com)",
	}
	return providerConfig{
		ID:       "anthropic",
		Label:    label,
		Status:   statusFromMissing(missing),
		Missing:  missing,
		Settings: settings,
		Secrets:  []string{"ANTHROPIC_API_KEY: " + presentOrMissing("ANTHROPIC_API_KEY")},
	}
}

func missingEnvVars(names []string) []string {
	missing := []string{}
	for _, name := range names {
		if configValue(name) == "" {
			missing = append(missing, name)
		}
	}
	return missing
}

func statusFromMissing(missing []string) string {
	if len(missing) == 0 {
		return "ok"
	}
	return "needs setup"
}

func presentOrMissing(name string) string {
	if configValue(name) == "" {
		return "missing"
	}
	return "set"
}

func presentOrMissingEither(names ...string) string {
	for _, name := range names {
		if configValue(name) != "" {
			return "set"
		}
	}
	return "missing"
}

func presentOrMissingValue(value string) string {
	if strings.TrimSpace(value) == "" {
		return "missing"
	}
	return "set"
}

func valueOrDefault(name, fallback string) string {
	value := configValue(name)
	if value == "" {
		return fallback
	}
	return value
}

func renderProviderDoctor(w io.Writer, provider providerConfig) {
	fmt.Fprintln(w, "Provider")
	fmt.Fprintln(w, provider.Label)
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Status")
	fmt.Fprintln(w, provider.Status)
	if len(provider.Settings) > 0 {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Settings")
		for _, item := range provider.Settings {
			fmt.Fprintf(w, "- %s\n", item)
		}
	}
	if len(provider.Secrets) > 0 {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Secrets")
		for _, item := range provider.Secrets {
			fmt.Fprintf(w, "- %s\n", item)
		}
	}
	if len(provider.Missing) > 0 {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Missing")
		for _, item := range provider.Missing {
			fmt.Fprintf(w, "- %s\n", item)
		}
	}
}

func renderInteractiveProviderDoctor(w io.Writer) {
	if shouldRefreshLocalBenchmarkForDoctor() {
		_ = currentLocalRuntimeCapabilities(context.Background())
	}
	renderProviderDoctor(w, detectProvider())
}

func workingWithLabel(provider providerConfig) string {
	if route := detectRoute(); route.Active {
		label := route.ToolLabel
		if provider.Label != "" && provider.Label != route.ToolLabel && provider.ID != "local-preview" {
			label += " via " + provider.Label
		}
		if route.ChosenAutomatically {
			return label + " (chosen automatically)"
		}
		return label
	}
	if configuredProviderMode() == "auto" {
		return provider.Label + " (chosen automatically)"
	}
	return provider.Label
}

func runNewWorkIntake(stdin io.Reader, stdout, stderr io.Writer) int {
	session := bufio.NewScanner(stdin)
	return runNewWorkIntakeWithScanner(session, stdout, stderr)
}

func runDirectTaskArgsIntake(args []string, stdout, stderr io.Writer) int {
	source := strings.TrimSpace(strings.Join(args, " "))
	if source == "" {
		fmt.Fprintln(stderr, "I need one line of source context to start this work.")
		return 1
	}
	if handled, exitCode := maybeHandleLocalTextFileEditIntent(source, stdout, stderr); handled {
		return exitCode
	}
	if handled, exitCode := maybeHandleDirectPromptIntent(source, stdout, stderr); handled {
		return exitCode
	}
	if maybeHandleSimpleAnswer(source, stdout) {
		return 0
	}
	if maybeHandleStandaloneQuestion(source, stdout) {
		return 0
	}
	if maybeHandleAmbiguousBareEntity(source, stdout) {
		return 0
	}
	envelope := classifyWorkEnvelope(starterChoice{}, source)
	inputItems, normalizedSource := inputItemsForSource(source)
	if strings.TrimSpace(normalizedSource) != "" {
		source = normalizedSource
		envelope = classifyWorkEnvelope(envelope.Choice, source)
	}

	request := providerGenerationRequest{
		Choice: envelope.Choice,
		Title:  deriveStarterTitle(envelope.Choice.DefaultName, source, envelope.Choice.PackID),
		Source: source,
	}
	_ = detectRouteForRequest(request)

	summary, err := bootstrapStarterWork(envelope.Choice, source, "quick", inputItems)
	if err != nil {
		fmt.Fprintf(stderr, "Could not start this work: %v\n", err)
		return 1
	}
	if isRepoReviewDirectTask(source) {
		_, snapshot, snapshotErr := applyRepoReviewSnapshot(summary, source)
		if snapshotErr != nil {
			fmt.Fprintf(stderr, "Could not prepare repo review snapshot: %v\n", snapshotErr)
			return 1
		}
		renderRepoReviewDirectTaskStarted(stdout, source, snapshot)
		return 0
	}
	renderDirectTaskStarted(stdout, summary, source)
	return 0
}

func runNewWorkIntakeWithScanner(session *bufio.Scanner, stdout, stderr io.Writer) int {
	_ = warmLocalRuntimeCapabilities()
	renderNewWorkPrompt(stdout)

	for {
		firstRaw, ok := readInputLine(session, stdout)
		if !ok {
			return 0
		}
		if isHelpInput(firstRaw) {
			fmt.Fprintln(stdout)
			renderNewWorkLauncher(stdout)
			fmt.Fprintln(stdout)
			continue
		}
		if isGreetingOnly(firstRaw) {
			fmt.Fprintln(stdout)
			fmt.Fprintln(stdout, "Hi.")
			fmt.Fprintln(stdout, "Describe the task when you're ready.")
			fmt.Fprintln(stdout)
			continue
		}
		if handled, exitCode := maybeHandleNewWorkUtilityIntent(firstRaw, stdout); handled {
			if exitCode != 0 {
				return exitCode
			}
			fmt.Fprintln(stdout)
			continue
		}
		if isAcknowledgementOnly(firstRaw) || isBareContinuationIntent(firstRaw) {
			fmt.Fprintln(stdout)
			renderNewWorkNoop(stdout)
			fmt.Fprintln(stdout)
			continue
		}
		if handled, exitCode := maybeHandleProviderSetupIntent(firstRaw, session, stdout, stderr); handled {
			if exitCode != 0 {
				fmt.Fprintln(stdout)
				renderNewWorkLauncher(stdout)
				continue
			}
			fmt.Fprintln(stdout)
			renderNewWorkLauncher(stdout)
			continue
		}
		if isSlashCommandInput(firstRaw) {
			fmt.Fprintf(stderr, "Unknown command %q.\n", strings.TrimSpace(firstRaw))
			return 1
		}
		return startNewWorkFromRawInput(firstRaw, session, stdout, stderr)
	}
}

func startNewWorkFromRawInput(firstRaw string, session *bufio.Scanner, stdout, stderr io.Writer) int {
	var source string
	choiceExplicit := false
	choice, err := resolveStarterChoice(firstRaw)
	if err != nil {
		source = strings.TrimSpace(firstRaw)
		choice = classifyStarterChoice(source)
	} else {
		choiceExplicit = true
		var ok bool
		source, ok = readPromptLine(session, stdout, sourcePromptForChoice(choice))
		if !ok {
			return 0
		}
		if strings.TrimSpace(source) == "" {
			fmt.Fprintln(stderr, "I need one line of source context to start this work.")
			return 1
		}
		if choice.PackID == "auto" {
			choice = classifyStarterChoice(source)
		}
	}
	if strings.TrimSpace(source) == "" {
		fmt.Fprintln(stderr, "I need one line of source context to start this work.")
		return 1
	}
	if handled, exitCode := maybeHandleLocalTextFileEditIntent(source, stdout, stderr); handled {
		return exitCode
	}
	if handled, exitCode := maybeHandleDirectPromptIntent(source, stdout, stderr); handled {
		return exitCode
	}
	if maybeHandleSimpleAnswer(source, stdout) {
		return 0
	}
	if !choiceExplicit && maybeHandleStandaloneQuestion(source, stdout) {
		return 0
	}
	if !choiceExplicit && maybeHandleAmbiguousBareEntity(source, stdout) {
		return 0
	}
	envelope := classifyWorkEnvelope(choice, source)
	inputItems, normalizedSource := inputItemsForSource(source)
	if strings.TrimSpace(normalizedSource) != "" {
		source = normalizedSource
		envelope = classifyWorkEnvelope(choice, source)
	}
	clarifiedSource, clarificationItem, ok := maybeClarifyStarterSource(envelope, session, stdout)
	if !ok {
		return 0
	}
	source = clarifiedSource
	envelope = classifyWorkEnvelope(envelope.Choice, source)
	if clarificationItem.InputID != "" {
		inputItems = append(inputItems, clarificationItem)
	}

	request := providerGenerationRequest{
		Choice: envelope.Choice,
		Title:  deriveStarterTitle(envelope.Choice.DefaultName, source, envelope.Choice.PackID),
		Source: source,
	}
	_ = detectRouteForRequest(request)

	summary, err := bootstrapStarterWork(envelope.Choice, source, "quick", inputItems)
	if err != nil {
		fmt.Fprintf(stderr, "Could not start this work: %v\n", err)
		return 1
	}

	renderFirstRunResult(stdout, summary)
	action, ok := readOptionalInputLine(session, stdout)
	if !ok || strings.TrimSpace(action) == "" {
		fmt.Fprintln(stdout)
		return 0
	}
	fmt.Fprintln(stdout)
	return handlePostResultAction(action, summary, session, stdout, stderr)
}

func readInputLine(scanner *bufio.Scanner, stdout io.Writer) (string, bool) {
	fmt.Fprint(stdout, "> ")
	if !scanner.Scan() {
		return "", false
	}
	return strings.TrimSpace(scanner.Text()), true
}

func readOptionalInputLine(scanner *bufio.Scanner, stdout io.Writer) (string, bool) {
	fmt.Fprint(stdout, "> ")
	if !scanner.Scan() {
		return "", false
	}
	return strings.TrimSpace(scanner.Text()), true
}

func readPromptLine(scanner *bufio.Scanner, stdout io.Writer, prompt string) (string, bool) {
	fmt.Fprintln(stdout)
	fmt.Fprintln(stdout, prompt)
	fmt.Fprint(stdout, "> ")
	if !scanner.Scan() {
		return "", false
	}
	return strings.TrimSpace(scanner.Text()), true
}

func resolveStarterChoice(raw string) (starterChoice, error) {
	if !looksLikeStarterMenuSelection(raw) {
		return starterChoice{}, fmt.Errorf("No starter flow matches %q yet.", raw)
	}
	choice := normalizeName(raw)
	switch choice {
	case "3", "i am not sure", "i'm not sure", "i’m not sure", "im not sure", "i m not sure", "not sure", "unsure", "help me finish this":
		return starterChoice{PackID: "auto", ChoiceLabel: "I am not sure", DefaultName: "Request Brief", State: "decided"}, nil
	case "plan this first", "plan first":
		return starterChoice{PackID: "auto", ChoiceLabel: "Plan", DefaultName: "Plan", State: "modeled"}, nil
	}
	for _, packID := range starterPackMenuOrder {
		profile, ok := starterProfileForPack(packID)
		if !ok {
			continue
		}
		for _, alias := range profile.MenuAliases {
			if choice == alias {
				resolved, ok := starterChoiceForPack(profile.PackID)
				if ok {
					return resolved, nil
				}
			}
		}
	}
	return starterChoice{}, fmt.Errorf("No starter flow matches %q yet.", raw)
}

func looksLikeStarterMenuSelection(raw string) bool {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return false
	}
	runes := []rune(trimmed)
	last := runes[len(runes)-1]
	return last != '.' && last != '!' && last != '?'
}

func isGreetingOnly(raw string) bool {
	switch normalizeName(raw) {
	case "hello", "hi", "hey", "hey there", "hello there", "good morning", "good afternoon", "good evening", "morning", "evening":
		return true
	default:
		return false
	}
}

func isAcknowledgementOnly(raw string) bool {
	switch normalizeName(raw) {
	case "thanks", "thank you", "thx", "ok", "okay", "k", "cool", "got it", "sounds good", "yes", "yep", "yeah", "no", "nope", "done":
		return true
	default:
		return false
	}
}

func isBareContinuationIntent(raw string) bool {
	switch normalizeName(raw) {
	case "continue":
		return true
	default:
		return false
	}
}

func isHelpInput(raw string) bool {
	switch interactiveCommandName(raw) {
	case "help", "what can you do", "?":
		return true
	default:
		return false
	}
}

func maybeHandleNewWorkUtilityIntent(raw string, stdout io.Writer) (bool, int) {
	if access, ok := commercialOnlyCommandAccess(raw); ok {
		fmt.Fprintln(stdout)
		renderFeatureAccessDenied(stdout, access)
		return true, 1
	}
	switch interactiveCommandName(raw) {
	case "status":
		fmt.Fprintln(stdout)
		renderNoCurrentWorkStatus(stdout)
		return true, 0
	case "doctor":
		fmt.Fprintln(stdout)
		renderInteractiveProviderDoctor(stdout)
		return true, 0
	case "route":
		fmt.Fprintln(stdout)
		renderRouteCostStatus(stdout)
		return true, 0
	case "route help":
		fmt.Fprintln(stdout)
		renderRouteSetupHelp(stdout)
		return true, 0
	case "memory":
		fmt.Fprintln(stdout)
		renderUserContextMemorySummary(stdout, "none")
		return true, 0
	case "memory inspect", "memory status":
		fmt.Fprintln(stdout)
		return true, runMemory([]string{"inspect"}, stdout, stdout)
	case "memory off", "memory disable":
		fmt.Fprintln(stdout)
		return true, runMemory([]string{"off"}, stdout, stdout)
	case "memory on", "memory enable":
		fmt.Fprintln(stdout)
		return true, runMemory([]string{"on"}, stdout, stdout)
	case "memory forget", "memory clear", "memory revoke":
		fmt.Fprintln(stdout)
		return true, runMemory([]string{"forget"}, stdout, stdout)
	case "permissions":
		fmt.Fprintln(stdout)
		renderSafePermissionsStatus(stdout)
		return true, 0
	case "init":
		fmt.Fprintln(stdout)
		renderNoInitRequired(stdout)
		fmt.Fprintln(stdout)
		renderNewWorkPrompt(stdout)
		return true, 0
	case "clear":
		fmt.Fprintln(stdout)
		fmt.Fprintln(stdout, "Nothing to clear yet.")
		fmt.Fprintln(stdout, "Describe the task when you're ready.")
		return true, 0
	default:
		return false, 0
	}
}

func sourcePromptForChoice(choice starterChoice) string {
	if choice.PackID == "auto" {
		return strings.Join([]string{
			"Describe the task in one sentence.",
			"Jini can answer, edit files, route to a configured tool, or ask one clarification.",
			"Nothing will be sent, booked, committed, or changed without a visible step.",
		}, "\n")
	}
	return "Describe the task in one sentence."
}

func classifyStarterChoice(source string) starterChoice {
	return classifyWorkEnvelope(starterChoice{}, source).Choice
}

func maybeClarifyStarterSource(envelope workEnvelope, scanner *bufio.Scanner, stdout io.Writer) (string, inputItem, bool) {
	if scanner == nil {
		return envelope.Source, inputItem{}, true
	}
	prompt, ok := clarificationPromptForEnvelope(envelope)
	if !ok {
		return envelope.Source, inputItem{}, true
	}
	answer, answered := readPromptLine(scanner, stdout, prompt)
	if !answered {
		return "", inputItem{}, false
	}
	answer = strings.TrimSpace(answer)
	if answer == "" || normalizeName(answer) == "skip" {
		return envelope.Source, inputItem{}, true
	}
	return mergeClarifiedSource(envelope.Source, answer), inputItem{
		InputID:   "clarified-scope",
		Kind:      "clarification",
		Title:     "Clarified scope",
		Status:    "processed",
		Preview:   compactPreview(answer, 120),
		OriginRef: answer,
	}, true
}

func classifyWorkEnvelope(explicitChoice starterChoice, source string) workEnvelope {
	source = strings.TrimSpace(source)
	choice := explicitChoice
	if strings.TrimSpace(choice.PackID) == "" {
		packID := detectStarterPackFromSource(source)
		resolved, ok := starterChoiceForPack(packID)
		if !ok {
			resolved = starterChoice{PackID: "general-work", ChoiceLabel: "Request Brief", DefaultName: "Request Brief", State: "decided"}
		}
		choice = resolved
	}
	profile := starterProfile(choice.PackID)
	return workEnvelope{
		Choice:         choice,
		Goal:           deriveStarterTitle(choice.DefaultName, source, choice.PackID),
		Source:         source,
		WorkClass:      strings.TrimSpace(profile.WorkClass),
		RequestCohort:  strings.TrimSpace(profile.RequestCohort),
		ArtifactFamily: strings.TrimSpace(profile.ArtifactFamily),
	}
}

func clarificationPromptForEnvelope(envelope workEnvelope) (string, bool) {
	return clarificationPromptForCohort(envelope.RequestCohort, envelope.Source)
}

func mergeClarifiedSource(source, answer string) string {
	base := strings.TrimSpace(source)
	scope := strings.TrimSpace(answer)
	if base == "" {
		return scope
	}
	if scope == "" {
		return base
	}
	if strings.HasSuffix(base, ".") || strings.HasSuffix(base, "!") || strings.HasSuffix(base, "?") {
		return base + " Scope: " + scope
	}
	return base + ". Scope: " + scope
}

func bootstrapStarterWork(choice starterChoice, source, detail string, inputItems []inputItem) (*workSummary, error) {
	if len(inputItems) == 0 {
		inputItems, _ = inputItemsForSource(source)
	}
	stateRoot := sessionStateRoot()
	workRoot := filepath.Join(stateRoot, "work")
	if err := os.MkdirAll(workRoot, 0o755); err != nil {
		return nil, err
	}

	title := deriveStarterTitle(choice.DefaultName, source, choice.PackID)
	workDir, err := uniqueWorkDir(workRoot, choice.PackID, title)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Join(workDir, "views"), 0o755); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Join(workDir, "artifacts"), 0o755); err != nil {
		return nil, err
	}

	if err := writeStarterWork(choice, workDir, title, source, detail); err != nil {
		return nil, err
	}
	if err := enrichSmartHyperlinksInViews(workDir, providerGenerationRequest{
		Choice: choice,
		Title:  title,
		Source: source,
	}); err != nil {
		return nil, err
	}
	if err := saveInputItems(workDir, inputItems); err != nil {
		return nil, err
	}
	initialSummary := &workSummary{
		Dir:        workDir,
		PackID:     choice.PackID,
		WorkUnitID: slugify(title),
		Title:      title,
		State:      choice.State,
		Views:      collectViews(workDir, choice.PackID),
		Missing:    inferMissing(choice.State, collectDetails(workDir), choice.PackID, source),
		Uncertain:  inferUncertain(choice.PackID, inferMissing(choice.State, collectDetails(workDir), choice.PackID, source), source),
		Doing:      inferDoing(choice.PackID, choice.State),
		NextStep:   inferNextStep(choice.PackID, collectViews(workDir, choice.PackID)),
		SafeToDo:   "Nothing has been sent yet. You can review before sharing.",
	}
	if err := saveThreadState(workDir, synthesizeThreadState(initialSummary)); err != nil {
		return nil, err
	}
	if err := maybeWriteProviderFirstDraft(context.Background(), choice, workDir, title, source); err != nil {
		return nil, err
	}
	if err := saveArtifactFeedbackBaseline(workDir); err != nil {
		return nil, err
	}

	current := &currentWork{
		PackDir:    workDir,
		PackID:     choice.PackID,
		WorkUnitID: slugify(title),
		Title:      title,
		State:      choice.State,
		Health:     inferHealthFromState(choice.State),
	}
	if err := saveCurrentWork(current); err != nil {
		return nil, err
	}
	summary, err := loadWorkSummary(workDir, current)
	if err != nil {
		return nil, err
	}
	return summary, nil
}

func saveCurrentWork(current *currentWork) error {
	if err := os.MkdirAll(sessionStateRoot(), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(map[string]any{
		"schema_version": "0.1.0",
		"context_type":   "JiniCurrentWork",
		"pack_dir":       current.PackDir,
		"pack_id":        current.PackID,
		"work_unit_id":   current.WorkUnitID,
		"title":          current.Title,
		"state":          current.State,
		"health":         current.Health,
	}, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(sessionStateRoot(), "current-work.json"), append(data, '\n'), 0o644)
}

func writeStarterWork(choice starterChoice, workDir, title, source, detail string) error {
	workUnit := strings.Join([]string{
		fmt.Sprintf("work_unit_id: %s", slugify(title)),
		fmt.Sprintf("title: %s", title),
		fmt.Sprintf("current_state: %s", choice.State),
		"",
	}, "\n")
	if err := os.WriteFile(filepath.Join(workDir, "work-unit.yaml"), []byte(workUnit), 0o644); err != nil {
		return err
	}

	profile, ok := starterProfileForPack(choice.PackID)
	if ok && profile.Writer != nil {
		return profile.Writer(workDir, title, source, detail)
	}
	return writeFirstUsefulPassStarterWork(workDir, title, source)
}

func writeMeetingStarterWork(workDir, title, source, detail string) error {
	return writeStarterArtifactPlan(workDir, buildMeetingArtifactPlan(title, source, detail))
}

func writeResearchStarterWork(workDir, title, source, detail string) error {
	return writeStarterArtifactPlan(workDir, buildResearchArtifactPlan(title, source, detail))
}

func writeTravelStarterWork(workDir, title, source, detail string) error {
	return writeStarterArtifactPlan(workDir, buildTravelArtifactPlan(title, source))
}

func writeFirstUsefulPassStarterWork(workDir, title, source string) error {
	return writeStarterArtifactPlan(workDir, buildFirstUsefulPassArtifactPlan(title, source))
}

func writeSimpleStarterWork(workDir, title, viewLabel, source string, bullets []string) error {
	return writeStarterArtifactPlan(workDir, buildSimpleArtifactPlan(title, viewLabel, source, bullets))
}

func writeStarterArtifacts(workDir string, artifactTypes []string) error {
	for index, artifactType := range artifactTypes {
		filename := fmt.Sprintf("%02d-%s.yaml", index+1, normalizeFilename(artifactType))
		content := fmt.Sprintf("artifact_type: %s\nstatus: ready\n", artifactType)
		if err := os.WriteFile(filepath.Join(workDir, "artifacts", filename), []byte(content), 0o644); err != nil {
			return err
		}
	}
	return nil
}

func starterTripDays(ctx travelStarterContext) []string {
	dayCount := ctx.DayCount
	if dayCount <= 0 {
		dayCount = 7
	}
	if dayCount > 14 {
		dayCount = 14
	}
	destination := ctx.Destination
	if destination == "" {
		destination = "the destination"
	}
	themes := []string{
		"First neighborhood anchor",
		"Headliner day",
		"Food and wandering day",
		"Flexible favorite day",
		"Day trip or second anchor",
		"Slow day and catch-up",
		"Last highlights and packing buffer",
		"Second-city or deeper exploration",
		"Open favorites day",
		"Departure buffer",
	}
	lines := []string{}
	for day := 1; day <= dayCount; day++ {
		switch {
		case day == 1:
			lines = append(lines,
				fmt.Sprintf("### Day %d: Arrive and settle into %s", day, destination),
				fmt.Sprintf("- Keep the first day light so arrival, check-in, and your first look at %s do not turn into a forced march.", destination),
			)
		case day == dayCount:
			lines = append(lines,
				fmt.Sprintf("### Day %d: Buffer and departure", day),
				"- Keep the final day intentionally light so checkout, bags, airport or rail transfer, and one last meal fit without stress.",
			)
		default:
			index := day - 2
			if index < len(ctx.MustDos) {
				mustDo := ctx.MustDos[index]
				lines = append(lines,
					fmt.Sprintf("### Day %d: %s", day, titleCase(cleanForTitle(mustDo))),
					fmt.Sprintf("- Build the day around %s, then leave enough room for meals, transit, queues, and one lighter backup if timing slips.", strings.ToLower(strings.TrimSpace(mustDo))),
				)
				continue
			}
			theme := themes[min(index, len(themes)-1)]
			lines = append(lines,
				fmt.Sprintf("### Day %d: %s", day, theme),
				fmt.Sprintf("- Use this day to balance one anchor in %s with slower time for food, neighborhoods, or recovery so the trip stays usable end to end.", destination),
			)
		}
	}
	return lines
}

func starterTripBudget(ctx travelStarterContext) []string {
	destination := firstNonEmpty(ctx.Destination, "the destination")
	items := []string{
		fmt.Sprintf("Lodging: choose the base first because where you stay in %s will shape daily transit, fatigue, and how much you can fit into each day.", destination),
		"Food: separate anchor meals from routine meals so the budget stays honest instead of getting eaten by convenience spending.",
		"Transit: include both arrival or departure transfer and the local movement needed to connect the trip's main anchors.",
	}
	if len(ctx.MustDos) > 0 {
		items = append(items, fmt.Sprintf("Tickets: reserve room for at least one paid anchor such as %s before filling the rest of the week.", strings.ToLower(ctx.MustDos[0])))
	}
	return items
}

func starterTripLogistics(ctx travelStarterContext) []string {
	items := []string{
		"Pick the base area before overcommitting the daily plan so each day does not inherit hidden transit cost.",
		"Lock arrival and departure transfer details before filling optional activities around them.",
	}
	if len(ctx.MustDos) > 0 {
		items = append(items, fmt.Sprintf("Reserve any timed anchors like %s before treating the rest of the itinerary as fixed.", strings.ToLower(ctx.MustDos[0])))
	} else {
		items = append(items, "Reserve the first timed anchor before treating the rest of the itinerary as fixed.")
	}
	return items
}

func starterTripContingencies(ctx travelStarterContext) []string {
	return []string{
		"Have one indoor or low-effort backup for every outdoor-heavy or reservation-heavy day.",
		"Leave one uncommitted slot so a sold-out booking or bad weather does not wreck the rest of the plan.",
		"Protect one slower day in the middle of longer trips so fatigue does not spill into the final third.",
	}
}

func bulletLines(items []string) string {
	lines := make([]string, 0, len(items)+1)
	for _, item := range items {
		lines = append(lines, "- "+item)
	}
	lines = append(lines, "")
	return strings.Join(lines, "\n")
}

func starterSourceBullets(source string, limit int, fallback []string) []string {
	parts := splitSourceFragments(source)
	out := make([]string, 0, min(limit, len(parts)))
	seen := map[string]bool{}
	for _, part := range parts {
		clean := strings.TrimSpace(part)
		if clean == "" {
			continue
		}
		if !strings.HasSuffix(clean, ".") && !strings.HasSuffix(clean, "!") && !strings.HasSuffix(clean, "?") {
			clean += "."
		}
		key := normalizeName(clean)
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, strings.ToUpper(clean[:1])+clean[1:])
		if len(out) >= limit {
			break
		}
	}
	if len(out) == 0 {
		return append([]string{}, fallback...)
	}
	return out
}

func splitSourceFragments(source string) []string {
	fields := strings.FieldsFunc(source, func(r rune) bool {
		switch r {
		case '\n', '\r', '.', '!', '?', ';':
			return true
		default:
			return false
		}
	})
	out := make([]string, 0, len(fields))
	for _, field := range fields {
		clean := strings.TrimSpace(field)
		if clean == "" {
			continue
		}
		out = append(out, clean)
	}
	return out
}

func starterMeetingOwnersToConfirm(source string, decisions []string) []string {
	items := []string{}
	for _, decision := range decisions {
		trimmed := strings.TrimSpace(strings.TrimSuffix(decision, "."))
		if trimmed == "" {
			continue
		}
		items = append(items, "Confirm owner and due date for: "+strings.ToLower(trimmed)+".")
		if len(items) >= 3 {
			break
		}
	}
	if containsAny(normalizeName(source), []string{"owner", "due date", "deadline"}) {
		items = append(items, "Reply with any missing owner, due date, or dependency before this note is treated as final.")
	}
	if len(items) == 0 {
		items = []string{
			"Confirm the owner for the highest-priority follow-up item.",
			"Confirm the due date or next checkpoint for each promised deliverable.",
		}
	}
	return dedupeStrings(items)
}

func starterMeetingOpenQuestions(source string) []string {
	items := []string{}
	normalized := normalizeName(source)
	switch {
	case containsAny(normalized, []string{"pricing", "launch"}):
		items = append(items, "Which launch or pricing detail is still blocked on a separate decision?")
	case containsAny(normalized, []string{"renewal", "risk"}):
		items = append(items, "Which renewal risk changes the immediate action plan if it stays unresolved?")
	}
	if containsAny(normalized, []string{"open question", "question", "unknown"}) {
		items = append(items, "Which open question needs a named owner before the next check-in?")
	}
	items = append(items, "Does any dependency or approval need to be called out before work starts?")
	return dedupeStrings(items)
}

func starterMeetingNextMoves(source string) []string {
	items := []string{
		"Send this note today so everyone is working from the same commitments.",
		"Ask each owner to reply with any date or dependency risk before the next check-in.",
	}
	if containsAny(normalizeName(source), []string{"launch", "release", "ship"}) {
		items = append(items, "Close any launch-blocking question before treating the plan as locked.")
	} else {
		items = append(items, "Close the highest-risk open question before the work is considered settled.")
	}
	return items
}

func starterResearchMustClear(source string) []string {
	items := []string{
		"Name the approval owner and the exact decision needed before build starts.",
		"Add a rollback note for the first released slice so failure handling is not implicit.",
		"Lock the first implementation slice to one user-facing outcome instead of a broad theme.",
	}
	normalized := normalizeName(source)
	if containsAny(normalized, []string{"handoff", "hand off"}) {
		items = append(items, "Confirm what the receiving team is expected to build first after handoff.")
	}
	return dedupeStrings(items)
}

func starterResearchFirstSlice(source string) []string {
	items := starterSourceBullets(source, 1, []string{
		"Start with the thinnest slice that proves the user-facing value before wider expansion.",
	})
	first := strings.TrimSpace(strings.TrimSuffix(items[0], "."))
	return []string{
		"Turn " + strings.ToLower(first) + " into the first thin slice that can be built and checked quickly.",
		"Keep the first slice small enough that approval, rollback, and success checks stay obvious.",
	}
}

func starterResearchWhoNeeds(source string) []string {
	items := []string{
		"Product owner: confirm approval and the success boundary for the first slice.",
		"Engineering owner: record rollback behavior and implementation scope before build starts.",
	}
	if containsAny(normalizeName(source), []string{"handoff", "hand off"}) {
		items = append(items, "Receiving team lead: confirm the handoff is specific enough to execute without guesswork.")
	}
	return items
}

func starterResearchStillToConfirm(source string) []string {
	items := []string{
		"Whether approval is already explicit or still only implied in discussion.",
		"Whether the rollback note belongs in the same artifact or a linked operational surface.",
	}
	if containsAny(normalizeName(source), []string{"notification", "notifications"}) {
		items = append(items, "Which notification path or user case should be the very first implementation slice.")
	}
	return items
}

func dedupeStrings(items []string) []string {
	out := make([]string, 0, len(items))
	seen := map[string]bool{}
	for _, item := range items {
		clean := strings.TrimSpace(item)
		if clean == "" {
			continue
		}
		key := normalizeName(clean)
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, clean)
	}
	return out
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func deriveStarterTitle(defaultName, source, packID string) string {
	if packID == "travel-plan" {
		ctx := parseTravelStarterContext(source)
		if ctx.DayCount > 0 && strings.TrimSpace(ctx.Destination) != "" {
			return fmt.Sprintf("%d Day %s Trip", ctx.DayCount, ctx.Destination)
		}
		if strings.TrimSpace(ctx.Destination) != "" {
			return fmt.Sprintf("Trip To %s", ctx.Destination)
		}
		source = ctx.BaseSource
	} else {
		source = primaryStarterTitleSource(source)
	}
	cleaned := titleCase(cleanForTitle(source))
	if strings.TrimSpace(cleaned) == "" {
		return defaultName
	}
	return cleaned
}

func primaryStarterTitleSource(source string) string {
	trimmed := strings.TrimSpace(source)
	lower := strings.ToLower(trimmed)
	if idx := strings.Index(lower, " scope: "); idx >= 0 {
		return strings.TrimSpace(strings.TrimRight(trimmed[:idx], ".!? "))
	}
	return trimmed
}

type travelStarterContext struct {
	BaseSource  string
	Destination string
	DayCount    int
	MustDos     []string
	Missing     []string
}

func parseTravelStarterContext(source string) travelStarterContext {
	ctx := travelStarterContext{
		BaseSource: primaryStarterTitleSource(source),
	}
	ctx.DayCount = extractTravelDayCount(ctx.BaseSource)
	ctx.Destination = extractTravelDestination(ctx.BaseSource)
	ctx.MustDos = extractTravelMustDos(source)
	ctx.Missing = travelMissingDimensions(source)
	return ctx
}

func travelMissingDimensions(source string) []string {
	return missingScopeDimensions(source, travelScopeDimensions)
}

func extractTravelDayCount(source string) int {
	re := regexp.MustCompile(`(?i)\b(\d+)\s*[- ]day\b`)
	match := re.FindStringSubmatch(source)
	if len(match) < 2 {
		return 0
	}
	count, err := strconv.Atoi(match[1])
	if err != nil || count <= 0 {
		return 0
	}
	return count
}

func extractTravelDestination(source string) string {
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)\b\d+\s*[- ]day\b\s+(.+?)\s+trip\b`),
		regexp.MustCompile(`(?i)\btrip\s+to\s+(.+?)(?:\s+for\b|\s+with\b|[,.]|$)`),
	}
	for _, pattern := range patterns {
		match := pattern.FindStringSubmatch(source)
		if len(match) < 2 {
			continue
		}
		destination := titleCase(cleanForTitle(match[1]))
		if strings.TrimSpace(destination) != "" {
			return destination
		}
	}
	return ""
}

func extractTravelMustDos(source string) []string {
	mustDos := []string{}
	seen := map[string]bool{}
	pattern := regexp.MustCompile(`(?i)^(.*?)(?:\s+(?:is|are))?\s+(?:a\s+)?must[- ]?(?:do|see)s?$`)
	fragments := strings.FieldsFunc(source, func(r rune) bool {
		return r == ',' || r == '\n' || r == '\r' || r == ';'
	})
	for _, fragment := range fragments {
		trimmed := strings.TrimSpace(fragment)
		normalized := normalizeName(trimmed)
		if !containsAny(normalized, []string{"must do", "must see", "must dos", "must sees"}) {
			continue
		}
		match := pattern.FindStringSubmatch(trimmed)
		if len(match) < 2 {
			continue
		}
		item := strings.TrimSpace(match[1])
		item = strings.Trim(item, " ,")
		item = strings.TrimPrefix(strings.ToLower(item), "that ")
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		key := normalizeName(item)
		if seen[key] {
			continue
		}
		seen[key] = true
		mustDos = append(mustDos, item)
	}
	return mustDos
}

func travelStillToConfirm(ctx travelStarterContext) []string {
	if len(ctx.Missing) > 0 {
		items := make([]string, 0, len(ctx.Missing))
		for _, missing := range ctx.Missing {
			items = append(items, titleCase(missing))
		}
		return items
	}
	return []string{"Nothing right now"}
}

func travelTaskList(ctx travelStarterContext) []string {
	if len(ctx.Missing) > 0 {
		items := make([]string, 0, len(ctx.Missing))
		for _, missing := range ctx.Missing {
			items = append(items, "Confirm "+strings.ToLower(missing)+".")
		}
		return items
	}
	return []string{"Nothing right now"}
}

func cleanForTitle(value string) string {
	var builder strings.Builder
	spacePending := false
	for _, r := range value {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			if spacePending && builder.Len() > 0 {
				builder.WriteRune(' ')
			}
			builder.WriteRune(r)
			spacePending = false
		default:
			spacePending = true
		}
	}
	return strings.TrimSpace(builder.String())
}

func normalizeFilename(value string) string {
	return strings.ReplaceAll(normalizeName(value), " ", "-")
}

func slugify(value string) string {
	slug := normalizeFilename(cleanForTitle(value))
	if slug == "" {
		return "jini-work"
	}
	return slug
}

func uniqueWorkDir(root, packID, title string) (string, error) {
	base := filepath.Join(root, fmt.Sprintf("%s-%s", packID, slugify(title)))
	candidate := base
	for index := 1; ; index++ {
		if _, err := os.Stat(candidate); errors.Is(err, os.ErrNotExist) {
			return candidate, nil
		}
		candidate = fmt.Sprintf("%s-%d", base, index)
		if index > 1000 {
			return "", fmt.Errorf("could not allocate work directory under %s", root)
		}
	}
}

func inferHealthFromState(state string) string {
	switch state {
	case "awaiting_verification":
		return "ready-to-verify"
	case "incident":
		return "active"
	default:
		return "ready-to-make"
	}
}

func runCheck(args []string, stdout, stderr io.Writer) int {
	if len(args) > 0 && exactCommandToken(args[0]) == "ship" {
		return runShipCheck(args[1:], stdout, stderr)
	}
	if len(args) > 0 && exactCommandToken(args[0]) == "competitor-watch" {
		return runCompetitorWatchCheck(args[1:], stdout, stderr)
	}
	summary, err := resolveSummary(args)
	if err != nil {
		fmt.Fprintf(stderr, "%v\n", err)
		return 1
	}

	renderCheck(stdout, summary)
	return 0
}

func runObserve(args []string, stdout, stderr io.Writer) int {
	summary, err := resolveSummary(nil)
	if err != nil {
		fmt.Fprintf(stderr, "%v\n", err)
		return 1
	}
	if len(args) == 0 || exactCommandToken(args[0]) == "status" {
		if err := scanExternalObservations(summary.Dir); err != nil {
			fmt.Fprintf(stderr, "Could not scan observed files: %v\n", err)
			return 1
		}
		renderExternalObservationStatus(stdout, summary.Dir)
		return 0
	}
	switch exactCommandToken(args[0]) {
	case "add":
		targetPath, connectorID, parseErr := parseObserveAddArgs(args[1:])
		if parseErr != nil {
			fmt.Fprintf(stderr, "%v\n", parseErr)
			fmt.Fprintln(stderr, "Usage: jini observe add [--connector github|jira|confluence|markdown] <external-file>")
			return 1
		}
		artifactPath := currentFeedbackArtifactPath(summary)
		item, err := addExternalObservation(summary.Dir, artifactPath, targetPath, connectorID)
		if err != nil {
			fmt.Fprintf(stderr, "Could not add observed file: %v\n", err)
			return 1
		}
		if err := scanExternalObservations(summary.Dir); err != nil {
			fmt.Fprintf(stderr, "Could not scan observed files: %v\n", err)
			return 1
		}
		fmt.Fprintln(stdout, "Observing external file")
		fmt.Fprintln(stdout, item.TargetPath)
		return 0
	case "scan":
		if err := scanExternalObservations(summary.Dir); err != nil {
			fmt.Fprintf(stderr, "Could not scan observed files: %v\n", err)
			return 1
		}
		renderExternalObservationStatus(stdout, summary.Dir)
		return 0
	default:
		fmt.Fprintf(stderr, "Unknown observe command %q.\n", args[0])
		fmt.Fprintln(stderr, "Try `jini observe`, `jini observe add [--connector github|jira|confluence|markdown] <external-file>`, or `jini observe scan`.")
		return 1
	}
}

func parseObserveAddArgs(args []string) (string, string, error) {
	if len(args) == 0 {
		return "", "", fmt.Errorf("missing external file to observe")
	}
	var targetPath string
	var connectorID string
	for index := 0; index < len(args); index++ {
		part := strings.TrimSpace(args[index])
		if part == "" {
			continue
		}
		if part == "--connector" {
			if index+1 >= len(args) || strings.TrimSpace(args[index+1]) == "" {
				return "", "", fmt.Errorf("missing connector value after --connector")
			}
			connectorID = strings.TrimSpace(args[index+1])
			index++
			continue
		}
		if strings.HasPrefix(part, "--connector=") {
			connectorID = strings.TrimSpace(strings.TrimPrefix(part, "--connector="))
			continue
		}
		if targetPath == "" {
			targetPath = part
			continue
		}
		return "", "", fmt.Errorf("too many arguments for observe add")
	}
	if strings.TrimSpace(targetPath) == "" {
		return "", "", fmt.Errorf("missing external file to observe")
	}
	if connectorID != "" && normalizeConnectorID(connectorID) == "" {
		return "", "", fmt.Errorf("unsupported connector %q", connectorID)
	}
	return targetPath, normalizeConnectorID(connectorID), nil
}

func runOpen(args []string, stdout, stderr io.Writer) int {
	summary, err := resolveSummary(nil)
	if err != nil {
		fmt.Fprintf(stderr, "%v\n", err)
		return 1
	}

	if len(args) == 0 {
		renderOpenShelf(stdout, summary)
		return 0
	}

	item, err := resolveOpenItem(summary, args[0])
	if err != nil {
		fmt.Fprintf(stderr, "%v\n", err)
		return 1
	}
	return openArtifactItem(summary, item, stdout, stderr)
}

func runContinue(stdout, stderr io.Writer) int {
	summary, err := resolveSummary(nil)
	if err != nil {
		fmt.Fprintf(stderr, "%v\n", err)
		return 1
	}
	if renderNextContinuation(stdout, summary) {
		return 0
	}
	if renderFocusedContinuation(stdout, summary) {
		return 0
	}
	renderCheck(stdout, summary)
	return 0
}

func resolveSummary(args []string) (*workSummary, error) {
	if len(args) > 0 && strings.TrimSpace(args[0]) != "" {
		return loadWorkSummary(args[0], nil)
	}
	current, err := loadCurrentWork()
	if err != nil {
		return nil, err
	}
	if current == nil {
		return nil, errors.New("Nothing is in progress yet. Run `jini` to start something.")
	}
	summary, err := loadWorkSummary(current.PackDir, current)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			_ = clearCurrentWork()
			return nil, errors.New("Remembered work is no longer available. Run `jini` to start something.")
		}
		return nil, errors.New("Could not load current work. Run `jini` to start again or pass a valid work directory.")
	}
	return summary, nil
}

func loadCurrentWork() (*currentWork, error) {
	path := filepath.Join(sessionStateRoot(), "current-work.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}

	var payload currentWork
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, err
	}
	if strings.TrimSpace(payload.PackDir) == "" {
		return nil, nil
	}
	return &payload, nil
}

func sessionStateRoot() string {
	if override := strings.TrimSpace(os.Getenv("JINI_STATE_DIR")); override != "" {
		return override
	}
	cwd, err := os.Getwd()
	if err != nil {
		return ".jini"
	}
	return filepath.Join(cwd, ".jini")
}

func clearCurrentWork() error {
	err := os.Remove(filepath.Join(sessionStateRoot(), "current-work.json"))
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}

func loadWorkSummary(dir string, current *currentWork) (*workSummary, error) {
	resolved, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}
	info, err := os.Stat(resolved)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("%s is not a work directory", resolved)
	}

	workUnitPath := filepath.Join(resolved, "work-unit.yaml")
	workUnit, _ := parseSimpleYAML(workUnitPath)

	packID := strings.TrimSpace(workUnit["pack_id"])
	if packID == "" && current != nil {
		packID = current.PackID
	}
	if packID == "" {
		packID = inferPackID(resolved)
	}

	title := strings.TrimSpace(workUnit["title"])
	if title == "" && current != nil {
		title = current.Title
	}
	if title == "" {
		title = filepath.Base(resolved)
	}

	state := strings.TrimSpace(workUnit["current_state"])
	if state == "" && current != nil {
		state = current.State
	}

	views := collectViews(resolved, packID)
	exports := collectExports(resolved)
	details := collectDetails(resolved)
	inputs := loadInputItems(resolved, packID)
	source := sourceFromInputItems(inputs)
	missing := inferMissing(state, details, packID, source)
	uncertain := inferUncertain(packID, missing, source)
	route := loadWorkRoute(resolved)
	_ = scanExternalObservations(resolved)
	route = loadWorkRoute(resolved)

	summary := &workSummary{
		Dir:    resolved,
		PackID: packID,
		WorkUnitID: firstNonEmpty(strings.TrimSpace(workUnit["work_unit_id"]), func() string {
			if current != nil {
				return current.WorkUnitID
			}
			return ""
		}()),
		Title:              title,
		WorkingWith:        workingWithLabelForSavedRoute(route),
		ModelLabel:         strings.TrimSpace(route.ModelLabel),
		ModelReason:        strings.TrimSpace(route.ModelReason),
		ModelFeedback:      strings.TrimSpace(route.ModelFeedback),
		EffortLevel:        strings.TrimSpace(route.EffortLevel),
		VerificationLevel:  strings.TrimSpace(route.VerificationLevel),
		VerificationReason: strings.TrimSpace(route.VerificationReason),
		RoutePolicy:        strings.TrimSpace(route.RoutePolicy),
		AutoMode:           autoModePolicyValue(route.AutoMode),
		RouteReason:        strings.TrimSpace(route.Reason),
		ContinuityReason:   strings.TrimSpace(route.ContinuityReason),
		CLIHandoffSummary:  formatCLIHandoffReceiptSummary(route.CLIHandoffReceipt),
		State:              state,
		Views:              views,
		Exports:            exports,
		Details:            details,
		Missing:            missing,
		Uncertain:          uncertain,
		Using:              inferUsing(packID),
		Doing:              inferDoing(packID, state),
		Progress:           inferProgress(state),
		SafeToDo:           "Nothing has been sent yet. You can review before sharing.",
	}
	summary.NextStep = inferNextStep(packID, summary.Views)
	summary.Thread = buildWorkThread(summary, inputs, loadThreadState(resolved, summary))
	refreshSemanticEnvelopeForSummary(summary)
	return summary, nil
}

func parseSimpleYAML(path string) (map[string]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return map[string]string{}, err
	}
	out := map[string]string{}
	for _, raw := range strings.Split(string(data), "\n") {
		line := strings.TrimRight(raw, "\r")
		if strings.TrimSpace(line) == "" {
			continue
		}
		if strings.HasPrefix(line, " ") || strings.HasPrefix(line, "-") || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		value = strings.Trim(value, `"'`)
		out[key] = value
	}
	return out, nil
}

func inferPackID(dir string) string {
	base := filepath.Base(dir)
	base = strings.TrimPrefix(base, "jini-sim-")
	switch {
	case strings.Contains(base, "prd"):
		return "research-prd"
	case strings.Contains(base, "meeting"):
		return "meeting-followup"
	case strings.Contains(base, "vendor"):
		return "vendor-selection"
	case strings.Contains(base, "incident"):
		return "incident-response"
	case strings.Contains(base, "trip") || strings.Contains(base, "travel"):
		return "travel-plan"
	default:
		return base
	}
}

func collectViews(root, packID string) []catalogItem {
	viewDir := filepath.Join(root, "views")
	entries, err := os.ReadDir(viewDir)
	if err != nil {
		return synthesizeViews(root, packID)
	}

	items := []catalogItem{}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}
		stem := strings.TrimSuffix(entry.Name(), ".md")
		item := viewCatalogItem(packID, stem, filepath.Join(viewDir, entry.Name()))
		items = append(items, item)
	}
	items = append(items, synthesizedPackViews(root, packID)...)
	return prioritizeViews(packID, dedupeItems(items))
}

func synthesizeViews(root, packID string) []catalogItem {
	return prioritizeViews(packID, synthesizedPackViews(root, packID))
}

func synthesizedPackViews(root, packID string) []catalogItem {
	return starterSynthesizedViews(root, packID)
}

func viewCatalogItem(packID, stem, path string) catalogItem {
	if item, ok := starterViewForStem(packID, stem, path); ok {
		return item
	}
	return catalogItem{
		ID:      normalizeName(stem),
		Label:   titleCase(stem),
		Path:    path,
		Aliases: []string{stem},
	}
}

func collectExports(root string) []catalogItem {
	candidates := []catalogItem{
		{ID: "markdown-wiki", Label: "Markdown Wiki", Path: filepath.Join(root, "exports", "wiki", "markdown", "overview.md"), Aliases: []string{"markdown", "wiki"}},
		{ID: "github-issues", Label: "GitHub Issues", Path: filepath.Join(root, "exports", "issues", "github", "README.md"), Aliases: []string{"github"}},
		{ID: "jira-issues", Label: "Jira Issues", Path: filepath.Join(root, "exports", "issues", "jira", "README.md"), Aliases: []string{"jira"}},
	}
	items := []catalogItem{}
	for _, item := range candidates {
		if fileExists(item.Path) {
			items = append(items, item)
		}
	}
	return items
}

func collectDetails(root string) []catalogItem {
	artifactDir := filepath.Join(root, "artifacts")
	entries, err := os.ReadDir(artifactDir)
	if err != nil {
		return nil
	}

	items := []catalogItem{}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yaml") {
			continue
		}
		path := filepath.Join(artifactDir, entry.Name())
		doc, _ := parseSimpleYAML(path)
		artifactType := firstNonEmpty(doc["artifact_type"], titleCase(strings.TrimSuffix(entry.Name(), ".yaml")))
		label := artifactType
		id := normalizeName(artifactType)
		if artifactType == "Tasks" {
			label = "Tasks Record"
			id = "tasks-record"
		}
		items = append(items, catalogItem{
			ID:      id,
			Label:   label,
			Path:    path,
			Aliases: []string{id, strings.TrimSuffix(entry.Name(), ".yaml")},
		})
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Label < items[j].Label })
	return items
}

func sourceFromInputItems(items []inputItem) string {
	parts := []string{}
	for _, item := range items {
		switch item.Kind {
		case "text", "clarification", "derived":
			text := strings.TrimSpace(item.OriginRef)
			if text == "" {
				text = strings.TrimSpace(item.Preview)
			}
			if text != "" {
				parts = append(parts, text)
			}
		}
	}
	return strings.Join(parts, ". ")
}

func inferMissing(state string, details []catalogItem, packID, source string) []string {
	available := map[string]bool{}
	for _, item := range details {
		available[item.Label] = true
	}
	missing := []string{}
	switch state {
	case "awaiting_verification", "operational", "incident":
		if !available["Approval"] {
			missing = append(missing, "Approval")
		}
		if state != "awaiting_verification" && !available["Evidence"] {
			missing = append(missing, "Evidence")
		}
	case "decided", "in_make":
		missing = append(missing, starterMissing(packID, state, details, source)...)
	}
	return missing
}

func inferUncertain(packID string, missing []string, source string) []string {
	return starterUncertain(packID, missing, source)
}

func inferUsing(packID string) string {
	return starterWorkingWith(packID)
}

func inferDoing(packID, state string) string {
	return starterDoing(packID, state)
}

func inferProgress(state string) string {
	switch state {
	case "awaiting_verification", "incident":
		return "3 of 4 steps done"
	case "decided", "in_make":
		return "2 of 4 steps done"
	case "modeled", "probed":
		return "1 of 4 steps done"
	case "operational", "retired":
		return "4 of 4 steps done"
	default:
		return "1 of 4 steps done"
	}
}

func inferNextStep(packID string, views []catalogItem) string {
	return starterNextStep(packID, views)
}

func prioritizeViews(packID string, items []catalogItem) []catalogItem {
	return starterPrioritizeViews(packID, items)
}

func resolveOpenItem(summary *workSummary, name string) (*catalogItem, error) {
	normalized := normalizeName(name)
	for _, item := range append(append([]catalogItem{}, summary.Views...), append(summary.Exports, summary.Details...)...) {
		candidates := append([]string{item.ID, item.Label}, item.Aliases...)
		for _, candidate := range candidates {
			if normalizeName(candidate) == normalized {
				return &item, nil
			}
		}
	}
	return nil, fmt.Errorf("Artifact not found: %q. Run `jini open` to list available artifacts.", name)
}

func renderNewWorkLauncher(w io.Writer) {
	fmt.Fprintln(w, "Jini")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Describe the task. Jini can answer, edit files, route to a configured tool, or ask one clarification.")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Examples:")
	fmt.Fprintln(w, "- Add a line to the matching .txt file in this folder")
	fmt.Fprintln(w, "- Turn meeting notes into something I can send")
	fmt.Fprintln(w, "- Check whether a plan is ready to hand off")
	fmt.Fprintln(w, "- Plan a 7 day Paris trip for two adults in October")
	fmt.Fprintln(w, "- Compare these vendors and recommend one")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Jini asks before sending, booking, committing, or running destructive changes.")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Type `help` for examples and commands.")
}

func renderPublicCommandInventory(w io.Writer) {
	fmt.Fprintln(w, "Jini")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "- `jini`")
	fmt.Fprintln(w, "- `jini <request>`")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Examples:")
	fmt.Fprintln(w, "- `jini \"what is the capital of france?\"`")
	fmt.Fprintln(w, "- `jini \"add a line to the matching .txt file in this folder\"`")
	fmt.Fprintln(w, "- `jini \"review this repo\"`")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Essential commands:")
	fmt.Fprintln(w, "- `jini status`, `jini continue`, `jini open`")
	fmt.Fprintln(w, "- `jini route`, `jini doctor`")
	fmt.Fprintln(w, "- `jini memory inspect`, `jini memory off`, `jini memory forget`")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Setup:")
	fmt.Fprintln(w, "- `jini route help`")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "More: `jini admin help`")
}

func renderAdminCommandInventory(w io.Writer) {
	fmt.Fprintln(w, "Admin and developer command inventory")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "- jini provider doctor")
	fmt.Fprintln(w, "- jini check ship")
	fmt.Fprintln(w, "- jini check competitor-watch")
	fmt.Fprintln(w, "- jini observe status")
	fmt.Fprintln(w, "- jini observe add <path>")
	fmt.Fprintln(w, "- jini open <artifact>")
	fmt.Fprintln(w, "- jini run")
	fmt.Fprintln(w, "- jini scorecard-gate")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Admin commands stay intentionally narrow. Broader release plumbing should graduate through `jini publish-readiness` before it becomes public.")
}

func renderNewWorkPrompt(w io.Writer) {
	fmt.Fprintln(w, "Jini")
}

func renderNoCurrentWorkStatus(w io.Writer) {
	fmt.Fprintln(w, "No current work yet.")
	fmt.Fprintln(w, "Describe the task, or run `jini commands`.")
}

func renderNoInitRequired(w io.Writer) {
	fmt.Fprintln(w, "No init step is required before first value.")
	fmt.Fprintln(w, "Describe a task. Jini can answer, edit files, route to a configured tool, or save the work state.")
}

func renderNoCurrentMemoryStatus(w io.Writer) {
	renderUserContextMemorySummary(w, "none")
}

func renderSafePermissionsStatus(w io.Writer) {
	fmt.Fprintln(w, "Permissions")
	fmt.Fprintln(w, "Nothing has been sent, published, booked, or changed.")
	fmt.Fprintln(w, "Jini asks before risky external actions and keeps low-risk drafts reviewable first.")
}

func renderRouteCostStatus(w io.Writer) {
	device := currentDeviceProfile()
	route := detectRoute()
	provider := detectLegacyProvider()
	if route.Active {
		provider = route.Provider
	}
	fmt.Fprintln(w, "Route and cost")
	fmt.Fprintf(w, "Current route: %s. Readiness: %s.\n", workingWithLabel(provider), routeReadinessLabel(provider))
	if route.Active && cliHandoffMode(route.ToolMode) {
		renderCLIHandoffRouteInspection(w, provider)
	}
	fmt.Fprintln(w, "Token posture: compact context first; avoid transcript replay unless quality or safety requires it.")
	fmt.Fprintln(w, "Continuity: offline and online work stitch into the same session.")
	fmt.Fprintf(w, "Route inputs: device %s; battery, thermal, and CLI throttle levels affect switching.\n", firstNonEmpty(device.DeviceClass, "unknown"))
	fmt.Fprintln(w, "Least-expense capable route is the default; use `doctor` to inspect setup or `auto` to restore automatic routing.")
}

func renderCLIHandoffRouteInspection(w io.Writer, provider providerConfig) {
	status := "configured"
	if strings.TrimSpace(provider.Status) != "" && provider.Status != "ok" && len(provider.Missing) > 0 {
		status = "needs setup"
	}
	fmt.Fprintf(w, "CLI handoff: %s\n", status)
	if status == "needs setup" {
		fmt.Fprintf(w, "Setup: %s\n", shipCLIHandoffSetupCategory(provider.Missing))
	}
	fmt.Fprintln(w, "Provider alias: disabled; Jini invokes the CLI subprocess only when running work.")
}

func formatCLIHandoffReceiptSummary(receipt *cliHandoffReceipt) []string {
	if receipt == nil {
		return nil
	}
	lines := []string{}
	label := firstNonEmpty(strings.TrimSpace(receipt.Label), strings.TrimSpace(receipt.Mode), "CLI handoff")
	lines = append(lines, label)
	status := "completed"
	if receipt.ExitStatus != 0 {
		status = "failed"
	}
	lines = append(lines, "Status: "+status)
	lines = append(lines, fmt.Sprintf(
		"Exit %d in %dms; prompt %d chars, stdout %d chars, stderr %d chars.",
		receipt.ExitStatus,
		receipt.DurationMS,
		receipt.PromptChars,
		receipt.StdoutChars,
		receipt.StderrChars,
	))
	if strings.TrimSpace(receipt.CompletedAt) != "" {
		lines = append(lines, "Completed: "+strings.TrimSpace(receipt.CompletedAt))
	}
	return lines
}

func renderCLIHandoffReceiptSummary(w io.Writer, lines []string) {
	if len(lines) == 0 {
		return
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Last CLI handoff")
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			fmt.Fprintln(w, line)
		}
	}
}

func routeReadinessLabel(provider providerConfig) string {
	return firstNonEmpty(strings.TrimSpace(provider.Status), "unknown")
}

func renderRouteDecisionCard(w io.Writer, request providerGenerationRequest, decision routeDecision) {
	if !decision.Active {
		return
	}
	features := classifyRouteFeatures(request)
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Jini will start with")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Tool")
	fmt.Fprintln(w, firstNonEmpty(strings.TrimSpace(decision.ToolLabel), "Not decided yet"))
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Provider")
	fmt.Fprintln(w, firstNonEmpty(strings.TrimSpace(decision.Provider.Label), "Local preview"))
	if strings.TrimSpace(decision.RoutePolicy) != "" {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "How chosen")
		fmt.Fprintln(w, decision.RoutePolicy)
	}
	renderAutoModePolicy(w, decision.AutoMode)
	if strings.TrimSpace(decision.ModelLabel) != "" {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Model")
		fmt.Fprintln(w, decision.ModelLabel)
	}
	if strings.TrimSpace(decision.ModelReason) != "" {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Why this model")
		fmt.Fprintln(w, decision.ModelReason)
	}
	if strings.TrimSpace(decision.EffortLevel) != "" {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Effort level")
		fmt.Fprintln(w, titleCase(decision.EffortLevel))
	}
	if strings.TrimSpace(decision.VerificationLevel) != "" {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Verification")
		fmt.Fprintln(w, decision.VerificationLevel)
	}
	if strings.TrimSpace(decision.VerificationReason) != "" {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Why this verification")
		fmt.Fprintln(w, decision.VerificationReason)
	}
	if strings.TrimSpace(decision.Reason) != "" {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Why this route")
		fmt.Fprintln(w, decision.Reason)
	}
	if strings.TrimSpace(decision.ContinuityReason) != "" {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Continuity")
		fmt.Fprintln(w, decision.ContinuityReason)
	}
	if lines := freshLocalMultimodalLearningViewLines(features); len(lines) > 0 {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Multimodal learning")
		for _, line := range lines {
			fmt.Fprintf(w, "- %s\n", line)
		}
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Want a different route?")
	fmt.Fprintln(w, "Type `Claude`, `Bedrock`, `Azure`, `Local`, `Preview`, or `Auto`.")
	fmt.Fprintln(w)
}

func renderAutoModePolicy(w io.Writer, policy autoModePolicy) {
	if !autoModePolicyEnabled(policy) {
		return
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Auto mode")
	if strings.TrimSpace(policy.FrameworkSwitching) != "" {
		fmt.Fprintf(w, "Frameworks: %s\n", policy.FrameworkSwitching)
	}
	if strings.TrimSpace(policy.ModelSwitching) != "" {
		fmt.Fprintf(w, "Models: %s\n", policy.ModelSwitching)
	}
	if strings.TrimSpace(policy.SpeedSwitching) != "" {
		fmt.Fprintf(w, "Speed: %s\n", policy.SpeedSwitching)
	}
	if strings.TrimSpace(policy.UserApprovalMode) != "" {
		fmt.Fprintf(w, "Approvals: %s\n", policy.UserApprovalMode)
	}
}

func maybeHandleProviderSetupIntent(raw string, scanner *bufio.Scanner, stdout, stderr io.Writer) (bool, int) {
	switch normalizeCommandName(raw) {
	case "claude":
		return true, runProviderSetupWizard("claude-api", scanner, stdout, stderr)
	case "bedrock":
		return true, runProviderSetupWizard("bedrock-sonnet", scanner, stdout, stderr)
	case "chatgpt":
		return true, runProviderSetupWizard("chatgpt", scanner, stdout, stderr)
	case "codex":
		return true, runProviderSetupWizard("codex", scanner, stdout, stderr)
	case "azure":
		return true, runProviderSetupWizard("azure-openai", scanner, stdout, stderr)
	case "local":
		return true, runProviderSetupWizard("local-slm", scanner, stdout, stderr)
	case "auto":
		return true, runProviderSetupWizard("auto", scanner, stdout, stderr)
	case "preview":
		return true, runProviderSetupWizard("local-preview", scanner, stdout, stderr)
	default:
		return false, 0
	}
}

func runProviderSetupWizard(mode string, scanner *bufio.Scanner, stdout, stderr io.Writer) int {
	fmt.Fprintln(stdout)
	if cliHandoffMode(mode) {
		provider := detectCLIHandoffProvider(mode)
		if provider.Status == "ok" {
			fmt.Fprintf(stdout, "%s is ready.\n", provider.Label)
			fmt.Fprintf(stdout, "Run `jini route set %s` to use it.\n", mode)
			return 0
		}
		fmt.Fprintln(stderr, provider.Label+" CLI handoff is not ready.")
		for _, missing := range provider.Missing {
			fmt.Fprintf(stderr, "- %s\n", missing)
		}
		return 1
	}
	switch mode {
	case "claude-api":
		fmt.Fprintln(stdout, "Claude")
		fmt.Fprintln(stdout, "Paste your Anthropic API key. Jini will route this repo through Claude for code work.")
		key, ok := readPromptLine(scanner, stdout, "Anthropic API key")
		if !ok || strings.TrimSpace(key) == "" {
			fmt.Fprintln(stderr, "Claude API setup needs an API key.")
			return 1
		}
		model, ok := readPromptLine(scanner, stdout, "Model (press Enter for sonnet)")
		if !ok {
			return 1
		}
		if err := saveRouterSettings("claude-api"); err != nil {
			fmt.Fprintf(stderr, "Could not save Claude API route: %v\n", err)
			return 1
		}
		if err := saveProviderSettings(map[string]string{
			"JINI_PROVIDER":     "claude",
			"ANTHROPIC_API_KEY": key,
			"JINI_MODEL":        firstNonEmpty(model, "sonnet"),
		}); err != nil {
			fmt.Fprintf(stderr, "Could not save Claude setup: %v\n", err)
			return 1
		}
	case "bedrock-sonnet":
		fmt.Fprintln(stdout, "Bedrock")
		settings, ok := readBedrockSetupInputs(scanner, stdout, stderr, false)
		if !ok {
			return 1
		}
		if err := saveRouterSettings("bedrock-sonnet"); err != nil {
			fmt.Fprintf(stderr, "Could not save Bedrock Sonnet route: %v\n", err)
			return 1
		}
		if err := saveProviderSettings(settings); err != nil {
			fmt.Fprintf(stderr, "Could not save Bedrock setup: %v\n", err)
			return 1
		}
	case "chatgpt", "azure-code", "azure-openai":
		targetLabel := map[string]string{
			"chatgpt":      "Azure",
			"azure-code":   "Azure",
			"azure-openai": "Azure",
		}[mode]
		fmt.Fprintln(stdout, targetLabel)
		endpoint, ok := readPromptLine(scanner, stdout, "Azure endpoint")
		if !ok || strings.TrimSpace(endpoint) == "" {
			fmt.Fprintln(stderr, targetLabel+" setup needs an endpoint.")
			return 1
		}
		apiKey, ok := readPromptLine(scanner, stdout, "Azure API key")
		if !ok || strings.TrimSpace(apiKey) == "" {
			fmt.Fprintln(stderr, targetLabel+" setup needs an API key.")
			return 1
		}
		deployment, ok := readPromptLine(scanner, stdout, "Azure deployment name")
		if !ok || strings.TrimSpace(deployment) == "" {
			fmt.Fprintln(stderr, targetLabel+" setup needs a deployment name.")
			return 1
		}
		if err := saveRouterSettings(mode); err != nil {
			fmt.Fprintf(stderr, "Could not save %s route: %v\n", targetLabel, err)
			return 1
		}
		if err := saveProviderSettings(map[string]string{
			"JINI_PROVIDER":            "azure-openai",
			"AZURE_OPENAI_ENDPOINT":    endpoint,
			"AZURE_OPENAI_API_KEY":     apiKey,
			"AZURE_OPENAI_DEPLOYMENT":  deployment,
			"AZURE_OPENAI_API_VERSION": "2024-10-21",
		}); err != nil {
			fmt.Fprintf(stderr, "Could not save Azure setup: %v\n", err)
			return 1
		}
	case "claude":
		fmt.Fprintln(stdout, "Claude")
		fmt.Fprintln(stdout, "Paste your Anthropic API key. Jini will save it only in this repo's .jini folder.")
		key, ok := readPromptLine(scanner, stdout, "Anthropic API key")
		if !ok || strings.TrimSpace(key) == "" {
			fmt.Fprintln(stderr, "Claude setup needs an API key.")
			return 1
		}
		model, ok := readPromptLine(scanner, stdout, "Model (press Enter for sonnet)")
		if !ok {
			return 1
		}
		if err := clearRouterSettings(); err != nil {
			fmt.Fprintf(stderr, "Could not clear saved tool route: %v\n", err)
			return 1
		}
		if err := saveProviderSettings(map[string]string{
			"JINI_PROVIDER":     "claude",
			"ANTHROPIC_API_KEY": key,
			"JINI_MODEL":        firstNonEmpty(model, "sonnet"),
		}); err != nil {
			fmt.Fprintf(stderr, "Could not save Claude setup: %v\n", err)
			return 1
		}
	case "bedrock":
		fmt.Fprintln(stdout, "Bedrock")
		settings, ok := readBedrockSetupInputs(scanner, stdout, stderr, true)
		if !ok {
			return 1
		}
		if err := clearRouterSettings(); err != nil {
			fmt.Fprintf(stderr, "Could not clear saved tool route: %v\n", err)
			return 1
		}
		if err := saveProviderSettings(settings); err != nil {
			fmt.Fprintf(stderr, "Could not save Bedrock setup: %v\n", err)
			return 1
		}
	case "auto":
		fmt.Fprintln(stdout, "Auto")
		if err := saveRouterSettings("auto"); err != nil {
			fmt.Fprintf(stderr, "Could not save auto tool mode: %v\n", err)
			return 1
		}
		if err := saveProviderSettings(map[string]string{
			"JINI_PROVIDER": "auto",
			"JINI_MODEL":    "auto",
		}); err != nil {
			fmt.Fprintf(stderr, "Could not save auto mode: %v\n", err)
			return 1
		}
	case "local-slm":
		fmt.Fprintln(stdout, "Local")
		fmt.Fprintln(stdout, "Point Jini at an OpenAI-compatible local model server. Jini will keep the exact endpoint in this repo's .jini folder.")
		endpoint, ok := readPromptLine(scanner, stdout, "Local SLM endpoint")
		if !ok || strings.TrimSpace(endpoint) == "" {
			fmt.Fprintln(stderr, "Local SLM setup needs an endpoint.")
			return 1
		}
		model, ok := readPromptLine(scanner, stdout, "Default local model")
		if !ok || strings.TrimSpace(model) == "" {
			fmt.Fprintln(stderr, "Local SLM setup needs a default model.")
			return 1
		}
		apiKey, ok := readPromptLine(scanner, stdout, "Local SLM API key (press Enter if not needed)")
		if !ok {
			return 1
		}
		fastModel, ok := readPromptLine(scanner, stdout, "Fast profile model (press Enter to reuse the default)")
		if !ok {
			return 1
		}
		workhorseModel, ok := readPromptLine(scanner, stdout, "Workhorse profile model (press Enter to reuse the default)")
		if !ok {
			return 1
		}
		deepModel, ok := readPromptLine(scanner, stdout, "Deep profile model (press Enter to reuse the default)")
		if !ok {
			return 1
		}
		multimodalModel, ok := readPromptLine(scanner, stdout, "Multimodal profile model (press Enter to reuse the default)")
		if !ok {
			return 1
		}
		if err := saveRouterSettings("auto"); err != nil {
			fmt.Fprintf(stderr, "Could not save Local SLM route mode: %v\n", err)
			return 1
		}
		if err := saveProviderSettings(map[string]string{
			"JINI_PROVIDER":                   "local-slm",
			"JINI_MODEL":                      "auto",
			"JINI_LOCAL_SLM_ENDPOINT":         endpoint,
			"JINI_LOCAL_SLM_MODEL":            model,
			"JINI_LOCAL_SLM_API_KEY":          apiKey,
			"JINI_LOCAL_SLM_FAST_MODEL":       firstNonEmpty(strings.TrimSpace(fastModel), strings.TrimSpace(model)),
			"JINI_LOCAL_SLM_WORKHORSE_MODEL":  firstNonEmpty(strings.TrimSpace(workhorseModel), strings.TrimSpace(model)),
			"JINI_LOCAL_SLM_DEEP_MODEL":       firstNonEmpty(strings.TrimSpace(deepModel), strings.TrimSpace(model)),
			"JINI_LOCAL_SLM_MULTIMODAL_MODEL": firstNonEmpty(strings.TrimSpace(multimodalModel), strings.TrimSpace(model)),
		}); err != nil {
			fmt.Fprintf(stderr, "Could not save Local SLM setup: %v\n", err)
			return 1
		}
		_ = saveLocalRuntimeCapabilities(benchmarkLocalRuntimeCapabilities(context.Background()))
	case "local-preview":
		fmt.Fprintln(stdout, "Preview")
		if err := saveRouterSettings("local-preview"); err != nil {
			fmt.Fprintf(stderr, "Could not save local preview route: %v\n", err)
			return 1
		}
		if err := saveProviderSettings(map[string]string{
			"JINI_PROVIDER": "local-preview",
			"JINI_MODEL":    "auto",
		}); err != nil {
			fmt.Fprintf(stderr, "Could not save local preview mode: %v\n", err)
			return 1
		}
	}

	provider := detectProvider()
	fmt.Fprintln(stdout)
	if provider.Status == "ok" {
		if mode == "local-slm" {
			fmt.Fprintf(stdout, "Setup saved. Working with %s / %s.\n", provider.Label, firstNonEmpty(strings.TrimSpace(configValue("JINI_LOCAL_SLM_MODEL")), "default local model"))
			return 0
		}
		fmt.Fprintf(stdout, "Setup saved. Working with %s.\n", workingWithLabel(provider))
		return 0
	}
	fmt.Fprintf(stderr, "Setup is still incomplete. Missing: %s\n", strings.Join(provider.Missing, ", "))
	return 1
}

func readBedrockSetupInputs(scanner *bufio.Scanner, stdout, stderr io.Writer, promptForModel bool) (map[string]string, bool) {
	region, ok := readPromptLine(scanner, stdout, "AWS region (press Enter for us-east-1)")
	if !ok {
		return nil, false
	}
	profile, ok := readPromptLine(scanner, stdout, "AWS profile name (press Enter to use access keys)")
	if !ok {
		return nil, false
	}

	values := map[string]string{
		"JINI_PROVIDER": "bedrock",
		"AWS_REGION":    firstNonEmpty(region, "us-east-1"),
		"JINI_MODEL":    "sonnet-4.6",
	}
	if strings.TrimSpace(profile) != "" {
		values["AWS_PROFILE"] = profile
		values["AWS_ACCESS_KEY_ID"] = ""
		values["AWS_SECRET_ACCESS_KEY"] = ""
		values["AWS_SESSION_TOKEN"] = ""
	} else {
		accessKeyID, ok := readPromptLine(scanner, stdout, "AWS access key ID")
		if !ok || strings.TrimSpace(accessKeyID) == "" {
			fmt.Fprintln(stderr, "Bedrock setup needs AWS profile or AWS access keys.")
			return nil, false
		}
		secretAccessKey, ok := readPromptLine(scanner, stdout, "AWS secret access key")
		if !ok || strings.TrimSpace(secretAccessKey) == "" {
			fmt.Fprintln(stderr, "Bedrock setup needs AWS secret access key.")
			return nil, false
		}
		sessionToken, ok := readPromptLine(scanner, stdout, "AWS session token (press Enter if not needed)")
		if !ok {
			return nil, false
		}
		values["AWS_PROFILE"] = ""
		values["AWS_ACCESS_KEY_ID"] = accessKeyID
		values["AWS_SECRET_ACCESS_KEY"] = secretAccessKey
		values["AWS_SESSION_TOKEN"] = sessionToken
	}
	if promptForModel {
		model, ok := readPromptLine(scanner, stdout, "Model (press Enter for sonnet-4.6)")
		if !ok {
			return nil, false
		}
		values["JINI_MODEL"] = firstNonEmpty(model, "sonnet-4.6")
	}
	return values, true
}

func renderFirstRunResult(w io.Writer, summary *workSummary) {
	fmt.Fprintln(w)

	item := firstResultItem(summary)
	if item == nil {
		fmt.Fprintln(w, "Work saved.")
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Work item")
		fmt.Fprintln(w)
		fmt.Fprintln(w, "No result file is ready yet. Jini still created the work record so the source context is not lost.")
		renderPostResultActions(w, summary, nil)
		return
	}

	fmt.Fprintln(w, "Artifact created.")
	fmt.Fprintln(w)
	fmt.Fprintln(w, item.Label)
	fmt.Fprintln(w)
	content, err := os.ReadFile(item.Path)
	if err != nil {
		fmt.Fprintf(w, "Could not read %q yet.\n", item.Label)
	} else {
		fmt.Fprint(w, strings.TrimSpace(string(content)))
		fmt.Fprintln(w)
	}
	fmt.Fprintln(w)
	renderPostResultActions(w, summary, item)
}

func renderDirectTaskStarted(w io.Writer, summary *workSummary, source string) {
	fmt.Fprintf(w, "Working on: %s\n", strings.TrimSpace(source))
	fmt.Fprintln(w, directTaskNextStepLine(summary, source))
	fmt.Fprintln(w, "Saved. Use `jini status` for full status or `jini open` for the draft.")
}

func directTaskNextStepLine(summary *workSummary, source string) string {
	if command := directTaskSuggestedCommand(source); command != "" {
		return fmt.Sprintf("Start with `%s`.", command)
	}
	if summary != nil && strings.TrimSpace(summary.NextStep) != "" {
		return "Next: " + strings.TrimRight(strings.TrimSpace(summary.NextStep), ".") + "."
	}
	return "Next: use `jini status` when you want the full state."
}

func directTaskSuggestedCommand(source string) string {
	normalized := normalizeName(source)
	if strings.Contains(normalized, "test") || strings.Contains(normalized, "tests") {
		return "make test"
	}
	if isRepoReviewDirectTask(source) {
		return "git status"
	}
	return ""
}

func currentFeedbackArtifactPath(summary *workSummary) string {
	if summary == nil {
		return ""
	}
	if item := currentArtifactItem(summary); item != nil {
		return item.Path
	}
	return ""
}

func firstResultItem(summary *workSummary) *catalogItem {
	if len(summary.Views) == 0 {
		return nil
	}
	return &summary.Views[0]
}

func renderPrimaryActionMenu(w io.Writer, summary *workSummary, heading, readyAction string) {
	if strings.TrimSpace(heading) != "" {
		fmt.Fprintln(w, heading)
		fmt.Fprintln(w)
	}
	fmt.Fprintln(w, "- Continue")
	fmt.Fprintf(w, "- %s\n", firstNonEmpty(strings.TrimSpace(readyAction), "Open"))
	fmt.Fprintln(w, "- Context")
	fmt.Fprintln(w, "- Missing")
	fmt.Fprintln(w, "- Expand")
	fmt.Fprintln(w, "Revision shortcuts: `Shorter`, `Executive`, `Checklist`.")
	if hasArtifactVersions(summary) {
		fmt.Fprintln(w, "- Versions")
		fmt.Fprintln(w, "- Undo")
	}
	fmt.Fprintln(w, "- Plan")
	fmt.Fprintln(w, "Paste a new request to start a different task.")
}

func renderCompactCurrentWorkChoices(w io.Writer, canSwitch bool) {
	fmt.Fprintln(w, "Actions")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "- Continue")
	fmt.Fprintln(w, "- Open")
	fmt.Fprintln(w, "- Missing")
	fmt.Fprintln(w, "- Plan")
	if canSwitch {
		fmt.Fprintln(w, "- Resume saved work")
	}
	fmt.Fprintln(w, "Paste a new request to start a different task.")
}

func renderWorkSwitchChoices(w io.Writer, includeStartNew bool) {
	if includeStartNew {
		fmt.Fprintln(w, "Type a number to open one, or paste a new request.")
		return
	}
	fmt.Fprintln(w, "Type a number or saved title.")
}

func renderSelectableWorkList(w io.Writer, heading string, active []*workSummary, includeStartNew bool) {
	if strings.TrimSpace(heading) != "" {
		fmt.Fprintln(w, heading)
		fmt.Fprintln(w)
	}
	for index, item := range active {
		fmt.Fprintf(w, "%d. %s\n", index+1, item.Title)
		fmt.Fprintf(w, "   Next: %s\n", item.NextStep)
	}
	fmt.Fprintln(w)
	renderWorkSwitchChoices(w, includeStartNew)
}

func renderArtifactFeedbackChoices(w io.Writer) {
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Feedback")
	fmt.Fprintln(w, "- Accept")
	fmt.Fprintln(w, "- Edit")
	fmt.Fprintln(w, "- Use")
	fmt.Fprintln(w, "- Share")
	fmt.Fprintln(w, "- Reject")
	fmt.Fprintln(w, "- Replace")
	fmt.Fprintln(w, "Advanced: `Upvote` or `Downvote`.")
}

func activeWorkOverviewLabels(active []*workSummary, currentTitle string) []string {
	labels := []string{}
	positions := map[string]int{}
	counts := map[string]int{}
	baseLabels := map[string]string{}
	currentKey := normalizeName(currentTitle)
	for _, item := range active {
		if item == nil {
			continue
		}
		label := firstNonEmpty(strings.TrimSpace(item.Title), "Untitled work")
		key := normalizeName(label)
		if index, ok := positions[key]; ok {
			counts[key]++
			if currentKey != "" && key == currentKey {
				labels[index] = fmt.Sprintf("%s (%d other saved)", baseLabels[key], counts[key])
			} else {
				labels[index] = fmt.Sprintf("%s (%d saved)", baseLabels[key], counts[key])
			}
			continue
		}
		positions[key] = len(labels)
		counts[key] = 1
		baseLabels[key] = label
		if currentKey != "" && key == currentKey {
			label = fmt.Sprintf("%s (another saved)", label)
		}
		labels = append(labels, label)
	}
	return labels
}

func renderOtherActiveWorkList(w io.Writer, active []*workSummary, currentTitle string, includeHint bool) {
	labels := activeWorkOverviewLabels(active, currentTitle)
	if len(labels) == 0 {
		return
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Other active work")
	for _, label := range labels {
		fmt.Fprintf(w, "- %s\n", label)
	}
	if includeHint {
		fmt.Fprintln(w, "Type a saved title to resume it.")
	}
}

func renderPostResultActions(w io.Writer, summary *workSummary, item *catalogItem) {
	renderPostResultContext(w, summary, item)
}

func renderNewWorkNoop(w io.Writer) {
	fmt.Fprintln(w, "Nothing to do yet.")
	fmt.Fprintln(w, "Paste the work when you're ready.")
}

func renderCurrentWorkNoop(w io.Writer) {
	fmt.Fprintln(w, "Nothing changed.")
	fmt.Fprintln(w, "Use `Continue`, `Open`, or paste a new request.")
}

func renderCurrentWorkMemoryStatus(w io.Writer, summary *workSummary) {
	if summary == nil {
		renderUserContextMemorySummary(w, "none")
		return
	}
	renderUserContextMemorySummary(w, summary.Title)
}

func renderPostResultContext(w io.Writer, summary *workSummary, item *catalogItem) {
	if summary == nil {
		return
	}
	workingWith := postResultWorkingWith(summary.Thread.WorkingWith)
	if len(workingWith) > 0 {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Working with")
		for _, value := range workingWith {
			fmt.Fprintf(w, "- %s\n", value)
		}
	}
	alsoReady := postResultAlsoReady(summary.Thread.ReadyNow, item)
	fmt.Fprintln(w)
	if len(alsoReady) > 0 {
		fmt.Fprintf(w, "Artifacts available: %s\n", strings.Join(alsoReady, ", "))
	}
	if len(summary.Thread.MultimodalLearning) > 0 {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Multimodal learning")
		for _, value := range summary.Thread.MultimodalLearning {
			fmt.Fprintf(w, "- %s\n", value)
		}
	}
	if item != nil && strings.TrimSpace(item.Path) != "" {
		fmt.Fprintf(w, "Saved artifact: %s\n", item.Path)
	}
	fmt.Fprintln(w, "Next commands: `jini continue`, `jini open`, or `jini status`.")
}

func richerUsefulItem(summary *workSummary) *catalogItem {
	return nextUsefulItem(summary)
}

func openShelfItems(summary *workSummary) []catalogItem {
	if summary == nil {
		return nil
	}
	items := make([]catalogItem, 0, len(summary.Views)+len(summary.Exports)+len(summary.Details))
	seen := map[string]bool{}
	appendUnique := func(group []catalogItem) {
		for _, item := range group {
			key := normalizeName(item.Label + ":" + item.Path)
			if seen[key] {
				continue
			}
			seen[key] = true
			items = append(items, item)
		}
	}
	appendUnique(summary.Views)
	appendUnique(summary.Exports)
	appendUnique(summary.Details)
	return items
}

func resolveInteractiveArtifactSelection(summary *workSummary, raw string) (*catalogItem, bool) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil, false
	}
	if index, err := strconv.Atoi(trimmed); err == nil {
		items := openShelfItems(summary)
		if index >= 1 && index <= len(items) {
			return &items[index-1], true
		}
	}
	item, err := resolveOpenItem(summary, raw)
	if err != nil {
		return nil, false
	}
	return item, true
}

func runInteractiveOpenShelf(summary *workSummary, scanner *bufio.Scanner, stdout, stderr io.Writer) int {
	renderOpenShelf(stdout, summary)
	if scanner == nil {
		return 0
	}
	choice, ok := readOptionalInputLine(scanner, stdout)
	if !ok || strings.TrimSpace(choice) == "" {
		return 0
	}
	fmt.Fprintln(stdout)
	item, ok := resolveInteractiveArtifactSelection(summary, choice)
	if !ok {
		fmt.Fprintln(stderr, "Type a number or artifact name to open one.")
		return 1
	}
	return openArtifactItem(summary, item, stdout, stderr)
}

func artifactThreadFocus(summary *workSummary, item *catalogItem) *threadFocus {
	if summary == nil || item == nil {
		return nil
	}
	return &threadFocus{Kind: "artifact", ArtifactPath: artifactRelativePath(summary.Dir, item.Path), ArtifactLabel: item.Label}
}

func focusArtifactSelection(summary *workSummary, item *catalogItem) {
	if summary == nil || item == nil {
		return
	}
	updateThreadFocus(summary.Dir, artifactThreadFocus(summary, item))
}

func recordAndFocusArtifactSelection(summary *workSummary, item *catalogItem) error {
	if summary == nil || item == nil {
		return nil
	}
	if err := recordPassiveArtifactObservation(summary.Dir, *item); err != nil {
		return err
	}
	focusArtifactSelection(summary, item)
	return nil
}

func renderSelectedArtifact(w io.Writer, summary *workSummary, item *catalogItem) error {
	if err := recordAndFocusArtifactSelection(summary, item); err != nil {
		return err
	}
	renderThreadSurface(w, summary, artifactThreadFocus(summary, item))
	return nil
}

func openArtifactItem(summary *workSummary, item *catalogItem, stdout, stderr io.Writer) int {
	if err := renderSelectedArtifact(stdout, summary, item); err != nil {
		fmt.Fprintf(stderr, "Could not open artifact: %v\n", err)
		return 1
	}
	return 0
}

func activeAskFocus(summary *workSummary) *threadFocus {
	if summary == nil {
		return nil
	}
	state := loadThreadState(summary.Dir, summary)
	if state.ActiveAsk == nil {
		return nil
	}
	return &threadFocus{Kind: "ask", AskID: state.ActiveAsk.AskID}
}

func renderThreadAsk(w io.Writer, summary *workSummary, ask *threadAsk) {
	if ask == nil {
		return
	}
	fmt.Fprintln(w, "Pending decision")
	fmt.Fprintln(w)
	fmt.Fprintln(w, ask.Prompt)
	if strings.TrimSpace(ask.Reason) != "" {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Why this matters")
		fmt.Fprintln(w, ask.Reason)
	}
	if len(ask.Options) > 0 {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Options")
		for _, item := range ask.Options {
			fmt.Fprintf(w, "- %s\n", item)
		}
	}
	if len(ask.AssumptionsIfSkipped) > 0 {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "If you skip this")
		for _, item := range ask.AssumptionsIfSkipped {
			fmt.Fprintf(w, "- %s\n", item)
		}
	}
	if strings.TrimSpace(summary.SafeToDo) != "" {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Safety")
		fmt.Fprintln(w, summary.SafeToDo)
	}
}

func resolveThreadSurface(summary *workSummary, state savedThreadState, focus *threadFocus) *threadFocus {
	if summary == nil || focus == nil {
		return nil
	}
	switch focus.Kind {
	case "artifact":
		if focusedArtifactItem(summary, focus) != nil {
			return focus
		}
	case "ask":
		if focusedThreadAsk(state, focus) != nil {
			return focus
		}
	case "missing":
		if state.ActiveAsk != nil {
			return activeAskFocus(summary)
		}
		return &threadFocus{Kind: "missing"}
	case "context":
		return &threadFocus{Kind: "context"}
	case "plan":
		return &threadFocus{Kind: "plan"}
	}
	return nil
}

func renderThreadSurface(w io.Writer, summary *workSummary, focus *threadFocus) bool {
	if summary == nil {
		return false
	}
	state := loadThreadState(summary.Dir, summary)
	resolved := resolveThreadSurface(summary, state, focus)
	if resolved == nil {
		return false
	}
	updateThreadFocus(summary.Dir, resolved)
	switch resolved.Kind {
	case "artifact":
		if item := focusedArtifactItem(summary, resolved); item != nil {
			renderItem(w, item)
			return true
		}
	case "ask":
		if ask := focusedThreadAsk(state, resolved); ask != nil {
			renderThreadAsk(w, summary, ask)
			return true
		}
	case "missing":
		renderMissingOnly(w, summary)
		return true
	case "context":
		renderContextCapsule(w, summary)
		return true
	case "plan":
		renderPlanFirst(w, summary)
		return true
	}
	return false
}

func renderFocusedContinuation(w io.Writer, summary *workSummary) bool {
	if summary == nil {
		return false
	}
	state := loadThreadState(summary.Dir, summary)
	if state.CurrentFocus != nil && renderThreadSurface(w, summary, state.CurrentFocus) {
		return true
	}
	if item := currentArtifactItem(summary); item != nil {
		return renderThreadSurface(w, summary, artifactThreadFocus(summary, item))
	}
	return false
}

func renderNextContinuation(w io.Writer, summary *workSummary) bool {
	if summary == nil {
		return false
	}
	if item := nextUsefulItem(summary); item != nil {
		return renderThreadSurface(w, summary, artifactThreadFocus(summary, item))
	}
	return false
}

func postResultWorkingWith(items []string) []string {
	out := []string{}
	for _, item := range items {
		clean := strings.TrimSpace(item)
		if clean == "" || strings.HasPrefix(clean, "Your request:") {
			continue
		}
		out = append(out, clean)
	}
	return out
}

func postResultAlsoReady(items []catalogItem, primary *catalogItem) []string {
	out := []string{}
	primaryID := ""
	if primary != nil {
		primaryID = primary.ID
	}
	for _, item := range items {
		if strings.TrimSpace(item.Label) == "" || item.ID == primaryID {
			continue
		}
		out = append(out, item.Label)
	}
	return out
}

func handlePostResultAction(action string, summary *workSummary, scanner *bufio.Scanner, stdout, stderr io.Writer) int {
	if resolution, resolved, err := resolveActiveAskAction(summary.Dir, summary, action); resolved {
		if err != nil {
			fmt.Fprintf(stderr, "Could not save the decision: %v\n", err)
			return 1
		}
		fmt.Fprintf(stdout, "Recorded decision: %s.\n", resolution)
		return 0
	}
	if access, ok := commercialOnlyCommandAccess(action); ok {
		renderFeatureAccessDenied(stdout, access)
		return 1
	}
	switch interactiveCommandName(action) {
	case "resume":
		if !renderFocusedContinuation(stdout, summary) {
			renderCheck(stdout, summary)
			return 0
		}
	case "continue":
		if !renderNextContinuation(stdout, summary) {
			renderCheck(stdout, summary)
			return 0
		}
	case "open":
		return runInteractiveOpenShelf(summary, scanner, stdout, stderr)
	case "status":
		renderCheck(stdout, summary)
	case "context":
		renderThreadSurface(stdout, summary, &threadFocus{Kind: "context"})
	case "missing":
		renderThreadSurface(stdout, summary, &threadFocus{Kind: "missing"})
	case "expand", "make it fuller", "make this fuller", "make it longer", "expand it", "expand this":
		if !renderNextContinuation(stdout, summary) {
			renderCheck(stdout, summary)
			return 0
		}
	case "shorter":
		return renderArtifactTransform(summary, "shorter", stdout, stderr)
	case "executive":
		return renderArtifactTransform(summary, "executive", stdout, stderr)
	case "checklist":
		return renderArtifactTransform(summary, "checklist", stdout, stderr)
	case "versions":
		renderArtifactVersions(stdout, summary)
	case "undo":
		item, err := undoLastArtifactChange(summary)
		if err != nil {
			fmt.Fprintf(stderr, "Could not restore the artifact: %v\n", err)
			return 1
		}
		fmt.Fprintln(stdout, "Restored the previous version.")
		fmt.Fprintln(stdout)
		renderItem(stdout, item)
	case "plan":
		renderThreadSurface(stdout, summary, &threadFocus{Kind: "plan"})
	case "accept":
		return recordArtifactFeedback(summary, "accepted-as-is", "Accept", stdout, stderr)
	case "edit":
		return recordArtifactFeedback(summary, "needed-light-edits", "Edit", stdout, stderr)
	case "use":
		return recordArtifactOutcome(summary, "used-this", "Use", stdout, stderr)
	case "share":
		return recordArtifactOutcome(summary, "shared-this", "Share", stdout, stderr)
	case "reject":
		return recordArtifactFeedback(summary, "not-useful", "Reject", stdout, stderr)
	case "replace":
		return recordArtifactOutcome(summary, "replaced-this", "Replace", stdout, stderr)
	case "start":
		renderNewWorkLauncher(stdout)
	default:
		if isSlashCommandInput(action) {
			fmt.Fprintf(stderr, "Unknown command %q.\n", strings.TrimSpace(action))
			return 1
		}
		if isAcknowledgementOnly(action) {
			renderCurrentWorkNoop(stdout)
			return 0
		}
		if item, ok := resolveInteractiveArtifactSelection(summary, action); ok {
			return openArtifactItem(summary, item, stdout, stderr)
		}
		renderCheck(stdout, summary)
	}
	return 0
}

func renderArtifactTransform(summary *workSummary, transform string, stdout, stderr io.Writer) int {
	item, err := applyArtifactTransform(summary, transform)
	if err != nil {
		fmt.Fprintf(stderr, "Could not revise the artifact: %v\n", err)
		return 1
	}
	renderItem(stdout, item)
	fmt.Fprintln(stdout)
	fmt.Fprintln(stdout, "Saved a restorable version. You can say `Versions` or `Undo`.")
	return 0
}

func recordArtifactFeedback(summary *workSummary, feedback, decision string, stdout, stderr io.Writer) int {
	if err := saveModelFeedback(summary.Dir, feedback, currentFeedbackArtifactPath(summary)); err != nil {
		fmt.Fprintf(stderr, "Could not save artifact feedback: %v\n", err)
		return 1
	}
	recordThreadDecision(summary.Dir, summary, decision)
	fmt.Fprintf(stdout, "Saved artifact feedback: %s.\n", feedback)
	return 0
}

func recordArtifactOutcome(summary *workSummary, outcome, decision string, stdout, stderr io.Writer) int {
	if err := saveArtifactOutcome(summary.Dir, outcome, currentFeedbackArtifactPath(summary)); err != nil {
		fmt.Fprintf(stderr, "Could not save artifact outcome: %v\n", err)
		return 1
	}
	recordThreadDecision(summary.Dir, summary, decision)
	fmt.Fprintf(stdout, "Saved artifact outcome: %s.\n", outcome)
	return 0
}

func nextUsefulItem(summary *workSummary) *catalogItem {
	if summary == nil {
		return nil
	}
	items := openShelfItems(summary)
	if len(items) == 0 {
		return nil
	}
	state := loadThreadState(summary.Dir, summary)
	if state.CurrentFocus != nil && state.CurrentFocus.Kind == "artifact" {
		for index, item := range items {
			if artifactRelativePath(summary.Dir, item.Path) == strings.TrimSpace(state.CurrentFocus.ArtifactPath) ||
				normalizeName(item.Label) == normalizeName(state.CurrentFocus.ArtifactLabel) {
				if index+1 < len(items) {
					copy := items[index+1]
					return &copy
				}
				return nil
			}
		}
	}
	preferred := []string{"owners-and-due-points", "missing-pieces-before-build", "next-actions", "task-list"}
	for _, id := range preferred {
		if item := findViewByID(summary, id); item != nil {
			return item
		}
	}
	if len(items) > 1 {
		copy := items[1]
		return &copy
	}
	copy := items[0]
	return &copy
}

func findViewByID(summary *workSummary, id string) *catalogItem {
	for index := range summary.Views {
		if summary.Views[index].ID == id {
			return &summary.Views[index]
		}
	}
	return nil
}

func renderItem(w io.Writer, item *catalogItem) {
	fmt.Fprintln(w, item.Label)
	fmt.Fprintln(w)
	content, err := os.ReadFile(item.Path)
	if err != nil {
		fmt.Fprintf(w, "Could not read %q yet.\n", item.Label)
		return
	}
	fmt.Fprint(w, strings.TrimSpace(string(content)))
	fmt.Fprintln(w)
}

func renderMissingOnly(w io.Writer, summary *workSummary) {
	fmt.Fprintln(w, "Still missing")
	if len(summary.Missing) == 0 {
		fmt.Fprintln(w, "- Nothing right now")
	} else {
		for _, item := range summary.Missing {
			fmt.Fprintf(w, "- %s\n", item)
		}
	}
	if len(summary.Uncertain) > 0 {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Not sure about")
		for _, item := range summary.Uncertain {
			fmt.Fprintf(w, "- %s\n", item)
		}
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Next step")
	fmt.Fprintln(w, summary.NextStep)
}

func renderPlanFirst(w io.Writer, summary *workSummary) {
	fmt.Fprintln(w, "Plan")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Goal")
	fmt.Fprintf(w, "- Finish %s without hiding what is missing.\n", summary.Title)
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Requirements")
	fmt.Fprintln(w, "- Keep the useful output visible.")
	fmt.Fprintln(w, "- Keep missing proof, approval, or owner gaps visible.")
	fmt.Fprintln(w, "- Do not send, change, book, or commit anything without review.")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Design")
	if len(summary.Views) == 0 {
		fmt.Fprintln(w, "- No ready surface exists yet; create the first useful pass before deeper execution.")
	} else {
		for _, item := range summary.Views {
			fmt.Fprintf(w, "- Use %s as an openable work surface.\n", item.Label)
		}
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Steps")
	fmt.Fprintln(w, "- Review what is ready.")
	fmt.Fprintln(w, "- Clear the highest-risk missing item.")
	fmt.Fprintln(w, "- Then continue only if the next action is safe.")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Run")
	fmt.Fprintln(w, "- Continue")
	fmt.Fprintln(w, "- Open")
	fmt.Fprintln(w, "- Missing")
}

func renderCurrentWorkLauncher(w io.Writer, summary *workSummary, interactive bool) {
	fmt.Fprintln(w, "Goal")
	fmt.Fprintln(w, summary.Thread.Goal)
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Working with")
	for _, item := range summary.Thread.WorkingWith {
		fmt.Fprintf(w, "- %s\n", item)
	}
	if len(summary.Thread.WorkingWith) == 0 {
		fmt.Fprintln(w, "- Nothing attached yet")
	}
	if strings.TrimSpace(summary.Thread.CurrentRoute) != "" {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "AI route")
		fmt.Fprintln(w, summary.Thread.CurrentRoute)
	}
	if interactive {
		if strings.TrimSpace(summary.Thread.UpNext) != "" {
			fmt.Fprintln(w)
			fmt.Fprintln(w, "Up next")
			fmt.Fprintln(w, summary.Thread.UpNext)
		}
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Ready now")
		if len(summary.Thread.ReadyNow) == 0 {
			fmt.Fprintln(w, "- Nothing is ready yet")
		} else {
			for _, item := range summary.Thread.ReadyNow {
				fmt.Fprintf(w, "- %s\n", item.Label)
			}
		}
		if len(summary.Thread.Blocked) > 0 {
			fmt.Fprintln(w)
			fmt.Fprintln(w, "Blocked")
			for _, item := range summary.Thread.Blocked {
				fmt.Fprintf(w, "- %s\n", item)
			}
		}
		other := otherActiveWorkSummaries(summary)
		renderOtherActiveWorkList(w, other, summary.Title, false)
		fmt.Fprintln(w)
		if strings.TrimSpace(summary.Thread.ModelLabel) != "" {
			renderArtifactFeedbackChoices(w)
		}
		renderCompactCurrentWorkChoices(w, len(other) > 0)
		return
	}
	if strings.TrimSpace(summary.Thread.ModelLabel) != "" {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Model")
		fmt.Fprintln(w, summary.Thread.ModelLabel)
	}
	if strings.TrimSpace(summary.Thread.ModelReason) != "" {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Why this model")
		fmt.Fprintln(w, summary.Thread.ModelReason)
	}
	if strings.TrimSpace(summary.Thread.ModelFeedback) != "" {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Model feedback")
		fmt.Fprintln(w, titleCase(summary.Thread.ModelFeedback))
	}
	if strings.TrimSpace(summary.Thread.RoutePolicy) != "" {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "How chosen")
		fmt.Fprintln(w, summary.Thread.RoutePolicy)
	}
	renderAutoModePolicy(w, summary.Thread.AutoMode)
	if strings.TrimSpace(summary.Thread.EffortLevel) != "" {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Effort level")
		fmt.Fprintln(w, titleCase(summary.Thread.EffortLevel))
	}
	if strings.TrimSpace(summary.Thread.VerificationLevel) != "" {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Verification")
		fmt.Fprintln(w, summary.Thread.VerificationLevel)
	}
	if strings.TrimSpace(summary.Thread.VerificationReason) != "" {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Why this verification")
		fmt.Fprintln(w, summary.Thread.VerificationReason)
	}
	if strings.TrimSpace(summary.Thread.RouteReason) != "" {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Why this route")
		fmt.Fprintln(w, summary.Thread.RouteReason)
	}
	if strings.TrimSpace(summary.Thread.ContinuityReason) != "" {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Continuity")
		fmt.Fprintln(w, summary.Thread.ContinuityReason)
	}
	if len(summary.Thread.MultimodalLearning) > 0 {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Multimodal learning")
		for _, line := range summary.Thread.MultimodalLearning {
			fmt.Fprintf(w, "- %s\n", line)
		}
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Just finished")
	for _, item := range summary.Thread.JustFinished {
		fmt.Fprintf(w, "- %s\n", item)
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Doing now")
	fmt.Fprintln(w, summary.Thread.DoingNow)
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Up next")
	fmt.Fprintln(w, summary.Thread.UpNext)
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Now")
	fmt.Fprintln(w, summary.Thread.Now)
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Done")
	for _, item := range summary.Thread.Done {
		fmt.Fprintf(w, "- %s\n", item)
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Need")
	fmt.Fprintln(w, summary.Thread.Need)
	if strings.TrimSpace(summary.Thread.NeedReason) != "" {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Why this matters")
		fmt.Fprintln(w, summary.Thread.NeedReason)
	}
	if len(summary.Thread.NeedOptions) > 0 {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Options")
		for _, item := range summary.Thread.NeedOptions {
			fmt.Fprintf(w, "- %s\n", item)
		}
	}
	if len(summary.Thread.Assumptions) > 0 {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "If you skip this")
		for _, item := range summary.Thread.Assumptions {
			fmt.Fprintf(w, "- %s\n", item)
		}
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Next")
	fmt.Fprintln(w, summary.Thread.Next)
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Ready now")
	for _, item := range summary.Thread.ReadyNow {
		fmt.Fprintf(w, "- %s\n", item.Label)
	}
	if len(summary.Thread.ReadyNow) == 0 {
		fmt.Fprintln(w, "- Nothing is ready yet")
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Blocked")
	if len(summary.Thread.Blocked) == 0 {
		fmt.Fprintln(w, "- Nothing right now")
	} else {
		for _, item := range summary.Thread.Blocked {
			fmt.Fprintf(w, "- %s\n", item)
		}
	}
	if len(summary.Thread.NotSureAbout) > 0 {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Not sure about")
		for _, item := range summary.Thread.NotSureAbout {
			fmt.Fprintf(w, "- %s\n", item)
		}
	}
	renderOtherActiveWorkList(w, otherActiveWorkSummaries(summary), summary.Title, false)
	fmt.Fprintln(w)
}

func renderCurrentWorkPrompt(w io.Writer, summary *workSummary) {
	renderNewWorkPrompt(w)
}

func renderCurrentWorkHelp(w io.Writer, summary *workSummary) {
	fmt.Fprintln(w, "Goal")
	fmt.Fprintln(w, summary.Thread.Goal)
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Working with")
	for _, item := range summary.Thread.WorkingWith {
		fmt.Fprintf(w, "- %s\n", item)
	}
	if len(summary.Thread.WorkingWith) == 0 {
		fmt.Fprintln(w, "- Nothing attached yet")
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Ready now")
	if len(summary.Thread.ReadyNow) == 0 {
		fmt.Fprintln(w, "- Nothing is ready yet")
	} else {
		for _, item := range summary.Thread.ReadyNow {
			fmt.Fprintf(w, "- %s\n", item.Label)
		}
	}
	if len(summary.Thread.Blocked) > 0 {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Blocked")
		for _, item := range summary.Thread.Blocked {
			fmt.Fprintf(w, "- %s\n", item)
		}
	}
	renderOtherActiveWorkList(w, otherActiveWorkSummaries(summary), summary.Title, false)
	if strings.TrimSpace(summary.Thread.ResumeTarget) != "" {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Resume")
		fmt.Fprintf(w, "- %s\n", summary.Thread.ResumeTarget)
	}
	if strings.TrimSpace(summary.Thread.ModelLabel) != "" {
		renderArtifactFeedbackChoices(w)
	}
	fmt.Fprintln(w)
	renderPrimaryActionMenu(w, summary, "Actions", "Open")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Commands: `open`, `status`, `doctor`, `memory`.")
}

func runActiveWorkLauncher(active []*workSummary, stdin io.Reader, stdout, stderr io.Writer) int {
	if stdin == nil {
		renderActiveWorkLauncher(stdout, active)
		return 0
	}
	return runActiveWorkPicker(active, bufio.NewScanner(stdin), stdout, stderr)
}

func runActiveWorkPicker(active []*workSummary, scanner *bufio.Scanner, stdout, stderr io.Writer) int {
	if len(active) == 0 {
		fmt.Fprintln(stdout, "No active work is saved yet.")
		return 0
	}
	renderActiveWorkLauncher(stdout, active)
	if scanner == nil {
		return 0
	}
	action, ok := readOptionalInputLine(scanner, stdout)
	if !ok || strings.TrimSpace(action) == "" {
		return 0
	}
	fmt.Fprintln(stdout)
	return handleActiveWorkSelection(action, active, scanner, stdout, stderr)
}

func renderActiveWorkLauncher(w io.Writer, active []*workSummary) {
	renderSelectableWorkList(w, "Active work", active, true)
}

func handleActiveWorkSelection(action string, active []*workSummary, scanner *bufio.Scanner, stdout, stderr io.Writer) int {
	selection, err := resolveActiveWorkSelection(action, active)
	if err != nil {
		switch interactiveCommandName(action) {
		case "start":
			if scanner == nil {
				renderNewWorkLauncher(stdout)
				return 0
			}
			return runNewWorkIntakeWithScanner(scanner, stdout, stderr)
		}
		fmt.Fprintf(stderr, "%v\n", err)
		return 1
	}
	return switchToWorkSelection(selection, stdout, stderr)
}

func runSwitchWorkPicker(current *workSummary, scanner *bufio.Scanner, stdout, stderr io.Writer) int {
	other := otherActiveWorkSummaries(current)
	if len(other) == 0 {
		fmt.Fprintln(stdout, "No other active work right now.")
		return 0
	}
	renderSelectableWorkList(stdout, "Saved work", other, false)
	if scanner == nil {
		return 0
	}
	choice, ok := readOptionalInputLine(scanner, stdout)
	if !ok || strings.TrimSpace(choice) == "" {
		return 0
	}
	fmt.Fprintln(stdout)
	selection, err := resolveActiveWorkSelection(choice, other)
	if err != nil {
		fmt.Fprintf(stderr, "%v\n", err)
		return 1
	}
	return switchToWorkSelection(selection, stdout, stderr)
}

func switchToWorkSelection(selection *workSummary, stdout, stderr io.Writer) int {
	if err := saveSummaryAsCurrent(selection); err != nil {
		fmt.Fprintf(stderr, "Could not switch work: %v\n", err)
		return 1
	}
	fmt.Fprintln(stdout, "Resumed")
	fmt.Fprintln(stdout, selection.Title)
	fmt.Fprintln(stdout)
	renderCheck(stdout, selection)
	return 0
}

func listActiveWorkSummaries(current *workSummary) ([]*workSummary, error) {
	root := filepath.Join(sessionStateRoot(), "work")
	entries, err := os.ReadDir(root)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	active := []*workSummary{}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		summary, loadErr := loadWorkSummary(filepath.Join(root, entry.Name()), nil)
		if loadErr != nil {
			continue
		}
		active = append(active, summary)
	}
	sort.SliceStable(active, func(i, j int) bool {
		return active[i].Title < active[j].Title
	})
	if current == nil {
		return active, nil
	}
	currentPath := ""
	if resolved, err := filepath.Abs(current.Dir); err == nil {
		currentPath = resolved
	}
	out := make([]*workSummary, 0, len(active))
	for _, item := range active {
		itemPath := item.Dir
		if resolved, err := filepath.Abs(item.Dir); err == nil {
			itemPath = resolved
		}
		if currentPath != "" && itemPath == currentPath {
			continue
		}
		out = append(out, item)
	}
	return out, nil
}

func otherActiveWorkSummaries(current *workSummary) []*workSummary {
	active, err := listActiveWorkSummaries(current)
	if err != nil {
		return nil
	}
	return active
}

func resolveActiveWorkSelection(action string, active []*workSummary) (*workSummary, error) {
	normalized := normalizeCommandName(action)
	for index, item := range active {
		if normalized == normalizeName(fmt.Sprintf("%d", index+1)) || normalized == normalizeName(item.Title) {
			return item, nil
		}
	}
	return nil, fmt.Errorf("Work not found. Enter a shown number or saved title.")
}

func saveSummaryAsCurrent(summary *workSummary) error {
	current := &currentWork{
		PackDir:    summary.Dir,
		PackID:     summary.PackID,
		WorkUnitID: summary.WorkUnitID,
		Title:      summary.Title,
		State:      summary.State,
		Health:     inferHealthFromState(summary.State),
	}
	return saveCurrentWork(current)
}

func renderCheck(w io.Writer, summary *workSummary) {
	fmt.Fprintln(w, "Goal")
	fmt.Fprintln(w, summary.Thread.Goal)
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Working with")
	for _, item := range summary.Thread.WorkingWith {
		fmt.Fprintf(w, "- %s\n", item)
	}
	if len(summary.Thread.WorkingWith) == 0 {
		fmt.Fprintln(w, "- Nothing attached yet")
	}
	if strings.TrimSpace(summary.Thread.CurrentRoute) != "" {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "AI route")
		fmt.Fprintln(w, summary.Thread.CurrentRoute)
	}
	renderCLIHandoffReceiptSummary(w, summary.Thread.CLIHandoffSummary)
	if strings.TrimSpace(summary.Thread.ModelLabel) != "" {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Model")
		fmt.Fprintln(w, summary.Thread.ModelLabel)
	}
	if strings.TrimSpace(summary.Thread.ModelReason) != "" {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Why this model")
		fmt.Fprintln(w, summary.Thread.ModelReason)
	}
	if strings.TrimSpace(summary.Thread.ModelFeedback) != "" {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Model feedback")
		fmt.Fprintln(w, titleCase(summary.Thread.ModelFeedback))
	}
	if strings.TrimSpace(summary.Thread.RoutePolicy) != "" {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "How chosen")
		fmt.Fprintln(w, summary.Thread.RoutePolicy)
	}
	renderAutoModePolicy(w, summary.Thread.AutoMode)
	if strings.TrimSpace(summary.Thread.EffortLevel) != "" {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Effort level")
		fmt.Fprintln(w, titleCase(summary.Thread.EffortLevel))
	}
	if strings.TrimSpace(summary.Thread.VerificationLevel) != "" {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Verification")
		fmt.Fprintln(w, summary.Thread.VerificationLevel)
	}
	if strings.TrimSpace(summary.Thread.VerificationReason) != "" {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Why this verification")
		fmt.Fprintln(w, summary.Thread.VerificationReason)
	}
	if strings.TrimSpace(summary.Thread.RouteReason) != "" {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Why this route")
		fmt.Fprintln(w, summary.Thread.RouteReason)
	}
	if strings.TrimSpace(summary.Thread.ContinuityReason) != "" {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Continuity")
		fmt.Fprintln(w, summary.Thread.ContinuityReason)
	}
	if len(summary.Thread.MultimodalLearning) > 0 {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Multimodal learning")
		for _, line := range summary.Thread.MultimodalLearning {
			fmt.Fprintf(w, "- %s\n", line)
		}
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Just finished")
	for _, item := range summary.Thread.JustFinished {
		fmt.Fprintf(w, "- %s\n", item)
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Doing now")
	fmt.Fprintln(w, summary.Thread.DoingNow)
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Up next")
	fmt.Fprintln(w, summary.Thread.UpNext)
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Now")
	fmt.Fprintln(w, summary.Thread.Now)
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Done")
	for _, item := range summary.Thread.Done {
		fmt.Fprintf(w, "- %s\n", item)
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Need")
	fmt.Fprintln(w, summary.Thread.Need)
	if strings.TrimSpace(summary.Thread.NeedReason) != "" {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Why this matters")
		fmt.Fprintln(w, summary.Thread.NeedReason)
	}
	if len(summary.Thread.NeedOptions) > 0 {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Options")
		for _, item := range summary.Thread.NeedOptions {
			fmt.Fprintf(w, "- %s\n", item)
		}
	}
	if len(summary.Thread.Assumptions) > 0 {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "If you skip this")
		for _, item := range summary.Thread.Assumptions {
			fmt.Fprintf(w, "- %s\n", item)
		}
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Next")
	fmt.Fprintln(w, summary.Thread.Next)
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Ready now")
	for _, item := range summary.Thread.ReadyNow {
		fmt.Fprintf(w, "- %s\n", item.Label)
	}
	if len(summary.Thread.ReadyNow) == 0 {
		fmt.Fprintln(w, "- Nothing is ready yet")
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Blocked")
	if len(summary.Thread.Blocked) == 0 {
		fmt.Fprintln(w, "- Nothing right now")
	} else {
		for _, item := range summary.Thread.Blocked {
			fmt.Fprintf(w, "- %s\n", item)
		}
	}
	fmt.Fprintln(w)
	if len(summary.Thread.NotSureAbout) > 0 {
		fmt.Fprintln(w, "Not sure about")
		for _, item := range summary.Thread.NotSureAbout {
			fmt.Fprintf(w, "- %s\n", item)
		}
		fmt.Fprintln(w)
	}
	fmt.Fprintln(w, "Safe to do")
	fmt.Fprintln(w, summary.Thread.SafeToDo)
}

func summaryWorkingWith(summary *workSummary) string {
	if summary != nil && strings.TrimSpace(summary.WorkingWith) != "" {
		return summary.WorkingWith
	}
	return workingWithLabel(detectProvider())
}

func renderOpenShelf(w io.Writer, summary *workSummary) {
	fmt.Fprintln(w, "Open")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Ready now")
	indexByPath := map[string]int{}
	nextIndex := 1
	for _, item := range summary.Views {
		fmt.Fprintf(w, "%d. %s\n", nextIndex, item.Label)
		indexByPath[item.Path] = nextIndex
		nextIndex++
	}
	if len(summary.Views) == 0 {
		fmt.Fprintln(w, "- Nothing is ready yet")
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Send / share")
	if len(summary.Exports) == 0 {
		fmt.Fprintln(w, "- Nothing to share yet")
	} else {
		for _, item := range summary.Exports {
			if index, ok := indexByPath[item.Path]; ok {
				fmt.Fprintf(w, "%d. %s\n", index, item.Label)
				continue
			}
			fmt.Fprintf(w, "%d. %s\n", nextIndex, item.Label)
			indexByPath[item.Path] = nextIndex
			nextIndex++
		}
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Details")
	if len(summary.Details) == 0 {
		fmt.Fprintln(w, "- No extra details yet")
	} else {
		for _, item := range summary.Details {
			if index, ok := indexByPath[item.Path]; ok {
				fmt.Fprintf(w, "%d. %s\n", index, item.Label)
				continue
			}
			fmt.Fprintf(w, "%d. %s\n", nextIndex, item.Label)
			indexByPath[item.Path] = nextIndex
			nextIndex++
		}
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Type a number or name to open one, or press Enter to go back.")
}

func dedupeItems(items []catalogItem) []catalogItem {
	seen := map[string]bool{}
	out := make([]catalogItem, 0, len(items))
	for _, item := range items {
		key := normalizeName(item.Label + ":" + item.Path)
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, item)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Label < out[j].Label })
	return out
}

func normalizeName(value string) string {
	if strings.TrimSpace(value) == "?" {
		return "?"
	}
	var builder strings.Builder
	for _, r := range strings.ToLower(strings.TrimSpace(value)) {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r) || unicode.IsSpace(r):
			builder.WriteRune(r)
		case r == '\'' || r == '’':
			continue
		default:
			builder.WriteRune(' ')
		}
	}
	return strings.Join(strings.Fields(builder.String()), " ")
}

func exactCommandToken(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func canonicalTopLevelCommand(value string) string {
	switch exactCommandToken(value) {
	case "help", "--help", "-h":
		return "help"
	case "commands", "admin", "check", "status", "continue", "doctor", "provider", "route", "memory", "permissions", "init", "new", "observe", "open", "run", "publish-readiness", "scorecard-gate":
		return exactCommandToken(value)
	default:
		return ""
	}
}

func canonicalHelpTopic(value string) string {
	switch exactCommandToken(value) {
	case "all", "--all":
		return "all"
	case "commands":
		return "commands"
	case "admin", "--admin":
		return "admin"
	default:
		return ""
	}
}

func isAdminHelpAlias(value string) bool {
	switch exactCommandToken(value) {
	case "help", "--help", "-h":
		return true
	default:
		return false
	}
}

func normalizeCommandName(value string) string {
	trimmed := strings.TrimSpace(value)
	if strings.HasPrefix(trimmed, "/") {
		return ""
	}
	return normalizeName(value)
}

func interactiveCommandName(value string) string {
	trimmed := strings.TrimSpace(value)
	if isSlashCommandInput(trimmed) {
		trimmed = strings.TrimPrefix(trimmed, "/")
	}
	return normalizeCommandName(trimmed)
}

func isSlashCommandInput(value string) bool {
	trimmed := strings.TrimSpace(value)
	if !strings.HasPrefix(trimmed, "/") {
		return false
	}
	if strings.Contains(trimmed[1:], "/") {
		return false
	}
	for _, r := range trimmed[1:] {
		if !(unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' || r == '_' || unicode.IsSpace(r)) {
			return false
		}
	}
	return len(trimmed) > 1
}

func containsAny(value string, candidates []string) bool {
	for _, candidate := range candidates {
		if strings.Contains(value, candidate) {
			return true
		}
	}
	return false
}

func titleCase(value string) string {
	parts := strings.Fields(strings.NewReplacer("-", " ", "_", " ").Replace(value))
	for i := range parts {
		if parts[i] == "" {
			continue
		}
		parts[i] = strings.ToUpper(parts[i][:1]) + strings.ToLower(parts[i][1:])
	}
	return strings.Join(parts, " ")
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
