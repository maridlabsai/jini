from __future__ import annotations

import unittest
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
DOC_PATH = ROOT / "specs" / "client-surfaces-and-free-tier.md"


class FreeTierValueSplitDocsTests(unittest.TestCase):
    def test_doc_spells_out_open_vs_paid_value_split(self) -> None:
        text = DOC_PATH.read_text(encoding="utf-8")
        for marker in (
            "## Adoption And Monetization Boundary",
            "free is manual proof and control",
            "paid is autonomous optimization and managed continuity",
            "basic throttle and route health visibility",
            "switch among configured tools, CLIs, providers, and local",
            "resume from saved state manually without transcript replay",
            "Commercial value is not a feature count",
            "## Open Version Value Proposition",
            "structural token savings",
            "context compaction",
            "artifact reuse instead of transcript replay",
            "offline-first continuation when the device can keep working locally",
            "explicit offline mode and reconciliation-debt visibility",
            "Claude Code, Codex, or",
            "GitHub CLI when the same Jini session goes offline",
            "## Upgrade Trigger",
            "Token Savings Ledger",
            "Throttle Radar",
            "Route Autopilot",
            "Auto Resume",
            "Cross-Platform Session Graph",
            "Managed Skill OS",
            "provider-specific optimization",
            "subscription-limit forecasting",
            "automatic fallback and resume",
            "changing the underlying platform or runtime target when throttle pressure",
            "changing the model or local profile for the task based on task shape",
            "automatic platform switching across managed routes when throttling or quotas",
            "task-shaped model and profile selection across local and managed routes",
            "online-capability detection, reconciliation scheduling, and managed recovery",
            "free structural efficiency patterns",
        ):
            with self.subTest(marker=marker):
                self.assertIn(marker, text)


if __name__ == "__main__":
    unittest.main()
