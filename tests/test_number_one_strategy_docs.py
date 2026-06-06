from __future__ import annotations

import unittest
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
RESEARCH_PATH = ROOT / "specs" / "number-one-product-research.md"
PRD_PATH = ROOT / "specs" / "number-one-platform-prd.md"
PLAN_PATH = ROOT / "specs" / "number-one-development-plan.md"
PLATFORM_OFFLINE_STRATEGY_PATH = ROOT / "specs" / "platform-offline-strategy.md"


def read(path: Path) -> str:
    return path.read_text(encoding="utf-8")


class NumberOneStrategyDocsTests(unittest.TestCase):
    maxDiff = None

    def test_strategy_docs_exist(self) -> None:
        self.assertTrue(RESEARCH_PATH.exists())
        self.assertTrue(PRD_PATH.exists())
        self.assertTrue(PLAN_PATH.exists())
        self.assertTrue(PLATFORM_OFFLINE_STRATEGY_PATH.exists())

    def test_research_uses_current_official_sources(self) -> None:
        text = read(RESEARCH_PATH)
        required_links = [
            "https://openai.com/index/introducing-codex/",
            "https://openai.com/index/work-with-codex-from-anywhere/",
            "https://docs.anthropic.com/en/docs/claude-code/sub-agents",
            "https://docs.anthropic.com/en/docs/claude-code/hooks",
            "https://docs.anthropic.com/en/docs/claude-code/memory",
            "https://docs.github.com/en/copilot/concepts/agents/copilot-cli/about-custom-agents",
            "https://docs.github.com/en/copilot/how-tos/copilot-cli/customize-copilot/add-skills",
            "https://docs.github.com/en/copilot/how-tos/copilot-cli/customize-copilot/add-mcp-servers",
            "https://docs.github.com/en/enterprise-server%403.19/github-cli/github-cli/using-github-cli-extensions",
            "https://kiro.dev/docs/steering/",
            "https://kiro.dev/docs/cli/hooks/",
            "https://github.com/google-gemini/gemini-cli/blob/main/README.md",
            "https://aider.chat/docs/repomap.html",
            "https://docs.cline.bot/features/plan-and-act",
            "https://block.github.io/goose/docs/getting-started/using-extensions/",
        ]
        for link in required_links:
            with self.subTest(link=link):
                self.assertIn(link, text)

    def test_research_covers_competitor_evolution_and_jini_wedge(self) -> None:
        text = read(RESEARCH_PATH)
        required_markers = [
            "## Competing Framework Evolution",
            "### OpenAI Codex",
            "### Claude Code",
            "### GitHub Copilot And GitHub CLI",
            "### Kiro",
            "### Gemini CLI",
            "### Aider, Cline, And Goose",
            "## What The Market Is Converging On",
            "## Where Competitors Are Still Weak",
            "## Jini Wedge",
            "## Self-Learning And Self-Correcting Requirement",
            "## Product Research Findings",
            "### Finding 1: Jini should be a work operating system, not a pack launcher",
            "### Finding 2: Quality must move upstream",
            "### Finding 5: Pace should come from policy and evaluation, not constant surface churn",
        ]
        for marker in required_markers:
            with self.subTest(marker=marker):
                self.assertIn(marker, text)

    def test_prd_defines_self_learning_cross_surface_product(self) -> None:
        text = read(PRD_PATH)
        required_markers = [
            "# Number One Platform PRD",
            "[jini-next-initiative-plan.md](./jini-next-initiative-plan.md)",
            "## Product Decision",
            "## Mission",
            "## Category Definition",
            "## Core Product Thesis",
            "## Product Requirements",
            "### 1. One Session Graph",
            "### 2. Environment Learning",
            "### 3. Workflow Learning",
            "### 4. Self-Correction Engine",
            "### 5. Upstream Quality Automation",
            "### 6. CLI Contract",
            "### 7. Desktop And Mobile Apps",
            "### 8. GitHub-Native System Of Record",
            "### 9. Frugal Route Policy",
            "### 10. Trust And Governance",
            "## Success Metrics",
            "## Score Exit Criteria",
            "`delivery-maturity >= 9.0`",
            "`memory-reliability >= 9.0`",
            "`adapter-portability >= 9.0`",
            "`token-efficiency >= 9.0`",
            "## Release Philosophy",
        ]
        for marker in required_markers:
            with self.subTest(marker=marker):
                self.assertIn(marker, text)

    def test_plan_links_workstreams_phases_and_score_lifts(self) -> None:
        text = read(PLAN_PATH)
        required_markers = [
            "# Number One Development Plan",
            "## Operating Cadence",
            "## Workstreams",
            "### Workstream A: Canonical Session And Artifact Graph",
            "### Workstream B: Environment And Workflow Learning",
            "### Workstream C: Self-Correction And Contract Repair",
            "### Workstream D: Upstream Quality Automation",
            "### Workstream E: CLI And App Surface Specialization",
            "### Workstream F: GitHub-Native System Of Record",
            "### Workstream G: Score And Competitive Operations",
            "## Phased Plan",
            "### Phase 0: Score Truth And Benchmark Control",
            "### Phase 1: Session Graph And Artifact-First Continuation",
            "### Phase 2: Workflow Learning And Upstream Quality",
            "### Phase 3: GitHub-Native Execution And Review",
            "### Phase 4: Self-Correcting Policy Engine",
            "### Phase 5: Score-Declare Win",
            "## Kill List",
            "## Sustainability Model",
            "## Final Exit Gate",
        ]
        for marker in required_markers:
            with self.subTest(marker=marker):
                self.assertIn(marker, text)

    def test_full_product_execution_plan_links_to_next_initiative_contract(self) -> None:
        text = read(ROOT / "specs" / "full-product-prd-execution-plan.md")
        self.assertIn(
            "[jini-next-initiative-plan.md](./jini-next-initiative-plan.md)",
            text,
        )

    def test_platform_offline_strategy_defines_form_factor_contracts(self) -> None:
        text = read(PLATFORM_OFFLINE_STRATEGY_PATH)
        required_markers = [
            "# Platform Offline Strategy",
            "## Cross-Platform Guarantees",
            "## macOS Strategy",
            "## Windows Strategy",
            "## Android Strategy",
            "## iOS Strategy",
            "### Offline Guarantees",
            "### Local Model Expectations",
            "### Sync Semantics",
            "### Route Policy",
            "## Sync Semantics",
            "## Route Policy",
            "## Shipping Prerequisites",
            "## Future Update Policy",
            "mobile is positioned as continuation and review",
            "desktop is positioned as the main offline authoring and artifact host",
            "future model updates flow through the registry, canary, promote, and",
        ]
        for marker in required_markers:
            with self.subTest(marker=marker):
                self.assertIn(marker, text)


if __name__ == "__main__":
    unittest.main()
