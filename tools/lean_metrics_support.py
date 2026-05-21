"""Support helpers for the lean-platform metrics surface.

These helpers keep the public validator focused on orchestration while this
module owns the lower-level route, alias, and command-sampling mechanics.
"""

from __future__ import annotations

import json
import os
import re
import subprocess
import sys
import tempfile
import time
from pathlib import Path
from typing import Any, Callable, Pattern


def collect_taught_commands(cli_doc_path: Path, taught_command_pattern: Pattern[str]) -> list[str]:
    """Return the canonical commands explicitly taught in the public CLI docs."""
    cli_doc = cli_doc_path.read_text(encoding="utf-8") if cli_doc_path.exists() else ""
    return sorted(
        {
            match.group(1).strip()
            for match in taught_command_pattern.finditer(cli_doc)
            if match.group(1).strip()
        }
    )


def scan_public_aliases(
    scan_paths: tuple[Path, ...],
    forbidden_aliases: tuple[str, ...],
    display_path: Callable[[Path], str],
) -> list[dict[str, Any]]:
    """Count any forbidden public command aliases still taught in docs."""
    alias_matches: list[dict[str, Any]] = []
    for path in scan_paths:
        text = path.read_text(encoding="utf-8") if path.exists() else ""
        normalized = text.lower()
        for alias in forbidden_aliases:
            count = len(re.findall(rf"\b{re.escape(alias)}\b", normalized))
            if count:
                alias_matches.append(
                    {
                        "path": display_path(path),
                        "alias": alias,
                        "count": count,
                    }
                )
    return alias_matches


def _quality_rank(value: str) -> int:
    normalized = str(value or "").strip().lower()
    return {"unknown": 0, "weak": 1, "usable": 2, "strong": 3}.get(normalized, 0)


def _reliability_rank(value: str) -> int:
    normalized = str(value or "").strip().lower()
    return {"unknown": 0, "fragile": 1, "usable": 2, "strong": 3}.get(normalized, 0)


def _route_history_trend(adapter_row: dict[str, Any], history: list[dict[str, Any]]) -> str:
    if len(history) < 2:
        return ""
    latest = history[-1]
    previous = history[:-1]
    row_stamp = str(adapter_row.get("benchmarked_at", "")).strip()
    latest_stamp = str(latest.get("benchmarked_at", "")).strip()
    if row_stamp and latest_stamp and row_stamp != latest_stamp:
        return ""
    latency_samples = [float(item.get("latency_ms", 0) or 0) for item in previous if float(item.get("latency_ms", 0) or 0) > 0]
    tps_samples = [float(item.get("tokens_per_second", 0) or 0) for item in previous if float(item.get("tokens_per_second", 0) or 0) > 0]
    avg_latency = (sum(latency_samples) / len(latency_samples)) if latency_samples else 0.0
    avg_tps = (sum(tps_samples) / len(tps_samples)) if tps_samples else 0.0
    latest_latency = float(latest.get("latency_ms", 0) or 0)
    latest_tps = float(latest.get("tokens_per_second", 0) or 0)
    latest_status = str(latest.get("status", "")).strip().lower()
    latest_quality = str(latest.get("quality_class", "")).strip().lower()
    latest_reliability = str(latest.get("structured_reliability", "")).strip().lower()
    if avg_latency > 0 and latest_latency >= avg_latency * 1.7:
        return "slower"
    if avg_tps > 0 and latest_tps > 0 and latest_tps <= avg_tps * 0.55:
        return "throughput-down"
    previous_had_issues = any(
        str(item.get("status", "")).strip().lower() in {"failed", "degraded"}
        or _quality_rank(str(item.get("quality_class", ""))) < _quality_rank("usable")
        or _reliability_rank(str(item.get("structured_reliability", ""))) < _reliability_rank("usable")
        for item in previous
    )
    if previous_had_issues and latest_status == "ok" and latest_quality == "strong" and latest_reliability == "strong":
        return "recovered"
    if latest_status == "ok" and latest_quality == "strong" and latest_reliability == "strong":
        return "stable"
    return ""


