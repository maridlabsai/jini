from __future__ import annotations

import re
import unittest
from pathlib import Path


CSS_PATH = Path(__file__).resolve().parents[1] / "docs" / "assets" / "css" / "style.css"
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
        self.assertRegex(
            CSS_TEXT,
            r"@media screen and \(max-width:\s*960px\)\s*\{[\s\S]*?html\s*\{[^}]*scroll-padding-top:\s*1\.25rem;",
            "When the header stops being sticky on smaller widths, anchor jumps should use a much smaller top offset.",
        )
        self.assertRegex(
            CSS_TEXT,
            r"@media screen and \(max-width:\s*960px\)\s*\{[\s\S]*?main\s+\[id\]\s*\{[^}]*scroll-margin-top:\s*1\.25rem;",
            "Smaller-width fragment targets should not keep the oversized desktop scroll margin.",
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

    def test_ticket_like_ui_surfaces_wrap_and_shrink(self) -> None:
        self.assertRegex(
            CSS_TEXT,
            r"\.page-intro-highlights span,\s*\.page-intro-links a,\s*\.site-footer-links a,\s*\.pill-list span,\s*\.pill-list a,\s*\.compat-pill,\s*\.site-signal-pill,\s*\.offer-card-eyebrow,\s*\.offer-card-contexts span,\s*\.hero-scene-label,\s*\.workflow-card code,\s*\.media-overlay-tags span,\s*\.media-artifact-stack span,\s*\.truth-rule-pill\s*\{[^}]*max-width:\s*100%;[^}]*white-space:\s*normal;[^}]*overflow-wrap:\s*anywhere;[^}]*word-break:\s*break-word;",
            "Chip, pill, and ticket-like UI across the site should wrap safely instead of forcing horizontal overflow.",
        )
        self.assertRegex(
            CSS_TEXT,
            r"\.truth-rule-pill\s*\{[^}]*flex-wrap:\s*wrap;[^}]*justify-content:\s*flex-start;",
            "Long truth-rule pills should be allowed to wrap instead of overflowing their band.",
        )

    def test_dense_grid_children_can_shrink(self) -> None:
        self.assertRegex(
            CSS_TEXT,
            r"\.hero-scene > \*,\s*\.offer-grid > \*,\s*\.offer-side-rail > \*,[\s\S]*?\.proof-carousel-grid > \*[\s\S]*?\{\s*min-width:\s*0;",
            "Critical marketing grids should allow child cards to shrink without overflow.",
        )

    def test_homepage_hero_uses_two_column_desktop_balance(self) -> None:
        self.assertRegex(
            CSS_TEXT,
            r"\.hero-title\s*\{[^}]*max-width:\s*16ch;",
            "Homepage hero title should allow a broader measure before desktop grid kicks in.",
        )
        self.assertRegex(
            CSS_TEXT,
            r"@media screen and \(min-width:\s*980px\)\s*\{[\s\S]*?\.hero-panel-marketing\s*\{[^}]*display:\s*grid;[^}]*grid-template-columns:\s*minmax\(0,\s*1\.14fr\)\s*minmax\(320px,\s*0\.86fr\);",
            "Common desktop-width homepage hero should use a two-column layout instead of trapping the headline in a narrow single lane.",
        )
        self.assertRegex(
            CSS_TEXT,
            r"@media screen and \(min-width:\s*980px\)\s*\{[\s\S]*?\.hero-scene\s*\{[^}]*grid-column:\s*2;[^}]*grid-template-columns:\s*1fr;",
            "Common desktop-width homepage hero scene should move into the right column and stack safely there.",
        )

    def test_subpage_intro_uses_desktop_two_column_balance(self) -> None:
        self.assertRegex(
            CSS_TEXT,
            r"@media screen and \(min-width:\s*980px\)\s*\{[\s\S]*?\.page-intro-card\s*\{[^}]*display:\s*grid;[^}]*grid-template-columns:\s*minmax\(0,\s*1\.08fr\)\s*minmax\(280px,\s*0\.72fr\);",
            "Subpage intro bands should use a desktop two-column layout instead of leaving a large empty panel on the right.",
        )
        self.assertRegex(
            CSS_TEXT,
            r"\.page-intro-aside\s*\{[^}]*display:\s*grid;[^}]*gap:\s*0\.72rem;",
            "Subpage intro cards should use a dedicated aside container so the desktop right column is structurally real instead of relying on loose child placement.",
        )
        self.assertRegex(
            CSS_TEXT,
            r"@media screen and \(min-width:\s*980px\)\s*\{[\s\S]*?\.page-intro-aside\s*\{[^}]*grid-column:\s*2;",
            "Desktop subpage intro highlights and quick links should occupy the second column through the shared aside container.",
        )

    def test_section_card_headlines_use_a_broader_measure(self) -> None:
        self.assertRegex(
            CSS_TEXT,
            r"\.section-card h2\s*\{[^}]*max-width:\s*22ch;",
            "Shared section-card headlines should use a broader measure so desktop content cards do not collapse into narrow left-side stacks.",
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
