from __future__ import annotations

import json
import os
import subprocess
import sys
import tempfile
import unittest
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
DOCTRINE_PATH = ROOT / "specs" / "lean-platform-doctrine.md"
GATE_PATH = ROOT / "specs" / "lean-platform-gate.md"
PUBLISH_TOOL_PATH = ROOT / "tools" / "jini_validate.py"
CLI = ["/usr/bin/python3", str(ROOT / "tools" / "jini.py")]


def read(path: Path) -> str:
    return path.read_text(encoding="utf-8")


class LeanPlatformDoctrineDocsTests(unittest.TestCase):
    maxDiff = None

    def setUp(self) -> None:
        self.temp_dir = tempfile.TemporaryDirectory(prefix="jini-lean-platform-tests-")
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

    def test_doctrine_and_gate_docs_exist(self) -> None:
        self.assertTrue(DOCTRINE_PATH.exists())
        self.assertTrue(GATE_PATH.exists())

    def test_doctrine_captures_lean_platform_rules(self) -> None:
        text = read(DOCTRINE_PATH)
        required_markers = [
            "# Lean Platform Doctrine",
            "## Mission",
            "the most dependable, cost-effective, and frictionless",
            "automating complex AI workflows across any environment",
            "## Core Principles",
            "### 1. Lowest Total Cost To Useful Outcome",
            "### 2. One Stable Surface",
            "### 3. Cheap By Default, Strong When Needed",
            "### 4. Visible Efficiency",
            "### 5. UX Second To None",
            "### 6. Sessions First, Surface Second",
            "### 7. Fewer Product Ideas, Better Execution",
            "## Buying Posture",
            "## Operating Rules",
            "### Command-Surface Discipline",
            "### Latency Discipline",
            "### UX Discipline",
            "### Continuity Discipline",
            "### Cost Discipline",
            "## Measures",
            "cost-per-successful-task",
            "time-to-first-useful-result",
            "resume-cost",
            "cross-surface-resume-success-rate",
            "recovery-time-after-interruption",
            "command-surface-count",
            "premium-route-regret-rate",
            "throttle-driven platform switching should be automatic",
            "task-shaped model selection should choose the smallest model or profile",
            "can do the job well, then escalate only when the task truly needs more depth",
            "whether Jini is currently operating in offline mode",
            "whether online capability is currently available",
            "what reconciliation debt was accrued while working offline",
            "Claude Code, Codex, or GitHub CLI",
            "throttle-avoided-interruption-rate",
            "platform-switch-success-rate",
            "task-to-model-fit accuracy",
            "offline-mode-transparency-rate",
            "reconciliation-debt-clearance-rate",
            "imported-context-reuse-rate",
            "## Reject Conditions",
            "compatibility aliases should not be taught",
            "add a supported surface that cannot resume the same session model",
            "hide that Jini is operating offline",
            "lose or silently ignore reconciliation debt",
        ]
        for marker in required_markers:
            with self.subTest(marker=marker):
                self.assertIn(marker, text)

    def test_gate_blocks_regressions_in_cost_latency_and_command_surface(self) -> None:
        gate = read(GATE_PATH)
        required_markers = [
            "# Lean Platform Gate",
            "## Gate Categories",
            "### 1. Cost Discipline",
            "### 2. Latency Discipline",
            "### 3. Command-Surface Discipline",
            "### 4. Visible Efficiency",
            "### 5. Buyability",
            "## Reject Conditions",
            "## Required Regression Inputs",
            "`lowest-total-cost-to-useful-outcome`",
            "`one-stable-surface`",
            "`cheap-by-default`",
            "`visible-efficiency`",
            "`fewer-product-ideas-better-execution`",
            "`cost-per-successful-task`",
            "`time-to-first-useful-result`",
            "`resume-cost`",
            "`command-surface-count`",
            "`route-feedback-health`",
            "`route-feedback-impact`",
            "`no-compatibility-aliases`",
            "`throttle-driven-platform-switching`",
            "`task-shaped-model-selection`",
        ]
        for marker in required_markers:
            with self.subTest(marker=marker):
                self.assertIn(marker, gate)

    def test_publish_readiness_includes_lean_platform_gate(self) -> None:
        publish_tool = read(PUBLISH_TOOL_PATH)
        self.assertIn("lean-platform-doctrine.md", publish_tool)
        self.assertIn("lean-platform-gate.md", publish_tool)
        self.assertIn("lean-platform", publish_tool)

        result = self.run_cli("publish-readiness", "--format", "json")
        self.assert_ok(result)
        report = json.loads(result.stdout)
        sections = {section["id"]: section for section in report["sections"]}
        self.assertEqual("ok", sections["lean-platform"]["status"])
        gate_check = next(
            check
            for check in sections["lean-platform"]["checks"]
            if check.get("path") == "specs/lean-platform-gate.md"
        )
        self.assertIn("`route-feedback-health`", gate_check["markers"])
        self.assertIn("`route-feedback-impact`", gate_check["markers"])
        runtime_check = next(
            check
            for check in sections["lean-platform"]["checks"]
            if check.get("path") == "runtime:lean-platform-metrics"
        )
        self.assertIn("route_feedback_health:tracked", runtime_check["markers"])
        self.assertIn("route_feedback_impact:measured-or-empty", runtime_check["markers"])
        self.assertIn("route_feedback_health", runtime_check["measurement"])
        self.assertIn("route_feedback_impact", runtime_check["measurement"])


if __name__ == "__main__":
    unittest.main()
