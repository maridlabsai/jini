from __future__ import annotations

import unittest
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
DOCS_DIR = ROOT / "docs"


class PublicDocsRenderSafetyTests(unittest.TestCase):
    def test_public_docs_do_not_use_markdown_wrappers(self) -> None:
        offenders: list[str] = []
        for path in sorted(DOCS_DIR.glob("*.md")):
            text = path.read_text(encoding="utf-8")
            if 'markdown="1"' in text or "markdown=1" in text:
                offenders.append(path.name)

        self.assertEqual(
            offenders,
            [],
            f"Public docs should use explicit HTML or plain Markdown, not markdown wrappers: {offenders}",
        )


if __name__ == "__main__":
    unittest.main()
