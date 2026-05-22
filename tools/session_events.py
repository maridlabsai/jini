from __future__ import annotations

import json
from pathlib import Path
from typing import Any


def append_session_event(
    session_dir: Path,
    *,
    event_type: str,
    recorded_at: str,
    payload: dict[str, Any],
) -> None:
    event_path = session_dir / "events.ndjson"
    event = {
        "recorded_at": recorded_at,
        "event_type": event_type,
        "payload": payload,
    }
    with event_path.open("a", encoding="utf-8") as handle:
        handle.write(json.dumps(event, sort_keys=True) + "\n")
