from __future__ import annotations

import unittest
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
ENGINEERING_PATH = ROOT / "specs" / "engineering-principles.md"
DOCTRINE_PATH = ROOT / "specs" / "lean-platform-doctrine.md"
SYSTEM_DESIGN_PATH = ROOT / "specs" / "cross-surface-session-system-and-dev-design.md"


def read(path: Path) -> str:
    return path.read_text(encoding="utf-8")


class EngineeringPrinciplesDocsTests(unittest.TestCase):
    def test_engineering_principles_doc_exists_and_has_required_markers(self) -> None:
        text = read(ENGINEERING_PATH)
        for marker in (
            "# Engineering Principles",
            "## SOLID Rules",
            "Single Responsibility Principle",
            "Dependency Inversion Principle",
            "## OOP Rules",
            "## Preferred Design Patterns",
            "Adapter",
            "Strategy",
            "Factory",
            "Facade",
            "Value Object",
            "## Reject Conditions",
        ):
            with self.subTest(marker=marker):
                self.assertIn(marker, text)

    def test_doctrine_and_system_design_anchor_engineering_rules(self) -> None:
        doctrine = read(DOCTRINE_PATH)
        design = read(SYSTEM_DESIGN_PATH)
        for marker in (
            "### Engineering Discipline",
            "implementation should follow SOLID by default",
            "prefer composition over inheritance",
        ):
            with self.subTest(marker=marker):
                self.assertIn(marker, doctrine)
        for marker in (
            "[engineering-principles.md](./engineering-principles.md)",
            "## Developer Structure Rules",
            "routing uses strategy-style policy objects",
            "surface integration stays behind adapter-style boundaries",
            "bundle and manifest construction uses factory-style builders",
            "facade-style orchestration",
        ):
            with self.subTest(marker=marker):
                self.assertIn(marker, design)


if __name__ == "__main__":
    unittest.main()
