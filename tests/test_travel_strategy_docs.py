from __future__ import annotations

import unittest
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
FRAMEWORK_PATH = ROOT / "specs" / "travel-curated-experience-framework.md"
REVIEW_PATH = ROOT / "specs" / "travel-curated-experience-framework-review.md"
GATE_PATH = ROOT / "specs" / "travel-curated-experience-framework-gate.md"


def read(path: Path) -> str:
    return path.read_text(encoding="utf-8")


class TravelStrategyDocsTests(unittest.TestCase):
    def test_framework_docs_exist(self) -> None:
        for path in (FRAMEWORK_PATH, REVIEW_PATH, GATE_PATH):
            with self.subTest(path=path):
                self.assertTrue(path.exists(), f"Missing required doc: {path}")

    def test_framework_has_required_sections(self) -> None:
        text = read(FRAMEWORK_PATH)
        required = [
            "# Travel Curated Experience Framework",
            "## Benchmark Lessons",
            "## Public Experience Layers",
            "### 1. Scoped Travel Brief",
            "### 2. Curated Option Set",
            "### 3. Day-By-Day Trip Object",
            "### 6. Confirmation-First Trust",
            "## Shipping Now Vs Later",
            "## Shared Invariants",
            "## Relationship To Commercial Repo",
        ]
        for marker in required:
            with self.subTest(marker=marker):
                self.assertIn(marker, text)

    def test_framework_uses_official_benchmark_inputs(self) -> None:
        text = read(FRAMEWORK_PATH)
        self.assertIn("https://layla.ai/about", text)
        self.assertIn("https://navan.com/product/navan-edge", text)
        self.assertIn("https://navan.com/blog/introducing-navan-edge", text)
        self.assertIn("Layla-like curation", text)
        self.assertIn("Navan-like confirmation", text)

    def test_review_and_gate_are_pass_shaped(self) -> None:
        review = read(REVIEW_PATH)
        gate = read(GATE_PATH)
        self.assertIn("## Review Personas", review)
        self.assertIn("## Round 1 Findings", review)
        self.assertIn("## Revisions Applied", review)
        self.assertIn("## Rationalized Position", review)
        self.assertIn("## Final Verdict", review)
        self.assertIn("`PASS`", review)
        self.assertIn("# Travel Curated Experience Framework Gate", gate)
        self.assertIn("## Gate Categories", gate)
        self.assertIn("### 5. Competitive Realism", gate)
        self.assertIn("## Reject Conditions", gate)


if __name__ == "__main__":
    unittest.main()
