from __future__ import annotations

import unittest
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
MAKEFILE_PATH = ROOT / "Makefile"
CONTRIBUTING_PATH = ROOT / "CONTRIBUTING.md"
PREVIEW_TOOL_PATH = ROOT / "tools" / "preview_docs.sh"
DOCS_GEMFILE_PATH = ROOT / "docs" / "Gemfile"


def read(path: Path) -> str:
    return path.read_text(encoding="utf-8")


class DeveloperExperienceDocsTests(unittest.TestCase):
    def test_makefile_exposes_standard_validation_targets(self) -> None:
        text = read(MAKEFILE_PATH)
        for marker in (
            ".PHONY: help test test-cli test-docs readiness preview-docs build-docs",
            "make test",
            "python3 -m unittest discover -s tests -v",
            "python3 -m unittest tests/test_jini_cli.py -v",
            "python3 tools/jini.py publish-readiness --format json",
            "make preview-docs",
            "make build-docs",
            "bash tools/preview_docs.sh serve",
            "bash tools/preview_docs.sh build",
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
            "make preview-docs",
            "make build-docs",
        ):
            with self.subTest(marker=marker):
                self.assertIn(marker, text)

    def test_repo_ships_reproducible_docs_preview_tooling(self) -> None:
        self.assertTrue(PREVIEW_TOOL_PATH.exists(), "docs preview helper should exist")
        self.assertTrue(DOCS_GEMFILE_PATH.exists(), "docs preview gemfile should exist")

        tool_text = read(PREVIEW_TOOL_PATH)
        gemfile_text = read(DOCS_GEMFILE_PATH)
        for marker in (
            "Usage: tools/preview_docs.sh",
            "bundle exec jekyll",
            "bundle install",
            "ca-certificates",
            "docker",
            "jekyll/jekyll:4",
            "vendor/bundle",
            "JINI_DOCS_EXTRA_CA_CERT",
            "jini-preview-extra-ca.crt",
            ".preview-config.local.yml",
            "exclude:",
            "--dry-run",
            "docs/Gemfile",
            "_config.yml",
        ):
            with self.subTest(marker=marker):
                self.assertIn(marker, tool_text)

        for marker in (
            'gem "jekyll"',
            'gem "webrick"',
        ):
            with self.subTest(marker=marker):
                self.assertIn(marker, gemfile_text)


if __name__ == "__main__":
    unittest.main()
