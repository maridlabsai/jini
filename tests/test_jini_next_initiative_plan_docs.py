from __future__ import annotations

import unittest
from pathlib import Path
import json
import os
import subprocess
import tempfile


ROOT = Path(__file__).resolve().parents[1]
PLAN_PATH = ROOT / "specs" / "jini-next-initiative-plan.md"
ARCHIVE_PATH = ROOT / "docs" / "archive" / "2026-06-02-codebase-snapshot-manifest.md"
PERSONAS_PATH = ROOT / "specs" / "dogfood-personas.yaml"
DOCTRINE_PATH = ROOT / "specs" / "lean-platform-doctrine.md"
PUBLISH_TOOL_PATH = ROOT / "tools" / "jini_validate.py"
CLI = ["/usr/bin/python3", str(ROOT / "tools" / "jini.py")]


def read(path: Path) -> str:
    return path.read_text(encoding="utf-8")


class JiniNextInitiativePlanDocsTests(unittest.TestCase):
    maxDiff = None

    def setUp(self) -> None:
        self.temp_dir = tempfile.TemporaryDirectory(prefix="jini-next-initiative-tests-")
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

    def test_plan_exists_and_covers_strategy_pillars(self) -> None:
        text = read(PLAN_PATH)
        required_markers = [
            "# Jini Next Initiative Plan",
            "## Archive Before Change",
            "## Planner Verdict",
            "## Architect Verdict",
            "move the execution kernel to Go",
            "keep Python as the compatibility shell",
            "## Persona Outcomes",
            "### Software Engineer",
            "### College Student",
            "### High School Student",
            "### Realtor",
            "## Contradictions To Resolve",
            "## State-Of-The-Art Requirements",
            "## Target Architecture",
            "### Layer 1: Go Session Kernel",
            "### Layer 3: Python Compatibility Shell",
            "### Offline Excellence",
            "### Accessibility",
            "### Security",
            "### Self-Learning",
            "### Traceability",
            "### Extensibility",
            "## Free And Commercial Split",
            "## Scorecard",
            "## SLO And SLA Framework",
            "CLI cold start p50",
            "offline continuation success rate",
            "adaptive platform switch success rate",
            "### Phase 0: Snapshot, Freeze, And Contract Cleanup",
            "### Phase 2: Go Kernel Slice",
            "### Phase 3: Offline-First Excellence",
            "### Phase 6: Commercial Adaptive Orchestration",
        ]
        for marker in required_markers:
            with self.subTest(marker=marker):
                self.assertIn(marker, text)

    def test_archive_manifest_records_free_and_commercial_snapshots(self) -> None:
        text = read(ARCHIVE_PATH)
        required_markers = [
            "# Codebase Snapshot Manifest",
            "fe5b94e1aa1d58b013f71d1843935ec795edf2e8",
            "b40ebffd4a01b2b389b0413117332129b973e4cd",
            "/Users/sharad.sharma/Developer/jini-archives/2026-06-02/jini-fe5b94e.bundle",
            "/Users/sharad.sharma/Developer/jini-archives/2026-06-02/jini-commercial-b40ebff.bundle",
            "git clone /Users/sharad.sharma/Developer/jini-archives/2026-06-02/jini-fe5b94e.bundle jini-restored",
            "git clone /Users/sharad.sharma/Developer/jini-archives/2026-06-02/jini-commercial-b40ebff.bundle jini-commercial-restored",
        ]
        for marker in required_markers:
            with self.subTest(marker=marker):
                self.assertIn(marker, text)

    def test_persona_catalog_covers_family_and_operator_panel(self) -> None:
        text = read(PERSONAS_PATH)
        required_markers = [
            "college-student-user",
            "high-school-student-user",
            "household-manager-user",
            "software-engineer-user",
            "realtor-user",
            "travel-advisor-user",
        ]
        for marker in required_markers:
            with self.subTest(marker=marker):
                self.assertIn(marker, text)

    def test_doctrine_no_longer_duplicates_offline_bullets(self) -> None:
        text = read(DOCTRINE_PATH)
        self.assertEqual(1, text.count("what reconciliation debt was accrued while working offline"))
        self.assertEqual(
            1,
            text.count(
                "When Jini is offline, it should still be able to benefit from locally available"
            ),
        )
        self.assertEqual(1, text.count("offline mode should be visible to the user instead of being inferred later"))

    def test_publish_readiness_includes_next_initiative_contract(self) -> None:
        publish_tool = read(PUBLISH_TOOL_PATH)
        self.assertIn("jini-next-initiative-plan.md", publish_tool)
        self.assertIn("2026-06-02-codebase-snapshot-manifest.md", publish_tool)
        self.assertIn("next-initiative", publish_tool)

        result = self.run_cli("publish-readiness", "--format", "json")
        self.assert_ok(result)
        report = json.loads(result.stdout)
        sections = {section["id"]: section for section in report["sections"]}
        self.assertEqual("ok", sections["next-initiative"]["status"])
        checks = sections["next-initiative"]["checks"]
        self.assertTrue(
            any(check.get("path") == "specs/jini-next-initiative-plan.md" for check in checks)
        )
        self.assertTrue(
            any(
                check.get("path") == "docs/archive/2026-06-02-codebase-snapshot-manifest.md"
                for check in checks
            )
        )


if __name__ == "__main__":
    unittest.main()