def load_route_report(
    session_state_root: Path,
    display_path: Callable[[Path], str],
) -> tuple[dict[str, Any], dict[str, Any]]:
    """Load the latest local runtime capability report when it exists."""
    local_route_path = session_state_root / "local-runtime-capabilities.json"
    route_evidence: dict[str, Any] = {
        "available": False,
        "path": display_path(local_route_path),
        "captured_at": "",
        "local_runtime_class": "",
        "adapter_count": 0,
        "ready_adapter_count": 0,
        "adapters": [],
    }
    payload: dict[str, Any] = {}
    if not local_route_path.exists():
        return route_evidence, payload
    try:
        loaded = json.loads(local_route_path.read_text(encoding="utf-8"))
    except (OSError, json.JSONDecodeError):
        loaded = {}
    if not isinstance(loaded, dict):
        return route_evidence, payload
    payload = loaded
    adapter_rows: list[dict[str, Any]] = []
    raw_adapters = payload.get("adapters", {})
    if isinstance(raw_adapters, dict):
        for adapter_id, row in sorted(raw_adapters.items()):
            if not isinstance(row, dict):
                continue
            adapter_rows.append(
                {
                    "adapter_id": str(adapter_id),
                    "status": str(row.get("status", "")),
                    "latency_ms": int(row.get("latency_ms", 0) or 0),
                    "warm_latency_ms": int(row.get("warm_latency_ms", 0) or 0),
                    "cold_start_cost_ms": int(row.get("cold_start_cost_ms", 0) or 0),
                    "tokens_per_second": float(row.get("tokens_per_second", 0) or 0),
                    "quality_class": str(row.get("quality_class", "")),
                    "structured_reliability": str(row.get("structured_reliability", "")),
                    "benchmarked_at": str(row.get("benchmarked_at", "")),
                }
            )
    route_evidence = {
        "available": bool(adapter_rows),
        "path": display_path(local_route_path),
        "captured_at": str(payload.get("captured_at", "")),
        "local_runtime_class": str(payload.get("local_runtime_class", "")),
        "adapter_count": len(adapter_rows),
        "ready_adapter_count": sum(1 for item in adapter_rows if item.get("status") in {"ok", "degraded"}),
        "adapters": adapter_rows,
    }
    return route_evidence, payload


def build_route_trend(route_evidence: dict[str, Any], payload: dict[str, Any]) -> dict[str, Any]:
    """Summarize whether local adapters are recovering or regressing over time."""
    route_trend_rows: list[dict[str, Any]] = []
    raw_history = payload.get("history", {})
    if isinstance(raw_history, dict):
        adapter_lookup = {
            str(item.get("adapter_id", "")): item
            for item in route_evidence.get("adapters", [])
            if str(item.get("adapter_id", "")).strip()
        }
        for adapter_id, entries in sorted(raw_history.items()):
            if not isinstance(entries, list):
                continue
            adapter_row = adapter_lookup.get(str(adapter_id))
            if not adapter_row:
                continue
            comparable_history = [item for item in entries if isinstance(item, dict)]
            trend = _route_history_trend(adapter_row, comparable_history)
            if not trend:
                continue
            route_trend_rows.append(
                {
                    "adapter_id": str(adapter_id),
                    "trend": trend,
                    "sample_count": len(comparable_history),
                    "latest_latency_ms": int(adapter_row.get("latency_ms", 0) or 0),
                    "latest_tokens_per_second": float(adapter_row.get("tokens_per_second", 0) or 0),
                }
            )
    return {
        "available": bool(route_trend_rows),
        "status": "measured" if route_trend_rows else "unavailable",
        "improving_count": sum(1 for item in route_trend_rows if item.get("trend") == "recovered"),
        "stable_count": sum(1 for item in route_trend_rows if item.get("trend") == "stable"),
        "regressing_count": sum(
            1
            for item in route_trend_rows
            if item.get("trend") in {"slower", "slower-watch", "throughput-down", "throughput-watch"}
        ),
        "adapters": route_trend_rows,
    }


