from __future__ import annotations

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
    return {
        "schema_version": "0.1.0",
        "projection_type": "JiniSessionProjection",
        "session_id": session_id,
        "updated_at": updated_at,
        "pack_dir": str(pack_dir),
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
        "route": {
            "provider_id": "session-kernel",
            "reason": "Derived from current pack state and ready artifacts.",
        },
        "cost_posture": {
            "current_path": "continue-existing-work" if continuation_saved_work else "start-fresh",
            "continuation_saved_work": continuation_saved_work,
        },
    }
