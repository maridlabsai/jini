# Auto/Ask Execution Mode Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement `jini mode [auto|ask]` plus Ask-mode approval before resuming throttled work, with a park record that makes `jini continue` guidance true, per `specs/auto-ask-execution-mode-design.md`.

**Architecture:** A global fail-closed mode setting (`~/.jini/mode.json`, `JINI_MODE` override); a `throttleApprover` interface consulted once per `runWithThrottleSurvival` call via a process-level default set at dispatch branches; a caller-owned park file (`sessionStateRoot()/throttle-park.json`) pre-written on Ask paths and deleted only on success; per-attempt timeouts and a single-hold ladder on the standalone-question path.

**Tech Stack:** Go stdlib only (no new dependencies). Tests via `go test ./internal/app`.

## Global Constraints

- Token frugality is P0; no verbose ceremony in any output copy (CLAUDE.md).
- Auto mode must remain byte-identical in behavior and output to commit 5e5d8d7's shipped path.
- Mode parsing fails CLOSED: any unreadable/unknown source resolves to "ask" with a one-line stderr warning. Absent file = normal "auto" default.
- All new state files: atomic temp+rename writes, 0o600.
- Prompt default is DECLINE: only explicit `y`/`yes` grants.
- "session is saved; `jini continue` resumes it" copy appears ONLY where a park record was written.
- Run `bash tools/run_required_gates.sh commit` before every commit; commit only after diff review.
- Commit trailer: `Co-Authored-By: Claude Fable 5 <noreply@anthropic.com>`.
- PRECONDITION: the CLI-handoff dogfood fix diff (concurrent work in app.go/cli_handoff.go/provider.go) must be committed before starting Task 2 or later. Task 1 touches only new files and may start immediately.

---

### Task 1: Execution mode settings (`execution_mode.go`)

**Files:**
- Create: `internal/app/execution_mode.go`
- Test: `internal/app/execution_mode_test.go`

**Interfaces:**
- Consumes: nothing new.
- Produces: `effectiveExecutionMode() string` (returns exactly `"auto"` or `"ask"`), `saveExecutionMode(mode string) error`, `executionModePath() (string, error)`, package var `executionModeWarnings io.Writer = os.Stderr`, test seam `executionModeHomeDir = os.UserHomeDir`.

- [ ] **Step 1: Write the failing tests**

```go
package app

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func withExecutionModeHome(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	oldHome := executionModeHomeDir
	executionModeHomeDir = func() (string, error) { return home, nil }
	t.Cleanup(func() { executionModeHomeDir = oldHome })
	t.Setenv("JINI_MODE", "")
	return home
}

func TestEffectiveExecutionModeDefaultsToAuto(t *testing.T) {
	withExecutionModeHome(t)
	if mode := effectiveExecutionMode(); mode != "auto" {
		t.Fatalf("expected auto, got %q", mode)
	}
}

func TestEffectiveExecutionModeRoundTrip(t *testing.T) {
	withExecutionModeHome(t)
	if err := saveExecutionMode("ask"); err != nil {
		t.Fatal(err)
	}
	if mode := effectiveExecutionMode(); mode != "ask" {
		t.Fatalf("expected ask, got %q", mode)
	}
	if err := saveExecutionMode("auto"); err != nil {
		t.Fatal(err)
	}
	if mode := effectiveExecutionMode(); mode != "auto" {
		t.Fatalf("expected auto, got %q", mode)
	}
}

func TestEffectiveExecutionModeEnvOverridesFile(t *testing.T) {
	withExecutionModeHome(t)
	if err := saveExecutionMode("auto"); err != nil {
		t.Fatal(err)
	}
	t.Setenv("JINI_MODE", "ask")
	if mode := effectiveExecutionMode(); mode != "ask" {
		t.Fatalf("expected ask from env, got %q", mode)
	}
}

func TestEffectiveExecutionModeCorruptFileFailsClosed(t *testing.T) {
	home := withExecutionModeHome(t)
	dir := filepath.Join(home, ".jini")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "mode.json"), []byte("{corrupt"), 0o600); err != nil {
		t.Fatal(err)
	}
	var warnings bytes.Buffer
	oldWarn := executionModeWarnings
	executionModeWarnings = &warnings
	defer func() { executionModeWarnings = oldWarn }()
	if mode := effectiveExecutionMode(); mode != "ask" {
		t.Fatalf("corrupt file must fail closed to ask, got %q", mode)
	}
	if !strings.Contains(warnings.String(), "mode setting unreadable") {
		t.Fatalf("expected warning, got %q", warnings.String())
	}
}

func TestEffectiveExecutionModeGarbageEnvFailsClosed(t *testing.T) {
	withExecutionModeHome(t)
	t.Setenv("JINI_MODE", "garbage")
	var warnings bytes.Buffer
	oldWarn := executionModeWarnings
	executionModeWarnings = &warnings
	defer func() { executionModeWarnings = oldWarn }()
	if mode := effectiveExecutionMode(); mode != "ask" {
		t.Fatalf("garbage env must fail closed to ask, got %q", mode)
	}
}

func TestEffectiveExecutionModeHomeDirFailureFailsClosed(t *testing.T) {
	oldHome := executionModeHomeDir
	executionModeHomeDir = func() (string, error) { return "", os.ErrPermission }
	defer func() { executionModeHomeDir = oldHome }()
	t.Setenv("JINI_MODE", "")
	var warnings bytes.Buffer
	oldWarn := executionModeWarnings
	executionModeWarnings = &warnings
	defer func() { executionModeWarnings = oldWarn }()
	if mode := effectiveExecutionMode(); mode != "ask" {
		t.Fatalf("homedir failure must fail closed to ask, got %q", mode)
	}
}

func TestSaveExecutionModeAtomicNoTempDroppings(t *testing.T) {
	home := withExecutionModeHome(t)
	if err := saveExecutionMode("ask"); err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(filepath.Join(home, ".jini"))
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if entry.Name() != "mode.json" {
			t.Fatalf("unexpected file left behind: %s", entry.Name())
		}
	}
	info, err := os.Stat(filepath.Join(home, ".jini", "mode.json"))
	if err != nil {
		t.Fatal(err)
	}
	if perm := info.Mode().Perm(); perm != 0o600 {
		t.Fatalf("expected 0600, got %o", perm)
	}
}

func TestSaveExecutionModeRejectsInvalid(t *testing.T) {
	withExecutionModeHome(t)
	if err := saveExecutionMode("banana"); err == nil {
		t.Fatal("expected error for invalid mode")
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/app -run TestEffectiveExecutionMode -v`
Expected: FAIL — `undefined: executionModeHomeDir` (compile error).

- [ ] **Step 3: Write the implementation**

