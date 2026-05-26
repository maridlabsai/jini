from __future__ import annotations

import sys
import unittest
from pathlib import Path

REPO_ROOT = Path(__file__).resolve().parents[1]
if str(REPO_ROOT) not in sys.path:
    sys.path.insert(0, str(REPO_ROOT))

from tools.yaml_compat import safe_load


DOCS_DIR = REPO_ROOT / "docs"
PUBLIC_MARKDOWN_DOCS = sorted(DOCS_DIR.glob("*.md"))


def _frontmatter_text(path: Path) -> str | None:
    lines = path.read_text(encoding="utf-8").splitlines()
    if lines[:1] != ["---"]:
        return None
    collected: list[str] = []
    for line in lines[1:]:
        if line == "---":
            return "\n".join(collected)
        collected.append(line)
    return None


class PublicDocsFrontmatterIntegrityTests(unittest.TestCase):
    def test_public_markdown_frontmatter_is_parseable_yaml(self) -> None:
        for path in PUBLIC_MARKDOWN_DOCS:
            with self.subTest(page=path.name):
                frontmatter = _frontmatter_text(path)
                if frontmatter is None:
                    continue
                parsed = safe_load(frontmatter or "")
                self.assertIsInstance(parsed, dict, f"{path.name} frontmatter should parse into a mapping")


if __name__ == "__main__":
    unittest.main()
