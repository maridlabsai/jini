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


def build_session_projection(
    *,
    session_id: str,
    pack_dir: Path,
    summary: dict[str, Any],
    artifact_catalog: dict[str, Any],
    updated_at: str,
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
    return {
        "schema_version": "0.1.0",
        "projection_type": "JiniSessionProjection",
        "session_id": session_id,
        "updated_at": updated_at,
        "pack_dir": str(pack_dir),
        "ready": ready_now,
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
