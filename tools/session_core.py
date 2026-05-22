from __future__ import annotations

from dataclasses import asdict, dataclass
from pathlib import Path
from typing import Any

import yaml


def normalize_session_id(raw: str) -> str:
    cleaned = "".join(ch.lower() if ch.isalnum() else "-" for ch in raw.strip())
    compact = "-".join(part for part in cleaned.split("-") if part)
    return compact or "session"


@dataclass
class CanonicalSession:
    session_id: str
    title: str
    goal: str
    status: str
    updated_at: str
    pack_dir: str
    work_unit_id: str
    current_artifact_id: str
    next_step: str
    review_safe: bool
    share_boundary: str

    def to_dict(self) -> dict[str, Any]:
        return asdict(self)

    def to_yaml(self) -> str:
        return yaml.safe_dump(self.to_dict(), sort_keys=False)

    @classmethod
    def from_dict(cls, payload: dict[str, Any]) -> "CanonicalSession":
        return cls(**payload)


def derive_session_id(pack_dir: Path, work_unit_id: str) -> str:
    candidate = work_unit_id.strip() or pack_dir.name
    return normalize_session_id(candidate)


def build_canonical_session(
    *,
    pack_dir: Path,
    work_unit_id: str,
    title: str,
    status: str,
    updated_at: str,
    current_artifact_id: str,
    next_step: str,
    review_safe: bool,
    share_boundary: str = "pack-local",
) -> CanonicalSession:
    session_id = derive_session_id(pack_dir, work_unit_id)
    return CanonicalSession(
        session_id=session_id,
        title=title,
        goal=title,
        status=status,
        updated_at=updated_at,
        pack_dir=str(pack_dir),
        work_unit_id=work_unit_id,
        current_artifact_id=current_artifact_id,
        next_step=next_step,
        review_safe=review_safe,
        share_boundary=share_boundary,
    )
