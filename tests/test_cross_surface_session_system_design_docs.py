from __future__ import annotations

import unittest
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
DESIGN_PATH = ROOT / "specs" / "cross-surface-session-system-and-dev-design.md"
PRD_PATH = ROOT / "specs" / "cross-surface-session-platform-prd.md"


def read(path: Path) -> str:
    return path.read_text(encoding="utf-8")


class CrossSurfaceSessionSystemDesignDocsTests(unittest.TestCase):
    def test_system_design_exists_and_has_required_sections(self) -> None:
        text = read(DESIGN_PATH)
        required_markers = [
            "# Cross-Surface Session System And Developer Design",
            "## Purpose",
            "## Design Goals",
            "## System Overview",
            "## Canonical Session Object",
            "session_id",
            "review_safe",
            "## Persistence Model",
            ".jini/",
            "events.ndjson",
            "projection.json",
            "## Resume Contract",
            "## Routing And Cost Engine",
            "## Surface Adapter Design",
            "## Sync And Identity Design",
            "## Module Design",
            "session_core.py",
            "route_engine.py",
            "## Testing Strategy",
            "## Delivery Sequence",
            "## First Runtime Slice",
            "## Out Of Scope For Slice 1",
            "## Migration From Current Runtime",
            "## Acceptance Proof For Slice 1",
            "## Definition Of Done",
            "connectivity_mode",
            "reconciliation_debt_count",
            "reconciliation_debt_cleared",
            "external_framework_context_imported",
        ]
        for marker in required_markers:
            with self.subTest(marker=marker):
                self.assertIn(marker, text)

    def test_design_is_explicitly_tied_to_charter(self) -> None:
        text = read(DESIGN_PATH)
        for marker in (
            "cost-optimizer first",
            "UX to feel second to none",
            "macOS",
            "Windows",
            "mobile",
            "CLI",
            "cheapest suitable route",
            "Continuation should be cheaper than restart",
            "Make the CLI speak the canonical session model",
            "`status` and `resume` should become the first public proof",
            "Jini should state that it is working in offline mode",
            "Claude Code",
            "Codex",
            "GitHub CLI",
        ):
            with self.subTest(marker=marker):
                self.assertIn(marker, text)

    def test_prd_links_to_design(self) -> None:
        text = read(PRD_PATH)
        self.assertIn(
            "[cross-surface-session-system-and-dev-design.md](./cross-surface-session-system-and-dev-design.md)",
            text,
        )


if __name__ == "__main__":
    unittest.main()
