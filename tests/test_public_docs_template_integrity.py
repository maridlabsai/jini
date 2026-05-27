from __future__ import annotations

import unittest
from pathlib import Path


LAYOUT_PATH = Path(__file__).resolve().parents[1] / "docs" / "_layouts" / "default.html"


class PublicDocsTemplateIntegrityTests(unittest.TestCase):
    def test_default_layout_escapes_frontmatter_driven_copy(self) -> None:
        text = LAYOUT_PATH.read_text(encoding="utf-8")
        for marker in (
            "{{ page.title | escape }}",
            "{{ site.title | escape }}",
            "{{ page.description | default: site.description | escape }}",
            '{{ page.eyebrow | default: "Jini" | escape }}',
            "{{ page.context_line | escape }}",
            "{{ item | escape }}",
            "{{ link.label | escape }}",
        ):
            with self.subTest(marker=marker):
                self.assertIn(marker, text)

    def test_default_layout_uses_structured_page_intro_regions(self) -> None:
        text = LAYOUT_PATH.read_text(encoding="utf-8")
        for marker in (
            '<div class="page-intro-main">',
            '<div class="page-intro-aside">',
            '<div class="page-intro-highlights" aria-label="Page highlights">',
            '<div class="page-intro-links">',
        ):
            with self.subTest(marker=marker):
                self.assertIn(marker, text)


if __name__ == "__main__":
    unittest.main()
