from __future__ import annotations

import re
import unittest
from pathlib import Path
from typing import Iterable


DOCS_DIR = Path(__file__).resolve().parents[1] / "docs"
PUBLIC_DOCS = sorted(DOCS_DIR.glob("*.md")) + sorted(DOCS_DIR.glob("*.html"))
SITE_TEMPLATES = sorted((DOCS_DIR / "_layouts").glob("*.html")) + sorted((DOCS_DIR / "_includes").glob("*.html"))
PUBLIC_LINK_SOURCES = PUBLIC_DOCS + SITE_TEMPLATES
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
        if path.is_file() and path.suffix == ".scss":
            mapping[f"/{path.relative_to(DOCS_DIR).with_suffix('.css').as_posix()}"] = path
    return mapping


TARGET_MAP = _doc_target_map()
LINK_RE = re.compile(r"\[[^\]]+\]\(([^)]+)\)|href=\"([^\"]+)\"|src=\"([^\"]+)\"")
LIQUID_RELATIVE_URL_RE = re.compile(
    r"""(?:href|src)=["']\{\{\s*'([^']+)'\s*(?:\|[^}]*)?relative_url\s*\}\}["']"""
)


def _frontmatter_lines(path: Path) -> list[str]:
    lines = path.read_text().splitlines()
    if lines[:1] != ["---"]:
        return []
    collected: list[str] = []
    for line in lines[1:]:
        if line == "---":
            break
        collected.append(line)
    return collected


def _quick_link_hrefs(path: Path) -> Iterable[str]:
    in_quick_links = False
    for line in _frontmatter_lines(path):
        if line.startswith("quick_links:"):
            in_quick_links = True
            continue
        if in_quick_links and re.match(r"^[A-Za-z0-9_]+:", line):
            break
        if in_quick_links:
            match = re.search(r"href:\s*(\S+)", line)
            if match:
                yield match.group(1).strip()


def _normalize_target(target: str) -> str:
    normalized = target.split("#", 1)[0].split("?", 1)[0]
    if normalized.startswith("./"):
        normalized = "/" + normalized[2:]
    return normalized


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
                normalized = _normalize_target(target)
                if not normalized:
                    continue
                if normalized not in TARGET_MAP:
                    unresolved.append(f"{path.name}: {normalized}")
        self.assertFalse(unresolved, "Unresolved local public-doc routes:\n" + "\n".join(unresolved))

    def test_static_liquid_relative_url_routes_resolve(self) -> None:
        unresolved: list[str] = []
        for path in PUBLIC_LINK_SOURCES:
            text = path.read_text()
            for target in LIQUID_RELATIVE_URL_RE.findall(text):
                normalized = _normalize_target(target)
                if not normalized:
                    continue
                if normalized not in TARGET_MAP:
                    unresolved.append(f"{path.relative_to(DOCS_DIR).as_posix()}: {normalized}")
        self.assertFalse(
            unresolved,
            "Unresolved Liquid relative_url routes in public docs chrome or bodies:\n" + "\n".join(unresolved),
        )

    def test_frontmatter_quick_links_use_site_relative_routes(self) -> None:
        offenders = []
        for path in DOCS_DIR.glob("*.md"):
            for href in _quick_link_hrefs(path):
                if not href.startswith("/"):
                    offenders.append(f"{path.name}: {href}")
        self.assertFalse(
            offenders,
            "Page quick_links should use site-relative routes:\n" + "\n".join(offenders),
        )

    def test_frontmatter_quick_links_resolve(self) -> None:
        unresolved = []
        for path in DOCS_DIR.glob("*.md"):
            for href in _quick_link_hrefs(path):
                if href not in TARGET_MAP:
                    unresolved.append(f"{path.name}: {href}")
        self.assertFalse(
            unresolved,
            "Page quick_links should resolve to public pages or assets:\n" + "\n".join(unresolved),
        )


if __name__ == "__main__":
    unittest.main()
