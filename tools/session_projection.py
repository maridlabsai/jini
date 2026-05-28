from __future__ import annotations

import re
from pathlib import Path
from typing import Any


def _build_snapshot_markdown(path: Path, *, max_chars: int = 700) -> tuple[str, bool]:
    try:
        raw = path.read_text(encoding="utf-8").strip()
    except OSError:
        return "", False
    if len(raw) <= max_chars:
        return raw, False

    min_boundary = int(max_chars * 0.55)
    cut = -1
    for marker in ("\n\n", "\n", " "):
        candidate = raw.rfind(marker, 0, max_chars)
        if candidate >= min_boundary:
            cut = candidate
            break
    if cut < 0:
        cut = max_chars
    preview = raw[:cut].rstrip()
    return preview + "\n\n...", True


def _default_focus_item(ready_now: list[dict[str, Any]]) -> dict[str, Any] | None:
    if not ready_now:
        return None
    preferred_ids = ("tasks", "next-actions", "decision")
    indexed = {str(item.get("id", "")).strip().lower(): item for item in ready_now}
    for preferred_id in preferred_ids:
        if preferred_id in indexed:
            return indexed[preferred_id]
    if len(ready_now) > 1:
        return ready_now[1]
    return ready_now[0]


def _projection_focus_item(
    ready_now: list[dict[str, Any]],
    previous_projection: dict[str, Any] | None,
) -> dict[str, Any] | None:
    if isinstance(previous_projection, dict):
        current_focus = previous_projection.get("current_focus", {})
        if isinstance(current_focus, dict):
            focus_id = str(current_focus.get("artifact_id", "")).strip()
            focus_label = str(current_focus.get("artifact_label", "")).strip().lower()
            for item in ready_now:
                item_id = str(item.get("id", "")).strip()
                item_label = str(item.get("label", "")).strip().lower()
                if focus_id and focus_id == item_id:
                    return item
                if focus_label and focus_label == item_label:
                    return item
    return _default_focus_item(ready_now)


def _artifact_delta_items(ready_now: list[dict[str, Any]]) -> list[dict[str, str]]:
    return [
        {
            "artifact_id": str(item.get("id", "")).strip(),
            "title": str(item.get("label", "")).strip(),
            "status": "ready",
        }
        for item in ready_now
        if str(item.get("id", "")).strip()
    ]


def _compact_preview(text: str, *, max_chars: int = 180) -> str:
    cleaned = " ".join(text.strip().split())
    if len(cleaned) <= max_chars:
        return cleaned
    cut = cleaned.rfind(" ", 0, max_chars)
    if cut < int(max_chars * 0.55):
        cut = max_chars
    return cleaned[:cut].rstrip() + "..."


def _input_origin(item: dict[str, Any]) -> str:
    return (
        str(item.get("origin_ref", "")).strip()
        or str(item.get("title", "")).strip()
        or str(item.get("input_id", "")).strip()
        or "input"
    )


def _input_extraction_status(item: dict[str, Any]) -> str:
    explicit = str(item.get("extraction_status", "")).strip()
    if explicit:
        return explicit
    status = str(item.get("status", "")).strip().lower()
    if status == "processed":
        return "extracted"
    if status == "failed":
        return "failed"
    if status == "received":
        return "pending"
    if status == "superseded":
        return "superseded"
    return "unknown"


def _input_failure_reason(item: dict[str, Any]) -> str:
    return (
        str(item.get("failure_reason", "")).strip()
        or str(item.get("error_message", "")).strip()
        or str(item.get("preview", "")).strip()
        or "Processing failed."
    )


