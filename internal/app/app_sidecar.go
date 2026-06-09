package app

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type macOSAppSidecar struct {
	completedByIdempotency map[string]appIdempotencyRecord
	events                 []appRPCEvent
	nextSequence           int64
	now                    func() time.Time
}

type appIdempotencyRecord struct {
	RequestDigest string
	Response      appRPCResponse
}

func isHiddenAppSidecarServeCommand(args []string) bool {
	if len(args) != 5 {
		return false
	}
	return exactCommandToken(args[0]) == "app" &&
		exactCommandToken(args[1]) == "serve" &&
		strings.TrimSpace(args[2]) == "--stdio" &&
		strings.TrimSpace(args[3]) == "--surface" &&
		exactCommandToken(args[4]) == "macos"
}

func runAppSidecarServe(stdin io.Reader, stdout, stderr io.Writer) int {
	if stdin == nil {
		fmt.Fprintln(stderr, "app sidecar requires stdin")
		return 1
	}
	sidecar := newMacOSAppSidecar()
	return sidecar.run(stdin, stdout, stderr)
}

func newMacOSAppSidecar() *macOSAppSidecar {
	return &macOSAppSidecar{
		completedByIdempotency: map[string]appIdempotencyRecord{},
		now:                    time.Now,
	}
}

func (sidecar *macOSAppSidecar) run(stdin io.Reader, stdout, stderr io.Writer) int {
	scanner := bufio.NewScanner(stdin)
	scanner.Buffer(make([]byte, 1024), 1024*1024)
	encoder := json.NewEncoder(stdout)
	for scanner.Scan() {
		line := bytes.TrimSpace(scanner.Bytes())
		if len(line) == 0 {
			continue
		}
		response := sidecar.handle(line)
		if err := encoder.Encode(response); err != nil {
			fmt.Fprintf(stderr, "app sidecar write failed: %v\n", err)
			return 1
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintf(stderr, "app sidecar read failed: %v\n", err)
		return 1
	}
	return 0
}

func (sidecar *macOSAppSidecar) handle(line []byte) appRPCResponse {
	var request appRPCRequest
	decoder := json.NewDecoder(bytes.NewReader(line))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		return appErrorResponse("", "invalid_request", "Request must be a JSON object.", "Send one valid JSON request per line.", true)
	}
	if response, ok := validateAppRPCRequest(request); !ok {
		return response
	}
	if isMutatingAppMethod(request.Method) {
		digest := appRequestDigest(request)
		if completed, ok := sidecar.completedByIdempotency[request.IdempotencyKey]; ok {
			if completed.RequestDigest != digest {
				return appErrorResponse(request.ID, "invalid_request", "idempotency_key was reused for a different request.", "Use a fresh idempotency_key for a new user action.", false)
			}
			response := completed.Response
			response.ID = request.ID
			return response
		}
	}

	response := sidecar.dispatch(request)
	if response.OK && isMutatingAppMethod(request.Method) {
		sidecar.completedByIdempotency[request.IdempotencyKey] = appIdempotencyRecord{
			RequestDigest: appRequestDigest(request),
			Response:      response,
		}
	}
	return response
}

func validateAppRPCRequest(request appRPCRequest) (appRPCResponse, bool) {
	if strings.TrimSpace(request.ID) == "" {
		return appErrorResponse("", "invalid_request", "Request id is required.", "Send every request with an id.", true), false
	}
	if request.ProtocolVersion != macOSAppProtocolVersion {
		return appErrorResponse(request.ID, "protocol_mismatch", "Unsupported app protocol version.", "Update the macOS app and Jini core together.", false), false
	}
	if request.Surface != "macos" {
		return appErrorResponse(request.ID, "invalid_request", "surface must be macos.", "Start the sidecar with --surface macos.", true), false
	}
	if strings.TrimSpace(request.Method) == "" {
		return appErrorResponse(request.ID, "invalid_request", "method is required.", "Send an allowlisted app method.", true), false
	}
	if isMutatingAppMethod(request.Method) && strings.TrimSpace(request.IdempotencyKey) == "" {
		return appErrorResponse(request.ID, "invalid_request", "idempotency_key is required for mutating requests.", "Retry with a stable idempotency_key for this user action.", true), false
	}
	return appRPCResponse{}, true
}

func isMutatingAppMethod(method string) bool {
	switch method {
	case "turn.submit", "project.open", "project.close", "session.resume", "turn.cancel", "turn.retry", "approval.resolve", "diagnostics.export":
		return true
	default:
		return false
	}
}

