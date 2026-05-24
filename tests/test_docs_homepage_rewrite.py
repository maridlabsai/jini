from __future__ import annotations

import unittest
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
HOMEPAGE_PATH = ROOT / "docs" / "index.md"
LAYOUT_PATH = ROOT / "docs" / "_layouts" / "default.html"
PLAN_PATH = ROOT / "specs" / "docs-homepage-rewrite-plan.md"
TRUST_PATH = ROOT / "docs" / "proof.md"
OUTPUTS_PATH = ROOT / "docs" / "state-and-artifacts.md"
SHOWCASE_DATA_PATH = ROOT / "docs" / "_data" / "showcase_media.json"


def read(path: Path) -> str:
    return path.read_text(encoding="utf-8")


class DocsHomepageRewriteTests(unittest.TestCase):
    def test_homepage_keeps_clear_install_compatibility_and_trust_markers(self) -> None:
        text = read(HOMEPAGE_PATH)
        for marker in (
            "AI work that has to survive week two",
            "Turn messy AI work into something you can actually send.",
            "The core shell stays open.",
            "Works with Claude Code",
            "Works with Codex",
            "CLI thread",
            "Desktop and mobile continuity",
            "Sendable output",
            "If you only need a one-off answer, use a raw model shell.",
            "Choose the lightest layer that still leaves behind usable work.",
            "After the meeting",
            "Before the handoff",
            "Before the decision",
            "Free orchestration core",
            "Jini earns the right to exist when the work has to leave chat with a sendable artifact, a safer handoff, and reasoning you can still explain later.",
            "Use the raw shell for one-shot answers.",
            "Use the free Jini shell when the work has to survive handoff.",
            "Add the paid optimizer only when the proof can be measured.",
            "Pay only for proof",
            "## What Jini leaves behind",
            "One session, not four different products",
            "macOS, Windows, mobile, and CLI",
            "## Where you can use it now",
            "## Small front door. Clear paid boundary.",
            "jini commands",
            "## Proof, kept brief",
            "Trust should support the pitch, not bury it.",
            "Free first. Paid only if the savings story is measurable.",
            "No live store claims before release. No fake telemetry. No hidden preview posture.",
            "Stored locally, not as product magic",
            "## See the product surface",
            "/assets/story/jini-showcase-strip.svg",
            "/assets/story/jini-output-strip.svg",
            "checked-in storyboard illustration built from the current public product posture",
            "second checked-in storyboard illustration built from the current public examples",
            "site.data.showcase_media.product_surface_cards",
            "site.data.showcase_media.output_cards",
            "site.data.proof_carousel.slides",
            "capture id and source",
            "checked-in proof carousel slides from the repo",
            "Start in the CLI today, then carry the same work forward as desktop and mobile come online.",
            "jini metrics",
        ):
            with self.subTest(marker=marker):
                self.assertIn(marker, text)

    def test_homepage_merges_old_standalone_sections_into_fewer_stops(self) -> None:
        text = read(HOMEPAGE_PATH)
        for marker in (
            "## Quickstart",
            "## One thread across surfaces",
            "## Why teams keep Jini around",
            "## See real outputs",
            "Three flagship jobs",
            "## Three ways to work",
            "Free now. Paid later only if it earns it.",
            "What stays free",
            "What becomes paid",
            "Good fit",
            "Not the best fit",
        ):
            with self.subTest(marker=marker):
                self.assertNotIn(marker, text)

    def test_showcase_data_carries_homepage_media_copy_and_truth_notes(self) -> None:
        text = read(SHOWCASE_DATA_PATH)
        for marker in (
            "Install and first run",
            "Metrics and route evidence",
            "Sendable follow-up",
            "Build-readiness check",
            "Recommendation memo",
            "Closure checklist",
            "Checked-in public illustration. Install posture only, with no live checkout or signed desktop-release claim.",
            "Checked-in public illustration. Session state and artifact posture only, not a public app-download claim.",
            "Checked-in public illustration. Route evidence shown here is product posture only, not live paid-savings telemetry.",
            "Checked-in public illustration. Decision-output posture only, not a live provider comparison or billing panel.",
            "Checked-in public illustration. Follow-up-output posture only, not a live inbox, mail send, or delivery screenshot.",
            "Checked-in public illustration. Readiness-output posture only, not a live issue tracker, PR gate, or deployment screen.",
            "Example artifact from the public incident-closure scenario.",
            "capture-install-panel",
            "capture-followup-panel",
            "capture-state-panel",
            "capture-readiness-panel",
            "capture-route-evidence-panel",
            "capture-decision-panel",
            "capture-incident-response",
            "current-public-example",
            "checked-in-public-illustration",
            "assets/story/jini-followup-panel.svg",
            "assets/story/jini-install-panel.svg",
            "assets/story/jini-readiness-panel.svg",
            "assets/story/jini-state-panel.svg",
            "assets/story/jini-route-evidence-panel.svg",
            "assets/story/jini-decision-panel.svg",
            "examples/incident-closure",
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
