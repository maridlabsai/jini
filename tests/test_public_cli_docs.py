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

    def test_cli_guide_keeps_support_commands_on_the_runtime_surface(self) -> None:
        text = read(CLI_PATH)
        for marker in (
            "jini status",
            "jini doctor",
        ):
            with self.subTest(marker=marker):
                self.assertIn(marker, text)
        for marker in (
            "jini check",
            "jini setup",
            "jini open",
            "jini metrics",
        ):
            with self.subTest(marker=marker):
                self.assertNotIn(marker, text)

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

    def test_install_page_explains_installer_provenance_summary(self) -> None:
        text = read(INSTALL_PATH)
        for marker in (
            "- install source: release binary",
            "- install source: source runtime (explicit source)",
            "- install source: source runtime (release-unavailable)",
            "- install source: source fallback (release validation failed: unsupported-public-command-surface)",
        ):
            with self.subTest(marker=marker):
                self.assertIn(marker, text)

    def test_install_page_maps_provenance_lines_to_support_actions(self) -> None:
        text = read(INSTALL_PATH)
        for marker in (
            "Install source line",
            "What it usually means",
            "What to do next",
            "Matching published release asset was available, passed Jini's public command check, and was accepted without any source fallback.",
            "No action needed unless you expected a source install for local development.",
            "The absence of <code>- support receipt: ...</code> and <code>- next step: ...</code> is expected on this path.",
            "No matching release asset was available for this machine or release channel.",
            "If this machine should have had a published release, file a release issue and include the receipt.",
            "Downloaded release binary did not support the current public command surface.",
            "Keep the source install, attach install-receipt.txt, and flag the stale release artifact.",
        ):
            with self.subTest(marker=marker):
                self.assertIn(marker, text)

    def test_install_page_explains_receipt_next_step_field(self) -> None:
        text = read(INSTALL_PATH)
        for marker in (
            "install-receipt.txt",
            "source_reason=",
            "release_validation=",
            "next_step=",
            "- support receipt: /path/to/install-receipt.txt (send version=, source_reason=, release_validation=, next_step=)",
        ):
            with self.subTest(marker=marker):
                self.assertIn(marker, text)
        self.assertGreaterEqual(
            text.count(
                "- support receipt: /path/to/install-receipt.txt (send version=, source_reason=, release_validation=, next_step=)"
            ),
            4,
        )

    def test_install_page_shows_source_path_terminal_example(self) -> None:
        text = read(INSTALL_PATH)
        for marker in (
            "Example source-path follow-up output:",
            "- install source: source runtime (release-unavailable)",
            "- support receipt: /Users/you/.local/bin/install-receipt.txt (send version=, source_reason=, release_validation=, next_step=)",
            "- next step: If this machine should have had a published release, file a release issue and include the receipt.",
        ):
            with self.subTest(marker=marker):
                self.assertIn(marker, text)

    def test_install_page_shows_release_binary_terminal_example(self) -> None:
        text = read(INSTALL_PATH)
        for marker in (
            "Example release-binary success output:",
            "- install source: release binary",
            "jini",
            "The healthy release-binary path stops there.",
            "It does not print a <code>- support receipt: ...</code> line or a <code>- next step: ...</code> line.",
        ):
            with self.subTest(marker=marker):
                self.assertIn(marker, text)

    def test_install_page_lists_minimal_support_receipt_keys(self) -> None:
        text = read(INSTALL_PATH)
        for marker in (
            "When support asks for install details on a source-path install, send the support receipt path plus only these receipt keys:",
            "install-receipt.txt</code> path from the printed <code>- support receipt: ...</code> line",
            "version=",
            "source_reason=",
            "release_validation=",
            "next_step=",
            "If the install output shows only <code>- install source: release binary</code> followed by <code>jini</code>, support can ignore this checklist.",
            "If <code>next_step=</code> is missing, the release-binary path likely succeeded and no extra follow-up was required.",
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
