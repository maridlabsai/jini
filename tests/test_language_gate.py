from __future__ import annotations

import subprocess
import sys
import tempfile
import unittest
from pathlib import Path

from tools import language_gate


ROOT = Path(__file__).resolve().parents[1]


class LanguageGateTests(unittest.TestCase):
    def test_scan_paths_flags_blocked_language(self) -> None:
        with tempfile.TemporaryDirectory() as tmpdir:
            path = Path(tmpdir) / "note.md"
            path.write_text("This is bullshit.\n", encoding="utf-8")

            matches = language_gate.scan_paths([path])

        self.assertEqual(1, len(matches))
        self.assertEqual("crude-dismissal", matches[0].label)
        self.assertEqual(1, matches[0].line)

    def test_scan_paths_skips_non_user_facing_files(self) -> None:
        with tempfile.TemporaryDirectory() as tmpdir:
            path = Path(tmpdir) / "note.py"
            path.write_text("print('bullshit')\n", encoding="utf-8")

            matches = language_gate.scan_paths([path])

        self.assertEqual([], matches)

    def test_changed_paths_includes_untracked_user_facing_files(self) -> None:
        with tempfile.NamedTemporaryFile(dir=ROOT, suffix=".md", prefix=".language-gate-", delete=False) as handle:
            path = Path(handle.name)
        try:
            path.write_text("Plain public note.\n", encoding="utf-8")

            paths = language_gate.changed_paths()

            self.assertIn(path, paths)
        finally:
            path.unlink(missing_ok=True)

    def test_cli_fails_on_blocked_language(self) -> None:
        with tempfile.TemporaryDirectory() as tmpdir:
            path = Path(tmpdir) / "note.md"
            path.write_text("This is bullshit.\n", encoding="utf-8")

            result = subprocess.run(
                [sys.executable, "tools/language_gate.py", str(path)],
                cwd=ROOT,
                text=True,
                stdout=subprocess.PIPE,
                stderr=subprocess.PIPE,
                check=False,
            )

        self.assertNotEqual(0, result.returncode)
        self.assertIn("language-gate: blocked language found", result.stderr)
        self.assertIn("crude-dismissal", result.stderr)

    def test_required_gate_runs_language_gate(self) -> None:
        gate_text = (ROOT / "tools" / "run_required_gates.sh").read_text(encoding="utf-8")
        self.assertIn("python3 tools/language_gate.py", gate_text)


if __name__ == "__main__":
    unittest.main()
