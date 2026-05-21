from __future__ import annotations

import unittest
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
HOMEPAGE_PATH = ROOT / "docs" / "index.md"
LAYOUT_PATH = ROOT / "docs" / "_layouts" / "default.html"
PLAN_PATH = ROOT / "specs" / "docs-homepage-rewrite-plan.md"
TRUST_PATH = ROOT / "docs" / "proof.md"
OUTPUTS_PATH = ROOT / "docs" / "state-and-artifacts.md"


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
            "## Who Jini is for",
            "Good fit",
            "Not the best fit",
            "## Why buy Jini",
            "Raw model shell",
            "Use a raw model shell when you only need a one-off answer.",
            "## What Jini writes",
            "## Commands that matter",
            "Stored locally",
            "Not stored as product magic",
            "## See the product surface",
            "Install and first run",
            "Metrics and route evidence",
            "hidden cloud memory",
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

    def test_supporting_pages_use_buyability_language(self) -> None:
        trust = read(TRUST_PATH)
        outputs = read(OUTPUTS_PATH)
        for marker in (
            "cheap to inspect",
            "What a buyer should be able to verify quickly",
            "Inspectability instead of product magic",
        ):
            with self.subTest(marker=marker):
                self.assertIn(marker, trust)
        for marker in (
            "Outputs are part of the buyability story",
            "deliverables, continuation, and explicit risk",
            "It should open deliverables, not storage concepts.",
            "No storage-first labels",
        ):
            with self.subTest(marker=marker):
                self.assertIn(marker, outputs)


if __name__ == "__main__":
    unittest.main()
