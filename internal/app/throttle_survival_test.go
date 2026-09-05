package app

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"
)

func withThrottleTestHarness(t *testing.T) (*bytes.Buffer, *[]time.Duration) {
	t.Helper()
	narration := &bytes.Buffer{}
	sleeps := &[]time.Duration{}
	previousNarration := throttleNarration
	previousSleep := throttleSleep
	previousWaits := throttleHoldWaits
	throttleNarration = narration
	throttleSleep = func(ctx context.Context, wait time.Duration) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		*sleeps = append(*sleeps, wait)
		return nil
	}
	throttleHoldWaits = []time.Duration{20 * time.Second, 40 * time.Second, 80 * time.Second}
	t.Cleanup(func() {
		throttleNarration = previousNarration
		throttleSleep = previousSleep
		throttleHoldWaits = previousWaits
	})
	return narration, sleeps
}

func TestIsThrottleSignalClassifiesThrottleAndQuotaEvents(t *testing.T) {
	for _, text := range []string{
		"anthropic API error: status 429 Too Many Requests",
		"Rate limit reached for requests",
		"You have exceeded your usage limit. Limit will reset at 3pm.",
		"ThrottlingException: rate exceeded",
		"RESOURCE_EXHAUSTED: quota exceeded for model",
		"the model is currently overloaded (status 529)",
	} {
		if !isThrottleSignal(text) {
			t.Fatalf("expected throttle signal for %q", text)
		}
	}
	for _, text := range []string{
		"connection refused",
		"invalid api key",
		"file not found",
		"context deadline exceeded",
		"model returned malformed JSON",
	} {
		if isThrottleSignal(text) {
			t.Fatalf("expected no throttle signal for %q", text)
		}
	}
}

func TestThrottleRetryAfterParsesAdvertisedWaits(t *testing.T) {
	for _, testCase := range []struct {
		text string
		want time.Duration
	}{
		{"retry after 30 seconds", 30 * time.Second},
		{"Retry-After: 15", 15 * time.Second},
		{"please try again in 2 minutes", 2 * time.Minute},
		{"limit resets in 90 seconds", 90 * time.Second},
	} {
		got, ok := throttleRetryAfter(testCase.text)
		if !ok || got != testCase.want {
			t.Fatalf("throttleRetryAfter(%q) = %s, %v; want %s", testCase.text, got, ok, testCase.want)
		}
	}
	if wait, ok := throttleRetryAfter("retry after 45 minutes"); !ok || wait != throttleRetryAfterCap {
		t.Fatalf("expected advertised wait above cap to clamp to %s, got %s (%v)", throttleRetryAfterCap, wait, ok)
	}
	if _, ok := throttleRetryAfter("no advertised wait here"); ok {
		t.Fatalf("expected no advertised wait")
	}
}

// A throttle with no configured fallback must not just wait silently — it must
// tell the user how to become resilient, or "keeps work moving" is unmet for a
// single-provider setup.
func TestThrottleHoldWithoutFallbackNudgesSetup(t *testing.T) {
	narration, _ := withThrottleTestHarness(t)
	attempts := 0
	_, _, err := runWithThrottleSurvival(context.Background(), "claude-code", func() string { return "" }, throttleSurvivalOptions{approver: autoApprover{}}, func() (string, error) {
		attempts++
		if attempts == 1 {
			return "", &throttledRouteError{label: "claude-code", underlying: errors.New("status 429")}
		}
		return "ok", nil
	})
	if err != nil {
		t.Fatalf("expected resume, got %v", err)
	}
	out := narration.String()
	for _, want := range []string{"No fallback route is configured", "jini route help"} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected no-fallback setup nudge %q; got:\n%s", want, out)
		}
	}
}

func TestRunWithThrottleSurvivalHoldsAndResumesSameRoute(t *testing.T) {
	narration, sleeps := withThrottleTestHarness(t)

	attempts := 0
	text, report, err := runWithThrottleSurvival(context.Background(), "claude-code", func() string { return "local-slm" }, throttleSurvivalOptions{approver: autoApprover{}}, func() (string, error) {
		attempts++
		if attempts <= 2 {
			return "", &throttledRouteError{label: "claude-code", underlying: errors.New("status 429")}
		}
		return "resumed result", nil
	})
	if err != nil {
		t.Fatalf("expected autonomous resume, got error: %v", err)
	}
	if text != "resumed result" {
		t.Fatalf("expected resumed result, got %q", text)
	}
	if attempts != 3 {
		t.Fatalf("expected 3 attempts, got %d", attempts)
	}
	if report.Holds != 2 || report.TotalHeld != 60*time.Second {
		t.Fatalf("expected 2 holds totaling 60s, got %+v", report)
	}
	if got := *sleeps; len(got) != 2 || got[0] != 20*time.Second || got[1] != 40*time.Second {
		t.Fatalf("expected backoff sleeps [20s 40s], got %v", got)
	}
	out := narration.String()
	for _, want := range []string{
		"claude-code is throttled.",
		"Holding the session; resuming automatically in 20s (hold 1 of 3).",
		"Fallback available: local-slm.",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("narration missing %q; got:\n%s", want, out)
		}
	}

	reason := appendThrottleSurvivalReason("Automatic route.", report)
	for _, want := range []string{
		"Route was throttled",
		"held the session for 1m0s across 2 automatic holds",
		"resumed on the same route without human intervention",
	} {
		if !strings.Contains(reason, want) {
			t.Fatalf("survival reason missing %q; got %q", want, reason)
		}
	}
}

