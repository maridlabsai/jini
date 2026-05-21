from __future__ import annotations

import unittest
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
MAKEFILE_PATH = ROOT / "Makefile"
CONTRIBUTING_PATH = ROOT / "CONTRIBUTING.md"


def read(path: Path) -> str:
    return path.read_text(encoding="utf-8")


class DeveloperExperienceDocsTests(unittest.TestCase):
    def test_makefile_exposes_standard_validation_targets(self) -> None:
        text = read(MAKEFILE_PATH)
        for marker in (
            ".PHONY: help test test-cli test-docs readiness",
            "make test",
            "python3 -m unittest discover -s tests -v",
            "python3 -m unittest tests/test_jini_cli.py -v",
            "python3 tools/jini.py publish-readiness --format json",
        ):
            with self.subTest(marker=marker):
                self.assertIn(marker, text)

    def test_contributing_points_developers_at_make_targets(self) -> None:
        text = read(CONTRIBUTING_PATH)
        for marker in (
            "make test-cli",
            "make test-docs",
            "make readiness",
            "make test",
        ):
            with self.subTest(marker=marker):
                self.assertIn(marker, text)


if __name__ == "__main__":
    unittest.main()
