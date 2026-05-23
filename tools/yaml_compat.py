from __future__ import annotations

import ast
import json
import re
from typing import Any

try:
    import yaml as _yaml
except ModuleNotFoundError:  # pragma: no cover - optional dependency
    _yaml = None


class YAMLError(ValueError):
    """Raised when the bundled YAML subset parser cannot understand input."""


def _strip_comment(value: str) -> str:
    in_single = False
    in_double = False
    escaped = False
    for index, char in enumerate(value):
        if escaped:
            escaped = False
            continue
        if char == "\\" and in_double:
            escaped = True
            continue
        if char == "'" and not in_double:
            in_single = not in_single
            continue
        if char == '"' and not in_single:
            in_double = not in_double
            continue
        if char == "#" and not in_single and not in_double:
            if index == 0 or value[index - 1].isspace():
                return value[:index].rstrip()
    return value.rstrip()


def _indent(line: str) -> int:
    return len(line) - len(line.lstrip(" "))


def _scalar(value: str) -> Any:
    token = _strip_comment(value.strip())
    if token == "":
        return ""
    if token in {"null", "~"}:
        return None
    if token == "true":
        return True
    if token == "false":
        return False
    if token.startswith(('"', "'")) and token.endswith(('"', "'")) and len(token) >= 2:
        return ast.literal_eval(token)
    if token.startswith("[") or token.startswith("{"):
        try:
            return ast.literal_eval(token)
        except (SyntaxError, ValueError) as exc:
            raise YAMLError(f"Unsupported inline YAML literal: {token}") from exc
    if re.fullmatch(r"-?\d+", token):
        return int(token)
    if re.fullmatch(r"-?\d+\.\d+", token):
        return float(token)
    return token


def _parse_inline_literal_block(lines: list[str], start: int, indent: int) -> tuple[Any, int]:
    chunks: list[str] = []
    balance = 0
    index = start
    while index < len(lines):
        line = lines[index]
        if not line.strip():
            index += 1
            continue
        line_indent = _indent(line)
        if line_indent < indent and chunks:
            break
        text = line[indent:] if line_indent >= indent else line.lstrip()
        chunks.append(text)
        balance += text.count("[") + text.count("{")
        balance -= text.count("]") + text.count("}")
        index += 1
        if balance <= 0:
            break
    try:
        return ast.literal_eval("\n".join(chunks)), index
    except (SyntaxError, ValueError) as exc:
        raise YAMLError(f"Unsupported multiline YAML literal starting at line {start + 1}") from exc


def _parse_block_scalar(lines: list[str], start: int, parent_indent: int, style: str) -> tuple[str, int]:
    index = start
    values: list[str] = []
    while index < len(lines):
        line = lines[index]
        if not line.strip():
            values.append("")
            index += 1
            continue
        line_indent = _indent(line)
        if line_indent <= parent_indent:
            break
        values.append(line[parent_indent + 2 :])
        index += 1
    if style == ">":
        folded: list[str] = []
        for value in values:
            if value == "":
                folded.append("\n")
                continue
            if folded and folded[-1] != "\n":
                folded.append(" ")
            folded.append(value.strip())
        return "".join(folded).strip(), index
    return "\n".join(values), index


def _parse_block(lines: list[str], start: int, indent: int) -> tuple[Any, int]:
    index = start
    while index < len(lines):
        candidate = lines[index]
        if candidate.strip() and not candidate.lstrip().startswith("#"):
            break
        index += 1
    if index >= len(lines):
        return {}, index
    line = lines[index]
    current_indent = _indent(line)
    if current_indent < indent:
        return {}, index
    if current_indent > indent:
        raise YAMLError(f"Unexpected indentation at line {index + 1}")
    if line.lstrip().startswith("[") or line.lstrip().startswith("{"):
        return _parse_inline_literal_block(lines, index, indent)
    if line.lstrip().startswith("- "):
        return _parse_list(lines, index, indent)
    return _parse_dict(lines, index, indent)


def _parse_list(lines: list[str], start: int, indent: int) -> tuple[list[Any], int]:
    items: list[Any] = []
    index = start
    while index < len(lines):
        line = lines[index]
        if not line.strip():
            index += 1
            continue
        line_indent = _indent(line)
        if line_indent < indent:
            break
        if line_indent != indent or not line.lstrip().startswith("- "):
            break
        remainder = line.lstrip()[2:].rstrip()
        if remainder == "":
            value, index = _parse_block(lines, index + 1, indent + 2)
            items.append(value)
            continue
        if remainder in {">", "|"}:
            value, index = _parse_block_scalar(lines, index + 1, indent, remainder)
            items.append(value)
            continue
        key, separator, raw_value = remainder.partition(":")
        if separator:
            mapping: dict[str, Any] = {}
            key = key.strip()
            raw_value = raw_value.strip()
            if raw_value in {">", "|"}:
                value, next_index = _parse_block_scalar(lines, index + 1, indent, raw_value)
            elif raw_value == "":
                value, next_index = _parse_block(lines, index + 1, indent + 2)
            else:
                value = _scalar(raw_value)
                next_index = index + 1
            mapping[key] = value
            while next_index < len(lines):
                extra = lines[next_index]
                if not extra.strip():
                    next_index += 1
                    continue
                extra_indent = _indent(extra)
                if extra_indent < indent + 2 or (extra_indent == indent and extra.lstrip().startswith("- ")):
                    break
                if extra_indent != indent + 2:
                    raise YAMLError(f"Unexpected list indentation at line {next_index + 1}")
                extra_key, extra_separator, extra_value = extra.strip().partition(":")
                if not extra_separator:
                    raise YAMLError(f"Expected mapping entry at line {next_index + 1}")
                extra_value = extra_value.strip()
                if extra_value in {">", "|"}:
                    parsed_value, next_index = _parse_block_scalar(lines, next_index + 1, indent + 2, extra_value)
                elif extra_value == "":
                    parsed_value, next_index = _parse_block(lines, next_index + 1, indent + 4)
                else:
                    parsed_value = _scalar(extra_value)
                    next_index += 1
                mapping[extra_key.strip()] = parsed_value
            items.append(mapping)
            index = next_index
            continue
        items.append(_scalar(remainder))
        index += 1
    return items, index


