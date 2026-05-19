from __future__ import annotations

import unittest
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
FRAMEWORK_PATH = ROOT / "specs" / "adaptive-response-rendering-framework.md"
REVIEW_PATH = ROOT / "specs" / "adaptive-response-rendering-framework-review.md"
GATE_PATH = ROOT / "specs" / "adaptive-response-rendering-framework-gate.md"
PUBLISH_TOOL_PATH = ROOT / "tools" / "jini_validate.py"
PRODUCT_CONTRACT_PATH = ROOT / "specs" / "product-rewrite-contract.md"


def read(path: Path) -> str:
    return path.read_text(encoding="utf-8")


class AdaptiveResponseDocsTests(unittest.TestCase):
    def test_framework_docs_exist(self) -> None:
        for path in (FRAMEWORK_PATH, REVIEW_PATH, GATE_PATH):
            with self.subTest(path=path):
                self.assertTrue(path.exists(), f"Missing required doc: {path}")

    def test_framework_has_required_architecture(self) -> None:
        text = read(FRAMEWORK_PATH)
        required = [
            "# Adaptive Response Rendering Framework",
            "## Product Rule",
            "## Open Framework Lessons",
            "### 1. Semantic Envelope",
            "### 2. Artifact Envelope",
            "### 3. Render Request",
            "### 4. Surface Renderer",
            "### Projector And Render Policy Split",
            "## Response Modes",
            "## Surface Guidance",
            "## Testing Strategy",
            "## Migration Plan",
            "## Gate Summary",
        ]
        for marker in required:
            with self.subTest(marker=marker):
                self.assertIn(marker, text)

    def test_framework_uses_official_open_framework_references(self) -> None:
        text = read(FRAMEWORK_PATH)
        required_links = [
            "https://github.com/langchain-ai/langgraph",
            "https://ai.pydantic.dev/",
            "https://github.com/All-Hands-AI/OpenHands",
            "https://docs.continue.dev/",
            "https://mastra.ai/docs",
            "https://docs.ag-ui.com/",
        ]
        for link in required_links:
            with self.subTest(link=link):
                self.assertIn(link, text)

    def test_framework_blocks_hard_coded_response_molds(self) -> None:
        text = read(FRAMEWORK_PATH)
        required = [
            "Jini must harden the semantic contract, not the sentence shape.",
            "tests assert this envelope first",
            "prose can vary if the envelope remains correct",
            "use-case profiles may configure artifact families",
            "core renderer remains generic",
            "test meaning, not prose",
            "The projector owns truth. The render policy owns emphasis.",
        ]
        for marker in required:
            with self.subTest(marker=marker):
                self.assertIn(marker, text)

    def test_gate_has_required_reject_conditions(self) -> None:
        gate = read(GATE_PATH)
        required = [
            "### 2. Adaptive Rendering",
            "greeting-only input does not create work",
            "### 4. Testability",
            "### 5. No Core Hard Coding",
            "core rendering branches grow one hard-coded format per use case",
            "## Required Regression Inputs",
            "semantic envelope tests",
            "publish-readiness pass",
        ]
        for marker in required:
            with self.subTest(marker=marker):
                self.assertIn(marker, gate)

    def test_review_records_expert_critique_and_pass(self) -> None:
        review = read(REVIEW_PATH)
        required = [
            "## Review Personas",
            "### Systems Architect",
            "### Application Developer",
            "### Test Engineer",
            "### UX Researcher",
            "### Product Designer",
            "### Product And Market Critic",
            "## Revisions Applied",
            "## Rationalized Position",
            "`PASS`",
        ]
        for marker in required:
            with self.subTest(marker=marker):
                self.assertIn(marker, review)

    def test_publish_readiness_and_contract_include_framework(self) -> None:
        publish_tool = read(PUBLISH_TOOL_PATH)
        product_contract = read(PRODUCT_CONTRACT_PATH)
        self.assertIn("adaptive-response-rendering-framework.md", publish_tool)
        self.assertIn("adaptive-response-rendering-framework.md", product_contract)


if __name__ == "__main__":
    unittest.main()
