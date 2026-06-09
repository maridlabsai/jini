package app

import "encoding/json"

const macOSAppProtocolVersion = "macos-app-v1"

type appRPCRequest struct {
	ProtocolVersion string          `json:"protocol_version"`
	ID              string          `json:"id"`
	IdempotencyKey  string          `json:"idempotency_key,omitempty"`
	Method          string          `json:"method"`
	Params          json.RawMessage `json:"params,omitempty"`
	Surface         string          `json:"surface"`
	SentAt          string          `json:"sent_at,omitempty"`
}

type appRPCResponse struct {
	ProtocolVersion string        `json:"protocol_version"`
	ID              string        `json:"id"`
	OK              bool          `json:"ok"`
	Result          any           `json:"result,omitempty"`
	Events          []appRPCEvent `json:"events"`
	Error           *appRPCError  `json:"error,omitempty"`
}

type appRPCError struct {
	Code        string `json:"code"`
	Message     string `json:"message"`
	Recovery    string `json:"recovery"`
	SafeToRetry bool   `json:"safe_to_retry"`
}

type appRPCEvent struct {
	ProtocolVersion string `json:"protocol_version"`
	EventID         string `json:"event_id"`
	Sequence        int64  `json:"sequence"`
	EventType       string `json:"event_type"`
	Surface         string `json:"surface"`
	SessionID       string `json:"session_id,omitempty"`
	Payload         any    `json:"payload"`
	EmittedAt       string `json:"emitted_at"`
}

type appSnapshotVM struct {
	AppVersion        string          `json:"app_version"`
	CoreVersion       string          `json:"core_version"`
	ProtocolVersion   string          `json:"protocol_version"`
	SelectedProjectID string          `json:"selected_project_id"`
	SelectedSessionID string          `json:"selected_session_id"`
	OnlineState       string          `json:"online_state"`
	OfflineState      string          `json:"offline_state"`
	RouteSummary      appRouteSummary `json:"route_summary"`
	SetupWarnings     []string        `json:"setup_warnings"`
	RecentProjects    []projectVM     `json:"recent_projects"`
	SafeStartMode     string          `json:"safe_start_mode"`
}

type projectVM struct {
	ProjectID        string `json:"project_id"`
	DisplayName      string `json:"display_name"`
	RootLabel        string `json:"root_label"`
	RootPathRedacted string `json:"root_path_redacted"`
	GitBranch        string `json:"git_branch"`
	GitDirty         bool   `json:"git_dirty"`
	JiniStateState   string `json:"jini_state_state"`
	RouteHealth      string `json:"route_health"`
	LastOpenedAt     string `json:"last_opened_at"`
}

type appRouteSummary struct {
	RouteID       string   `json:"route_id"`
	RouteKind     string   `json:"route_kind"`
	Label         string   `json:"label"`
	Status        string   `json:"status"`
	Selected      bool     `json:"selected"`
	Reason        string   `json:"reason"`
	TokenPosture  string   `json:"token_posture"`
	PowerPosture  string   `json:"power_posture"`
	OfflineState  string   `json:"offline_state"`
	SetupGuidance []string `json:"setup_guidance"`
}

type routeHelpVM struct {
	Title string   `json:"title"`
	Lines []string `json:"lines"`
}

type transientResponseVM struct {
	Kind           string `json:"kind"`
	RequestID      string `json:"request_id"`
	AssistantText  string `json:"assistant_text"`
	CreatedAt      string `json:"created_at"`
	CreatesSession bool   `json:"creates_session"`
	RouteVisible   bool   `json:"route_visible"`
}

type appSubscribeParams struct {
	LastEventID  string `json:"last_event_id,omitempty"`
	LastSequence int64  `json:"last_sequence,omitempty"`
}

type appSubscribeResult struct {
	Subscribed               bool          `json:"subscribed"`
	LastEventID              string        `json:"last_event_id,omitempty"`
	LastSequence             int64         `json:"last_sequence"`
	ProjectionResyncRequired bool          `json:"projection_resync_required"`
	ReplayedEvents           []appRPCEvent `json:"replayed_events"`
}

type turnSubmitParams struct {
	Text string `json:"text"`
}