def _input_extraction_summary(item: dict[str, Any], extraction_status: str) -> str:
    explicit = str(item.get("extraction_summary", "")).strip()
    if explicit:
        return _compact_preview(explicit, max_chars=180)

    origin = _input_origin(item)
    preview = str(item.get("preview", "")).strip()
    kind = str(item.get("kind", "")).strip().lower()

    if extraction_status == "failed":
        return _compact_preview(f"Could not process {origin}: {_input_failure_reason(item)}", max_chars=180)
    if extraction_status in {"pending", "received"}:
        return _compact_preview(f"Waiting to extract {origin}.", max_chars=180)
    if extraction_status == "superseded":
        return _compact_preview(f"{origin} was superseded by a newer input.", max_chars=180)

    suffix = f": {preview}" if preview else "."
    if kind == "text":
        derived_count = len(
            [
                artifact_id
                for artifact_id in item.get("derived_artifact_ids", [])
                if str(artifact_id).strip()
            ]
        )
        if derived_count:
            return _compact_preview(
                f"Captured text request from {origin} and linked it to {derived_count} artifact(s).",
                max_chars=180,
            )
        return _compact_preview(f"Captured text request from {origin}.", max_chars=180)
    if kind == "image":
        return _compact_preview(f"Observed image input from {origin}{suffix}", max_chars=180)
    if kind == "audio":
        return _compact_preview(f"Transcribed audio input from {origin}{suffix}", max_chars=180)
    if kind == "link":
        return _compact_preview(f"Fetched link input from {origin}{suffix}", max_chars=180)
    if kind == "file" and origin.lower().endswith(".pdf"):
        return _compact_preview(f"Parsed PDF input from {origin}{suffix}", max_chars=180)
    if kind == "file":
        return _compact_preview(f"Parsed file input from {origin}{suffix}", max_chars=180)
    if kind == "derived":
        return _compact_preview(f"Derived working input from {origin}{suffix}", max_chars=180)
    return _compact_preview(f"Extracted {origin}{suffix}", max_chars=180)


def normalize_input_item(item: dict[str, Any]) -> dict[str, Any]:
    normalized = dict(item)
    extraction_status = _input_extraction_status(normalized)
    normalized["extraction_status"] = extraction_status
    normalized["extraction_summary"] = _input_extraction_summary(normalized, extraction_status)
    if extraction_status == "failed":
        normalized["failure_reason"] = _input_failure_reason(normalized)
    elif "failure_reason" in normalized and not str(normalized.get("failure_reason", "")).strip():
        normalized.pop("failure_reason", None)
    return normalized


def build_input_items(
    *,
    session_id: str,
    summary: dict[str, Any],
    ready_now: list[dict[str, Any]],
    updated_at: str,
    previous_projection: dict[str, Any] | None = None,
) -> list[dict[str, Any]]:
    previous_initial: dict[str, Any] = {}
    preserved_items: list[dict[str, Any]] = []
    if isinstance(previous_projection, dict):
        previous_items = previous_projection.get("input_items")
        if isinstance(previous_items, list) and previous_items:
            for item in previous_items:
                if isinstance(item, dict) and item.get("input_id") == "initial-request":
                    previous_initial = item
                elif isinstance(item, dict):
                    preserved_items.append(normalize_input_item(item))

    work_unit = summary.get("work_unit", {})
    title = str(work_unit.get("title", "")).strip() or "Initial work request"
    purpose = str(work_unit.get("purpose", "")).strip() or title
    source_actor = str(work_unit.get("owner_actor_id", "")).strip() or "user"
    created_at = (
        str(work_unit.get("created_at", "")).strip()
        or str(previous_initial.get("created_at", "")).strip()
        or updated_at
    )
    derived_artifact_ids = [
        str(item.get("id", "")).strip()
        for item in ready_now
        if str(item.get("id", "")).strip()
    ]

    initial_item = normalize_input_item(
        {
            "input_id": "initial-request",
            "thread_id": session_id,
            "kind": "text",
            "title": title,
            "source_actor": source_actor,
            "status": "processed",
            "preview": _compact_preview(purpose),
            "origin_ref": "work-unit.yaml",
            "derived_artifact_ids": derived_artifact_ids,
            "created_at": created_at,
            "updated_at": updated_at,
        }
    )

    return [
        initial_item,
        *preserved_items,
    ]


def _card_artifact_id(label: str) -> str:
    slug = re.sub(r"[^a-z0-9]+", "-", label.strip().lower()).strip("-")
    return slug or "artifact"


def _snapshot_summary(snapshot_markdown: str, fallback: str) -> str:
    for raw_line in snapshot_markdown.splitlines():
        line = raw_line.strip()
        if not line:
            continue
        if line.startswith("#"):
            line = line.lstrip("#").strip()
        if line:
            return _compact_preview(line, max_chars=120)
    return fallback