def build_route_cost(route_evidence: dict[str, Any]) -> dict[str, Any]:
    """Summarize the cheapest measured local route posture when adapters are ready."""
    ready_route_adapters = [
        item
        for item in route_evidence.get("adapters", [])
        if item.get("status") in {"ok", "degraded"}
    ]
    cheapest_ready_adapter = None
    if ready_route_adapters:
        cheapest_ready_adapter = min(
            ready_route_adapters,
            key=lambda item: (
                int(item.get("cold_start_cost_ms", 0) or 0),
                int(item.get("warm_latency_ms", 0) or 0),
                -float(item.get("tokens_per_second", 0) or 0),
            ),
        )
    return {
        "available": bool(ready_route_adapters),
        "status": "measured" if ready_route_adapters else "unavailable",
        "basis": "local-runtime-benchmark" if ready_route_adapters else "none",
        "posture": "zero-external-api-spend" if ready_route_adapters else "unknown",
        "ready_adapter_count": len(ready_route_adapters),
        "avg_ready_warm_latency_ms": round(
            sum(float(item.get("warm_latency_ms", 0) or 0) for item in ready_route_adapters) / len(ready_route_adapters),
            1,
        )
        if ready_route_adapters
        else None,
        "avg_ready_tokens_per_second": round(
            sum(float(item.get("tokens_per_second", 0) or 0) for item in ready_route_adapters)
            / len(ready_route_adapters),
            1,
        )
        if ready_route_adapters
        else None,
        "cheapest_ready_adapter": (
            {
                "adapter_id": str(cheapest_ready_adapter.get("adapter_id", "")),
                "cold_start_cost_ms": int(cheapest_ready_adapter.get("cold_start_cost_ms", 0) or 0),
                "warm_latency_ms": int(cheapest_ready_adapter.get("warm_latency_ms", 0) or 0),
                "tokens_per_second": float(cheapest_ready_adapter.get("tokens_per_second", 0) or 0),
                "quality_class": str(cheapest_ready_adapter.get("quality_class", "")),
                "structured_reliability": str(cheapest_ready_adapter.get("structured_reliability", "")),
            }
            if cheapest_ready_adapter
            else None
        ),
    }


def collect_core_command_samples(root: Path, cli_path: Path) -> tuple[list[dict[str, Any]], dict[str, Any]]:
    """Sample only the canonical public commands so metrics stays cheap and stable."""
    command_samples: list[dict[str, Any]] = []
    provider_evidence: dict[str, Any] = {
        "available": False,
        "provider_id": "",
        "label": "",
        "status": "",
    }

    def sample_cli(
        args: list[str],
        *,
        env_overrides: dict[str, str] | None = None,
        parse_json_output: bool = False,
    ) -> tuple[dict[str, Any], dict[str, Any] | None]:
        env = os.environ.copy()
        if env_overrides:
            env.update(env_overrides)
        started = time.perf_counter()
        completed = subprocess.run(
            [sys.executable, str(cli_path), *args],
            cwd=root,
            check=False,
            capture_output=True,
            text=True,
            timeout=30,
            env=env,
        )
        duration_ms = round((time.perf_counter() - started) * 1000, 1)
        sample = {
            "command": " ".join(["jini", *args]),
            "duration_ms": duration_ms,
            "exit_code": int(completed.returncode),
        }
        parsed_json: dict[str, Any] | None = None
        if parse_json_output:
            try:
                loaded = json.loads(completed.stdout)
            except json.JSONDecodeError:
                loaded = None
            if isinstance(loaded, dict):
                parsed_json = loaded
        return sample, parsed_json

    with tempfile.TemporaryDirectory(prefix="jini-metrics-") as tmp_dir:
        tmp_root = Path(tmp_dir)
        sample, _ = sample_cli(["setup", "--harness", "codex", "--prefix", str(tmp_root / "setup-prefix")])
        command_samples.append(sample)
        doctor_sample, doctor_payload = sample_cli(
            ["doctor", "--format", "json"],
            env_overrides={"JINI_PROVIDER": "local-preview"},
            parse_json_output=True,
        )
        command_samples.append(doctor_sample)
        if isinstance(doctor_payload, dict):
            provider_evidence = {
                "available": bool(str(doctor_payload.get("provider_id", "")).strip()),
                "provider_id": str(doctor_payload.get("provider_id", "")),
                "label": str(doctor_payload.get("label", "")),
                "status": str(doctor_payload.get("status", "")),
            }
        sample, _ = sample_cli(["status", "packs/research-prd/examples/research-prd-v1"])
        command_samples.append(sample)
        sample, _ = sample_cli(["open", "prd", "--from", "packs/research-prd/examples/research-prd-v1", "--print-path"])
        command_samples.append(sample)
    return command_samples, provider_evidence


def build_latency_sample(command_samples: list[dict[str, Any]]) -> dict[str, Any]:
    """Reduce raw command timings into the latency summary used by the gate."""
    successful_durations = [float(item["duration_ms"]) for item in command_samples if int(item["exit_code"]) == 0]
    return {
        "sample_count": len(command_samples),
        "successful_sample_count": len(successful_durations),
        "max_ms": round(max(successful_durations), 1) if successful_durations else None,
        "min_ms": round(min(successful_durations), 1) if successful_durations else None,
        "avg_ms": round(sum(successful_durations) / len(successful_durations), 1) if successful_durations else None,
    }
