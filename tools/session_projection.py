from __future__ import annotations

from pathlib import Path
from typing import Any


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
        }
        for item in artifact_catalog.get("ready_now", [])
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
