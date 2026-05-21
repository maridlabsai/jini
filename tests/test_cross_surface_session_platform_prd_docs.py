from __future__ import annotations

import unittest
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
DOCTRINE_PATH = ROOT / "specs" / "lean-platform-doctrine.md"
PRD_PATH = ROOT / "specs" / "cross-surface-session-platform-prd.md"
FULL_PRD_PATH = ROOT / "specs" / "full-product-prd.md"


def read(path: Path) -> str:
    return path.read_text(encoding="utf-8")


class CrossSurfaceSessionPlatformPrdDocsTests(unittest.TestCase):
    def test_prd_exists_and_captures_charter(self) -> None:
        text = read(PRD_PATH)
        required_markers = [
            "# Cross-Surface Session Platform PRD",
            "## Product Charter",
            "## Non-Negotiable Product Order",
            "### 1. Cost Optimizer First",
            "### 2. UX Second To None",
            "### 3. Session Continuity Across Form Factors",
            "Supported form factors:",
            "macOS",
            "Windows",
            "mobile",
            "CLI",
            "## Canonical Session Model",
            "stable session id",
            "ready state",
            "missing state",
            "route evidence",
            "review-safe state",
            "## Surface Contract",
            "## Routing And Cost Contract",
            "## UX Rules",
            "## Trust Rules",
            "## Primary Metrics",
            "cross-surface-resume-success-rate",
            "## Rollout Priorities",
            "### Phase 1: Canonical Session Contract",
            "### Phase 2: Surface Parity",
            "### Phase 3: Cost And Continuity Evidence",
            "### Phase 4: UX Hardening",
            "## Definition Of Done",
        ]
        for marker in required_markers:
            with self.subTest(marker=marker):
                self.assertIn(marker, text)

    def test_doctrine_explicitly_includes_ux_and_cross_surface_continuity(self) -> None:
        text = read(DOCTRINE_PATH)
        for marker in (
            "### 5. UX Second To None",
            "### 6. Sessions First, Surface Second",
            "### UX Discipline",
            "### Continuity Discipline",
            "cross-surface-resume-success-rate",
            "recovery-time-after-interruption",
            "add a supported surface that cannot resume the same session model",
        ):
            with self.subTest(marker=marker):
                self.assertIn(marker, text)

    def test_full_prd_links_to_new_session_platform_prd(self) -> None:
        text = read(FULL_PRD_PATH)
        self.assertIn(
            "[cross-surface-session-platform-prd.md](./cross-surface-session-platform-prd.md)",
            text,
        )


if __name__ == "__main__":
    unittest.main()
