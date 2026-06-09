package app_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/maridlabsai/jini/internal/app"
)

func TestMacOSAppSidecarHandshakeReturnsTaskPromptSnapshot(t *testing.T) {
	responses := runMacOSAppSidecar(t, `{"protocol_version":"macos-app-v1","id":"req_handshake","idempotency_key":"idem_handshake","method":"app.handshake","surface":"macos","params":{}}`)

	response := responses[0]
	requireSidecarOK(t, response)
	if got := stringField(t, response, "protocol_version"); got != "macos-app-v1" {
		t.Fatalf("expected protocol version macos-app-v1, got %q", got)
	}

	result := objectField(t, response, "result")
	if got := stringField(t, result, "safe_start_mode"); got != "task_prompt" {
		t.Fatalf("expected task prompt safe start, got %q", got)
	}
	if got := stringField(t, result, "selected_session_id"); got != "" {
		t.Fatalf("expected no auto-opened session, got %q", got)
	}
	if got := stringField(t, result, "protocol_version"); got != "macos-app-v1" {
		t.Fatalf("expected snapshot protocol version, got %q", got)
	}

	events := arrayField(t, response, "events")
	if len(events) != 1 {
		t.Fatalf("expected handshake to emit one readiness event, got %#v", events)
	}
	event := objectValue(t, events[0])
	if got := stringField(t, event, "event_type"); got != "app.ready" {
		t.Fatalf("expected app.ready event, got %q", got)
	}
	if got := numberField(t, event, "sequence"); got != 1 {
		t.Fatalf("expected first event sequence 1, got %v", got)
	}
}

func TestMacOSAppSidecarSimpleQuestionReturnsTransientResponse(t *testing.T) {
	responses := runMacOSAppSidecar(t, `{"protocol_version":"macos-app-v1","id":"req_turn","idempotency_key":"idem_turn","method":"turn.submit","surface":"macos","params":{"text":"what is the capital of france?"}}`)

	response := responses[0]
	requireSidecarOK(t, response)
	result := objectField(t, response, "result")
	if got := stringField(t, result, "kind"); got != "compact_answer" {
		t.Fatalf("expected compact answer, got %q", got)
	}
	if got := stringField(t, result, "assistant_text"); got != "Paris." {
		t.Fatalf("expected Paris answer, got %q", got)
	}
	if got := boolField(t, result, "creates_session"); got {
		t.Fatalf("expected transient answer not to create a session")
	}
	if got := boolField(t, result, "route_visible"); got {
		t.Fatalf("expected simple answer not to show route chrome")
	}

	encoded, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("marshal response: %v", err)
	}
	for _, unwanted := range []string{"session_id", "Task Snapshot", "Saved:", "Result ready"} {
		if strings.Contains(string(encoded), unwanted) {
			t.Fatalf("expected transient answer to avoid %q, got %s", unwanted, string(encoded))
		}
	}
}

func TestMacOSAppSidecarTurnSubmitRequiresIdempotencyKey(t *testing.T) {
	responses := runMacOSAppSidecar(t, `{"protocol_version":"macos-app-v1","id":"req_missing_idem","method":"turn.submit","surface":"macos","params":{"text":"what is the capital of france?"}}`)

	response := responses[0]
	if ok := boolField(t, response, "ok"); ok {
		t.Fatalf("expected turn.submit without idempotency key to fail, got %#v", response)
	}
	err := objectField(t, response, "error")
	if got := stringField(t, err, "code"); got != "invalid_request" {
		t.Fatalf("expected invalid_request, got %q", got)
	}
	if !strings.Contains(stringField(t, err, "message"), "idempotency_key") {
		t.Fatalf("expected idempotency recovery message, got %#v", err)
	}
}

func TestMacOSAppSidecarSubscribeAcceptsReplayCursor(t *testing.T) {
	responses := runMacOSAppSidecar(t, `{"protocol_version":"macos-app-v1","id":"req_subscribe","idempotency_key":"idem_subscribe","method":"app.subscribe","surface":"macos","params":{"last_event_id":"evt_000001","last_sequence":1}}`)

	response := responses[0]
	requireSidecarOK(t, response)
	result := objectField(t, response, "result")
	if got := boolField(t, result, "subscribed"); !got {
		t.Fatalf("expected subscribed result, got %#v", result)
	}
	if got := boolField(t, result, "projection_resync_required"); got {
		t.Fatalf("expected current empty event log not to require projection resync")
	}
	if got := arrayField(t, result, "replayed_events"); len(got) != 0 {
		t.Fatalf("expected no replayed events for empty sidecar, got %#v", got)
	}
}