```go
package app

// Auto/Ask execution mode — specs/auto-ask-execution-mode-design.md.
// GLOBAL setting (~/.jini/mode.json), deliberately diverging from the
// per-project sessionStateRoot() pattern: supervision is a property of the
// user, not the workspace. Parsing fails CLOSED: a corrupted safety toggle
// must degrade to supervision ("ask"), not autonomy — this deliberately
// diverges from router_settings.go's fail-open load. An absent file is not
// an error; that is the normal "auto" default.

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const (
	executionModeAuto = "auto"
	executionModeAsk  = "ask"
)

var (
	executionModeWarnings io.Writer = os.Stderr
	executionModeHomeDir            = os.UserHomeDir
)

type savedExecutionMode struct {
	SchemaVersion string `json:"schema_version"`
	ContextType   string `json:"context_type"`
	Mode          string `json:"mode"`
}

func executionModePath() (string, error) {
	home, err := executionModeHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".jini", "mode.json"), nil
}

// effectiveExecutionMode resolves JINI_MODE env, then ~/.jini/mode.json,
// then "auto". Every unreadable or unknown source fails closed to "ask".
func effectiveExecutionMode() string {
	if raw := strings.TrimSpace(os.Getenv("JINI_MODE")); raw != "" {
		return parseExecutionMode(raw, "JINI_MODE")
	}
	path, err := executionModePath()
	if err != nil {
		return warnExecutionModeUnreadable("home directory")
	}
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return executionModeAuto
	}
	if err != nil {
		return warnExecutionModeUnreadable(path)
	}
	var payload savedExecutionMode
	if err := json.Unmarshal(data, &payload); err != nil {
		return warnExecutionModeUnreadable(path)
	}
	return parseExecutionMode(payload.Mode, path)
}

func parseExecutionMode(raw, source string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case executionModeAuto:
		return executionModeAuto
	case executionModeAsk:
		return executionModeAsk
	default:
		return warnExecutionModeUnreadable(source)
	}
}

func warnExecutionModeUnreadable(source string) string {
	if executionModeWarnings != nil {
		fmt.Fprintf(executionModeWarnings, "mode setting unreadable (%s); treating as Ask (supervised)\n", source)
	}
	return executionModeAsk
}

func saveExecutionMode(mode string) error {
	if mode != executionModeAuto && mode != executionModeAsk {
		return fmt.Errorf("unknown mode %q", mode)
	}
	path, err := executionModePath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	payload := savedExecutionMode{SchemaVersion: "0.1.0", ContextType: "JiniExecutionMode", Mode: mode}
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), "mode-*.json.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	if _, err := tmp.Write(append(data, '\n')); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return err
	}
	if err := tmp.Chmod(0o600); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return err
	}
	return os.Rename(tmpName, path)
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/app -run 'TestEffectiveExecutionMode|TestSaveExecutionMode' -v`
Expected: all PASS.

- [ ] **Step 5: Gate + commit**

```bash
bash tools/run_required_gates.sh commit
git add internal/app/execution_mode.go internal/app/execution_mode_test.go
git commit -m "feat: global fail-closed Auto/Ask execution mode setting

Co-Authored-By: Claude Fable 5 <noreply@anthropic.com>"
```

---

### Task 2: `jini mode` command surface

**Files:**
- Modify: `internal/app/app.go` (dispatch switch ~line 112; `canonicalTopLevelCommand` ~line 5393; `validateNativeArgs` ~line 246; help listing — locate via `grep -n '"route"' internal/app/app.go` in the help renderer)
- Modify: `internal/app/go_migration_test.go:90` (allowed-command map)
- Modify: `specs/product-settling-decisions.md` (named spec amendment)
- Create: `runMode` in `internal/app/execution_mode.go`
- Test: `internal/app/execution_mode_test.go`

**Interfaces:**
- Consumes: Task 1's `effectiveExecutionMode()`, `saveExecutionMode(mode)`.
- Produces: `runMode(args []string, stdout, stderr io.Writer) int`; top-level command token `"mode"`.

- [ ] **Step 1: Write the failing tests**

```go
func TestRunModeBarePrintsEffectiveModeAndHint(t *testing.T) {
	withExecutionModeHome(t)
	var out bytes.Buffer
	if code := runMode(nil, &out, io.Discard); code != 0 {
		t.Fatalf("exit %d", code)
	}
	got := out.String()
	if !strings.Contains(got, "Auto — Jini picks for you and keeps going.") {
		t.Fatalf("missing auto copy: %q", got)
	}
	if !strings.Contains(got, "Switch with `jini mode ask`.") {
		t.Fatalf("missing switch hint: %q", got)
	}
}

func TestRunModeSwitchToAsk(t *testing.T) {
	withExecutionModeHome(t)
	var out bytes.Buffer
	if code := runMode([]string{"ask"}, &out, io.Discard); code != 0 {
		t.Fatalf("exit %d", code)
	}
	if !strings.Contains(out.String(), "Ask — Jini checks with you before side effects and before resuming throttled work.") {
		t.Fatalf("missing ask copy: %q", out.String())
	}
	if mode := effectiveExecutionMode(); mode != "ask" {
		t.Fatalf("mode not persisted, got %q", mode)
	}
}

func TestRunModeBareNamesEnvOverride(t *testing.T) {
	withExecutionModeHome(t)
	if err := saveExecutionMode("auto"); err != nil {
		t.Fatal(err)
	}
	t.Setenv("JINI_MODE", "ask")
	var out bytes.Buffer
	runMode(nil, &out, io.Discard)
	if !strings.Contains(out.String(), "(from JINI_MODE; saved setting is Auto)") {
		t.Fatalf("env override not named: %q", out.String())
	}
}

func TestRunModeInvalidArg(t *testing.T) {
	withExecutionModeHome(t)
	var errOut bytes.Buffer
	if code := runMode([]string{"banana"}, io.Discard, &errOut); code != 1 {
		t.Fatalf("expected exit 1, got %d", code)
	}
	if !strings.Contains(errOut.String(), "Unknown mode \"banana\". Use `jini mode auto` or `jini mode ask`.") {
		t.Fatalf("wrong rejection copy: %q", errOut.String())
	}
}

func TestModeIsARoutedTopLevelCommand(t *testing.T) {
	if canonicalTopLevelCommand("mode") != "mode" {
		t.Fatal("mode not a canonical top-level command")
	}
	if err := validateNativeArgs([]string{"mode", "ask"}); err != nil {
		t.Fatalf("mode ask must validate: %v", err)
	}
	if err := validateNativeArgs([]string{"mode", "banana"}); err != nil {
		t.Fatalf("mode banana must pass validation so runMode's friendly error is reachable: %v", err)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/app -run 'TestRunMode|TestModeIsARouted' -v`
Expected: FAIL — `undefined: runMode`.

- [ ] **Step 3: Implement `runMode` (append to execution_mode.go)**

