from __future__ import annotations

import json
import subprocess
import sys
import unittest
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
BOUNDARY_PATH = ROOT / "specs" / "public-repo-boundary.md"


class PublicRepoBoundaryDocsTests(unittest.TestCase):
    def test_boundary_doc_makes_app_code_private(self) -> None:
        text = BOUNDARY_PATH.read_text(encoding="utf-8")
        for marker in (
            "Commercial desktop/mobile app implementation also belongs in the private",
            "app-shell source code",
            "native wrapper projects",
            "host manifests",
            "desktop/mobile app implementation directories or generated app bundles",
            "public docs that describe the commercial apps and their distribution posture",
        ):
            with self.subTest(marker=marker):
                self.assertIn(marker, text)

    def test_validate_public_boundary_reports_apps_glob_as_forbidden(self) -> None:
        result = subprocess.run(
            [
                sys.executable,
                str(ROOT / "tools" / "jini.py"),
                "validate-public-boundary",
                "--format",
                "json",
            ],
            check=True,
            capture_output=True,
            text=True,
        )
        report = json.loads(result.stdout)
        checks = {check["id"]: check for check in report["checks"]}
        forbidden_paths = checks["forbidden-paths"]
        self.assertIn("apps/**", forbidden_paths["forbidden_globs"])
        self.assertEqual(forbidden_paths["status"], "ok")


if __name__ == "__main__":
    unittest.main()