func (sidecar *macOSAppSidecar) dispatch(request appRPCRequest) appRPCResponse {
	switch request.Method {
	case "app.handshake":
		return sidecar.ok(request, sidecar.snapshot(), []appRPCEvent{
			sidecar.nextEvent("app.ready", "", map[string]any{
				"safe_start_mode": "task_prompt",
			}),
		})
	case "app.snapshot":
		return sidecar.ok(request, sidecar.snapshot(), nil)
	case "app.subscribe":
		params, err := decodeAppParams[appSubscribeParams](request.Params)
		if err != nil {
			return appErrorResponse(request.ID, "invalid_request", "app.subscribe params are invalid.", "Send last_event_id and last_sequence as optional cursor fields.", true)
		}
		return sidecar.ok(request, sidecar.subscribe(params), nil)
	case "project.listRecent":
		return sidecar.ok(request, []projectVM{}, nil)
	case "route.status":
		return sidecar.ok(request, routeSummaryFromDecision(detectRoute()), nil)
	case "route.help":
		return sidecar.ok(request, appRouteHelp(), nil)
	case "diagnostics.preview":
		if _, err := decodeAppParams[struct{}](request.Params); err != nil {
			return appErrorResponse(request.ID, "invalid_request", "diagnostics.preview params are invalid.", "Send an empty params object.", true)
		}
		return sidecar.ok(request, sidecar.diagnosticsPreview(), nil)
	case "diagnostics.export":
		if _, err := decodeAppParams[struct{}](request.Params); err != nil {
			return appErrorResponse(request.ID, "invalid_request", "diagnostics.export params are invalid.", "Send an empty params object.", true)
		}
		return sidecar.exportDiagnostics(request)
	case "turn.submit":
		params, err := decodeAppParams[turnSubmitParams](request.Params)
		if err != nil || strings.TrimSpace(params.Text) == "" {
			return appErrorResponse(request.ID, "invalid_request", "turn.submit requires text.", "Send params.text with the user request.", true)
		}
		return sidecar.submitTurn(request, params)
	default:
		return appErrorResponse(request.ID, "invalid_request", "Unsupported app method.", "Use an allowlisted macOS app method.", false)
	}
}

func (sidecar *macOSAppSidecar) ok(request appRPCRequest, result any, events []appRPCEvent) appRPCResponse {
	if events == nil {
		events = []appRPCEvent{}
	}
	return appRPCResponse{
		ProtocolVersion: macOSAppProtocolVersion,
		ID:              request.ID,
		OK:              true,
		Result:          result,
		Events:          events,
	}
}

func (sidecar *macOSAppSidecar) snapshot() appSnapshotVM {
	version := currentJiniVersion()
	return appSnapshotVM{
		AppVersion:        version,
		CoreVersion:       version,
		ProtocolVersion:   macOSAppProtocolVersion,
		SelectedProjectID: "",
		SelectedSessionID: "",
		OnlineState:       "online",
		OfflineState:      "available",
		RouteSummary:      routeSummaryFromDecision(detectRoute()),
		SetupWarnings:     []string{},
		RecentProjects:    []projectVM{},
		SafeStartMode:     "task_prompt",
	}
}

func routeSummaryFromDecision(decision routeDecision) appRouteSummary {
	routeID := firstNonEmpty(decision.ToolMode, "auto")
	providerStatus := strings.TrimSpace(decision.Provider.Status)
	guidance := []string{}
	if len(decision.Provider.Missing) > 0 {
		guidance = append(guidance, decision.Provider.Missing...)
	}
	return appRouteSummary{
		RouteID:       routeID,
		RouteKind:     appRouteKind(routeID, decision.Provider),
		Label:         firstNonEmpty(decision.ToolLabel, titleCase(routeID)),
		Status:        routeAppStatus(providerStatus),
		Selected:      true,
		Reason:        strings.TrimSpace(decision.Reason),
		TokenPosture:  "frugal",
		PowerPosture:  "normal",
		OfflineState:  "available",
		SetupGuidance: guidance,
	}
}

func routeAppStatus(providerStatus string) string {
	switch strings.TrimSpace(providerStatus) {
	case "", "ok":
		return "available"
	case "missing":
		return "needs_setup"
	default:
		return normalizeName(providerStatus)
	}
}

func appRouteKind(routeID string, provider providerConfig) string {
	if cliHandoffMode(routeID) || provider.ID == "cli-handoff" {
		return "cli_handoff"
	}
	if strings.Contains(routeID, "local") || strings.Contains(provider.ID, "local") {
		return "local_model"
	}
	if routeID == "auto" {
		return "gateway"
	}
	if provider.ID == "claude" || provider.ID == "azure-openai" || provider.ID == "bedrock" || strings.Contains(provider.ID, "api") {
		return "provider_api"
	}
	return "offline"
}