```go
func executionModeDisplayLine(mode string) string {
	if mode == executionModeAsk {
		return "Ask — Jini checks with you before side effects and before resuming throttled work."
	}
	return "Auto — Jini picks for you and keeps going."
}

func runMode(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		mode := effectiveExecutionMode()
		line := executionModeDisplayLine(mode)
		if env := strings.TrimSpace(os.Getenv("JINI_MODE")); env != "" {
			saved := executionModeAuto
			if path, err := executionModePath(); err == nil {
				if data, readErr := os.ReadFile(path); readErr == nil {
					var payload savedExecutionMode
					if json.Unmarshal(data, &payload) == nil && strings.EqualFold(payload.Mode, executionModeAsk) {
						saved = executionModeAsk
					}
				}
			}
			line = strings.TrimSuffix(line, ".") + fmt.Sprintf(" (from JINI_MODE; saved setting is %s).", titleCase(saved))
		}
		fmt.Fprintln(stdout, line)
		if mode == executionModeAsk {
			fmt.Fprintln(stdout, "Switch with `jini mode auto`.")
		} else {
			fmt.Fprintln(stdout, "Switch with `jini mode ask`.")
		}
		return 0
	}
	requested := strings.ToLower(strings.TrimSpace(args[0]))
	if requested != executionModeAuto && requested != executionModeAsk {
		fmt.Fprintf(stderr, "Unknown mode %q. Use `jini mode auto` or `jini mode ask`.\n", args[0])
		return 1
	}
	if err := saveExecutionMode(requested); err != nil {
		fmt.Fprintln(stderr, err.Error())
		return 1
	}
	fmt.Fprintln(stdout, executionModeDisplayLine(requested))
	if strings.TrimSpace(os.Getenv("JINI_MODE")) != "" {
		fmt.Fprintln(stdout, "Note: JINI_MODE is set and still wins in this shell.")
	}
	return 0
}
```

- [ ] **Step 4: Register the command (app.go, four sites)**

In the dispatch switch (after `case "memory":`):

```go
		case "mode":
			return runMode(args[1:], stdout, stderr)
```

In `canonicalTopLevelCommand`, add `"mode"` to the recognized-token case list:

```go
	case "commands", "admin", "check", "status", "continue", "doctor", "provider", "route", "memory", "mode", "permissions", "init", "new", "observe", "open", "run", "publish-readiness", "scorecard-gate":
```

