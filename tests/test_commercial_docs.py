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
            "{% for surface in site.data.public_surfaces.surfaces %}",
            "| {{ surface.name }} | {{ surface.badge }} | {{ surface.current_state }} |",
            "docs/_data/public_surfaces.json",
            "30-day free trial",
            "$1/month subscription",
            "What stays free vs what becomes paid",
        ):
            with self.subTest(marker=marker):
                self.assertIn(marker, text)

    def test_commercial_page_states_open_downloads_and_paid_optimizer(self) -> None:
        text = read(COMMERCIAL_PATH)
        for marker in (
            "The shell stays free.",
            "30-day free trial before asking for the $1/month subscription",
            "What the $1 license unlocks",
            "What stays free",
            "What downloads are free",
            "## The short version",
            "## Free shell vs paid optimizer",
            "| Need | Free shell | Paid optimizer |",
            "Predict provider limits before they block work",
            "What is not open source",
            "desktop/mobile app source code lives in the commercial repo only",
            "Distribution rule of thumb",
            "distribute directly from the website first when platform policy allows it",
            "When the paid layer earns the right to exist",
            "when the proof can be shown before payment, not explained after payment",
            "## Free app surfaces, when ready",
            "## Current readiness and payment status",
            "{% for surface in site.data.public_surfaces.surfaces %}",
            "| {{ surface.name }} | {{ surface.badge }} | {{ surface.current_state }} | {{ surface.activation }} |",
            "docs/_data/public_surfaces.json",
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
        self.assertIn("Commercial License starts with a 30-day free trial and becomes $1/month once checkout and entitlement activation are live", text)
        self.assertIn("<h2>What you get today</h2>", text)
        self.assertIn("{% for surface in site.data.public_surfaces.surfaces %}", text)
        self.assertIn("| {{ surface.name }} | {{ surface.current_state }} | {{ surface.next_step }} |", text)
        self.assertIn("docs/_data/public_surfaces.json", text)


if __name__ == "__main__":
    unittest.main()
