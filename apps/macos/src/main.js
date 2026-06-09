import "./styles.css";
import { JiniSidecarClient } from "./sidecar.js";

const client = new JiniSidecarClient();
const SIDECAR_METHOD_TRACE = ["turn.submit", "diagnostics.export"];

const state = {
  snapshot: null,
  route: null,
  routeHelp: null,
  projects: [],
  messages: [],
  events: [],
  diagnostics: null,
  busy: false,
  error: ""
};

const app = document.querySelector("#app");

function render() {
  const route = state.route ?? state.snapshot?.route_summary;
  app.innerHTML = `
    <div class="shell" data-sidecar-methods="${escapeHTML(SIDECAR_METHOD_TRACE.join(","))}">
      <aside class="sidebar" aria-label="Project and session browser">
        <div class="brand">Jini</div>
        <button class="primary small" data-action="open-project">Open project</button>

        <section>
          <h2>Project</h2>
          ${renderProjects()}
        </section>

        <section>
          <h2>Sessions</h2>
          <div class="empty">No session selected. Ask a task or open a project.</div>
        </section>
      </aside>

      <main class="thread" aria-label="Task composer and compact answer UI">
        <header class="toolbar">
          <div>
            <h1>Describe the task.</h1>
            <p>${escapeHTML(route?.label ?? "Auto")} · ${escapeHTML(route?.token_posture ?? "frugal")} tokens · ${escapeHTML(route?.offline_state ?? "offline available")}</p>
          </div>
          <button class="secondary" data-action="route-help">Route help</button>
        </header>

        <div class="messages" aria-live="polite">
          ${renderMessages()}
        </div>

        ${state.error ? `<div class="error">${escapeHTML(state.error)}</div>` : ""}

        <form class="composer" data-action="submit-task">
          <label class="visually-hidden" for="task-input">Task composer</label>
          <textarea id="task-input" name="task" rows="3" placeholder="Ask a question or describe the file work."></textarea>
          <div class="composer-row">
            <span>${state.busy ? "Working..." : "Go core decides answer, route, approval, or diff."}</span>
            <button class="primary" type="submit" ${state.busy ? "disabled" : ""}>Send</button>
          </div>
        </form>
      </main>

      <aside class="inspector" aria-label="Diffs, artifacts, route, approvals, and diagnostics">
        ${renderPanel("Progress", renderProgress())}
        ${renderPanel("Diffs", `<div class="empty">Diffs will appear after approved file edits.</div>`)}
        ${renderPanel("Artifacts", `<div class="empty">Generated artifacts and receipts appear here.</div>`)}
        ${renderPanel("Route", renderRoute(route))}
        ${renderPanel("Approvals", `<div class="empty">Side effects require Go approval before they run.</div>`)}
        ${renderPanel("Diagnostics", renderDiagnostics())}
      </aside>
    </div>
  `;
}

function renderProjects() {
  if (!state.projects.length) {
    return `<div class="empty">Open a local folder to inspect sessions.</div>`;
  }
  return state.projects.map((project) => `
    <button class="project-card">
      <strong>${escapeHTML(project.display_name)}</strong>
      <span>${escapeHTML(project.root_path_redacted)}</span>
    </button>
  `).join("");
}

function renderMessages() {
  if (!state.messages.length) {
    return `
      <div class="welcome">
        <p>Task in, useful result out.</p>
        <p>Simple questions answer compactly. File edits and commands request approval before side effects.</p>
      </div>
    `;
  }
  return state.messages.map((message) => `
    <article class="message ${message.role}">
      <div class="role">${message.role === "user" ? "You" : "Jini"}</div>
      <p>${escapeHTML(message.text)}</p>
    </article>
  `).join("");
}

function renderProgress() {
  const last = state.events[state.events.length - 1];
  if (!last) {
    return `<div class="empty">No active task.</div>`;
  }
  return `
    <dl>
      <dt>Last event</dt>
      <dd>${escapeHTML(last.event_type ?? last.type ?? "event")}</dd>
    </dl>
  `;
}

function renderRoute(route) {
  if (!route) {
    return `<div class="empty">Route status unavailable.</div>`;
  }
  const guidance = route.setup_guidance?.length
    ? `<ul>${route.setup_guidance.map((line) => `<li>${escapeHTML(line)}</li>`).join("")}</ul>`
    : `<div class="empty">No route setup action needed.</div>`;
  const help = state.routeHelp?.lines?.length
    ? `<details><summary>${escapeHTML(state.routeHelp.title)}</summary><pre>${escapeHTML(state.routeHelp.lines.join("\n"))}</pre></details>`
    : "";
  return `
    <dl>
      <dt>Selected</dt>
      <dd>${escapeHTML(route.label)} (${escapeHTML(route.route_kind)})</dd>
      <dt>Status</dt>
      <dd>${escapeHTML(route.status)}</dd>
      <dt>Reason</dt>
      <dd>${escapeHTML(route.reason || "Chosen automatically")}</dd>
    </dl>
    ${guidance}
    ${help}
  `;
}

function renderDiagnostics() {
  const exported = state.diagnostics
    ? `<p>Saved ${escapeHTML(state.diagnostics.bundle_path_redacted)}</p>`
    : `<p class="empty">Export a redacted support bundle when debugging app/core setup.</p>`;
  return `
    ${exported}
    <button class="secondary" data-action="diagnostics-export">Export diagnostics</button>
  `;
}

function renderPanel(title, body) {
  return `
    <section class="panel">
      <h2>${escapeHTML(title)}</h2>
      ${body}
    </section>
  `;
}

function escapeHTML(value) {
  return String(value ?? "")
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;");
}

app.addEventListener("submit", async (event) => {
  if (!event.target.matches('[data-action="submit-task"]')) {
    return;
  }
  event.preventDefault();
  const textarea = event.target.elements.task;
  const text = textarea.value.trim();
  if (!text) {
    return;
  }
  textarea.value = "";
  state.error = "";
  state.busy = true;
  state.messages.push({ role: "user", text });
  render();

  try {
    const result = await client.submitTurn(text);
    state.messages.push({ role: "assistant", text: result.assistant_text ?? "Request accepted." });
  } catch (error) {
    state.error = error.response?.error?.recovery ?? error.message;
    state.messages.push({ role: "assistant", text: error.response?.error?.message ?? "Jini could not complete that request." });
  } finally {
    state.busy = false;
    render();
  }
});

app.addEventListener("click", async (event) => {
  const action = event.target.closest("[data-action]")?.dataset.action;
  if (!action) {
    return;
  }
  if (action === "route-help") {
    state.routeHelp = await client.routeHelp();
    render();
  }
  if (action === "diagnostics-export") {
    try {
      state.diagnostics = await client.exportDiagnostics();
      state.error = "";
    } catch (error) {
      state.error = error.response?.error?.recovery ?? error.message;
    }
    render();
  }
  if (action === "open-project") {
    state.error = "Project picker is not implemented yet. Use the CLI for project-specific work in this slice.";
    render();
  }
});

client.onEvent((event) => {
  state.events.push(event.event ?? event);
  render();
});

render();

try {
  state.snapshot = await client.start();
  state.route = await client.routeStatus();
  state.projects = await client.listRecentProjects();
  state.error = "";
} catch (error) {
  state.error = error.response?.error?.recovery ?? error.message;
}
render();