In `validateNativeArgs`, add a case (loose — len ≤ 2, any second arg passes so `runMode`'s friendly rejection is reachable instead of the generic pre-dispatch "Unsupported arguments" error):

```go
	case "mode":
		if len(args) <= 2 {
			return nil
		}
```

In the help renderer, add one line next to the `route` entry (match surrounding format exactly when implementing):

```
mode        Show or switch Auto/Ask execution mode
```

In `internal/app/go_migration_test.go:90`, add `"mode": true,` to the `allowed` map, keeping alphabetical order.

- [ ] **Step 5: Record the spec amendment**

Append to `specs/product-settling-decisions.md`:

```markdown
## 2026-07-21: `jini mode` joins the taught command surface

The Auto/Ask execution mode (PRD §Execution modes) ships as top-level
`jini mode [auto|ask]`. The allowed-command map in
`internal/app/go_migration_test.go` gains `mode` as a deliberate surface
addition — one command, no aliases, no subtree — per
`specs/auto-ask-execution-mode-design.md`. Command-surface discipline
(engineering-gate-matrix §Command-Surface Discipline) was weighed: a mode
toggle is the PRD's own "one obvious action" requirement.
```

- [ ] **Step 6: Run tests to verify they pass**

Run: `go test ./internal/app -run 'TestRunMode|TestModeIsARouted|TestOfficialGoOnly' -v`
Expected: all PASS (including the migration-surface test with the new map entry).

- [ ] **Step 7: Gate + commit**

```bash
bash tools/run_required_gates.sh commit
git add internal/app/execution_mode.go internal/app/execution_mode_test.go internal/app/app.go internal/app/go_migration_test.go specs/product-settling-decisions.md
git commit -m "feat: jini mode command — Auto/Ask toggle with named spec amendment

Co-Authored-By: Claude Fable 5 <noreply@anthropic.com>"
```

---

### Task 3: Approver seam in throttle survival

**Files:**
- Modify: `internal/app/throttle_survival.go`
- Modify: `internal/app/provider.go:129,155` (call sites gain options argument)
- Modify: `internal/app/simple_answers.go` (standalone profile; see Task 5 for park/copy)
- Test: `internal/app/throttle_survival_test.go`

**Interfaces:**
- Consumes: Task 1's `effectiveExecutionMode()`.
- Produces:

```go
type throttleApprovalRequest struct {
	Label        string
	Wait         time.Duration
	FallbackHint string
	Hold         int
	TaskTitle    string
}
type throttleApprovalDecision int
const (
	approvalGranted throttleApprovalDecision = iota
	approvalDeclined
)
type throttleApprover interface {
	Approve(ctx context.Context, req throttleApprovalRequest) (throttleApprovalDecision, error)
}
type autoApprover struct{}
type failClosedApprover struct{}
type throttleSurvivalOptions struct {
	approver       throttleApprover // nil = throttleApproverForProcess
	attemptTimeout time.Duration    // 0 = no per-attempt timeout
	holdWaits      []time.Duration  // nil = throttleHoldWaits (20/40/80)
	taskTitle      string
}
var throttleApproverForProcess throttleApprover = autoApprover{}
type throttleDeclinedError struct{ label, fallbackHint string; underlying error }
// runWithThrottleSurvival(ctx, label string, fallbackHint func() string,
//     opts throttleSurvivalOptions, attempt func() (string, error))
//     (string, throttleSurvivalReport, error)
```

- [ ] **Step 1: Write the failing tests (append to throttle_survival_test.go, following its existing override style)**

```go
type recordingApprover struct {
	calls    int
	decision throttleApprovalDecision
	lastReq  throttleApprovalRequest
}

func (a *recordingApprover) Approve(_ context.Context, req throttleApprovalRequest) (throttleApprovalDecision, error) {
	a.calls++
	a.lastReq = req
	return a.decision, nil
}

func TestRunWithThrottleSurvivalAutoApproverNeverConsultedOutput(t *testing.T) {
	// Auto path byte-identical: same narration as the shipped suite's
	// hold test, with an approver present.
	var narration bytes.Buffer
	oldNarration := throttleNarration
	throttleNarration = &narration
	defer func() { throttleNarration = oldNarration }()
	oldSleep := throttleSleep
	throttleSleep = func(context.Context, time.Duration) error { return nil }
	defer func() { throttleSleep = oldSleep }()

	attempts := 0
	text, report, err := runWithThrottleSurvival(context.Background(), "Claude API route", nil, throttleSurvivalOptions{approver: autoApprover{}}, func() (string, error) {
		attempts++
		if attempts == 1 {
			return "", &throttledRouteError{label: "Claude API route", underlying: errors.New("status 429")}
		}
		return "done", nil
	})
	if err != nil || text != "done" || report.Holds != 1 {
		t.Fatalf("text=%q holds=%d err=%v", text, report.Holds, err)
	}
	if !strings.Contains(narration.String(), "Holding the session; resuming automatically") {
		t.Fatalf("auto path narration changed: %q", narration.String())
	}
}

func TestRunWithThrottleSurvivalApproverConsultedOncePerCall(t *testing.T) {
	oldSleep := throttleSleep
	throttleSleep = func(context.Context, time.Duration) error { return nil }
	defer func() { throttleSleep = oldSleep }()
	approver := &recordingApprover{decision: approvalGranted}
	attempts := 0
	_, _, err := runWithThrottleSurvival(context.Background(), "route", nil, throttleSurvivalOptions{approver: approver, taskTitle: "fix the parser"}, func() (string, error) {
		attempts++
		if attempts <= 3 {
			return "", &throttledRouteError{label: "route", underlying: errors.New("rate limit")}
		}
		return "ok", nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if approver.calls != 1 {
		t.Fatalf("approver consulted %d times, want 1", approver.calls)
	}
	if approver.lastReq.TaskTitle != "fix the parser" || approver.lastReq.Hold != 1 {
		t.Fatalf("bad request: %+v", approver.lastReq)
	}
}

func TestRunWithThrottleSurvivalDeclineReturnsTypedError(t *testing.T) {
	approver := &recordingApprover{decision: approvalDeclined}
	_, report, err := runWithThrottleSurvival(context.Background(), "route", func() string { return "local-fast" }, throttleSurvivalOptions{approver: approver}, func() (string, error) {
		return "", &throttledRouteError{label: "route", underlying: errors.New("rate limit")}
	})
	var declined *throttleDeclinedError
	if !errors.As(err, &declined) {
		t.Fatalf("expected throttleDeclinedError, got %v", err)
	}
	if report.Holds != 0 {
		t.Fatalf("declined call must not hold, got %d", report.Holds)
	}
	if !strings.Contains(err.Error(), "jini route set local-fast") {
		t.Fatalf("declined error must name fallback: %v", err)
	}
}

func TestRunWithThrottleSurvivalFailClosedApproverNeverPrompts(t *testing.T) {
	_, _, err := runWithThrottleSurvival(context.Background(), "route", nil, throttleSurvivalOptions{approver: failClosedApprover{}}, func() (string, error) {
		return "", &throttledRouteError{label: "route", underlying: errors.New("status 429")}
	})
	var declined *throttleDeclinedError
	if !errors.As(err, &declined) {
		t.Fatalf("expected declined, got %v", err)
	}
}

func TestRunWithThrottleSurvivalCustomHoldLadder(t *testing.T) {
	oldSleep := throttleSleep
	var waits []time.Duration
	throttleSleep = func(_ context.Context, wait time.Duration) error {
		waits = append(waits, wait)
		return nil
	}
	defer func() { throttleSleep = oldSleep }()
	attempts := 0
	_, _, err := runWithThrottleSurvival(context.Background(), "route", nil, throttleSurvivalOptions{approver: autoApprover{}, holdWaits: []time.Duration{20 * time.Second}}, func() (string, error) {
		attempts++
		return "", &throttledRouteError{label: "route", underlying: errors.New("rate limit")}
	})
	if err == nil {
		t.Fatal("expected exhaustion")
	}
	if len(waits) != 1 || waits[0] != 20*time.Second {
		t.Fatalf("single-hold ladder not honored: %v", waits)
	}
	if attempts != 2 {
		t.Fatalf("expected 2 attempts (1 + 1 hold), got %d", attempts)
	}
}

func TestRunWithThrottleSurvivalPerAttemptTimeout(t *testing.T) {
	_, _, err := runWithThrottleSurvival(context.Background(), "route", nil, throttleSurvivalOptions{approver: autoApprover{}, attemptTimeout: 10 * time.Millisecond}, func() (string, error) {
		// Attempt closure receives its budget via throttleAttemptContext.
		ctx := throttleAttemptContext()
		<-ctx.Done()
		return "", ctx.Err()
	})
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected deadline exceeded passthrough (non-throttle, no hold), got %v", err)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/app -run TestRunWithThrottleSurvival -v`
Expected: FAIL — `undefined: throttleSurvivalOptions` (compile error; existing tests also break on the signature — fix them in Step 3 by passing `throttleSurvivalOptions{approver: autoApprover{}}`).

- [ ] **Step 3: Implement (throttle_survival.go)**

Add the types from the Interfaces block verbatim, plus:

```go
func (e *throttleDeclinedError) Error() string {
	guidance := "The session is saved; `jini continue` resumes it."
	if e.fallbackHint != "" {
		guidance = fmt.Sprintf("The session is saved; `jini continue` resumes it, or switch with `jini route set %s`.", e.fallbackHint)
	}
	return fmt.Sprintf("%s resume was not approved. %s Underlying: %v", e.label, guidance, e.underlying)
}

func (e *throttleDeclinedError) Unwrap() error { return e.underlying }

func (autoApprover) Approve(context.Context, throttleApprovalRequest) (throttleApprovalDecision, error) {
	return approvalGranted, nil
}

func (failClosedApprover) Approve(context.Context, throttleApprovalRequest) (throttleApprovalDecision, error) {
	return approvalDeclined, nil
}

// throttleAttemptContext exposes the per-attempt context to the attempt
// closure without changing its signature; set for the duration of one
// attempt only. Single-goroutine per survival call by construction.
var currentAttemptContext context.Context = context.Background()

func throttleAttemptContext() context.Context { return currentAttemptContext }
```

Rework `runWithThrottleSurvival`:

```go
func runWithThrottleSurvival(ctx context.Context, label string, fallbackHint func() string, opts throttleSurvivalOptions, attempt func() (string, error)) (string, throttleSurvivalReport, error) {
	report := throttleSurvivalReport{}
	approver := opts.approver
	if approver == nil {
		approver = throttleApproverForProcess
	}
	holdWaits := opts.holdWaits
	if holdWaits == nil {
		holdWaits = throttleHoldWaits
	}
	var lastErr error
	hint := ""
	hintResolved := false
	resolveHint := func() string {
		if !hintResolved {
			hintResolved = true
			if fallbackHint != nil {
				hint = fallbackHint()
			}
		}
		return hint
	}
	runAttempt := func() (string, error) {
		if opts.attemptTimeout > 0 {
			attemptCtx, cancel := context.WithTimeout(ctx, opts.attemptTimeout)
			defer cancel()
			currentAttemptContext = attemptCtx
		} else {
			currentAttemptContext = ctx
		}
		defer func() { currentAttemptContext = context.Background() }()
		return attempt()
	}
	approved := false
	for hold := 0; hold <= len(holdWaits); hold++ {
		text, err := runAttempt()
		if err == nil {
			return text, report, nil
		}
		var throttled *throttledRouteError
		if !errors.As(err, &throttled) {
			return "", report, err
		}
		lastErr = err
		if hold == len(holdWaits) {
			break
		}
		wait := holdWaits[hold]
		if throttled.retryAfter > 0 {
			wait = throttled.retryAfter
			if wait > throttleRetryAfterCap {
				wait = throttleRetryAfterCap
			}
		}
		if !approved {
			decision, approveErr := approver.Approve(ctx, throttleApprovalRequest{
				Label:        label,
				Wait:         wait,
				FallbackHint: resolveHint(),
				Hold:         hold + 1,
				TaskTitle:    opts.taskTitle,
			})
			if approveErr != nil {
				return "", report, approveErr
			}
			if decision != approvalGranted {
				return "", report, &throttleDeclinedError{label: label, fallbackHint: resolveHint(), underlying: lastErr}
			}
			approved = true
		}
		narrateThrottleHold(label, resolveHint(), wait, hold+1, len(holdWaits))
		if sleepErr := throttleSleep(ctx, wait); sleepErr != nil {
			return "", report, sleepErr
		}
		report.Holds++
		report.TotalHeld += wait
	}
	return "", report, throttleExhaustionError(label, resolveHint(), report, lastErr)
}
```

Update the two call sites in `provider.go` (129, 155) to pass `throttleSurvivalOptions{taskTitle: request.Title}` (approver nil → process default; the sidecar and every existing path therefore keep Auto behavior). Update every existing test caller to pass `throttleSurvivalOptions{approver: autoApprover{}}`.

Note on `attemptTimeout` plumbing to the real providers: `generateProviderText` and `runCLIHandoff` already accept `ctx`. Where `attemptTimeout > 0`, the attempt closures must use `throttleAttemptContext()` instead of the captured outer `ctx` — change the closure bodies at provider.go:129-135 and provider.go:155-158 to `runCLIHandoff(throttleAttemptContext(), ...)` and `generateProviderText(throttleAttemptContext(), ...)`.

- [ ] **Step 4: Run the full package tests**

Run: `go test ./internal/app -run 'TestRunWithThrottleSurvival|TestIsThrottleSignal|TestClassifyCLIThrottleOutput' -v`
Expected: all PASS — including the untouched shipped tests, proving Auto-path equivalence.

- [ ] **Step 5: Gate + commit**

```bash
bash tools/run_required_gates.sh commit
git add internal/app/throttle_survival.go internal/app/throttle_survival_test.go internal/app/provider.go
git commit -m "feat: throttle survival approver seam — Auto byte-identical, typed decline

Co-Authored-By: Claude Fable 5 <noreply@anthropic.com>"
```

---

### Task 4: CLI prompt approver + dispatch-branch conduit

**Files:**
- Modify: `internal/app/throttle_survival.go` (add `cliPromptApprover`)
- Modify: `internal/app/app.go` (`runDirectTaskArgsIntake` ~1982 sets prompt approver; `runLauncher` and `runNewWorkIntake` set fail-closed)
- Modify: `internal/app/app_sidecar.go` (serve loop sets fail-closed)
- Test: `internal/app/throttle_survival_test.go`

**Interfaces:**
- Consumes: Task 3's approver types; Task 1's `effectiveExecutionMode()`.
- Produces:

```go
type cliPromptApprover struct{}
var (
	throttlePromptInput  io.Reader = os.Stdin
	throttlePromptOutput io.Writer = os.Stderr
	throttlePromptIsTTY  func() bool // default: term check on os.Stdin fd
)
func configureThrottleApproverForEntry(oneShot bool)
```

- [ ] **Step 1: Write the failing tests**

```go
func withPromptIO(t *testing.T, input string) *bytes.Buffer {
	t.Helper()
	oldIn, oldOut, oldTTY := throttlePromptInput, throttlePromptOutput, throttlePromptIsTTY
	out := &bytes.Buffer{}
	throttlePromptInput = strings.NewReader(input)
	throttlePromptOutput = out
	throttlePromptIsTTY = func() bool { return true }
	t.Cleanup(func() {
		throttlePromptInput, throttlePromptOutput, throttlePromptIsTTY = oldIn, oldOut, oldTTY
	})
	return out
}

func TestCLIPromptApproverExplicitYesGrants(t *testing.T) {
	out := withPromptIO(t, "y\n")
	decision, err := cliPromptApprover{}.Approve(context.Background(), throttleApprovalRequest{Label: "Claude API route", Wait: 20 * time.Second})
	if err != nil || decision != approvalGranted {
		t.Fatalf("decision=%v err=%v", decision, err)
	}
	if !strings.Contains(out.String(), "Resume automatically when capacity returns? [y/N]") {
		t.Fatalf("prompt copy wrong: %q", out.String())
	}
}

func TestCLIPromptApproverBareEnterDeclines(t *testing.T) {
	withPromptIO(t, "\n")
	decision, err := cliPromptApprover{}.Approve(context.Background(), throttleApprovalRequest{Label: "route"})
	if err != nil || decision != approvalDeclined {
		t.Fatalf("bare Enter must decline: decision=%v err=%v", decision, err)
	}
}

func TestCLIPromptApproverBufferedNewlineThenYesDeclinesOnFirstLine(t *testing.T) {
	// A stray buffered newline must not auto-approve: first line wins, and
	// it is empty → decline.
	withPromptIO(t, "\ny\n")
	decision, _ := cliPromptApprover{}.Approve(context.Background(), throttleApprovalRequest{Label: "route"})
	if decision != approvalDeclined {
		t.Fatal("stray newline auto-approved")
	}
}

func TestCLIPromptApproverNoTTYDeclinesWithoutPrompting(t *testing.T) {
	out := withPromptIO(t, "y\n")
	throttlePromptIsTTY = func() bool { return false }
	decision, err := cliPromptApprover{}.Approve(context.Background(), throttleApprovalRequest{Label: "route"})
	if err != nil || decision != approvalDeclined {
		t.Fatalf("no-TTY must decline: %v %v", decision, err)
	}
	if out.Len() != 0 {
		t.Fatalf("must not prompt without a TTY: %q", out.String())
	}
}

func TestCLIPromptApproverCtxCancelDeclines(t *testing.T) {
	withPromptIO(t, "") // reader that never yields a line
	throttlePromptInput = blockingReader{}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := cliPromptApprover{}.Approve(ctx, throttleApprovalRequest{Label: "route"})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected ctx cancellation, got %v", err)
	}
}

type blockingReader struct{}

func (blockingReader) Read([]byte) (int, error) { select {} }

func TestConfigureThrottleApproverForEntry(t *testing.T) {
	withExecutionModeHome(t)
	oldApprover := throttleApproverForProcess
	defer func() { throttleApproverForProcess = oldApprover }()
	if err := saveExecutionMode("ask"); err != nil {
		t.Fatal(err)
	}
	configureThrottleApproverForEntry(true)
	if _, ok := throttleApproverForProcess.(cliPromptApprover); !ok {
		t.Fatalf("one-shot ask must get cliPromptApprover, got %T", throttleApproverForProcess)
	}
	configureThrottleApproverForEntry(false)
	if _, ok := throttleApproverForProcess.(failClosedApprover); !ok {
		t.Fatalf("launcher ask must get failClosedApprover, got %T", throttleApproverForProcess)
	}
	if err := saveExecutionMode("auto"); err != nil {
		t.Fatal(err)
	}
	configureThrottleApproverForEntry(true)
	if _, ok := throttleApproverForProcess.(autoApprover); !ok {
		t.Fatalf("auto must keep autoApprover, got %T", throttleApproverForProcess)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/app -run 'TestCLIPromptApprover|TestConfigureThrottleApprover' -v`
Expected: FAIL — `undefined: cliPromptApprover`.

- [ ] **Step 3: Implement**

Append to throttle_survival.go:

```go
type cliPromptApprover struct{}

var (
	throttlePromptInput  io.Reader = os.Stdin
	throttlePromptOutput io.Writer = os.Stderr
	throttlePromptIsTTY  func() bool = func() bool {
		info, err := os.Stdin.Stat()
		if err != nil {
			return false
		}
		return info.Mode()&os.ModeCharDevice != 0
	}
)

func (cliPromptApprover) Approve(ctx context.Context, req throttleApprovalRequest) (throttleApprovalDecision, error) {
	if throttlePromptIsTTY == nil || !throttlePromptIsTTY() {
		return approvalDeclined, nil
	}
	fmt.Fprintf(throttlePromptOutput, "%s is throttled (retry in %s). Resume automatically when capacity returns? [y/N] ", req.Label, req.Wait)
	type answer struct {
		line string
		err  error
	}
	answers := make(chan answer, 1)
	go func() {
		// The first line wins: a stray buffered newline reads as an empty
		// line and declines — it must never auto-approve. This goroutine
		// leaks if ctx cancels while blocked on stdin; accepted for a
		// one-shot process.
		reader := bufio.NewReader(throttlePromptInput)
		line, err := reader.ReadString('\n')
		answers <- answer{line: line, err: err}
	}()
	select {
	case <-ctx.Done():
		return approvalDeclined, ctx.Err()
	case got := <-answers:
		if got.err != nil && strings.TrimSpace(got.line) == "" {
			return approvalDeclined, nil
		}
		switch strings.ToLower(strings.TrimSpace(got.line)) {
		case "y", "yes":
			return approvalGranted, nil
		default:
			return approvalDeclined, nil
		}
	}
}

// configureThrottleApproverForEntry sets the process-level approver at a
// dispatch branch. oneShot must be true ONLY where jini owns stdin (the
// direct-task and standalone-question one-shot paths) — never for the
// interactive launcher (its Scanner owns stdin) or the sidecar (stdin is a
// protocol pipe).
func configureThrottleApproverForEntry(oneShot bool) {
	if effectiveExecutionMode() != executionModeAsk {
		throttleApproverForProcess = autoApprover{}
		return
	}
	if oneShot {
		throttleApproverForProcess = cliPromptApprover{}
		return
	}
	throttleApproverForProcess = failClosedApprover{}
}
```

Add `"bufio"` to the imports.

Set-sites (one line each, first statement of the function):
- `runDirectTaskArgsIntake` (app.go ~1982): `configureThrottleApproverForEntry(true)`
- `runLauncher`: `configureThrottleApproverForEntry(false)`
- `runNewWorkIntake`: `configureThrottleApproverForEntry(false)`
- sidecar serve entry (app_sidecar.go, the function containing the scanner loop at ~57): `configureThrottleApproverForEntry(false)`

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/app -run 'TestCLIPromptApprover|TestConfigureThrottleApprover' -v`
Expected: all PASS.

- [ ] **Step 5: Gate + commit**

```bash
bash tools/run_required_gates.sh commit
git add internal/app/throttle_survival.go internal/app/throttle_survival_test.go internal/app/app.go internal/app/app_sidecar.go
git commit -m "feat: Ask-mode throttle-resume prompt on one-shot path, fail-closed elsewhere

Co-Authored-By: Claude Fable 5 <noreply@anthropic.com>"
```

---

### Task 5: Park record + `jini continue` + caller integration

**Files:**
- Create: `internal/app/throttle_park.go`
- Modify: `internal/app/provider.go` (`generateWithConfiguredProviderDecision`: park pre-write/update/clear)
- Modify: `internal/app/simple_answers.go` (standalone profile, error passthrough)
- Modify: `internal/app/app.go` (`runContinue` ~3165 checks park first; `saveCurrentWork` ~2506 clears park)
- Test: `internal/app/throttle_park_test.go`

**Interfaces:**
- Consumes: Task 3's `throttleDeclinedError`, options struct; Task 1's `effectiveExecutionMode()`.
- Produces:

```go
type throttlePark struct {
	SchemaVersion string `json:"schema_version"`
	ContextType   string `json:"context_type"` // "JiniThrottlePark"
	Prompt        string `json:"prompt"`
	RouteLabel    string `json:"route_label"`
	Reason        string `json:"reason"`
	FallbackHint  string `json:"fallback_hint"`
	ParkedAt      string `json:"parked_at"` // RFC3339
}
func writeThrottlePark(prompt, routeLabel string) error
func updateThrottleParkError(reason, fallbackHint string)
func loadThrottlePark() *throttlePark // nil on absent or corrupt
func clearThrottlePark()
```

- [ ] **Step 1: Write the failing tests**

```go
package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestThrottleParkRoundTrip(t *testing.T) {
	t.Setenv("JINI_STATE_DIR", t.TempDir())
	if err := writeThrottlePark("refactor the parser", "Claude API route"); err != nil {
		t.Fatal(err)
	}
	updateThrottleParkError("throttled: status 429", "local-fast")
	park := loadThrottlePark()
	if park == nil {
		t.Fatal("park not loaded")
	}
	if park.Prompt != "refactor the parser" || park.FallbackHint != "local-fast" {
		t.Fatalf("bad park: %+v", park)
	}
	if _, err := time.Parse(time.RFC3339, park.ParkedAt); err != nil {
		t.Fatalf("bad timestamp: %v", err)
	}
	info, err := os.Stat(filepath.Join(sessionStateRoot(), "throttle-park.json"))
	if err != nil {
		t.Fatal(err)
	}
	if perm := info.Mode().Perm(); perm != 0o600 {
		t.Fatalf("expected 0600, got %o", perm)
	}
	clearThrottlePark()
	if loadThrottlePark() != nil {
		t.Fatal("park not cleared")
	}
}

func TestThrottleParkCorruptFileLoadsNil(t *testing.T) {
	t.Setenv("JINI_STATE_DIR", t.TempDir())
	if err := os.MkdirAll(sessionStateRoot(), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sessionStateRoot(), "throttle-park.json"), []byte("{corrupt"), 0o600); err != nil {
		t.Fatal(err)
	}
	if loadThrottlePark() != nil {
		t.Fatal("corrupt park must load as nil")
	}
}

