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
            "Free orchestration core",
            "Only the CLI is installable now",
            "## Free app surfaces, when ready",
            "## Current availability",
            "Desktop and Android should distribute directly first where policy allows",
            "Commercial License checkout | Planned. Not live yet",
            "$1/month Commercial License",
            "What stays free vs what becomes paid",
        ):
            with self.subTest(marker=marker):
                self.assertIn(marker, text)

    def test_commercial_page_states_open_downloads_and_paid_optimizer(self) -> None:
        text = read(COMMERCIAL_PATH)
        for marker in (
            "The shell stays free.",
            "What the $1 license unlocks",
            "What stays free",
            "What downloads are free",
            "## The short version",
            "What is not open source",
            "desktop/mobile app source code lives in the commercial repo only",
            "Distribution rule of thumb",
            "distribute directly from the website first when platform policy allows it",
            "## Free app surfaces, when ready",
            "## Current readiness and payment status",
            "| macOS app shell | Preview only. Not downloadable yet | Buy on the website, then sign in |",
            "| iOS companion app | Preview only. Not on the App Store yet | Sign in with an existing paid account |",
            "| Android companion app | Preview only. Not downloadable yet. Direct-first when policy allows, with Play Store secondary | Sign in with an existing paid account |",
            "| Commercial License checkout | Planned. Not live yet | Website checkout + account entitlement |",
            "Payment integration is not live yet.",
            "not yet release-ready for direct public download or store rollout",
            "What the paid layer must prove before renewal",
            "month-to-date token savings",
            "provider headroom preserved",
        ):
            with self.subTest(marker=marker):
                self.assertIn(marker, text)

    def test_install_page_points_to_non_cli_apps(self) -> None:
        text = read(INSTALL_PATH)
        self.assertIn("Today the CLI is the only installable surface.", text)
        self.assertIn("preview-only, not publicly downloadable yet", text)
        self.assertIn("Desktop and Android should distribute directly first where policy allows", text)
        self.assertIn("Commercial License is $1/month once checkout and entitlement activation are live", text)
        self.assertIn("<h2>What you get today</h2>", text)
        self.assertIn("| CLI | Installable now on macOS and Linux |", text)
        self.assertIn("| macOS app shell | Preview only. Not downloadable yet |", text)
        self.assertIn("| Commercial License | Planned. Not live yet |", text)


if __name__ == "__main__":
    unittest.main()