def _parse_dict(lines: list[str], start: int, indent: int) -> tuple[dict[str, Any], int]:
    payload: dict[str, Any] = {}
    index = start
    while index < len(lines):
        line = lines[index]
        if not line.strip():
            index += 1
            continue
        line_indent = _indent(line)
        if line_indent < indent:
            break
        if line_indent != indent:
            raise YAMLError(f"Unexpected mapping indentation at line {index + 1}")
        stripped = line.strip()
        if stripped.startswith("- "):
            break
        key, separator, raw_value = stripped.partition(":")
        if not separator:
            raise YAMLError(f"Expected mapping entry at line {index + 1}")
        key = key.strip()
        raw_value = raw_value.strip()
        if raw_value in {">", "|"}:
            value, index = _parse_block_scalar(lines, index + 1, indent, raw_value)
        elif raw_value == "":
            value, index = _parse_block(lines, index + 1, indent + 2)
        else:
            value = _scalar(raw_value)
            index += 1
        payload[key] = value
    return payload, index


def safe_load(text: str) -> Any:
    if _yaml is not None:
        class _Loader(_yaml.SafeLoader):
            """Safe YAML loader that keeps timestamps as strings."""

        for key, resolvers in list(_Loader.yaml_implicit_resolvers.items()):
            _Loader.yaml_implicit_resolvers[key] = [
                resolver for resolver in resolvers if resolver[0] != "tag:yaml.org,2002:timestamp"
            ]
        return _yaml.load(text, Loader=_Loader)

    stripped = text.strip()
    if not stripped:
        return None
    if stripped.startswith("{") or stripped.startswith("["):
        return ast.literal_eval(stripped)
    lines = text.splitlines()
    parsed, index = _parse_block(lines, 0, 0)
    while index < len(lines):
        if lines[index].strip() and not lines[index].lstrip().startswith("#"):
            raise YAMLError(f"Trailing YAML content at line {index + 1}")
        index += 1
    return parsed


def _dump_scalar(value: Any) -> str:
    if value is None:
        return "null"
    if value is True:
        return "true"
    if value is False:
        return "false"
    if isinstance(value, (int, float)):
        return str(value)
    return json.dumps(str(value))


def _dump_lines(value: Any, indent: int = 0) -> list[str]:
    prefix = " " * indent
    if isinstance(value, dict):
        lines: list[str] = []
        for key, item in value.items():
            if isinstance(item, (dict, list)):
                lines.append(f"{prefix}{key}:")
                lines.extend(_dump_lines(item, indent + 2))
            else:
                lines.append(f"{prefix}{key}: {_dump_scalar(item)}")
        return lines or [f"{prefix}{{}}"]
    if isinstance(value, list):
        lines = []
        for item in value:
            if isinstance(item, dict) and item:
                iterator = iter(item.items())
                first_key, first_value = next(iterator)
                if isinstance(first_value, (dict, list)):
                    lines.append(f"{prefix}- {first_key}:")
                    lines.extend(_dump_lines(first_value, indent + 4))
                else:
                    lines.append(f"{prefix}- {first_key}: {_dump_scalar(first_value)}")
                for key, nested_value in iterator:
                    if isinstance(nested_value, (dict, list)):
                        lines.append(f"{prefix}  {key}:")
                        lines.extend(_dump_lines(nested_value, indent + 4))
                    else:
                        lines.append(f"{prefix}  {key}: {_dump_scalar(nested_value)}")
            elif isinstance(item, list):
                lines.append(f"{prefix}-")
                lines.extend(_dump_lines(item, indent + 2))
            else:
                lines.append(f"{prefix}- {_dump_scalar(item)}")
        return lines or [f"{prefix}[]"]
    return [f"{prefix}{_dump_scalar(value)}"]


def safe_dump(data: Any, *, sort_keys: bool = False) -> str:
    if _yaml is not None:
        return _yaml.safe_dump(data, sort_keys=sort_keys)
    if sort_keys and isinstance(data, dict):
        data = {key: data[key] for key in sorted(data.keys())}
    return "\n".join(_dump_lines(data)) + "\n"