func TestMacOSAppSidecarSubscribeReplaysEventsAfterCursor(t *testing.T) {
	responses := runMacOSAppSidecar(t,
		`{"protocol_version":"macos-app-v1","id":"req_handshake","idempotency_key":"idem_handshake","method":"app.handshake","surface":"macos","params":{}}`,
		`{"protocol_version":"macos-app-v1","id":"req_subscribe","idempotency_key":"idem_subscribe","method":"app.subscribe","surface":"macos","params":{"last_sequence":0}}`,
	)

	requireSidecarOK(t, responses[0])
	requireSidecarOK(t, responses[1])
	result := objectField(t, responses[1], "result")
	replayed := arrayField(t, result, "replayed_events")
	if len(replayed) != 1 {
		t.Fatalf("expected one replayed app.ready event, got %#v", replayed)
	}
	event := objectValue(t, replayed[0])
	if got := stringField(t, event, "event_type"); got != "app.ready" {
		t.Fatalf("expected replayed app.ready event, got %q", got)
	}
	if got := numberField(t, event, "sequence"); got != 1 {
		t.Fatalf("expected replayed event sequence 1, got %v", got)
	}
}

func TestMacOSAppSidecarRejectsUnknownEnvelopeFields(t *testing.T) {
	responses := runMacOSAppSidecar(t, `{"protocol_version":"macos-app-v1","id":"req_bad","idempotency_key":"idem_bad","method":"app.handshake","surface":"macos","unexpected":true,"params":{}}`)

	response := responses[0]
	if ok := boolField(t, response, "ok"); ok {
		t.Fatalf("expected unknown envelope field to fail, got %#v", response)
	}
	err := objectField(t, response, "error")
	if got := stringField(t, err, "code"); got != "invalid_request" {
		t.Fatalf("expected invalid_request, got %q", got)
	}
}

func TestMacOSAppSidecarRejectsIdempotencyKeyReuseForDifferentRequest(t *testing.T) {
	responses := runMacOSAppSidecar(t,
		`{"protocol_version":"macos-app-v1","id":"req_turn_1","idempotency_key":"idem_turn","method":"turn.submit","surface":"macos","params":{"text":"what is the capital of france?"}}`,
		`{"protocol_version":"macos-app-v1","id":"req_turn_2","idempotency_key":"idem_turn","method":"turn.submit","surface":"macos","params":{"text":"what is the capital of germany?"}}`,
	)

	requireSidecarOK(t, responses[0])
	if ok := boolField(t, responses[1], "ok"); ok {
		t.Fatalf("expected changed request with reused idempotency key to fail, got %#v", responses[1])
	}
	err := objectField(t, responses[1], "error")
	if got := stringField(t, err, "code"); got != "invalid_request" {
		t.Fatalf("expected invalid_request, got %q", got)
	}
	if !strings.Contains(stringField(t, err, "message"), "different request") {
		t.Fatalf("expected digest mismatch message, got %#v", err)
	}
}

func TestMacOSAppSidecarRouteStatusPreservesCLIHandoffKind(t *testing.T) {
	t.Setenv("JINI_TOOL", "codex")
	responses := runMacOSAppSidecar(t, `{"protocol_version":"macos-app-v1","id":"req_route","idempotency_key":"idem_route","method":"route.status","surface":"macos","params":{}}`)

	response := responses[0]
	requireSidecarOK(t, response)
	result := objectField(t, response, "result")
	if got := stringField(t, result, "route_id"); got != "codex" {
		t.Fatalf("expected codex route id, got %q", got)
	}
	if got := stringField(t, result, "route_kind"); got != "cli_handoff" {
		t.Fatalf("expected codex route to stay cli_handoff, got %q", got)
	}
	if got := stringField(t, result, "route_kind"); got == "provider_api" {
		t.Fatalf("expected route status not to alias Codex CLI to provider API")
	}
}

func TestMacOSAppSidecarRouteHelpReturnsSetupLines(t *testing.T) {
	responses := runMacOSAppSidecar(t, `{"protocol_version":"macos-app-v1","id":"req_route_help","idempotency_key":"idem_route_help","method":"route.help","surface":"macos","params":{}}`)

	response := responses[0]
	requireSidecarOK(t, response)
	result := objectField(t, response, "result")
	if got := stringField(t, result, "title"); got != "Route setup" {
		t.Fatalf("expected Route setup title, got %q", got)
	}
	lines := arrayField(t, result, "lines")
	encoded, err := json.Marshal(lines)
	if err != nil {
		t.Fatalf("marshal route help lines: %v", err)
	}
	for _, want := range []string{"jini route list", "jini doctor", "local-preview"} {
		if !strings.Contains(string(encoded), want) {
			t.Fatalf("expected route help to contain %q, got %s", want, string(encoded))
		}
	}
}

