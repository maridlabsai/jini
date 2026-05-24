from __future__ import annotations

import unittest
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
HOMEPAGE_PATH = ROOT / "docs" / "index.md"
COMMERCIAL_PATH = ROOT / "docs" / "commercial.md"
PROOF_PATH = ROOT / "docs" / "proof.md"


def read(path: Path) -> str:
    return path.read_text(encoding="utf-8")


class PublicValueNarrativeDocsTests(unittest.TestCase):
    def test_homepage_leads_with_free_shell_and_paid_proof_split(self) -> None:
        text = read(HOMEPAGE_PATH)
        for marker in (
            "Turn messy AI work into something you can actually send.",
            "Jini turns rough notes, transcripts, screenshots, and drafts into follow-ups, readiness checks, and decision memos that survive handoff.",
            "The paid layer stays narrow: it only enters when Jini can prove it saved money or kept work moving.",
            "After the meeting",
            "Before the handoff",
            "Before the decision",
            "Use the raw shell for one-shot answers.",
            "Use the free Jini shell when the work has to survive handoff.",
            "Add the paid optimizer only when the proof can be measured.",
            "planned 30-day free trial",
            "The free shell should already be enough to finish serious work.",
        ):
            with self.subTest(marker=marker):
                self.assertIn(marker, text)

    def test_commercial_page_makes_free_downloads_and_paid_boundary_plain(self) -> None:
        text = read(COMMERCIAL_PATH)
        for marker in (
            "Jini should be easy to adopt and hard to overpay for.",
            "The shell stays free.",
            "App downloads stay free when each surface is live.",
            "paywall prompt should appear before downgrade",
            "## Free shell vs paid optimizer",
            "when the proof can be shown before payment, not explained after payment",
            "Do not charge for the shell.",
            "Do not charge for app downloads.",
            "Charge only for the cost-saver and continuity layer that proves its value at runtime.",
        ):
            with self.subTest(marker=marker):
                self.assertIn(marker, text)

    def test_proof_page_requires_paid_layer_to_earn_renewal_with_evidence(self) -> None:
        text = read(PROOF_PATH)
        for marker in (
            "First, the free shell should make its value obvious fast.",
            "{{ site.data.public_proof.hero.eyebrow }}",
            "{{ site.data.public_proof.hero.headline }}",
            "{{ site.data.public_proof.hero.body }}",
            "{{ card.value }}",
            "{{ site.data.public_proof.sections[2].bullets[0] }}",
            "before anyone is asked to pay or renew",
            "What the paid layer should prove before renewal",
            "Month-to-date savings",
            "Provider headroom preserved",
            "Throttles avoided or recovered",
            "Sessions resumed without babysitting",
            "How public proof should be fed",
            "docs/_data/public_proof.json",
        ):
            with self.subTest(marker=marker):
                self.assertIn(marker, text)


if __name__ == "__main__":
    unittest.main()