func TestRunWithThrottleSurvivalHonorsAdvertisedRetryAfter(t *testing.T) {
	_, sleeps := withThrottleTestHarness(t)

	attempts := 0
	_, _, err := runWithThrottleSurvival(context.Background(), "anthropic", nil, throttleSurvivalOptions{approver: autoApprover{}}, func() (string, error) {
		attempts++
		if attempts == 1 {
			return "", &throttledRouteError{label: "anthropic", retryAfter: 33 * time.Second, underlying: errors.New("rate limit")}
		}
		return "ok", nil
	})
	if err != nil {
		t.Fatalf("expected resume, got %v", err)
	}
	if got := *sleeps; len(got) != 1 || got[0] != 33*time.Second {
		t.Fatalf("expected advertised 33s hold, got %v", got)
	}
}

func TestRunWithThrottleSurvivalNeverWaitsOnNonThrottleErrors(t *testing.T) {
	_, sleeps := withThrottleTestHarness(t)

	attempts := 0
	_, _, err := runWithThrottleSurvival(context.Background(), "codex", nil, throttleSurvivalOptions{approver: autoApprover{}}, func() (string, error) {
		attempts++
		return "", errors.New("invalid api key")
	})
	if err == nil || !strings.Contains(err.Error(), "invalid api key") {
		t.Fatalf("expected passthrough error, got %v", err)
	}
	if attempts != 1 || len(*sleeps) != 0 {
		t.Fatalf("expected one attempt and no holds, got attempts=%d sleeps=%v", attempts, *sleeps)
	}
}

func TestRunWithThrottleSurvivalExhaustionNamesRecoveryAndFallback(t *testing.T) {
	_, sleeps := withThrottleTestHarness(t)

	attempts := 0
	_, report, err := runWithThrottleSurvival(context.Background(), "claude-code", func() string { return "local-slm" }, throttleSurvivalOptions{approver: autoApprover{}}, func() (string, error) {
		attempts++
		return "", &throttledRouteError{label: "claude-code", underlying: errors.New("usage limit reached")}
	})
	if err == nil {
		t.Fatalf("expected exhaustion error")
	}
	if attempts != 4 || len(*sleeps) != 3 {
		t.Fatalf("expected 4 attempts and 3 holds, got attempts=%d sleeps=%v", attempts, *sleeps)
	}
	if report.Holds != 3 || report.TotalHeld != 140*time.Second {
		t.Fatalf("expected 3 holds totaling 140s, got %+v", report)
	}
	for _, want := range []string{
		"still throttled after 4 automatic resume attempts",
		"held 2m20s total",
		"`jini continue` resumes it",
		"`jini route set local-slm`",
	} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("exhaustion error missing %q; got %v", want, err)
		}
	}
}

func TestRunWithThrottleSurvivalStopsWhenContextCancelled(t *testing.T) {
	withThrottleTestHarness(t)

	ctx, cancel := context.WithCancel(context.Background())
	attempts := 0
	_, _, err := runWithThrottleSurvival(ctx, "anthropic", nil, throttleSurvivalOptions{approver: autoApprover{}}, func() (string, error) {
		attempts++
		cancel()
		return "", &throttledRouteError{label: "anthropic", underlying: errors.New("rate limit")}
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context cancellation to stop holds, got %v", err)
	}
	if attempts != 1 {
		t.Fatalf("expected a single attempt after cancel, got %d", attempts)
	}
}

func TestClassifyThrottleErrorWrapsOnlyThrottleMessages(t *testing.T) {
	err := classifyThrottleError("anthropic", fmt.Errorf("anthropic API error: status 429, retry after 20 seconds"))
	var throttled *throttledRouteError
	if !errors.As(err, &throttled) {
		t.Fatalf("expected throttledRouteError, got %T: %v", err, err)
	}
	if throttled.retryAfter != 20*time.Second {
		t.Fatalf("expected 20s advertised wait, got %s", throttled.retryAfter)
	}
	plain := classifyThrottleError("anthropic", errors.New("invalid api key"))
	if errors.As(plain, &throttled) && plain != nil && isThrottleSignal(plain.Error()) {
		t.Fatalf("expected plain error to pass through, got %v", plain)
	}
	if classifyThrottleError("anthropic", nil) != nil {
		t.Fatalf("expected nil passthrough")
	}
}

func TestClassifyCLIThrottleOutputDetectsFromRawOutputOnly(t *testing.T) {
	sanitized := errors.New("Claude Code failed: exit status 1 (stdout 0 chars, stderr 42 chars; output omitted)")
	err := classifyCLIThrottleOutput("Claude Code", "", "usage limit reached, try again in 3 minutes", sanitized)
	var throttled *throttledRouteError
	if !errors.As(err, &throttled) {
		t.Fatalf("expected throttle classification from raw stderr, got %v", err)
	}
	if throttled.retryAfter != 3*time.Minute {
		t.Fatalf("expected 3m advertised wait, got %s", throttled.retryAfter)
	}
	if !strings.Contains(err.Error(), "output omitted") || strings.Contains(err.Error(), "usage limit reached") {
		t.Fatalf("throttle error must keep the sanitized message and never leak raw output; got %v", err)
	}
	plain := classifyCLIThrottleOutput("Claude Code", "", "boom", sanitized)
	if errors.As(plain, &throttled) {
		t.Fatalf("expected non-throttle output to pass through sanitized, got %v", plain)
	}
}

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
		ctx := throttleAttemptContext()
		<-ctx.Done()
		return "", ctx.Err()
	})
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected deadline exceeded passthrough (non-throttle, no hold), got %v", err)
	}
}

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

type blockingReader struct{}

func (blockingReader) Read([]byte) (int, error) { select {} }

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
	withPromptIO(t, "")
	throttlePromptInput = blockingReader{}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := cliPromptApprover{}.Approve(ctx, throttleApprovalRequest{Label: "route"})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected ctx cancellation, got %v", err)
	}
}

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
