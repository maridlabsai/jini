package app

import (
	"fmt"
	"net/http"
)

// Typed credential errors — a BYO user whose key is wrong, revoked, or
// rate-limited should get an actionable message that names the credential to
// check, not a bare HTTP status code. This classifies a remote provider's
// non-2xx response into a small set of kinds with a fix-oriented message.
//
// The 429 message deliberately contains "HTTP 429" / "rate-limited" so throttle
// survival (classifyThrottleError → isThrottleSignal) still recognizes the
// rate-limit event; changing that wording would silently break throttle holds.

type providerCredentialErrorKind int

const (
	credentialErrorUnknown providerCredentialErrorKind = iota
	credentialErrorInvalid       // 401 — key wrong/revoked
	credentialErrorForbidden     // 403 — key lacks permission / account disabled
	credentialErrorRateLimited   // 429 — throttled
	credentialErrorModelNotFound // 404 — model/endpoint not found
	credentialErrorServer        // 5xx — provider-side failure
)

type providerCredentialError struct {
	Provider string
	KeyEnv   string
	Status   int
	Kind     providerCredentialErrorKind
	message  string
}

func (e *providerCredentialError) Error() string { return e.message }

// classifyProviderHTTPError returns a typed, actionable error for a non-2xx
// status from a remote provider, or nil for any 2xx status. providerLabel is
// the human name ("Claude API"); keyEnv is the credential the user should check
// ("ANTHROPIC_API_KEY").
func classifyProviderHTTPError(providerLabel, keyEnv string, status int) error {
	if status >= 200 && status <= 299 {
		return nil
	}
	e := &providerCredentialError{Provider: providerLabel, KeyEnv: keyEnv, Status: status}
	switch {
	case status == http.StatusUnauthorized:
		e.Kind = credentialErrorInvalid
		e.message = fmt.Sprintf("%s rejected the credential (HTTP 401). Check %s — the key may be wrong, revoked, or from a different account.", providerLabel, keyEnv)
	case status == http.StatusForbidden:
		e.Kind = credentialErrorForbidden
		e.message = fmt.Sprintf("%s denied access (HTTP 403). The credential in %s lacks permission for this model, or the account is disabled.", providerLabel, keyEnv)
	case status == http.StatusTooManyRequests:
		e.Kind = credentialErrorRateLimited
		e.message = fmt.Sprintf("%s rate-limited this request (HTTP 429). Wait and retry, or switch routes with `jini route`.", providerLabel)
	case status == http.StatusNotFound:
		e.Kind = credentialErrorModelNotFound
		e.message = fmt.Sprintf("%s could not find the requested model or endpoint (HTTP 404). Check the configured model id and endpoint.", providerLabel)
	case status >= 500:
		e.Kind = credentialErrorServer
		e.message = fmt.Sprintf("%s had a server error (HTTP %d). Retry shortly.", providerLabel, status)
	default:
		e.Kind = credentialErrorUnknown
		e.message = fmt.Sprintf("%s request failed with HTTP %d. Run `jini doctor`.", providerLabel, status)
	}
	return e
}