func TestMacOSAppSidecarDiagnosticsPreviewIsRedactedAndDoesNotWriteBundle(t *testing.T) {
	stateDir := t.TempDir()
	t.Setenv("ANTHROPIC_API_KEY", "sk-test-secret")

	responses := runMacOSAppSidecarWithState(t, stateDir, `{"protocol_version":"macos-app-v1","id":"req_diagnostics_preview","method":"diagnostics.preview","surface":"macos","params":{}}`)

	response := responses[0]
	requireSidecarOK(t, response)
	result := objectField(t, response, "result")
	if got := stringField(t, result, "kind"); got != "diagnostics_preview" {
		t.Fatalf("expected diagnostics preview, got %q", got)
	}
	if got := boolField(t, result, "redacted"); !got {
		t.Fatalf("expected diagnostics preview to be redacted")
	}
	if got := boolField(t, result, "writes_bundle"); got {
		t.Fatalf("expected preview not to write a diagnostics bundle")
	}
	encoded, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("marshal diagnostics preview: %v", err)
	}
	for _, forbidden := range []string{"sk-test-secret", "ANTHROPIC_API_KEY", stateDir} {
		if strings.Contains(string(encoded), forbidden) {
			t.Fatalf("expected diagnostics preview to redact %q, got %s", forbidden, string(encoded))
		}
	}
	if _, err := os.Stat(filepath.Join(stateDir, "diagnostics")); !os.IsNotExist(err) {
		t.Fatalf("expected diagnostics preview not to create bundle dir, stat error: %v", err)
	}
}

func TestMacOSAppSidecarDiagnosticsExportWritesRedactedBundle(t *testing.T) {
	stateDir := t.TempDir()
	t.Setenv("OPENAI_API_KEY", "sk-test-secret")

	responses := runMacOSAppSidecarWithState(t, stateDir, `{"protocol_version":"macos-app-v1","id":"req_diagnostics_export","idempotency_key":"idem_diagnostics_export","method":"diagnostics.export","surface":"macos","params":{}}`)

	response := responses[0]
	requireSidecarOK(t, response)
	result := objectField(t, response, "result")
	if got := stringField(t, result, "kind"); got != "diagnostics_export" {
		t.Fatalf("expected diagnostics export, got %q", got)
	}
	if got := boolField(t, result, "redacted"); !got {
		t.Fatalf("expected diagnostics export to be redacted")
	}
	bundlePath := stringField(t, result, "bundle_path")
	if !strings.HasPrefix(bundlePath, filepath.Join(stateDir, "diagnostics")+string(os.PathSeparator)) {
		t.Fatalf("expected bundle under diagnostics state dir, got %q", bundlePath)
	}
	if got := stringField(t, result, "bundle_path_redacted"); strings.Contains(got, stateDir) {
		t.Fatalf("expected redacted bundle path not to expose full state dir, got %q", got)
	}

	data, err := os.ReadFile(bundlePath)
	if err != nil {
		t.Fatalf("read diagnostics bundle: %v", err)
	}
	var bundle map[string]any
	if err := json.Unmarshal(data, &bundle); err != nil {
		t.Fatalf("decode diagnostics bundle: %v", err)
	}
	if got := stringField(t, bundle, "schema_version"); got != "jini-macos-diagnostics-v1" {
		t.Fatalf("expected diagnostics schema, got %q", got)
	}
	encoded := string(data)
	for _, forbidden := range []string{"sk-test-secret", "OPENAI_API_KEY", stateDir, "raw_prompt", "artifact_body"} {
		if strings.Contains(encoded, forbidden) {
			t.Fatalf("expected diagnostics bundle to redact %q, got %s", forbidden, encoded)
		}
	}
}

