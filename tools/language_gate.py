#!/usr/bin/env python3
from __future__ import annotations

import argparse
import re
import subprocess
import sys
from dataclasses import dataclass
from pathlib import Path
from typing import Iterable


ROOT = Path(__file__).resolve().parents[1]

USER_FACING_SUFFIXES = {
    ".md",
    ".markdown",
    ".rst",
    ".txt",
}

SKIP_PARTS = {
    ".git",
    ".venv",
    "__pycache__",
    "vendor",
    "bundle-preview",
}

PROFANITY_PATTERNS = [
    ("expletive-f-word", re.compile(r"\bfuck(?:ing|ed|s)?\b", re.IGNORECASE)),
    ("expletive-s-word", re.compile(r"\bshit(?:ty)?\b", re.IGNORECASE)),
    ("crude-dismissal", re.compile(r"\b(?:bullshit|crap)\b", re.IGNORECASE)),
    ("personal-insult", re.compile(r"\b(?:asshole|bitch|bastard|dick)\b", re.IGNORECASE)),
    ("crude-anger", re.compile(r"\b(?:piss|pissed)\b", re.IGNORECASE)),
]


@dataclass(frozen=True)
class Match:
    path: Path
    line: int
    column: int
    label: str
    text: str


def display_path(path: Path) -> str:
    try:
        return str(path.resolve().relative_to(ROOT))
    except ValueError:
        return str(path)


def is_user_facing_path(path: Path) -> bool:
    if path.suffix.lower() not in USER_FACING_SUFFIXES:
        return False
    return not any(part in SKIP_PARTS for part in path.parts)


def changed_paths() -> list[Path]:
    names: set[str] = set()
    for args in (
        ["git", "diff", "--name-only", "--cached"],
        ["git", "diff", "--name-only"],
        ["git", "ls-files", "--others", "--exclude-standard"],
    ):
        result = subprocess.run(args, cwd=ROOT, text=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE, check=False)
        if result.returncode != 0:
            raise RuntimeError(result.stderr.strip() or f"{' '.join(args)} failed")
        names.update(line.strip() for line in result.stdout.splitlines() if line.strip())
    return sorted(ROOT / name for name in names)


def iter_scan_paths(paths: Iterable[Path]) -> list[Path]:
    scan_paths: list[Path] = []
    for path in paths:
        resolved = path if path.is_absolute() else ROOT / path
        if resolved.exists() and resolved.is_file() and is_user_facing_path(resolved):
            scan_paths.append(resolved)
    return scan_paths


def scan_file(path: Path) -> list[Match]:
    matches: list[Match] = []
    text = path.read_text(encoding="utf-8")
    for line_number, line in enumerate(text.splitlines(), start=1):
        for label, pattern in PROFANITY_PATTERNS:
            match = pattern.search(line)
            if match:
                matches.append(
                    Match(
                        path=path,
                        line=line_number,
                        column=match.start() + 1,
                        label=label,
                        text=line.strip(),
                    )
                )
    return matches


def scan_paths(paths: Iterable[Path]) -> list[Match]:
    matches: list[Match] = []
    for path in iter_scan_paths(paths):
        matches.extend(scan_file(path))
    return matches


def parse_args(argv: list[str]) -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Scan changed user-facing files for blocked language.")
    parser.add_argument(
        "paths",
        nargs="*",
        help="Specific paths to scan. When omitted, staged and unstaged changed files are scanned.",
    )
    return parser.parse_args(argv)


def main(argv: list[str]) -> int:
    args = parse_args(argv)
    paths = [Path(path) for path in args.paths] if args.paths else changed_paths()
    scan_targets = iter_scan_paths(paths)
    if not scan_targets:
        print("language-gate: no changed user-facing files to scan")
        return 0

    matches = scan_paths(scan_targets)
    if not matches:
        print(f"language-gate: scanned {len(scan_targets)} user-facing file(s); no blocked language found")
        return 0

    print("language-gate: blocked language found", file=sys.stderr)
    for match in matches:
        print(
            f"{display_path(match.path)}:{match.line}:{match.column}: "
            f"{match.label}: {match.text}",
            file=sys.stderr,
        )
    return 1


if __name__ == "__main__":
    raise SystemExit(main(sys.argv[1:]))
