from __future__ import annotations

import unittest
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
SLICE_PATH = ROOT / "specs" / "skills-and-delegation-slice.md"


def read(path: Path) -> str:
    return path.read_text(encoding="utf-8")


class SkillsAndDelegationSliceDocsTests(unittest.TestCase):
    def test_slice_doc_exists(self) -> None:
        self.assertTrue(SLICE_PATH.exists(), f"Missing required doc: {SLICE_PATH}")

    def test_slice_has_required_sections(self) -> None:
        text = read(SLICE_PATH)
        required = [
            "# Skills And Delegation Slice",
            "## UX Contract",
            "### 2. Explicit Delegation Uses One Verb",
            "### 3. Discovery Uses One Noun",
            "## File Format",
            "## Skill Discovery Roots",
            "### skill.yaml",
            "### prompt.md",
            "## Runtime Behavior",
            "### 1. `skills`",
            "### 2. `delegate <skill-id>`",
            "## Output Contract",
            "## Shipping Sequence",
            "## Tests",
            "## Non-Goals",
            "## Success Criteria",
        ]
        for marker in required:
            with self.subTest(marker=marker):
                self.assertIn(marker, text)

    def test_slice_commits_to_boring_explicit_surface(self) -> None:
        text = read(SLICE_PATH)
        required_markers = [
            "- one discovery command: `skills`",
            "- one execution command: `delegate`",
            "Phase 1 delegation is explicit only.",
            "The explicit command is:",
            "- `delegate`",
            "The discovery command is:",
            "- `skills`",
            "Jini should not yet ship:",
            "- automatic delegation by default",
            "- visible subagent trees",
            "- agent-role theater",
        ]
        for marker in required_markers:
            with self.subTest(marker=marker):
                self.assertIn(marker, text)

    def test_slice_defines_file_layout_and_state_contract(self) -> None:
        text = read(SLICE_PATH)
        required_markers = [
            "- project: `.jini/skills/`",
            "- user: `~/.jini/skills/`",
            ".jini/skills/reviewer/",
            "skill.yaml",
            "prompt.md",
            "work/<work-id>/delegations/<timestamp>-<skill-id>/",
            "request.json",
            "result.json",
            "summary.md",
            "`active_delegation_id`",
            "`active_skill_id`",
            "`active_delegation_status`",
        ]
        for marker in required_markers:
            with self.subTest(marker=marker):
                self.assertIn(marker, text)


if __name__ == "__main__":
    unittest.main()
