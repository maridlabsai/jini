from __future__ import annotations

import unittest
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
README_PATH = ROOT / "README.md"
CLI_PATH = ROOT / "docs" / "cli.md"
INSTALL_PATH = ROOT / "docs" / "install.md"
SIMPLE_PATH = ROOT / "docs" / "simple.md"


def read(path: Path) -> str:
    return path.read_text(encoding="utf-8")


class PublicCliDocsTests(unittest.TestCase):
    def test_readme_points_public_users_at_commands_and_admin_help(self) -> None:
        text = read(README_PATH)
        for marker in (
            "jini commands",
            "small public command catalog",
            "jini admin help",
            "routes, bundles, or release plumbing",
        ):
            with self.subTest(marker=marker):
                self.assertIn(marker, text)

    def test_cli_guide_promotes_commands_as_public_catalog(self) -> None:
        text = read(CLI_PATH)
        for marker in (
            "title: Command Catalog",
            "Command catalog",
            "jini commands",
            "small product-facing command catalog",
            "jini admin help",
            "routes, bundles, or release plumbing",
            "Support commands when Jini points you there",
            "not because they are navigating a command tree by hand",
        ):
            with self.subTest(marker=marker):
                self.assertIn(marker, text)

    def test_install_page_uses_same_public_vs_admin_boundary(self) -> None:
        text = read(INSTALL_PATH)
        for marker in (
            "jini commands",
            "small public command list",
            "jini admin help",
            "routes, bundles, or release plumbing",
        ):
            with self.subTest(marker=marker):
                self.assertIn(marker, text)

    def test_quickstart_uses_same_public_vs_admin_boundary(self) -> None:
        text = read(SIMPLE_PATH)
        for marker in (
            "jini commands",
            "small public command list",
            "jini admin help",
            "routes, bundles, or release plumbing",
        ):
            with self.subTest(marker=marker):
                self.assertIn(marker, text)


if __name__ == "__main__":
    unittest.main()
