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
            "Use Jini free across providers now. Pay only when it proves it saves money.",
            "Jini is the free orchestration shell.",
            "The paid layer only starts to matter if Jini can measurably keep costs down or keep work moving.",
            "30-day free trial",
            "proof before payment",
            "paywall before downgrade",
            "## Free now. Paid later only if it earns it.",
            "The paid layer should not exist as a generic upgrade tax.",
        ):
            with self.subTest(marker=marker):
                self.assertIn(marker, text)

    def test_commercial_page_makes_free_downloads_and_paid_boundary_plain(self) -> None:
        text = read(COMMERCIAL_PATH)
        for marker in (
            "Jini should be easy to adopt and hard to overpay for.",
            "The shell stays free.",
            "Planned app downloads stay free.",
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
            "Public proof snapshot",
            "5 surfaces",
            "41% savings",
            "1 auto resume",
            "Public proof is acceptable only when it clearly marks preview posture and current release limits.",
            "before anyone is asked to pay or renew",
            "What the paid layer should prove before renewal",
            "Month-to-date savings",
            "Provider headroom preserved",
            "Throttles avoided or recovered",
            "Sessions resumed without babysitting",
        ):
            with self.subTest(marker=marker):
                self.assertIn(marker, text)


if __name__ == "__main__":
    unittest.main()