func TestThrottleParkAgeDisclosure(t *testing.T) {
	old := throttlePark{SchemaVersion: "0.1.0", ContextType: "JiniThrottlePark", Prompt: "old task", ParkedAt: time.Now().Add(-72 * time.Hour).Format(time.RFC3339)}
	line := throttleParkResumeLine(&old)
	if !strings.Contains(line, "parked 3 days ago") {
		t.Fatalf("age not disclosed: %q", line)
	}
	fresh := throttlePark{ParkedAt: time.Now().Format(time.RFC3339), Prompt: "new task"}
	if strings.Contains(throttleParkResumeLine(&fresh), "parked") {
		t.Fatal("fresh park must not disclose age")
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/app -run TestThrottlePark -v`
Expected: FAIL — `undefined: writeThrottlePark`.

- [ ] **Step 3: Implement `throttle_park.go`**

```go
package app

// Throttle park record — specs/auto-ask-execution-mode-design.md.
// CALLER-owned: pre-written before runWithThrottleSurvival on Ask paths,
// updated with reason/fallback on the error return, deleted only on SUCCESS
// (never on approval grant), so Ctrl-C at any stage leaves work resumable.
// One slot per state dir: concurrent processes are last-writer-wins on
// write, and one process's delete-on-success may remove another's park —
// accepted loss modes for a single-user CLI. Atomic rename prevents tearing.

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type throttlePark struct {
	SchemaVersion string `json:"schema_version"`
	ContextType   string `json:"context_type"`
	Prompt        string `json:"prompt"`
	RouteLabel    string `json:"route_label"`
	Reason        string `json:"reason"`
	FallbackHint  string `json:"fallback_hint"`
	ParkedAt      string `json:"parked_at"`
}

func throttleParkPath() string {
	return filepath.Join(sessionStateRoot(), "throttle-park.json")
}

func writeThrottlePark(prompt, routeLabel string) error {
	park := throttlePark{
		SchemaVersion: "0.1.0",
		ContextType:   "JiniThrottlePark",
		Prompt:        prompt,
		RouteLabel:    routeLabel,
		ParkedAt:      time.Now().Format(time.RFC3339),
	}
	return persistThrottlePark(park)
}

func updateThrottleParkError(reason, fallbackHint string) {
	park := loadThrottlePark()
	if park == nil {
		return
	}
	park.Reason = reason
	park.FallbackHint = fallbackHint
	_ = persistThrottlePark(*park)
}

func persistThrottlePark(park throttlePark) error {
	if err := os.MkdirAll(sessionStateRoot(), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(park, "", "  ")
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(sessionStateRoot(), "throttle-park-*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	if _, err := tmp.Write(append(data, '\n')); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return err
	}
	if err := tmp.Chmod(0o600); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return err
	}
	return os.Rename(tmpName, throttleParkPath())
}

func loadThrottlePark() *throttlePark {
	data, err := os.ReadFile(throttleParkPath())
	if err != nil {
		return nil
	}
	var park throttlePark
	if err := json.Unmarshal(data, &park); err != nil {
		return nil
	}
	if park.ContextType != "JiniThrottlePark" {
		return nil
	}
	return &park
}

func clearThrottlePark() {
	_ = os.Remove(throttleParkPath())
}

func throttleParkResumeLine(park *throttlePark) string {
	excerpt := park.Prompt
	if len(excerpt) > 60 {
		excerpt = excerpt[:57] + "..."
	}
	parkedAt, err := time.Parse(time.RFC3339, park.ParkedAt)
	if err == nil {
		if age := time.Since(parkedAt); age > time.Hour {
			days := int(age.Hours() / 24)
			if days >= 1 {
				return fmt.Sprintf("parked %d days ago: %s — resuming", days, excerpt)
			}
			return fmt.Sprintf("parked %d hours ago: %s — resuming", int(age.Hours()), excerpt)
		}
	}
	return fmt.Sprintf("Resuming: %s", excerpt)
}
```

- [ ] **Step 4: Wire the callers**

In `generateWithConfiguredProviderDecision` (provider.go), at the top of both route branches (CLI handoff before `runWithThrottleSurvival` at ~129, provider before ~155), when Ask mode and the prompt is non-empty:

```go
	askMode := effectiveExecutionMode() == executionModeAsk
	parkPrompt := strings.TrimSpace(request.Source)
	if askMode && parkPrompt != "" {
		_ = writeThrottlePark(parkPrompt, cliHandoffLabel(decision.ToolMode)) // provider branch: routeLabel
	}
```

On the error return of each branch, before surfacing the error, when the error is throttle-family:

```go
	if askMode && parkPrompt != "" && isThrottleFamilyError(err) {
		updateThrottleParkError(err.Error(), "")
		return "", true, decision, err
	}
	if askMode && parkPrompt != "" {
		clearThrottlePark() // non-throttle error: nothing to resume
	}
```

On the success return of each branch: `if askMode && parkPrompt != "" { clearThrottlePark() }`.

Add to throttle_survival.go:

```go
func isThrottleFamilyError(err error) bool {
	var throttled *throttledRouteError
	var declined *throttleDeclinedError
	return errors.As(err, &throttled) || errors.As(err, &declined) || isThrottleExhaustionError(err)
}
```

Make exhaustion typed so the family check works — replace `throttleExhaustionError`'s `fmt.Errorf` with:

```go
type throttleExhaustedError struct{ message string; underlying error }

func (e *throttleExhaustedError) Error() string { return e.message }
func (e *throttleExhaustedError) Unwrap() error { return e.underlying }

func isThrottleExhaustionError(err error) bool {
	var exhausted *throttleExhaustedError
	return errors.As(err, &exhausted)
}

func throttleExhaustionError(label, fallbackHint string, report throttleSurvivalReport, lastErr error) error {
	guidance := "The session is saved; `jini continue` resumes it."
	if fallbackHint != "" {
		guidance = fmt.Sprintf("The session is saved; `jini continue` resumes it, or switch with `jini route set %s`.", fallbackHint)
	}
	return &throttleExhaustedError{
		message:    fmt.Sprintf("%s is still throttled after %d automatic resume attempts (held %s total). %s Underlying: %v", label, report.Holds+1, report.TotalHeld, guidance, lastErr),
		underlying: lastErr,
	}
}
```

In `standaloneQuestionSetupMessage` (simple_answers.go:48), insert as the FIRST check:

```go
	if err != nil && isThrottleFamilyError(err) {
		return err.Error()
	}
```

In `maybeHandleStandaloneQuestion` (simple_answers.go:37), replace the whole-call deadline with a cancellable ctx and pass the standalone profile via the request:

```go
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	request.Standalone = true
```

Add `Standalone bool` to `providerGenerationRequest` (find the struct with `grep -n "type providerGenerationRequest" internal/app/*.go`). In `generateWithConfiguredProviderDecision`, build the options once at the top:

```go
	opts := throttleSurvivalOptions{taskTitle: request.Title}
	if request.Standalone {
		opts.attemptTimeout = standaloneQuestionTimeout()
		opts.holdWaits = []time.Duration{20 * time.Second}
	}
```

and pass `opts` at both `runWithThrottleSurvival` call sites.

In `runContinue` (app.go:3165), insert before `resolveSummary`:

```go
	if park := loadThrottlePark(); park != nil {
		fmt.Fprintln(stdout, throttleParkResumeLine(park))
		code := runDirectTaskArgsIntake([]string{park.Prompt}, stdout, stderr)
		if code == 0 {
			clearThrottlePark()
		}
		return code
	}
```

In `saveCurrentWork` (app.go:2506), first line of the function: `clearThrottlePark()` — new work always clears a stale park.

- [ ] **Step 5: Add integration tests**

```go
func TestStandaloneThrottleFamilyErrorPassesThrough(t *testing.T) {
	declined := &throttleDeclinedError{label: "Claude API route", fallbackHint: "local-fast", underlying: errors.New("rate limit")}
	message := standaloneQuestionSetupMessage(routeDecision{ToolMode: "claude-api"}, declined)
	if !strings.Contains(message, "jini continue") || !strings.Contains(message, "jini route set local-fast") {
		t.Fatalf("throttle-family error was swallowed: %q", message)
	}
}

func TestRunContinueResumesPark(t *testing.T) {
	t.Setenv("JINI_STATE_DIR", t.TempDir())
	if err := writeThrottlePark("what is the capital of France?", "route"); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	code := runContinue(&out, io.Discard)
	if code != 0 {
		t.Fatalf("exit %d: %s", code, out.String())
	}
	if !strings.Contains(out.String(), "Paris.") {
		t.Fatalf("parked prompt not re-run: %q", out.String())
	}
	if loadThrottlePark() != nil {
		t.Fatal("park not cleared after successful resume")
	}
}

func TestSaveCurrentWorkClearsStalePark(t *testing.T) {
	t.Setenv("JINI_STATE_DIR", t.TempDir())
	if err := writeThrottlePark("stale", "route"); err != nil {
		t.Fatal(err)
	}
	// Any successful new-work save clears the park; construct the minimal
	// currentWork used elsewhere in app_internal_test.go.
	if err := saveCurrentWork(&currentWork{PackDir: t.TempDir()}); err != nil {
		t.Fatal(err)
	}
	if loadThrottlePark() != nil {
		t.Fatal("stale park survived new work")
	}
}
```

(If `runContinue`'s park branch routes the resumed prompt through the simple-answer path, the capital-of-France park resolves locally with no provider — that is the point of the test fixture. Verify `runDirectTaskArgsIntake` handles a plain question; it calls `maybeHandleStandaloneQuestion` at app.go:1997 and `maybeHandleSimpleAnswer` upstream — adjust the assertion to the actual local answer path if output differs, but the park-cleared assertion is non-negotiable.)

- [ ] **Step 6: Run the full package**

Run: `go test ./internal/app`
Expected: PASS.

- [ ] **Step 7: Gate + commit**

```bash
bash tools/run_required_gates.sh commit
git add internal/app/throttle_park.go internal/app/throttle_park_test.go internal/app/provider.go internal/app/simple_answers.go internal/app/app.go internal/app/throttle_survival.go internal/app/throttle_survival_test.go
git commit -m "feat: throttle park record — jini continue resumes parked throttled work

Co-Authored-By: Claude Fable 5 <noreply@anthropic.com>"
```

---

### Task 6: Live transcript evidence + trace update

**Files:**
- Modify: `specs/prd-implementation-trace.md` (move Auto/Ask + Ask-approval slices from backlog to Implemented; leave Autopilot switching + dodge counter in backlog)
- Evidence: transcript in the commit message / `.jini` evidence per repo convention

**Interfaces:**
- Consumes: everything above, built binary.
- Produces: release-claimable evidence per the repo's test-evidence doctrine.

- [ ] **Step 1: Build and run the live transcript**

```bash
go build -o /tmp/jini-mode-test ./cmd/jini
# Fake throttling downstream CLI (same harness family as 5e5d8d7's transcript):
cat > /tmp/fake-throttle-cli <<'EOF'
#!/bin/bash
STATE=/tmp/fake-throttle-state
if [ ! -f "$STATE" ]; then
  touch "$STATE"
  echo "error: rate limit exceeded, retry after 2 seconds" >&2
  exit 1
fi
rm -f "$STATE"
echo "resumed answer after throttle"
EOF
chmod +x /tmp/fake-throttle-cli
rm -f /tmp/fake-throttle-state

# Ask mode, decline path:
/tmp/jini-mode-test mode ask
printf 'n\n' | JINI_CLI_HANDOFF_SKIP_TRUST_CHECK=1 JINI_TOOL=codex JINI_CODEX_CLI=/tmp/fake-throttle-cli /tmp/jini-mode-test "summarize the release notes"
# Expected: [y/N] prompt on stderr, decline, error naming `jini continue`; throttle-park.json exists.
/tmp/jini-mode-test continue
# Expected: parked prompt re-runs, fake CLI succeeds on second call, park cleared.

# Ask mode, grant path:
rm -f /tmp/fake-throttle-state
printf 'y\n' | JINI_CLI_HANDOFF_SKIP_TRUST_CHECK=1 JINI_TOOL=codex JINI_CODEX_CLI=/tmp/fake-throttle-cli /tmp/jini-mode-test "summarize the release notes"
# Expected: prompt, grant, 2s advertised hold honored, same-route resume, answer printed.

/tmp/jini-mode-test mode auto
```

Capture the verbatim terminal output for the commit message and evidence records.

- [ ] **Step 2: Update the trace**

In `specs/prd-implementation-trace.md`: add an Implemented row for the Auto/Ask mode surface + Ask-mode throttle-resume approval naming `execution_mode.go`, `throttle_survival.go` approver seam, `throttle_park.go`, and the test names; shrink the backlog bullet to the remaining slices (paid Autopilot mid-task switching, throttle-dodge counter). Add the consistency/refine-draft throttle bypass to Residual hardening.

- [ ] **Step 3: Gate + commit**

```bash
bash tools/run_required_gates.sh commit
git add specs/prd-implementation-trace.md
git commit -m "docs: trace Auto/Ask mode + Ask-resume approval as implemented with transcript evidence

Co-Authored-By: Claude Fable 5 <noreply@anthropic.com>"
```

---

## Self-Review Notes

- Spec coverage: mode file + fail-closed parsing (Task 1); `jini mode` + registration + allowed map + settling amendment (Task 2); approver interface + options + ladder/attempt timeout + typed decline/exhaustion (Task 3); prompt approver + conduit set-sites (Task 4); park + lifecycle + `jini continue` + copy discipline + standalone passthrough/profile (Task 5); transcript + trace (Task 6). The design's "Known exemption" (consistency/refine bypass) is recorded in Task 6 Step 2 as residual hardening — deliberately not fixed here.
- Type consistency: `throttleSurvivalOptions`, `throttleApprovalRequest`, `throttleDeclinedError`, `throttleExhaustedError`, `configureThrottleApproverForEntry`, park function names are identical across Tasks 3–5.
- Anchors are line-approximate (concurrent dogfood-fix diff will shift them); every modify instruction also names the function, which is authoritative.
