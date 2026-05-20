from __future__ import annotations

import unittest
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
DOC_PATH = ROOT / "specs" / "delight-gap-closure.md"


def read(path: Path) -> str:
    return path.read_text(encoding="utf-8")


class DelightGapClosureDocsTests(unittest.TestCase):
    def test_doc_exists(self) -> None:
        self.assertTrue(DOC_PATH.exists())

    def test_doc_covers_competitor_understanding_ux_and_tech(self) -> None:
        text = read(DOC_PATH)
        required_markers = [
            "# Delight Gap Closure",
            "Claude Code overview:",
            "ChatGPT Projects:",
            "ChatGPT Canvas:",
            "OpenAI Codex docs:",
            "Work with Codex from anywhere:",
            "## Critique Rule",
            "If two or more critique sources independently say the same delight feature",
            "the default action is to remove, demote, hide, or collapse it",
            "Keeping the feature requires explicit proof",
            "## Next Delight Slice",
            "### 1. Inspectable Context Capsule",
            "### 2. Quick Artifact Rewrite Shortcuts",
            "### 3. Revision History And Undo",
            "#### Competitor pattern",
            "#### UX design",
            "#### Technical design",
            "#### Pass criteria",
            "## Why These Three",
            "## Delivery Order",
        ]
        for marker in required_markers:
            with self.subTest(marker=marker):
                self.assertIn(marker, text)


if __name__ == "__main__":
    unittest.main()