def _source_input_ids_for_artifact(artifact_id: str, input_items: list[dict[str, Any]]) -> list[str]:
    matches = [
        str(item.get("input_id", "")).strip()
        for item in input_items
        if isinstance(item, dict)
        and artifact_id in {str(value).strip() for value in item.get("derived_artifact_ids", [])}
        and str(item.get("input_id", "")).strip()
    ]
    if matches:
        return matches
    return [
        str(item.get("input_id", "")).strip()
        for item in input_items
        if isinstance(item, dict) and str(item.get("input_id", "")).strip()
    ][:1]


def build_artifact_shelf(
    *,
    session_id: str,
    summary: dict[str, Any],
    ready_now: list[dict[str, Any]],
    input_items: list[dict[str, Any]],
    updated_at: str,
) -> dict[str, Any]:
    ready_cards: list[dict[str, Any]] = []
    for item in ready_now:
        artifact_id = str(item.get("id", "")).strip()
        if not artifact_id:
            continue
        title = str(item.get("label", "")).strip() or artifact_id
        snapshot = str(item.get("snapshot_markdown", "")).strip()
        ready_cards.append(
            {
                "schema_version": "0.1.0",
                "card_type": "JiniArtifactCard",
                "artifact_id": artifact_id,
                "thread_id": session_id,
                "artifact_type": title,
                "title": title,
                "status": "ready",
                "summary": _snapshot_summary(snapshot, f"{title} is ready."),
                "preview": _compact_preview(snapshot or title, max_chars=220),
                "open_action": {
                    "label": "Open",
                    "command": f"jini open {artifact_id}",
                },
                "export_actions": [
                    {
                        "label": "Show",
                        "command": f"jini show {artifact_id}",
                    },
                    {
                        "label": "Print path",
                        "command": f"jini open {artifact_id} --print-path",
                    },
                ],
                "source_input_ids": _source_input_ids_for_artifact(artifact_id, input_items),
                "updated_at": updated_at,
            }
        )

    needs_input_cards = [
        {
            "schema_version": "0.1.0",
            "card_type": "JiniArtifactCard",
            "artifact_id": _card_artifact_id(str(artifact_type)),
            "thread_id": session_id,
            "artifact_type": str(artifact_type).strip(),
            "title": str(artifact_type).strip(),
            "status": "needs_input",
            "summary": f"Missing required artifact: {str(artifact_type).strip()}",
            "preview": "Jini needs this before the current stage can move forward.",
            "open_action": {
                "label": "Inspect status",
                "command": "jini status",
            },
            "export_actions": [],
            "source_input_ids": [],
            "updated_at": updated_at,
        }
        for artifact_type in summary.get("missing_stage_required", [])
        if str(artifact_type).strip()
    ]

    missing_blocker_prefixes = {
        f"Missing required artifact: {str(artifact_type).strip()}"
        for artifact_type in summary.get("missing_stage_required", [])
        if str(artifact_type).strip()
    }
    blocked_cards = [
        {
            "schema_version": "0.1.0",
            "card_type": "JiniArtifactCard",
            "artifact_id": f"blocked-{index}",
            "thread_id": session_id,
            "artifact_type": "Blocker",
            "title": "Blocked work",
            "status": "blocked",
            "summary": blocker,
            "preview": blocker,
            "open_action": {
                "label": "Inspect status",
                "command": "jini status",
            },
            "export_actions": [],
            "source_input_ids": [],
            "updated_at": updated_at,
        }
        for index, blocker in enumerate(
            [
                str(item).strip()
                for item in summary.get("blockers", [])
                if str(item).strip() and str(item).strip() not in missing_blocker_prefixes
            ],
            start=1,
        )
    ]

    return {
        "schema_version": "0.1.0",
        "shelf_type": "JiniArtifactShelf",
        "groups": ["ready_now", "needs_input", "blocked"],
        "ready_now": {
            "label": "Ready now",
            "cards": ready_cards,
        },
        "needs_input": {
            "label": "Needs input",
            "cards": needs_input_cards,
        },
        "blocked": {
            "label": "Blocked",
            "cards": blocked_cards,
        },
    }