func appRouteHelp() routeHelpVM {
	var buffer bytes.Buffer
	renderRouteSetupHelp(&buffer)
	return routeHelpVM{
		Title: "Route setup",
		Lines: strings.Split(strings.TrimRight(buffer.String(), "\n"), "\n"),
	}
}

func (sidecar *macOSAppSidecar) diagnosticsPreview() diagnosticsPreviewVM {
	return diagnosticsPreviewVM{
		Kind:         "diagnostics_preview",
		CreatedAt:    sidecar.now().UTC().Format(time.RFC3339),
		Redacted:     true,
		WritesBundle: false,
		Included:     diagnosticsIncludedCategories(),
		Excluded:     diagnosticsExcludedCategories(),
	}
}

func (sidecar *macOSAppSidecar) exportDiagnostics(request appRPCRequest) appRPCResponse {
	createdAt := sidecar.now().UTC()
	stateRoot := sessionStateRoot()
	diagnosticsDir := filepath.Join(stateRoot, "diagnostics")
	if err := os.MkdirAll(diagnosticsDir, 0o755); err != nil {
		return appErrorResponse(request.ID, "io_error", "Could not create diagnostics directory.", "Check local file permissions and try diagnostics export again.", true)
	}

	bundleName := "jini-diagnostics-" + createdAt.Format("20060102T150405Z") + ".json"
	bundlePath := filepath.Join(diagnosticsDir, bundleName)
	bundle := sidecar.diagnosticsBundle(createdAt)
	data, err := json.MarshalIndent(bundle, "", "  ")
	if err != nil {
		return appErrorResponse(request.ID, "internal_error", "Could not render diagnostics bundle.", "Try again after restarting Jini.", true)
	}
	data = append(data, '\n')
	if err := os.WriteFile(bundlePath, data, 0o600); err != nil {
		return appErrorResponse(request.ID, "io_error", "Could not write diagnostics bundle.", "Check local file permissions and try diagnostics export again.", true)
	}

	return sidecar.ok(request, diagnosticsExportVM{
		Kind:               "diagnostics_export",
		BundleName:         bundleName,
		BundlePath:         bundlePath,
		BundlePathRedacted: redactDiagnosticsBundlePath(bundlePath),
		CreatedAt:          createdAt.Format(time.RFC3339),
		Redacted:           true,
		Included:           diagnosticsIncludedCategories(),
		Excluded:           diagnosticsExcludedCategories(),
	}, nil)
}

func (sidecar *macOSAppSidecar) diagnosticsBundle(createdAt time.Time) map[string]any {
	version := currentJiniVersion()
	return map[string]any{
		"schema_version":    "jini-macos-diagnostics-v1",
		"generated_at":      createdAt.UTC().Format(time.RFC3339),
		"app_version":       version,
		"core_version":      version,
		"protocol_version":  macOSAppProtocolVersion,
		"surface":           "macos",
		"online_state":      "online",
		"offline_state":     "available",
		"route_summary":     diagnosticsRouteSummary(),
		"session_ids":       []string{},
		"sync_debt_count":   0,
		"command_classes":   diagnosticsCommandClasses(),
		"included":          diagnosticsIncludedCategories(),
		"excluded":          diagnosticsExcludedCategories(),
		"redaction_summary": diagnosticsRedactionSummary(),
	}
}

func diagnosticsCommandClasses() []string {
	return []string{
		"app.handshake",
		"app.snapshot",
		"app.subscribe",
		"project.listRecent",
		"route.status",
		"route.help",
		"turn.submit",
		"diagnostics.preview",
		"diagnostics.export",
	}
}

func diagnosticsRouteSummary() appRouteSummary {
	summary := routeSummaryFromDecision(detectRoute())
	summary.Reason = redactDiagnosticsText(summary.Reason)
	for index, guidance := range summary.SetupGuidance {
		summary.SetupGuidance[index] = redactDiagnosticsText(guidance)
	}
	return summary
}

func diagnosticsIncludedCategories() []string {
	return []string{
		"app and core versions",
		"protocol version",
		"route summary",
		"offline and sync state",
		"allowed command classes",
		"redaction summary",
	}
}

func diagnosticsExcludedCategories() []string {
	return []string{
		"secrets",
		"provider payloads",
		"prompt text",
		"artifact content",
		"full local paths",
	}
}

