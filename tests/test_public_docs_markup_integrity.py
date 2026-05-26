from __future__ import annotations

import re
import unittest
from pathlib import Path


DOCS_DIR = Path(__file__).resolve().parents[1] / "docs"
PUBLIC_PAGES = sorted(DOCS_DIR.glob("*.md"))

BLOCK_OPEN_RE = re.compile(
    r"<(div|section|article|aside|header|footer|main|nav|table|thead|tbody|tr|td|th|ul|ol|li|p|h[1-6]|pre|code)(\s|>)",
    re.IGNORECASE,
)
BLOCK_CLOSE_RE = re.compile(
    r"</(div|section|article|aside|header|footer|main|nav|table|thead|tbody|tr|td|th|ul|ol|li|p|h[1-6]|pre|code)>",
    re.IGNORECASE,
)
MD_BLOCK_RE = re.compile(r"^(\s*)([-*] |\d+\. |##+ |\| |```|> )")
TAG_RE = re.compile(r"<(/?)([a-zA-Z0-9:-]+)(?:\s[^<>]*?)?(\/?)>")
VOID_TAGS = {"area", "base", "br", "col", "hr", "img", "input", "link", "meta", "source", "wbr"}


def _content_lines(path: Path) -> list[str]:
    lines = path.read_text().splitlines()
    if lines[:1] != ["---"]:
        return lines
    for i in range(1, len(lines)):
        if lines[i] == "---":
            return lines[i + 1 :]
    return lines


class PublicDocsMarkupIntegrityTests(unittest.TestCase):
    def test_public_pages_do_not_use_markdown_wrappers(self) -> None:
        for path in PUBLIC_PAGES:
            text = path.read_text()
            self.assertNotIn('markdown="1"', text, f"{path.name} should not rely on markdown wrappers")
            self.assertNotIn("markdown=1", text, f"{path.name} should not rely on markdown wrappers")

    def test_public_pages_do_not_mix_markdown_blocks_inside_html_blocks(self) -> None:
        for path in PUBLIC_PAGES:
            depth = 0
            pre_depth = 0
            findings: list[str] = []
            for line_no, line in enumerate(_content_lines(path), 1):
                opens = [tag.lower() for tag, _ in BLOCK_OPEN_RE.findall(line)]
                closes = [tag.lower() for tag in BLOCK_CLOSE_RE.findall(line)]
                if depth > 0 and pre_depth == 0 and MD_BLOCK_RE.search(line):
                    findings.append(f"{line_no}: {line}")
                depth += len(opens) - len(closes)
                pre_depth += opens.count("pre") + opens.count("code")
                pre_depth -= closes.count("pre") + closes.count("code")
                depth = max(depth, 0)
                pre_depth = max(pre_depth, 0)
            self.assertFalse(
                findings,
                f"{path.name} should not place Markdown block syntax inside raw HTML blocks:\n" + "\n".join(findings),
            )

    def test_public_pages_have_balanced_block_html(self) -> None:
        for path in PUBLIC_PAGES:
            stack: list[tuple[str, int]] = []
            issues: list[str] = []
            for line_no, line in enumerate(_content_lines(path), 1):
                for slash, tag, selfclose in TAG_RE.findall(line):
                    tag = tag.lower()
                    if tag in VOID_TAGS or selfclose:
                        continue
                    if slash:
                        if stack and stack[-1][0] == tag:
                            stack.pop()
                            continue
                        idx = next((i for i in range(len(stack) - 1, -1, -1) if stack[i][0] == tag), None)
                        if idx is None:
                            issues.append(f"unmatched </{tag}> at line {line_no}")
                        else:
                            issues.append(f"crossed </{tag}> at line {line_no}")
                            stack = stack[:idx]
                    else:
                        stack.append((tag, line_no))
            issues.extend(f"unclosed <{tag}> from line {line_no}" for tag, line_no in stack)
            self.assertFalse(issues, f"{path.name} has malformed block HTML:\n" + "\n".join(issues))

    def test_public_pre_blocks_wrap_code_tags(self) -> None:
        bad_pages = []
        for path in PUBLIC_PAGES:
            text = path.read_text()
            if re.search(r"<pre>(?!<code)", text):
                bad_pages.append(path.name)
        self.assertFalse(bad_pages, f"Public pages should use <pre><code> consistently: {', '.join(bad_pages)}")

    def test_public_table_headers_use_column_scope(self) -> None:
        bad_pages = []
        for path in PUBLIC_PAGES:
            text = path.read_text()
            if re.search(r"<th(?=[\s>])(?![^>]*scope=)", text):
                bad_pages.append(path.name)
        self.assertFalse(
            bad_pages,
            f"Public HTML tables should declare header scope for accessibility: {', '.join(bad_pages)}",
        )

    def test_public_explicit_html_tables_use_row_headers(self) -> None:
        bad_pages = []
        row_header_re = re.compile(r"<tbody>.*?<tr>\s*<td>", re.DOTALL)
        for path in PUBLIC_PAGES:
            text = path.read_text()
            if row_header_re.search(text):
                bad_pages.append(path.name)
        self.assertFalse(
            bad_pages,
            f"Public HTML tables should use row headers in the first body cell: {', '.join(bad_pages)}",
        )


if __name__ == "__main__":
    unittest.main()
