import { Command } from "@tauri-apps/plugin-shell";

const PROTOCOL_VERSION = "macos-app-v1";
const SURFACE = "macos";
const SIDECAR_ARGS = ["app", "serve", "--stdio", "--surface", "macos"];

export class JiniSidecarClient {
  constructor() {
    this.child = null;
    this.buffer = "";
    this.nextRequestNumber = 1;
    this.pending = new Map();
    this.listeners = new Set();
    this.demoMode = !("__TAURI_INTERNALS__" in window);
  }

  async start() {
    if (this.demoMode) {
      return this.send("app.handshake", {});
    }

    const command = Command.sidecar("binaries/jini-sidecar", SIDECAR_ARGS);
    command.stdout.on("data", (chunk) => this.acceptOutput(String(chunk)));
    command.stderr.on("data", (chunk) => this.emit({ type: "stderr", text: String(chunk) }));
    command.on("close", () => this.emit({ type: "core_closed" }));
    command.on("error", (error) => this.emit({ type: "core_error", text: String(error) }));
    this.child = await command.spawn();

    const snapshot = await this.send("app.handshake", {});
    await this.send("app.subscribe", { last_sequence: 0 });
    return snapshot;
  }

  onEvent(listener) {
    this.listeners.add(listener);
    return () => this.listeners.delete(listener);
  }

  async submitTurn(text) {
    return this.send("turn.submit", { text }, { mutating: true });
  }

  async exportDiagnostics() {
    return this.send("diagnostics.export", {}, { mutating: true });
  }

  async routeStatus() {
    return this.send("route.status", {});
  }

  async routeHelp() {
    return this.send("route.help", {});
  }

  async listRecentProjects() {
    return this.send("project.listRecent", {});
  }

  async send(method, params = {}, options = {}) {
    const requestID = `req_${Date.now()}_${this.nextRequestNumber++}`;
    const request = {
      protocol_version: "macos-app-v1",
      id: requestID,
      method,
      params,
      surface: SURFACE,
      sent_at: new Date().toISOString()
    };
    if (options.mutating) {
      request.idempotency_key = `${method}_${requestID}`;
    }

    if (this.demoMode) {
      return demoResponse(method, params, requestID);
    }

    if (!this.child) {
      throw new Error("Jini core is not running.");
    }

    const result = new Promise((resolve, reject) => {
      this.pending.set(requestID, { resolve, reject });
    });
    await this.child.write(`${JSON.stringify(request)}\n`);
    return result;
  }

  acceptOutput(chunk) {
    this.buffer += chunk;
    const lines = this.buffer.split(/\r?\n/);
    this.buffer = lines.pop() ?? "";
    for (const line of lines) {
      const trimmed = line.trim();
      if (trimmed.length > 0) {
        this.acceptLine(trimmed);
      }
    }
  }

  acceptLine(line) {
    let message;
    try {
      message = JSON.parse(line);
    } catch (error) {
      this.emit({ type: "decode_error", text: String(error), line });
      return;
    }

    for (const event of message.events ?? []) {
      this.emit({ type: "event", event });
    }

    const pending = this.pending.get(message.id);
    if (!pending) {
      this.emit({ type: "orphan_response", response: message });
      return;
    }
    this.pending.delete(message.id);
    if (message.ok) {
      pending.resolve(message.result);
    } else {
      pending.reject(Object.assign(new Error(message.error?.message ?? "Jini request failed."), { response: message }));
    }
  }

  emit(event) {
    for (const listener of this.listeners) {
      listener(event);
    }
  }
}

function demoResponse(method, params, requestID) {
  const createdAt = new Date().toISOString();
  if (method === "turn.submit" && /capital of france/i.test(params.text ?? "")) {
    return {
      kind: "compact_answer",
      request_id: requestID,
      assistant_text: "Paris.",
      created_at: createdAt,
      creates_session: false,
      route_visible: false
    };
  }
  if (method === "route.status") {
    return {
      route_id: "auto",
      route_kind: "gateway",
      label: "Auto",
      status: "available",
      selected: true,
      reason: "Demo renderer fallback",
      token_posture: "frugal",
      power_posture: "normal",
      offline_state: "available",
      setup_guidance: []
    };
  }
  if (method === "route.help") {
    return { title: "Route setup", lines: ["Run jini route help in the CLI for full setup guidance."] };
  }
  if (method === "project.listRecent") {
    return [];
  }
  if (method === "diagnostics.export") {
    return {
      kind: "diagnostics_export",
      bundle_name: "demo.json",
      bundle_path: "",
      bundle_path_redacted: "$JINI_STATE_DIR/diagnostics/demo.json",
      created_at: createdAt,
      redacted: true,
      included: ["app and core versions", "route summary"],
      excluded: ["secrets", "prompt text", "artifact content", "full local paths"]
    };
  }
  return {
    app_version: "0.1.0",
    core_version: "0.1.0",
    protocol_version: PROTOCOL_VERSION,
    selected_project_id: "",
    selected_session_id: "",
    online_state: "online",
    offline_state: "available",
    route_summary: {
      route_id: "auto",
      route_kind: "gateway",
      label: "Auto",
      status: "available",
      selected: true,
      reason: "Demo renderer fallback",
      token_posture: "frugal",
      power_posture: "normal",
      offline_state: "available",
      setup_guidance: []
    },
    setup_warnings: [],
    recent_projects: [],
    safe_start_mode: "task_prompt"
  };
}