func redactDiagnosticsText(text string) string {
	redacted := strings.ReplaceAll(text, sessionStateRoot(), "$JINI_STATE_DIR")
	if home := homeDir(); home != "" && home != "." {
		redacted = strings.ReplaceAll(redacted, home, "~")
	}
	for _, env := range os.Environ() {
		name, value, ok := strings.Cut(env, "=")
		if !ok || len(value) < 6 || !looksLikeSensitiveEnvName(name) {
			continue
		}
		redacted = strings.ReplaceAll(redacted, value, "[redacted]")
	}
	return redacted
}

func looksLikeSensitiveEnvName(name string) bool {
	upper := strings.ToUpper(strings.TrimSpace(name))
	return strings.Contains(upper, "KEY") ||
		strings.Contains(upper, "TOKEN") ||
		strings.Contains(upper, "SECRET") ||
		strings.Contains(upper, "PASSWORD") ||
		strings.Contains(upper, "CREDENTIAL")
}

func diagnosticsRedactionSummary() map[string]string {
	return map[string]string{
		"secrets":           "excluded",
		"provider_payloads": "excluded",
		"prompt_text":       "excluded",
		"artifact_content":  "excluded",
		"full_local_paths":  "excluded",
	}
}

func redactDiagnosticsBundlePath(bundlePath string) string {
	stateRoot := sessionStateRoot()
	if rel, err := filepath.Rel(stateRoot, bundlePath); err == nil && !strings.HasPrefix(rel, "..") {
		return filepath.ToSlash(filepath.Join("$JINI_STATE_DIR", rel))
	}
	return filepath.Join("$JINI_DIAGNOSTICS_DIR", filepath.Base(bundlePath))
}

func (sidecar *macOSAppSidecar) subscribe(params appSubscribeParams) appSubscribeResult {
	replayed := []appRPCEvent{}
	for _, event := range sidecar.events {
		if event.Sequence > params.LastSequence {
			replayed = append(replayed, event)
		}
	}
	return appSubscribeResult{
		Subscribed:               true,
		LastEventID:              params.LastEventID,
		LastSequence:             params.LastSequence,
		ProjectionResyncRequired: false,
		ReplayedEvents:           replayed,
	}
}

func (sidecar *macOSAppSidecar) submitTurn(request appRPCRequest, params turnSubmitParams) appRPCResponse {
	if answer, ok := simpleCapitalAnswer(params.Text); ok {
		return sidecar.ok(request, transientResponseVM{
			Kind:           "compact_answer",
			RequestID:      request.ID,
			AssistantText:  answer,
			CreatedAt:      sidecar.now().UTC().Format(time.RFC3339),
			CreatesSession: false,
			RouteVisible:   false,
		}, nil)
	}
	if looksLikeStandaloneQuestion(params.Text) {
		return sidecar.ok(request, transientResponseVM{
			Kind:           "compact_answer",
			RequestID:      request.ID,
			AssistantText:  "I don't know locally.",
			CreatedAt:      sidecar.now().UTC().Format(time.RFC3339),
			CreatesSession: false,
			RouteVisible:   false,
		}, nil)
	}
	return appErrorResponse(request.ID, "approval_required", "This app action is not implemented in the sidecar yet.", "Use the Jini CLI for this action until the macOS approval and diff flow is available.", false)
}

func (sidecar *macOSAppSidecar) nextEvent(eventType, sessionID string, payload any) appRPCEvent {
	sidecar.nextSequence++
	event := appRPCEvent{
		ProtocolVersion: macOSAppProtocolVersion,
		EventID:         fmt.Sprintf("evt_%06d", sidecar.nextSequence),
		Sequence:        sidecar.nextSequence,
		EventType:       eventType,
		Surface:         "macos",
		SessionID:       sessionID,
		Payload:         payload,
		EmittedAt:       sidecar.now().UTC().Format(time.RFC3339),
	}
	sidecar.events = append(sidecar.events, event)
	return event
}

func appErrorResponse(id, code, message, recovery string, safeToRetry bool) appRPCResponse {
	return appRPCResponse{
		ProtocolVersion: macOSAppProtocolVersion,
		ID:              id,
		OK:              false,
		Events:          []appRPCEvent{},
		Error: &appRPCError{
			Code:        code,
			Message:     message,
			Recovery:    recovery,
			SafeToRetry: safeToRetry,
		},
	}
}

func decodeAppParams[T any](raw json.RawMessage) (T, error) {
	var params T
	if len(raw) == 0 {
		return params, nil
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&params); err != nil {
		return params, err
	}
	return params, nil
}

func appRequestDigest(request appRPCRequest) string {
	hash := sha256.Sum256([]byte(strings.Join([]string{
		request.ProtocolVersion,
		request.Surface,
		request.Method,
		string(bytes.TrimSpace(request.Params)),
	}, "\x00")))
	return fmt.Sprintf("sha256:%x", hash[:])
}
