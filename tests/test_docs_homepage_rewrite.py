from __future__ import annotations

import unittest
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
HOMEPAGE_PATH = ROOT / "docs" / "index.md"
LAYOUT_PATH = ROOT / "docs" / "_layouts" / "default.html"
PLAN_PATH = ROOT / "specs" / "docs-homepage-rewrite-plan.md"
TRUST_PATH = ROOT / "docs" / "proof.md"
OUTPUTS_PATH = ROOT / "docs" / "state-and-artifacts.md"


def read(path: Path) -> str:
    return path.read_text(encoding="utf-8")


class DocsHomepageRewriteTests(unittest.TestCase):
    def test_homepage_keeps_clear_install_compatibility_and_trust_markers(self) -> None:
        text = read(HOMEPAGE_PATH)
        for marker in (
            "Free shell for AI work",
            "Use Jini free across providers now. Pay only when it proves it saves money.",
            "The core shell stays open.",
            "Works with Claude Code",
            "Works with Codex",
            "If you only need a one-off answer, use a raw model shell.",
            "Use a raw model shell when",
            "Use Jini when",
            "Free orchestration core",
            "Cross-provider by default",
            "Pay only for proof",
            "## Pricing promise",
            "proof before payment",
            "paywall before downgrade",
            "## Who Jini is for",
            "Good fit",
            "Not the best fit",
            "## Free now. Paid later only if it earns it.",
            "What stays free",
            "What becomes paid",
            "## What Jini writes",
            "## Start anywhere. Resume anywhere.",
            "One session, not four different products",
            "macOS, Windows, mobile, and CLI",
            "## Cross-surface rollout",
            "## Commands that matter",
            "jini commands",
            "Stored locally",
            "Not stored as product magic",
            "## What stays free vs what becomes paid",
            "## See the product surface",
            "Install and first run",
            "Metrics and route evidence",
            "Current example capture. No live checkout or signed desktop-release claim.",
            "Current example capture. Route evidence shown here is interface proof, not live paid-savings telemetry.",
            "Sendable follow-up",
            "Build-readiness check",
            "Recommendation memo",
            "Closure checklist",
            "Example artifact from the public meeting-follow-up scenario.",
            "Example artifact from the public spec-readiness scenario.",
            "Example artifact from the public vendor-choice scenario.",
            "Example artifact from the public incident-closure scenario.",
            "free app downloads once each surface is live",
            "jini metrics",
        ):
            with self.subTest(marker=marker):
                self.assertIn(marker, text)

    def test_layout_uses_product_facing_nav_labels(self) -> None:
        text = read(LAYOUT_PATH)
        for marker in (
            "Quickstart",
            "Outputs",
            "Proof",
            "Command Catalog",
            "Free shell now. Paid only if it earns it.",
            "30-day trial",
            "$1 after proof",
        ):
            with self.subTest(marker=marker):
                self.assertIn(marker, text)

    def test_plan_documents_scope_and_nuget_decision(self) -> None:
        text = read(PLAN_PATH)
        for marker in (
            "Docs Homepage Rewrite Plan",
            "Use the Buddy site as a structure reference",
            "NuGet is not part of this first pass",
            "Homebrew",
            "pipx / PyPI",
            "winget / scoop",
        ):
            with self.subTest(marker=marker):
                self.assertIn(marker, text)

    def test_supporting_pages_use_buyability_language(self) -> None:
        trust = read(TRUST_PATH)
        outputs = read(OUTPUTS_PATH)
        for marker in (
            "free shell should make its value obvious",
            "What a buyer should be able to verify quickly",
            "Inspectability instead of product magic",
            "What the paid layer should prove before renewal",
        ):
            with self.subTest(marker=marker):
                self.assertIn(marker, trust)
        for marker in (
            "Outputs are part of the buyability story",
            "deliverables, continuation, and explicit risk",
            "It should open deliverables, not storage concepts.",
            "No storage-first labels",
            "macOS, Windows, mobile, and CLI",
        ):
            with self.subTest(marker=marker):
                self.assertIn(marker, outputs)


if __name__ == "__main__":
    unittest.main()
