from __future__ import annotations

import unittest
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
ROLES_PATH = ROOT / "specs" / "product-review-roles.md"
RESEARCH_PATH = ROOT / "specs" / "friction-reduction-research.md"


def read(path: Path) -> str:
    return path.read_text(encoding="utf-8")


class ProductReviewRolesDocsTests(unittest.TestCase):
    def test_roles_anchor_ux_research_and_design_on_charter(self) -> None:
        text = read(ROLES_PATH)
        for marker in (
            "Jini is a cost optimizer first",
            "UX must feel second to none",
            "the same session should survive across macOS, Windows, mobile, and CLI",
            "the most dependable, cost-effective, and frictionless",
            "dependable and frictionless across any environment",
            "pay premium cost for work that should have stayed cheap",
            "restart work because the session cannot resume cleanly on another surface",
            "hide cost posture when it changes user trust or behavior",
            "make cross-surface resume feel like separate products instead of one session",
            "route cost posture is legible when it matters",
            "supported surfaces preserve one resumable session model",
        ):
            with self.subTest(marker=marker):
                self.assertIn(marker, text)

    def test_friction_research_has_explicit_ux_research_and_design_anchor(self) -> None:
        text = read(RESEARCH_PATH)
        for marker in (
            "Jini is a cost optimizer first",
            "Jini UX must be second to none",
            "Jini must preserve one resumable session across macOS, Windows, mobile, and",
            "## UX Research And Design Anchor",
            "it does not make the cheapest suitable path feel like the normal path",
            "it makes device switching feel like starting over",
            "cheaper to finish",
            "easier to resume",
            "cross-surface parity decisions",
        ):
            with self.subTest(marker=marker):
                self.assertIn(marker, text)


if __name__ == "__main__":
    unittest.main()
