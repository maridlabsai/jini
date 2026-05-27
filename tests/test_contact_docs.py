from __future__ import annotations

import unittest
from pathlib import Path


CONTACT_DOC = Path(__file__).resolve().parents[1] / "docs" / "contact.md"


class ContactDocsTests(unittest.TestCase):
    def test_contact_page_exposes_clickable_commercial_contact(self) -> None:
        text = CONTACT_DOC.read_text(encoding="utf-8")
        self.assertIn(
            'href="mailto:maridlabsai@gmail.com"',
            text,
            "The contact page should expose a clickable commercial contact route.",
        )
        self.assertIn(
            '<a href="https://github.com/maridlabsai/jini/issues">Issues</a>',
            text,
            "The contact page should keep the public Issues path explicit.",
        )
        self.assertIn(
            '<a href="https://github.com/maridlabsai/jini/discussions">Discussions</a>',
            text,
            "The contact page should keep the public Discussions path explicit.",
        )


if __name__ == "__main__":
    unittest.main()
