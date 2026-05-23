from __future__ import annotations

import json
from pathlib import Path
from typing import Any

try:
    from .session_core import CanonicalSession
    from .session_events import append_session_event
    from .yaml_compat import YAMLError, safe_load
except ImportError:  # pragma: no cover - script execution path
    from session_core import CanonicalSession
    from session_events import append_session_event
    from yaml_compat import YAMLError, safe_load


class SessionStore:
    def __init__(self, state_root: Path) -> None:
        self.state_root = state_root
        self.sessions_root = state_root / "sessions"
        self.sessions_root.mkdir(parents=True, exist_ok=True)

    def current_session_path(self) -> Path:
        return self.state_root / "current-session.json"

    def session_dir(self, session_id: str) -> Path:
        path = self.sessions_root / session_id
        path.mkdir(parents=True, exist_ok=True)
        return path

    def save(
        self,
        session: CanonicalSession,
        *,
        projection: dict[str, Any],
        source: str,
    ) -> dict[str, Any]:
        session_dir = self.session_dir(session.session_id)
        (session_dir / "session.yaml").write_text(session.to_yaml(), encoding="utf-8")
        (session_dir / "projection.json").write_text(json.dumps(projection, indent=2) + "\n", encoding="utf-8")
        append_session_event(
            session_dir,
            event_type="session_refreshed",
            recorded_at=session.updated_at,
            payload={
                "session_id": session.session_id,
                "source": source,
                "pack_dir": session.pack_dir,
                "status": session.status,
                "next_step": session.next_step,
            },
        )
        pointer = {
            "schema_version": "0.1.0",
            "pointer_type": "JiniCurrentSession",
            "updated_at": session.updated_at,
            "session_id": session.session_id,
            "pack_dir": session.pack_dir,
            "work_unit_id": session.work_unit_id,
            "title": session.title,
            "status": session.status,
            "source": source,
        }
        self.current_session_path().write_text(json.dumps(pointer, indent=2) + "\n", encoding="utf-8")
        return pointer

    def load_current_pointer(self) -> dict[str, Any] | None:
        return self._load_json(self.current_session_path())

    def load_session(self, session_id: str) -> CanonicalSession | None:
        session_path = self.sessions_root / session_id / "session.yaml"
        if not session_path.exists():
            return None
        try:
            payload = safe_load(session_path.read_text(encoding="utf-8"))
        except (OSError, YAMLError):
            return None
        if not isinstance(payload, dict):
            return None
        return CanonicalSession.from_dict(payload)

    def load_projection(self, session_id: str) -> dict[str, Any] | None:
        return self._load_json(self.sessions_root / session_id / "projection.json")

    def _load_json(self, path: Path) -> dict[str, Any] | None:
        if not path.exists():
            return None
        try:
            payload = json.loads(path.read_text(encoding="utf-8"))
        except (OSError, json.JSONDecodeError):
            return None
        return payload if isinstance(payload, dict) else None