def build_progress_snapshot(summary: dict[str, Any], ready_now: list[dict[str, Any]]) -> dict[str, Any]:
    work_unit = summary.get("work_unit", {})
    title = str(work_unit.get("title", "")).strip()
    state = str(work_unit.get("current_state", "")).strip()
    health = str(summary.get("health", "")).strip()
    next_operation = str(summary.get("next_operation", "")).strip()
    task_summary = summary.get("task_summary", {})
    missing_now = [str(item).strip() for item in summary.get("missing_stage_required", []) if str(item).strip()]
    blockers = [str(item).strip() for item in summary.get("blockers", []) if str(item).strip()]
    ready_labels = [str(item.get("label", "")).strip() or str(item.get("id", "")).strip() for item in ready_now]

    if ready_labels:
        working_with_summary = f"{len(ready_labels)} ready artifact(s): {', '.join(ready_labels[:3])}"
    else:
        working_with_summary = "No ready artifacts yet"

    total_tasks = int(task_summary.get("total", 0) or 0)
    done_tasks = int(task_summary.get("done", 0) or 0)
    if total_tasks:
        done = f"{done_tasks}/{total_tasks} tasks completed"
    elif ready_labels:
        done = f"{len(ready_labels)} ready artifact(s) available"
    else:
        done = "No completed artifact evidence yet"

    if missing_now:
        need = "Missing now: " + ", ".join(missing_now[:3])
    elif blockers:
        need = blockers[0]
    else:
        need = "Nothing stage-critical is missing right now"

    return {
        "goal": title,
        "working_with_summary": working_with_summary,
        "now": f"{state} / {health}".strip(" /"),
        "done": done,
        "need": need,
        "next": next_operation,
        "safe_to_do": not bool(summary.get("validation_errors")) and not bool(missing_now),
    }


def _active_asks(summary: dict[str, Any]) -> list[dict[str, Any]]:
    validation_errors = [str(item).strip() for item in summary.get("validation_errors", []) if str(item).strip()]
    missing_now = [str(item).strip() for item in summary.get("missing_stage_required", []) if str(item).strip()]
    blockers = [str(item).strip() for item in summary.get("blockers", []) if str(item).strip()]

    if validation_errors:
        return [
            {
                "ask_id": "fix-validation",
                "prompt": "Fix validation errors before continuing.",
                "reason": validation_errors[0],
                "blocking": True,
            }
        ]
    if missing_now:
        return [
            {
                "ask_id": f"provide-{missing_now[0].lower()}",
                "prompt": f"Provide {missing_now[0]} before the next state transition.",
                "reason": "The current stage requires this artifact.",
                "blocking": True,
            }
        ]
    if blockers:
        return [
            {
                "ask_id": "resolve-blocker",
                "prompt": blockers[0],
                "reason": "This blocks the next safe transition.",
                "blocking": True,
            }
        ]
    return []


def build_turn_record(
    *,
    session_id: str,
    summary: dict[str, Any],
    ready_now: list[dict[str, Any]],
    input_items: list[dict[str, Any]] | None = None,
    updated_at: str,
    previous_projection: dict[str, Any] | None = None,
) -> dict[str, Any]:
    current_state = str(summary.get("work_unit", {}).get("current_state", "")).strip()
    current_next = str(summary.get("next_operation", "")).strip()
    ready_ids = [str(item.get("id", "")).strip() for item in ready_now if str(item.get("id", "")).strip()]
    input_ids = [
        str(item.get("input_id", "")).strip()
        for item in input_items or []
        if isinstance(item, dict) and str(item.get("input_id", "")).strip()
    ]

    if isinstance(previous_projection, dict):
        previous_turn = previous_projection.get("turn_record")
        previous_ready_ids = [
            str(item.get("id", "")).strip()
            for item in previous_projection.get("ready", [])
            if isinstance(item, dict) and str(item.get("id", "")).strip()
        ]
        unchanged_state = str(previous_projection.get("state", "")).strip() == current_state
        unchanged_next = str(previous_projection.get("next", "")).strip() == current_next
        if isinstance(previous_turn, dict) and previous_ready_ids == ready_ids and unchanged_state and unchanged_next:
            previous_input_ids = [
                str(item).strip()
                for item in previous_turn.get("user_input_ids", [])
                if str(item).strip()
            ]
            if previous_input_ids == input_ids:
                return dict(previous_turn)

    state_changes: list[dict[str, str]] = []
    previous_state = ""
    previous_next = ""
    if isinstance(previous_projection, dict):
        previous_state = str(previous_projection.get("state", "")).strip()
        previous_next = str(previous_projection.get("next", "")).strip()
    if previous_state != current_state:
        state_changes.append({"field": "state", "before": previous_state, "after": current_state})
    if previous_next != current_next:
        state_changes.append({"field": "next", "before": previous_next, "after": current_next})

    ready_count = len(ready_ids)
    artifact_phrase = f"{ready_count} ready artifact(s)" if ready_count else "no ready artifacts"
    return {
        "schema_version": "0.1.0",
        "record_type": "JiniTurnRecord",
        "turn_id": f"{session_id}-latest",
        "thread_id": session_id,
        "user_input_ids": input_ids,
        "assistant_message": f"Session projection refreshed with {artifact_phrase}; next is {current_next}.",
        "artifacts_created": _artifact_delta_items(ready_now),
        "artifacts_updated": [],
        "state_changes": state_changes,
        "asks_opened": _active_asks(summary),
        "asks_resolved": [],
        "route_decision": {
            "provider_id": "session-kernel",
            "reason": "Derived from current pack state and ready artifacts.",
        },
        "started_at": updated_at,
        "completed_at": updated_at,
    }


