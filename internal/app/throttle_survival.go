package app

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Autonomous throttle survival — the number one P0 requirement in
// specs/number-one-platform-prd.md. When a route is throttled or
// quota-limited, Jini detects it, holds the session, and resumes work on its
// own without a human babysitting the terminal. The free tier self-resumes on
// the same route the moment capacity returns and names viable fallbacks.

// throttledRouteError marks an error as a throttle/quota event on a route.
type throttledRouteError struct {
	label      string
	retryAfter time.Duration
	underlying error
}

func (e *throttledRouteError) Error() string {
	if e.retryAfter > 0 {
		return fmt.Sprintf("%s is throttled (retry after %s): %v", e.label, e.retryAfter, e.underlying)
	}
	return fmt.Sprintf("%s is throttled: %v", e.label, e.underlying)
}

func (e *throttledRouteError) Unwrap() error {
	return e.underlying
}

// Injection points so tests can observe holds without real waiting.
var (
	throttleHoldWaits               = []time.Duration{20 * time.Second, 40 * time.Second, 80 * time.Second}
	throttleRetryAfterCap           = 10 * time.Minute
	throttleSleep                   = sleepWithContext
	throttleNarration     io.Writer = os.Stderr
)

var throttleSignalFragments = []string{
	"rate limit",
	"rate-limit",
	"ratelimit",
	"too many requests",
	"usage limit",
	"quota exceeded",
	"quota has been exhausted",
	"insufficient quota",
	"resource exhausted",
	"resource_exhausted",
	"overloaded",
	"capacity constraints",
	"throttl",
	"status 429",
	"status: 429",
	"code 429",
	"http 429",
	"429 too many",
	"error 429",
	"status 529",
	"limit reached",
	"limit will reset",
}

// isThrottleSignal reports whether text looks like a throttle/quota event.
func isThrottleSignal(text string) bool {
	message := strings.ToLower(text)
	return containsAny(message, throttleSignalFragments)
}

var throttleRetryAfterPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)retry[- ]after[:\s]+(\d+)\s*(seconds?|secs?|s\b|minutes?|mins?|m\b)?`),
	regexp.MustCompile(`(?i)try again in\s+(\d+)\s*(seconds?|secs?|s\b|minutes?|mins?|m\b)`),
	regexp.MustCompile(`(?i)resets? in\s+(\d+)\s*(seconds?|secs?|s\b|minutes?|mins?|m\b)`),
}

// throttleRetryAfter extracts an advertised wait from throttle output.
func throttleRetryAfter(text string) (time.Duration, bool) {
	for _, pattern := range throttleRetryAfterPatterns {
		match := pattern.FindStringSubmatch(text)
		if match == nil {
			continue
		}
		value, err := strconv.Atoi(match[1])
		if err != nil || value <= 0 {
			continue
		}
		unit := time.Second
		if len(match) > 2 && strings.HasPrefix(strings.ToLower(strings.TrimSpace(match[2])), "m") {
			unit = time.Minute
		}
		wait := time.Duration(value) * unit
		if wait > throttleRetryAfterCap {
			wait = throttleRetryAfterCap
		}
		return wait, true
	}
	return 0, false
}

// classifyThrottleError wraps err as a throttledRouteError when its message
// looks like a throttle event. Provider SDK errors carry status text, so the
// message is the detection surface.
func classifyThrottleError(label string, err error) error {
	if err == nil {
		return nil
	}
	var already *throttledRouteError
	if errors.As(err, &already) {
		return err
	}
	if !isThrottleSignal(err.Error()) {
		return err
	}
	wait, _ := throttleRetryAfter(err.Error())
	return &throttledRouteError{label: label, retryAfter: wait, underlying: err}
}

// classifyCLIThrottleOutput inspects raw downstream CLI output (which never
// leaves this process — receipts keep only char counts) and wraps the
// sanitized execution error as a throttle event when the output says so.
func classifyCLIThrottleOutput(label, stdout, stderr string, sanitized error) error {
	if sanitized == nil {
		return nil
	}
	combined := stdout + "\n" + stderr
	if !isThrottleSignal(combined) && !isThrottleSignal(sanitized.Error()) {
		return sanitized
	}
	wait, _ := throttleRetryAfter(combined)
	return &throttledRouteError{label: label, retryAfter: wait, underlying: sanitized}
}

type throttleSurvivalReport struct {
	Holds     int
	TotalHeld time.Duration
}

// throttleFallbackHint names one viable fallback route for narration and
// exhaustion guidance; empty when none is ready. Called lazily — only when a
// throttle actually happens — so normal requests never pay the probe.
func throttleFallbackHint(request providerGenerationRequest, decision routeDecision) func() string {
	return func() string {
		availability := detectRuntimeAvailability(request)
		mode := strings.TrimSpace(availability.OfflineRouteMode)
		if mode == "" || mode == decision.ToolMode || mode == "local-preview" {
			return ""
		}
		return mode
	}
}

// throttleApprovalRequest describes one pending resume so an approver can put
// the decision to whoever is supervising the session.
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

// throttleApprover decides whether a throttled route may hold and resume.
// Ask mode swaps in a prompting approver; paid Autopilot and remote mobile
// approvals are further implementations of this same seam.
type throttleApprover interface {
	Approve(ctx context.Context, req throttleApprovalRequest) (throttleApprovalDecision, error)
}

// autoApprover is the free-tier Auto-mode default: resume without asking.
type autoApprover struct{}

// failClosedApprover declines rather than assuming autonomy. Mode parsing
// degrades to this so an unreadable setting costs supervision, not safety.
type failClosedApprover struct{}

func (autoApprover) Approve(context.Context, throttleApprovalRequest) (throttleApprovalDecision, error) {
	return approvalGranted, nil
}

func (failClosedApprover) Approve(context.Context, throttleApprovalRequest) (throttleApprovalDecision, error) {
	return approvalDeclined, nil
}

var throttleApproverForProcess throttleApprover = autoApprover{}

type throttleSurvivalOptions struct {
	approver       throttleApprover // nil = throttleApproverForProcess
	attemptTimeout time.Duration    // 0 = no per-attempt timeout
	holdWaits      []time.Duration  // nil = throttleHoldWaits
	taskTitle      string
}

// throttleDeclinedError reports that a resume was offered and refused. It is
// distinct from exhaustion: nothing was held, and the work is parked.
type throttleDeclinedError struct {
	label        string
	fallbackHint string
	underlying   error
}

func (e *throttleDeclinedError) Error() string {
	guidance := "The session is saved; `jini continue` resumes it."
	if e.fallbackHint != "" {
		guidance = fmt.Sprintf("The session is saved; `jini continue` resumes it, or switch with `jini route set %s`.", e.fallbackHint)
	}
	return fmt.Sprintf("%s resume was not approved. %s Underlying: %v", e.label, guidance, e.underlying)
}

func (e *throttleDeclinedError) Unwrap() error { return e.underlying }

// currentAttemptContext exposes the per-attempt context to the attempt closure
// without changing its signature. One survival call runs its attempts on a
// single goroutine by construction, so this is set for the duration of one
// attempt only.
var currentAttemptContext context.Context = context.Background()

func throttleAttemptContext() context.Context { return currentAttemptContext }

// runWithThrottleSurvival runs attempt, and on throttle errors holds the
// session and retries the same route automatically. It never waits on
// non-throttle errors. fallbackHint is evaluated lazily on the first hold and
// names a viable fallback route for narration and the exhaustion error. The
// approver is consulted once per call, before the first hold: Auto grants
// silently, so its narration and timing are byte-identical to the pre-seam
// behavior.
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

func narrateThrottleHold(label, fallbackHint string, wait time.Duration, attempt, total int) {
	if throttleNarration == nil {
		return
	}
	suffix := ""
	if fallbackHint != "" {
		suffix = fmt.Sprintf(" Fallback available: %s.", fallbackHint)
	}
	fmt.Fprintf(
		throttleNarration,
		"%s is throttled. Holding the session; resuming automatically in %s (hold %d of %d).%s\n",
		label,
		wait,
		attempt,
		total,
		suffix,
	)
}

func throttleExhaustionError(label, fallbackHint string, report throttleSurvivalReport, lastErr error) error {
	guidance := "The session is saved; `jini continue` resumes it."
	if fallbackHint != "" {
		guidance = fmt.Sprintf(
			"The session is saved; `jini continue` resumes it, or switch with `jini route set %s`.",
			fallbackHint,
		)
	}
	return fmt.Errorf(
		"%s is still throttled after %d automatic resume attempts (held %s total). %s Underlying: %v",
		label,
		report.Holds+1,
		report.TotalHeld,
		guidance,
		lastErr,
	)
}

// appendThrottleSurvivalReason records autonomous survival on the route
// decision so receipts and trust surfaces show the dodge.
func appendThrottleSurvivalReason(reason string, report throttleSurvivalReport) string {
	if report.Holds == 0 {
		return reason
	}
	note := fmt.Sprintf(
		"Route was throttled; Jini held the session for %s across %d automatic holds and resumed on the same route without human intervention.",
		report.TotalHeld,
		report.Holds,
	)
	if strings.TrimSpace(reason) == "" {
		return note
	}
	return strings.TrimSpace(reason) + " " + note
}

func sleepWithContext(ctx context.Context, wait time.Duration) error {
	if wait <= 0 {
		return nil
	}
	timer := time.NewTimer(wait)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
