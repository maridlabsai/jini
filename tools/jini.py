#!/usr/bin/env python3
"""Go-first Jini launcher with legacy Python compatibility behind the Go boundary."""

from __future__ import annotations

import os
import shutil
import subprocess
import sys
from pathlib import Path


ROOT = Path(__file__).resolve().parent.parent
LOCAL_GO = ROOT.parent / ".local-go" / "bin" / "go"

NATIVE_SINGLE_TOKEN_COMMANDS = {
    "help",
    "--help",
    "-h",
    "commands",
    "init",
    "memory",
    "new",
    "permissions",
    "route",
    "doctor",
}

NATIVE_MULTI_SHAPE_COMMANDS = {"observe"}


def _normalize_command(value: str) -> str:
    return value.strip().lower()


def _optional_format(argv: list[str]) -> str | None:
    if len(argv) >= 2 and argv[-2] == "--format":
        return _normalize_command(argv[-1])
    if argv and argv[-1].startswith("--format="):
        return _normalize_command(argv[-1].split("=", 1)[1])
    return None


def _matches_optional_format_shape(argv: list[str], *, command_words: int) -> bool:
    suffix = argv[command_words:]
    if len(suffix) == 2 and suffix[0] == "--format" and _normalize_command(suffix[1]) in {"json", "text"}:
        return True
    if len(suffix) == 1 and suffix[0].startswith("--format=") and _normalize_command(suffix[0].split("=", 1)[1]) in {"json", "text"}:
        return True
    return False


def _should_use_go(argv: list[str]) -> bool:
    if not argv:
        return True

    first = _normalize_command(argv[0])
    format_name = _optional_format(argv)
    if format_name is not None:
        if first == "doctor":
            return _matches_optional_format_shape(argv, command_words=1)
        if first == "provider":
            if _matches_optional_format_shape(argv, command_words=1):
                return True
            return len(argv) >= 2 and _normalize_command(argv[1]) == "doctor" and _matches_optional_format_shape(
                argv, command_words=2
            )
        return False

    if first == "help":
        return len(argv) == 1 or (len(argv) == 2 and _normalize_command(argv[1]) == "--all")
    if first in NATIVE_SINGLE_TOKEN_COMMANDS:
        return len(argv) == 1
    if first in NATIVE_MULTI_SHAPE_COMMANDS:
        return True
    if first == "admin":
        return len(argv) <= 2 and (len(argv) == 1 or _normalize_command(argv[1]) in {"help", "--help", "-h"})
    if first == "check":
        return len(argv) <= 2
    if first == "provider":
        return len(argv) == 1 or (len(argv) == 2 and _normalize_command(argv[1]) == "doctor")
    if first in {"status", "continue", "open"}:
        return len(argv) == 1
    if first == "run":
        return len(argv) == 1 or (len(argv) == 2 and _normalize_command(argv[1]) in {"new", "--new"})
    return False


def _go_command(argv: list[str]) -> list[str]:
    if LOCAL_GO.is_file() and os.access(LOCAL_GO, os.X_OK):
        return [str(LOCAL_GO), "run", "./cmd/jini", *argv]
    if shutil.which("go"):
        return ["go", "run", "./cmd/jini", *argv]
    built_binary = ROOT / "jini"
    if built_binary.is_file() and os.access(built_binary, os.X_OK):
        return [str(built_binary), *argv]
    return []


def main(argv: list[str] | None = None) -> int:
    args = list(sys.argv[1:] if argv is None else argv)
    if not _should_use_go(args):
        from jini_validate import main as legacy_main

        original_argv = sys.argv[:]
        try:
            sys.argv = [str(Path(__file__).resolve()), *args]
            return legacy_main()
        finally:
            sys.argv = original_argv
    command = _go_command(args)
    if not command:
        from jini_validate import main as legacy_main

        original_argv = sys.argv[:]
        try:
            sys.argv = [str(Path(__file__).resolve()), *args]
            return legacy_main()
        finally:
            sys.argv = original_argv
    env = dict(os.environ)
    env.setdefault("JINI_SOURCE_DIR", str(ROOT))
    env.setdefault("JINI_USE_LEGACY_FRONT_DOOR", "1")
    env.setdefault("JINI_CALLER_CWD", os.getcwd())
    env.setdefault("JINI_LEGACY_PYTHON", os.path.realpath(sys.executable))
    env.setdefault("GOCACHE", "/private/tmp/jini-go-cache")
    completed = subprocess.run(command, cwd=ROOT, env=env)
    return completed.returncode


if __name__ == "__main__":
    raise SystemExit(main())
