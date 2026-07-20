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

// runWithThrottleSurvival runs attempt, and on throttle errors holds the
// session and retries the same route automatically. It never waits on
// non-throttle errors. fallbackHint is evaluated lazily on the first hold and
// names a viable fallback route for narration and the exhaustion error.
func runWithThrottleSurvival(ctx context.Context, label string, fallbackHint func() string, attempt func() (string, error)) (string, throttleSurvivalReport, error) {
	report := throttleSurvivalReport{}
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
	for hold := 0; hold <= len(throttleHoldWaits); hold++ {
		text, err := attempt()
		if err == nil {
			return text, report, nil
		}
		var throttled *throttledRouteError
		if !errors.As(err, &throttled) {
			return "", report, err
		}
		lastErr = err
		if hold == len(throttleHoldWaits) {
			break
		}
		wait := throttleHoldWaits[hold]
		if throttled.retryAfter > 0 {
			wait = throttled.retryAfter
			if wait > throttleRetryAfterCap {
				wait = throttleRetryAfterCap
			}
		}
		narrateThrottleHold(label, resolveHint(), wait, hold+1, len(throttleHoldWaits))
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