func TestMacOSAppSidecarDiagnosticsRejectsUnknownParams(t *testing.T) {
	responses := runMacOSAppSidecar(t, `{"protocol_version":"macos-app-v1","id":"req_bad_diagnostics","idempotency_key":"idem_bad_diagnostics","method":"diagnostics.export","surface":"macos","params":{"include_raw_prompts":true}}`)

	response := responses[0]
	if ok := boolField(t, response, "ok"); ok {
		t.Fatalf("expected diagnostics export with unknown params to fail, got %#v", response)
	}
	err := objectField(t, response, "error")
	if got := stringField(t, err, "code"); got != "invalid_request" {
		t.Fatalf("expected invalid_request, got %q", got)
	}
	if !strings.Contains(stringField(t, err, "message"), "params are invalid") {
		t.Fatalf("expected invalid params message, got %#v", err)
	}
}

func TestMacOSAppSidecarCommandIsNotPublicInventory(t *testing.T) {
	for _, args := range [][]string{
		{"commands"},
		{"help", "all"},
		{"help", "admin"},
	} {
		var stdout bytes.Buffer
		exitCode := app.Run(args, &stdout, &stdout)
		if exitCode != 0 {
			t.Fatalf("expected %v to succeed, got %d with output:\n%s", args, exitCode, stdout.String())
		}
		for _, unwanted := range []string{"app serve", "macos-app-v1", "--surface macos", "sidecar"} {
			if strings.Contains(stdout.String(), unwanted) {
				t.Fatalf("expected command inventory %v not to expose app sidecar %q, got:\n%s", args, unwanted, stdout.String())
			}
		}
	}
}

func runMacOSAppSidecar(t *testing.T, requestLines ...string) []map[string]any {
	t.Helper()
	return runMacOSAppSidecarWithState(t, t.TempDir(), requestLines...)
}

func runMacOSAppSidecarWithState(t *testing.T, stateDir string, requestLines ...string) []map[string]any {
	t.Helper()
	t.Setenv("JINI_STATE_DIR", stateDir)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	stdin := strings.NewReader(strings.Join(requestLines, "\n") + "\n")
	exitCode := app.RunInteractive([]string{"app", "serve", "--stdio", "--surface", "macos"}, stdin, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("expected sidecar exit 0, got %d\nstdout:\n%s\nstderr:\n%s", exitCode, stdout.String(), stderr.String())
	}

	lines := strings.Split(strings.TrimSpace(stdout.String()), "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) == "" {
		t.Fatalf("expected sidecar response, got empty output")
	}
	responses := make([]map[string]any, 0, len(lines))
	for _, line := range lines {
		var decoded map[string]any
		if err := json.Unmarshal([]byte(line), &decoded); err != nil {
			t.Fatalf("decode sidecar response %q: %v", line, err)
		}
		responses = append(responses, decoded)
	}
	return responses
}

func requireSidecarOK(t *testing.T, response map[string]any) {
	t.Helper()
	if ok := boolField(t, response, "ok"); !ok {
		t.Fatalf("expected ok sidecar response, got %#v", response)
	}
}

func objectField(t *testing.T, values map[string]any, key string) map[string]any {
	t.Helper()
	value, ok := values[key]
	if !ok {
		t.Fatalf("expected object field %q in %#v", key, values)
	}
	return objectValue(t, value)
}

func objectValue(t *testing.T, value any) map[string]any {
	t.Helper()
	object, ok := value.(map[string]any)
	if !ok {
		t.Fatalf("expected object value, got %#v", value)
	}
	return object
}

func arrayField(t *testing.T, values map[string]any, key string) []any {
	t.Helper()
	value, ok := values[key]
	if !ok {
		t.Fatalf("expected array field %q in %#v", key, values)
	}
	array, ok := value.([]any)
	if !ok {
		t.Fatalf("expected array field %q, got %#v", key, value)
	}
	return array
}

func stringField(t *testing.T, values map[string]any, key string) string {
	t.Helper()
	value, ok := values[key]
	if !ok {
		t.Fatalf("expected string field %q in %#v", key, values)
	}
	text, ok := value.(string)
	if !ok {
		t.Fatalf("expected string field %q, got %#v", key, value)
	}
	return text
}

func boolField(t *testing.T, values map[string]any, key string) bool {
	t.Helper()
	value, ok := values[key]
	if !ok {
		t.Fatalf("expected bool field %q in %#v", key, values)
	}
	result, ok := value.(bool)
	if !ok {
		t.Fatalf("expected bool field %q, got %#v", key, value)
	}
	return result
}

func numberField(t *testing.T, values map[string]any, key string) float64 {
	t.Helper()
	value, ok := values[key]
	if !ok {
		t.Fatalf("expected number field %q in %#v", key, values)
	}
	number, ok := value.(float64)
	if !ok {
		t.Fatalf("expected number field %q, got %#v", key, value)
	}
	return number
}
