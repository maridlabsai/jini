from __future__ import annotations

import unittest
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
HOMEPAGE_PATH = ROOT / "docs" / "index.md"
LAYOUT_PATH = ROOT / "docs" / "_layouts" / "default.html"
PLAN_PATH = ROOT / "specs" / "docs-homepage-rewrite-plan.md"


def read(path: Path) -> str:
    return path.read_text(encoding="utf-8")


class DocsHomepageRewriteTests(unittest.TestCase):
    def test_homepage_keeps_clear_install_compatibility_and_trust_markers(self) -> None:
        text = read(HOMEPAGE_PATH)
        for marker in (
            "One shell for AI work",
            "Cheap by default. Strong when needed.",
            "Works with Claude Code",
            "Works with Codex",
            "## What Jini writes",
            "## Commands that matter",
            "## Why trust it",
            "jini metrics",
        ):
            with self.subTest(marker=marker):
                self.assertIn(marker, text)

    def test_layout_uses_product_facing_nav_labels(self) -> None:
        text = read(LAYOUT_PATH)
        for marker in ("Quickstart", "Outputs", "Trust"):
            with self.subTest(marker=marker):
                self.assertIn(marker, text)

    def test_plan_documents_scope_and_nuget_decision(self) -> None:
        text = read(PLAN_PATH)
        for marker in (
            "Docs Homepage Rewrite Plan",
            "Use the Buddy site as a structure reference",
            "NuGet is not part of this first pass",
            "Homebrew",
            "pipx / PyPI",
            "winget / scoop",
        ):
            with self.subTest(marker=marker):
                self.assertIn(marker, text)


if __name__ == "__main__":
    unittest.main()
