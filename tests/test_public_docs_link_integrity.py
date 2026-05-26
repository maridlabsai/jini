from __future__ import annotations

import re
import unittest
from pathlib import Path


DOCS_DIR = Path(__file__).resolve().parents[1] / "docs"
PUBLIC_DOCS = sorted(DOCS_DIR.glob("*.md")) + sorted(DOCS_DIR.glob("*.html"))
STATIC_ASSET_SUFFIXES = {
    ".css",
    ".js",
    ".json",
    ".png",
    ".jpg",
    ".jpeg",
    ".svg",
    ".ico",
    ".webp",
}
ROOT_TARGETS = {
    "/": DOCS_DIR / "index.md",
}


def _doc_target_map() -> dict[str, Path]:
    mapping = dict(ROOT_TARGETS)
    for path in DOCS_DIR.glob("*.md"):
        if path.stem == "index":
            continue
        mapping[f"/{path.stem}.html"] = path
    for path in DOCS_DIR.rglob("*"):
        if path.is_file() and path.suffix in STATIC_ASSET_SUFFIXES:
            mapping[f"/{path.relative_to(DOCS_DIR).as_posix()}"] = path
    return mapping


TARGET_MAP = _doc_target_map()
LINK_RE = re.compile(r"\[[^\]]+\]\(([^)]+)\)|href=\"([^\"]+)\"|src=\"([^\"]+)\"")


class PublicDocsLinkIntegrityTests(unittest.TestCase):
    def test_public_pages_do_not_use_page_relative_html_routes(self) -> None:
        offenders = []
        for path in PUBLIC_DOCS:
            text = path.read_text()
            if re.search(r'href="\./[^"]+\.html"', text):
                offenders.append(path.name)
        self.assertFalse(
            offenders,
            f"Public pages should use site-relative routes for internal HTML links: {', '.join(offenders)}",
        )

    def test_static_local_routes_resolve(self) -> None:
        unresolved: list[str] = []
        for path in PUBLIC_DOCS:
            text = path.read_text()
            for match in LINK_RE.finditer(text):
                target = next(group for group in match.groups() if group)
                if not target or target.startswith(("http://", "https://", "mailto:", "tel:", "#")):
                    continue
                if "{{" in target or "{%" in target:
                    continue
                normalized = target.split("#", 1)[0]
                if not normalized:
                    continue
                if normalized.startswith("./"):
                    normalized = "/" + normalized[2:]
                if normalized not in TARGET_MAP:
                    unresolved.append(f"{path.name}: {normalized}")
        self.assertFalse(unresolved, "Unresolved local public-doc routes:\n" + "\n".join(unresolved))


if __name__ == "__main__":
    unittest.main()
