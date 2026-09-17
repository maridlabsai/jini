package app

import (
	"errors"
	"strings"
	"testing"
)

func TestClassifyProviderHTTPError(t *testing.T) {
	cases := []struct {
		status   int
		wantKind providerCredentialErrorKind
		wantIn   string // substring the message must contain
	}{
		{401, credentialErrorInvalid, "ANTHROPIC_API_KEY"},
		{403, credentialErrorForbidden, "permission"},
		{429, credentialErrorRateLimited, "HTTP 429"},
		{404, credentialErrorModelNotFound, "model"},
		{500, credentialErrorServer, "server error"},
		{503, credentialErrorServer, "server error"},
		{418, credentialErrorUnknown, "HTTP 418"},
	}
	for _, tc := range cases {
		err := classifyProviderHTTPError("Claude API", "ANTHROPIC_API_KEY", tc.status)
		if err == nil {
			t.Fatalf("status %d: expected an error", tc.status)
		}
		var ce *providerCredentialError
		if !errors.As(err, &ce) {
			t.Fatalf("status %d: expected *providerCredentialError, got %T", tc.status, err)
		}
		if ce.Kind != tc.wantKind {
			t.Fatalf("status %d: kind = %v, want %v", tc.status, ce.Kind, tc.wantKind)
		}
		if !strings.Contains(err.Error(), tc.wantIn) {
			t.Fatalf("status %d: message %q missing %q", tc.status, err.Error(), tc.wantIn)
		}
	}
}

// A 2xx status must not produce an error.
func TestClassifyProviderHTTPErrorSuccess(t *testing.T) {
	for _, status := range []int{200, 201, 204, 299} {
		if err := classifyProviderHTTPError("Claude API", "ANTHROPIC_API_KEY", status); err != nil {
			t.Fatalf("status %d: expected nil, got %v", status, err)
		}
	}
}

// The 429 credential error must still be recognized as a throttle signal so
// throttle survival keeps holding/switching rather than surfacing it raw.
func TestRateLimitCredentialErrorStaysThrottleSignal(t *testing.T) {
	err := classifyProviderHTTPError("Claude API", "ANTHROPIC_API_KEY", 429)
	wrapped := classifyThrottleError("Claude API", err)
	var throttled *throttledRouteError
	if !errors.As(wrapped, &throttled) {
		t.Fatalf("429 credential error must classify as a throttle event, got %T: %v", wrapped, wrapped)
	}
}
