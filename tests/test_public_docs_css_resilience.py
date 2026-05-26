from __future__ import annotations

import re
import unittest
from pathlib import Path


CSS_PATH = Path(__file__).resolve().parents[1] / "docs" / "assets" / "css" / "style.scss"
CSS_TEXT = CSS_PATH.read_text()


class PublicDocsCssResilienceTests(unittest.TestCase):
    def test_fragment_targets_account_for_sticky_header(self) -> None:
        self.assertRegex(
            CSS_TEXT,
            r"html\s*\{[^}]*scroll-padding-top:\s*7\.6rem;",
            "The site should reserve top scroll padding for the sticky header.",
        )
        self.assertRegex(
            CSS_TEXT,
            r"main\s+\[id\]\s*\{[^}]*scroll-margin-top:\s*7\.6rem;",
            "Anchored sections should offset for the sticky header.",
        )

    def test_command_and_code_surfaces_wrap_long_content(self) -> None:
        self.assertRegex(
            CSS_TEXT,
            r"\.hero-install-main code\s*\{[^}]*overflow-wrap:\s*anywhere;[^}]*word-break:\s*break-word;",
            "Install commands should wrap instead of overflowing narrow screens.",
        )
        self.assertRegex(
            CSS_TEXT,
            r"\.markdown-body pre code\s*\{[^}]*overflow-wrap:\s*anywhere;[^}]*word-break:\s*break-word;",
            "Rendered code blocks should wrap long tokens safely.",
        )
        self.assertRegex(
            CSS_TEXT,
            r"\.shell-panel pre\s*\{[^}]*overflow-wrap:\s*anywhere;[^}]*word-break:\s*break-word;",
            "Dark shell examples should wrap long tokens safely.",
        )

    def test_dense_grid_children_can_shrink(self) -> None:
        self.assertRegex(
            CSS_TEXT,
            r"\.hero-scene > \*,\s*\.offer-grid > \*,\s*\.offer-side-rail > \*,[\s\S]*?\.proof-carousel-grid > \*[\s\S]*?\{\s*min-width:\s*0;",
            "Critical marketing grids should allow child cards to shrink without overflow.",
        )

    def test_scrollable_surfaces_keep_touch_overflow_support(self) -> None:
        self.assertRegex(
            CSS_TEXT,
            r"\.surface-story table,\s*\.economics-table table\s*\{[^}]*overflow-x:\s*auto;[^}]*-webkit-overflow-scrolling:\s*touch;",
            "HTML tables should stay scrollable on narrow touch screens.",
        )
        self.assertRegex(
            CSS_TEXT,
            r"\.markdown-body pre\s*\{[^}]*overflow-x:\s*auto;[^}]*-webkit-overflow-scrolling:\s*touch;",
            "Generic pre blocks should preserve horizontal scrolling on touch devices.",
        )


if __name__ == "__main__":
    unittest.main()
