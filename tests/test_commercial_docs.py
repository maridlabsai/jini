from __future__ import annotations

import unittest
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
HOMEPAGE_PATH = ROOT / "docs" / "index.md"
COMMERCIAL_PATH = ROOT / "docs" / "commercial.md"
INSTALL_PATH = ROOT / "docs" / "install.md"


def read(path: Path) -> str:
    return path.read_text(encoding="utf-8")


class CommercialDocsTests(unittest.TestCase):
    def test_homepage_mentions_non_cli_apps_and_subscription_boundary(self) -> None:
        text = read(HOMEPAGE_PATH)
        for marker in (
            "Apps, not just CLI",
            "macOS, Windows, iOS, and Android downloads",
            "## Desktop and mobile apps",
            "$1/month Commercial License",
        ):
            with self.subTest(marker=marker):
                self.assertIn(marker, text)

    def test_commercial_page_states_open_downloads_and_paid_optimizer(self) -> None:
        text = read(COMMERCIAL_PATH)
        for marker in (
            "Desktop and mobile app shells are planned as free downloads.",
            "$1/month Commercial License",
            "What downloads are free",
            "What the $1 license unlocks",
            "Commercial app surfaces",
            "Distribution and payment status",
            "planned as direct website downloads",
            "planned as a free App Store companion app",
            "planned as a free Play Store companion app",
            "Payment integration is not live yet.",
            "not yet release-ready for direct public store rollout",
            "Why people keep renewing each month",
            "provider headroom preserved",
        ):
            with self.subTest(marker=marker):
                self.assertIn(marker, text)

    def test_install_page_points_to_non_cli_apps(self) -> None:
        text = read(INSTALL_PATH)
        self.assertIn("commercial desktop and mobile apps should be free downloads", text)
        self.assertIn("$1/month Commercial License", text)


if __name__ == "__main__":
    unittest.main()
