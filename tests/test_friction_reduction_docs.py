from __future__ import annotations

import json
import os
import subprocess
import sys
import tempfile
import unittest
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
RESEARCH_PATH = ROOT / "specs" / "friction-reduction-research.md"
GATE_PATH = ROOT / "specs" / "friction-reduction-gate.md"
PUBLISH_TOOL_PATH = ROOT / "tools" / "jini_validate.py"
CLI = [sys.executable, str(ROOT / "tools" / "jini.py")]


def read(path: Path) -> str:
    return path.read_text(encoding="utf-8")


class FrictionReductionDocsTests(unittest.TestCase):
    maxDiff = None

    def setUp(self) -> None:
        self.temp_dir = tempfile.TemporaryDirectory(prefix="jini-friction-tests-")
        self.tmp = Path(self.temp_dir.name)

    def tearDown(self) -> None:
        self.temp_dir.cleanup()

    def run_cli(self, *args: object) -> subprocess.CompletedProcess[str]:
        env = dict(os.environ)
        env["JINI_STATE_DIR"] = str((self.tmp / ".jini").resolve())
        return subprocess.run(
            [*CLI, *[str(arg) for arg in args]],
            cwd=ROOT,
            text=True,
            capture_output=True,
            env=env,
        )

    def assert_ok(self, result: subprocess.CompletedProcess[str]) -> None:
        if result.returncode != 0:
            self.fail(
                f"Expected command to succeed.\nSTDOUT:\n{result.stdout}\nSTDERR:\n{result.stderr}"
            )

    def test_research_and_gate_docs_exist(self) -> None:
        self.assertTrue(RESEARCH_PATH.exists())
        self.assertTrue(GATE_PATH.exists())

    def test_research_uses_official_competitor_sources(self) -> None:
        text = read(RESEARCH_PATH)
        required_links = [
            "https://help.openai.com/en/articles/11369540/",
            "https://platform.openai.com/docs/codex",
            "https://openai.com/index/work-with-codex-from-anywhere/",
            "https://help.openai.com/en/articles/10169521-using-projects-in-chatgpt",
            "https://help.openai.com/en/articles/9930697-what-is-the-canvas-feature-in-chatgpt-and-how-do-i-use-it",
            "https://help.openai.com/en/articles/10291617-scheduled-tasks-in-chatgpt",
            "https://docs.anthropic.com/en/docs/claude-code/overview",
            "https://docs.anthropic.com/en/docs/claude-code/slash-commands",
            "https://docs.anthropic.com/en/docs/claude-code/hooks",
            "https://docs.anthropic.com/en/docs/claude-code/memory",
            "https://support.anthropic.com/en/articles/9487310-what-are-artifacts-and-how-do-i-use-them",
        ]
        for link in required_links:
            with self.subTest(link=link):
                self.assertIn(link, text)

    def test_research_translates_competitors_into_jini_principles(self) -> None:
        text = read(RESEARCH_PATH)
        required_markers = [
            "## Competitive Lessons",
            "### Codex",
            "### ChatGPT",
            "### Claude",
            "## Jini Friction Principles",
            "### 1. One Prompt Before Taxonomy",
            "### 2. No Empty-Shell Noise",
            "### 4. Artifact Escalation",
            "### 5. Continue Anywhere",
            "### 8. Best Productivity With Least Expense",
            "## What Jini Should Not Do",
            "## Measurement",
            "time to first useful result",
            "premium-route regret rate after user edits",
        ]
        for marker in required_markers:
            with self.subTest(marker=marker):
                self.assertIn(marker, text)

    def test_gate_blocks_high_friction_regressions(self) -> None:
        gate = read(GATE_PATH)
        required_markers = [
            "# Friction Reduction Gate",
            "## Gate Categories",
            "### 1. First-Minute Simplicity",
            "### 2. Natural Intent Handling",
            "### 3. Continue-Anywhere Work State",
            "### 4. Artifact Escalation",
            "### 5. Setup Doctor And Self-Healing",
            "### 6. Cost And Route Minimalism",
            "### 7. Trust Without Ceremony",
            "## Reject Conditions",
            "## Required Regression Inputs",
            "`continue-anywhere`",
            "`artifact-escalation`",
            "`setup-doctor`",
            "`best-productivity-least-expense`",
            "`visible-trust`",
        ]
        for marker in required_markers:
            with self.subTest(marker=marker):
                self.assertIn(marker, gate)

    def test_publish_readiness_includes_friction_reduction_gate(self) -> None:
        publish_tool = read(PUBLISH_TOOL_PATH)
        self.assertIn("friction-reduction-research.md", publish_tool)
        self.assertIn("friction-reduction-gate.md", publish_tool)
        self.assertIn("friction-reduction", publish_tool)

        result = self.run_cli("publish-readiness", "--format", "json")
        self.assert_ok(result)
        report = json.loads(result.stdout)
        sections = {section["id"]: section for section in report["sections"]}
        self.assertEqual("ok", sections["friction-reduction"]["status"])


if __name__ == "__main__":
    unittest.main()