def build_session_projection(
    *,
    session_id: str,
    pack_dir: Path,
    summary: dict[str, Any],
    artifact_catalog: dict[str, Any],
    updated_at: str,
    previous_projection: dict[str, Any] | None = None,
) -> dict[str, Any]:
    ready_now = [
        {
            "id": item["id"],
            "label": item["label"],
            "resolved_path": item["resolved_path"],
            "snapshot_markdown": snapshot_markdown,
            "snapshot_trimmed": snapshot_trimmed,
        }
        for item in artifact_catalog.get("ready_now", [])
        for snapshot_markdown, snapshot_trimmed in [
            _build_snapshot_markdown(Path(str(item.get("resolved_path", "")).strip()))
        ]
    ]
    missing = list(summary.get("missing_stage_required", []))
    task_summary = summary.get("task_summary", {})
    continuation_saved_work = bool(ready_now) or int(task_summary.get("done", 0) or 0) > 0
    focus_item = _projection_focus_item(ready_now, previous_projection)
    progress_snapshot = build_progress_snapshot(summary, ready_now)
    input_items = build_input_items(
        session_id=session_id,
        summary=summary,
        ready_now=ready_now,
        updated_at=updated_at,
        previous_projection=previous_projection,
    )
    artifact_shelf = build_artifact_shelf(
        session_id=session_id,
        summary=summary,
        ready_now=ready_now,
        input_items=input_items,
        updated_at=updated_at,
    )
    turn_record = build_turn_record(
        session_id=session_id,
        summary=summary,
        ready_now=ready_now,
        input_items=input_items,
        updated_at=updated_at,
        previous_projection=previous_projection,
    )
    return {
        "schema_version": "0.1.0",
        "projection_type": "JiniSessionProjection",
        "session_id": session_id,
        "updated_at": updated_at,
        "pack_dir": str(pack_dir),
        "state": str(summary.get("work_unit", {}).get("current_state", "")),
        "ready": ready_now,
        "current_focus": (
            {
                "kind": "artifact",
                "artifact_id": str(focus_item.get("id", "")).strip(),
                "artifact_label": str(focus_item.get("label", "")).strip(),
            }
            if isinstance(focus_item, dict)
            else {}
        ),
        "missing": missing,
        "next": str(summary.get("next_operation", "")),
        "input_items": input_items,
        "artifact_shelf": artifact_shelf,
        "artifact_cards": [
            *artifact_shelf["ready_now"]["cards"],
            *artifact_shelf["needs_input"]["cards"],
            *artifact_shelf["blocked"]["cards"],
        ],
        "progress_snapshot": progress_snapshot,
        "turn_record": turn_record,
        "route": {
            "provider_id": "session-kernel",
            "reason": "Derived from current pack state and ready artifacts.",
        },
        "cost_posture": {
            "current_path": "continue-existing-work" if continuation_saved_work else "start-fresh",
            "continuation_saved_work": continuation_saved_work,
        },
    }
