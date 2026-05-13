#!/usr/bin/env python3
"""Lightweight validator for Jini WorkUnits, artifacts, and packs.

This validator intentionally supports a small subset of JSON Schema so the repo
remains self-contained. Supported keywords are:

- type
- properties
- required
- enum
- minimum
- minItems
- items
- additionalProperties
- pattern
"""

from __future__ import annotations

import argparse
import hashlib
import json
import os
import re
import shlex
import shutil
import subprocess
import sys
import tempfile
import time
from collections import Counter
from copy import deepcopy
from datetime import datetime, timezone
from itertools import zip_longest
from pathlib import Path
from typing import Any

import yaml

try:
    import tomllib
except ModuleNotFoundError:  # pragma: no cover - fallback for older Python
    tomllib = None


ROOT = Path(__file__).resolve().parent.parent
SCHEMA_ROOT = ROOT / "schemas"
REGISTRY_PATH = SCHEMA_ROOT / "schema-registry.json"
PACKS_ROOT = ROOT / "packs"
LEARNING_ROOT = ROOT / "learning"
POLICY_ROOT = LEARNING_ROOT / "policies"
LEARNING_EVENTS_ROOT = LEARNING_ROOT / "events"
LEARNING_EVENTS_PATH = LEARNING_EVENTS_ROOT / "runtime-events.jsonl"
FRAMEWORK_EVOLUTION_ROOT = LEARNING_ROOT / "framework-evolution"
COMPETITIVE_KPI_PATH = ROOT / "specs" / "competitive-kpis.yaml"
GOLDEN_BENCHMARK_PATH = ROOT / "specs" / "golden-competitive-benchmark.yaml"
INSTALL_MANIFEST_PATH = ROOT / "distribution" / "install-manifest.yaml"
ADAPTER_REGISTRY_PATH = ROOT / "distribution" / "adapter-registry.yaml"
VERSION_PATH = ROOT / "VERSION"
READY_ARTIFACT_STATUSES = {"reviewed", "approved", "merged"}
DONE_TASK_STATUSES = {"done", "complete", "completed", "verified", "closed"}
UNRESOLVED_TASK_STATES = {"pending", "blocked", "todo", "open", "not-started", "not_started"}
VERIFY_STATES = {"awaiting_verification", "operational"}
HIGH_CONTROL_PROFILES = {"Critical", "Regulated"}
RUNTIME_CONSENT_CATEGORIES = ("write", "command", "publish")
HARVEST_CATEGORY_ORDER = ("test", "verify", "startup", "demo", "docs")
PUBLIC_EXAMPLE_SPECS: dict[str, dict[str, Any]] = {
    "meeting-followup": {
        "label": "Meeting Follow-up",
        "pack_id": "meeting-followup",
        "source_kind": "compiled",
        "scenario": (
            "A normal weekly meeting ends with notes, action items, and implied commitments "
            "that usually get split across docs, chat, and memory."
        ),
        "title": "Weekly Product Review Follow-up",
        "purpose": "Turn one meeting into decisions owners and next steps",
        "owner": "meeting-owner",
        "approvers": ["team-lead"],
        "stakeholders": [],
        "daily_value": [
            "The meeting owner stops guessing what people heard.",
            "Action items stop living only in chat threads.",
            "Approvers can see what is still missing before work starts.",
            "The next person inherits state, not just notes.",
        ],
    },
    "research-prd": {
        "label": "Research To PRD Handoff",
        "pack_id": "research-prd",
        "source_kind": "bundled",
        "source_path": "packs/research-prd/examples/research-prd-v1",
        "scenario": (
            "Research exists and the team agrees something should be built, but the handoff "
            "still needs a truthful boundary between drafted, verified, and approved work."
        ),
        "daily_value": [
            "Product and engineering stop arguing from different versions of the rationale.",
            "People can see whether the spec is ready or merely drafted.",
            "Verification becomes a visible stage instead of an implied one.",
            "The handoff keeps its source trail attached to the work.",
        ],
    },
    "vendor-selection": {
        "label": "Vendor Selection",
        "pack_id": "vendor-selection",
        "source_kind": "compiled",
        "scenario": (
            "Several vendors look plausible and the team needs an approval-ready recommendation "
            "without losing tradeoffs, scoring, and objections."
        ),
        "title": "Vendor Evaluation",
        "purpose": "Compare shortlisted vendors and prepare an approval-ready recommendation",
        "owner": "procurement-lead",
        "approvers": ["finance-approver"],
        "stakeholders": [],
        "daily_value": [
            "The recommendation survives beyond the meeting where it was made.",
            "Tradeoffs stay attached to the final answer.",
            "Finance or leadership can see the approval path without re-asking for context.",
            "The team can revisit the decision later without reconstructing it from memory.",
        ],
    },
    "incident-response": {
        "label": "Incident Response",
        "pack_id": "incident-response",
        "source_kind": "compiled",
        "scenario": (
            "The immediate outage is over, but the operational work still needs rollback, "
            "verification evidence, and honest closure state."
        ),
        "title": "Checkout Latency Incident",
        "purpose": "Stabilize the checkout path with explicit rollback and verification",
        "owner": "incident-commander",
        "approvers": ["service-owner"],
        "stakeholders": [],
        "daily_value": [
            "Responders stop treating recovery as closure.",
            "Rollback context stays visible while pressure is high.",
            "Verification evidence gets attached before the story drifts.",
            "Closure becomes a real state, not an assumption.",
        ],
    },
}
LINEAR_STATE_ORDER = [
    "intake",
    "scoped",
    "probed",
    "modeled",
    "decided",
    "in_make",
    "awaiting_verification",
    "operational",
    "retired",
]
NEXT_STATE_BY_PROGRESS = {
    state: LINEAR_STATE_ORDER[idx + 1]
    for idx, state in enumerate(LINEAR_STATE_ORDER[:-1])
}
NEXT_OPERATION_BY_STATE = {
    "intake": "Scope",
    "scoped": "Probe",
    "probed": "Model",
    "modeled": "Decide",
    "decided": "Make",
    "in_make": "Make",
    "awaiting_verification": "Verify",
    "operational": "Maintain",
    "reopened": "Probe",
    "incident": "Verify",
    "retired": "Archive",
}
STATE_REQUIRED_ARTIFACTS = {
    "intake": [],
    "scoped": ["Brief", "Assumptions"],
    "probed": ["Brief", "Assumptions"],
    "modeled": ["Brief", "Assumptions", "Spec", "Signals", "Rollback", "Runbook"],
    "decided": ["Brief", "Assumptions", "Spec", "Decision", "Plan", "Tasks", "Signals", "Rollback", "Runbook"],
    "in_make": ["Brief", "Assumptions", "Spec", "Decision", "Plan", "Tasks", "Signals", "Rollback", "Runbook"],
    "awaiting_verification": ["Brief", "Assumptions", "Spec", "Decision", "Plan", "Tasks", "Signals", "Rollback", "Runbook", "Evidence"],
    "operational": ["Brief", "Assumptions", "Spec", "Decision", "Plan", "Tasks", "Signals", "Rollback", "Runbook", "Evidence", "Approval"],
    "reopened": ["Brief", "Assumptions"],
    "incident": ["Brief", "Assumptions", "Decision", "Plan", "Tasks", "Signals", "Rollback", "Runbook", "Evidence"],
    "retired": ["Brief", "Decision", "Evidence"],
}


class JiniYamlLoader(yaml.SafeLoader):
    """Safe YAML loader that keeps timestamps as strings."""


for key, resolvers in list(JiniYamlLoader.yaml_implicit_resolvers.items()):
    JiniYamlLoader.yaml_implicit_resolvers[key] = [
        resolver for resolver in resolvers if resolver[0] != "tag:yaml.org,2002:timestamp"
    ]


def load_document(path: Path) -> Any:
    text = path.read_text(encoding="utf-8")
    if path.suffix.lower() == ".json":
        return json.loads(text)
    return yaml.load(text, Loader=JiniYamlLoader)


def dump_document(path: Path, data: Any) -> None:
    if path.suffix.lower() == ".json":
        path.write_text(json.dumps(data, indent=2) + "\n", encoding="utf-8")
        return
    path.write_text(yaml.safe_dump(data, sort_keys=False), encoding="utf-8")


def load_version() -> str:
    try:
        return VERSION_PATH.read_text(encoding="utf-8").strip() or "0.0.0"
    except OSError:
        return "0.0.0"


def type_name(value: Any) -> str:
    if isinstance(value, bool):
        return "boolean"
    if isinstance(value, int):
        return "integer"
    if isinstance(value, float):
        return "number"
    if isinstance(value, str):
        return "string"
    if isinstance(value, list):
        return "array"
    if isinstance(value, dict):
        return "object"
    if value is None:
        return "null"
    return type(value).__name__


def validate(instance: Any, schema: dict[str, Any], path: str = "$") -> list[str]:
    errors: list[str] = []

    expected_type = schema.get("type")
    if expected_type:
        actual_type = type_name(instance)
        if expected_type == "number":
            ok = actual_type in {"integer", "number"}
        else:
            ok = actual_type == expected_type
        if not ok:
            return [f"{path}: expected type {expected_type}, got {actual_type}"]

    if "enum" in schema and instance not in schema["enum"]:
        errors.append(f"{path}: expected one of {schema['enum']}, got {instance!r}")

    if "minimum" in schema and isinstance(instance, (int, float)) and instance < schema["minimum"]:
        errors.append(f"{path}: expected minimum {schema['minimum']}, got {instance}")

    if "pattern" in schema and isinstance(instance, str):
        if re.match(schema["pattern"], instance) is None:
            errors.append(f"{path}: value {instance!r} does not match pattern {schema['pattern']}")

    if expected_type == "array":
        if "minItems" in schema and len(instance) < schema["minItems"]:
            errors.append(f"{path}: expected at least {schema['minItems']} items, got {len(instance)}")
        item_schema = schema.get("items")
        if item_schema:
            for idx, item in enumerate(instance):
                errors.extend(validate(item, item_schema, f"{path}[{idx}]"))

    if expected_type == "object":
        required = schema.get("required", [])
        for key in required:
            if key not in instance:
                errors.append(f"{path}: missing required field {key!r}")

        properties = schema.get("properties", {})
        for key, subschema in properties.items():
            if key in instance:
                errors.extend(validate(instance[key], subschema, f"{path}.{key}"))

        if schema.get("additionalProperties") is False:
            extra = sorted(set(instance.keys()) - set(properties.keys()))
            for key in extra:
                errors.append(f"{path}: unexpected field {key!r}")

    return errors


def load_registry() -> dict[str, Any]:
    return json.loads(REGISTRY_PATH.read_text(encoding="utf-8"))


def load_competitive_kpis() -> dict[str, Any]:
    scorecard = load_document(COMPETITIVE_KPI_PATH)
    if not isinstance(scorecard, dict):
        raise ValueError("competitive KPI scorecard must be a mapping")
    dimensions = scorecard.get("dimensions")
    if not isinstance(dimensions, list) or not dimensions:
        raise ValueError("competitive KPI scorecard must define non-empty dimensions")
    return scorecard


def load_golden_benchmark() -> dict[str, Any]:
    benchmark = load_document(GOLDEN_BENCHMARK_PATH)
    if not isinstance(benchmark, dict):
        raise ValueError("golden benchmark must be a mapping")
    scenarios = benchmark.get("scenarios")
    competitors = benchmark.get("competitors")
    if not isinstance(scenarios, list) or not scenarios:
        raise ValueError("golden benchmark must define non-empty scenarios")
    if not isinstance(competitors, list) or len(competitors) < 2:
        raise ValueError("golden benchmark must define at least two competitors")
    return benchmark


def golden_benchmark_digest() -> str:
    return hashlib.sha256(GOLDEN_BENCHMARK_PATH.read_text(encoding="utf-8").encode("utf-8")).hexdigest()[:16]


def load_adapter_registry() -> dict[str, Any]:
    registry = load_document(ADAPTER_REGISTRY_PATH)
    if not isinstance(registry, dict):
        raise ValueError("adapter registry must be a mapping")
    adapters = registry.get("adapters")
    if not isinstance(adapters, list) or not adapters:
        raise ValueError("adapter registry must define non-empty adapters")
    return registry


def build_competitive_kpi_summary(
    scorecard: dict[str, Any],
    *,
    dimension: str | None = None,
    limit: int | None = None,
) -> dict[str, Any]:
    target_default = float(scorecard.get("target_score", 9.0))
    entries: list[dict[str, Any]] = []

    for raw_entry in scorecard["dimensions"]:
        if not isinstance(raw_entry, dict):
            raise ValueError("competitive KPI entries must be mappings")
        entry = deepcopy(raw_entry)
        competitor = entry.get("strongest_competitor")
        if not isinstance(competitor, dict):
            raise ValueError(f"KPI {entry.get('id', '<unknown>')} is missing strongest_competitor metadata")

        current_score = float(entry["current_score"])
        target_score = float(entry.get("target_score", target_default))
        competitor_score = float(competitor["score"])

        entry["current_score"] = current_score
        entry["target_score"] = target_score
        entry["gap_to_target"] = round(max(0.0, target_score - current_score), 1)
        entry["competitive_gap"] = round(max(0.0, competitor_score - current_score), 1)
        entry["strength_status"] = "leading" if current_score >= competitor_score else "trailing"
        entries.append(entry)

    if dimension:
        needle = dimension.strip().lower()
        filtered = [
            entry
            for entry in entries
            if needle in entry["id"].lower() or needle in entry["label"].lower()
        ]
        if not filtered:
            raise KeyError(f"Unknown KPI dimension {dimension!r}")
        entries = filtered

    entries.sort(
        key=lambda item: (
            int(item.get("execution_order", 999)),
            -item["gap_to_target"],
            -item["competitive_gap"],
            item["id"],
        )
    )

    if dimension is None and limit is not None:
        entries = entries[:limit]

    return {
        "updated_at": scorecard.get("updated_at"),
        "target_score": target_default,
        "comparison_set": scorecard.get("comparison_set", []),
        "dimension_filter": dimension,
        "dimensions": entries,
    }


def print_competitive_kpi_summary(summary: dict[str, Any]) -> None:
    print(f"UPDATED {summary.get('updated_at', 'unknown')}")
    print(f"TARGET  {summary['target_score']:.1f}")
    comparison_set = summary.get("comparison_set") or []
    if comparison_set:
        print("VERSUS  " + ", ".join(str(item) for item in comparison_set))

    if summary.get("dimension_filter"):
        for item in summary["dimensions"]:
            strongest = item["strongest_competitor"]
            print(f"DIM    {item['label']} ({item['id']})")
            print(f"CUR    {item['current_score']:.1f}")
            print(f"TGT    {item['target_score']:.1f}")
            print(f"BEST   {strongest['name']} ({float(strongest['score']):.1f})")
            print(f"STATE  {item['strength_status']}")
            print(f"GAP    target={item['gap_to_target']:.1f} competitor={item['competitive_gap']:.1f}")
            print(f"EDGE   {strongest['advantage']}")
            print(f"WEAK   {item['weakness']}")
            print(f"TURN   {item['conversion_strategy']}")
            print(f"KPI    {item['kpi']}")
            print("NEXT")
            for step in item.get("next_build_steps", []):
                print(f"  - {step}")
            print("DONE")
            for criterion in item.get("exit_criteria", []):
                print(f"  - {criterion}")
        return

    print("TOP")
    for index, item in enumerate(summary["dimensions"], start=1):
        strongest = item["strongest_competitor"]
        print(
            f"{index:>2}. {item['id']} | cur {item['current_score']:.1f} | "
            f"tgt {item['target_score']:.1f} | best {strongest['name']} "
            f"({float(strongest['score']):.1f}) | gap {item['gap_to_target']:.1f}"
        )
        print(f"    weak: {item['weakness']}")
        next_steps = item.get("next_build_steps", [])
        if next_steps:
            print(f"    next: {next_steps[0]}")


def golden_benchmark_report_dir() -> Path:
    path = LEARNING_ROOT / "golden-benchmark" / "reports"
    path.mkdir(parents=True, exist_ok=True)
    return path


def next_golden_benchmark_report_path() -> Path:
    stamp = datetime.now(timezone.utc).strftime("%Y%m%dT%H%M%S%fZ")
    return golden_benchmark_report_dir() / f"golden-benchmark-{stamp}.json"


def format_benchmark_value(value: Any, context: dict[str, str]) -> Any:
    if isinstance(value, str):
        return value.format(**context)
    if isinstance(value, list):
        return [format_benchmark_value(item, context) for item in value]
    if isinstance(value, dict):
        return {key: format_benchmark_value(item, context) for key, item in value.items()}
    return value


def run_benchmark_command(args: list[str], *, cwd: Path = ROOT) -> subprocess.CompletedProcess[str]:
    cli_path = ROOT / "tools" / "jini.py"
    command = [sys.executable, str(cli_path), *args]
    env = os.environ.copy()
    env["JINI_GOLDEN_BENCHMARK_ACTIVE"] = "1"
    return subprocess.run(
        command,
        cwd=cwd,
        check=False,
        capture_output=True,
        text=True,
        timeout=45,
        env=env,
    )


def evaluate_benchmark_assertions(actual: Any, assertion: dict[str, Any]) -> list[str]:
    failures: list[str] = []
    if "equals" in assertion:
        expected = assertion["equals"]
        if actual != expected:
            failures.append(f"expected `{expected}` got `{actual}`")
    if "contains" in assertion:
        needle = assertion["contains"]
        if isinstance(actual, list):
            if not any(str(needle) in str(item) for item in actual):
                failures.append(f"expected list to contain `{needle}`")
        elif str(needle) not in str(actual):
            failures.append(f"expected `{actual}` to contain `{needle}`")
    if "not_contains" in assertion:
        needle = assertion["not_contains"]
        if isinstance(actual, list):
            if any(str(needle) in str(item) for item in actual):
                failures.append(f"expected list to exclude `{needle}`")
        elif str(needle) in str(actual):
            failures.append(f"expected `{actual}` to exclude `{needle}`")
    if "minimum" in assertion:
        try:
            numeric = float(actual)
        except (TypeError, ValueError):
            failures.append(f"expected numeric value for minimum check, got `{actual}`")
        else:
            if numeric < float(assertion["minimum"]):
                failures.append(f"expected minimum `{assertion['minimum']}` got `{numeric}`")
    if "maximum" in assertion:
        try:
            numeric = float(actual)
        except (TypeError, ValueError):
            failures.append(f"expected numeric value for maximum check, got `{actual}`")
        else:
            if numeric > float(assertion["maximum"]):
                failures.append(f"expected maximum `{assertion['maximum']}` got `{numeric}`")
    return failures


def execute_golden_benchmark_check(check: dict[str, Any], context: dict[str, str]) -> dict[str, Any]:
    check_id = str(check.get("id", "check")).strip() or "check"
    raw_command = check.get("command", [])
    if not isinstance(raw_command, list) or not raw_command:
        return {
            "id": check_id,
            "status": "error",
            "detail": "Benchmark check is missing command arguments.",
        }
    args = [str(item) for item in format_benchmark_value(raw_command, context)]
    completed = run_benchmark_command(args)
    if completed.returncode != 0:
        return {
            "id": check_id,
            "status": "failed",
            "detail": f"Command exited with code {completed.returncode}: {trim_output(completed.stderr or completed.stdout)}",
            "command": args,
        }
    try:
        payload = json.loads(completed.stdout)
    except json.JSONDecodeError:
        return {
            "id": check_id,
            "status": "failed",
            "detail": "Benchmark command did not emit valid JSON output.",
            "command": args,
        }
    path = str(check.get("path", "")).strip()
    actual = payload
    if path:
        try:
            actual = resolve_json_path(payload, path)
        except KeyError:
            return {
                "id": check_id,
                "status": "failed",
                "detail": f"JSON path `{path}` was not present in output.",
                "command": args,
            }
    failures = evaluate_benchmark_assertions(actual, check)
    return {
        "id": check_id,
        "status": "ok" if not failures else "failed",
        "detail": "ok" if not failures else "; ".join(failures),
        "command": args,
        "path": path,
        "actual": actual,
    }


def run_benchmark_setup_step(step: dict[str, Any], context: dict[str, str]) -> dict[str, Any]:
    step_id = str(step.get("id", "setup")).strip() or "setup"
    raw_command = step.get("command", [])
    if not isinstance(raw_command, list) or not raw_command:
        return {"id": step_id, "status": "error", "detail": "Benchmark setup is missing command arguments."}
    args = [str(item) for item in format_benchmark_value(raw_command, context)]
    completed = run_benchmark_command(args)
    return {
        "id": step_id,
        "status": "ok" if completed.returncode == 0 else "failed",
        "detail": "ok" if completed.returncode == 0 else trim_output(completed.stderr or completed.stdout),
        "command": args,
    }


def build_golden_benchmark_projection() -> dict[str, Any]:
    benchmark = load_golden_benchmark()
    scorecard = build_competitive_kpi_summary(load_competitive_kpis(), limit=None)
    dimension_lookup = {
        str(item.get("id", "")): item
        for item in scorecard.get("dimensions", [])
        if isinstance(item, dict)
    }
    competitor_rows = benchmark.get("competitors", [])
    competitor_ids = [str(item.get("id", "")).strip() for item in competitor_rows if isinstance(item, dict)]
    scenarios: list[dict[str, Any]] = []
    aggregate_weight = 0.0
    jini_total = 0.0
    competitor_totals: dict[str, float] = {competitor_id: 0.0 for competitor_id in competitor_ids}

    for raw_scenario in benchmark.get("scenarios", []):
        if not isinstance(raw_scenario, dict):
            continue
        dimensions = [str(item) for item in raw_scenario.get("dimensions", []) if str(item).strip()]
        dimension_scores = [float(dimension_lookup[item]["current_score"]) for item in dimensions if item in dimension_lookup]
        jini_score = round(sum(dimension_scores) / len(dimension_scores), 2) if dimension_scores else 0.0
        weight = float(raw_scenario.get("weight", 1.0))
        competitor_scores = {
            competitor_id: float(raw_scenario.get("competitor_scores", {}).get(competitor_id, 0.0))
            for competitor_id in competitor_ids
        }
        aggregate_weight += weight
        jini_total += jini_score * weight
        for competitor_id, score in competitor_scores.items():
            competitor_totals[competitor_id] += score * weight
        scenarios.append(
            {
                "id": str(raw_scenario.get("id", "")),
                "label": str(raw_scenario.get("label", "")),
                "purpose": str(raw_scenario.get("purpose", "")),
                "weight": weight,
                "dimensions": dimensions,
                "dimension_scores": {item: float(dimension_lookup[item]["current_score"]) for item in dimensions if item in dimension_lookup},
                "base_jini_score": jini_score,
                "validated_jini_score": jini_score,
                "check_pass_rate": 1.0,
                "status": "projected",
                "setup": [],
                "checks": [],
                "competitor_scores": competitor_scores,
            }
        )

    jini_aggregate = round(jini_total / aggregate_weight, 2) if aggregate_weight else 0.0
    competitor_aggregates = {
        competitor_id: round(total / aggregate_weight, 2) if aggregate_weight else 0.0
        for competitor_id, total in competitor_totals.items()
    }
    strongest_competitor = ""
    strongest_score = 0.0
    if competitor_aggregates:
        strongest_competitor, strongest_score = max(competitor_aggregates.items(), key=lambda item: item[1])
    return {
        "schema_version": benchmark.get("schema_version", "0.1.0"),
        "report_type": "JiniGoldenBenchmarkProjection",
        "generated_at": now_utc(),
        "benchmark_id": benchmark.get("benchmark_id", "golden-benchmark"),
        "benchmark_label": benchmark.get("label", "Jini golden benchmark"),
        "updated_at": benchmark.get("updated_at", ""),
        "last_verified_at": benchmark.get("last_verified_at", benchmark.get("updated_at", "")),
        "methodology": benchmark.get("methodology", {}),
        "dataset_path": display_path(GOLDEN_BENCHMARK_PATH),
        "dataset_digest": golden_benchmark_digest(),
        "report_path": "",
        "competitors": competitor_rows,
        "scenario_count": len(scenarios),
        "overall": {
            "status": "projected-leading" if jini_aggregate >= strongest_score else "projected-trailing",
            "jini_score": jini_aggregate,
            "competitor_scores": competitor_aggregates,
            "strongest_competitor": strongest_competitor,
            "strongest_competitor_score": strongest_score,
            "failed_scenarios": [],
        },
        "scenarios": scenarios,
    }


def build_golden_benchmark_report() -> tuple[dict[str, Any], Path]:
    benchmark = load_golden_benchmark()
    scorecard = build_competitive_kpi_summary(load_competitive_kpis(), limit=None)
    dimension_lookup = {
        str(item.get("id", "")): item
        for item in scorecard.get("dimensions", [])
        if isinstance(item, dict)
    }
    competitor_rows = benchmark.get("competitors", [])
    competitor_ids = [str(item.get("id", "")).strip() for item in competitor_rows if isinstance(item, dict)]
    report_path = next_golden_benchmark_report_path()
    temp_root = Path(tempfile.mkdtemp(prefix="jini-golden-benchmark-"))
    try:
        scenarios: list[dict[str, Any]] = []
        aggregate_weight = 0.0
        jini_total = 0.0
        competitor_totals: dict[str, float] = {competitor_id: 0.0 for competitor_id in competitor_ids}

        for raw_scenario in benchmark.get("scenarios", []):
            if not isinstance(raw_scenario, dict):
                continue
            scenario_id = str(raw_scenario.get("id", "")).strip()
            if not scenario_id:
                continue
            scenario_root = temp_root / scenario_id
            scenario_root.mkdir(parents=True, exist_ok=True)
            context = {
                "repo_root": str(ROOT),
                "temp_root": str(temp_root),
                "scenario_root": str(scenario_root),
            }
            setup_results = [
                run_benchmark_setup_step(step, context)
                for step in raw_scenario.get("setup", [])
                if isinstance(step, dict)
            ]
            checks = [
                execute_golden_benchmark_check(check, context)
                for check in raw_scenario.get("checks", [])
                if isinstance(check, dict)
            ]
            setup_failed = any(item["status"] != "ok" for item in setup_results)
            passed_checks = sum(1 for item in checks if item["status"] == "ok")
            check_total = len(checks)
            pass_rate = 0.0 if check_total == 0 else passed_checks / check_total
            dimensions = [str(item) for item in raw_scenario.get("dimensions", []) if str(item).strip()]
            dimension_scores = [float(dimension_lookup[item]["current_score"]) for item in dimensions if item in dimension_lookup]
            base_jini_score = round(sum(dimension_scores) / len(dimension_scores), 2) if dimension_scores else 0.0
            validated_score = round(base_jini_score * (0.6 + 0.4 * pass_rate), 2) if not setup_failed else 0.0
            weight = float(raw_scenario.get("weight", 1.0))
            competitor_scores = {
                competitor_id: float(raw_scenario.get("competitor_scores", {}).get(competitor_id, 0.0))
                for competitor_id in competitor_ids
            }
            if not setup_failed:
                aggregate_weight += weight
                jini_total += validated_score * weight
                for competitor_id, score in competitor_scores.items():
                    competitor_totals[competitor_id] += score * weight
            scenarios.append(
                {
                    "id": scenario_id,
                    "label": str(raw_scenario.get("label", scenario_id)),
                    "purpose": str(raw_scenario.get("purpose", "")).strip(),
                    "weight": weight,
                    "dimensions": dimensions,
                    "dimension_scores": {item: float(dimension_lookup[item]["current_score"]) for item in dimensions if item in dimension_lookup},
                    "base_jini_score": base_jini_score,
                    "validated_jini_score": validated_score,
                    "check_pass_rate": round(pass_rate, 3),
                    "status": "failed" if setup_failed or pass_rate < 1.0 else "ok",
                    "setup": setup_results,
                    "checks": checks,
                    "competitor_scores": competitor_scores,
                }
            )

        jini_aggregate = round(jini_total / aggregate_weight, 2) if aggregate_weight else 0.0
        competitor_aggregates = {
            competitor_id: round(total / aggregate_weight, 2) if aggregate_weight else 0.0
            for competitor_id, total in competitor_totals.items()
        }
        strongest_competitor = ""
        strongest_score = 0.0
        if competitor_aggregates:
            strongest_competitor, strongest_score = max(competitor_aggregates.items(), key=lambda item: item[1])
        report = {
            "schema_version": benchmark.get("schema_version", "0.1.0"),
            "report_type": "JiniGoldenBenchmarkValidation",
            "generated_at": now_utc(),
            "benchmark_id": benchmark.get("benchmark_id", "golden-benchmark"),
            "benchmark_label": benchmark.get("label", "Jini golden benchmark"),
            "updated_at": benchmark.get("updated_at", ""),
            "last_verified_at": benchmark.get("last_verified_at", benchmark.get("updated_at", "")),
            "methodology": benchmark.get("methodology", {}),
            "dataset_path": display_path(GOLDEN_BENCHMARK_PATH),
            "dataset_digest": golden_benchmark_digest(),
            "report_path": display_path(report_path),
            "competitors": competitor_rows,
            "scenario_count": len(scenarios),
            "overall": {
                "status": "leading" if jini_aggregate >= strongest_score else "trailing",
                "jini_score": jini_aggregate,
                "competitor_scores": competitor_aggregates,
                "strongest_competitor": strongest_competitor,
                "strongest_competitor_score": strongest_score,
                "failed_scenarios": [scenario["id"] for scenario in scenarios if scenario["status"] != "ok"],
            },
            "scenarios": scenarios,
        }
        dump_document(report_path, report)
        return report, report_path
    finally:
        shutil.rmtree(temp_root, ignore_errors=True)


def print_golden_benchmark_report(report: dict[str, Any]) -> None:
    print(f"BENCH   {report.get('benchmark_label', '')}")
    print(f"DATASET {report.get('dataset_path', '')}")
    if report.get("dataset_digest"):
        print(f"DIGEST  {report.get('dataset_digest', '')}")
    if report.get("last_verified_at"):
        print(f"VERIFY  {report.get('last_verified_at', '')}")
    print(f"STATUS  {report.get('overall', {}).get('status', '')}")
    print(f"Jini   {report.get('overall', {}).get('jini_score', 0.0):.2f}")
    strongest_competitor = report.get("overall", {}).get("strongest_competitor", "")
    if strongest_competitor:
        strongest_score = report.get("overall", {}).get("strongest_competitor_score", 0.0)
        print(f"BEST    {strongest_competitor} ({strongest_score:.2f})")
    print("SCENARIOS")
    for scenario in report.get("scenarios", []):
        print(
            f"  - {scenario['id']} | {scenario['status']} | "
            f"jini={scenario['validated_jini_score']:.2f} | pass={scenario['check_pass_rate']:.3f}"
        )


def framework_review_dir() -> Path:
    path = FRAMEWORK_EVOLUTION_ROOT / "reviews"
    path.mkdir(parents=True, exist_ok=True)
    return path


def framework_experiment_dir() -> Path:
    path = FRAMEWORK_EVOLUTION_ROOT / "experiments"
    path.mkdir(parents=True, exist_ok=True)
    return path


def framework_outcome_dir() -> Path:
    path = FRAMEWORK_EVOLUTION_ROOT / "outcomes"
    path.mkdir(parents=True, exist_ok=True)
    return path


def next_framework_review_path() -> Path:
    stamp = datetime.now(timezone.utc).strftime("%Y%m%dT%H%M%S%fZ")
    return framework_review_dir() / f"framework-review-{stamp}.json"


def latest_framework_review_path() -> Path | None:
    reviews = sorted(framework_review_dir().glob("framework-review-*.json"))
    if not reviews:
        return None
    return reviews[-1]


def next_framework_experiment_path(dimension_id: str) -> Path:
    stamp = datetime.now(timezone.utc).strftime("%Y%m%dT%H%M%S%fZ")
    return framework_experiment_dir() / f"{slugify(dimension_id)}-experiment-{stamp}.json"


def next_framework_outcome_path(dimension_id: str) -> Path:
    stamp = datetime.now(timezone.utc).strftime("%Y%m%dT%H%M%S%fZ")
    return framework_outcome_dir() / f"{slugify(dimension_id)}-outcome-{stamp}.json"


def framework_adoption_constraints() -> list[str]:
    return [
        "Any novice should be able to reach first value without knowing Python, internal jargon, or bundle-level architecture first.",
        "Time to first useful output must be short enough that a new user reaches value before reading deep docs.",
        "Install, activation, and first-run trust must be visible and low-risk across supported targets.",
        "The cheapest adequate path must feel like the default path, not an expert mode.",
        "Memory must change behavior at the right step without forcing the user to restate context.",
        "Portable execution edges must be real enough that users can stay in their preferred runtime or system.",
        "Anything that adds friction or kernel pressure without proven adoption value should be removed, demoted, or consolidated.",
    ]


def framework_adoption_weights() -> dict[str, float]:
    return {
        "workflow-rigor": 1.0,
        "delivery-maturity": 1.8,
        "packaging-install": 1.7,
        "memory-reliability": 1.5,
        "adapter-portability": 1.7,
        "token-efficiency": 1.6,
        "advanced-set-breadth": 1.2,
        "learning-maturity": 1.4,
        "flexibility": 1.1,
        "governance": 0.9,
        "core-simplicity": 1.0,
    }


def framework_user_value_reason(dimension_id: str) -> str:
    reasons = {
        "workflow-rigor": "Users adopt a framework more easily when the protocol stays legible under pressure.",
        "delivery-maturity": "Users keep a framework when it reduces operator decisions and gets them to output fast.",
        "packaging-install": "Adoption stalls when installation, activation, or first-run trust is weak.",
        "memory-reliability": "Reliable memory reduces repeated setup and makes the framework feel dependable over time.",
        "adapter-portability": "Users stay when the framework works in their preferred runtime and work systems.",
        "token-efficiency": "Low-cost default loops make the product cheap enough to become habitual.",
        "advanced-set-breadth": "Breadth matters when one framework can replace many fragmented workflows.",
        "learning-maturity": "Governed learning is what turns improvement from manual taste into a compounding system.",
        "flexibility": "Flexibility expands the addressable user base without forcing a new framework per domain.",
        "governance": "Trust and auditability are retention levers when work becomes important or shared.",
        "core-simplicity": "A small core lowers learning cost and prevents collapse under feature pressure.",
    }
    return reasons.get(dimension_id, "Users adopt systems that save time, reduce friction, and stay understandable.")


def framework_reward_signals(dimension_id: str) -> list[str]:
    signals = {
        "workflow-rigor": ["manual-yaml-repair-reduced", "legal-next-step-clarity", "compiler-success-rate"],
        "delivery-maturity": ["manual-decisions-reduced", "time-to-first-useful-output", "execute-flow-completion"],
        "packaging-install": ["install-success-rate", "doctor-clean-rate", "first-run-trust"],
        "memory-reliability": ["context-restatement-reduced", "memory-hit-rate", "stale-memory-reduction"],
        "adapter-portability": ["live-edge-count", "portable-apply-success-rate", "adapter-conformance-pass-rate"],
        "token-efficiency": ["average-compact-tokens", "context-reuse-rate", "deep-execution-rate"],
        "advanced-set-breadth": ["implemented-surface-count", "cross-domain-pack-coverage", "routine-reuse-rate"],
        "learning-maturity": ["experiment-reward", "policy-rollout-win-rate", "rollback-safety"],
        "flexibility": ["domains-supported", "pack-reuse-rate", "kernel-change-avoidance"],
        "governance": ["fresh-evidence-rate", "approval-gap-detection", "replayable-publish-rate"],
        "core-simplicity": ["new-kernel-concepts-avoided", "docs-teachability", "operator-surface-clarity"],
    }
    return signals.get(dimension_id, ["score-delta"])


def framework_cleanup_signals(dimension_id: str) -> list[str]:
    signals = {
        "workflow-rigor": ["manual-steps-removed", "yaml-surgery-reduced", "duplicate-lifecycle-actions-removed"],
        "delivery-maturity": ["manual-decisions-removed", "flags-removed", "time-to-first-output-reduced"],
        "packaging-install": ["bundle-choice-friction-reduced", "kit-adoption-rate", "activation-confusion-reduced"],
        "memory-reliability": ["context-restatement-removed", "memory-prompting-reduced", "stale-memory-surface-reduced"],
        "adapter-portability": ["staged-only-dead-ends-removed", "adapter-fallback-clarified", "portable-edge-friction-reduced"],
        "token-efficiency": ["redundant-context-calls-removed", "default-token-path-simplified", "deep-path-overuse-reduced"],
        "advanced-set-breadth": ["kernel-leakage-removed", "implicit-surface-reduced", "advanced-surface-clarity"],
        "learning-maturity": ["ungoverned-optimization-removed", "rollout-risk-reduced", "policy-drift-surface-reduced"],
        "flexibility": ["domain-specific-kernel-pressure-removed", "special-case-rules-reduced", "pack-reuse-clarity"],
        "governance": ["manual-proof-burden-reduced", "duplicate-approval-steps-removed", "audit-friction-reduced"],
        "core-simplicity": ["kernel-surface-removed", "duplicate-concepts-removed", "docs-teachability-improved"],
    }
    return signals.get(dimension_id, ["friction-removed"])


def framework_cleanup_steps(entry: dict[str, Any]) -> list[str]:
    dimension_id = str(entry.get("id", ""))
    label = str(entry.get("label", "this dimension")).lower()
    steps_by_dimension = {
        "workflow-rigor": [
            "Remove benchmark flows that still require manual YAML repair or duplicate lifecycle actions.",
            "Demote internal-only protocol detail out of default operator guidance when it does not change the next action.",
            "Consolidate overlapping execution and verification guidance so one obvious path remains.",
        ],
        "delivery-maturity": [
            "Remove avoidable flags and manual decisions from the common guided flow by encoding stronger defaults.",
            "Demote repo-specific setup detail out of first-run guidance when it can be inferred automatically.",
            "Collapse overlapping run, refresh, and verify steps into one default operator path where possible.",
        ],
        "packaging-install": [
            "Remove bundle-level choice overload from first-run examples by preferring curated kits.",
            "Demote target-specific detail behind catalog, doctor, and activation outputs instead of requiring manual repo reading.",
            "Consolidate install, verify, and activation guidance so the trust path is obvious from the CLI itself.",
        ],
        "memory-reliability": [
            "Remove manual memory restatement from more flows by auto-attaching bounded memory cues where they already exist.",
            "Demote stale or redundant memory surfaces that do not change execution or verification behavior.",
            "Consolidate overlapping memory summaries into the compact reload path before adding new memory artifacts.",
        ],
        "adapter-portability": [
            "Remove staged-only dead ends where a local-apply or live edge can be the default path.",
            "Demote adapter-specific operator detail out of pack guidance when the registry can carry it.",
            "Consolidate fallback order so one canonical adapter contract remains visible to operators.",
        ],
        "token-efficiency": [
            "Remove redundant context regeneration where existing compact artifacts can be reused safely.",
            "Demote high-token paths from default examples when a bounded local-first path already exists.",
            "Consolidate overlapping resume and handoff surfaces so cheaper reloads become the obvious path.",
        ],
        "advanced-set-breadth": [
            "Remove advanced behavior that leaks into the kernel when it belongs in packs, routines, or adapters.",
            "Demote speculative advanced surfaces that are not yet operationally useful.",
            "Consolidate overlapping advanced docs into clearer packaged surfaces before broadening further.",
        ],
        "learning-maturity": [
            "Remove ungoverned optimization paths that bypass candidate review, approval, or rollback.",
            "Demote telemetry that does not influence bounded policy or framework decisions.",
            "Consolidate learning surfaces so the operator can see one clear path from experiment to outcome.",
        ],
        "flexibility": [
            "Remove domain-specific special cases that pressure the kernel instead of the pack layer.",
            "Demote one-off workflow logic into domain packs before broadening the core story.",
            "Consolidate cross-domain examples so reuse is easier to perceive than novelty.",
        ],
        "governance": [
            "Remove duplicate proof and approval steps that do not improve trust.",
            "Demote heavy governance detail out of the default operator path when automation can surface it just in time.",
            "Consolidate audit surfaces so operators see one clear proof path per transition.",
        ],
        "core-simplicity": [
            "Remove or demote any surface that behaves like a new kernel concept without earning that promotion.",
            "Consolidate overlapping terminology and commands that teach the same idea twice.",
            "Demote advanced implementation detail into packs, adapters, or routines where it belongs.",
        ],
    }
    steps = steps_by_dimension.get(dimension_id)
    if steps:
        return steps
    return [
        f"Remove or demote any surface that adds friction without materially improving {label}.",
        "Consolidate overlapping operator surfaces before adding new ones.",
        "Prefer clearer defaults over broader exposed surface area.",
    ]


def build_framework_cleanup_experiment(entry: dict[str, Any]) -> dict[str, Any]:
    gap = float(entry.get("gap_to_target", 0.0))
    weight = float(entry.get("adoption_weight", framework_adoption_weights().get(str(entry.get("id", "")), 1.0)))
    cleanup_steps = framework_cleanup_steps(entry)
    reward_signals = framework_cleanup_signals(str(entry.get("id", "")))
    expected_delta = round(min(gap, 0.2 if gap > 0 else 0.1), 1)
    return {
        "rank": 1,
        "change_type": "subtractive",
        "title": f"{entry['label']}: remove or demote friction before adding surface",
        "hypothesis": (
            f"If Jini removes, demotes, or consolidates low-value surface area before adding more for "
            f"{entry['label'].lower()}, then adoption should improve because {framework_user_value_reason(str(entry['id'])).lower()}"
        ),
        "build_steps": cleanup_steps,
        "expected_score_delta": expected_delta,
        "success_signals": reward_signals[:3],
        "reward_model": {
            "primary_metric": entry["id"],
            "adoption_weight": weight,
            "expected_score_delta": expected_delta,
            "secondary_signals": reward_signals[:3],
            "change_type": "subtractive",
        },
    }


def build_framework_experiment_candidates(entry: dict[str, Any]) -> list[dict[str, Any]]:
    weight = float(entry.get("adoption_weight", framework_adoption_weights().get(entry["id"], 1.0)))
    reward_signals = framework_reward_signals(str(entry["id"]))
    gap = float(entry.get("gap_to_target", 0.0))
    experiments: list[dict[str, Any]] = [build_framework_cleanup_experiment(entry)]
    for index, step in enumerate(entry.get("next_build_steps", [])[:3], start=2):
        expected_delta = round(min(gap, max(0.1, 0.4 - ((index - 1) * 0.1))), 1)
        experiments.append(
            {
                "rank": index,
                "change_type": "additive",
                "title": f"{entry['label']}: {step}",
                "hypothesis": (
                    f"If Jini {step[0].lower() + step[1:]}, then {entry['label'].lower()} should improve "
                    f"because {framework_user_value_reason(str(entry['id'])).lower()}"
                ),
                "build_steps": [step, *entry.get("next_build_steps", [])[index:index + 2]],
                "expected_score_delta": expected_delta,
                "success_signals": reward_signals[:3],
                "reward_model": {
                    "primary_metric": entry["id"],
                    "adoption_weight": weight,
                    "expected_score_delta": expected_delta,
                    "secondary_signals": reward_signals[:3],
                    "change_type": "additive",
                },
            }
        )
    return experiments


def build_framework_review(
    *,
    dimension: str | None = None,
    limit: int = 5,
) -> tuple[dict[str, Any], Path]:
    scorecard = load_competitive_kpis()
    summary = build_competitive_kpi_summary(scorecard, dimension=dimension, limit=None)
    weights = framework_adoption_weights()
    prioritized: list[dict[str, Any]] = []
    for raw_entry in summary["dimensions"]:
        entry = deepcopy(raw_entry)
        weight = float(weights.get(entry["id"], 1.0))
        base_priority = (
            float(entry.get("gap_to_target", 0.0)) * 1.8
            + float(entry.get("competitive_gap", 0.0)) * 1.2
            + (0.4 if entry.get("strength_status") == "trailing" else 0.0)
        )
        entry["adoption_weight"] = weight
        entry["why_users_care"] = framework_user_value_reason(str(entry["id"]))
        entry["reward_signals"] = framework_reward_signals(str(entry["id"]))
        entry["cleanup_signals"] = framework_cleanup_signals(str(entry["id"]))
        entry["cleanup_candidates"] = framework_cleanup_steps(entry)
        entry["priority_score"] = round(base_priority * weight, 2)
        entry["recommended_experiments"] = build_framework_experiment_candidates(entry)
        prioritized.append(entry)

    prioritized.sort(
        key=lambda item: (-float(item.get("priority_score", 0.0)), int(item.get("execution_order", 999)), item["id"])
    )
    if limit > 0:
        prioritized = prioritized[:limit]

    review_path = next_framework_review_path()
    payload = {
        "schema_version": "0.1.0",
        "review_type": "JiniFrameworkEvolutionReview",
        "generated_at": now_utc(),
        "review_path": display_path(review_path),
        "updated_at": scorecard.get("updated_at"),
        "target_score": float(scorecard.get("target_score", 9.0)),
        "dimension_filter": dimension,
        "adoption_constraints": framework_adoption_constraints(),
        "prioritized_dimensions": prioritized,
        "best_next_dimension": prioritized[0]["id"] if prioritized else "",
        "best_next_experiment": prioritized[0]["recommended_experiments"][0] if prioritized and prioritized[0].get("recommended_experiments") else {},
    }
    review_path.write_text(json.dumps(payload, indent=2) + "\n", encoding="utf-8")
    if prioritized:
        append_learning_event(
            "framework-review",
            {
                "dimension_filter": dimension or "",
                "top_dimension": prioritized[0]["id"],
                "priority_score": prioritized[0]["priority_score"],
                "review_path": display_path(review_path),
            },
        )
    return payload, review_path


def print_framework_review(review: dict[str, Any]) -> None:
    print(f"UPDATED {review.get('updated_at', '')}")
    print(f"REVIEW {review.get('review_path', '')}")
    print(f"TARGET {review.get('target_score', 9.0):.1f}")
    if review.get("best_next_dimension"):
        print(f"NEXT   {review['best_next_dimension']}")
    print("ADOPTION")
    for item in review.get("adoption_constraints", []):
        print(f"  - {item}")
    print("PRIORITIES")
    for entry in review.get("prioritized_dimensions", []):
        strongest = entry.get("strongest_competitor", {})
        print(
            f"  - {entry.get('id', '')} | priority={entry.get('priority_score', 0.0):.2f} | "
            f"cur={entry.get('current_score', 0.0):.1f} | best={strongest.get('name', '')}"
        )
        print(f"    weak: {entry.get('weakness', '')}")
        cleanup_candidates = entry.get("cleanup_candidates", [])
        if cleanup_candidates:
            print(f"    prune: {cleanup_candidates[0]}")
        experiments = entry.get("recommended_experiments", [])
        if experiments:
            print(f"    experiment: {experiments[0].get('title', '')}")


def stage_framework_experiment(
    *,
    review_path: Path | None = None,
    dimension: str | None = None,
    index: int = 1,
) -> tuple[dict[str, Any], Path]:
    selected_review_path = resolve_display_path(str(review_path)) if review_path is not None else latest_framework_review_path()
    if selected_review_path is None or not selected_review_path.exists():
        raise ValueError("No framework review report is available to stage")
    review = load_json_file(selected_review_path)
    if not isinstance(review, dict):
        raise ValueError("Framework review report must be a mapping")
    dimensions = review.get("prioritized_dimensions", [])
    if not isinstance(dimensions, list) or not dimensions:
        raise ValueError("Framework review report has no prioritized dimensions")

    selected_entry: dict[str, Any] | None = None
    if dimension:
        needle = dimension.strip().lower()
        for entry in dimensions:
            if not isinstance(entry, dict):
                continue
            if needle in str(entry.get("id", "")).lower() or needle in str(entry.get("label", "")).lower():
                selected_entry = entry
                break
        if selected_entry is None:
            focused_review, focused_review_path = build_framework_review(dimension=dimension, limit=1)
            selected_review_path = focused_review_path
            dimensions = focused_review.get("prioritized_dimensions", [])
            for entry in dimensions:
                if isinstance(entry, dict):
                    selected_entry = entry
                    break
            if selected_entry is None:
                raise ValueError(f"Dimension {dimension!r} is not present in the selected review")
    else:
        for entry in dimensions:
            if isinstance(entry, dict):
                selected_entry = entry
                break
    if selected_entry is None:
        raise ValueError("No framework dimension is available to stage")

    experiments = selected_entry.get("recommended_experiments", [])
    if not isinstance(experiments, list) or not experiments:
        raise ValueError("Selected dimension does not define recommended experiments")
    if index < 1 or index > len(experiments):
        raise ValueError(f"Experiment index must be between 1 and {len(experiments)}")
    selected_experiment = experiments[index - 1]

    experiment_path = next_framework_experiment_path(str(selected_entry.get("id", "framework")))
    payload = {
        "schema_version": "0.1.0",
        "experiment_type": "JiniFrameworkEvolutionExperiment",
        "generated_at": now_utc(),
        "experiment_id": experiment_path.stem,
        "experiment_path": display_path(experiment_path),
        "status": "proposed",
        "source_review_path": display_path(selected_review_path),
        "dimension_id": selected_entry.get("id", ""),
        "dimension_label": selected_entry.get("label", ""),
        "current_score": float(selected_entry.get("current_score", 0.0)),
        "target_score": float(selected_entry.get("target_score", 0.0)),
        "priority_score": float(selected_entry.get("priority_score", 0.0)),
        "adoption_weight": float(selected_entry.get("adoption_weight", 1.0)),
        "strongest_competitor": selected_entry.get("strongest_competitor", {}),
        "critique": selected_entry.get("weakness", ""),
        "conversion_strategy": selected_entry.get("conversion_strategy", ""),
        "hypothesis": selected_experiment.get("hypothesis", ""),
        "change_type": selected_experiment.get("change_type", "additive"),
        "change_plan": selected_experiment.get("build_steps", []),
        "expected_score_delta": float(selected_experiment.get("expected_score_delta", 0.0)),
        "reward_model": selected_experiment.get("reward_model", {}),
        "success_signals": selected_experiment.get("success_signals", []),
        "exit_criteria": selected_entry.get("exit_criteria", [])[:3],
        "implemented_assets": selected_entry.get("implemented_assets", []),
    }
    experiment_path.write_text(json.dumps(payload, indent=2) + "\n", encoding="utf-8")
    append_learning_event(
        "framework-experiment-staged",
        {
            "experiment_id": payload["experiment_id"],
            "dimension_id": payload["dimension_id"],
            "priority_score": payload["priority_score"],
            "expected_score_delta": payload["expected_score_delta"],
            "source_review_path": payload["source_review_path"],
        },
    )
    return payload, experiment_path


def print_framework_experiment(experiment: dict[str, Any]) -> None:
    print(f"EXPERIMENT {experiment.get('experiment_id', '')}")
    print(f"DIM       {experiment.get('dimension_id', '')}")
    print(f"TYPE      {experiment.get('change_type', 'additive')}")
    print(f"STATUS    {experiment.get('status', '')}")
    print(f"CURRENT   {experiment.get('current_score', 0.0):.1f}")
    print(f"TARGET    {experiment.get('target_score', 0.0):.1f}")
    print(f"EXPECTED  +{experiment.get('expected_score_delta', 0.0):.1f}")
    print(f"REVIEW    {experiment.get('source_review_path', '')}")
    print(f"HYPOTHESIS {experiment.get('hypothesis', '')}")
    for step in experiment.get("change_plan", []):
        print(f"  - {step}")


def record_framework_experiment_outcome(
    experiment_path: Path,
    *,
    actor: str,
    result: str,
    score_delta: float,
    adoption_signals: list[str] | None = None,
    notes: list[str] | None = None,
) -> tuple[dict[str, Any], Path]:
    experiment_path = resolve_display_path(str(experiment_path))
    experiment = load_json_file(experiment_path)
    if not isinstance(experiment, dict):
        raise ValueError("Framework experiment must be a mapping")
    signals = [item.strip() for item in (adoption_signals or []) if item and item.strip()]
    normalized_notes = [item.strip() for item in (notes or []) if item and item.strip()]
    reward_model = experiment.get("reward_model", {})
    weight = float(reward_model.get("adoption_weight", experiment.get("adoption_weight", 1.0)) or 1.0)
    signal_bonus = min(len(signals) * 0.03, 0.15)
    if result == "success":
        computed_reward = round((score_delta * weight) + signal_bonus, 3)
    elif result == "mixed":
        computed_reward = round((score_delta * weight * 0.5) + (signal_bonus * 0.5), 3)
    else:
        computed_reward = round(min(0.0, (score_delta * weight) - 0.2), 3)

    outcome_path = next_framework_outcome_path(str(experiment.get("dimension_id", "framework")))
    outcome = {
        "schema_version": "0.1.0",
        "outcome_type": "JiniFrameworkEvolutionOutcome",
        "generated_at": now_utc(),
        "outcome_id": outcome_path.stem,
        "outcome_path": display_path(outcome_path),
        "experiment_id": experiment.get("experiment_id", ""),
        "dimension_id": experiment.get("dimension_id", ""),
        "change_type": experiment.get("change_type", "additive"),
        "result": result,
        "actor": actor,
        "score_delta": score_delta,
        "adoption_signals": signals,
        "notes": normalized_notes,
        "computed_reward": computed_reward,
        "source_experiment_path": display_path(experiment_path),
    }
    outcome_path.write_text(json.dumps(outcome, indent=2) + "\n", encoding="utf-8")
    experiment["status"] = "completed"
    experiment["completed_at"] = outcome["generated_at"]
    experiment["completed_by"] = actor
    experiment["latest_outcome_path"] = display_path(outcome_path)
    experiment["latest_outcome_reward"] = computed_reward
    experiment_path.write_text(json.dumps(experiment, indent=2) + "\n", encoding="utf-8")
    append_learning_event(
        "framework-experiment-recorded",
        {
            "experiment_id": experiment.get("experiment_id", ""),
            "dimension_id": experiment.get("dimension_id", ""),
            "result": result,
            "score_delta": score_delta,
            "computed_reward": computed_reward,
        },
    )
    return outcome, outcome_path


def print_framework_outcome(outcome: dict[str, Any]) -> None:
    print(f"OUTCOME  {outcome.get('outcome_id', '')}")
    print(f"EXPERIMENT {outcome.get('experiment_id', '')}")
    print(f"DIM      {outcome.get('dimension_id', '')}")
    print(f"RESULT   {outcome.get('result', '')}")
    print(f"DELTA    {float(outcome.get('score_delta', 0.0)):+.1f}")
    print(f"REWARD   {float(outcome.get('computed_reward', 0.0)):+.3f}")
    for signal in outcome.get("adoption_signals", []):
        print(f"  - signal: {signal}")


def build_framework_evolution_backtest(
    *,
    limit: int = 100,
) -> dict[str, Any]:
    outcomes = sorted(framework_outcome_dir().glob("*.json"))
    if limit > 0:
        outcomes = outcomes[-limit:]
    per_dimension: dict[str, dict[str, Any]] = {}
    for path in outcomes:
        try:
            outcome = load_json_file(path)
        except (OSError, json.JSONDecodeError):
            continue
        if not isinstance(outcome, dict):
            continue
        dimension_id = str(outcome.get("dimension_id", "")).strip() or "unknown"
        bucket = per_dimension.setdefault(
            dimension_id,
            {
                "dimension_id": dimension_id,
                "experiments": 0,
                "subtractive": 0,
                "additive": 0,
                "successes": 0,
                "mixed": 0,
                "failed": 0,
                "score_deltas": [],
                "rewards": [],
            },
        )
        bucket["experiments"] += 1
        change_type = str(outcome.get("change_type", "additive")).strip() or "additive"
        if change_type == "subtractive":
            bucket["subtractive"] += 1
        else:
            bucket["additive"] += 1
        status = str(outcome.get("result", "")).strip()
        if status == "success":
            bucket["successes"] += 1
        elif status == "mixed":
            bucket["mixed"] += 1
        else:
            bucket["failed"] += 1
        bucket["score_deltas"].append(float(outcome.get("score_delta", 0.0)))
        bucket["rewards"].append(float(outcome.get("computed_reward", 0.0)))

    scorecard = load_competitive_kpis()
    review, _review_path = build_framework_review(limit=0)
    priority_lookup = {
        item["id"]: item
        for item in review.get("prioritized_dimensions", [])
        if isinstance(item, dict)
    }
    dimension_summaries: list[dict[str, Any]] = []
    for dimension_id, bucket in per_dimension.items():
        deltas = bucket.pop("score_deltas")
        rewards = bucket.pop("rewards")
        experiments = int(bucket["experiments"])
        avg_delta = round(sum(deltas) / experiments, 3) if experiments else 0.0
        avg_reward = round(sum(rewards) / experiments, 3) if experiments else 0.0
        base_priority = float(priority_lookup.get(dimension_id, {}).get("priority_score", 0.0))
        learned_priority = round(max(0.0, base_priority - avg_reward), 2) if avg_reward > 0 else round(base_priority + abs(avg_reward), 2)
        dimension_summaries.append(
            {
                **bucket,
                "average_score_delta": avg_delta,
                "average_reward": avg_reward,
                "base_priority": base_priority,
                "learned_priority": learned_priority,
            }
        )

    dimension_summaries.sort(key=lambda item: (-float(item.get("learned_priority", 0.0)), item["dimension_id"]))
    recommended = ""
    if dimension_summaries:
        recommended = dimension_summaries[0]["dimension_id"]
    elif review.get("prioritized_dimensions"):
        recommended = review["prioritized_dimensions"][0]["id"]

    payload = {
        "schema_version": "0.1.0",
        "report_type": "JiniFrameworkEvolutionBacktest",
        "generated_at": now_utc(),
        "updated_at": scorecard.get("updated_at"),
        "outcome_count": len(outcomes),
        "dimension_summaries": dimension_summaries,
        "recommended_next_focus": recommended,
    }
    append_learning_event(
        "framework-evolution-backtest",
        {
            "outcome_count": len(outcomes),
            "recommended_next_focus": recommended,
        },
    )
    return payload


def print_framework_evolution_backtest(backtest: dict[str, Any]) -> None:
    print(f"UPDATED {backtest.get('updated_at', '')}")
    print(f"OUTCOMES {backtest.get('outcome_count', 0)}")
    if backtest.get("recommended_next_focus"):
        print(f"NEXT    {backtest['recommended_next_focus']}")
    print("DIMENSIONS")
    for item in backtest.get("dimension_summaries", []):
        print(
            f"  - {item.get('dimension_id', '')} | experiments={item.get('experiments', 0)} | "
            f"success={item.get('successes', 0)} | avg_delta={item.get('average_score_delta', 0.0):+.3f} | "
            f"avg_reward={item.get('average_reward', 0.0):+.3f}"
        )


def load_install_manifest() -> dict[str, Any]:
    manifest = load_document(INSTALL_MANIFEST_PATH)
    if not isinstance(manifest, dict):
        raise ValueError("install manifest must be a mapping")
    bundles = manifest.get("bundles")
    targets = manifest.get("targets")
    kits = manifest.get("kits", [])
    if not isinstance(bundles, list) or not bundles:
        raise ValueError("install manifest must define non-empty bundles")
    if not isinstance(targets, list) or not targets:
        raise ValueError("install manifest must define non-empty targets")
    if kits and not isinstance(kits, list):
        raise ValueError("install manifest kits must be a list when provided")
    default_kit_id = manifest.get("default_kit_id", "")
    if default_kit_id:
        if not isinstance(default_kit_id, str):
            raise ValueError("install manifest default_kit_id must be a string when provided")
        valid_kit_ids = {
            entry.get("id")
            for entry in kits
            if isinstance(entry, dict) and isinstance(entry.get("id"), str)
        }
        if default_kit_id not in valid_kit_ids:
            raise ValueError("install manifest default_kit_id must reference a declared kit")
    return manifest


def build_install_catalog(
    *,
    target: str | None = None,
    kind: str | None = None,
) -> dict[str, Any]:
    manifest = load_install_manifest()
    source = manifest.get("source", {}) if isinstance(manifest.get("source", {}), dict) else {}
    target_filter = target.strip() if target else ""
    kind_filter = kind.strip().lower() if kind else ""
    bundles: list[dict[str, Any]] = []
    for raw_bundle in manifest.get("bundles", []):
        if not isinstance(raw_bundle, dict):
            continue
        compatible_targets = raw_bundle.get("compatible_targets", [])
        if not isinstance(compatible_targets, list):
            compatible_targets = []
        bundle_kind = str(raw_bundle.get("kind", "bundle"))
        if target_filter and target_filter not in compatible_targets:
            continue
        if kind_filter and bundle_kind.lower() != kind_filter:
            continue
        bundles.append(
            {
                "id": raw_bundle.get("id", ""),
                "label": raw_bundle.get("label", raw_bundle.get("id", "")),
                "kind": bundle_kind,
                "version": str(raw_bundle.get("version", manifest.get("version", "0.1.0"))),
                "summary": raw_bundle.get("summary", ""),
                "permission_risk": raw_bundle.get("permission_risk", "unknown"),
                "review_before_use": bool(raw_bundle.get("review_before_use", False)),
                "compatible_targets": compatible_targets,
                "depends_on": raw_bundle.get("depends_on", []),
                "activation_steps": raw_bundle.get("activation_steps", []),
                "migration_notes": raw_bundle.get("migration_notes", []),
                "universal_paths": expand_manifest_paths(raw_bundle.get("universal_paths", [])),
            }
        )
    kits: list[dict[str, Any]] = []
    for raw_kit in manifest.get("kits", []):
        if not isinstance(raw_kit, dict):
            continue
        recommended_targets = raw_kit.get("recommended_targets", [])
        if not isinstance(recommended_targets, list):
            recommended_targets = []
        if target_filter and target_filter not in recommended_targets:
            continue
        kits.append(
            {
                "id": raw_kit.get("id", ""),
                "label": raw_kit.get("label", raw_kit.get("id", "")),
                "audience": str(raw_kit.get("audience", "")).strip(),
                "summary": raw_kit.get("summary", ""),
                "bundle_ids": raw_kit.get("bundle_ids", []),
                "recommended_targets": recommended_targets,
            }
        )
    selected_target = infer_get_started_target(manifest, preferred_target=target_filter)
    beginner_kit = find_install_kit(
        manifest,
        kit_id=str(manifest.get("default_kit_id", "")).strip() or None,
        audience="beginner",
    )
    power_kit = find_install_kit(manifest, audience="power-user")
    recommended_paths: dict[str, Any] = {}
    cli = cli_invocation()
    if beginner_kit is not None:
        beginner_kit_id = str(beginner_kit.get("id", "")).strip()
        recommended_paths["beginner"] = {
            "kit_id": beginner_kit_id,
            "target": selected_target,
            "commands": [
                f"{cli} start --harness {selected_target}",
                *[item["command"] for item in build_install_trust_path(target_id=selected_target, kit_id=beginner_kit_id)],
            ],
        }
    if power_kit is not None:
        power_kit_id = str(power_kit.get("id", "")).strip()
        recommended_paths["power_user"] = {
            "kit_id": power_kit_id,
            "target": selected_target,
            "commands": [
                f"{cli} catalog-bundles --target {selected_target} --format json",
                f"{cli} plan-install --kit {power_kit_id} --harness {selected_target}",
                f"{cli} show-adapters",
            ],
        }
    return {
        "schema_version": "0.1.0",
        "catalog_type": "JiniInstallCatalog",
        "generated_at": now_utc(),
        "manifest_version": str(manifest.get("version", "0.1.0")),
        "default_kit_id": str(manifest.get("default_kit_id", "")),
        "source": {
            "name": str(source.get("name", "Jini")),
            "revision": str(source.get("revision", "workspace")),
            "trust_notice": str(source.get("trust_notice", "")).strip(),
        },
        "target_filter": target_filter,
        "kind_filter": kind_filter,
        "recommended_paths": recommended_paths,
        "targets": manifest.get("targets", []),
        "kits": kits,
        "bundles": bundles,
    }


def print_install_catalog(catalog: dict[str, Any]) -> None:
    print(f"SOURCE  {catalog['source']['name']}")
    print(f"REV     {catalog['source']['revision']}")
    print(f"MANIFEST {catalog['manifest_version']}")
    if catalog.get("default_kit_id"):
        print(f"DEFAULT {catalog['default_kit_id']}")
    if catalog["source"].get("trust_notice"):
        print(f"NOTICE  {catalog['source']['trust_notice']}")
    beginner_path = catalog.get("recommended_paths", {}).get("beginner", {})
    if beginner_path:
        print("START")
        print(f"  target: {beginner_path.get('target', '')}")
        print(f"  kit: {beginner_path.get('kit_id', '')}")
        for command in beginner_path.get("commands", []):
            print(f"    {command}")
    print("KITS")
    for kit in catalog.get("kits", []):
        targets = ", ".join(kit.get("recommended_targets", []))
        audience = kit.get("audience", "")
        suffix = f" | audience={audience}" if audience else ""
        print(f"  - {kit['id']} | bundles={len(kit.get('bundle_ids', []))} | targets={targets}{suffix}")
    power_path = catalog.get("recommended_paths", {}).get("power_user", {})
    if power_path:
        print("ADVANCED")
        for command in power_path.get("commands", []):
            print(f"    {command}")
    print("BUNDLES")
    print(
        f"  {len(catalog.get('bundles', []))} bundle(s) available. "
        "Use `--format json` when you need bundle-level detail."
    )


def build_harness_catalog() -> dict[str, Any]:
    manifest = load_install_manifest()
    cli = cli_invocation()
    harnesses: list[dict[str, Any]] = []
    for raw_target in manifest.get("targets", []):
        if not isinstance(raw_target, dict):
            continue
        harness_id = str(raw_target.get("id", "")).strip()
        if not harness_id:
            continue
        label = str(raw_target.get("label", harness_id)).strip() or harness_id
        harnesses.append(
            {
                "id": harness_id,
                "label": label,
                "destination_root": str(raw_target.get("destination_root", "")).strip(),
                "risk_notice": str(raw_target.get("risk_notice", "")).strip(),
                "start_command": f"{cli} start --harness {harness_id}",
                "run_command": f"{cli} execute-flow /path/to/work --repo /path/to/repo --harness {harness_id}",
            }
        )
    return {
        "schema_version": "0.1.0",
        "catalog_type": "JiniHarnessCatalog",
        "generated_at": now_utc(),
        "summary": "Bring your own harness. Jini keeps state, artifacts, and next steps coherent above it.",
        "harnesses": harnesses,
    }


def print_harness_catalog(report: dict[str, Any]) -> None:
    print("Jini works above coding harnesses.")
    print("Use any harness to execute. Jini keeps the work, state, and next steps coherent.")
    print()
    print("HARNESSES")
    for item in report.get("harnesses", []):
        print(f"  - {item.get('label', '')} ({item.get('id', '')})")
        if item.get("risk_notice"):
            print(f"    note:  {item.get('risk_notice', '')}")
        print(f"    start: {item.get('start_command', '')}")
        print(f"    run:   {item.get('run_command', '')}")


def build_publish_readiness() -> dict[str, Any]:
    pack_catalog = build_pack_catalog()
    install_catalog = build_install_catalog()
    scorecard = build_competitive_kpi_summary(load_competitive_kpis())

    doc_paths = [
        ROOT / "README.md",
        ROOT / "WHITEPAPER.md",
        ROOT / "specs" / "learning-system.md",
        ROOT / "specs" / "install-packaging.md",
        ROOT / "specs" / "competitive-kpis.yaml",
        ROOT / "specs" / "golden-competitive-benchmark.yaml",
        ROOT / "distribution" / "install-manifest.yaml",
    ]
    doc_checks = [{"path": display_path(path), "exists": path.exists()} for path in doc_paths]

    install_manifest = load_install_manifest()
    target_checks: list[dict[str, Any]] = []
    for target in install_manifest.get("targets", []):
        if not isinstance(target, dict):
            continue
        target_id = str(target.get("id", "")).strip()
        if not target_id:
            continue
        readme_path = ROOT / "distribution" / "targets" / target_id / "README.md"
        doctor_expectations = target.get("doctor_expectations", {}) if isinstance(target.get("doctor_expectations", {}), dict) else {}
        readme_markers = [str(marker) for marker in doctor_expectations.get("readme_markers", []) if str(marker).strip()]
        smoke_commands = doctor_expectations.get("smoke_commands", [])
        smoke_command_count = len(smoke_commands) if isinstance(smoke_commands, list) else 0
        readme_text = readme_path.read_text(encoding="utf-8") if readme_path.exists() else ""
        missing_markers = [marker for marker in readme_markers if marker not in readme_text]
        target_checks.append(
            {
                "target_id": target_id,
                "readme_path": display_path(readme_path),
                "readme_exists": readme_path.exists(),
                "doctor_markers": readme_markers,
                "smoke_command_count": smoke_command_count,
                "missing_markers": missing_markers,
                "status": "ok" if readme_path.exists() and not missing_markers and smoke_command_count > 0 else "warning",
            }
        )

    pack_checks: list[dict[str, Any]] = []
    for entry in pack_catalog.get("packs", []):
        pack_id = str(entry.get("pack_id", "")).strip()
        if not pack_id:
            continue
        pack_root = ROOT / "packs" / pack_id
        readme_path = pack_root / "README.md"
        context_path = pack_root / "context" / "benchmark-context.md"
        pack_checks.append(
            {
                "pack_id": pack_id,
                "target_profile": str(entry.get("target_profile", "")),
                "readme_exists": readme_path.exists(),
                "context_exists": context_path.exists(),
                "status": "ok" if readme_path.exists() and context_path.exists() else "warning",
            }
        )

    front_door_doc_paths = [
        ROOT / "README.md",
        ROOT / "docs" / "index.md",
        ROOT / "docs" / "install.md",
        ROOT / "docs" / "cli.md",
        ROOT / "docs" / "simple.md",
    ]
    forbidden_novice_patterns = [
        "python3 tools/jini.py",
        "python3 tools/jini_validate.py",
    ]
    front_door_checks: list[dict[str, Any]] = []
    for path in front_door_doc_paths:
        text = path.read_text(encoding="utf-8") if path.exists() else ""
        limited_text = "\n".join(text.splitlines()[:220])
        matches = [pattern for pattern in forbidden_novice_patterns if pattern in limited_text]
        front_door_checks.append(
            {
                "path": display_path(path),
                "exists": path.exists(),
                "forbidden_matches": matches,
                "status": "ok" if path.exists() and not matches else "warning",
            }
        )

    novice_guide = build_get_started_guide(target="codex", audience="beginner")
    novice_commands = novice_guide.get("beginner_path", {}).get("commands", [])
    if not isinstance(novice_commands, list):
        novice_commands = []
    novice_checks = [
        {
            "id": "beginner-command-count",
            "command_count": len(novice_commands),
            "max_allowed": 4,
            "status": "ok" if 0 < len(novice_commands) <= 4 else "warning",
        },
        {
            "id": "beginner-command-prefix",
            "all_jini_commands": all(str(command).startswith("jini ") for command in novice_commands),
            "status": "ok" if novice_commands and all(str(command).startswith("jini ") for command in novice_commands) else "warning",
        },
        {
            "id": "beginner-proof-command",
            "present": any(
                token in str(command)
                for command in novice_commands
                for token in ("try-example research-prd", "example research-prd")
            ),
            "status": "ok"
            if any(
                token in str(command)
                for command in novice_commands
                for token in ("try-example research-prd", "example research-prd")
            )
            else "warning",
        },
        {
            "id": "beginner-no-bundle-catalog",
            "present": any("catalog-bundles" in str(command) for command in novice_commands),
            "status": "ok" if not any("catalog-bundles" in str(command) for command in novice_commands) else "warning",
        },
        {
            "id": "cli-guide-no-python-requirement",
            "present": "You do not need to know Python to use Jini." in ((ROOT / "docs" / "cli.md").read_text(encoding="utf-8") if (ROOT / "docs" / "cli.md").exists() else ""),
            "status": "ok" if "You do not need to know Python to use Jini." in ((ROOT / "docs" / "cli.md").read_text(encoding="utf-8") if (ROOT / "docs" / "cli.md").exists() else "") else "warning",
        },
        {
            "id": "readme-plain-words-entry",
            "present": "**In plain words:**" in ((ROOT / "README.md").read_text(encoding="utf-8") if (ROOT / "README.md").exists() else ""),
            "status": "ok" if "**In plain words:**" in ((ROOT / "README.md").read_text(encoding="utf-8") if (ROOT / "README.md").exists() else "") else "warning",
        },
        {
            "id": "homepage-plain-words-entry",
            "present": "**In plain words:**" in ((ROOT / "docs" / "index.md").read_text(encoding="utf-8") if (ROOT / "docs" / "index.md").exists() else ""),
            "status": "ok" if "**In plain words:**" in ((ROOT / "docs" / "index.md").read_text(encoding="utf-8") if (ROOT / "docs" / "index.md").exists() else "") else "warning",
        },
        {
            "id": "simple-guide-exists",
            "present": (ROOT / "docs" / "simple.md").exists(),
            "status": "ok" if (ROOT / "docs" / "simple.md").exists() else "warning",
        },
        {
            "id": "simple-guide-core-questions",
            "present": all(
                phrase in ((ROOT / "docs" / "simple.md").read_text(encoding="utf-8") if (ROOT / "docs" / "simple.md").exists() else "")
                for phrase in ("What is done?", "What happens next?", "What is still missing?")
            ),
            "status": "ok"
            if all(
                phrase in ((ROOT / "docs" / "simple.md").read_text(encoding="utf-8") if (ROOT / "docs" / "simple.md").exists() else "")
                for phrase in ("What is done?", "What happens next?", "What is still missing?")
            )
            else "warning",
        },
    ]

    required_dimensions = {
        "workflow-rigor": 9.0,
        "delivery-maturity": 8.8,
        "packaging-install": 8.5,
        "advanced-set-breadth": 8.8,
    }
    dimension_lookup = {
        str(item.get("id", "")): item
        for item in scorecard.get("dimensions", [])
        if isinstance(item, dict)
    }
    score_checks: list[dict[str, Any]] = []
    for dimension_id, threshold in required_dimensions.items():
        item = dimension_lookup.get(dimension_id)
        current_score = float(item.get("current_score", 0.0)) if item else 0.0
        score_checks.append(
            {
                "dimension_id": dimension_id,
                "current_score": current_score,
                "threshold": threshold,
                "status": "ok" if current_score >= threshold else "warning",
            }
        )

    leadership_checks: list[dict[str, Any]] = []
    for item in scorecard.get("dimensions", []):
        if not isinstance(item, dict) or not item.get("leadership_guard"):
            continue
        dimension_id = str(item.get("id", "")).strip()
        competitor = item.get("strongest_competitor", {}) if isinstance(item.get("strongest_competitor", {}), dict) else {}
        competitor_name = str(competitor.get("name", "")).strip()
        competitor_score = float(competitor.get("score", 0.0)) if competitor else 0.0
        current_score = float(item.get("current_score", 0.0))
        margin = round(current_score - competitor_score, 2)
        leadership_checks.append(
            {
                "dimension_id": dimension_id,
                "current_score": current_score,
                "competitor": competitor_name,
                "competitor_score": competitor_score,
                "margin": margin,
                "position": "ahead" if margin > 0 else "tied" if margin == 0 else "behind",
                "status": "ok" if current_score >= competitor_score else "warning",
            }
        )

    commercial_kit = next((kit for kit in install_catalog.get("kits", []) if kit.get("id") == "vendor-decision-kit"), None)
    breadth_summary = {
        "pack_count": int(pack_catalog.get("pack_count", 0)),
        "target_profiles": sorted(
            {
                str(item.get("target_profile", ""))
                for item in pack_catalog.get("packs", [])
                if str(item.get("target_profile", "")).strip()
            }
        ),
        "kit_count": len(install_catalog.get("kits", [])),
        "commercial_kit_present": commercial_kit is not None,
        "status": "ok"
        if int(pack_catalog.get("pack_count", 0)) >= 6 and commercial_kit is not None
        else "warning",
    }

    sections = [
        {
            "id": "docs",
            "label": "Public docs",
            "status": "ok" if all(item["exists"] for item in doc_checks) else "warning",
            "checks": doc_checks,
        },
        {
            "id": "install",
            "label": "Install surface",
            "status": "ok" if all(item["status"] == "ok" for item in target_checks) else "warning",
            "checks": target_checks,
        },
        {
            "id": "novice",
            "label": "Novice usability",
            "status": "ok"
            if all(item["status"] == "ok" for item in front_door_checks) and all(item["status"] == "ok" for item in novice_checks)
            else "warning",
            "checks": [
                *front_door_checks,
                *novice_checks,
            ],
        },
        {
            "id": "breadth",
            "label": "Advanced surface breadth",
            "status": breadth_summary["status"],
            "checks": [breadth_summary, *pack_checks],
        },
        {
            "id": "scores",
            "label": "Publishing score gates",
            "status": "ok" if all(item["status"] == "ok" for item in score_checks) else "warning",
            "checks": score_checks,
        },
        {
            "id": "leadership",
            "label": "Lead preservation gates",
            "status": "ok" if all(item["status"] == "ok" for item in leadership_checks) else "warning",
            "checks": leadership_checks,
        },
    ]
    overall_status = "ok" if all(section["status"] == "ok" for section in sections) else "warning"

    return {
        "schema_version": "0.1.0",
        "result_type": "JiniPublishReadiness",
        "generated_at": now_utc(),
        "status": overall_status,
        "default_kit_id": install_catalog.get("default_kit_id", ""),
        "pack_count": int(pack_catalog.get("pack_count", 0)),
        "bundle_count": len(install_catalog.get("bundles", [])),
        "kit_count": len(install_catalog.get("kits", [])),
        "target_count": len(install_catalog.get("targets", [])),
        "sections": sections,
    }


def print_publish_readiness(report: dict[str, Any]) -> None:
    print(f"STATUS  {report.get('status', 'unknown')}")
    print(f"PACKS   {report.get('pack_count', 0)}")
    print(f"BUNDLES {report.get('bundle_count', 0)}")
    print(f"KITS    {report.get('kit_count', 0)}")
    print(f"TARGETS {report.get('target_count', 0)}")
    if report.get("default_kit_id"):
        print(f"DEFAULT {report['default_kit_id']}")
    print("CHECKS")
    for section in report.get("sections", []):
        print(f"  - {section['label']}: {section['status']}")


def resolve_json_path(document: dict[str, Any], dotted_path: str) -> Any:
    current: Any = document
    for part in dotted_path.split("."):
        if isinstance(current, list):
            if not part.isdigit():
                raise KeyError(dotted_path)
            index = int(part)
            if index < 0 or index >= len(current):
                raise KeyError(dotted_path)
            current = current[index]
            continue
        if not isinstance(current, dict) or part not in current:
            raise KeyError(dotted_path)
        current = current[part]
    return current


def run_install_smoke_probe(
    cli_path: Path,
    universal_root: Path,
    target_id: str,
    probe: dict[str, Any],
) -> dict[str, Any]:
    probe_id = str(probe.get("id", "smoke")).strip() or "smoke"
    raw_args = probe.get("args", [])
    if not isinstance(raw_args, list) or not raw_args:
        return {
            "id": f"smoke-{probe_id}",
            "status": "warning",
            "detail": "Smoke probe is missing command arguments.",
        }
    args = [str(item).format(target_id=target_id) for item in raw_args]
    command = [sys.executable, str(cli_path), *args]
    try:
        completed = subprocess.run(
            command,
            cwd=universal_root,
            check=False,
            capture_output=True,
            text=True,
            timeout=20,
        )
    except (OSError, subprocess.TimeoutExpired) as exc:
        return {
            "id": f"smoke-{probe_id}",
            "status": "warning",
            "detail": f"Smoke probe failed to execute: {exc}",
        }
    if completed.returncode != 0:
        return {
            "id": f"smoke-{probe_id}",
            "status": "warning",
            "detail": f"Smoke probe exited with code {completed.returncode}: {trim_output(completed.stderr or completed.stdout)}",
        }
    expectations = probe.get("json_expectations", {})
    if isinstance(expectations, dict) and expectations:
        try:
            payload = json.loads(completed.stdout)
        except json.JSONDecodeError:
            return {
                "id": f"smoke-{probe_id}",
                "status": "warning",
                "detail": "Smoke probe did not emit valid JSON output.",
            }
        mismatches: list[str] = []
        for key, expected_value in expectations.items():
            expected = str(expected_value).format(target_id=target_id)
            try:
                actual = resolve_json_path(payload, str(key))
            except KeyError:
                mismatches.append(f"{key}: missing")
                continue
            if str(actual) != expected:
                mismatches.append(f"{key}: expected `{expected}` got `{actual}`")
        if mismatches:
            return {
                "id": f"smoke-{probe_id}",
                "status": "warning",
                "detail": f"Smoke probe output mismatch: {', '.join(mismatches)}.",
            }
    return {
        "id": f"smoke-{probe_id}",
        "status": "ok",
        "detail": f"Smoke probe `{probe_id}` executed successfully for `{target_id}`.",
    }


def expand_manifest_paths(paths: list[str]) -> list[str]:
    expanded: list[str] = []
    for raw_path in paths:
        path = ROOT / raw_path
        expanded.append(display_path(path.resolve() if path.exists() else path))
    return expanded


def infer_get_started_target(manifest: dict[str, Any], preferred_target: str | None = None) -> str:
    if preferred_target and preferred_target.strip():
        return preferred_target.strip()
    default_kit_id = str(manifest.get("default_kit_id", "")).strip()
    for raw_kit in manifest.get("kits", []):
        if not isinstance(raw_kit, dict):
            continue
        if str(raw_kit.get("id", "")).strip() != default_kit_id:
            continue
        recommended_targets = raw_kit.get("recommended_targets", [])
        if isinstance(recommended_targets, list):
            for item in recommended_targets:
                if isinstance(item, str) and item.strip():
                    return item.strip()
    for raw_target in manifest.get("targets", []):
        if isinstance(raw_target, dict):
            target_id = str(raw_target.get("id", "")).strip()
            if target_id:
                return target_id
    return "codex"


def build_install_trust_path(
    *,
    target_id: str,
    kit_id: str,
    prefix: str = "/tmp/jini-stage",
) -> list[dict[str, str]]:
    cli = cli_invocation()
    return [
        {
            "id": "preview",
            "label": "Preview curated install",
            "command": f"{cli} plan-install --kit {kit_id} --harness {target_id}",
        },
        {
            "id": "install",
            "label": "Install curated bundle set",
            "command": f"{cli} install-bundles --kit {kit_id} --harness {target_id} --prefix {prefix}",
        },
        {
            "id": "verify",
            "label": "Verify trust and activation readiness",
            "command": f"{cli} doctor-install --kit {kit_id} --harness {target_id} --prefix {prefix}",
        },
    ]


def find_install_kit(manifest: dict[str, Any], *, kit_id: str | None = None, audience: str | None = None) -> dict[str, Any] | None:
    kits = manifest.get("kits", [])
    if kit_id:
        for raw_kit in kits:
            if isinstance(raw_kit, dict) and str(raw_kit.get("id", "")).strip() == kit_id:
                return raw_kit
    normalized_audience = (audience or "").strip().lower()
    if normalized_audience:
        for raw_kit in kits:
            if not isinstance(raw_kit, dict):
                continue
            if str(raw_kit.get("audience", "")).strip().lower() == normalized_audience:
                return raw_kit
    return None


def build_get_started_guide(
    *,
    target: str | None = None,
    audience: str | None = None,
) -> dict[str, Any]:
    manifest = load_install_manifest()
    selected_target = infer_get_started_target(manifest, preferred_target=target)
    beginner_kit = find_install_kit(
        manifest,
        kit_id=str(manifest.get("default_kit_id", "")).strip() or None,
        audience="beginner",
    )
    power_kit = find_install_kit(manifest, audience="power-user")
    if beginner_kit is None:
        raise ValueError("install manifest must define a beginner-compatible default kit")
    if power_kit is None:
        raise ValueError("install manifest must define a power-user kit")

    selected_audience = (audience or "both").strip().lower()
    if selected_audience not in {"beginner", "power-user", "both"}:
        raise ValueError("audience must be one of: beginner, power-user, both")

    beginner_kit_id = str(beginner_kit.get("id", "")).strip()
    power_kit_id = str(power_kit.get("id", "")).strip()
    beginner_trust_path = build_install_trust_path(
        target_id=selected_target,
        kit_id=beginner_kit_id,
    )
    cli = cli_invocation()
    beginner_path = {
        "audience": "beginner",
        "label": "Beginner Path",
        "goal": "Reach a safe first success with the smallest guided surface.",
        "kit_id": beginner_kit_id,
        "target": selected_target,
        "commands": [
            *[item["command"] for item in beginner_trust_path],
            f"{cli} example research-prd",
        ],
        "trust_path": beginner_trust_path,
        "notes": [
            "Bundle-level detail is intentionally demoted from the beginner path.",
            "The trust path is preview, install, verify, then prove value on one common example.",
        ],
    }
    power_path = {
        "audience": "power-user",
        "label": "Power-User Path",
        "goal": "Open the same system up for inspection, composition, and deeper execution control.",
        "kit_id": power_kit_id,
        "target": selected_target,
        "commands": [
            f"{cli} catalog-bundles --format json",
            f"{cli} plan-install --kit {power_kit_id} --harness {selected_target}",
            f"{cli} catalog-packs",
            f"{cli} show-adapters",
            f"{cli} review-framework --format json --limit 5",
            f"{cli} execute-flow /tmp/my-pack --repo /path/to/repo --harness codex --activate-runtime --consent write --consent publish",
        ],
        "notes": [
            "Use JSON outputs when you want bundle-by-bundle or adapter-level detail.",
        ],
    }
    guide = {
        "schema_version": "0.1.0",
        "guide_type": "JiniGettingStartedGuide",
        "generated_at": now_utc(),
        "audience": selected_audience,
        "target": selected_target,
        "shared_model": [
            "Same kernel and lifecycle for both audiences.",
            "Beginners get the smallest safe path first.",
            "Power users get the same system with more inspectable depth, not a different framework.",
        ],
        "beginner_path": beginner_path,
        "power_user_path": power_path,
        "recommended_path": beginner_path if selected_audience != "power-user" else power_path,
    }
    return guide


def print_get_started_guide(guide: dict[str, Any]) -> None:
    print(f"HARNESS {guide.get('target', '')}")
    print(f"AUDIENCE {guide.get('audience', 'both')}")
    print("SHARED")
    for item in guide.get("shared_model", []):
        print(f"  - {item}")
    if guide.get("audience") in {"beginner", "both"}:
        beginner = guide.get("beginner_path", {})
        print("BEGINNER")
        print(f"  kit: {beginner.get('kit_id', '')}")
        print(f"  goal: {beginner.get('goal', '')}")
        for command in beginner.get("commands", []):
            print(f"    {command}")
    if guide.get("audience") in {"power-user", "both"}:
        power = guide.get("power_user_path", {})
        print("POWER")
        print(f"  kit: {power.get('kit_id', '')}")
        print(f"  goal: {power.get('goal', '')}")
        for command in power.get("commands", []):
            print(f"    {command}")


def build_public_example_proof(
    example_id: str,
    *,
    output_path: Path | None = None,
    registry: dict[str, Any] | None = None,
) -> dict[str, Any]:
    spec = PUBLIC_EXAMPLE_SPECS.get(example_id)
    if spec is None:
        raise ValueError(f"Unknown public example {example_id!r}")
    selected_registry = registry or load_registry()
    source_kind = str(spec.get("source_kind", "compiled")).strip()
    generated = False
    compile_warnings: list[str] = []
    validation_warnings: list[str] = []
    validation_errors: list[str] = []

    if source_kind == "bundled":
        source_path = ROOT / str(spec.get("source_path", "")).strip()
        if not source_path.exists():
            raise FileNotFoundError(f"Bundled example path not found: {source_path}")
        pack_dir = source_path
    else:
        temp_root: Path | None = None
        if output_path is None:
            temp_root = Path(tempfile.mkdtemp(prefix=f"jini-example-{slugify(example_id)}-"))
            target_output = temp_root / slugify(example_id)
        else:
            target_output = output_path.expanduser()
        owner = str(spec.get("owner", "")).strip()
        approvers = [str(item).strip() for item in spec.get("approvers", []) if str(item).strip()]
        raw_stakeholders = [str(item).strip() for item in spec.get("stakeholders", []) if str(item).strip()]
        stakeholders = list(dict.fromkeys([owner, *raw_stakeholders]))
        operator = owner
        rollback_authority = approvers[0] if approvers else owner
        service_owner = owner
        pack_dir = write_compiled_pack(
            pack_id=str(spec.get("pack_id", "")).strip(),
            output_dir=target_output,
            work_unit_id=f"example-{slugify(example_id)}",
            title=str(spec.get("title", spec.get("label", example_id))).strip(),
            purpose=str(spec.get("purpose", spec.get("scenario", ""))).strip(),
            owner_actor_id=owner,
            approver_actor_ids=approvers,
            stakeholder_actor_ids=stakeholders,
            branch_id="main",
            operator_actor_id=operator,
            rollback_authority_actor_id=rollback_authority,
            service_owner_actor_id=service_owner,
            context_path=None,
        )
        generated = True
        validation_errors, validation_warnings = validate_pack(pack_dir, selected_registry)
        if validation_errors:
            raise ValueError("Public example failed validation:\n- " + "\n- ".join(validation_errors))
        compile_warnings = materialize_compile_outputs(pack_dir, selected_registry)

    summary = summarise_pack(pack_dir, selected_registry)
    tasks = summary.get("task_summary", {})
    evidence_doc = summary.get("evidence_doc") or {}
    future_missing = [
        artifact_type
        for artifact_type in summary.get("missing_full_required", [])
        if artifact_type not in summary.get("missing_stage_required", [])
    ]
    cli = cli_invocation()
    report = {
        "schema_version": "0.1.0",
        "example_type": "JiniPublicExampleProof",
        "generated_at": now_utc(),
        "example_id": example_id,
        "label": str(spec.get("label", example_id)),
        "pack_id": summary.get("pack_id", ""),
        "scenario": str(spec.get("scenario", "")).strip(),
        "source_kind": "generated" if generated else "bundled",
        "generated": generated,
        "path": display_path(pack_dir),
        "health": summary.get("health", ""),
        "state": summary.get("work_unit", {}).get("current_state", ""),
        "next_operation": summary.get("next_operation", ""),
        "control_packs": summary.get("control_packs", []),
        "missing_now": summary.get("missing_stage_required", []),
        "missing_later": future_missing,
        "task_summary": {
            "total": int(tasks.get("total", 0) or 0),
            "done": int(tasks.get("done", 0) or 0),
            "unresolved": int(tasks.get("unresolved", 0) or 0),
            "statuses": {str(key): int(value) for key, value in dict(tasks.get("counts", {})).items()},
        },
        "evidence_summary": {
            "present": bool(evidence_doc),
            "claims": len(evidence_doc.get("claims", [])) if isinstance(evidence_doc, dict) else 0,
            "risks": len(evidence_doc.get("risks", [])) if isinstance(evidence_doc, dict) else 0,
            "target": str(evidence_doc.get("target_artifact_id", "")) if isinstance(evidence_doc, dict) else "",
        },
        "daily_value": list(spec.get("daily_value", [])),
        "try_command": f"{cli} example {example_id}",
        "continue_with": [
            f"{cli} outcome {display_path(pack_dir)}",
            f"{cli} next {display_path(pack_dir)}",
        ],
        "warnings": [*validation_warnings, *compile_warnings],
    }
    return report


def print_public_example_proof(report: dict[str, Any]) -> None:
    print(f"EXAMPLE {report.get('label', '')}")
    print(f"PACK    {report.get('pack_id', '')}")
    print(f"PATH    {report.get('path', '')}")
    print(f"SOURCE  {report.get('source_kind', '')}")
    print(f"STATE   {report.get('state', '')}")
    print(f"HEALTH  {report.get('health', '')}")
    print(f"NEXT    {report.get('next_operation', '')}")
    scenario = str(report.get("scenario", "")).strip()
    if scenario:
        print(f"WHY     {scenario}")
    control_packs = report.get("control_packs", [])
    if control_packs:
        print(f"CTRL    {', '.join(control_packs)}")
    if report.get("missing_now"):
        print("MISSING-NOW")
        for item in report["missing_now"]:
            print(f"  - {item}")
    if report.get("missing_later"):
        print("MISSING-LATER")
        for item in report["missing_later"]:
            print(f"  - {item}")
    task_summary = report.get("task_summary", {})
    total = int(task_summary.get("total", 0) or 0)
    if total:
        print("TASKS")
        print(f"  done:       {task_summary.get('done', 0)}/{total}")
        print(f"  unresolved: {task_summary.get('unresolved', 0)}/{total}")
    evidence_summary = report.get("evidence_summary", {})
    if evidence_summary.get("present"):
        print("EVIDENCE")
        if evidence_summary.get("target"):
            print(f"  target: {evidence_summary['target']}")
        print(f"  claims: {evidence_summary.get('claims', 0)}")
        print(f"  risks:  {evidence_summary.get('risks', 0)}")
    if report.get("daily_value"):
        print("VALUE")
        for item in report["daily_value"]:
            print(f"  - {item}")
    if report.get("continue_with"):
        print("CONTINUE")
        for item in report["continue_with"]:
            print(f"  - {item}")
    for warning in report.get("warnings", []):
        print(f"WARN   {warning}")


def resolve_install_root(raw_root: str, prefix: Path | None = None) -> Path:
    if prefix is None:
        return Path(raw_root).expanduser()
    trimmed = raw_root.replace("~", "", 1).lstrip("/")
    return prefix / trimmed


def install_receipt_dir(prefix: Path | None = None) -> Path:
    if prefix is None:
        return Path("~/.jini/install-receipts").expanduser()
    return prefix / ".jini" / "install-receipts"


def pack_policy_candidate_dir(pack_dir: Path) -> Path:
    path = pack_dir / "runtime" / "policy-candidates"
    path.mkdir(parents=True, exist_ok=True)
    return path


def pack_policy_rollout_dir(pack_dir: Path) -> Path:
    path = pack_dir / "runtime" / "policy-rollouts"
    path.mkdir(parents=True, exist_ok=True)
    return path


def active_policy_rollout_path(pack_dir: Path, policy_id: str) -> Path:
    rollout_dir = pack_policy_rollout_dir(pack_dir)
    return rollout_dir / f"{slugify(policy_id)}-active.json"


def manifest_source_to_repo_relative(source_path: str) -> str:
    candidate = Path(source_path)
    if not candidate.is_absolute() and (ROOT / candidate).exists():
        return str(candidate)
    resolved = candidate.resolve()
    try:
        return str(resolved.relative_to(ROOT))
    except ValueError:
        return str(candidate)


def materialize_install_path(source_relative: str, destination_root: Path, link_mode: str) -> Path:
    source = ROOT / source_relative
    destination = destination_root / source_relative
    destination.parent.mkdir(parents=True, exist_ok=True)
    if destination.exists() or destination.is_symlink():
        remove_installed_path(destination)
    if link_mode == "symlink":
        destination.symlink_to(source.resolve(), target_is_directory=source.is_dir())
        return destination
    if source.is_dir():
        shutil.copytree(source, destination)
    else:
        shutil.copy2(source, destination)
    return destination


def remove_installed_path(path: Path) -> None:
    if path.is_symlink() or path.is_file():
        path.unlink(missing_ok=True)
        return
    if path.is_dir():
        shutil.rmtree(path)


def maybe_prune_empty_parents(path: Path, stop_at: Path) -> None:
    current = path
    while current != stop_at and current.exists():
        try:
            current.rmdir()
        except OSError:
            break
        current = current.parent


def plan_install(
    *,
    bundle_ids: list[str] | None = None,
    kit_ids: list[str] | None = None,
    target_ids: list[str] | None = None,
    link_mode: str = "auto",
    prefix: Path | None = None,
) -> dict[str, Any]:
    manifest = load_install_manifest()
    manifest_version = str(manifest.get("version", "0.1.0"))
    source = manifest.get("source", {}) if isinstance(manifest.get("source", {}), dict) else {}
    source_name = str(source.get("name", "Jini"))
    source_path = display_path(ROOT)
    source_revision = str(source.get("revision", "workspace"))
    trust_notice = str(source.get("trust_notice", "")).strip()

    target_index = {
        entry["id"]: entry
        for entry in manifest["targets"]
        if isinstance(entry, dict) and isinstance(entry.get("id"), str)
    }
    bundle_index = {
        entry["id"]: entry
        for entry in manifest["bundles"]
        if isinstance(entry, dict) and isinstance(entry.get("id"), str)
    }
    kit_index = {
        entry["id"]: entry
        for entry in manifest.get("kits", [])
        if isinstance(entry, dict) and isinstance(entry.get("id"), str)
    }
    default_kit_id = str(manifest.get("default_kit_id", "")).strip()

    selected_target_ids = target_ids or list(target_index.keys())
    requested_kit_ids = list(kit_ids or [])
    if not requested_kit_ids and not bundle_ids and default_kit_id:
        requested_kit_ids = [default_kit_id]
    unknown_kits = [kit for kit in requested_kit_ids if kit not in kit_index]
    if unknown_kits:
        raise ValueError(f"Unknown install kit(s): {', '.join(unknown_kits)}")
    selected_bundle_ids: list[str] = []
    if requested_kit_ids:
        for kit_id in requested_kit_ids:
            bundle_list = kit_index[kit_id].get("bundle_ids", [])
            if not isinstance(bundle_list, list):
                continue
            for bundle_id in bundle_list:
                if isinstance(bundle_id, str) and bundle_id not in selected_bundle_ids:
                    selected_bundle_ids.append(bundle_id)
    if bundle_ids:
        for bundle_id in bundle_ids:
            if bundle_id not in selected_bundle_ids:
                selected_bundle_ids.append(bundle_id)
    if not selected_bundle_ids:
        selected_bundle_ids = list(bundle_index.keys())

    unknown_targets = [target for target in selected_target_ids if target not in target_index]
    if unknown_targets:
        raise ValueError(f"Unknown install target(s): {', '.join(unknown_targets)}")
    unknown_bundles = [bundle for bundle in selected_bundle_ids if bundle not in bundle_index]
    if unknown_bundles:
        raise ValueError(f"Unknown install bundle(s): {', '.join(unknown_bundles)}")

    targets: list[dict[str, Any]] = []
    for target_id in selected_target_ids:
        target_doc = target_index[target_id]
        destination_root = str(resolve_install_root(str(target_doc.get("destination_root", "")), prefix))
        shim_root = str(resolve_install_root(str(target_doc.get("shim_root", "")), prefix))
        targets.append(
            {
                "id": target_id,
                "label": target_doc.get("label", target_id),
                "destination_root": destination_root,
                "shim_root": shim_root,
                "default_link_mode": target_doc.get("default_link_mode", "symlink"),
                "risk_notice": target_doc.get("risk_notice", ""),
                "activation_steps": target_doc.get("activation_steps", []),
                "doctor_expectations": target_doc.get("doctor_expectations", {}),
            }
        )

    bundle_plans: list[dict[str, Any]] = []
    risk_notices: list[str] = []

    for bundle_id in selected_bundle_ids:
        bundle = bundle_index[bundle_id]
        universal_paths = bundle.get("universal_paths", [])
        if not isinstance(universal_paths, list):
            raise ValueError(f"Bundle {bundle_id!r} has invalid universal_paths")
        compatible_targets = bundle.get("compatible_targets", list(target_index.keys()))
        if not isinstance(compatible_targets, list):
            compatible_targets = list(target_index.keys())
        target_shims = bundle.get("target_shims", {})
        if not isinstance(target_shims, dict):
            target_shims = {}
        permission_risk = str(bundle.get("permission_risk", "unknown"))
        review_before_use = bool(bundle.get("review_before_use", False))
        bundle_version = str(bundle.get("version", manifest_version))
        depends_on = bundle.get("depends_on", [])
        if not isinstance(depends_on, list):
            depends_on = []
        activation_steps = bundle.get("activation_steps", [])
        if not isinstance(activation_steps, list):
            activation_steps = []
        migration_notes = bundle.get("migration_notes", [])
        if not isinstance(migration_notes, list):
            migration_notes = []

        installations: list[dict[str, Any]] = []
        for target in targets:
            target_id = target["id"]
            if target_id not in compatible_targets:
                continue
            effective_link_mode = target["default_link_mode"] if link_mode == "auto" else link_mode
            universal_destination = f"{target['destination_root']}/bundles/{bundle_id}"
            shim_paths = target_shims.get(target_id, [])
            if not isinstance(shim_paths, list):
                shim_paths = []
            shim_destination = f"{target['shim_root']}/{bundle_id}" if shim_paths else ""
            installations.append(
                {
                    "target_id": target_id,
                    "target_label": target["label"],
                    "link_mode": effective_link_mode,
                    "universal_destination": universal_destination,
                    "shim_paths": expand_manifest_paths(shim_paths),
                    "shim_destination": shim_destination,
                    "activation_steps": target.get("activation_steps", []),
                }
            )

        bundle_plan = {
            "id": bundle_id,
            "label": bundle.get("label", bundle_id),
            "kind": bundle.get("kind", "bundle"),
            "version": bundle_version,
            "summary": bundle.get("summary", ""),
            "permission_risk": permission_risk,
            "review_before_use": review_before_use,
            "depends_on": depends_on,
            "activation_steps": activation_steps,
            "migration_notes": migration_notes,
            "universal_paths": expand_manifest_paths(universal_paths),
            "installations": installations,
        }
        bundle_plans.append(bundle_plan)

        if review_before_use:
            risk_notices.append(f"{bundle_plan['label']}: review before activation because it affects agent behavior.")
        risk_notices.append(f"{bundle_plan['label']}: permission risk `{permission_risk}`.")

    receipt_seed = json.dumps(
        {
            "bundles": selected_bundle_ids,
            "kits": requested_kit_ids,
            "targets": selected_target_ids,
            "link_mode": link_mode,
            "version": manifest_version,
        },
        sort_keys=True,
    )
    receipt_id = hashlib.sha256(receipt_seed.encode("utf-8")).hexdigest()[:12]
    manifest_digest = hashlib.sha256(INSTALL_MANIFEST_PATH.read_text(encoding="utf-8").encode("utf-8")).hexdigest()[:16]

    selected_kits = []
    for kit_id in requested_kit_ids:
        kit = kit_index[kit_id]
        selected_kits.append(
            {
                "id": kit_id,
                "label": kit.get("label", kit_id),
                "summary": kit.get("summary", ""),
                "bundle_ids": kit.get("bundle_ids", []),
                "recommended_targets": kit.get("recommended_targets", []),
            }
        )
    cli = cli_invocation()
    install_command = [cli, "install-bundles"]
    doctor_command = [cli, "doctor-install"]
    if requested_kit_ids:
        for kit_id in requested_kit_ids:
            install_command.extend(["--kit", kit_id])
            doctor_command.extend(["--kit", kit_id])
    elif bundle_ids:
        for bundle_id in bundle_ids:
            install_command.extend(["--bundle", bundle_id])
            doctor_command.extend(["--bundle", bundle_id])
    for target_id in selected_target_ids:
        install_command.extend(["--target", target_id])
        doctor_command.extend(["--target", target_id])
    if prefix is not None:
        resolved_prefix = str(prefix.resolve())
        install_command.extend(["--prefix", resolved_prefix])
        doctor_command.extend(["--prefix", resolved_prefix])
    recommended_next_steps = [
        {
            "id": "install",
            "label": "Materialize the selected kit or bundle set",
            "command": " ".join(install_command),
        },
        {
            "id": "verify",
            "label": "Verify trust, receipts, and activation readiness",
            "command": " ".join(doctor_command),
        },
    ]

    return {
        "schema_version": "0.1.0",
        "plan_type": "JiniInstallPlan",
        "generated_at": now_utc(),
        "manifest_version": manifest_version,
        "default_kit_id": default_kit_id,
        "source": {
            "name": source_name,
            "path": source_path,
            "revision": source_revision,
            "trust_notice": trust_notice,
        },
        "selected_targets": targets,
        "selected_kits": selected_kits,
        "selected_bundles": bundle_plans,
        "link_mode": link_mode,
        "prefix": str(prefix.resolve()) if prefix is not None else "",
        "recommended_next_steps": recommended_next_steps,
        "risk_notices": list(dict.fromkeys(risk_notices)),
        "receipt": {
            "receipt_id": receipt_id,
            "manifest_digest": manifest_digest,
            "summary": f"Dry-run only. No files were written for {len(bundle_plans)} bundle(s) across {len(targets)} target(s).",
        },
    }


def print_install_plan(plan: dict[str, Any]) -> None:
    print(f"SOURCE  {plan['source']['name']}")
    print(f"PATH    {plan['source']['path']}")
    print(f"REV     {plan['source']['revision']}")
    print(f"MODE    {plan['link_mode']}")
    print(f"TARGETS {', '.join(target['label'] for target in plan['selected_targets'])}")
    if plan.get("selected_kits"):
        print(f"KITS    {', '.join(kit['label'] for kit in plan['selected_kits'])}")
    print(f"BUNDLES {len(plan['selected_bundles'])}")
    if plan["source"].get("trust_notice"):
        print(f"NOTICE  {plan['source']['trust_notice']}")
    if plan.get("recommended_next_steps"):
        print("TRUST PATH")
        for step in plan["recommended_next_steps"]:
            print(f"  - {step['label']}: {step['command']}")
    print("PLAN")
    for bundle in plan["selected_bundles"]:
        print(
            f"  - {bundle['label']} [{bundle['kind']}] risk={bundle['permission_risk']} "
            f"universal={len(bundle['universal_paths'])}"
        )
        for install in bundle["installations"]:
            shim_note = (
                f" shim={len(install['shim_paths'])} -> {install['shim_destination']}"
                if install["shim_paths"]
                else " shim=0"
            )
            print(
                f"    {install['target_label']}: {install['link_mode']} "
                f"-> {install['universal_destination']}{shim_note}"
            )
    print("RISKS")
    for notice in plan["risk_notices"]:
        print(f"  - {notice}")
    print("RECEIPT")
    print(f"  ID   {plan['receipt']['receipt_id']}")
    print(f"  DIGEST {plan['receipt']['manifest_digest']}")
    print(f"  NOTE {plan['receipt']['summary']}")


def write_install_receipt(
    plan: dict[str, Any],
    *,
    installs: list[dict[str, Any]],
    prefix: Path | None = None,
    status: str = "installed",
) -> Path:
    receipt_dir = install_receipt_dir(prefix)
    receipt_dir.mkdir(parents=True, exist_ok=True)
    receipt_path = receipt_dir / f"{plan['receipt']['receipt_id']}.json"
    payload = {
        "schema_version": "0.1.0",
        "receipt_type": "JiniInstallReceipt",
        "generated_at": now_utc(),
        "status": status,
        "plan": {
            "receipt_id": plan["receipt"]["receipt_id"],
            "manifest_digest": plan["receipt"]["manifest_digest"],
            "manifest_version": plan["manifest_version"],
            "source": plan["source"],
            "selected_targets": plan["selected_targets"],
            "selected_kits": [
                {
                    "id": kit["id"],
                    "label": kit["label"],
                }
                for kit in plan.get("selected_kits", [])
            ],
            "selected_bundles": [
                {
                    "id": bundle["id"],
                    "label": bundle["label"],
                    "kind": bundle["kind"],
                }
                for bundle in plan["selected_bundles"]
            ],
            "link_mode": plan["link_mode"],
            "prefix": plan.get("prefix", ""),
        },
        "installs": installs,
    }
    receipt_path.write_text(json.dumps(payload, indent=2) + "\n", encoding="utf-8")
    return receipt_path


def apply_install_plan(
    *,
    plan: dict[str, Any],
    prefix: Path | None = None,
    status: str = "installed",
) -> dict[str, Any]:
    installs: list[dict[str, Any]] = []
    for bundle in plan["selected_bundles"]:
        for installation in bundle["installations"]:
            universal_root = Path(installation["universal_destination"])
            shim_root = Path(installation["shim_destination"]) if installation["shim_destination"] else None
            written_paths: list[str] = []
            for relative_path in bundle["universal_paths"]:
                written = materialize_install_path(relative_path, universal_root, installation["link_mode"])
                written_paths.append(str(written))
            for relative_path in installation["shim_paths"]:
                repo_relative = manifest_source_to_repo_relative(relative_path)
                if shim_root is None:
                    continue
                written = materialize_install_path(repo_relative, shim_root, installation["link_mode"])
                written_paths.append(str(written))
            installs.append(
                {
                    "bundle_id": bundle["id"],
                    "bundle_label": bundle["label"],
                    "target_id": installation["target_id"],
                    "target_label": installation["target_label"],
                    "link_mode": installation["link_mode"],
                    "universal_destination": str(universal_root),
                    "shim_destination": str(shim_root) if shim_root is not None else "",
                    "written_paths": written_paths,
                }
            )

    receipt_path = write_install_receipt(plan, installs=installs, prefix=prefix, status=status)
    return {
        "schema_version": "0.1.0",
        "result_type": "JiniInstallResult",
        "generated_at": now_utc(),
        "status": status,
        "installs": installs,
        "risk_notices": plan["risk_notices"],
        "receipt_path": str(receipt_path),
    }


def install_bundles(
    *,
    bundle_ids: list[str] | None = None,
    kit_ids: list[str] | None = None,
    target_ids: list[str] | None = None,
    link_mode: str = "auto",
    prefix: Path | None = None,
) -> dict[str, Any]:
    plan = plan_install(
        bundle_ids=bundle_ids,
        kit_ids=kit_ids,
        target_ids=target_ids,
        link_mode=link_mode,
        prefix=prefix,
    )
    return apply_install_plan(plan=plan, prefix=prefix, status="installed")


def update_bundles(
    *,
    bundle_ids: list[str] | None = None,
    kit_ids: list[str] | None = None,
    target_ids: list[str] | None = None,
    link_mode: str = "auto",
    prefix: Path | None = None,
) -> dict[str, Any]:
    plan = plan_install(
        bundle_ids=bundle_ids,
        kit_ids=kit_ids,
        target_ids=target_ids,
        link_mode=link_mode,
        prefix=prefix,
    )
    return apply_install_plan(plan=plan, prefix=prefix, status="updated")


def uninstall_bundles(
    *,
    bundle_ids: list[str] | None = None,
    kit_ids: list[str] | None = None,
    target_ids: list[str] | None = None,
    prefix: Path | None = None,
) -> dict[str, Any]:
    plan = plan_install(
        bundle_ids=bundle_ids,
        kit_ids=kit_ids,
        target_ids=target_ids,
        link_mode="auto",
        prefix=prefix,
    )
    removals: list[dict[str, Any]] = []
    for bundle in plan["selected_bundles"]:
        for installation in bundle["installations"]:
            universal_root = Path(installation["universal_destination"])
            shim_root = Path(installation["shim_destination"]) if installation["shim_destination"] else None
            removed_paths: list[str] = []
            if universal_root.exists() or universal_root.is_symlink():
                remove_installed_path(universal_root)
                removed_paths.append(str(universal_root))
                maybe_prune_empty_parents(universal_root.parent, Path(plan["selected_targets"][0]["destination_root"]))
            if shim_root is not None and (shim_root.exists() or shim_root.is_symlink()):
                remove_installed_path(shim_root)
                removed_paths.append(str(shim_root))
                maybe_prune_empty_parents(shim_root.parent, Path(plan["selected_targets"][0]["shim_root"]))
            removals.append(
                {
                    "bundle_id": bundle["id"],
                    "target_id": installation["target_id"],
                    "removed_paths": removed_paths,
                }
            )
    receipt_path = write_install_receipt(plan, installs=removals, prefix=prefix, status="uninstalled")
    return {
        "schema_version": "0.1.0",
        "result_type": "JiniUninstallResult",
        "generated_at": now_utc(),
        "status": "uninstalled",
        "removals": removals,
        "receipt_path": str(receipt_path),
    }


def doctor_install(
    *,
    bundle_ids: list[str] | None = None,
    kit_ids: list[str] | None = None,
    target_ids: list[str] | None = None,
    prefix: Path | None = None,
) -> dict[str, Any]:
    plan = plan_install(
        bundle_ids=bundle_ids,
        kit_ids=kit_ids,
        target_ids=target_ids,
        link_mode="auto",
        prefix=prefix,
    )
    checks: list[dict[str, Any]] = []
    overall_missing = False
    overall_warning = False
    current_manifest_digest = str(plan.get("receipt", {}).get("manifest_digest", ""))
    target_index = {
        str(target.get("id", "")): target
        for target in plan.get("selected_targets", [])
        if isinstance(target, dict)
    }

    receipt_dir = install_receipt_dir(prefix)
    receipt_files = sorted(receipt_dir.glob("*.json")) if receipt_dir.exists() else []
    parsed_receipts: list[dict[str, Any]] = []
    for receipt_file in receipt_files:
        try:
            receipt_doc = json.loads(receipt_file.read_text(encoding="utf-8"))
        except (OSError, json.JSONDecodeError):
            continue
        parsed_receipts.append(
            {
                "path": str(receipt_file),
                "generated_at": str(receipt_doc.get("generated_at", "")),
                "status": str(receipt_doc.get("status", "unknown")),
                "installs": receipt_doc.get("installs", []),
                "manifest_digest": str(receipt_doc.get("plan", {}).get("manifest_digest", "")),
            }
        )

    for bundle in plan["selected_bundles"]:
        for installation in bundle["installations"]:
            universal_root = Path(installation["universal_destination"])
            shim_root = Path(installation["shim_destination"]) if installation["shim_destination"] else None
            present_paths: list[str] = []
            missing_paths: list[str] = []
            for relative_path in bundle["universal_paths"]:
                expected = universal_root / relative_path
                if expected.exists() or expected.is_symlink():
                    present_paths.append(str(expected))
                else:
                    missing_paths.append(str(expected))
            for relative_path in installation["shim_paths"]:
                repo_relative = manifest_source_to_repo_relative(relative_path)
                expected = (shim_root / repo_relative) if shim_root is not None else None
                if expected is None:
                    continue
                if expected.exists() or expected.is_symlink():
                    present_paths.append(str(expected))
                else:
                    missing_paths.append(str(expected))
            receipt_matches = []
            receipt_statuses: list[tuple[str, str]] = []
            receipt_records: list[dict[str, Any]] = []
            for receipt_doc in parsed_receipts:
                installs = receipt_doc.get("installs", [])
                for item in installs:
                    if not isinstance(item, dict):
                        continue
                    if item.get("bundle_id") != bundle["id"] or item.get("target_id") != installation["target_id"]:
                        continue
                    receipt_matches.append(receipt_doc["path"])
                    receipt_statuses.append((receipt_doc["generated_at"], receipt_doc["status"]))
                    receipt_records.append(
                        {
                            "path": receipt_doc["path"],
                            "generated_at": receipt_doc["generated_at"],
                            "status": receipt_doc["status"],
                            "manifest_digest": receipt_doc.get("manifest_digest", ""),
                            "install": item,
                        }
                    )
            receipt_statuses.sort(key=lambda item: item[0])
            receipt_records.sort(key=lambda item: item["generated_at"])
            latest_receipt = receipt_records[-1] if receipt_records else None
            activation_steps = list(dict.fromkeys(
                [str(step) for step in installation.get("activation_steps", []) if str(step).strip()]
                + [str(step) for step in bundle.get("activation_steps", []) if str(step).strip()]
            ))
            health_checks: list[dict[str, Any]] = []
            health_checks.append(
                {
                    "id": "receipt-present",
                    "status": "ok" if receipt_matches else "warning",
                    "detail": "Latest install receipt is available for this bundle and target."
                    if receipt_matches
                    else "No matching install receipt was found for this bundle and target.",
                }
            )
            if latest_receipt is not None:
                latest_manifest_digest = str(latest_receipt.get("manifest_digest", "")).strip()
                digest_matches = bool(latest_manifest_digest) and latest_manifest_digest == current_manifest_digest
                health_checks.append(
                    {
                        "id": "receipt-manifest-current",
                        "status": "ok" if digest_matches else "warning",
                        "detail": "Latest install receipt matches the current install manifest digest."
                        if digest_matches
                        else "Latest install receipt was generated from a different install manifest digest.",
                    }
                )
                expected_link_mode = str(latest_receipt["install"].get("link_mode", installation["link_mode"]))
                inspected_paths = [
                    Path(path)
                    for path in latest_receipt["install"].get("written_paths", [])
                    if isinstance(path, str) and (Path(path).exists() or Path(path).is_symlink())
                ]
                if inspected_paths:
                    if expected_link_mode == "symlink":
                        mismatched = [str(path) for path in inspected_paths if not path.is_symlink()]
                        link_ok = not mismatched
                    else:
                        mismatched = [str(path) for path in inspected_paths if path.is_symlink()]
                        link_ok = not mismatched
                    health_checks.append(
                        {
                            "id": "install-link-mode",
                            "status": "ok" if link_ok else "warning",
                            "detail": f"Installed paths match the recorded `{expected_link_mode}` link mode."
                            if link_ok
                            else f"Some installed paths do not match the recorded `{expected_link_mode}` link mode: {', '.join(mismatched[:4])}.",
                        }
                    )
                else:
                    health_checks.append(
                        {
                            "id": "install-link-mode",
                            "status": "warning",
                            "detail": "No installed paths were available for behavioral link-mode validation.",
                        }
                    )
            if shim_root is not None:
                target_doc = target_index.get(installation["target_id"], {})
                doctor_expectations = target_doc.get("doctor_expectations", {}) if isinstance(target_doc, dict) else {}
                shim_readme = None
                for relative_path in installation["shim_paths"]:
                    repo_relative = manifest_source_to_repo_relative(relative_path)
                    candidate = shim_root / repo_relative
                    if candidate.name == "README.md":
                        shim_readme = candidate
                        break
                if shim_readme is not None and shim_readme.exists():
                    shim_text = shim_readme.read_text(encoding="utf-8")
                    for token, check_id in (
                        ("plan-install", "shim-documents-plan"),
                        ("doctor-install", "shim-documents-doctor"),
                        ("activate-runtime-target", "shim-documents-activate"),
                    ):
                        health_checks.append(
                            {
                                "id": check_id,
                                "status": "ok" if token in shim_text else "warning",
                                "detail": f"Shim README {'documents' if token in shim_text else 'does not document'} `{token}`.",
                            }
                        )
                    expected_markers = [
                        str(marker)
                        for marker in doctor_expectations.get("readme_markers", [])
                        if str(marker).strip()
                    ]
                    if expected_markers:
                        missing_markers = [marker for marker in expected_markers if marker not in shim_text]
                        health_checks.append(
                            {
                                "id": "shim-target-behavior",
                                "status": "ok" if not missing_markers else "warning",
                                "detail": f"Shim README contains the target-specific behavioral markers for `{installation['target_id']}`."
                                if not missing_markers
                                else f"Shim README is missing target-specific behavioral markers: {', '.join(missing_markers)}.",
                            }
                        )
                    cli_path = universal_root / "tools" / "jini.py"
                    smoke_commands = doctor_expectations.get("smoke_commands", [])
                    if cli_path.exists() and isinstance(smoke_commands, list):
                        for raw_probe in smoke_commands:
                            if not isinstance(raw_probe, dict):
                                continue
                            health_checks.append(
                                run_install_smoke_probe(
                                    cli_path,
                                    universal_root,
                                    installation["target_id"],
                                    raw_probe,
                                )
                            )
                else:
                    health_checks.append(
                        {
                            "id": "shim-readme-present",
                            "status": "missing",
                            "detail": "Target shim README is missing from the installed shim root.",
                        }
                    )
            if bundle["id"] == "jini-core":
                cli_path = universal_root / "tools" / "jini.py"
                validator_path = universal_root / "tools" / "jini_validate.py"
                health_checks.append(
                    {
                        "id": "core-cli-entrypoint",
                        "status": "ok" if cli_path.exists() else "missing",
                        "detail": "Jini CLI entrypoint is present." if cli_path.exists() else "Jini CLI entrypoint is missing.",
                    }
                )
                health_checks.append(
                    {
                        "id": "core-validator-entrypoint",
                        "status": "ok" if validator_path.exists() else "missing",
                        "detail": "Jini validator entrypoint is present." if validator_path.exists() else "Jini validator entrypoint is missing.",
                    }
                )
                activation_base = shim_root if shim_root is not None else universal_root
                activation_parent = activation_base / "runtime-handoffs"
                activation_roots: list[str] = []
                healthy_activation_roots: list[str] = []
                activation_target_mismatches: list[str] = []
                if activation_parent.exists():
                    for candidate in sorted(path for path in activation_parent.iterdir() if path.is_dir()):
                        activation_roots.append(str(candidate))
                        required = [
                            candidate / "handoff.json",
                            candidate / "compact-context.json",
                            candidate / "execution-checklist.json",
                            candidate / "Jini-RUNTIME.md",
                        ]
                        if all(path.exists() for path in required):
                            healthy_activation_roots.append(str(candidate))
                            try:
                                handoff_doc = load_json_file(candidate / "handoff.json")
                            except (OSError, json.JSONDecodeError, ValueError):
                                activation_target_mismatches.append(f"{candidate}: unreadable handoff.json")
                            else:
                                selected_target = str(
                                    handoff_doc.get("runtime_target", {}).get("selected", {}).get("id", "")
                                ).strip()
                                if selected_target and selected_target != installation["target_id"]:
                                    activation_target_mismatches.append(
                                        f"{candidate}: handoff selects `{selected_target}` instead of `{installation['target_id']}`"
                                    )
                if healthy_activation_roots:
                    activation_status = "ready"
                    health_checks.append(
                        {
                            "id": "runtime-activation",
                            "status": "ok",
                            "detail": "At least one runtime activation bundle is present and complete.",
                        }
                    )
                elif activation_roots:
                    activation_status = "invalid"
                    health_checks.append(
                        {
                            "id": "runtime-activation",
                            "status": "warning",
                            "detail": "Runtime activation roots exist, but required activation files are incomplete.",
                        }
                    )
                else:
                    activation_status = "inactive"
                    health_checks.append(
                        {
                            "id": "runtime-activation",
                            "status": "info",
                            "detail": "No runtime activation bundle has been materialized yet for this target.",
                        }
                    )
                if healthy_activation_roots:
                    health_checks.append(
                        {
                            "id": "runtime-activation-target-match",
                            "status": "ok" if not activation_target_mismatches else "warning",
                            "detail": "Runtime activation bundles target the same runtime target as the install."
                            if not activation_target_mismatches
                            else f"Some runtime activation bundles do not match the installed target: {', '.join(activation_target_mismatches[:3])}.",
                        }
                    )
            else:
                activation_roots = []
                healthy_activation_roots = []
                activation_status = "not-applicable"

            derived_status = "ok"
            if missing_paths or any(item["status"] == "missing" for item in health_checks):
                derived_status = "missing"
            elif any(item["status"] == "warning" for item in health_checks):
                derived_status = "warning"
            overall_missing = overall_missing or derived_status == "missing"
            overall_warning = overall_warning or derived_status == "warning"
            checks.append(
                {
                    "bundle_id": bundle["id"],
                    "target_id": installation["target_id"],
                    "status": derived_status,
                    "present_paths": present_paths,
                    "missing_paths": missing_paths,
                    "receipt_paths": receipt_matches,
                    "latest_receipt_status": receipt_statuses[-1][1] if receipt_statuses else "missing",
                    "activation_steps": activation_steps,
                    "health_checks": health_checks,
                    "activation_status": activation_status,
                    "activation_roots": activation_roots,
                    "healthy_activation_roots": healthy_activation_roots,
                }
            )

    return {
        "schema_version": "0.1.0",
        "result_type": "JiniInstallDoctorReport",
        "generated_at": now_utc(),
        "status": "missing" if overall_missing else "warning" if overall_warning else "ok",
        "checks": checks,
    }


def list_packs() -> list[tuple[str, Path, dict[str, Any]]]:
    packs: list[tuple[str, Path, dict[str, Any]]] = []
    if not PACKS_ROOT.exists():
        return packs
    for pack_dir in sorted(p for p in PACKS_ROOT.iterdir() if p.is_dir()):
        manifest_path = pack_dir / "pack.yaml"
        if not manifest_path.exists():
            continue
        manifest = load_document(manifest_path)
        packs.append((manifest["pack_id"], pack_dir, manifest))
    return packs


def pack_map() -> dict[str, tuple[Path, dict[str, Any]]]:
    return {name: (path, manifest) for name, path, manifest in list_packs()}


def build_pack_catalog() -> dict[str, Any]:
    entries: list[dict[str, Any]] = []
    for pack_id, pack_dir, manifest in list_packs():
        bootstrap_mode, rationale = recommended_bootstrap_mode(pack_id, manifest)
        entries.append(
            {
                "pack_id": pack_id,
                "path": display_path(pack_dir),
                "target_profile": manifest.get("target_profile", ""),
                "extensions": list(manifest.get("extensions", [])),
                "control_packs": list(manifest.get("control_packs", [])),
                "emits": list(manifest.get("emits", [])),
                "compiled_flow": list(manifest.get("compiled_flow", [])),
                "bootstrap_mode": bootstrap_mode,
                "bootstrap_rationale": rationale,
            }
        )
    return {
        "schema_version": "0.1.0",
        "catalog_type": "JiniPackCatalog",
        "generated_at": now_utc(),
        "pack_count": len(entries),
        "packs": entries,
    }


def print_pack_catalog(catalog: dict[str, Any]) -> None:
    print(f"PACKS {catalog['pack_count']}")
    for pack in catalog.get("packs", []):
        print(
            f"- {pack['pack_id']} | profile={pack['target_profile']} | "
            f"bootstrap={pack['bootstrap_mode']} | flow={', '.join(pack['compiled_flow'])}"
        )


def default_personal_tool_inventory() -> list[dict[str, str]]:
    return [
        {
            "id": "jini-cli",
            "kind": "cli",
            "scope": "local",
            "location": display_path(ROOT / "tools" / "jini.py"),
            "description": "Core Jini CLI for packs, memory, routines, installs, and verification.",
        },
        {
            "id": "adapter-registry",
            "kind": "registry",
            "scope": "local",
            "location": display_path(ADAPTER_REGISTRY_PATH),
            "description": "Canonical adapter catalog with layers, capabilities, and maturity.",
        },
        {
            "id": "install-manifest",
            "kind": "manifest",
            "scope": "local",
            "location": display_path(INSTALL_MANIFEST_PATH),
            "description": "Manifest-driven bundle catalog for installation and activation planning.",
        },
        {
            "id": "pack-catalog",
            "kind": "catalog",
            "scope": "local",
            "location": display_path(PACKS_ROOT),
            "description": "Implemented Jini advanced set across compiled packs and examples.",
        },
        {
            "id": "golden-benchmark",
            "kind": "benchmark",
            "scope": "local",
            "location": display_path(GOLDEN_BENCHMARK_PATH),
            "description": "Weighted golden dataset used to compare Jini against Kiro and Hermes.",
        },
        {
            "id": "framework-evolution",
            "kind": "learning-loop",
            "scope": "local",
            "location": display_path(ROOT / "learning" / "framework-evolution"),
            "description": "Governed experiment, outcome, and backtest loop for framework improvement.",
        },
        {
            "id": "workspace-steering",
            "kind": "steering",
            "scope": "workspace",
            "location": ".jini/steering",
            "description": "Workspace steering docs inspired by spec-first agents, adapted to Jini semantics.",
        },
    ]


def render_personal_tools_markdown(inventory: list[dict[str, str]]) -> str:
    lines = [
        "# Tools",
        "",
        "This is the operator-facing tool inventory for the local Jini personal OS home.",
        "",
    ]
    for entry in inventory:
        lines.extend(
            [
                f"## {entry['id']}",
                f"- Kind: `{entry['kind']}`",
                f"- Scope: `{entry['scope']}`",
                f"- Location: `{entry['location']}`",
                f"- Description: {entry['description']}",
                "",
            ]
        )
    return "\n".join(lines)


def default_personal_routines() -> list[dict[str, Any]]:
    return [
        {
            "schema_version": "0.1.0",
            "routine_id": "dream-memory",
            "title": "Dream Memory",
            "mode": "local",
            "runner": "builtin",
            "builtin_id": "dream-memory",
            "enabled": True,
            "summary": "Compress daily memory lines into long-term memory.",
            "cadence": "nightly",
            "outputs": ["memory/long-term.md"],
        },
        {
            "schema_version": "0.1.0",
            "routine_id": "daily-brief",
            "title": "Daily Brief",
            "mode": "local",
            "runner": "builtin",
            "builtin_id": "daily-brief",
            "enabled": True,
            "summary": "Generate a concise brief from today's daily notes and long-term memory.",
            "cadence": "daily",
            "outputs": ["outputs/briefs/"],
        },
        {
            "schema_version": "0.1.0",
            "routine_id": "golden-benchmark",
            "title": "Golden Benchmark",
            "mode": "local",
            "runner": "builtin",
            "builtin_id": "golden-benchmark",
            "enabled": True,
            "summary": "Run the golden competitor benchmark and render a concise score brief.",
            "cadence": "on-demand",
            "outputs": ["outputs/benchmarks/"],
        },
        {
            "schema_version": "0.1.0",
            "routine_id": "framework-review",
            "title": "Framework Review",
            "mode": "local",
            "runner": "builtin",
            "builtin_id": "framework-review",
            "enabled": True,
            "summary": "Render the current framework priorities and learned focus into a brief.",
            "cadence": "on-demand",
            "outputs": ["outputs/reviews/"],
        },
        {
            "schema_version": "0.1.0",
            "routine_id": "weekly-planning",
            "title": "Weekly Planning",
            "mode": "remote",
            "runner": "staged-remote",
            "enabled": True,
            "summary": "Stage a remote weekly planning run without assuming a cloud runner exists.",
            "cadence": "weekly",
            "outputs": ["runtime/remote-runs/"],
        },
        {
            "schema_version": "0.1.0",
            "routine_id": "publish-readiness",
            "title": "Publish Readiness",
            "mode": "local",
            "runner": "builtin",
            "builtin_id": "publish-readiness",
            "enabled": True,
            "summary": "Render a concise public-release readiness brief from the current Jini repo state.",
            "cadence": "on-demand",
            "outputs": ["outputs/release/"],
        },
    ]


def load_personal_home(home_root: Path) -> dict[str, Any]:
    manifest_path = home_root / "home.yaml"
    if not manifest_path.exists():
        raise ValueError(f"Personal home is not initialized at {display_path(home_root)}")
    manifest = load_document(manifest_path)
    if not isinstance(manifest, dict):
        raise ValueError("home.yaml must be a mapping")
    return manifest


def write_personal_home(home_root: Path, manifest: dict[str, Any]) -> None:
    dump_document(home_root / "home.yaml", manifest)


def memory_limit_config(manifest: dict[str, Any]) -> dict[str, int]:
    memory = manifest.get("memory", {}) if isinstance(manifest.get("memory", {}), dict) else {}
    limits = memory.get("limits", {}) if isinstance(memory.get("limits", {}), dict) else {}
    return {
        "long_term_char_limit": int(limits.get("long_term_char_limit", 4000) or 4000),
        "daily_compact_threshold_lines": int(limits.get("daily_compact_threshold_lines", 40) or 40),
        "stale_after_days": int(limits.get("stale_after_days", 3) or 3),
    }


def parse_markdown_front_matter(path: Path) -> tuple[dict[str, Any], str]:
    text = path.read_text(encoding="utf-8")
    if not text.startswith("---\n"):
        return {}, text
    parts = text.split("\n---\n", 1)
    if len(parts) != 2:
        return {}, text
    raw_meta = parts[0][4:]
    body = parts[1]
    try:
        meta = yaml.safe_load(raw_meta) or {}
    except yaml.YAMLError:
        meta = {}
    if not isinstance(meta, dict):
        meta = {}
    return meta, body


def steering_dirs(workspace_root: Path) -> list[tuple[str, Path]]:
    resolved = workspace_root.expanduser().resolve()
    return [
        ("jini", resolved / ".jini" / "steering"),
        ("kiro", resolved / ".kiro" / "steering"),
    ]


def bootstrap_workspace_steering(workspace_root: Path) -> dict[str, Any]:
    resolved = workspace_root.expanduser().resolve()
    repo_context = inspect_repo_context(resolved, repo_path=resolved)
    steering_root = resolved / ".jini" / "steering"
    if steering_root.exists() and any(steering_root.iterdir()):
        raise FileExistsError(f"Steering directory already exists and is not empty: {steering_root}")
    steering_root.mkdir(parents=True, exist_ok=True)

    def write_doc(filename: str, metadata: dict[str, Any], lines: list[str]) -> Path:
        path = steering_root / filename
        payload = "---\n" + yaml.safe_dump(metadata, sort_keys=False).strip() + "\n---\n\n" + "\n".join(lines) + "\n"
        path.write_text(payload, encoding="utf-8")
        return path

    repo_name = resolved.name
    docs = [
        (
            "product.md",
            {"title": "Product Context", "inclusion": "always"},
            [
                f"# Product Context",
                "",
                f"- Workspace: `{repo_name}`",
                "- Intent: describe the product purpose, user outcomes, and non-negotiable constraints here.",
                "- Scope: keep this focused on durable product truths rather than task-local notes.",
                "",
            ],
        ),
        (
            "tech.md",
            {"title": "Technical Context", "inclusion": "always"},
            [
                "# Technical Context",
                "",
                "- Detected entrypoints:",
                *(
                    [f"  - {item}" for item in repo_context.get("next_actions", [])[:5]]
                    if repo_context.get("next_actions")
                    else ["  - Add the real build, test, startup, and deployment constraints here."]
                ),
                "",
            ],
        ),
        (
            "structure.md",
            {"title": "Structure Context", "inclusion": "always"},
            [
                "# Structure Context",
                "",
                "- Key docs and directories:",
                *(
                    [f"  - {entry.get('path', entry.get('label', ''))}" for entry in repo_context.get("entrypoints", {}).get("docs", [])[:6]]
                    if repo_context.get("entrypoints", {}).get("docs")
                    else ["  - Add the stable repo boundaries, ownership seams, and important directories here."]
                ),
                "",
            ],
        ),
        (
            "testing.md",
            {
                "title": "Testing Context",
                "inclusion": "auto",
                "file_match": ["tests/**", "**/*test*", "**/*spec*", "scripts/**", ".github/workflows/**"],
            },
            [
                "# Testing Context",
                "",
                "- Preferred verification surfaces:",
                *(
                    [f"  - {target.get('command') or target.get('path') or target.get('label', '')}" for target in repo_context.get("verification_targets", [])[:6]]
                    if repo_context.get("verification_targets")
                    else ["  - Add the real verification commands and evidence expectations here."]
                ),
                "- Jini rule: refresh evidence through `harvest-evidence` before approval or promotion when the repo provides executable verification targets.",
                "",
            ],
        ),
    ]
    created_paths = [str(write_doc(filename, metadata, lines)) for filename, metadata, lines in docs]
    return {
        "schema_version": "0.1.0",
        "result_type": "JiniWorkspaceSteeringBootstrap",
        "generated_at": now_utc(),
        "workspace_root": str(resolved),
        "steering_root": str(steering_root),
        "created_paths": created_paths,
    }


def steering_document_active(
    metadata: dict[str, Any],
    *,
    target_path: Path | None = None,
) -> bool:
    inclusion = str(metadata.get("inclusion", "always")).strip().lower()
    if inclusion == "always":
        return True
    if inclusion == "manual":
        return False
    if inclusion == "auto":
        return True
    file_match = metadata.get("file_match", [])
    if inclusion == "filematch" and isinstance(file_match, list) and target_path is not None:
        import fnmatch

        normalized = str(target_path).replace("\\", "/")
        return any(fnmatch.fnmatch(normalized, pattern) for pattern in file_match if isinstance(pattern, str))
    return False


def build_steering_summary(workspace_root: Path, *, target_path: Path | None = None) -> dict[str, Any]:
    resolved = workspace_root.expanduser().resolve()
    evaluation_target: Path | None = None
    if target_path is not None:
        try:
            evaluation_target = target_path.expanduser().resolve().relative_to(resolved)
        except ValueError:
            evaluation_target = target_path
    documents: list[dict[str, Any]] = []
    active_paths: list[str] = []
    for source, steering_root in steering_dirs(resolved):
        if not steering_root.exists():
            continue
        for path in sorted(steering_root.glob("*.md")):
            metadata, body = parse_markdown_front_matter(path)
            active = steering_document_active(metadata, target_path=evaluation_target)
            entry = {
                "id": path.stem,
                "title": str(metadata.get("title", path.stem.replace("-", " ").title())),
                "inclusion": str(metadata.get("inclusion", "always")),
                "file_match": metadata.get("file_match", []) if isinstance(metadata.get("file_match", []), list) else [],
                "path": display_path(path),
                "source": source,
                "active": active,
                "char_count": len(body),
            }
            documents.append(entry)
            if active:
                active_paths.append(entry["path"])
    return {
        "schema_version": "0.1.0",
        "steering_type": "JiniWorkspaceSteeringSummary",
        "generated_at": now_utc(),
        "workspace_root": str(resolved),
        "found": bool(documents),
        "documents": documents,
        "active_paths": active_paths,
    }


def print_steering_summary(summary: dict[str, Any]) -> None:
    print(f"WORKSPACE {summary['workspace_root']}")
    print(f"FOUND    {summary['found']}")
    for doc in summary.get("documents", []):
        active = "active" if doc.get("active") else "inactive"
        print(
            f"- {doc['id']} | {doc['source']} | {doc['inclusion']} | {active} | {doc['path']}"
        )


def bootstrap_home(
    home_root: Path,
    *,
    owner_name: str = "",
    assistant_name: str = "Jini",
) -> dict[str, Any]:
    root = home_root.expanduser().resolve()
    if root.exists() and any(root.iterdir()):
        raise FileExistsError(f"Home directory already exists and is not empty: {root}")
    root.mkdir(parents=True, exist_ok=True)
    for relative in (
        "memory/daily",
        "routines/local",
        "routines/remote",
        "outputs/briefs",
        "outputs/benchmarks",
        "outputs/reviews",
        "outputs/release",
        "runtime/remote-runs",
    ):
        (root / relative).mkdir(parents=True, exist_ok=True)

    inventory = default_personal_tool_inventory()
    created_at = now_utc()
    manifest = {
        "schema_version": "0.1.0",
        "home_type": "JiniPersonalHome",
        "created_at": created_at,
        "updated_at": created_at,
        "assistant_name": assistant_name,
        "owner_name": owner_name,
        "memory": {
            "daily_dir": "memory/daily",
            "long_term_path": "memory/long-term.md",
            "limits": {
                "long_term_char_limit": 4000,
                "daily_compact_threshold_lines": 40,
                "stale_after_days": 3,
            },
        },
        "tools": {
            "inventory": inventory,
        },
        "routines": {
            "local_dir": "routines/local",
            "remote_dir": "routines/remote",
        },
    }
    write_personal_home(root, manifest)

    (root / "soul.md").write_text(
        "\n".join(
            [
                f"# {assistant_name} Soul",
                "",
                "- Be concise, direct, and reliable.",
                "- Prefer durable state over repeated chat context.",
                "- Default to the cheapest adequate path without weakening guardrails.",
                "",
            ]
        ),
        encoding="utf-8",
    )
    (root / "user.md").write_text(
        "\n".join(
            [
                "# User",
                "",
                f"- Name: {owner_name or 'Unknown'}",
                "- Preferences: add durable preferences here as they become clear.",
                "- Constraints: add recurring constraints, working style, and priorities here.",
                "",
            ]
        ),
        encoding="utf-8",
    )
    (root / "tools.md").write_text(render_personal_tools_markdown(inventory), encoding="utf-8")
    (root / "memory" / "long-term.md").write_text(
        "# Long-Term Memory\n\n- Durable notes will be synthesized here by `dream-memory`.\n",
        encoding="utf-8",
    )
    for routine in default_personal_routines():
        routine_dir = root / "routines" / routine["mode"]
        dump_document(routine_dir / f"{routine['routine_id']}.yaml", routine)

    return {
        "schema_version": "0.1.0",
        "result_type": "JiniPersonalHomeBootstrap",
        "generated_at": created_at,
        "home_root": str(root),
        "created_paths": [
            str(root / "home.yaml"),
            str(root / "soul.md"),
            str(root / "user.md"),
            str(root / "tools.md"),
            str(root / "memory" / "long-term.md"),
        ],
    }


def append_memory_line(
    home_root: Path,
    *,
    line: str,
    date_text: str | None = None,
) -> dict[str, Any]:
    if not line.strip():
        raise ValueError("append-memory requires a non-empty --line")
    manifest = load_personal_home(home_root)
    date_value = date_text or datetime.now(timezone.utc).strftime("%Y-%m-%d")
    daily_path = home_root / str(manifest["memory"]["daily_dir"]) / f"{date_value}.md"
    prefix = "- " if daily_path.exists() else f"# {date_value}\n\n- "
    with daily_path.open("a", encoding="utf-8") as handle:
        handle.write(prefix + line.strip() + "\n")
    manifest["updated_at"] = now_utc()
    write_personal_home(home_root, manifest)
    return {
        "schema_version": "0.1.0",
        "result_type": "JiniMemoryAppendResult",
        "generated_at": manifest["updated_at"],
        "daily_path": str(daily_path),
        "date": date_value,
        "line": line.strip(),
    }


def extract_memory_lines(path: Path) -> list[str]:
    lines: list[str] = []
    for raw in path.read_text(encoding="utf-8").splitlines():
        stripped = raw.strip()
        if not stripped or stripped.startswith("#"):
            continue
        if stripped.startswith("- "):
            stripped = stripped[2:]
        lines.append(stripped)
    return lines


def dream_memory(home_root: Path) -> dict[str, Any]:
    manifest = load_personal_home(home_root)
    daily_dir = home_root / str(manifest["memory"]["daily_dir"])
    long_term_path = home_root / str(manifest["memory"]["long_term_path"])
    limits = memory_limit_config(manifest)
    source_files = sorted(daily_dir.glob("*.md"))
    collected: list[str] = []
    seen: set[str] = set()
    for path in source_files:
        for line in extract_memory_lines(path):
            if line in seen:
                continue
            seen.add(line)
            collected.append(line)
    lines = [
        "# Long-Term Memory",
        "",
        f"- Updated: {now_utc()}",
        f"- Source files: {len(source_files)}",
        "",
        "## Durable Notes",
    ]
    note_lines = [f"- {item}" for item in collected] if collected else ["- No daily memory lines were available to compress."]
    original_note_count = len(note_lines)
    tail_lines = ["", "## Provenance"]
    if source_files:
        tail_lines.extend([f"- {display_path(path)}" for path in source_files])
    else:
        tail_lines.append("- No source files")
    lines.extend(note_lines)
    lines.extend(tail_lines)
    rendered = "\n".join(lines) + "\n"
    char_limit = limits["long_term_char_limit"]
    while len(rendered) > char_limit and len(note_lines) > 1:
        note_lines = note_lines[:-1]
        compacted_lines = lines[:6] + note_lines + tail_lines
        rendered = "\n".join(compacted_lines) + "\n"
        lines = compacted_lines
    long_term_path.write_text(rendered, encoding="utf-8")
    manifest["updated_at"] = now_utc()
    manifest["last_dream_at"] = manifest["updated_at"]
    write_personal_home(home_root, manifest)
    return {
        "schema_version": "0.1.0",
        "result_type": "JiniDreamMemoryResult",
        "generated_at": manifest["updated_at"],
        "long_term_path": str(long_term_path),
        "source_files": [str(path) for path in source_files],
        "source_line_count": len(collected),
        "char_limit": char_limit,
        "char_count": len(rendered),
        "compacted": len(note_lines) < original_note_count,
    }


def list_personal_tools(home_root: Path) -> dict[str, Any]:
    manifest = load_personal_home(home_root)
    inventory = manifest.get("tools", {}).get("inventory", [])
    return {
        "schema_version": "0.1.0",
        "result_type": "JiniPersonalToolList",
        "generated_at": now_utc(),
        "home_root": str(home_root.resolve()),
        "tools": inventory,
    }


def load_routine(path: Path) -> dict[str, Any]:
    routine = load_document(path)
    if not isinstance(routine, dict):
        raise ValueError(f"Routine file must be a mapping: {display_path(path)}")
    return routine


def list_personal_routines(home_root: Path, *, mode: str | None = None) -> dict[str, Any]:
    manifest = load_personal_home(home_root)
    routines: list[dict[str, Any]] = []
    mode_dirs = {
        "local": home_root / str(manifest["routines"]["local_dir"]),
        "remote": home_root / str(manifest["routines"]["remote_dir"]),
    }
    selected_modes = [mode] if mode in {"local", "remote"} else ["local", "remote"]
    for selected_mode in selected_modes:
        for path in sorted(mode_dirs[selected_mode].glob("*.yaml")):
            routine = load_routine(path)
            routines.append(
                {
                    "routine_id": routine.get("routine_id", ""),
                    "title": routine.get("title", ""),
                    "mode": routine.get("mode", selected_mode),
                    "runner": routine.get("runner", ""),
                    "summary": routine.get("summary", ""),
                    "path": str(path),
                    "enabled": bool(routine.get("enabled", True)),
                }
            )
    return {
        "schema_version": "0.1.0",
        "result_type": "JiniPersonalRoutineList",
        "generated_at": now_utc(),
        "home_root": str(home_root.resolve()),
        "routines": routines,
    }


def render_daily_brief(home_root: Path) -> Path:
    load_personal_home(home_root)
    today = datetime.now(timezone.utc).strftime("%Y-%m-%d")
    daily_path = home_root / "memory" / "daily" / f"{today}.md"
    long_term_path = home_root / "memory" / "long-term.md"
    daily_lines = extract_memory_lines(daily_path) if daily_path.exists() else []
    long_term_lines = extract_memory_lines(long_term_path) if long_term_path.exists() else []
    output_dir = home_root / "outputs" / "briefs"
    output_dir.mkdir(parents=True, exist_ok=True)
    stamp = datetime.now(timezone.utc).strftime("%Y%m%dT%H%M%SZ")
    output_path = output_dir / f"daily-brief-{stamp}.md"
    lines = [
        "# Daily Brief",
        "",
        f"- Generated: {now_utc()}",
        f"- Day: {today}",
        "",
        "## Today",
    ]
    if daily_lines:
        lines.extend([f"- {item}" for item in daily_lines[:10]])
    else:
        lines.append("- No daily memory captured yet.")
    lines.extend(["", "## Durable Anchors"])
    if long_term_lines:
        lines.extend([f"- {item}" for item in long_term_lines[:10]])
    else:
        lines.append("- No long-term memory available yet.")
    output_path.write_text("\n".join(lines) + "\n", encoding="utf-8")
    return output_path


def render_publish_readiness_brief(home_root: Path) -> Path:
    load_personal_home(home_root)
    report = build_publish_readiness()
    output_dir = home_root / "outputs" / "release"
    output_dir.mkdir(parents=True, exist_ok=True)
    stamp = datetime.now(timezone.utc).strftime("%Y%m%dT%H%M%SZ")
    output_path = output_dir / f"publish-readiness-{stamp}.md"
    lines = [
        "# Publish Readiness",
        "",
        f"- Generated: {report.get('generated_at', now_utc())}",
        f"- Status: {report.get('status', 'unknown')}",
        f"- Default Kit: {report.get('default_kit_id', '')}",
        f"- Packs: {report.get('pack_count', 0)}",
        f"- Bundles: {report.get('bundle_count', 0)}",
        f"- Kits: {report.get('kit_count', 0)}",
        "",
        "## Sections",
    ]
    for section in report.get("sections", []):
        lines.append(f"- {section.get('label', '')}: {section.get('status', '')}")
    output_path.write_text("\n".join(lines) + "\n", encoding="utf-8")
    return output_path


def render_golden_benchmark_brief(home_root: Path) -> Path:
    load_personal_home(home_root)
    if os.environ.get("JINI_GOLDEN_BENCHMARK_ACTIVE") == "1":
        report = build_golden_benchmark_projection()
        report_path: Path | None = None
    else:
        report, generated_report_path = build_golden_benchmark_report()
        report_path = generated_report_path
    output_dir = home_root / "outputs" / "benchmarks"
    output_dir.mkdir(parents=True, exist_ok=True)
    stamp = datetime.now(timezone.utc).strftime("%Y%m%dT%H%M%SZ")
    output_path = output_dir / f"golden-benchmark-{stamp}.md"
    overall = report.get("overall", {})
    lines = [
        "# Golden Benchmark",
        "",
        f"- Generated: {report.get('generated_at', now_utc())}",
        f"- Dataset Digest: {report.get('dataset_digest', '')}",
        f"- Dataset Verified: {report.get('last_verified_at', report.get('updated_at', ''))}",
        f"- Status: {overall.get('status', 'unknown')}",
        f"- Jini Score: {overall.get('jini_score', 0.0)}",
        f"- Strongest Competitor: {overall.get('strongest_competitor', 'unknown')}",
        f"- Strongest Competitor Score: {overall.get('strongest_competitor_score', 0.0)}",
        f"- Report Path: {display_path(report_path)}" if report_path is not None else "- Report Path: projected-only",
        "",
        "## Competitors",
    ]
    for competitor in report.get("competitors", []):
        lines.append(f"- {competitor.get('label', competitor.get('id', ''))}: {competitor.get('rationale', '')}")
    lines.extend(["", "## Scenarios"])
    for scenario in report.get("scenarios", []):
        competitor_scores = scenario.get("competitor_scores", {})
        lines.append(
            "- "
            f"{scenario.get('label', scenario.get('id', ''))}: "
            f"Jini {scenario.get('validated_jini_score', 0.0)} | "
            f"Kiro {competitor_scores.get('kiro', 0.0)} | "
            f"Hermes {competitor_scores.get('hermes', 0.0)} | "
            f"status={scenario.get('status', 'unknown')}"
        )
    if overall.get("failed_scenarios"):
        lines.extend(["", "## Failed Scenarios"])
        lines.extend([f"- {item}" for item in overall.get("failed_scenarios", [])])
    output_path.write_text("\n".join(lines) + "\n", encoding="utf-8")
    return output_path


def render_framework_review_brief(home_root: Path) -> Path:
    load_personal_home(home_root)
    review, review_path = build_framework_review(limit=5)
    backtest = build_framework_evolution_backtest()
    output_dir = home_root / "outputs" / "reviews"
    output_dir.mkdir(parents=True, exist_ok=True)
    stamp = datetime.now(timezone.utc).strftime("%Y%m%dT%H%M%SZ")
    output_path = output_dir / f"framework-review-{stamp}.md"
    lines = [
        "# Framework Review",
        "",
        f"- Generated: {review.get('generated_at', now_utc())}",
        f"- Review Path: {display_path(Path(str(review_path)))}",
        f"- Best Next Dimension: {review.get('best_next_dimension', '')}",
        f"- Learned Next Focus: {backtest.get('recommended_next_focus', '')}",
        f"- Outcome Count: {backtest.get('outcome_count', 0)}",
        "",
        "## Adoption Constraints",
    ]
    lines.extend([f"- {item}" for item in review.get("adoption_constraints", [])[:6]])
    lines.extend(["", "## Prioritized Dimensions"])
    for entry in review.get("prioritized_dimensions", []):
        strongest = entry.get("strongest_competitor", {})
        lines.append(
            "- "
            f"{entry.get('label', entry.get('id', ''))}: "
            f"current={entry.get('current_score', 0.0)} "
            f"target={entry.get('target_score', 0.0)} "
            f"best={strongest.get('name', '')}"
        )
    if backtest.get("dimension_summaries"):
        lines.extend(["", "## Learned Outcomes"])
        for item in backtest.get("dimension_summaries", [])[:5]:
            lines.append(
                "- "
                f"{item.get('dimension_id', '')}: "
                f"experiments={item.get('experiments', 0)} "
                f"avg_delta={item.get('average_score_delta', 0.0):+.3f} "
                f"avg_reward={item.get('average_reward', 0.0):+.3f}"
            )
    output_path.write_text("\n".join(lines) + "\n", encoding="utf-8")
    return output_path


def locate_routine(home_root: Path, routine_id: str, mode: str | None = None) -> tuple[Path, dict[str, Any]]:
    catalog = list_personal_routines(home_root, mode=mode)
    for item in catalog["routines"]:
        if item["routine_id"] != routine_id:
            continue
        path = Path(item["path"])
        return path, load_routine(path)
    raise ValueError(f"Unknown routine_id {routine_id!r}")


def run_personal_routine(home_root: Path, routine_id: str, *, mode: str | None = None) -> dict[str, Any]:
    routine_path, routine = locate_routine(home_root, routine_id, mode=mode)
    selected_mode = str(routine.get("mode", ""))
    runner = str(routine.get("runner", ""))
    if not bool(routine.get("enabled", True)):
        raise ValueError(f"Routine {routine_id!r} is disabled")

    if runner == "builtin":
        builtin_id = str(routine.get("builtin_id", "")).strip()
        if builtin_id == "dream-memory":
            result = dream_memory(home_root)
            output_paths = [result["long_term_path"]]
        elif builtin_id == "daily-brief":
            output_path = render_daily_brief(home_root)
            result = {
                "schema_version": "0.1.0",
                "result_type": "JiniDailyBriefResult",
                "generated_at": now_utc(),
                "output_path": str(output_path),
            }
            output_paths = [str(output_path)]
        elif builtin_id == "publish-readiness":
            output_path = render_publish_readiness_brief(home_root)
            result = {
                "schema_version": "0.1.0",
                "result_type": "JiniPublishReadinessRoutineResult",
                "generated_at": now_utc(),
                "output_path": str(output_path),
            }
            output_paths = [str(output_path)]
        elif builtin_id == "golden-benchmark":
            output_path = render_golden_benchmark_brief(home_root)
            result = {
                "schema_version": "0.1.0",
                "result_type": "JiniGoldenBenchmarkRoutineResult",
                "generated_at": now_utc(),
                "output_path": str(output_path),
            }
            output_paths = [str(output_path)]
        elif builtin_id == "framework-review":
            output_path = render_framework_review_brief(home_root)
            result = {
                "schema_version": "0.1.0",
                "result_type": "JiniFrameworkReviewRoutineResult",
                "generated_at": now_utc(),
                "output_path": str(output_path),
            }
            output_paths = [str(output_path)]
        else:
            raise ValueError(f"Unsupported builtin routine {builtin_id!r}")
        return {
            "schema_version": "0.1.0",
            "result_type": "JiniRoutineRunResult",
            "generated_at": now_utc(),
            "routine_id": routine_id,
            "mode": selected_mode,
            "runner": runner,
            "status": "executed",
            "output_paths": output_paths,
            "receipt_path": "",
            "details": result,
        }

    if runner == "shell":
        if not bool(routine.get("trusted_local", False)):
            raise ValueError(
                f"Routine {routine_id!r} uses raw shell execution and requires trusted_local: true"
            )
        argv = routine.get("argv", [])
        if isinstance(argv, list) and argv:
            completed = subprocess.run(
                [str(item) for item in argv],
                cwd=home_root,
                check=False,
                capture_output=True,
                text=True,
            )
            command = " ".join(shlex.quote(str(item)) for item in argv)
        else:
            command = str(routine.get("command", "")).strip()
            if not command:
                raise ValueError(f"Routine {routine_id!r} is missing command")
            completed = subprocess.run(
                command,
                cwd=home_root,
                shell=True,
                executable="/bin/bash",
                check=False,
                capture_output=True,
                text=True,
            )
        return {
            "schema_version": "0.1.0",
            "result_type": "JiniRoutineRunResult",
            "generated_at": now_utc(),
            "routine_id": routine_id,
            "mode": selected_mode,
            "runner": runner,
            "status": "executed" if completed.returncode == 0 else "failed",
            "output_paths": [],
            "receipt_path": "",
            "details": {
                "command": command,
                "exit_code": completed.returncode,
                "stdout_excerpt": trim_output(completed.stdout),
                "stderr_excerpt": trim_output(completed.stderr),
            },
        }

    if runner == "staged-remote":
        receipt_dir = home_root / "runtime" / "remote-runs"
        receipt_dir.mkdir(parents=True, exist_ok=True)
        stamp = datetime.now(timezone.utc).strftime("%Y%m%dT%H%M%SZ")
        receipt_path = receipt_dir / f"{routine_id}-{stamp}.json"
        receipt = {
            "schema_version": "0.1.0",
            "receipt_type": "JiniRemoteRoutineReceipt",
            "generated_at": now_utc(),
            "routine_id": routine_id,
            "mode": selected_mode,
            "runner": runner,
            "status": "staged-remote",
            "summary": routine.get("summary", ""),
            "outputs": routine.get("outputs", []),
            "path": display_path(routine_path),
        }
        receipt_path.write_text(json.dumps(receipt, indent=2) + "\n", encoding="utf-8")
        return {
            "schema_version": "0.1.0",
            "result_type": "JiniRoutineRunResult",
            "generated_at": now_utc(),
            "routine_id": routine_id,
            "mode": selected_mode,
            "runner": runner,
            "status": "staged-remote",
            "output_paths": [],
            "receipt_path": str(receipt_path),
            "details": receipt,
        }

    raise ValueError(f"Unsupported routine runner {runner!r}")


def home_binding_path(pack_dir: Path) -> Path:
    runtime_dir = pack_dir / "runtime"
    runtime_dir.mkdir(parents=True, exist_ok=True)
    return runtime_dir / "home-binding.json"


def bind_personal_home(
    pack_dir: Path,
    home_root: Path,
    *,
    updated_by: str = "cli",
) -> Path:
    resolved_home = home_root.expanduser().resolve()
    load_personal_home(resolved_home)
    payload = {
        "schema_version": "0.1.0",
        "binding_type": "JiniHomeBinding",
        "updated_at": now_utc(),
        "updated_by": updated_by,
        "home_root": str(resolved_home),
    }
    path = home_binding_path(pack_dir)
    path.write_text(json.dumps(payload, indent=2) + "\n", encoding="utf-8")
    return path


def load_home_binding(pack_dir: Path) -> dict[str, Any] | None:
    path = home_binding_path(pack_dir)
    if not path.exists():
        return None
    try:
        payload = load_json_file(path)
    except (json.JSONDecodeError, OSError):
        return None
    return payload if isinstance(payload, dict) else None


def resolve_home_binding(
    pack_dir: Path,
    *,
    explicit_home: Path | None = None,
) -> dict[str, Any]:
    if explicit_home is not None:
        resolved = explicit_home.expanduser().resolve()
        load_personal_home(resolved)
        return {
            "bound": True,
            "home_root": resolved,
            "source": "cli",
            "binding_path": "",
        }

    bound = load_home_binding(pack_dir)
    if not bound:
        return {
            "bound": False,
            "home_root": None,
            "source": "",
            "binding_path": "",
        }

    raw_home = str(bound.get("home_root", "")).strip()
    if not raw_home:
        return {
            "bound": False,
            "home_root": None,
            "source": "",
            "binding_path": display_path(home_binding_path(pack_dir)),
        }
    resolved = Path(raw_home).expanduser().resolve()
    load_personal_home(resolved)
    return {
        "bound": True,
        "home_root": resolved,
        "source": "binding",
        "binding_path": display_path(home_binding_path(pack_dir)),
    }


def recent_personal_daily_lines(home_root: Path, limit: int = 5) -> list[str]:
    manifest = load_personal_home(home_root)
    daily_dir = home_root / str(manifest["memory"]["daily_dir"])
    daily_files = sorted(daily_dir.glob("*.md"))
    if not daily_files:
        return []
    lines = extract_memory_lines(daily_files[-1])
    return lines[-limit:]


def build_memory_status(home_root: Path) -> dict[str, Any]:
    manifest = load_personal_home(home_root)
    limits = memory_limit_config(manifest)
    daily_dir = home_root / str(manifest["memory"]["daily_dir"])
    long_term_path = home_root / str(manifest["memory"]["long_term_path"])
    daily_files = sorted(daily_dir.glob("*.md"))
    daily_line_count = sum(len(extract_memory_lines(path)) for path in daily_files)
    long_term_text = long_term_path.read_text(encoding="utf-8") if long_term_path.exists() else ""
    long_term_chars = len(long_term_text)
    char_limit = limits["long_term_char_limit"]
    char_ratio = round(long_term_chars / max(1, char_limit), 3)
    last_dream_at = iso_to_datetime(manifest.get("last_dream_at"))
    latest_daily_at = None
    if daily_files:
        latest_daily_at = max(datetime.fromtimestamp(path.stat().st_mtime, tz=timezone.utc) for path in daily_files)
    daily_since_dream = 0
    if latest_daily_at is not None:
        if last_dream_at is None:
            daily_since_dream = len(daily_files)
        else:
            daily_since_dream = sum(
                1
                for path in daily_files
                if datetime.fromtimestamp(path.stat().st_mtime, tz=timezone.utc) > last_dream_at.astimezone(timezone.utc)
            )
    stale_signals: list[str] = []
    recommended_action = ""
    if daily_line_count >= limits["daily_compact_threshold_lines"] or daily_since_dream > 0 and last_dream_at is None:
        stale_signals.append("Daily memory has accumulated enough lines that dream-memory should run.")
        recommended_action = "dream-memory"
    if last_dream_at is not None:
        age_days = int((datetime.now(timezone.utc) - last_dream_at.astimezone(timezone.utc)).total_seconds() // 86400)
        if age_days >= limits["stale_after_days"] and daily_since_dream > 0:
            stale_signals.append(
                f"Last dream-memory run is {age_days} day(s) old while new daily notes have accumulated."
            )
            recommended_action = "dream-memory"
    if char_ratio >= 0.9:
        stale_signals.append("Long-term memory is close to its configured character budget.")
        if not recommended_action:
            recommended_action = "review-memory"
    return {
        "schema_version": "0.1.0",
        "status_type": "JiniMemoryStatus",
        "generated_at": now_utc(),
        "home_root": str(home_root.resolve()),
        "long_term_path": str(long_term_path),
        "daily_file_count": len(daily_files),
        "daily_line_count": daily_line_count,
        "daily_since_dream": daily_since_dream,
        "long_term_chars": long_term_chars,
        "long_term_char_limit": char_limit,
        "long_term_char_ratio": char_ratio,
        "last_dream_at": str(manifest.get("last_dream_at", "")),
        "recommended_action": recommended_action,
        "stale_signals": stale_signals,
    }


def build_personal_home_context(home_root: Path) -> dict[str, Any]:
    memory_status = build_memory_status(home_root)
    manifest = load_personal_home(home_root)
    long_term_path = home_root / str(manifest["memory"]["long_term_path"])
    long_term_lines: list[str] = []
    if long_term_path.exists():
        for line in extract_memory_lines(long_term_path):
            if line.startswith("Updated:") or line.startswith("Source files:"):
                continue
            if "memory/daily/" in line or line == "No source files":
                continue
            long_term_lines.append(line)
    recent_daily = recent_personal_daily_lines(home_root, limit=5)
    routines = list_personal_routines(home_root)["routines"]
    routine_ids = [str(item.get("routine_id", "")).strip() for item in routines if item.get("routine_id")]
    tools = list_personal_tools(home_root)["tools"]
    tool_ids = [str(item.get("id", "")).strip() for item in tools if item.get("id")]
    stale_signals: list[str] = []
    if not long_term_lines:
        stale_signals.append("Bound home has no durable long-term memory yet.")
    elif not manifest.get("last_dream_at"):
        stale_signals.append("Bound home has durable notes but no recorded dream-memory compression timestamp.")
    stale_signals.extend(memory_status.get("stale_signals", []))
    return {
        "bound": True,
        "home_root": str(home_root.resolve()),
        "assistant_name": str(manifest.get("assistant_name", "")),
        "owner_name": str(manifest.get("owner_name", "")),
        "last_dream_at": str(manifest.get("last_dream_at", "")),
        "long_term_memory": long_term_lines[:8],
        "recent_daily_lines": recent_daily,
        "routine_ids": routine_ids,
        "tool_ids": tool_ids,
        "stale_signals": stale_signals,
        "memory_status": memory_status,
    }


def append_home_observation(
    pack_dir: Path,
    *,
    home_binding: dict[str, Any],
    line: str,
) -> dict[str, Any]:
    if not home_binding.get("bound") or home_binding.get("home_root") is None:
        return {
            "appended": False,
            "home_root": "",
            "daily_path": "",
            "line": "",
        }
    result = append_memory_line(Path(home_binding["home_root"]), line=line)
    return {
        "appended": True,
        "home_root": str(Path(home_binding["home_root"]).resolve()),
        "daily_path": result["daily_path"],
        "line": result["line"],
    }


def display_path(path: Path) -> str:
    try:
        return str(path.relative_to(ROOT))
    except ValueError:
        return str(path)


def cli_invocation() -> str:
    override = os.environ.get("JINI_CLI_COMMAND", "").strip()
    if override:
        return override
    executable = Path(sys.argv[0]).name
    if executable == "jini":
        return "jini"
    return "jini"


CLI_ALIAS_MAP: dict[str, str] = {
    "start": "get-started",
    "example": "try-example",
    "next": "execution-checklist",
    "resume": "compact-context",
}


def normalize_cli_argv(argv: list[str]) -> list[str]:
    if not argv:
        return []
    normalized = list(argv)
    normalized[0] = CLI_ALIAS_MAP.get(normalized[0], normalized[0])
    return normalized


def print_cli_overview() -> None:
    cli = cli_invocation()
    print(f"Jini CLI {load_version()}")
    print("Finish work with less rework, faster handoffs, and clearer next steps.")
    print()
    print("START HERE")
    print(f"  {cli} help")
    print(f"  {cli} example research-prd")
    print(f"  {cli} start --harness codex")
    print(f"  {cli} harnesses")
    print()
    print("OUTCOME LAYER")
    print(f"  {cli} outcome /path/to/work")
    print(f"  {cli} next /path/to/work --repo /path/to/repo --intent verify")
    print(f"  {cli} resume /path/to/work --repo /path/to/repo --intent verify --max-chars 900")
    print(f"  {cli} execute-flow /path/to/work --repo /path/to/repo --harness codex")
    print()
    print("INSTALL")
    print(f"  {cli} plan-install --kit starter-kit --harness codex")
    print(f"  {cli} install-bundles --kit starter-kit --harness codex --prefix /tmp/jini-stage")
    print(f"  {cli} doctor-install --kit starter-kit --harness codex --prefix /tmp/jini-stage")
    print()
    print("ALIASES")
    print(f"  {cli} start   -> {cli} get-started")
    print(f"  {cli} example -> {cli} try-example")
    print(f"  {cli} next    -> {cli} execution-checklist")
    print(f"  {cli} resume  -> {cli} compact-context")
    print()
    print("MORE")
    print(f"  {cli} help --all")
    print(f"  {cli} <command> --help")


def resolve_display_path(path_text: str) -> Path:
    path = Path(path_text).expanduser()
    if path.is_absolute():
        return path
    return ROOT / path


def trim_output(text: str | bytes | None, limit: int = 1200) -> str:
    if text is None:
        return ""
    if isinstance(text, bytes):
        normalized = text.decode("utf-8", errors="replace")
    else:
        normalized = text
    normalized = normalized.strip()
    if len(normalized) <= limit:
        return normalized
    return normalized[-limit:]


def runtime_events_path(pack_dir: Path) -> Path:
    runtime_dir = pack_dir / "runtime"
    runtime_dir.mkdir(parents=True, exist_ok=True)
    return runtime_dir / "events.jsonl"


def append_learning_event(
    event_type: str,
    payload: dict[str, Any],
    *,
    pack_dir: Path | None = None,
) -> dict[str, str]:
    event = {
        "schema_version": "0.1.0",
        "event_type": event_type,
        "recorded_at": now_utc(),
        **payload,
    }
    line = json.dumps(event, sort_keys=True) + "\n"

    LEARNING_EVENTS_ROOT.mkdir(parents=True, exist_ok=True)
    LEARNING_EVENTS_PATH.write_text(
        (LEARNING_EVENTS_PATH.read_text(encoding="utf-8") if LEARNING_EVENTS_PATH.exists() else "") + line,
        encoding="utf-8",
    )
    paths = {"global_path": display_path(LEARNING_EVENTS_PATH)}
    if pack_dir is not None:
        pack_event_path = runtime_events_path(pack_dir)
        pack_event_path.write_text(
            (pack_event_path.read_text(encoding="utf-8") if pack_event_path.exists() else "") + line,
            encoding="utf-8",
        )
        paths["pack_path"] = display_path(pack_event_path)
    return paths


def read_learning_events(
    *,
    path: Path | None = None,
    limit: int = 20,
    event_type: str | None = None,
    work_unit_id: str | None = None,
) -> list[dict[str, Any]]:
    source_path = path if path is not None else LEARNING_EVENTS_PATH
    if not source_path.exists():
        return []
    entries: list[dict[str, Any]] = []
    for line in source_path.read_text(encoding="utf-8").splitlines():
        if not line.strip():
            continue
        try:
            entry = json.loads(line)
        except json.JSONDecodeError:
            continue
        if event_type and entry.get("event_type") != event_type:
            continue
        if work_unit_id and entry.get("work_unit_id") != work_unit_id:
            continue
        if isinstance(entry, dict):
            entries.append(entry)
    if limit > 0:
        return entries[-limit:]
    return entries


def run_git(repo_root: Path, *args: str) -> str | None:
    try:
        result = subprocess.run(
            ["git", "-C", str(repo_root), *args],
            check=False,
            capture_output=True,
            text=True,
        )
    except OSError:
        return None
    if result.returncode != 0:
        return None
    return result.stdout.strip()


def parse_make_targets(path: Path) -> list[str]:
    targets: list[str] = []
    for line in path.read_text(encoding="utf-8").splitlines():
        match = re.match(r"^([A-Za-z0-9_.-]+):(?:\s|$)", line)
        if not match:
            continue
        target = match.group(1)
        if target.startswith("."):
            continue
        targets.append(target)
    return list(dict.fromkeys(targets))


def parse_just_targets(path: Path) -> list[str]:
    targets: list[str] = []
    for line in path.read_text(encoding="utf-8").splitlines():
        stripped = line.strip()
        if not stripped or stripped.startswith("#"):
            continue
        match = re.match(r"^([A-Za-z0-9_.-]+):(?:\s|$)", stripped)
        if not match:
            continue
        targets.append(match.group(1))
    return list(dict.fromkeys(targets))


def detect_node_package_manager(repo_root: Path) -> str:
    if (repo_root / "pnpm-lock.yaml").exists():
        return "pnpm"
    if (repo_root / "yarn.lock").exists():
        return "yarn"
    if (repo_root / "bun.lock").exists() or (repo_root / "bun.lockb").exists():
        return "bun"
    return "npm"


def command_for_script(package_manager: str, script_name: str) -> tuple[str, list[str]]:
    if package_manager == "npm":
        if script_name == "test":
            return ("npm test", ["npm", "test"])
        return (f"npm run {script_name}", ["npm", "run", script_name])
    if package_manager == "yarn":
        return (f"yarn {script_name}", ["yarn", script_name])
    if package_manager == "pnpm":
        return (f"pnpm {script_name}", ["pnpm", script_name])
    if package_manager == "bun":
        return (f"bun run {script_name}", ["bun", "run", script_name])
    return (f"{package_manager} {script_name}", [package_manager, script_name])


def command_for_script_path(script_path: Path) -> tuple[str, list[str]]:
    suffix = script_path.suffix.lower()
    rendered = display_path(script_path)
    if suffix == ".sh":
        return (f"sh {rendered}", ["sh", rendered])
    if suffix == ".py":
        return (f"python3 {rendered}", ["python3", rendered])
    if suffix in {".js", ".mjs", ".cjs"}:
        return (f"node {rendered}", ["node", rendered])
    return (rendered, [rendered])


def iso_to_datetime(value: Any) -> datetime | None:
    if not isinstance(value, str) or not value.strip():
        return None
    normalized = value.strip()
    if normalized.endswith("Z"):
        normalized = normalized[:-1] + "+00:00"
    try:
        return datetime.fromisoformat(normalized)
    except ValueError:
        return None


def age_summary(value: Any) -> str:
    moment = iso_to_datetime(value)
    if moment is None:
        return ""
    delta = datetime.now(timezone.utc) - moment.astimezone(timezone.utc)
    seconds = max(0, int(delta.total_seconds()))
    if seconds < 60:
        return f"{seconds}s ago"
    if seconds < 3600:
        return f"{seconds // 60}m ago"
    if seconds < 86400:
        return f"{seconds // 3600}h ago"
    return f"{seconds // 86400}d ago"


def estimate_tokens(value: Any) -> int:
    if isinstance(value, str):
        text = value
    else:
        text = json.dumps(value, sort_keys=True)
    return max(1, (len(text) + 3) // 4)


def artifact_for_entrypoint(category: str) -> str:
    if category in {"test", "verify", "demo"}:
        return "Evidence"
    if category in {"startup", "operations"}:
        return "Runbook"
    if category == "docs":
        return "Spec"
    return "Plan"


def append_repo_entry(
    repo_context: dict[str, Any],
    category: str,
    *,
    label: str,
    path: Path | None = None,
    command: str | None = None,
    argv: list[str] | None = None,
    source: str | None = None,
) -> None:
    entry: dict[str, Any] = {"label": label}
    if path is not None:
        entry["path"] = display_path(path)
    if command is not None:
        entry["command"] = command
    if argv is not None:
        entry["argv"] = [str(item) for item in argv]
    if source is not None:
        entry["source"] = source
    repo_context["entrypoints"].setdefault(category, []).append(entry)


def inspect_repo_context(pack_dir: Path, repo_path: Path | None = None) -> dict[str, Any]:
    requested = repo_path.resolve() if repo_path is not None else None
    repo_root = requested if requested is not None else None

    context: dict[str, Any] = {
        "requested_path": display_path(requested) if requested is not None else "",
        "repo_root": "",
        "discovered": False,
        "git": {
            "tracked": False,
            "branch": "",
            "dirty": False,
            "dirty_files": 0,
        },
        "entrypoints": {
            "build": [],
            "test": [],
            "startup": [],
            "verify": [],
            "demo": [],
            "docs": [],
            "operations": [],
        },
        "verification_targets": [],
        "steering": {
            "found": False,
            "documents": [],
            "active_paths": [],
        },
        "next_actions": [],
        "notes": [],
    }

    if repo_root is None:
        search_roots = [pack_dir, *pack_dir.parents]
        marker_names = (
            ".git",
            "package.json",
            "pyproject.toml",
            "Cargo.toml",
            "go.mod",
            "Justfile",
            "justfile",
            "tox.ini",
            "noxfile.py",
            "docker-compose.yml",
            "docker-compose.yaml",
            "compose.yml",
            "compose.yaml",
            "Makefile",
        )
        for candidate in search_roots:
            if any((candidate / marker).exists() for marker in marker_names):
                repo_root = candidate
                break
            if candidate == ROOT:
                break

    if repo_root is None or not repo_root.exists():
        context["notes"].append("No repo context detected. Pass --repo to enable repo-aware guidance.")
        return context

    context["discovered"] = True
    context["repo_root"] = display_path(repo_root)
    context["steering"] = build_steering_summary(repo_root)

    git_tracked = False
    if (repo_root / ".git").exists():
        git_tracked = True
    else:
        inside = run_git(repo_root, "rev-parse", "--is-inside-work-tree")
        git_tracked = inside == "true"

    context["git"]["tracked"] = git_tracked
    if git_tracked:
        branch = run_git(repo_root, "rev-parse", "--abbrev-ref", "HEAD") or ""
        dirty_output = run_git(repo_root, "status", "--porcelain") or ""
        dirty_files = len([line for line in dirty_output.splitlines() if line.strip()])
        context["git"]["branch"] = branch
        context["git"]["dirty"] = dirty_files > 0
        context["git"]["dirty_files"] = dirty_files

    readme_path = repo_root / "README.md"
    if readme_path.exists():
        append_repo_entry(context, "docs", label="Project README", path=readme_path, source="README.md")

    docs_dir = repo_root / "docs"
    if docs_dir.exists():
        append_repo_entry(context, "docs", label="Docs directory", path=docs_dir, source="docs/")

    workflows_dir = repo_root / ".github" / "workflows"
    if workflows_dir.exists():
        append_repo_entry(
            context,
            "docs",
            label="CI workflows",
            path=workflows_dir,
            source=".github/workflows/",
        )

    makefile_path = repo_root / "Makefile"
    if makefile_path.exists():
        make_targets = parse_make_targets(makefile_path)
        for target in make_targets:
            lowered = target.lower()
            category = "operations"
            if any(token in lowered for token in ("test", "check", "lint")):
                category = "test"
            elif any(token in lowered for token in ("verify", "smoke", "health")):
                category = "verify"
            elif any(token in lowered for token in ("demo", "walkthrough")):
                category = "demo"
            elif any(token in lowered for token in ("run", "start", "up", "boot", "dev")):
                category = "startup"
            elif any(token in lowered for token in ("build", "compile")):
                category = "build"
            append_repo_entry(
                context,
                category,
                label=f"make {target}",
                command=f"make {target}",
                argv=["make", target],
                source="Makefile",
            )

    for just_name in ("Justfile", "justfile"):
        justfile_path = repo_root / just_name
        if not justfile_path.exists():
            continue
        for target in parse_just_targets(justfile_path):
            lowered = target.lower()
            category = "operations"
            if any(token in lowered for token in ("test", "check", "lint")):
                category = "test"
            elif any(token in lowered for token in ("verify", "smoke", "health")):
                category = "verify"
            elif "demo" in lowered:
                category = "demo"
            elif any(token in lowered for token in ("run", "start", "up", "boot", "dev")):
                category = "startup"
            elif any(token in lowered for token in ("build", "compile")):
                category = "build"
            append_repo_entry(
                context,
                category,
                label=f"just {target}",
                command=f"just {target}",
                argv=["just", target],
                source=just_name,
            )

    package_json_path = repo_root / "package.json"
    if package_json_path.exists():
        try:
            package_manager = detect_node_package_manager(repo_root)
            package_doc = json.loads(package_json_path.read_text(encoding="utf-8"))
            scripts = package_doc.get("scripts", {})
        except (OSError, json.JSONDecodeError):
            context["notes"].append("package.json exists but could not be parsed for script guidance.")
        else:
            if isinstance(scripts, dict):
                for script_name in sorted(scripts):
                    lowered = script_name.lower()
                    category = "operations"
                    if lowered == "test" or "test" in lowered or "lint" in lowered or "check" in lowered:
                        category = "test"
                    elif "verify" in lowered or "smoke" in lowered:
                        category = "verify"
                    elif "demo" in lowered:
                        category = "demo"
                    elif lowered in {"dev", "start"} or "start" in lowered or "serve" in lowered:
                        category = "startup"
                    elif "build" in lowered or "compile" in lowered:
                        category = "build"
                    display_command, argv = command_for_script(package_manager, script_name)
                    append_repo_entry(
                        context,
                        category,
                        label=f"{package_manager}:{script_name}",
                        command=display_command,
                        argv=argv,
                        source="package.json",
                    )

    pyproject_path = repo_root / "pyproject.toml"
    if pyproject_path.exists():
        append_repo_entry(context, "docs", label="Python project config", path=pyproject_path, source="pyproject.toml")
        if tomllib is not None:
            try:
                pyproject_doc = tomllib.loads(pyproject_path.read_text(encoding="utf-8"))
            except (OSError, tomllib.TOMLDecodeError):
                context["notes"].append("pyproject.toml exists but could not be parsed for test guidance.")
            else:
                tool_section = pyproject_doc.get("tool", {}) if isinstance(pyproject_doc, dict) else {}
                if isinstance(tool_section, dict) and "pytest" in tool_section:
                    append_repo_entry(
                        context,
                        "test",
                        label="pytest",
                        command="pytest",
                        argv=["pytest"],
                        source="pyproject.toml",
                    )

    cargo_toml_path = repo_root / "Cargo.toml"
    if cargo_toml_path.exists():
        append_repo_entry(context, "docs", label="Rust project config", path=cargo_toml_path, source="Cargo.toml")
        append_repo_entry(
            context, "build", label="cargo build", command="cargo build", argv=["cargo", "build"], source="Cargo.toml"
        )
        append_repo_entry(
            context, "test", label="cargo test", command="cargo test", argv=["cargo", "test"], source="Cargo.toml"
        )
        append_repo_entry(
            context, "startup", label="cargo run", command="cargo run", argv=["cargo", "run"], source="Cargo.toml"
        )

    go_mod_path = repo_root / "go.mod"
    if go_mod_path.exists():
        append_repo_entry(context, "docs", label="Go module", path=go_mod_path, source="go.mod")
        append_repo_entry(
            context, "build", label="go build ./...", command="go build ./...", argv=["go", "build", "./..."], source="go.mod"
        )
        append_repo_entry(
            context, "test", label="go test ./...", command="go test ./...", argv=["go", "test", "./..."], source="go.mod"
        )
        append_repo_entry(
            context, "startup", label="go run .", command="go run .", argv=["go", "run", "."], source="go.mod"
        )

    requirements_path = repo_root / "requirements.txt"
    if requirements_path.exists():
        append_repo_entry(
            context,
            "docs",
            label="Python requirements",
            path=requirements_path,
            source="requirements.txt",
        )

    pytest_ini_path = repo_root / "pytest.ini"
    if pytest_ini_path.exists():
        append_repo_entry(context, "test", label="pytest", command="pytest", argv=["pytest"], source="pytest.ini")

    tox_ini_path = repo_root / "tox.ini"
    if tox_ini_path.exists():
        append_repo_entry(context, "test", label="tox", command="tox", argv=["tox"], source="tox.ini")

    noxfile_path = repo_root / "noxfile.py"
    if noxfile_path.exists():
        append_repo_entry(context, "test", label="nox", command="nox", argv=["nox"], source="noxfile.py")

    for name in ("docker-compose.yml", "docker-compose.yaml", "compose.yml", "compose.yaml"):
        compose_path = repo_root / name
        if compose_path.exists():
            append_repo_entry(
                context,
                "startup",
                label=f"docker compose up ({name})",
                path=compose_path,
                command=f"docker compose -f {display_path(compose_path)} up",
                argv=["docker", "compose", "-f", display_path(compose_path), "up"],
                source=name,
            )
            append_repo_entry(
                context,
                "operations",
                label=f"docker compose ps ({name})",
                path=compose_path,
                command=f"docker compose -f {display_path(compose_path)} ps",
                argv=["docker", "compose", "-f", display_path(compose_path), "ps"],
                source=name,
            )

    dockerfile_path = repo_root / "Dockerfile"
    if dockerfile_path.exists():
        append_repo_entry(context, "build", label="Dockerfile", path=dockerfile_path, source="Dockerfile")

    for folder_name in ("tests", "test", "__tests__"):
        test_dir = repo_root / folder_name
        if test_dir.exists():
            append_repo_entry(context, "test", label=f"Test directory ({folder_name})", path=test_dir, source=folder_name)

    manage_py_path = repo_root / "manage.py"
    if manage_py_path.exists():
        append_repo_entry(
            context,
            "startup",
            label="Django dev server",
            command="python3 manage.py runserver",
            argv=["python3", "manage.py", "runserver"],
            path=manage_py_path,
            source="manage.py",
        )
        append_repo_entry(
            context,
            "verify",
            label="Django system check",
            command="python3 manage.py check",
            argv=["python3", "manage.py", "check"],
            path=manage_py_path,
            source="manage.py",
        )

    for python_entry in ("app.py", "main.py"):
        python_path = repo_root / python_entry
        if python_path.exists():
            append_repo_entry(
                context,
                "startup",
                label=f"python3 {python_entry}",
                command=f"python3 {python_entry}",
                argv=["python3", python_entry],
                path=python_path,
                source=python_entry,
            )

    scripts_dir = repo_root / "scripts"
    if scripts_dir.exists():
        for script_path in sorted(scripts_dir.iterdir()):
            if not script_path.is_file():
                continue
            lowered = script_path.name.lower()
            category = "operations"
            if any(token in lowered for token in ("test", "check")):
                category = "test"
            elif any(token in lowered for token in ("verify", "health", "smoke")):
                category = "verify"
            elif "demo" in lowered:
                category = "demo"
            elif any(token in lowered for token in ("start", "run", "boot", "up")):
                category = "startup"
            display_command, argv = command_for_script_path(script_path)
            append_repo_entry(
                context,
                category,
                label=script_path.name,
                path=script_path,
                command=display_command,
                argv=argv,
                source="scripts/",
            )

    if not any(context["entrypoints"].values()):
        context["notes"].append("Repo detected, but no standard startup, test, build, or docs entrypoints were found.")

    verification_targets: list[dict[str, Any]] = []
    for category in ("test", "verify", "startup", "demo", "docs"):
        for entry in context["entrypoints"].get(category, [])[:3]:
            target: dict[str, Any] = {
                "artifact": artifact_for_entrypoint(category),
                "category": category,
                "label": entry["label"],
            }
            if "command" in entry:
                target["command"] = entry["command"]
            if "argv" in entry:
                target["argv"] = deepcopy(entry["argv"])
            if "path" in entry:
                target["path"] = entry["path"]
            verification_targets.append(target)
    context["verification_targets"] = verification_targets
    return context


def detect_repo_stack(repo_root: Path) -> list[str]:
    markers = [
        ("node", repo_root / "package.json"),
        ("python", repo_root / "pyproject.toml"),
        ("python", repo_root / "requirements.txt"),
        ("rust", repo_root / "Cargo.toml"),
        ("go", repo_root / "go.mod"),
        ("docker", repo_root / "Dockerfile"),
        ("docker-compose", repo_root / "docker-compose.yml"),
        ("docker-compose", repo_root / "compose.yml"),
        ("django", repo_root / "manage.py"),
    ]
    stack: list[str] = []
    for label, path in markers:
        if path.exists() and label not in stack:
            stack.append(label)
    return stack


def build_repo_map(
    pack_dir: Path,
    *,
    repo_path: Path | None = None,
    max_entries: int = 8,
) -> dict[str, Any]:
    repo_context = inspect_repo_context(pack_dir, repo_path=repo_path)
    if not repo_context.get("discovered"):
        raise ValueError("repo-map requires a repo or worktree path")
    repo_root = resolve_display_path(str(repo_context.get("repo_root", "")))
    top_level_files: list[str] = []
    top_level_dirs: list[str] = []
    for child in sorted(repo_root.iterdir(), key=lambda item: item.name):
        if child.name in {".git", ".jini", ".kiro", "__pycache__"}:
            continue
        if child.is_dir():
            top_level_dirs.append(child.name)
        else:
            top_level_files.append(child.name)
    entrypoint_summary = {
        category: [
            entry.get("command") or entry.get("path") or entry.get("label", "")
            for entry in repo_context.get("entrypoints", {}).get(category, [])[: max(1, min(3, max_entries))]
        ]
        for category in ("build", "test", "startup", "verify", "demo", "docs", "operations")
    }
    steering = repo_context.get("steering", {})
    return {
        "schema_version": "0.1.0",
        "map_type": "JiniRepoMap",
        "generated_at": now_utc(),
        "repo_root": str(repo_root),
        "git": repo_context.get("git", {}),
        "detected_stack": detect_repo_stack(repo_root),
        "top_level_files": top_level_files[:max_entries],
        "top_level_dirs": top_level_dirs[:max_entries],
        "entrypoints": entrypoint_summary,
        "verification_targets": repo_context.get("verification_targets", [])[:max_entries],
        "steering": steering,
    }


def print_repo_map(repo_map: dict[str, Any]) -> None:
    print(f"REPO   {repo_map['repo_root']}")
    git_info = repo_map.get("git", {})
    if git_info.get("tracked"):
        print(
            f"GIT    branch={git_info.get('branch', '') or 'unknown'} dirty={git_info.get('dirty', False)} files={git_info.get('dirty_files', 0)}"
        )
    if repo_map.get("detected_stack"):
        print(f"STACK  {', '.join(repo_map['detected_stack'])}")
    if repo_map.get("top_level_dirs"):
        print("DIRS")
        for item in repo_map["top_level_dirs"]:
            print(f"  - {item}")
    if repo_map.get("top_level_files"):
        print("FILES")
        for item in repo_map["top_level_files"]:
            print(f"  - {item}")
    steering = repo_map.get("steering", {})
    if steering.get("found"):
        print("STEERING")
        for path in steering.get("active_paths", [])[:6]:
            print(f"  - {path}")
    print("ENTRYPOINTS")
    for category, entries in repo_map.get("entrypoints", {}).items():
        if not entries:
            continue
        print(f"  {category}:")
        for item in entries:
            print(f"    - {item}")


def apply_repo_guidance(
    recommendation: dict[str, Any],
    repo_context: dict[str, Any],
) -> None:
    recommendation["repo_context"] = repo_context
    rationale = recommendation["rationale"]
    tool_order = recommendation["tool_order"]

    if not repo_context.get("discovered"):
        rationale.append("No repo context was detected; pass --repo to get concrete file and command guidance")
        return

    git_info = repo_context.get("git", {})
    if git_info.get("tracked"):
        branch = git_info.get("branch") or "unknown"
        dirty_files = int(git_info.get("dirty_files", 0))
        if dirty_files > 0:
            rationale.append(f"Repo branch `{branch}` has {dirty_files} dirty file(s); prefer targeted changes and explicit verification")
        else:
            rationale.append(f"Repo branch `{branch}` is clean; verification can target the detected entrypoints directly")
    steering = repo_context.get("steering", {})
    if steering.get("found"):
        active = steering.get("active_paths", [])
        rationale.append(
            f"Workspace steering is available with {len(active) or len(steering.get('documents', []))} steering document(s)."
        )

    intent = recommendation["intent"]
    next_actions: list[str] = []
    startup_entries = repo_context["entrypoints"].get("startup", [])
    test_entries = repo_context["entrypoints"].get("test", [])
    verify_entries = repo_context["entrypoints"].get("verify", [])
    doc_entries = repo_context["entrypoints"].get("docs", [])
    demo_entries = repo_context["entrypoints"].get("demo", [])

    if intent in {"model", "decide", "make"}:
        if startup_entries:
            entry = startup_entries[0]
            detail = entry.get("path") or entry.get("command") or entry["label"]
            next_actions.append(f"Inspect startup surface `{detail}` before changing delivery or ops claims")
        if test_entries:
            entry = test_entries[0]
            detail = entry.get("command") or entry.get("path") or entry["label"]
            next_actions.append(f"Use `{detail}` as the primary local verification path for make-stage work")
        if doc_entries:
            entry = doc_entries[0]
            detail = entry.get("path") or entry.get("command") or entry["label"]
            next_actions.append(f"Anchor spec or plan updates against `{detail}`")
    elif intent == "verify":
        preferred = verify_entries or test_entries or startup_entries
        for entry in preferred[:2]:
            detail = entry.get("command") or entry.get("path") or entry["label"]
            next_actions.append(f"Bind verification evidence to `{detail}`")
        if demo_entries:
            entry = demo_entries[0]
            detail = entry.get("command") or entry.get("path") or entry["label"]
            next_actions.append(f"Use `{detail}` as a demo-check surface before operational promotion")
        next_actions.append(
            "Run `harvest-evidence` against the detected repo verification targets before approval or operational promotion"
        )
    elif intent in {"scope", "probe", "research"}:
        if doc_entries:
            entry = doc_entries[0]
            detail = entry.get("path") or entry.get("command") or entry["label"]
            next_actions.append(f"Read `{detail}` first to ground project context before widening search")
        if startup_entries:
            entry = startup_entries[0]
            detail = entry.get("path") or entry.get("command") or entry["label"]
            next_actions.append(f"Use `{detail}` to understand the real delivery topology")
    else:
        if doc_entries:
            entry = doc_entries[0]
            detail = entry.get("path") or entry.get("command") or entry["label"]
            next_actions.append(f"Use `{detail}` as the canonical local narrative surface")

    if steering.get("found") and steering.get("active_paths"):
        next_actions.insert(0, f"Load steering docs first: {', '.join(steering['active_paths'][:3])}")

    if startup_entries or test_entries or verify_entries:
        tool_order.insert(0, "repo entrypoints and worktree state")
        recommendation["context_policy"] = "repo-targeted"
    repo_context["next_actions"] = next_actions


def snapshot_artifact(path: Path, doc: dict[str, Any], artifact_type: str) -> dict[str, Any]:
    return {
        "artifact_type": artifact_type,
        "path": display_path(path),
        "artifact_id": doc.get("artifact_id", ""),
        "status": doc.get("status", ""),
        "revision": int(doc.get("revision", 0)),
        "updated_at": doc.get("updated_at", ""),
    }


def build_memory_context(
    summary: dict[str, Any],
    intent: str,
    *,
    latest_harvest: dict[str, Any] | None = None,
    latest_run: dict[str, Any] | None = None,
    home_binding: dict[str, Any] | None = None,
    repo_context: dict[str, Any] | None = None,
) -> dict[str, Any]:
    latest_by_type = summary["latest_by_type"]
    task_summary = summary["task_summary"]
    state = str(summary["work_unit"].get("current_state", ""))
    artifact_order = {
        "scope": ["Brief", "Assumptions"],
        "probe": ["Brief", "Assumptions", "Sources", "Literature", "Method"],
        "research": ["Brief", "Assumptions", "Sources", "Literature", "Method"],
        "model": ["Brief", "Assumptions", "Spec", "Decision"],
        "decide": ["Spec", "Decision", "Plan", "Tasks"],
        "make": ["Spec", "Decision", "Plan", "Tasks", "Runbook", "Signals", "Rollback"],
        "verify": ["Tasks", "Evidence", "Approval", "Runbook", "Signals", "Rollback"],
        "publish": ["Tasks", "Evidence", "Approval", "Publication"],
        "issues": ["Tasks", "Plan"],
        "wiki": ["Brief", "Spec", "Plan", "Tasks"],
        "export": ["Plan", "Tasks", "Evidence"],
    }
    selected_types = artifact_order.get(intent, ["Brief", "Spec", "Plan", "Tasks"])
    recent_artifacts = [
        snapshot_artifact(*latest_by_type[artifact_type], artifact_type)
        for artifact_type in selected_types
        if artifact_type in latest_by_type
    ]

    indexes = []
    for path in (ROOT / "knowledge" / "index.md", ROOT / "projects" / "index.md", ROOT / "people" / "index.md"):
        if path.exists():
            indexes.append({"label": path.parent.name, "path": display_path(path)})

    stale_signals: list[str] = []
    home_context: dict[str, Any] = {
        "bound": False,
        "home_root": "",
        "assistant_name": "",
        "owner_name": "",
        "last_dream_at": "",
        "long_term_memory": [],
        "recent_daily_lines": [],
        "routine_ids": [],
        "tool_ids": [],
        "stale_signals": [],
    }
    if intent == "verify":
        if "Evidence" not in latest_by_type:
            stale_signals.append("No Evidence artifact is captured yet for verification.")
        if task_summary["unresolved"] > 0:
            stale_signals.append(
                f"{task_summary['unresolved']} unresolved task(s) remain before verification can be trusted."
            )
        if latest_harvest is None:
            stale_signals.append("No harvest report has been recorded for the current verification surface.")
        elif latest_harvest.get("readiness") != "ready":
            stale_signals.append(
                f"Latest harvest is `{latest_harvest.get('readiness', '')}` with evidence status "
                f"`{latest_harvest.get('evidence_status', '')}`."
            )

    evidence_entry = latest_by_type.get("Evidence")
    spec_entry = latest_by_type.get("Spec")
    if evidence_entry is not None and spec_entry is not None:
        evidence_doc = evidence_entry[1]
        target_revision = int(evidence_doc.get("target_revision", 0) or 0)
        spec_revision = int(spec_entry[1].get("revision", 0) or 0)
        if target_revision and spec_revision and target_revision < spec_revision:
            stale_signals.append(
                f"Evidence targets Spec revision {target_revision}, but the latest Spec is revision {spec_revision}."
            )

    if latest_run is not None and latest_run.get("state_after") != state:
        stale_signals.append(
            f"Latest run recorded state `{latest_run.get('state_after', '')}`, but the canonical state is `{state}`."
        )

    if home_binding and home_binding.get("bound") and home_binding.get("home_root") is not None:
        home_context = build_personal_home_context(Path(home_binding["home_root"]))
        indexes.append({"label": "personal-home", "path": home_context["home_root"]})
        stale_signals.extend(home_context.get("stale_signals", []))
    steering_paths: list[str] = []
    if repo_context is not None:
        steering = repo_context.get("steering", {})
        if isinstance(steering, dict):
            steering_paths = [str(item) for item in steering.get("active_paths", [])[:4]]

    freshness: list[dict[str, str]] = []
    for artifact in recent_artifacts[:4]:
        freshness.append(
            {
                "artifact_type": artifact["artifact_type"],
                "updated_at": artifact["updated_at"],
                "age": age_summary(artifact["updated_at"]),
            }
        )
    if latest_harvest is not None:
        freshness.append(
            {
                "artifact_type": "Harvest",
                "updated_at": str(latest_harvest.get("generated_at", "")),
                "age": age_summary(latest_harvest.get("generated_at", "")),
            }
        )
    if home_context.get("bound"):
        freshness.append(
            {
                "artifact_type": "PersonalHome",
                "updated_at": str(home_context.get("last_dream_at", "")),
                "age": age_summary(home_context.get("last_dream_at", "")),
            }
        )

    resume_items = [
        f"State anchor: `{state}` with health `{summary.get('health', '')}` and next operation `{summary.get('next_operation', '')}`.",
    ]
    if task_summary["total"] > 0:
        resume_items.append(
            f"Task status: {task_summary['done']}/{task_summary['total']} done, {task_summary['unresolved']} unresolved."
        )
    if latest_harvest is not None:
        resume_items.append(
            f"Harvest anchor: `{latest_harvest.get('readiness', '')}` / `{latest_harvest.get('evidence_status', '')}` "
            f"from `{latest_harvest.get('path', '')}` ({age_summary(latest_harvest.get('generated_at', '')) or 'unknown age'})."
        )
    for artifact in recent_artifacts[:4]:
        resume_items.append(
            f"{artifact['artifact_type']} rev {artifact['revision']} is `{artifact['status']}` at `{artifact['path']}`."
        )
    if home_context.get("bound"):
        if home_context.get("long_term_memory"):
            resume_items.append(f"Home memory: {home_context['long_term_memory'][0]}")
        if home_context.get("recent_daily_lines"):
            resume_items.append(f"Recent daily note: {home_context['recent_daily_lines'][-1]}")
        if home_context.get("routine_ids"):
            resume_items.append(
                "Home routines: " + ", ".join(home_context["routine_ids"][:3])
            )
        memory_status = home_context.get("memory_status", {})
        if memory_status.get("recommended_action"):
            resume_items.append(f"Home memory action: {memory_status['recommended_action']}")
    if steering_paths:
        resume_items.append("Steering docs: " + ", ".join(steering_paths[:3]))

    return {
        "state_anchor": state,
        "recent_artifacts": recent_artifacts,
        "indexes": indexes,
        "stale_signals": stale_signals,
        "freshness": freshness,
        "resume_items": resume_items,
        "home": home_context,
        "steering_paths": steering_paths,
    }


def unresolved_task_items(summary: dict[str, Any], limit: int = 3) -> list[dict[str, str]]:
    latest_by_type = summary["latest_by_type"]
    tasks_entry = latest_by_type.get("Tasks")
    if tasks_entry is None:
        return []
    _, tasks_doc = tasks_entry
    tasks = list(tasks_doc.get("tasks", []))
    statuses = [str(value).strip() for value in tasks_doc.get("status_per_task", [])]
    ownership = [str(value).strip() for value in tasks_doc.get("ownership", [])]
    items: list[dict[str, str]] = []
    for task, status, owner in zip_longest(tasks, statuses, ownership, fillvalue=""):
        if str(status).strip().lower() in DONE_TASK_STATUSES:
            continue
        items.append(
            {
                "task": str(task).strip(),
                "status": str(status).strip() or "pending",
                "owner": str(owner).strip(),
            }
        )
        if len(items) >= limit:
            break
    return items


def latest_harvest_report_summary(pack_dir: Path) -> dict[str, Any] | None:
    harvest_dir = pack_dir / "runtime" / "harvests"
    if not harvest_dir.exists():
        return None
    candidates = sorted(harvest_dir.glob("evidence-harvest-*.json"))
    if not candidates:
        return None
    report_path = candidates[-1]
    try:
        report = load_document(report_path)
    except (OSError, json.JSONDecodeError, yaml.YAMLError):
        return None
    if not isinstance(report, dict):
        return None
    summary = report.get("summary", {})
    return {
        "path": display_path(report_path),
        "generated_at": report.get("generated_at", ""),
        "readiness": report.get("readiness", ""),
        "evidence_status": report.get("evidence_status", ""),
        "passed": summary.get("passed", 0) if isinstance(summary, dict) else 0,
        "failed": summary.get("failed", 0) if isinstance(summary, dict) else 0,
        "timed_out": summary.get("timed_out", 0) if isinstance(summary, dict) else 0,
    }


def latest_run_report_summary(pack_dir: Path) -> dict[str, Any] | None:
    report_path = pack_dir / "runtime" / "last-run.json"
    if not report_path.exists():
        return None
    try:
        report = load_document(report_path)
    except (OSError, json.JSONDecodeError, yaml.YAMLError):
        return None
    if not isinstance(report, dict):
        return None
    return {
        "path": display_path(report_path),
        "generated_at": report.get("generated_at", ""),
        "mode": report.get("mode", ""),
        "intent": report.get("intent", ""),
        "state_after": report.get("state_after", ""),
        "health_after": report.get("health_after", ""),
        "actions": len(report.get("actions", [])),
        "blockers": len(report.get("blockers", [])),
    }


def build_compact_context(
    pack_dir: Path,
    registry: dict[str, Any],
    *,
    intent: str | None = None,
    repo_path: Path | None = None,
    home_path: Path | None = None,
    runtime_target: str | None = None,
    max_items: int = 5,
    max_chars: int = 1200,
) -> dict[str, Any]:
    summary = summarise_pack(pack_dir, registry)
    chosen_intent = (intent or summary.get("next_operation", "Make")).strip().lower()
    recommendation = recommend_execution(
        pack_dir,
        registry,
        intent=chosen_intent,
        repo_path=repo_path,
        home_path=home_path,
        runtime_target=runtime_target,
    )
    memory_context = recommendation["memory_context"]
    repo_context = recommendation["repo_context"]
    home_context = memory_context.get("home", {})
    recent_artifacts = memory_context.get("recent_artifacts", [])[: max(1, max_items)]
    resume_items = memory_context.get("resume_items", [])[: max(2, max_items + 1)]
    stale_signals = memory_context.get("stale_signals", [])[: max(1, max_items)]
    home_memory = home_context.get("long_term_memory", [])[: max(1, max_items)]
    unresolved_tasks = unresolved_task_items(summary, limit=max(1, max_items))
    repo_actions = repo_context.get("next_actions", [])[: max(1, max_items)]
    latest_run = latest_run_report_summary(pack_dir)
    latest_harvest = latest_harvest_report_summary(pack_dir)

    def shorten_path(path_text: Any) -> str:
        if not isinstance(path_text, str) or not path_text.strip():
            return ""
        path = Path(path_text)
        try:
            return str(path.resolve().relative_to(pack_dir.resolve()))
        except (OSError, ValueError):
            return path.name or path_text

    compact_recent_artifacts = [
        {
            "artifact_type": item.get("artifact_type", ""),
            "path": shorten_path(item.get("path", "")),
            "status": item.get("status", ""),
            "revision": item.get("revision", 0),
        }
        for item in recent_artifacts
    ]
    compact = {
        "schema_version": "0.1.0",
        "context_type": "JiniCompactContext",
        "generated_at": now_utc(),
        "pack_id": summary.get("pack_id", ""),
        "work_unit_id": summary["work_unit"].get("work_unit_id", ""),
        "intent": recommendation["intent"],
        "state": summary["work_unit"].get("current_state", ""),
        "health": summary.get("health", ""),
        "profile_id": summary["work_unit"].get("profile_id", ""),
        "execution_class": recommendation["execution_class"],
        "next_operation": summary.get("next_operation", ""),
        "recent_artifacts": compact_recent_artifacts,
        "resume_items": resume_items,
        "stale_signals": stale_signals,
        "home_memory": home_memory,
        "freshness": memory_context.get("freshness", [])[: max(1, max_items)],
        "unresolved_tasks": unresolved_tasks,
        "repo_root": repo_context.get("repo_root", ""),
        "repo_actions": repo_actions,
        "steering": memory_context.get("steering_paths", [])[: max(1, max_items)],
        "runtime_target": {
            "selected": recommendation["runtime_guidance"]["selected"]["id"],
            "fallbacks": recommendation["runtime_guidance"].get("fallbacks", [])[:3],
        },
        "latest_run": (
            {
                "generated_at": latest_run.get("generated_at", ""),
                "state_after": latest_run.get("state_after", ""),
                "health_after": latest_run.get("health_after", ""),
            }
            if latest_run is not None
            else None
        ),
        "latest_harvest": (
            {
                "generated_at": latest_harvest.get("generated_at", ""),
                "readiness": latest_harvest.get("readiness", ""),
                "evidence_status": latest_harvest.get("evidence_status", ""),
            }
            if latest_harvest is not None
            else None
        ),
    }

    def compact_chars(payload: dict[str, Any]) -> int:
        return len(json.dumps(payload, sort_keys=True))

    def shrink_compact(payload: dict[str, Any]) -> bool:
        for key in ("steering", "resume_items", "home_memory", "repo_actions", "stale_signals", "unresolved_tasks", "recent_artifacts", "freshness"):
            values = payload.get(key, [])
            if isinstance(values, list) and len(values) > 1:
                payload[key] = values[:-1]
                return True
        runtime_target = payload.get("runtime_target")
        if isinstance(runtime_target, dict):
            fallbacks = runtime_target.get("fallbacks", [])
            if isinstance(fallbacks, list) and len(fallbacks) > 1:
                runtime_target["fallbacks"] = fallbacks[:-1]
                return True
            if isinstance(fallbacks, list) and len(fallbacks) == 1:
                runtime_target["fallbacks"] = []
                return True
            if runtime_target.get("selected") and "fallbacks" in runtime_target:
                payload["runtime_target"] = {"selected": runtime_target["selected"]}
                return True
        for key in ("latest_harvest", "latest_run"):
            value = payload.get(key)
            if isinstance(value, dict) and len(value) > 2:
                preferred_keys = {
                    "latest_harvest": ["readiness", "evidence_status", "generated_at"],
                    "latest_run": ["state_after", "health_after", "generated_at"],
                }
                payload[key] = {item: value.get(item, "") for item in preferred_keys[key]}
                return True
        if isinstance(payload.get("repo_root"), str) and payload["repo_root"]:
            payload["repo_root"] = Path(payload["repo_root"]).name
            return True
        for key in ("resume_items", "home_memory", "repo_actions", "stale_signals"):
            values = payload.get(key, [])
            if isinstance(values, list) and values:
                last = str(values[-1])
                if len(last) > 96:
                    values[-1] = last[:93] + "..."
                    return True
        return False

    baseline_tokens = estimate_tokens(compact)
    original_chars = compact_chars(compact)
    if max_chars > 0:
        for _ in range(32):
            if compact_chars(compact) <= max_chars:
                break
            if not shrink_compact(compact):
                break
        if compact_chars(compact) > max_chars:
            compact["recent_artifacts"] = compact["recent_artifacts"][:1]
            compact["resume_items"] = compact["resume_items"][:1]
            compact["stale_signals"] = compact["stale_signals"][:1]
            compact["home_memory"] = compact["home_memory"][:1]
            compact["freshness"] = []
            compact["repo_actions"] = compact["repo_actions"][:1]
            compact["steering"] = compact.get("steering", [])[:1]
            compact["unresolved_tasks"] = [
                {
                    "task": item.get("task", ""),
                    "status": item.get("status", ""),
                }
                for item in compact.get("unresolved_tasks", [])[:1]
            ]
            for _ in range(16):
                if compact_chars(compact) <= max_chars:
                    break
                if not shrink_compact(compact):
                    break
        if compact_chars(compact) > max_chars:
            compact["recent_artifacts"] = [
                {
                    "artifact_type": item.get("artifact_type", ""),
                    "revision": item.get("revision", 0),
                }
                for item in compact.get("recent_artifacts", [])[:1]
            ]
            compact["unresolved_tasks"] = [
                {"task": item.get("task", "")}
                for item in compact.get("unresolved_tasks", [])[:1]
            ]
            compact["repo_actions"] = []
            compact["steering"] = []
            if compact.get("latest_run") is None:
                compact.pop("latest_run", None)
            if compact.get("latest_harvest") is None:
                compact.pop("latest_harvest", None)
            compact["runtime_target"] = {
                "selected": compact.get("runtime_target", {}).get("selected", "")
            }
            compact.pop("profile_id", None)
            compact.pop("next_operation", None)
        if compact_chars(compact) > max_chars:
            compact["recent_artifacts"] = []
            compact.pop("repo_root", None)
            compact["resume_items"] = [item[:64] + ("..." if len(item) > 64 else "") for item in compact["resume_items"][:1]]
            compact["home_memory"] = [item[:64] + ("..." if len(item) > 64 else "") for item in compact["home_memory"][:1]]
            compact["stale_signals"] = [item[:56] + ("..." if len(item) > 56 else "") for item in compact["stale_signals"][:1]]
            compact["steering"] = [item[:48] + ("..." if len(item) > 48 else "") for item in compact.get("steering", [])[:1]]
            compact["unresolved_tasks"] = [
                {"task": item.get("task", "")[:48] + ("..." if len(item.get("task", "")) > 48 else "")}
                for item in compact.get("unresolved_tasks", [])[:1]
            ]
        if compact_chars(compact) > max_chars:
            compact.pop("latest_run", None)
            compact.pop("latest_harvest", None)
        if compact_chars(compact) > max_chars:
            compact["home_memory"] = []
            compact["resume_items"] = compact["resume_items"][:1]
        if compact_chars(compact) > max_chars:
            for key in ("recent_artifacts", "home_memory", "freshness", "repo_actions", "stale_signals", "steering"):
                if compact.get(key) == []:
                    compact.pop(key, None)
    final_chars = compact_chars(compact)
    compact["token_budget"] = {
        "max_chars": max_chars,
        "estimated_tokens_before_trim": baseline_tokens,
        "estimated_tokens": estimate_tokens(compact),
        "estimated_chars": final_chars,
        "trimmed": final_chars < original_chars,
        "compression_ratio": round(final_chars / max(1, original_chars), 3),
    }
    return compact


def print_compact_context(compact: dict[str, Any]) -> None:
    print(f"PACK   {compact['pack_id']}")
    print(f"WORK   {compact['work_unit_id']}")
    print(f"STATE  {compact['state']}")
    print(f"HEALTH {compact['health']}")
    print(f"INTENT {compact['intent']}")
    print(f"CLASS  {compact['execution_class']}")
    token_budget = compact.get("token_budget", {})
    if token_budget:
        print(
            f"TOKENS est={token_budget.get('estimated_tokens', 0)} "
            f"chars={token_budget.get('estimated_chars', 0)} trimmed={token_budget.get('trimmed', False)}"
        )
    print("RESUME")
    for item in compact.get("resume_items", []):
        print(f"  - {item}")
    if compact.get("home_memory"):
        print("HOME")
        for item in compact["home_memory"]:
            print(f"  - {item}")
    if compact.get("unresolved_tasks"):
        print("TASKS")
        for task in compact["unresolved_tasks"]:
            owner_suffix = f" owner={task['owner']}" if task.get("owner") else ""
            print(f"  - [{task['status']}] {task['task']}{owner_suffix}")
    if compact.get("repo_actions"):
        print("REPO")
        for item in compact["repo_actions"]:
            print(f"  - {item}")
    if compact.get("steering"):
        print("STEERING")
        for item in compact["steering"]:
            print(f"  - {item}")
    if compact.get("stale_signals"):
        print("STALE")
        for item in compact["stale_signals"]:
            print(f"  - {item}")
    runtime_target = compact.get("runtime_target", {})
    if runtime_target.get("selected"):
        print(f"RUNTIME {runtime_target['selected']}")


def build_execution_checklist(
    pack_dir: Path,
    registry: dict[str, Any],
    *,
    intent: str | None = None,
    repo_path: Path | None = None,
    home_path: Path | None = None,
    runtime_target: str | None = None,
) -> dict[str, Any]:
    cli = cli_invocation()
    summary = summarise_pack(pack_dir, registry)
    recommendation = recommend_execution(
        pack_dir,
        registry,
        intent=intent,
        repo_path=repo_path,
        home_path=home_path,
        runtime_target=runtime_target,
    )
    memory_context = recommendation["memory_context"]
    repo_context = recommendation["repo_context"]
    pack_path = display_path(pack_dir)
    items: list[dict[str, Any]] = []

    items.append(
        {
            "phase": "state",
            "status": "ready" if not summary["blockers"] else "blocked",
            "kind": "state-anchor",
            "description": (
                f"Current state is `{summary['work_unit'].get('current_state', '')}` "
                f"with health `{summary.get('health', '')}` and next operation `{summary.get('next_operation', '')}`."
            ),
        }
    )

    for task in unresolved_task_items(summary):
        items.append(
            {
                "phase": "tasks",
                "status": "blocked" if recommendation["intent"] == "verify" else "recommended",
                "kind": "task",
                "description": f"Close unresolved task `{task['task']}` before promotion.",
                "owner": task.get("owner", ""),
            }
        )

    if repo_context.get("discovered"):
        for target in select_verification_targets(repo_context, ["test", "verify", "startup", "demo"], 4):
            detail = target.get("command") or target.get("path") or target.get("label")
            items.append(
                {
                    "phase": "repo",
                    "status": "recommended",
                    "kind": "repo-target",
                    "description": f"Use `{detail}` as a concrete verification surface.",
                    "command": target.get("command", ""),
                    "path": target.get("path", ""),
                    "category": target.get("category", ""),
                }
            )

    if recommendation["intent"] == "verify":
        items.append(
            {
                "phase": "evidence",
                "status": "recommended" if repo_context.get("discovered") else "blocked",
                "kind": "command",
                "description": "Refresh canonical evidence from local verification surfaces before approval or promotion.",
                "command": f"{cli} harvest-evidence {pack_path} --author <actor> --repo <repo>",
            }
        )

    home_context = memory_context.get("home", {})
    if home_context.get("bound"):
        items.append(
            {
                "phase": "memory",
                "status": "recommended",
                "kind": "home",
                "description": (
                    f"Use bound home `{home_context.get('home_root', '')}` to keep durable notes and routines aligned with this pack."
                ),
            }
        )

    runtime_guidance = recommendation.get("runtime_guidance", {})
    selected_runtime = runtime_guidance.get("selected", {})
    if selected_runtime.get("id"):
        runtime_id = selected_runtime["id"]
        items.append(
            {
                "phase": "adapter",
                "status": "recommended",
                "kind": "runtime-target",
                "description": f"Preferred runtime target for portable guidance is `{runtime_id}`.",
                "command": f"{cli} plan-install --bundle jini-core --target {runtime_id}",
            }
        )

    active_policy = recommendation.get("active_policy", {})
    if active_policy.get("policy_id"):
        items.append(
            {
                "phase": "policy",
                "status": "recommended",
                "kind": "active-rollout",
                "description": (
                    f"Active policy rollout `{active_policy.get('candidate_id', '')}` is in effect for "
                    f"`{active_policy.get('policy_id', '')}`."
                ),
            }
        )

    latest_harvest = latest_harvest_report_summary(pack_dir)
    if latest_harvest is not None:
        items.append(
            {
                "phase": "evidence",
                "status": "ready" if latest_harvest.get("readiness") == "ready" else "blocked",
                "kind": "harvest-report",
                "description": (
                    f"Latest harvest is `{latest_harvest.get('readiness')}` with evidence status "
                    f"`{latest_harvest.get('evidence_status')}`."
                ),
                "path": latest_harvest.get("path", ""),
            }
        )

    if memory_context.get("stale_signals"):
        for signal in memory_context["stale_signals"]:
            items.append(
                {
                    "phase": "memory",
                    "status": "blocked",
                    "kind": "stale-signal",
                    "description": signal,
                }
            )

    if "Approval" not in summary["latest_by_type"] and summary["work_unit"].get("current_state") == "awaiting_verification":
        items.append(
            {
                "phase": "promotion",
                "status": "recommended",
                "kind": "approval",
                "description": "Capture Approval after evidence is fresh and ready.",
                "command": f"{cli} capture-approval {pack_path} --author <actor> --approver-actor <approver> --scope operational-readiness",
            }
        )

    return {
        "schema_version": "0.1.0",
        "checklist_type": "JiniExecutionChecklist",
        "generated_at": now_utc(),
        "pack_id": summary.get("pack_id", ""),
        "work_unit_id": summary["work_unit"].get("work_unit_id", ""),
        "intent": recommendation["intent"],
        "execution_class": recommendation["execution_class"],
        "state": summary["work_unit"].get("current_state", ""),
        "repo_root": repo_context.get("repo_root", ""),
        "runtime_target": {
            "selected": recommendation["runtime_guidance"]["selected"]["id"],
            "fallbacks": recommendation["runtime_guidance"].get("fallbacks", [])[:3],
        },
        "active_policy": recommendation.get("active_policy", {}),
        "items": items,
    }


def print_execution_checklist(checklist: dict[str, Any]) -> None:
    print(f"PACK   {checklist['pack_id']}")
    print(f"WORK   {checklist['work_unit_id']}")
    print(f"STATE  {checklist['state']}")
    print(f"INTENT {checklist['intent']}")
    print(f"CLASS  {checklist.get('execution_class', '')}")
    if checklist.get("repo_root"):
        print(f"REPO   {checklist['repo_root']}")
    runtime_target = checklist.get("runtime_target", {})
    if runtime_target.get("selected"):
        print(f"RUNTIME {runtime_target['selected']}")
    print("CHECKLIST")
    for item in checklist.get("items", []):
        detail = item["description"]
        command = item.get("command", "")
        if command:
            detail += f" command={command}"
        print(f"  - [{item['status']}] {item['phase']} {item['kind']}: {detail}")


def build_adapter_summary(
    *,
    capability: str | None = None,
) -> dict[str, Any]:
    registry = load_adapter_registry()
    entries: list[dict[str, Any]] = []
    for adapter in registry["adapters"]:
        if not isinstance(adapter, dict):
            continue
        capabilities = adapter.get("capabilities", [])
        if capability and capability not in capabilities:
            continue
        entries.append(adapter)
    entries.sort(key=lambda item: (item.get("layer", ""), item.get("id", "")))
    return {
        "updated_at": registry.get("updated_at", ""),
        "schema_version": registry.get("schema_version", "0.1.0"),
        "capability_filter": capability,
        "adapters": entries,
    }


def print_adapter_summary(summary: dict[str, Any]) -> None:
    print(f"UPDATED {summary['updated_at']}")
    if summary.get("capability_filter"):
        print(f"FILTER  {summary['capability_filter']}")
    for adapter in summary.get("adapters", []):
        capabilities = ", ".join(adapter.get("capabilities", []))
        print(
            f"- {adapter.get('id', '')} | {adapter.get('layer', '')} | "
            f"{adapter.get('maturity', '')} | {capabilities}"
        )


def build_adapter_conformance_summary() -> dict[str, Any]:
    registry = load_adapter_registry()
    install_manifest = load_install_manifest()
    manifest_targets = {
        target["id"]: target
        for target in install_manifest.get("targets", [])
        if isinstance(target, dict) and target.get("id")
    }
    bundle_targets: set[str] = set()
    for bundle in install_manifest.get("bundles", []):
        if isinstance(bundle, dict):
            bundle_targets.update(bundle.get("compatible_targets", []))

    checks: list[dict[str, Any]] = []
    for adapter in registry["adapters"]:
        if not isinstance(adapter, dict):
            continue
        adapter_id = str(adapter.get("id", "")).strip()
        layer = str(adapter.get("layer", "")).strip()
        capabilities = list(adapter.get("capabilities", []))
        passed = True
        notes: list[str] = []

        if layer == "runtime-target":
            shim_path = ROOT / "distribution" / "targets" / adapter_id / "README.md"
            if not shim_path.exists():
                passed = False
                notes.append(f"Missing shim metadata at {display_path(shim_path)}")
            else:
                shim_text = shim_path.read_text(encoding="utf-8")
                if "activate-runtime-target" not in shim_text:
                    passed = False
                    notes.append("Runtime target shim does not document activate-runtime-target")
            if adapter_id not in manifest_targets:
                passed = False
                notes.append("Target is not present in install-manifest targets")
            if adapter_id not in bundle_targets:
                passed = False
                notes.append("No install bundle currently declares compatibility with this target")
            if "runtime-activate" in capabilities and "runtime-handoff" not in capabilities:
                passed = False
                notes.append("Runtime activation requires runtime-handoff capability")
        elif layer == "issue-system":
            if adapter_id not in {"jira", "github"}:
                passed = False
                notes.append("Issue-system adapter is not wired into export surfaces")
            if "issues-export" not in capabilities:
                passed = False
                notes.append("Issue-system adapter is missing issues-export capability")
            if adapter_id == "jira" and "issues-publish-plan" not in capabilities:
                passed = False
                notes.append("Jira adapter should expose issues-publish-plan")
            if adapter_id == "jira" and "issues-execute-bridge" not in capabilities:
                passed = False
                notes.append("Jira adapter should expose issues-execute-bridge")
            if adapter_id == "github":
                if "issues-publish-plan" not in capabilities:
                    passed = False
                    notes.append("GitHub adapter should expose issues-publish-plan")
                if "issues-apply-local" not in capabilities:
                    passed = False
                    notes.append("GitHub adapter should expose issues-apply-local")
        elif layer == "wiki-system":
            if adapter_id not in {"confluence", "markdown"}:
                passed = False
                notes.append("Wiki-system adapter is not wired into export surfaces")
            if "wiki-export" not in capabilities:
                passed = False
                notes.append("Wiki-system adapter is missing wiki-export capability")
            if adapter_id == "confluence" and "wiki-execute-bridge" not in capabilities:
                passed = False
                notes.append("Confluence adapter should expose wiki-execute-bridge")
            if adapter_id == "markdown":
                if "wiki-publish-plan" not in capabilities:
                    passed = False
                    notes.append("Markdown adapter should expose wiki-publish-plan")
                if "wiki-apply-local" not in capabilities:
                    passed = False
                    notes.append("Markdown adapter should expose wiki-apply-local")

        checks.append(
            {
                "id": adapter_id,
                "layer": layer,
                "maturity": adapter.get("maturity", ""),
                "status": "ok" if passed else "missing",
                "capabilities": capabilities,
                "notes": notes,
            }
        )

    overall = "ok" if all(check["status"] == "ok" for check in checks) else "missing"
    return {
        "updated_at": registry.get("updated_at", ""),
        "schema_version": registry.get("schema_version", "0.1.0"),
        "status": overall,
        "checks": checks,
    }


def print_adapter_conformance(summary: dict[str, Any]) -> None:
    print(f"UPDATED {summary['updated_at']}")
    print(f"STATUS  {summary['status']}")
    for check in summary.get("checks", []):
        capabilities = ", ".join(check.get("capabilities", []))
        print(f"- [{check['status']}] {check['id']} | {check['layer']} | {capabilities}")
        for note in check.get("notes", []):
            print(f"    note: {note}")


def build_learning_snapshot(
    *,
    path: Path | None = None,
    limit: int = 200,
) -> dict[str, Any]:
    source_path = path if path is not None else LEARNING_EVENTS_PATH
    events = read_learning_events(path=source_path, limit=limit)
    by_type = Counter(str(event.get("event_type", "")) for event in events)
    by_pack = Counter(str(event.get("pack_id", "")) for event in events if event.get("pack_id"))
    execution_classes = Counter(str(event.get("execution_class", "")) for event in events if event.get("execution_class"))
    harvest_readiness = Counter(str(event.get("readiness", "")) for event in events if event.get("event_type") == "harvest-evidence")
    runtime_targets = Counter(str(event.get("runtime_target", "")) for event in events if event.get("runtime_target"))
    compression_ratios: list[float] = []
    memory_write_count = 0
    home_bound_count = 0
    state_transitions: list[str] = []
    for event in events:
        if event.get("event_type") == "run-pack":
            before = event.get("state_before", "")
            after = event.get("state_after", "")
            state_transitions.append(f"{before}->{after}")
        ratio = event.get("compression_ratio")
        if isinstance(ratio, (int, float)) and ratio > 0:
            compression_ratios.append(float(ratio))
        if event.get("memory_appended"):
            memory_write_count += 1
        if event.get("home_bound"):
            home_bound_count += 1
    return {
        "path": display_path(source_path),
        "limit": limit,
        "event_count": len(events),
        "event_types": dict(sorted(by_type.items())),
        "packs": dict(sorted(by_pack.items())),
        "execution_classes": dict(sorted(execution_classes.items())),
        "harvest_readiness": dict(sorted(harvest_readiness.items())),
        "runtime_targets": dict(sorted(runtime_targets.items())),
        "memory_write_count": memory_write_count,
        "home_bound_count": home_bound_count,
        "average_compaction_ratio": round(sum(compression_ratios) / len(compression_ratios), 3) if compression_ratios else 0.0,
        "recent_state_transitions": state_transitions[-10:],
        "events": events,
    }


def print_learning_snapshot(snapshot: dict[str, Any]) -> None:
    print(f"PATH   {snapshot['path']}")
    print(f"COUNT  {snapshot['event_count']}")
    if snapshot.get("event_types"):
        print("TYPES")
        for key, value in snapshot["event_types"].items():
            print(f"  - {key}: {value}")
    if snapshot.get("execution_classes"):
        print("CLASS")
        for key, value in snapshot["execution_classes"].items():
            print(f"  - {key}: {value}")
    if snapshot.get("harvest_readiness"):
        print("HARVEST")
        for key, value in snapshot["harvest_readiness"].items():
            print(f"  - {key}: {value}")
    if snapshot.get("runtime_targets"):
        print("RUNTIME")
        for key, value in snapshot["runtime_targets"].items():
            print(f"  - {key}: {value}")
    print(f"MEMORY {snapshot.get('memory_write_count', 0)}")
    print(f"RATIO  {snapshot.get('average_compaction_ratio', 0.0)}")


def print_learning_events(events: list[dict[str, Any]]) -> None:
    if not events:
        print("No learning events recorded")
        return
    for event in events:
        work_unit = event.get("work_unit_id", "")
        print(
            f"- {event.get('recorded_at', '')} | {event.get('event_type', '')} "
            f"| work={work_unit or '-'}"
        )


def adapter_priority(adapter: dict[str, Any]) -> int:
    explicit = adapter.get("priority")
    if isinstance(explicit, int):
        return explicit
    maturity = str(adapter.get("maturity", ""))
    defaults = {
        "staged-publish": 90,
        "local-activation": 80,
        "local-apply": 75,
        "local-fallback": 70,
        "install-shim": 60,
        "export-only": 50,
    }
    return defaults.get(maturity, 10)


def build_adapter_resolution(
    *,
    capability: str,
    layer: str | None = None,
    preferred: str | None = None,
) -> dict[str, Any]:
    if not capability.strip():
        raise ValueError("resolve-adapter requires a non-empty capability")

    registry = load_adapter_registry()
    matches: list[dict[str, Any]] = []
    for adapter in registry["adapters"]:
        if not isinstance(adapter, dict):
            continue
        if layer and adapter.get("layer") != layer:
            continue
        capabilities = adapter.get("capabilities", [])
        if capability not in capabilities:
            continue
        entry = deepcopy(adapter)
        entry["priority"] = adapter_priority(adapter)
        matches.append(entry)

    if not matches:
        raise ValueError(f"No adapters expose capability {capability!r}")

    matches.sort(key=lambda item: (-int(item.get("priority", 0)), item.get("id", "")))
    selected = matches[0]
    notes: list[str] = []
    if preferred:
        preferred_match = next((item for item in matches if item.get("id") == preferred), None)
        if preferred_match is not None:
            selected = preferred_match
            notes.append(f"Preferred adapter `{preferred}` was selected explicitly.")
        else:
            notes.append(f"Preferred adapter `{preferred}` does not expose capability `{capability}`.")

    fallbacks = [item["id"] for item in matches if item["id"] != selected["id"]]
    notes.append(
        f"Selected `{selected['id']}` because it is the highest-priority adapter for `{capability}`."
    )
    return {
        "schema_version": registry.get("schema_version", "0.1.0"),
        "updated_at": registry.get("updated_at", ""),
        "capability": capability,
        "layer": layer or "",
        "selected": selected,
        "fallbacks": fallbacks,
        "matches": matches,
        "notes": notes,
    }


def print_adapter_resolution(resolution: dict[str, Any]) -> None:
    selected = resolution["selected"]
    print(f"UPDATED   {resolution['updated_at']}")
    print(f"CAPABILITY {resolution['capability']}")
    if resolution.get("layer"):
        print(f"LAYER     {resolution['layer']}")
    print(
        f"SELECTED  {selected['id']} | {selected.get('layer', '')} | "
        f"{selected.get('maturity', '')} | priority={selected.get('priority', 0)}"
    )
    if resolution.get("fallbacks"):
        print(f"FALLBACKS {', '.join(resolution['fallbacks'])}")
    for note in resolution.get("notes", []):
        print(f"NOTE      {note}")


def build_adapter_matrix() -> dict[str, Any]:
    registry = load_adapter_registry()
    capabilities: dict[str, list[str]] = {}
    layers: dict[str, list[str]] = {}
    for adapter in registry["adapters"]:
        if not isinstance(adapter, dict):
            continue
        adapter_id = str(adapter.get("id", ""))
        layer = str(adapter.get("layer", ""))
        layers.setdefault(layer, []).append(adapter_id)
        for capability in adapter.get("capabilities", []):
            capabilities.setdefault(str(capability), []).append(adapter_id)
    return {
        "schema_version": registry.get("schema_version", "0.1.0"),
        "updated_at": registry.get("updated_at", ""),
        "layers": {key: sorted(value) for key, value in sorted(layers.items())},
        "capabilities": {key: sorted(value) for key, value in sorted(capabilities.items())},
    }


def print_adapter_matrix(matrix: dict[str, Any]) -> None:
    print(f"UPDATED {matrix['updated_at']}")
    print("LAYERS")
    for layer, adapters in matrix.get("layers", {}).items():
        print(f"  - {layer}: {', '.join(adapters)}")
    print("CAPABILITIES")
    for capability, adapters in matrix.get("capabilities", {}).items():
        print(f"  - {capability}: {', '.join(adapters)}")


def build_routing_backtest(
    *,
    path: Path | None = None,
    limit: int = 200,
    min_samples: int = 1,
) -> dict[str, Any]:
    events = read_learning_events(path=path, limit=limit)
    by_bucket: dict[tuple[str, str], dict[str, Any]] = {}
    recommendations: list[dict[str, Any]] = []

    for event in events:
        event_type = str(event.get("event_type", ""))
        intent = str(event.get("intent", "")).strip()
        execution_class = str(event.get("execution_class", "")).strip()
        if not intent or not execution_class:
            continue
        key = (intent, execution_class)
        bucket = by_bucket.setdefault(
            key,
            {
                "intent": intent,
                "execution_class": execution_class,
                "samples": 0,
                "successes": 0,
                "failures": 0,
                "bounded": 0,
                "event_types": Counter(),
            },
        )
        bucket["samples"] += 1
        bucket["event_types"][event_type] += 1
        success = False
        bounded = False
        if event_type == "run-pack":
            success = int(event.get("blocker_count", 0) or 0) == 0 or event.get("state_after") != event.get("state_before")
        elif event_type == "harvest-evidence":
            readiness = str(event.get("readiness", ""))
            evidence_status = str(event.get("evidence_status", ""))
            success = readiness == "ready" and evidence_status in READY_ARTIFACT_STATUSES
            bounded = readiness != "ready"
        elif event_type in {"compact-context", "execution-checklist"}:
            success = int(event.get("stale_signal_count", 0) or 0) == 0
        if success:
            bucket["successes"] += 1
        else:
            bucket["failures"] += 1
        if bounded:
            bucket["bounded"] += 1

    entries: list[dict[str, Any]] = []
    intent_groups: dict[str, list[dict[str, Any]]] = {}
    for bucket in by_bucket.values():
        samples = int(bucket["samples"])
        if samples < max(1, min_samples):
            continue
        success_rate = bucket["successes"] / samples if samples else 0.0
        entry = {
            "intent": bucket["intent"],
            "execution_class": bucket["execution_class"],
            "samples": samples,
            "successes": int(bucket["successes"]),
            "failures": int(bucket["failures"]),
            "bounded": int(bucket["bounded"]),
            "success_rate": round(success_rate, 3),
            "event_types": dict(sorted(bucket["event_types"].items())),
        }
        entries.append(entry)
        intent_groups.setdefault(bucket["intent"], []).append(entry)

    class_rank = {"cheap": 0, "standard": 1, "deep": 2}
    for intent, group in sorted(intent_groups.items()):
        ordered = sorted(
            group,
            key=lambda item: (-float(item["success_rate"]), -int(item["samples"]), class_rank.get(str(item["execution_class"]), 99)),
        )
        best = ordered[0]
        rationale = (
            f"`{best['execution_class']}` observed success rate {best['success_rate']:.3f} "
            f"across {best['samples']} sample(s) for intent `{intent}`."
        )
        recommendations.append(
            {
                "intent": intent,
                "recommended_execution_class": best["execution_class"],
                "samples": best["samples"],
                "success_rate": best["success_rate"],
                "rationale": rationale,
            }
        )

    entries.sort(key=lambda item: (item["intent"], class_rank.get(str(item["execution_class"]), 99)))
    return {
        "path": display_path(path if path is not None else LEARNING_EVENTS_PATH),
        "limit": limit,
        "min_samples": min_samples,
        "event_count": len(events),
        "buckets": entries,
        "policy_recommendations": recommendations,
    }


def print_routing_backtest(backtest: dict[str, Any]) -> None:
    print(f"PATH   {backtest['path']}")
    print(f"COUNT  {backtest['event_count']}")
    print("BUCKETS")
    for bucket in backtest.get("buckets", []):
        print(
            f"  - {bucket['intent']} | {bucket['execution_class']} | "
            f"samples={bucket['samples']} success={bucket['success_rate']:.3f} bounded={bucket['bounded']}"
        )
    if backtest.get("policy_recommendations"):
        print("POLICY")
        for item in backtest["policy_recommendations"]:
            print(
                f"  - {item['intent']}: {item['recommended_execution_class']} "
                f"({item['success_rate']:.3f} over {item['samples']} sample(s))"
            )


def next_runtime_handoff_path(pack_dir: Path, runtime_target_id: str) -> Path:
    handoff_dir = pack_dir / "runtime" / "handoffs"
    handoff_dir.mkdir(parents=True, exist_ok=True)
    stamp = datetime.now(timezone.utc).strftime("%Y%m%dT%H%M%S%fZ")
    slug = re.sub(r"[^a-z0-9-]+", "-", runtime_target_id.strip().lower()).strip("-") or "runtime"
    return handoff_dir / f"runtime-handoff-{slug}-{stamp}.json"


def latest_runtime_handoff_path(pack_dir: Path, runtime_target_id: str | None = None) -> Path | None:
    handoff_dir = pack_dir / "runtime" / "handoffs"
    if not handoff_dir.exists():
        return None
    candidates = sorted(handoff_dir.glob("runtime-handoff-*.json"))
    if runtime_target_id:
        slug = re.sub(r"[^a-z0-9-]+", "-", runtime_target_id.strip().lower()).strip("-") or "runtime"
        candidates = [path for path in candidates if f"runtime-handoff-{slug}-" in path.name]
    if not candidates:
        return None
    return candidates[-1]


def next_runtime_activation_receipt_path(pack_dir: Path, runtime_target_id: str) -> Path:
    activation_dir = pack_dir / "runtime" / "activations"
    activation_dir.mkdir(parents=True, exist_ok=True)
    stamp = datetime.now(timezone.utc).strftime("%Y%m%dT%H%M%S%fZ")
    slug = re.sub(r"[^a-z0-9-]+", "-", runtime_target_id.strip().lower()).strip("-") or "runtime"
    return activation_dir / f"runtime-activation-{slug}-{stamp}.json"


def runtime_activation_root(install_result: dict[str, Any], target_id: str, pack_id: str, work_unit_id: str) -> Path:
    install_record = next(
        (item for item in install_result.get("installs", []) if item.get("target_id") == target_id),
        None,
    )
    if not isinstance(install_record, dict):
        raise ValueError(f"Install result does not include target {target_id!r}")
    shim_destination = str(install_record.get("shim_destination", "")).strip()
    universal_destination = str(install_record.get("universal_destination", "")).strip()
    root = Path(shim_destination or universal_destination)
    slug_source = pack_id or work_unit_id or "work-unit"
    return root / "runtime-handoffs" / slugify(slug_source)


def render_runtime_activation_markdown(handoff: dict[str, Any]) -> str:
    runtime_target = handoff.get("runtime_target", {}).get("selected", {})
    repo_context = handoff.get("repo_context", {})
    compact_context = handoff.get("compact_context", {})
    execution_checklist = handoff.get("execution_checklist", {})
    repo_map = handoff.get("repo_map", {}) if isinstance(handoff.get("repo_map", {}), dict) else {}
    active_policy = handoff.get("active_policy", {}) if isinstance(handoff.get("active_policy", {}), dict) else {}
    lines = [
        "# Jini Runtime Activation",
        "",
        f"- Pack: `{handoff.get('pack_id', '')}`",
        f"- WorkUnit: `{handoff.get('work_unit_id', '')}`",
        f"- State: `{handoff.get('state', '')}`",
        f"- Intent: `{handoff.get('intent', '')}`",
        f"- Execution Class: `{handoff.get('execution_class', '')}`",
        f"- Runtime Target: `{runtime_target.get('id', '')}`",
        "",
        "## Resume",
    ]
    for item in compact_context.get("resume_items", [])[:5]:
        lines.append(f"- {item}")
    if compact_context.get("home_memory"):
        lines.extend(["", "## Home Memory"])
        for item in compact_context.get("home_memory", [])[:4]:
            lines.append(f"- {item}")
    if repo_context.get("repo_root"):
        lines.extend(["", "## Repo Context", f"- Root: `{repo_context.get('repo_root', '')}`"])
        for item in repo_context.get("next_actions", [])[:4]:
            lines.append(f"- {item}")
    steering = repo_map.get("steering", {})
    if isinstance(steering, dict) and steering.get("active_paths"):
        lines.extend(["", "## Steering"])
        for path in steering.get("active_paths", [])[:4]:
            lines.append(f"- `{path}`")
    if active_policy.get("policy_id"):
        lines.extend(
            [
                "",
                "## Active Policy",
                f"- Policy: `{active_policy.get('policy_id', '')}`",
                f"- Candidate: `{active_policy.get('candidate_id', '')}`",
            ]
        )
        if active_policy.get("intent_overrides"):
            lines.append(f"- Overrides: `{json.dumps(active_policy.get('intent_overrides', {}), sort_keys=True)}`")
    lines.extend(["", "## Checklist"])
    for item in execution_checklist.get("items", [])[:6]:
        lines.append(f"- [{item.get('status', '')}] {item.get('description', '')}")
    lines.extend(["", "## Guardrails"])
    for key, value in handoff.get("guardrails", {}).items():
        lines.append(f"- `{key}`: `{value}`")
    return "\n".join(lines).rstrip() + "\n"


def activate_runtime_target(
    pack_dir: Path,
    registry: dict[str, Any],
    *,
    handoff_path: Path | None = None,
    intent: str | None = None,
    repo_path: Path | None = None,
    home_path: Path | None = None,
    runtime_target: str | None = None,
    prefix: Path | None = None,
    max_items: int = 5,
    max_chars: int = 1200,
) -> tuple[dict[str, Any], Path]:
    selected_handoff_path = handoff_path
    if selected_handoff_path is None:
        selected_handoff_path = latest_runtime_handoff_path(pack_dir, runtime_target_id=runtime_target)
    if selected_handoff_path is not None and selected_handoff_path.exists():
        handoff = load_json_file(selected_handoff_path)
    else:
        handoff, selected_handoff_path = build_runtime_handoff(
            pack_dir,
            registry,
            intent=intent,
            repo_path=repo_path,
            home_path=home_path,
            runtime_target=runtime_target,
            max_items=max_items,
            max_chars=max_chars,
        )

    if not isinstance(handoff, dict):
        raise ValueError("Runtime handoff must be a mapping")
    runtime_doc = handoff.get("runtime_target", {}).get("selected", {})
    target_id = str(runtime_doc.get("id", "")).strip()
    if not target_id:
        raise ValueError("Runtime handoff is missing a selected runtime target")

    install_result = install_bundles(
        bundle_ids=["jini-core"],
        target_ids=[target_id],
        prefix=prefix,
    )
    activation_root = runtime_activation_root(
        install_result,
        target_id,
        str(handoff.get("pack_id", "")),
        str(handoff.get("work_unit_id", "")),
    )
    if activation_root.exists():
        remove_installed_path(activation_root)
    activation_root.mkdir(parents=True, exist_ok=True)

    handoff_copy_path = activation_root / "handoff.json"
    handoff_copy_path.write_text(json.dumps(handoff, indent=2) + "\n", encoding="utf-8")
    compact_context_path = activation_root / "compact-context.json"
    compact_context_path.write_text(json.dumps(handoff.get("compact_context", {}), indent=2) + "\n", encoding="utf-8")
    execution_checklist_path = activation_root / "execution-checklist.json"
    execution_checklist_path.write_text(json.dumps(handoff.get("execution_checklist", {}), indent=2) + "\n", encoding="utf-8")
    if isinstance(handoff.get("repo_map"), dict):
        repo_map_path = activation_root / "repo-map.json"
        repo_map_path.write_text(json.dumps(handoff.get("repo_map", {}), indent=2) + "\n", encoding="utf-8")
    else:
        repo_map_path = None
    activation_markdown_path = activation_root / "Jini-RUNTIME.md"
    activation_markdown_path.write_text(render_runtime_activation_markdown(handoff), encoding="utf-8")

    home_binding = resolve_home_binding(pack_dir, explicit_home=home_path)
    home_observation = append_home_observation(
        pack_dir,
        home_binding=home_binding,
        line=(
            f"Activated runtime target {target_id} for {handoff.get('pack_id', '')} "
            f"at state {handoff.get('state', '')} with intent {handoff.get('intent', '')}."
        ),
    )

    receipt = {
        "schema_version": "0.1.0",
        "activation_type": "JiniRuntimeActivation",
        "generated_at": now_utc(),
        "pack_id": handoff.get("pack_id", ""),
        "work_unit_id": handoff.get("work_unit_id", ""),
        "runtime_target": target_id,
        "state": handoff.get("state", ""),
        "intent": handoff.get("intent", ""),
        "execution_class": handoff.get("execution_class", ""),
        "source_handoff_path": display_path(selected_handoff_path),
        "activation_root": str(activation_root),
        "activation_files": [
            str(handoff_copy_path),
            str(compact_context_path),
            str(execution_checklist_path),
            *( [str(repo_map_path)] if repo_map_path is not None else [] ),
            str(activation_markdown_path),
        ],
        "install_receipt_path": install_result.get("receipt_path", ""),
        "install_status": install_result.get("status", ""),
        "home_observation": home_observation,
        "guardrails": handoff.get("guardrails", {}),
    }
    receipt_path = next_runtime_activation_receipt_path(pack_dir, target_id)
    receipt_path.write_text(json.dumps(receipt, indent=2) + "\n", encoding="utf-8")
    append_learning_event(
        "activate-runtime-target",
        {
            "pack_id": receipt["pack_id"],
            "work_unit_id": receipt["work_unit_id"],
            "runtime_target": target_id,
            "intent": receipt["intent"],
            "execution_class": receipt["execution_class"],
            "activation_root": str(activation_root),
            "activation_file_count": len(receipt["activation_files"]),
            "memory_appended": bool(home_observation.get("appended")),
            "home_bound": bool(home_binding.get("bound")),
        },
        pack_dir=pack_dir,
    )
    return receipt, receipt_path


def print_runtime_activation(activation: dict[str, Any]) -> None:
    print(f"PACK   {activation['pack_id']}")
    print(f"WORK   {activation['work_unit_id']}")
    print(f"TARGET {activation['runtime_target']}")
    print(f"STATE  {activation['state']}")
    print(f"INTENT {activation['intent']}")
    print(f"CLASS  {activation['execution_class']}")
    print(f"ROOT   {activation['activation_root']}")
    print(f"HANDOFF {activation['source_handoff_path']}")
    print(f"INSTALL {activation.get('install_receipt_path', '')}")
    for path in activation.get("activation_files", []):
        print(f"  - {path}")


def next_policy_review_path(pack_dir: Path) -> Path:
    review_dir = pack_dir / "runtime" / "policy-reviews"
    review_dir.mkdir(parents=True, exist_ok=True)
    stamp = datetime.now(timezone.utc).strftime("%Y%m%dT%H%M%S%fZ")
    return review_dir / f"policy-review-{stamp}.json"


def build_runtime_handoff(
    pack_dir: Path,
    registry: dict[str, Any],
    *,
    intent: str | None = None,
    repo_path: Path | None = None,
    home_path: Path | None = None,
    runtime_target: str | None = None,
    max_items: int = 5,
    max_chars: int = 1200,
    summary: dict[str, Any] | None = None,
    recommendation: dict[str, Any] | None = None,
    compact: dict[str, Any] | None = None,
    checklist: dict[str, Any] | None = None,
    repo_map: dict[str, Any] | None = None,
    home_binding: dict[str, Any] | None = None,
) -> tuple[dict[str, Any], Path]:
    summary = summary or summarise_pack(pack_dir, registry)
    recommendation = recommendation or recommend_execution(
        pack_dir,
        registry,
        intent=intent,
        repo_path=repo_path,
        home_path=home_path,
        runtime_target=runtime_target,
    )
    compact = compact or build_compact_context(
        pack_dir,
        registry,
        intent=recommendation["intent"],
        repo_path=repo_path,
        home_path=home_path,
        runtime_target=runtime_target,
        max_items=max_items,
        max_chars=max_chars,
    )
    checklist = checklist or build_execution_checklist(
        pack_dir,
        registry,
        intent=recommendation["intent"],
        repo_path=repo_path,
        home_path=home_path,
        runtime_target=runtime_target,
    )
    runtime_guidance = recommendation["runtime_guidance"]
    selected_runtime = runtime_guidance["selected"]
    install_preview = plan_install(
        bundle_ids=["jini-core"],
        target_ids=[selected_runtime["id"]],
    )
    install_preview.setdefault("targets", deepcopy(install_preview.get("selected_targets", [])))
    install_preview.setdefault("bundles", deepcopy(install_preview.get("selected_bundles", [])))
    repo_map = repo_map if repo_map is not None else (build_repo_map(pack_dir, repo_path=repo_path) if repo_path is not None else None)
    home_binding = home_binding or resolve_home_binding(pack_dir, explicit_home=home_path)
    handoff_path = next_runtime_handoff_path(pack_dir, str(selected_runtime.get("id", "")))
    repo_context = recommendation["repo_context"]
    checklist_items = checklist.get("items", [])
    compact_budget = compact.get("token_budget", {})

    handoff_steps = [
        (
            f"Load compact context first; target `{selected_runtime['id']}` can resume in about "
            f"{compact_budget.get('estimated_tokens', 0)} token(s)."
        ),
        (
            f"Prefer runtime target `{selected_runtime['id']}` and fall back to "
            f"{', '.join(runtime_guidance.get('fallbacks', [])[:3]) or 'no fallback'}."
        ),
    ]
    for item in checklist_items[:4]:
        handoff_steps.append(item["description"])

    handoff = {
        "schema_version": "0.1.0",
        "handoff_type": "JiniRuntimeHandoff",
        "generated_at": now_utc(),
        "pack_id": summary.get("pack_id", ""),
        "work_unit_id": summary["work_unit"].get("work_unit_id", ""),
        "intent": recommendation["intent"],
        "state": summary["work_unit"].get("current_state", ""),
        "execution_class": recommendation["execution_class"],
        "runtime_target": runtime_guidance,
        "handoff_path": display_path(handoff_path),
        "repo_context": {
            "discovered": bool(repo_context.get("discovered")),
            "repo_root": repo_context.get("repo_root", ""),
            "notes": repo_context.get("notes", []),
            "next_actions": repo_context.get("next_actions", []),
            "verification_targets": repo_context.get("verification_targets", [])[:4],
        },
        "repo_map": repo_map,
        "home_binding": {
            "bound": bool(home_binding.get("bound")),
            "home_root": str(home_binding.get("home_root", "")) if home_binding.get("home_root") else "",
            "source": home_binding.get("source", ""),
            "binding_path": home_binding.get("binding_path", ""),
        },
        "compact_context": compact,
        "execution_checklist": checklist,
        "install_plan": install_preview,
        "active_policy": recommendation.get("active_policy", {}),
        "handoff_steps": handoff_steps,
        "guardrails": {
            "writes_require_consent": True,
            "publish_requires_consent": True,
            "semantic_state_changes_stay_in_jini": True,
        },
    }
    handoff_path.write_text(json.dumps(handoff, indent=2) + "\n", encoding="utf-8")
    append_learning_event(
        "stage-runtime-handoff",
        {
            "pack_id": handoff["pack_id"],
            "work_unit_id": handoff["work_unit_id"],
            "intent": handoff["intent"],
            "execution_class": handoff["execution_class"],
            "runtime_target": selected_runtime["id"],
            "fallback_count": len(runtime_guidance.get("fallbacks", [])),
            "checklist_item_count": len(checklist_items),
            "stale_signal_count": len(compact.get("stale_signals", [])),
            "estimated_tokens": compact_budget.get("estimated_tokens", 0),
            "compression_ratio": compact_budget.get("compression_ratio", 0.0),
            "home_bound": bool(home_binding.get("bound")),
        },
        pack_dir=pack_dir,
    )
    return handoff, handoff_path


def print_runtime_handoff(handoff: dict[str, Any]) -> None:
    runtime_target = handoff.get("runtime_target", {}).get("selected", {})
    print(f"PACK   {handoff['pack_id']}")
    print(f"WORK   {handoff['work_unit_id']}")
    print(f"STATE  {handoff['state']}")
    print(f"INTENT {handoff['intent']}")
    print(f"CLASS  {handoff['execution_class']}")
    if runtime_target.get("id"):
        print(f"TARGET {runtime_target['id']}")
    if handoff.get("repo_context", {}).get("repo_root"):
        print(f"REPO   {handoff['repo_context']['repo_root']}")
    if handoff.get("home_binding", {}).get("bound"):
        print(f"HOME   {handoff['home_binding']['home_root']}")
    print(f"HANDOFF {handoff['handoff_path']}")
    compact_budget = handoff.get("compact_context", {}).get("token_budget", {})
    if compact_budget:
        print(
            f"TOKENS est={compact_budget.get('estimated_tokens', 0)} "
            f"chars={compact_budget.get('estimated_chars', 0)}"
        )
    print("STEPS")
    for item in handoff.get("handoff_steps", []):
        print(f"  - {item}")


def build_policy_review(
    *,
    pack_dir: Path | None = None,
    limit: int = 200,
    min_samples: int = 1,
) -> tuple[dict[str, Any], Path | None]:
    source_path = runtime_events_path(pack_dir) if pack_dir is not None else LEARNING_EVENTS_PATH
    snapshot = build_learning_snapshot(path=source_path, limit=limit)
    backtest = build_routing_backtest(path=source_path, limit=limit, min_samples=min_samples)
    snapshot_summary = deepcopy(snapshot)
    snapshot_summary.pop("events", None)

    event_types = snapshot_summary.get("event_types", {})
    harvest_readiness = snapshot_summary.get("harvest_readiness", {})
    runtime_targets = snapshot_summary.get("runtime_targets", {})
    coverage = {
        "compact-context": int(event_types.get("compact-context", 0)),
        "execution-checklist": int(event_types.get("execution-checklist", 0)),
        "stage-runtime-handoff": int(event_types.get("stage-runtime-handoff", 0)),
        "run-pack": int(event_types.get("run-pack", 0)),
        "harvest-evidence": int(event_types.get("harvest-evidence", 0)),
    }
    coverage_gaps = [key for key, value in coverage.items() if value == 0]

    policy_candidates: list[dict[str, Any]] = []
    for recommendation in backtest.get("policy_recommendations", []):
        policy_candidates.append(
            {
                "kind": "routing-default",
                "priority": 1,
                "intent": recommendation["intent"],
                "proposed_execution_class": recommendation["recommended_execution_class"],
                "samples": recommendation["samples"],
                "confidence": recommendation["success_rate"],
                "rationale": recommendation["rationale"],
            }
        )

    bounded_harvests = int(harvest_readiness.get("bounded", 0))
    ready_harvests = int(harvest_readiness.get("ready", 0))
    promotion_rationale = (
        f"{bounded_harvests} bounded harvest(s) are present; promotion should stay blocked until ready evidence is observed."
    )
    if coverage["harvest-evidence"] == 0:
        promotion_rationale = (
            "No harvest-evidence events were observed; promotion gating should stay strict until proof exists."
        )
    elif bounded_harvests == 0:
        promotion_rationale = (
            f"{ready_harvests} ready harvest(s) were observed; keep fresh ready harvest evidence as a hard promotion gate."
        )
    policy_candidates.append(
        {
            "kind": "promotion-gate",
            "priority": 1,
            "proposed_rule": "Require fresh ready harvest evidence before approval or operational promotion.",
            "readiness": {
                "ready": ready_harvests,
                "bounded": bounded_harvests,
            },
            "rationale": promotion_rationale,
        }
    )

    average_compaction_ratio = float(snapshot_summary.get("average_compaction_ratio", 0.0) or 0.0)
    if average_compaction_ratio > 0:
        policy_candidates.append(
            {
                "kind": "compact-reload",
                "priority": 2,
                "proposed_rule": "Prefer compact-context for routine resumptions before broad reloads.",
                "confidence": round(max(0.0, 1.0 - average_compaction_ratio), 3),
                "rationale": (
                    f"Observed average compaction ratio is {average_compaction_ratio:.3f}, which indicates "
                    "compact resume slices are preserving context while reducing transport cost."
                ),
            }
        )

    if int(snapshot_summary.get("memory_write_count", 0)) > 0 or int(snapshot_summary.get("home_bound_count", 0)) > 0:
        policy_candidates.append(
            {
                "kind": "memory-loop",
                "priority": 2,
                "proposed_rule": "Keep home-bound memory append active and route stale homes toward dream-memory.",
                "memory_writes": int(snapshot_summary.get("memory_write_count", 0)),
                "home_bound_events": int(snapshot_summary.get("home_bound_count", 0)),
                "rationale": (
                    "Recent runtime traces show durable memory writes, so memory compression and resurfacing should "
                    "remain part of the default execution loop."
                ),
            }
        )

    if runtime_targets:
        top_runtime = sorted(runtime_targets.items(), key=lambda item: (-int(item[1]), item[0]))[0]
        policy_candidates.append(
            {
                "kind": "runtime-target",
                "priority": 2,
                "proposed_rule": f"Keep collecting target-specific handoff traces for `{top_runtime[0]}`.",
                "target": top_runtime[0],
                "samples": int(top_runtime[1]),
                "rationale": (
                    f"`{top_runtime[0]}` is the most observed runtime target in the recent trace set, which makes it "
                    "the best current edge for adapter and routing evaluation."
                ),
            }
        )

    policy_candidates.sort(key=lambda item: (int(item.get("priority", 99)), item.get("kind", "")))
    report: dict[str, Any] = {
        "schema_version": "0.1.0",
        "review_type": "JiniPolicyReview",
        "generated_at": now_utc(),
        "source_path": display_path(source_path),
        "report_path": "",
        "guardrails": {
            "mutation_allowed": False,
            "requires_human_approval": True,
            "requires_backtest": True,
            "rollback_required": True,
        },
        "learning_snapshot": snapshot_summary,
        "routing_backtest": backtest,
        "runtime_targets": runtime_targets,
        "event_coverage": coverage,
        "coverage_gaps": coverage_gaps,
        "policy_candidates": policy_candidates,
    }

    report_path: Path | None = None
    if pack_dir is not None:
        summary = summarise_pack(pack_dir, load_registry())
        report["pack_id"] = summary.get("pack_id", "")
        report["work_unit_id"] = summary["work_unit"].get("work_unit_id", "")
        report_path = next_policy_review_path(pack_dir)
        report["report_path"] = display_path(report_path)
        report_path.write_text(json.dumps(report, indent=2) + "\n", encoding="utf-8")
        append_learning_event(
            "policy-review",
            {
                "pack_id": report.get("pack_id", ""),
                "work_unit_id": report.get("work_unit_id", ""),
                "event_count": snapshot_summary.get("event_count", 0),
                "candidate_count": len(policy_candidates),
                "runtime_target_count": len(runtime_targets),
                "bounded_harvests": bounded_harvests,
                "memory_appended": int(snapshot_summary.get("memory_write_count", 0)) > 0,
            },
            pack_dir=pack_dir,
        )
    return report, report_path


def print_policy_review(review: dict[str, Any]) -> None:
    print(f"SOURCE {review['source_path']}")
    if review.get("pack_id"):
        print(f"PACK   {review['pack_id']}")
    if review.get("work_unit_id"):
        print(f"WORK   {review['work_unit_id']}")
    if review.get("report_path"):
        print(f"REPORT {review['report_path']}")
    print("GUARD  mutation_allowed=False approval_required=True")
    print(f"EVENTS {review.get('learning_snapshot', {}).get('event_count', 0)}")
    runtime_targets = review.get("runtime_targets", {})
    if runtime_targets:
        print("RUNTIME")
        for target, count in runtime_targets.items():
            print(f"  - {target}: {count}")
    print("CANDIDATES")
    for candidate in review.get("policy_candidates", []):
        detail = candidate.get("proposed_rule") or candidate.get("proposed_execution_class") or ""
        print(f"  - [{candidate.get('kind', '')}] {detail}")


def latest_policy_review_path(pack_dir: Path) -> Path | None:
    review_dir = pack_dir / "runtime" / "policy-reviews"
    if not review_dir.exists():
        return None
    candidates = sorted(review_dir.glob("policy-review-*.json"))
    if not candidates:
        return None
    return candidates[-1]


def next_policy_candidate_path(pack_dir: Path, policy_id: str) -> Path:
    candidate_dir = pack_policy_candidate_dir(pack_dir)
    stamp = datetime.now(timezone.utc).strftime("%Y%m%dT%H%M%S%fZ")
    return candidate_dir / f"{slugify(policy_id)}-candidate-{stamp}.json"


def next_policy_rollout_path(pack_dir: Path, policy_id: str) -> Path:
    rollout_dir = pack_policy_rollout_dir(pack_dir)
    stamp = datetime.now(timezone.utc).strftime("%Y%m%dT%H%M%S%fZ")
    return rollout_dir / f"{slugify(policy_id)}-rollout-{stamp}.json"


def load_active_policy_rollout(pack_dir: Path, policy_id: str = "runtime-routing") -> dict[str, Any] | None:
    path = active_policy_rollout_path(pack_dir, policy_id)
    if not path.exists():
        return None
    try:
        rollout = load_json_file(path)
    except (OSError, json.JSONDecodeError):
        return None
    if not isinstance(rollout, dict):
        return None
    if rollout.get("status") != "active":
        return None
    return rollout


def stage_policy_candidate(
    pack_dir: Path,
    *,
    review_path: Path | None = None,
) -> tuple[dict[str, Any], Path]:
    selected_review_path = review_path or latest_policy_review_path(pack_dir)
    if selected_review_path is None or not selected_review_path.exists():
        raise ValueError("No policy review report is available to stage")
    review = load_json_file(selected_review_path)
    if not isinstance(review, dict):
        raise ValueError("Policy review report must be a mapping")

    routing_overrides: dict[str, str] = {}
    recommended_runtime_target = ""
    promotion_gate_required = False
    compact_reload_preferred = False
    memory_loop_required = False
    candidate_items = []
    for item in review.get("policy_candidates", []):
        if not isinstance(item, dict):
            continue
        candidate_items.append(item)
        kind = str(item.get("kind", "")).strip()
        if kind == "routing-default":
            intent = str(item.get("intent", "")).strip()
            execution_class = str(item.get("proposed_execution_class", "")).strip()
            if intent and execution_class in {"cheap", "standard", "deep"}:
                routing_overrides[intent] = execution_class
        elif kind == "runtime-target":
            recommended_runtime_target = str(item.get("target", "")).strip()
        elif kind == "promotion-gate":
            promotion_gate_required = True
        elif kind == "compact-reload":
            compact_reload_preferred = True
        elif kind == "memory-loop":
            memory_loop_required = True

    policy_id = "runtime-routing"
    candidate_path = next_policy_candidate_path(pack_dir, policy_id)
    candidate_id = candidate_path.stem
    payload = {
        "schema_version": "0.1.0",
        "candidate_type": "JiniPolicyCandidate",
        "generated_at": now_utc(),
        "candidate_id": candidate_id,
        "policy_id": policy_id,
        "status": "proposed",
        "pack_id": review.get("pack_id", ""),
        "work_unit_id": review.get("work_unit_id", ""),
        "source_review_path": display_path(selected_review_path),
        "intent_overrides": routing_overrides,
        "recommended_runtime_target": recommended_runtime_target,
        "promotion_gate_required": promotion_gate_required,
        "compact_reload_preferred": compact_reload_preferred,
        "memory_loop_required": memory_loop_required,
        "guardrails": review.get("guardrails", {}),
        "candidate_items": candidate_items,
        "review_summary": {
            "event_count": review.get("learning_snapshot", {}).get("event_count", 0),
            "runtime_targets": review.get("runtime_targets", {}),
        },
    }
    candidate_path.write_text(json.dumps(payload, indent=2) + "\n", encoding="utf-8")
    append_learning_event(
        "policy-candidate-staged",
        {
            "pack_id": payload["pack_id"],
            "work_unit_id": payload["work_unit_id"],
            "candidate_id": candidate_id,
            "policy_id": policy_id,
            "override_count": len(routing_overrides),
            "recommended_runtime_target": recommended_runtime_target,
        },
        pack_dir=pack_dir,
    )
    return payload, candidate_path


def approve_policy_candidate(
    pack_dir: Path,
    candidate_path: Path,
    *,
    approver: str,
) -> tuple[dict[str, Any], Path]:
    candidate = load_json_file(candidate_path)
    if not isinstance(candidate, dict):
        raise ValueError("Policy candidate must be a mapping")
    policy_id = str(candidate.get("policy_id", "")).strip() or "runtime-routing"
    candidate["status"] = "approved"
    candidate["approved_at"] = now_utc()
    candidate["approved_by"] = approver
    candidate_path.write_text(json.dumps(candidate, indent=2) + "\n", encoding="utf-8")

    rollout_path = next_policy_rollout_path(pack_dir, policy_id)
    rollout = {
        "schema_version": "0.1.0",
        "rollout_type": "JiniPolicyRollout",
        "generated_at": now_utc(),
        "rollout_id": rollout_path.stem,
        "policy_id": policy_id,
        "candidate_id": candidate.get("candidate_id", ""),
        "status": "active",
        "approved_by": approver,
        "source_candidate_path": display_path(candidate_path),
        "intent_overrides": candidate.get("intent_overrides", {}),
        "recommended_runtime_target": candidate.get("recommended_runtime_target", ""),
        "promotion_gate_required": bool(candidate.get("promotion_gate_required", False)),
        "compact_reload_preferred": bool(candidate.get("compact_reload_preferred", False)),
        "memory_loop_required": bool(candidate.get("memory_loop_required", False)),
        "guardrails": candidate.get("guardrails", {}),
    }
    rollout_path.write_text(json.dumps(rollout, indent=2) + "\n", encoding="utf-8")
    active_path = active_policy_rollout_path(pack_dir, policy_id)
    active_path.write_text(json.dumps(rollout, indent=2) + "\n", encoding="utf-8")
    append_learning_event(
        "policy-rollout-activated",
        {
            "pack_id": candidate.get("pack_id", ""),
            "work_unit_id": candidate.get("work_unit_id", ""),
            "candidate_id": candidate.get("candidate_id", ""),
            "policy_id": policy_id,
            "approver": approver,
            "override_count": len(candidate.get("intent_overrides", {})),
            "recommended_runtime_target": candidate.get("recommended_runtime_target", ""),
        },
        pack_dir=pack_dir,
    )
    return rollout, rollout_path


def rollback_policy_candidate(
    pack_dir: Path,
    candidate_path: Path,
    *,
    actor: str,
    reason: str,
) -> tuple[dict[str, Any], Path]:
    candidate = load_json_file(candidate_path)
    if not isinstance(candidate, dict):
        raise ValueError("Policy candidate must be a mapping")
    policy_id = str(candidate.get("policy_id", "")).strip() or "runtime-routing"
    active_path = active_policy_rollout_path(pack_dir, policy_id)
    active_rollout = load_active_policy_rollout(pack_dir, policy_id=policy_id)
    if active_rollout is None:
        raise ValueError(f"No active rollout is present for policy {policy_id!r}")
    if str(active_rollout.get("candidate_id", "")).strip() != str(candidate.get("candidate_id", "")).strip():
        raise ValueError("Active rollout does not match the requested candidate")

    rollback_path = next_policy_rollout_path(pack_dir, f"{policy_id}-rollback")
    rollback = {
        "schema_version": "0.1.0",
        "rollback_type": "JiniPolicyRollback",
        "generated_at": now_utc(),
        "policy_id": policy_id,
        "candidate_id": candidate.get("candidate_id", ""),
        "status": "rolled-back",
        "rolled_back_by": actor,
        "reason": reason,
        "source_candidate_path": display_path(candidate_path),
        "source_rollout_path": display_path(active_path),
    }
    rollback_path.write_text(json.dumps(rollback, indent=2) + "\n", encoding="utf-8")
    active_path.unlink(missing_ok=True)
    candidate["status"] = "rolled-back"
    candidate["rolled_back_at"] = rollback["generated_at"]
    candidate["rolled_back_by"] = actor
    candidate["rollback_reason"] = reason
    candidate_path.write_text(json.dumps(candidate, indent=2) + "\n", encoding="utf-8")
    append_learning_event(
        "policy-rollout-rolled-back",
        {
            "pack_id": candidate.get("pack_id", ""),
            "work_unit_id": candidate.get("work_unit_id", ""),
            "candidate_id": candidate.get("candidate_id", ""),
            "policy_id": policy_id,
            "actor": actor,
        },
        pack_dir=pack_dir,
    )
    return rollback, rollback_path


def print_policy_candidate(candidate: dict[str, Any]) -> None:
    print(f"POLICY {candidate.get('policy_id', '')}")
    print(f"STATUS {candidate.get('status', '')}")
    print(f"ID     {candidate.get('candidate_id', '')}")
    print(f"REVIEW {candidate.get('source_review_path', '')}")
    if candidate.get("intent_overrides"):
        print("OVERRIDES")
        for intent, execution_class in sorted(candidate.get("intent_overrides", {}).items()):
            print(f"  - {intent}: {execution_class}")


def print_policy_rollout(rollout: dict[str, Any]) -> None:
    print(f"POLICY {rollout.get('policy_id', '')}")
    print(f"STATUS {rollout.get('status', '')}")
    print(f"CANDIDATE {rollout.get('candidate_id', '')}")
    if rollout.get("intent_overrides"):
        print("OVERRIDES")
        for intent, execution_class in sorted(rollout.get("intent_overrides", {}).items()):
            print(f"  - {intent}: {execution_class}")
    if rollout.get("recommended_runtime_target"):
        print(f"RUNTIME {rollout.get('recommended_runtime_target', '')}")


def next_harvest_report_path(pack_dir: Path) -> Path:
    report_dir = pack_dir / "runtime" / "harvests"
    report_dir.mkdir(parents=True, exist_ok=True)
    stamp = datetime.now(timezone.utc).strftime("%Y%m%dT%H%M%S%fZ")
    return report_dir / f"evidence-harvest-{stamp}.json"


def select_verification_targets(
    repo_context: dict[str, Any],
    categories: list[str],
    max_targets: int,
) -> list[dict[str, str]]:
    selected: list[dict[str, str]] = []
    seen: set[tuple[str, str, str]] = set()
    targets = repo_context.get("verification_targets", [])
    for category in categories:
        for target in targets:
            if target.get("category") != category:
                continue
            identity = (
                str(target.get("category", "")),
                str(target.get("command", "")),
                str(target.get("path", "")),
            )
            if identity in seen:
                continue
            selected.append(deepcopy(target))
            seen.add(identity)
            if len(selected) >= max_targets:
                return selected
    return selected


def run_verification_target(
    repo_root: Path,
    target: dict[str, Any],
    timeout_seconds: int,
) -> dict[str, Any]:
    result: dict[str, Any] = {
        "category": target.get("category", ""),
        "artifact": target.get("artifact", ""),
        "label": target.get("label", ""),
        "command": target.get("command", ""),
        "path": target.get("path", ""),
        "source": target.get("source", ""),
        "started_at": now_utc(),
        "duration_ms": 0,
        "status": "skipped",
        "exit_code": None,
        "stdout_excerpt": "",
        "stderr_excerpt": "",
    }

    command = result["command"].strip()
    argv = [str(item) for item in target.get("argv", [])] if isinstance(target.get("argv"), list) else []
    path_text = result["path"].strip()
    started = time.perf_counter()

    if argv:
        try:
            completed = subprocess.run(
                argv,
                cwd=repo_root,
                check=False,
                capture_output=True,
                text=True,
                timeout=timeout_seconds,
            )
        except subprocess.TimeoutExpired as exc:
            result["status"] = "timed_out"
            result["stdout_excerpt"] = trim_output(exc.stdout)
            result["stderr_excerpt"] = trim_output(exc.stderr)
        except OSError as exc:
            result["status"] = "failed"
            result["stderr_excerpt"] = str(exc)
        else:
            result["status"] = "passed" if completed.returncode == 0 else "failed"
            result["exit_code"] = completed.returncode
            result["stdout_excerpt"] = trim_output(completed.stdout)
            result["stderr_excerpt"] = trim_output(completed.stderr)
    elif command:
        try:
            completed = subprocess.run(
                shlex.split(command),
                cwd=repo_root,
                check=False,
                capture_output=True,
                text=True,
                timeout=timeout_seconds,
            )
        except ValueError as exc:
            result["status"] = "failed"
            result["stderr_excerpt"] = str(exc)
        except subprocess.TimeoutExpired as exc:
            result["status"] = "timed_out"
            result["stdout_excerpt"] = trim_output(exc.stdout)
            result["stderr_excerpt"] = trim_output(exc.stderr)
        except OSError as exc:
            result["status"] = "failed"
            result["stderr_excerpt"] = str(exc)
        else:
            result["status"] = "passed" if completed.returncode == 0 else "failed"
            result["exit_code"] = completed.returncode
            result["stdout_excerpt"] = trim_output(completed.stdout)
            result["stderr_excerpt"] = trim_output(completed.stderr)
    elif path_text:
        observed_path = resolve_display_path(path_text)
        result["status"] = "observed" if observed_path.exists() else "missing"
    result["duration_ms"] = int((time.perf_counter() - started) * 1000)
    return result


def summarize_harvest_result(result: dict[str, Any]) -> str:
    detail = result.get("command") or result.get("path") or result.get("label") or "<unknown>"
    status = str(result.get("status", "")).upper()
    extras: list[str] = []
    exit_code = result.get("exit_code")
    if exit_code is not None:
        extras.append(f"exit={exit_code}")
    duration_ms = int(result.get("duration_ms", 0) or 0)
    if duration_ms > 0:
        extras.append(f"{duration_ms}ms")
    summary = f"{status} `{detail}`"
    if extras:
        summary += f" ({', '.join(extras)})"
    stderr_excerpt = str(result.get("stderr_excerpt", "")).strip()
    if stderr_excerpt and status != "PASSED":
        summary += f" stderr: {stderr_excerpt}"
    return summary

def canonical_artifact_type(registry: dict[str, Any], artifact_type: str) -> str:
    for canonical, meta in registry["artifacts"].items():
        if artifact_type == canonical:
            return canonical
    return artifact_type


def resolve_pack_manifest(pack_dir: Path) -> tuple[dict[str, Any] | None, Path | None]:
    pack_instance_path = pack_dir / "pack-instance.yaml"
    if pack_instance_path.exists():
        pack_instance = load_document(pack_instance_path)
        pack_id = pack_instance.get("pack_id")
        if isinstance(pack_id, str):
            manifest_path = PACKS_ROOT / pack_id / "pack.yaml"
            if manifest_path.exists():
                return load_document(manifest_path), manifest_path

    for candidate in [pack_dir, *pack_dir.parents]:
        manifest_path = candidate / "pack.yaml"
        if manifest_path.exists():
            return load_document(manifest_path), manifest_path
        if candidate == ROOT:
            break
    return None, None


def load_bootstrap_policy(pack_id: str) -> dict[str, Any] | None:
    policy_path = POLICY_ROOT / f"{pack_id}-bootstrap-policy.yaml"
    if not policy_path.exists():
        return None
    return load_document(policy_path)


def recommended_bootstrap_mode(pack_id: str, manifest: dict[str, Any]) -> tuple[str, list[str]]:
    policy = load_bootstrap_policy(pack_id)
    rationale: list[str] = []
    if policy is not None:
        action = str(policy.get("recommended_action", "")).strip()
        if action in {"init-pack", "compile-pack"}:
            rationale.extend(policy.get("rationale", []))
            return action, rationale

    if manifest.get("target_profile") in HIGH_CONTROL_PROFILES:
        rationale.append("High-control target profile defaults to compile-pack")
        return "compile-pack", rationale
    if len(manifest.get("emits", [])) >= 8:
        rationale.append("Large emitted artifact surface defaults to compile-pack")
        return "compile-pack", rationale
    rationale.append("Fallback bootstrap mode is init-pack")
    return "init-pack", rationale


def load_pack_artifacts(
    pack_dir: Path,
    registry: dict[str, Any],
) -> tuple[list[tuple[Path, dict[str, Any], str]], dict[str, tuple[Path, dict[str, Any]]]]:
    artifact_records: list[tuple[Path, dict[str, Any], str]] = []
    latest_by_type: dict[str, tuple[Path, dict[str, Any]]] = {}
    artifact_dir = pack_dir / "artifacts"
    artifact_paths = sorted(artifact_dir.glob("*.y*ml")) + sorted(artifact_dir.glob("*.json"))
    for artifact_path in artifact_paths:
        artifact_doc = load_document(artifact_path)
        artifact_type = canonical_artifact_type(registry, artifact_doc.get("artifact_type", ""))
        artifact_records.append((artifact_path, artifact_doc, artifact_type))
        existing = latest_by_type.get(artifact_type)
        if existing is None or artifact_doc.get("revision", 0) >= existing[1].get("revision", 0):
            latest_by_type[artifact_type] = (artifact_path, artifact_doc)
    return artifact_records, latest_by_type


def summarise_task_status(tasks_doc: dict[str, Any] | None) -> dict[str, Any]:
    if not tasks_doc:
        return {"total": 0, "done": 0, "unresolved": 0, "counts": Counter(), "blocked_by": []}

    statuses = [str(value).strip().lower() for value in tasks_doc.get("status_per_task", [])]
    counts = Counter(statuses)
    done = sum(count for status, count in counts.items() if status in DONE_TASK_STATUSES)
    unresolved = sum(count for status, count in counts.items() if status not in DONE_TASK_STATUSES)
    return {
        "total": len(statuses),
        "done": done,
        "unresolved": unresolved,
        "counts": counts,
        "blocked_by": tasks_doc.get("blocked_by", []),
    }


def next_artifact_prefix(artifact_dir: Path) -> str:
    max_prefix = 0
    for candidate in artifact_dir.glob("*.y*ml"):
        match = re.match(r"^([0-9]+)-", candidate.name)
        if match:
            max_prefix = max(max_prefix, int(match.group(1)))
    for candidate in artifact_dir.glob("*.json"):
        match = re.match(r"^([0-9]+)-", candidate.name)
        if match:
            max_prefix = max(max_prefix, int(match.group(1)))
    return f"{max_prefix + 1:02d}"


def required_artifacts_for_state(state: str, manifest: dict[str, Any] | None) -> list[str]:
    stage_required = list(STATE_REQUIRED_ARTIFACTS.get(state, []))
    if manifest is None:
        return stage_required
    allowed = set(manifest.get("emits", []))
    return [artifact for artifact in stage_required if artifact in allowed]


def derive_blockers(
    work_unit_doc: dict[str, Any],
    latest_by_type: dict[str, tuple[Path, dict[str, Any]]],
    missing_stage_required: list[str],
    validation_errors: list[str],
) -> list[str]:
    blockers: list[str] = []
    state = work_unit_doc.get("current_state", "")
    profile = work_unit_doc.get("profile_id", "")

    if validation_errors:
        blockers.append(f"{len(validation_errors)} validation error(s) must be fixed")

    for artifact_type in missing_stage_required:
        blockers.append(f"Missing required artifact: {artifact_type}")

    for artifact_type, (_, artifact_doc) in latest_by_type.items():
        if artifact_type in STATE_REQUIRED_ARTIFACTS.get(state, []):
            if artifact_doc.get("status") not in READY_ARTIFACT_STATUSES:
                blockers.append(
                    f"{artifact_type} status {artifact_doc.get('status')!r} is not ready for state {state}"
                )

    if profile in HIGH_CONTROL_PROFILES and not work_unit_doc.get("approver_actor_ids"):
        blockers.append(f"{profile} profile requires approver_actor_ids")

    evidence = latest_by_type.get("Evidence")
    if state in VERIFY_STATES:
        if evidence is None:
            blockers.append(f"State {state} requires Evidence")
        elif evidence[1].get("status") not in READY_ARTIFACT_STATUSES:
            blockers.append(
                f"Evidence status {evidence[1].get('status')!r} is not ready for state {state}"
            )

    if state == "operational":
        for artifact_type in ("Runbook", "Signals", "Rollback", "Approval"):
            if artifact_type not in latest_by_type:
                blockers.append(f"Operational state requires {artifact_type}")

    tasks_doc = latest_by_type.get("Tasks", (None, None))[1]
    task_summary = summarise_task_status(tasks_doc)
    if state in VERIFY_STATES and task_summary["unresolved"] > 0:
        blockers.append(
            f"State {state} conflicts with unresolved task statuses: "
            f"{task_summary['unresolved']} of {task_summary['total']} task(s) not done"
        )

    return blockers


def infer_pack_health(
    work_unit_doc: dict[str, Any],
    validation_errors: list[str],
    blockers: list[str],
) -> str:
    state = work_unit_doc.get("current_state", "")
    next_operation = NEXT_OPERATION_BY_STATE.get(state, "Unknown")
    if validation_errors:
        return "invalid"
    if blockers:
        return "blocked"
    if state == "operational":
        return "operational"
    if next_operation == "Verify":
        return "ready-to-verify"
    if next_operation == "Make":
        return "ready-to-make"
    if next_operation == "Decide":
        return "ready-to-decide"
    if next_operation == "Model":
        return "ready-to-model"
    if next_operation == "Probe":
        return "ready-to-probe"
    if next_operation == "Make":
        return "in-make"
    if next_operation in {"Scope", "Probe", "Model", "Decide"}:
        return "in-design"
    return "in-progress"


def summarise_pack(pack_dir: Path, registry: dict[str, Any]) -> dict[str, Any]:
    errors, warnings = validate_pack(pack_dir, registry)
    work_unit_path = pack_dir / "work-unit.yaml"
    work_unit_doc = load_document(work_unit_path)
    manifest, manifest_path = resolve_pack_manifest(pack_dir)
    artifact_records, latest_by_type = load_pack_artifacts(pack_dir, registry)

    full_required_artifacts = manifest.get("emits", []) if manifest else []
    stage_required_artifacts = required_artifacts_for_state(work_unit_doc.get("current_state", ""), manifest)
    present_types = sorted(latest_by_type.keys())
    missing_full_required = sorted(set(full_required_artifacts) - set(present_types))
    missing_stage_required = sorted(set(stage_required_artifacts) - set(present_types))
    ready_stage_required = [
        artifact_type
        for artifact_type in stage_required_artifacts
        if artifact_type in latest_by_type
        and latest_by_type[artifact_type][1].get("status") in READY_ARTIFACT_STATUSES
    ]
    blockers = derive_blockers(work_unit_doc, latest_by_type, missing_stage_required, errors)
    tasks_doc = latest_by_type.get("Tasks", (None, None))[1]
    task_summary = summarise_task_status(tasks_doc)
    evidence_doc = latest_by_type.get("Evidence", (None, None))[1]
    health = infer_pack_health(work_unit_doc, errors, blockers)

    pack_id = manifest.get("pack_id", "") if manifest else ""
    compiled_flow = manifest.get("compiled_flow", []) if manifest else []
    control_packs = manifest.get("control_packs", []) if manifest else []

    return {
        "pack_dir": pack_dir,
        "pack_id": pack_id,
        "manifest": manifest,
        "manifest_path": manifest_path,
        "work_unit": work_unit_doc,
        "validation_errors": errors,
        "validation_warnings": warnings,
        "full_required_artifacts": full_required_artifacts,
        "stage_required_artifacts": stage_required_artifacts,
        "missing_full_required": missing_full_required,
        "missing_stage_required": missing_stage_required,
        "ready_stage_required": ready_stage_required,
        "present_types": present_types,
        "artifact_records": artifact_records,
        "latest_by_type": latest_by_type,
        "health": health,
        "next_operation": NEXT_OPERATION_BY_STATE.get(work_unit_doc.get("current_state", ""), "Unknown"),
        "compiled_flow": compiled_flow,
        "control_packs": control_packs,
        "task_summary": task_summary,
        "blockers": blockers,
        "evidence_doc": evidence_doc,
    }


def transition_blockers(summary: dict[str, Any], target_state: str) -> list[str]:
    work_unit_doc = summary["work_unit"]
    manifest = summary["manifest"]
    latest_by_type = summary["latest_by_type"]
    validation_errors = summary["validation_errors"]
    task_summary = summary["task_summary"]
    profile = work_unit_doc.get("profile_id", "")
    blockers: list[str] = []

    if validation_errors:
        blockers.append(f"{len(validation_errors)} validation error(s) must be fixed first")

    required_artifacts = required_artifacts_for_state(target_state, manifest)
    for artifact_type in required_artifacts:
        if artifact_type not in latest_by_type:
            blockers.append(f"Target state {target_state} requires missing artifact {artifact_type}")
            continue
        artifact_doc = latest_by_type[artifact_type][1]
        if artifact_doc.get("status") not in READY_ARTIFACT_STATUSES:
            blockers.append(
                f"Target state {target_state} requires {artifact_type} to be ready; "
                f"found status {artifact_doc.get('status')!r}"
            )

    if profile in HIGH_CONTROL_PROFILES and not work_unit_doc.get("approver_actor_ids"):
        blockers.append(f"{profile} profile requires approver_actor_ids")

    if target_state in {"awaiting_verification", "operational"} and task_summary["unresolved"] > 0:
        blockers.append(
            f"Target state {target_state} requires all tasks done; "
            f"{task_summary['unresolved']} unresolved task(s) remain"
        )

    if target_state in VERIFY_STATES:
        evidence = latest_by_type.get("Evidence")
        if evidence is None:
            blockers.append(f"Target state {target_state} requires Evidence")
        elif evidence[1].get("status") not in READY_ARTIFACT_STATUSES:
            blockers.append(
                f"Target state {target_state} requires Evidence to be ready; "
                f"found status {evidence[1].get('status')!r}"
            )

    if target_state == "operational":
        for artifact_type in ("Runbook", "Signals", "Rollback", "Approval"):
            if artifact_type not in latest_by_type:
                blockers.append(f"Operational transition requires {artifact_type}")
            elif latest_by_type[artifact_type][1].get("status") not in READY_ARTIFACT_STATUSES:
                blockers.append(
                    f"Operational transition requires ready {artifact_type}; "
                    f"found status {latest_by_type[artifact_type][1].get('status')!r}"
                )

    return blockers


def advance_pack_state(
    pack_dir: Path,
    registry: dict[str, Any],
    target_state: str | None = None,
) -> tuple[str, str]:
    summary = summarise_pack(pack_dir, registry)
    current_state = summary["work_unit"].get("current_state", "")

    if current_state in {"reopened", "incident", "retired"}:
        raise ValueError(f"advance-pack does not support automatic promotion from state {current_state!r}")

    if current_state not in NEXT_STATE_BY_PROGRESS:
        raise ValueError(f"No forward transition defined from state {current_state!r}")

    expected_next = NEXT_STATE_BY_PROGRESS[current_state]
    destination = target_state or expected_next
    if destination != expected_next:
        raise ValueError(
            f"advance-pack only allows the next linear state; from {current_state!r} "
            f"expected {expected_next!r}, got {destination!r}"
        )

    blockers = transition_blockers(summary, destination)
    if blockers:
        raise ValueError("Cannot advance pack:\n- " + "\n- ".join(blockers))

    work_unit_path = pack_dir / "work-unit.yaml"
    work_unit_doc = summary["work_unit"]
    work_unit_doc["current_state"] = destination
    work_unit_doc["updated_at"] = now_utc()
    dump_document(work_unit_path, work_unit_doc)
    return current_state, destination


def capture_evidence(
    pack_dir: Path,
    registry: dict[str, Any],
    author_actor_id: str,
    claims: list[str],
    test_results: list[str],
    review_results: list[str],
    operational_results: list[str],
    residual_risks: list[str],
    approver_actor_ids: list[str] | None = None,
    target_artifact_type: str | None = None,
    references: list[str] | None = None,
    status: str = "reviewed",
) -> Path:
    summary = summarise_pack(pack_dir, registry)
    work_unit = summary["work_unit"]
    latest_by_type = summary["latest_by_type"]
    artifact_dir = pack_dir / "artifacts"

    if not claims:
        raise ValueError("capture-evidence requires at least one --claim")
    if not test_results:
        raise ValueError("capture-evidence requires at least one --test-result")
    if not review_results:
        raise ValueError("capture-evidence requires at least one --review-result")
    if not operational_results:
        raise ValueError("capture-evidence requires at least one --operational-result")

    target_type = target_artifact_type or "Spec"
    if target_type not in latest_by_type:
        raise ValueError(f"capture-evidence could not find target artifact type {target_type!r}")

    _, target_doc = latest_by_type[target_type]
    existing = latest_by_type.get("Evidence")
    timestamp = now_utc()
    approvers = approver_actor_ids if approver_actor_ids is not None else work_unit.get("approver_actor_ids", [])

    if existing is None:
        artifact_path = artifact_dir / f"{next_artifact_prefix(artifact_dir)}-evidence.yaml"
        evidence = base_artifact(
            f"evidence-{work_unit['work_unit_id']}",
            "Evidence",
            work_unit["work_unit_id"],
            work_unit["branch_id"],
            author_actor_id,
            approvers,
            timestamp,
            status=status,
            references=[],
        )
    else:
        artifact_path, evidence = existing
        evidence["revision"] = int(evidence.get("revision", 0)) + 1
        evidence["updated_at"] = timestamp
        evidence["author_actor_id"] = author_actor_id
        evidence["approver_actor_ids"] = approvers
        evidence["status"] = status

    evidence["references"] = list(
        dict.fromkeys(
            [
                target_doc["artifact_id"],
                *(references or []),
            ]
        )
    )
    evidence["target_artifact_id"] = target_doc["artifact_id"]
    evidence["target_revision"] = int(target_doc.get("revision", 1))
    evidence["claims_validated"] = claims
    evidence["test_results"] = test_results
    evidence["review_results"] = review_results
    evidence["operational_results"] = operational_results
    evidence["residual_risks"] = residual_risks

    dump_document(artifact_path, evidence)
    return artifact_path


def harvest_evidence(
    pack_dir: Path,
    registry: dict[str, Any],
    author_actor_id: str,
    repo_path: Path | None = None,
    home_path: Path | None = None,
    categories: list[str] | None = None,
    claims: list[str] | None = None,
    residual_risks: list[str] | None = None,
    approver_actor_ids: list[str] | None = None,
    target_artifact_type: str | None = None,
    references: list[str] | None = None,
    status: str = "reviewed",
    timeout_seconds: int = 20,
    max_targets: int = 5,
) -> tuple[dict[str, Any], Path, Path]:
    if timeout_seconds < 1:
        raise ValueError("harvest-evidence requires --timeout-seconds >= 1")
    if max_targets < 1:
        raise ValueError("harvest-evidence requires --max-targets >= 1")

    summary = summarise_pack(pack_dir, registry)
    work_unit = summary["work_unit"]
    latest_by_type = summary["latest_by_type"]
    target_type = target_artifact_type or "Spec"
    if target_type not in latest_by_type:
        raise ValueError(f"harvest-evidence could not find target artifact type {target_type!r}")

    repo_context = inspect_repo_context(pack_dir, repo_path=repo_path)
    if not repo_context.get("discovered"):
        raise ValueError("harvest-evidence requires repo context; pass --repo or run from within a repo-backed pack")

    selected_categories = categories or list(HARVEST_CATEGORY_ORDER)
    repo_root = resolve_display_path(str(repo_context.get("repo_root", "")))
    selected_targets = select_verification_targets(repo_context, selected_categories, max_targets)
    if not selected_targets:
        raise ValueError(
            "harvest-evidence could not find any repo verification targets for the requested categories"
        )

    target_path, target_doc = latest_by_type[target_type]
    results = [run_verification_target(repo_root, target, timeout_seconds) for target in selected_targets]
    counts = Counter(result["status"] for result in results)
    command_targets = sum(1 for result in results if result.get("command"))
    readiness = "ready"
    if command_targets == 0 or counts["failed"] or counts["timed_out"] or counts["missing"]:
        readiness = "bounded"

    effective_status = status
    auto_downgraded = False
    if readiness != "ready" and effective_status in READY_ARTIFACT_STATUSES:
        effective_status = "draft"
        auto_downgraded = True

    report_path = next_harvest_report_path(pack_dir)
    report_reference = display_path(report_path)

    claims_validated = list(dict.fromkeys([
        f"Harvested repo verification evidence for `{display_path(pack_dir)}` against {target_type} revision {int(target_doc.get('revision', 1))}.",
        f"{counts['passed']} target(s) passed, {counts['failed']} failed, {counts['timed_out']} timed out, and {counts['observed']} passive surface(s) were observed.",
        *((claims or [])),
    ]))

    test_results: list[str] = []
    review_results: list[str] = []
    operational_results: list[str] = []
    derived_risks = list(residual_risks or [])

    for result in results:
        summary_line = summarize_harvest_result(result)
        category = str(result.get("category", ""))
        if category in {"test", "verify"}:
            test_results.append(summary_line)
        elif category == "docs":
            review_results.append(summary_line)
        else:
            operational_results.append(summary_line)

        if result["status"] in {"failed", "timed_out", "missing"}:
            derived_risks.append(
                f"{summary_line} prevented the harvest from producing ready verification evidence."
            )

    if not test_results:
        test_results.append("No explicit test or verify command was harvested; automated proof remains bounded.")
    if not review_results:
        review_results.append(
            f"Reviewed harvest report `{report_reference}` and the current repo verification target set for the active revision."
        )
    if not operational_results:
        operational_results.append(
            "No startup or demo command was harvested; operational validation remains bounded to non-runtime repo surfaces."
        )
    if command_targets == 0:
        derived_risks.append(
            "No executable repo verification commands were harvested; the resulting evidence is observational only."
        )
    if auto_downgraded:
        derived_risks.append(
            "Evidence status was downgraded to draft because the harvested checks did not establish ready verification proof."
        )

    report: dict[str, Any] = {
        "schema_version": "0.1.0",
        "report_type": "JiniEvidenceHarvestReport",
        "generated_at": now_utc(),
        "pack_id": summary.get("pack_id", ""),
        "work_unit_id": work_unit.get("work_unit_id", ""),
        "repo_root": display_path(repo_root),
        "requested_categories": selected_categories,
        "timeout_seconds": timeout_seconds,
        "max_targets": max_targets,
        "target_artifact_type": target_type,
        "target_artifact_id": target_doc.get("artifact_id", ""),
        "target_artifact_path": display_path(target_path),
        "target_revision": int(target_doc.get("revision", 1) or 1),
        "selected_targets": selected_targets,
        "results": results,
        "summary": {
            "total": len(results),
            "command_targets": command_targets,
            "passed": counts["passed"],
            "failed": counts["failed"],
            "timed_out": counts["timed_out"],
            "observed": counts["observed"],
            "missing": counts["missing"],
            "skipped": counts["skipped"],
        },
        "readiness": readiness,
        "evidence_status": effective_status,
        "auto_downgraded": auto_downgraded,
        "claims_validated": claims_validated,
        "test_results": test_results,
        "review_results": review_results,
        "operational_results": operational_results,
        "residual_risks": list(dict.fromkeys(derived_risks)),
        "report_path": report_reference,
        "evidence_artifact_path": "",
    }
    report_path.write_text(json.dumps(report, indent=2) + "\n", encoding="utf-8")

    evidence_path = capture_evidence(
        pack_dir,
        registry,
        author_actor_id=author_actor_id,
        claims=claims_validated,
        test_results=test_results,
        review_results=review_results,
        operational_results=operational_results,
        residual_risks=report["residual_risks"],
        approver_actor_ids=approver_actor_ids,
        target_artifact_type=target_type,
        references=[report_reference, *(references or [])],
        status=effective_status,
    )

    report["evidence_artifact_path"] = display_path(evidence_path)
    report_path.write_text(json.dumps(report, indent=2) + "\n", encoding="utf-8")
    home_binding = resolve_home_binding(pack_dir, explicit_home=home_path)
    report["home_binding"] = {
        "bound": bool(home_binding.get("bound")),
        "home_root": str(home_binding.get("home_root", "")) if home_binding.get("home_root") else "",
        "source": home_binding.get("source", ""),
        "binding_path": home_binding.get("binding_path", ""),
    }
    report["memory_append"] = append_home_observation(
        pack_dir,
        home_binding=home_binding,
        line=(
            f"harvest-evidence for {work_unit.get('title', summary.get('pack_id', 'pack'))}: "
            f"readiness={report['readiness']}, evidence_status={report['evidence_status']}, "
            f"passed={report['summary']['passed']}, failed={report['summary']['failed']}."
        ),
    )
    report_path.write_text(json.dumps(report, indent=2) + "\n", encoding="utf-8")
    recommendation = recommend_execution(pack_dir, registry, intent="verify", repo_path=repo_path, home_path=home_path)
    append_learning_event(
        "harvest-evidence",
        {
            "pack_id": summary.get("pack_id", ""),
            "work_unit_id": work_unit.get("work_unit_id", ""),
            "intent": "verify",
            "execution_class": recommendation["execution_class"],
            "repo_root": display_path(repo_root),
            "readiness": report["readiness"],
            "evidence_status": report["evidence_status"],
            "passed": report["summary"]["passed"],
            "failed": report["summary"]["failed"],
            "timed_out": report["summary"]["timed_out"],
            "home_bound": bool(home_binding.get("bound")),
            "memory_appended": bool(report["memory_append"]["appended"]),
            "runtime_target": recommendation["runtime_guidance"]["selected"]["id"],
        },
        pack_dir=pack_dir,
    )
    return report, report_path, evidence_path


def print_harvest_report(report: dict[str, Any]) -> None:
    print(f"PACK   {report['pack_id']}")
    print(f"WORK   {report['work_unit_id']}")
    print(f"REPO   {report['repo_root']}")
    print(
        f"TARGET {report['target_artifact_type']} "
        f"{report['target_artifact_id']} r{report['target_revision']}"
    )
    print(f"READY  {report['readiness']}")
    print(f"STATUS {report['evidence_status']}")
    print(f"REPORT {report['report_path']}")
    print(f"EVID   {report['evidence_artifact_path']}")
    summary = report["summary"]
    print(
        f"CHECKS total={summary['total']} commands={summary['command_targets']} "
        f"passed={summary['passed']} failed={summary['failed']} timed_out={summary['timed_out']}"
    )
    print("TARGETS")
    for result in report["results"]:
        print(f"  - {summarize_harvest_result(result)}")
    if report["residual_risks"]:
        print("RISKS")
        for risk in report["residual_risks"]:
            print(f"  - {risk}")


def capture_approval(
    pack_dir: Path,
    registry: dict[str, Any],
    author_actor_id: str,
    approver_actor_id: str,
    approval_scope: str,
    waivers: list[str],
    conditions: list[str],
    target_artifact_type: str | None = None,
    status: str = "approved",
) -> Path:
    summary = summarise_pack(pack_dir, registry)
    work_unit = summary["work_unit"]
    latest_by_type = summary["latest_by_type"]
    artifact_dir = pack_dir / "artifacts"

    if not approver_actor_id.strip():
        raise ValueError("capture-approval requires a non-empty --approver-actor")
    if not approval_scope.strip():
        raise ValueError("capture-approval requires a non-empty --scope")

    target_type = target_artifact_type or "Evidence"
    if target_type not in latest_by_type:
        raise ValueError(f"capture-approval could not find target artifact type {target_type!r}")

    target_path, target_doc = latest_by_type[target_type]
    timestamp = now_utc()
    existing = latest_by_type.get("Approval")
    approver_ids = list(dict.fromkeys([approver_actor_id, *work_unit.get("approver_actor_ids", [])]))

    if existing is None:
        artifact_path = artifact_dir / f"{next_artifact_prefix(artifact_dir)}-approval.yaml"
        approval = base_artifact(
            f"approval-{work_unit['work_unit_id']}",
            "Approval",
            work_unit["work_unit_id"],
            work_unit["branch_id"],
            author_actor_id,
            approver_ids,
            timestamp,
            status=status,
            references=[],
        )
    else:
        artifact_path, approval = existing
        approval["revision"] = int(approval.get("revision", 0)) + 1
        approval["updated_at"] = timestamp
        approval["author_actor_id"] = author_actor_id
        approval["approver_actor_ids"] = approver_ids
        approval["status"] = status

    approval["references"] = list(
        dict.fromkeys(
            [
                target_doc["artifact_id"],
                *[
                    artifact_doc.get("artifact_id", "")
                    for _, artifact_doc in summary["latest_by_type"].values()
                    if artifact_doc.get("artifact_id")
                ],
            ]
        )
    )
    approval["approved_object_id"] = target_doc["artifact_id"]
    approval["approved_revision"] = int(target_doc.get("revision", 1))
    approval["approver_actor_id"] = approver_actor_id
    approval["approval_scope"] = approval_scope
    approval["waivers"] = waivers
    approval["conditions"] = conditions

    dump_document(artifact_path, approval)
    return artifact_path


def capture_publication(
    pack_dir: Path,
    registry: dict[str, Any],
    author_actor_id: str,
    input_path: Path,
    publication_scope: str,
    approver_actor_ids: list[str] | None = None,
    status: str = "reviewed",
) -> Path:
    summary = summarise_pack(pack_dir, registry)
    work_unit = summary["work_unit"]
    latest_by_type = summary["latest_by_type"]
    artifact_dir = pack_dir / "artifacts"

    if not input_path.exists():
        raise ValueError(f"capture-publication input does not exist: {input_path}")
    if not publication_scope.strip():
        raise ValueError("capture-publication requires a non-empty --scope")

    payload = load_json_file(input_path)
    records = payload.get("records", [])
    if not isinstance(records, list) or not records:
        raise ValueError("capture-publication input must contain a non-empty records array")

    normalized_records: list[dict[str, Any]] = []
    references: list[str] = []
    for idx, record in enumerate(records, start=1):
        if not isinstance(record, dict):
            raise ValueError(f"capture-publication record {idx} must be an object")
        normalized = {
            "adapter": str(record.get("adapter", "")).strip(),
            "target_kind": str(record.get("target_kind", "")).strip(),
            "source_ref": str(record.get("source_ref", "")).strip(),
            "external_id": str(record.get("external_id", "")).strip(),
            "external_url": str(record.get("external_url", "")).strip(),
            "published_at": str(record.get("published_at", "")).strip(),
            "publication_status": str(record.get("publication_status", "")).strip(),
            "notes": record.get("notes", []),
        }
        for key in ("adapter", "target_kind", "source_ref", "external_id", "external_url", "published_at", "publication_status"):
            if not normalized[key]:
                raise ValueError(f"capture-publication record {idx} is missing {key}")
        if not isinstance(normalized["notes"], list):
            raise ValueError(f"capture-publication record {idx} notes must be an array")
        normalized_records.append(normalized)
        references.extend(
            item for item in (normalized["source_ref"], normalized["external_id"], normalized["external_url"]) if item
        )

    existing = latest_by_type.get("Publication")
    approvers = approver_actor_ids if approver_actor_ids is not None else work_unit.get("approver_actor_ids", [])
    timestamp = now_utc()

    if existing is None:
        artifact_path = artifact_dir / f"{next_artifact_prefix(artifact_dir)}-publication.yaml"
        publication = base_artifact(
            f"publication-{work_unit['work_unit_id']}",
            "Publication",
            work_unit["work_unit_id"],
            work_unit["branch_id"],
            author_actor_id,
            approvers,
            timestamp,
            status=status,
            references=[],
        )
    else:
        artifact_path, publication = existing
        publication["revision"] = int(publication.get("revision", 0)) + 1
        publication["updated_at"] = timestamp
        publication["author_actor_id"] = author_actor_id
        publication["approver_actor_ids"] = approvers
        publication["status"] = status

    publication["publication_scope"] = publication_scope
    publication["records"] = normalized_records
    publication["references"] = list(dict.fromkeys(references))

    dump_document(artifact_path, publication)
    return artifact_path


def render_tasks_markdown(summary: dict[str, Any]) -> str:
    work_unit = summary["work_unit"]
    latest_by_type = summary["latest_by_type"]
    tasks_doc = latest_by_type.get("Tasks", (None, None))[1]
    if tasks_doc is None:
        raise ValueError("export-tasks requires a Tasks artifact")
    plan_doc = latest_by_type.get("Plan", (None, None))[1]

    task_rows = list(
        zip_longest(
            tasks_doc.get("tasks", []),
            tasks_doc.get("ownership", []),
            tasks_doc.get("status_per_task", []),
            tasks_doc.get("deliverables", []),
            tasks_doc.get("output_notes", []),
            tasks_doc.get("output_refs", []),
            fillvalue="",
        )
    )

    lines: list[str] = [
        f"# Tasks: {work_unit.get('title', '')}",
        "",
        f"- WorkUnit: `{work_unit.get('work_unit_id', '')}`",
        f"- State: `{work_unit.get('current_state', '')}`",
        f"- Profile: `{work_unit.get('profile_id', '')}`",
        f"- Health: `{summary.get('health', '')}`",
        f"- Next: `{summary.get('next_operation', '')}`",
        "",
        "## Summary",
        f"- Total tasks: `{summary['task_summary']['total']}`",
        f"- Done: `{summary['task_summary']['done']}`",
        f"- Unresolved: `{summary['task_summary']['unresolved']}`",
        "",
        "## Task Board",
    ]

    for idx, (task, owner, status, deliverable, output_note, output_ref) in enumerate(task_rows, start=1):
        normalized = str(status).strip().lower()
        checkbox = "x" if normalized in DONE_TASK_STATUSES else " "
        lines.append(f"{idx}. [{checkbox}] {task}")
        if owner:
            lines.append(f"Owner: `{owner}`")
        if status:
            lines.append(f"Status: `{status}`")
        if deliverable:
            lines.append(f"Deliverable: {deliverable}")
        if output_note:
            lines.append(f"Output: {output_note}")
        if output_ref:
            lines.append(f"Refs: {output_ref}")
        lines.append("")

    blocked_by = tasks_doc.get("blocked_by", [])
    if blocked_by:
        lines.extend(["## Blockers", *[f"- {item}" for item in blocked_by], ""])

    if summary["blockers"]:
        lines.extend(["## Transition Blockers", *[f"- {item}" for item in summary["blockers"]], ""])

    if plan_doc:
        milestones = plan_doc.get("milestones", [])
        gates = plan_doc.get("acceptance_gates", [])
        if milestones:
            lines.extend(["## Milestones", *[f"- {item}" for item in milestones], ""])
        if gates:
            lines.extend(["## Acceptance Gates", *[f"- {item}" for item in gates], ""])

    evidence_doc = summary.get("evidence_doc")
    if evidence_doc:
        lines.extend(
            [
                "## Evidence",
                f"- Target: `{evidence_doc.get('target_artifact_id')}` revision `{evidence_doc.get('target_revision')}`",
                *[f"- Claim: {item}" for item in evidence_doc.get("claims_validated", [])],
                "",
            ]
        )

    return "\n".join(lines).rstrip() + "\n"


def export_tasks(
    pack_dir: Path,
    registry: dict[str, Any],
    output_path: Path | None = None,
) -> Path:
    summary = summarise_pack(pack_dir, registry)
    markdown = render_tasks_markdown(summary)
    if output_path is None:
        output_path = pack_dir / "views" / "tasks.md"
    output_path = output_path.expanduser()
    output_path.parent.mkdir(parents=True, exist_ok=True)
    output_path.write_text(markdown, encoding="utf-8")
    return output_path


def build_task_sync_payload(summary: dict[str, Any]) -> dict[str, Any]:
    work_unit = summary["work_unit"]
    latest_by_type = summary["latest_by_type"]
    tasks_path, tasks_doc = latest_by_type.get("Tasks", (None, None))
    if tasks_doc is None or tasks_path is None:
        raise ValueError("sync-tasks requires a Tasks artifact")
    plan_doc = latest_by_type.get("Plan", (None, None))[1]
    evidence_doc = summary.get("evidence_doc")

    task_rows = list(
        zip_longest(
            tasks_doc.get("tasks", []),
            tasks_doc.get("ownership", []),
            tasks_doc.get("status_per_task", []),
            tasks_doc.get("deliverables", []),
            tasks_doc.get("output_notes", []),
            tasks_doc.get("output_refs", []),
            fillvalue="",
        )
    )

    tasks: list[dict[str, Any]] = []
    for idx, (task, owner, status, deliverable, output_note, output_ref) in enumerate(task_rows, start=1):
        refs = [item.strip() for item in str(output_ref).split(",") if item.strip()]
        normalized_status = str(status).strip().lower()
        tasks.append(
            {
                "task_id": f"{work_unit.get('work_unit_id', '')}-task-{idx}",
                "index": idx,
                "title": task,
                "owner": owner,
                "status": status or "pending",
                "done": normalized_status in DONE_TASK_STATUSES,
                "deliverable": deliverable,
                "output_note": output_note,
                "output_refs": refs,
            }
        )

    payload: dict[str, Any] = {
        "schema_version": "0.1.0",
        "export_type": "JiniTaskSync",
        "generated_at": now_utc(),
        "pack_id": summary.get("pack_id", ""),
        "work_unit": {
            "work_unit_id": work_unit.get("work_unit_id", ""),
            "title": work_unit.get("title", ""),
            "purpose": work_unit.get("purpose", ""),
            "state": work_unit.get("current_state", ""),
            "profile_id": work_unit.get("profile_id", ""),
            "health": summary.get("health", ""),
            "next_operation": summary.get("next_operation", ""),
            "owner_actor_id": work_unit.get("owner_actor_id", ""),
            "approver_actor_ids": work_unit.get("approver_actor_ids", []),
            "stakeholder_actor_ids": work_unit.get("stakeholder_actor_ids", []),
        },
        "source": {
            "pack_path": display_path(summary["pack_dir"]),
            "tasks_artifact_path": display_path(tasks_path),
            "tasks_artifact_id": tasks_doc.get("artifact_id", ""),
            "tasks_revision": tasks_doc.get("revision", 0),
        },
        "summary": {
            "total_tasks": summary["task_summary"]["total"],
            "done_tasks": summary["task_summary"]["done"],
            "unresolved_tasks": summary["task_summary"]["unresolved"],
            "blocked_by": tasks_doc.get("blocked_by", []),
            "transition_blockers": summary["blockers"],
        },
        "milestones": plan_doc.get("milestones", []) if plan_doc else [],
        "acceptance_gates": plan_doc.get("acceptance_gates", []) if plan_doc else [],
        "tasks": tasks,
    }

    if evidence_doc:
        payload["evidence"] = {
            "artifact_id": evidence_doc.get("artifact_id", ""),
            "revision": evidence_doc.get("revision", 0),
            "target_artifact_id": evidence_doc.get("target_artifact_id", ""),
            "target_revision": evidence_doc.get("target_revision", 0),
            "claims_validated": evidence_doc.get("claims_validated", []),
            "residual_risks": evidence_doc.get("residual_risks", []),
        }

    return payload


def sync_tasks(
    pack_dir: Path,
    registry: dict[str, Any],
    output_path: Path | None = None,
) -> Path:
    summary = summarise_pack(pack_dir, registry)
    payload = build_task_sync_payload(summary)
    if output_path is None:
        output_path = pack_dir / "exports" / "tasks-sync.json"
    output_path = output_path.expanduser()
    output_path.parent.mkdir(parents=True, exist_ok=True)
    output_path.write_text(json.dumps(payload, indent=2) + "\n", encoding="utf-8")
    return output_path


def slugify(value: str) -> str:
    normalized = re.sub(r"[^a-z0-9]+", "-", value.strip().lower())
    normalized = normalized.strip("-")
    return normalized or "item"


def build_github_issue_exports(sync_payload: dict[str, Any]) -> tuple[dict[str, Any], list[tuple[str, str]]]:
    work_unit = sync_payload["work_unit"]
    base_labels = [
        "jini",
        f"pack:{sync_payload.get('pack_id', 'unknown')}",
        f"profile:{work_unit.get('profile_id', '').lower()}",
        f"state:{work_unit.get('state', '').lower()}",
    ]

    bundle_issues: list[dict[str, Any]] = []
    markdown_docs: list[tuple[str, str]] = []
    acceptance_gates = sync_payload.get("acceptance_gates", [])
    milestones = sync_payload.get("milestones", [])
    evidence = sync_payload.get("evidence", {})

    for issue in sync_payload.get("tasks", []):
        status = str(issue.get("status", "pending")).strip().lower()
        labels = [*base_labels, f"task-status:{status}"]
        title = issue.get("title", "")
        deliverable = issue.get("deliverable", "")
        owner = issue.get("owner", "")
        output_note = issue.get("output_note", "")
        output_refs = issue.get("output_refs", [])

        lines = [
            f"# {title}",
            "",
            "## Context",
            f"- WorkUnit: `{work_unit.get('work_unit_id', '')}`",
            f"- Title: {work_unit.get('title', '')}",
            f"- Purpose: {work_unit.get('purpose', '')}",
            f"- Profile: `{work_unit.get('profile_id', '')}`",
            f"- State: `{work_unit.get('state', '')}`",
            "",
            "## Task",
            f"- Task ID: `{issue.get('task_id', '')}`",
            f"- Status: `{issue.get('status', '')}`",
            f"- Owner: `{owner}`" if owner else "- Owner: `unassigned`",
            f"- Deliverable: {deliverable}" if deliverable else "- Deliverable: _not specified_",
            "",
        ]
        if milestones:
            lines.extend(["## Milestones", *[f"- {item}" for item in milestones], ""])
        if acceptance_gates:
            lines.extend(["## Acceptance Gates", *[f"- {item}" for item in acceptance_gates], ""])
        if evidence:
            lines.extend(
                [
                    "## Evidence Context",
                    f"- Target: `{evidence.get('target_artifact_id', '')}` revision `{evidence.get('target_revision', 0)}`",
                    *[f"- Claim: {item}" for item in evidence.get("claims_validated", [])],
                    "",
                ]
            )
        if output_note or output_refs:
            lines.append("## Current Output")
            if output_note:
                lines.append(f"- Note: {output_note}")
            if output_refs:
                lines.extend([f"- Ref: {item}" for item in output_refs])
            lines.append("")

        body = "\n".join(lines).rstrip() + "\n"
        bundle_issues.append(
            {
                "task_id": issue.get("task_id", ""),
                "title": title,
                "body_markdown": body,
                "labels": labels,
                "assignee_hint": owner,
                "status": issue.get("status", ""),
                "deliverable": deliverable,
                "source_task_index": issue.get("index", 0),
            }
        )
        filename = f"{int(issue.get('index', 0)):02d}-{slugify(title)}.md"
        markdown_docs.append((filename, body))

    bundle = {
        "schema_version": "0.1.0",
        "export_type": "JiniGithubIssues",
        "generated_at": sync_payload.get("generated_at", now_utc()),
        "pack_id": sync_payload.get("pack_id", ""),
        "work_unit": work_unit,
        "source": sync_payload.get("source", {}),
        "summary": sync_payload.get("summary", {}),
        "issues": bundle_issues,
    }
    return bundle, markdown_docs


def build_jira_issue_exports(sync_payload: dict[str, Any]) -> tuple[dict[str, Any], list[tuple[str, str]]]:
    work_unit = sync_payload["work_unit"]
    base_labels = [
        "jini",
        f"pack-{slugify(sync_payload.get('pack_id', 'unknown'))}",
        f"profile-{slugify(work_unit.get('profile_id', ''))}",
        f"state-{slugify(work_unit.get('state', ''))}",
    ]

    bundle_issues: list[dict[str, Any]] = []
    markdown_docs: list[tuple[str, str]] = []
    acceptance_gates = sync_payload.get("acceptance_gates", [])
    milestones = sync_payload.get("milestones", [])
    evidence = sync_payload.get("evidence", {})

    for issue in sync_payload.get("tasks", []):
        status = str(issue.get("status", "pending")).strip().lower()
        title = issue.get("title", "")
        deliverable = issue.get("deliverable", "")
        owner = issue.get("owner", "")
        output_note = issue.get("output_note", "")
        output_refs = issue.get("output_refs", [])
        labels = [*base_labels, f"task-status-{slugify(status)}"]

        issue_type_hint = "Task"
        if "bug" in title.lower() or "fix" in title.lower():
            issue_type_hint = "Bug"

        lines = [
            f"# {title}",
            "",
            "## Jini Context",
            f"- WorkUnit: `{work_unit.get('work_unit_id', '')}`",
            f"- Pack: `{sync_payload.get('pack_id', '')}`",
            f"- Purpose: {work_unit.get('purpose', '')}",
            f"- Profile: `{work_unit.get('profile_id', '')}`",
            f"- State: `{work_unit.get('state', '')}`",
            f"- Health: `{work_unit.get('health', '')}`",
            "",
            "## Jira Mapping",
            f"- Issue Type Hint: `{issue_type_hint}`",
            f"- Assignee Hint: `{owner}`" if owner else "- Assignee Hint: `unassigned`",
            f"- Status Hint: `{issue.get('status', '')}`",
            "",
            "## Deliverable",
            deliverable or "_not specified_",
            "",
        ]
        if milestones:
            lines.extend(["## Milestones", *[f"- {item}" for item in milestones], ""])
        if acceptance_gates:
            lines.extend(["## Acceptance Gates", *[f"- {item}" for item in acceptance_gates], ""])
        if evidence:
            lines.extend(
                [
                    "## Evidence Context",
                    f"- Target: `{evidence.get('target_artifact_id', '')}` revision `{evidence.get('target_revision', 0)}`",
                    *[f"- Claim: {item}" for item in evidence.get("claims_validated", [])],
                    "",
                ]
            )
        if output_note or output_refs:
            lines.append("## Current Output")
            if output_note:
                lines.append(f"- Note: {output_note}")
            if output_refs:
                lines.extend([f"- Ref: {item}" for item in output_refs])
            lines.append("")

        description = "\n".join(lines).rstrip() + "\n"
        bundle_issues.append(
            {
                "task_id": issue.get("task_id", ""),
                "summary": title,
                "description_markdown": description,
                "labels": labels,
                "issue_type_hint": issue_type_hint,
                "assignee_hint": owner,
                "status_hint": issue.get("status", ""),
                "deliverable": deliverable,
                "source_task_index": issue.get("index", 0),
            }
        )
        filename = f"{int(issue.get('index', 0)):02d}-{slugify(title)}.md"
        markdown_docs.append((filename, description))

    bundle = {
        "schema_version": "0.1.0",
        "export_type": "JiniJiraIssues",
        "generated_at": sync_payload.get("generated_at", now_utc()),
        "pack_id": sync_payload.get("pack_id", ""),
        "work_unit": work_unit,
        "source": sync_payload.get("source", {}),
        "summary": sync_payload.get("summary", {}),
        "issues": bundle_issues,
    }
    return bundle, markdown_docs


def export_issues(
    pack_dir: Path,
    registry: dict[str, Any],
    adapter: str = "github",
    output_dir: Path | None = None,
) -> Path:
    summary = summarise_pack(pack_dir, registry)
    sync_payload = build_task_sync_payload(summary)
    if adapter == "github":
        bundle, markdown_docs = build_github_issue_exports(sync_payload)
    elif adapter == "jira":
        bundle, markdown_docs = build_jira_issue_exports(sync_payload)
    else:
        raise ValueError(f"Unsupported adapter {adapter!r}; supported adapters are 'github' and 'jira'")

    if output_dir is None:
        output_dir = pack_dir / "exports" / "issues" / adapter
    output_dir = output_dir.expanduser()
    output_dir.mkdir(parents=True, exist_ok=True)

    (output_dir / "issues.json").write_text(json.dumps(bundle, indent=2) + "\n", encoding="utf-8")

    index_lines = [
        f"# Issue Export: {summary.get('pack_id', '')}",
        "",
        f"- WorkUnit: `{summary['work_unit'].get('work_unit_id', '')}`",
        f"- Adapter: `{adapter}`",
        f"- Generated: `{bundle.get('generated_at', '')}`",
        "",
        "## Files",
    ]

    for issue, (filename, body) in zip(bundle["issues"], markdown_docs):
        (output_dir / filename).write_text(body, encoding="utf-8")
        index_lines.append(f"- [{filename}](./{filename}) -> `{issue['task_id']}`")

    (output_dir / "README.md").write_text("\n".join(index_lines).rstrip() + "\n", encoding="utf-8")
    return output_dir


def build_confluence_page_export(summary: dict[str, Any]) -> tuple[dict[str, Any], list[tuple[str, str]]]:
    work_unit = summary["work_unit"]
    pack_dir: Path = summary["pack_dir"]
    latest_by_type = summary["latest_by_type"]
    brief_doc = latest_by_type.get("Brief", (None, None))[1]
    plan_doc = latest_by_type.get("Plan", (None, None))[1]
    evidence_doc = summary.get("evidence_doc")

    prd_path = pack_dir / "views" / "prd.md"
    tasks_path = pack_dir / "views" / "tasks.md"
    prd_body = prd_path.read_text(encoding="utf-8").strip() if prd_path.exists() else ""
    tasks_body = tasks_path.read_text(encoding="utf-8").strip() if tasks_path.exists() else render_tasks_markdown(summary).strip()

    overview_lines = [
        f"# {work_unit.get('title', '')}",
        "",
        "## Jini WorkUnit",
        f"- WorkUnit: `{work_unit.get('work_unit_id', '')}`",
        f"- Pack: `{summary.get('pack_id', '')}`",
        f"- Profile: `{work_unit.get('profile_id', '')}`",
        f"- State: `{work_unit.get('current_state', '')}`",
        f"- Health: `{summary.get('health', '')}`",
        f"- Next: `{summary.get('next_operation', '')}`",
        "",
        "## Purpose",
        work_unit.get("purpose", ""),
        "",
    ]
    if brief_doc:
        overview_lines.extend(
            [
                "## Scope Summary",
                brief_doc.get("scope_summary", ""),
                "",
                "## Success Criteria",
                *[f"- {item}" for item in brief_doc.get("success_criteria", [])],
                "",
            ]
        )
    if plan_doc:
        overview_lines.extend(
            [
                "## Milestones",
                *[f"- {item}" for item in plan_doc.get("milestones", [])],
                "",
                "## Acceptance Gates",
                *[f"- {item}" for item in plan_doc.get("acceptance_gates", [])],
                "",
            ]
        )
    if evidence_doc:
        overview_lines.extend(
            [
                "## Evidence Snapshot",
                f"- Target: `{evidence_doc.get('target_artifact_id', '')}` revision `{evidence_doc.get('target_revision', 0)}`",
                *[f"- Claim: {item}" for item in evidence_doc.get("claims_validated", [])],
                "",
            ]
        )

    overview = "\n".join(overview_lines).rstrip() + "\n"
    pages = [("overview.md", overview), ("tasks.md", tasks_body + ("\n" if not tasks_body.endswith("\n") else ""))]
    if prd_body:
        pages.insert(1, ("prd.md", prd_body + ("\n" if not prd_body.endswith("\n") else "")))

    bundle = {
        "schema_version": "0.1.0",
        "export_type": "JiniConfluencePages",
        "generated_at": now_utc(),
        "pack_id": summary.get("pack_id", ""),
        "work_unit": {
            "work_unit_id": work_unit.get("work_unit_id", ""),
            "title": work_unit.get("title", ""),
            "purpose": work_unit.get("purpose", ""),
            "state": work_unit.get("current_state", ""),
            "profile_id": work_unit.get("profile_id", ""),
        },
        "pages": [
            {
                "title": work_unit.get("title", ""),
                "slug": "overview",
                "body_markdown": overview,
                "role": "overview",
            },
            *(
                [
                    {
                        "title": f"{work_unit.get('title', '')} PRD",
                        "slug": "prd",
                        "body_markdown": prd_body + ("\n" if prd_body and not prd_body.endswith("\n") else ""),
                        "role": "prd",
                    }
                ]
                if prd_body
                else []
            ),
            {
                "title": f"{work_unit.get('title', '')} Tasks",
                "slug": "tasks",
                "body_markdown": tasks_body + ("\n" if not tasks_body.endswith("\n") else ""),
                "role": "tasks",
            },
        ],
    }
    return bundle, pages


def build_markdown_page_export(summary: dict[str, Any]) -> tuple[dict[str, Any], list[tuple[str, str]]]:
    bundle, pages = build_confluence_page_export(summary)
    markdown_bundle = deepcopy(bundle)
    markdown_bundle["export_type"] = "JiniMarkdownDocs"
    return markdown_bundle, pages


def export_wiki(
    pack_dir: Path,
    registry: dict[str, Any],
    adapter: str = "confluence",
    output_dir: Path | None = None,
) -> Path:
    summary = summarise_pack(pack_dir, registry)
    if adapter == "confluence":
        bundle, pages = build_confluence_page_export(summary)
    elif adapter == "markdown":
        bundle, pages = build_markdown_page_export(summary)
    else:
        raise ValueError(
            f"Unsupported adapter {adapter!r}; supported adapters are 'confluence' and 'markdown'"
        )

    if output_dir is None:
        output_dir = pack_dir / "exports" / "wiki" / adapter
    output_dir = output_dir.expanduser()
    output_dir.mkdir(parents=True, exist_ok=True)

    (output_dir / "pages.json").write_text(json.dumps(bundle, indent=2) + "\n", encoding="utf-8")
    index_lines = [
        f"# Wiki Export: {summary.get('pack_id', '')}",
        "",
        f"- WorkUnit: `{summary['work_unit'].get('work_unit_id', '')}`",
        f"- Adapter: `{adapter}`",
        f"- Generated: `{bundle.get('generated_at', '')}`",
        "",
        "## Pages",
    ]
    for filename, body in pages:
        (output_dir / filename).write_text(body, encoding="utf-8")
        index_lines.append(f"- [{filename}](./{filename})")
    (output_dir / "README.md").write_text("\n".join(index_lines).rstrip() + "\n", encoding="utf-8")
    return output_dir


def stable_publish_key(*parts: str) -> str:
    payload = "::".join(part.strip() for part in parts if part)
    return hashlib.sha256(payload.encode("utf-8")).hexdigest()[:16]


def build_rate_limit_policy(adapter: str) -> dict[str, Any]:
    return {
        "adapter": adapter,
        "dispatch_mode": "serialized",
        "max_parallel": 1,
        "retry_strategy": "exponential-backoff",
        "initial_delay_seconds": 2,
        "max_delay_seconds": 60,
        "max_attempts": 5,
        "on_quota_uncertainty": "fallback-to-file-exports",
        "on_rate_limit": "stop-and-preserve-remaining-items",
    }


def runtime_consent_path(pack_dir: Path) -> Path:
    return pack_dir / "runtime" / "consent.json"


def atlassian_targets_path(pack_dir: Path) -> Path:
    return pack_dir / "runtime" / "atlassian-targets.json"


def load_runtime_consent(pack_dir: Path) -> dict[str, bool]:
    path = runtime_consent_path(pack_dir)
    default = {category: False for category in RUNTIME_CONSENT_CATEGORIES}
    if not path.exists():
        return default

    try:
        payload = load_json_file(path)
    except (json.JSONDecodeError, OSError):
        return default

    stored = payload.get("categories", {})
    if not isinstance(stored, dict):
        return default
    return {
        category: bool(stored.get(category, False))
        for category in RUNTIME_CONSENT_CATEGORIES
    }


def persist_runtime_consent(
    pack_dir: Path,
    categories: dict[str, bool],
    granted_by: str = "cli",
) -> Path:
    path = runtime_consent_path(pack_dir)
    path.parent.mkdir(parents=True, exist_ok=True)
    payload = {
        "schema_version": "0.1.0",
        "type": "JiniRuntimeConsent",
        "updated_at": now_utc(),
        "granted_by": granted_by,
        "categories": {
            category: bool(categories.get(category, False))
            for category in RUNTIME_CONSENT_CATEGORIES
        },
    }
    path.write_text(json.dumps(payload, indent=2) + "\n", encoding="utf-8")
    return path


def load_atlassian_targets(pack_dir: Path) -> dict[str, Any] | None:
    path = atlassian_targets_path(pack_dir)
    if not path.exists():
        return None
    try:
        payload = load_json_file(path)
    except (json.JSONDecodeError, OSError):
        return None
    return payload if isinstance(payload, dict) else None


def bind_atlassian_targets(
    pack_dir: Path,
    cloud_id: str,
    site_url: str,
    jira_project_key: str | None = None,
    confluence_space_key: str | None = None,
    confluence_space_id: str | None = None,
    updated_by: str = "cli",
) -> Path:
    if not cloud_id.strip():
        raise ValueError("bind-atlassian requires a non-empty --cloud-id")
    if not site_url.strip():
        raise ValueError("bind-atlassian requires a non-empty --site-url")

    path = atlassian_targets_path(pack_dir)
    path.parent.mkdir(parents=True, exist_ok=True)
    payload = {
        "schema_version": "0.1.0",
        "type": "JiniAtlassianTargets",
        "updated_at": now_utc(),
        "updated_by": updated_by,
        "cloud_id": cloud_id.strip(),
        "site_url": site_url.strip(),
        "jira": {
            "project_key": (jira_project_key or "").strip(),
        },
        "confluence": {
            "space_key": (confluence_space_key or "").strip(),
            "space_id": (confluence_space_id or "").strip(),
        },
    }
    path.write_text(json.dumps(payload, indent=2) + "\n", encoding="utf-8")
    return path


def resolve_atlassian_targets(
    pack_dir: Path,
    *,
    cloud_id: str | None = None,
    site_url: str | None = None,
    project_key: str | None = None,
    space_key: str | None = None,
    space_id: str | None = None,
) -> dict[str, Any]:
    bound = load_atlassian_targets(pack_dir) or {}
    jira_bound = bound.get("jira", {}) if isinstance(bound.get("jira", {}), dict) else {}
    confluence_bound = (
        bound.get("confluence", {}) if isinstance(bound.get("confluence", {}), dict) else {}
    )
    return {
        "cloud_id": (cloud_id or bound.get("cloud_id", "") or "").strip(),
        "site_url": (site_url or bound.get("site_url", "") or "").strip(),
        "project_key": (project_key or jira_bound.get("project_key", "") or "").strip(),
        "space_key": (space_key or confluence_bound.get("space_key", "") or "").strip(),
        "space_id": (space_id or confluence_bound.get("space_id", "") or "").strip(),
        "bound": bool(bound),
        "path": display_path(atlassian_targets_path(pack_dir)) if bound else "",
    }


def append_run_action(
    actions: list[dict[str, Any]],
    *,
    category: str,
    command: str,
    status: str,
    message: str,
    output_path: Path | None = None,
) -> None:
    action = {
        "category": category,
        "command": command,
        "status": status,
        "message": message,
    }
    if output_path is not None:
        action["output_path"] = display_path(output_path)
    actions.append(action)


def load_json_file(path: Path) -> dict[str, Any]:
    return json.loads(path.read_text(encoding="utf-8"))


def publish_issues(
    pack_dir: Path,
    registry: dict[str, Any],
    adapter: str = "jira",
    output_dir: Path | None = None,
    project_key: str | None = None,
    cloud_id: str | None = None,
    site_url: str | None = None,
) -> Path:
    cli = cli_invocation()
    if adapter not in {"jira", "github"}:
        raise ValueError("publish-issues supports only 'jira' and 'github'")
    export_dir = export_issues(pack_dir, registry, adapter=adapter)
    bundle = load_json_file(export_dir / "issues.json")
    work_unit = bundle["work_unit"]
    targets = (
        resolve_atlassian_targets(
            pack_dir,
            cloud_id=cloud_id,
            site_url=site_url,
            project_key=project_key,
        )
        if adapter == "jira"
        else {
            "cloud_id": "",
            "site_url": "",
            "project_key": "",
            "bound": False,
            "path": "",
        }
    )
    publish_root = output_dir.expanduser() if output_dir else pack_dir / "exports" / "publish" / "issues" / adapter
    publish_root.mkdir(parents=True, exist_ok=True)

    payload_dir = publish_root / "payloads"
    payload_dir.mkdir(parents=True, exist_ok=True)
    items: list[dict[str, Any]] = []

    for idx, issue in enumerate(bundle.get("issues", []), start=1):
        issue_key = stable_publish_key(
            adapter,
            bundle.get("pack_id", ""),
            work_unit.get("work_unit_id", ""),
            issue.get("task_id", ""),
            str(issue.get("source_task_index", "")),
        )
        payload = {
            "idempotency_key": issue_key,
            "cloud_id": targets["cloud_id"],
            "site_url": targets["site_url"],
            "project_key": targets["project_key"],
            "summary": issue.get("summary", ""),
            "description_markdown": issue.get("description_markdown", ""),
            "labels": issue.get("labels", []),
            "issue_type_hint": issue.get("issue_type_hint", "Task"),
            "assignee_hint": issue.get("assignee_hint", ""),
            "status_hint": issue.get("status_hint", ""),
            "source_task_id": issue.get("task_id", ""),
        }
        payload_name = f"{idx:02d}-{slugify(issue.get('summary', 'issue'))}.json"
        payload_path = payload_dir / payload_name
        payload_path.write_text(json.dumps(payload, indent=2) + "\n", encoding="utf-8")
        items.append(
            {
                "sequence": idx,
                "task_id": issue.get("task_id", ""),
                "summary": issue.get("summary", ""),
                "idempotency_key": issue_key,
                "payload_path": display_path(payload_path),
            }
        )

    publish_plan = {
        "schema_version": "0.1.0",
        "publish_type": "JiniIssuePublishPlan",
        "generated_at": now_utc(),
        "adapter": adapter,
        "execution_mode": (
            "connector-ready"
            if adapter == "jira" and (targets["cloud_id"] and targets["project_key"])
            else "local-apply" if adapter == "github" else "staged-only"
        ),
        "target": {
            "cloud_id": targets["cloud_id"],
            "site_url": targets["site_url"],
            "project_key": targets["project_key"],
            "configured": bool(targets["cloud_id"] and targets["project_key"]),
            "bound_config_path": targets["path"],
        },
        "work_unit": work_unit,
        "source_bundle": display_path(export_dir / "issues.json"),
        "rate_limit_policy": build_rate_limit_policy(adapter),
        "items": items,
        "notes": [
            (
                "This plan is staged in serialized order so an external Jira publisher can replay it safely."
                if adapter == "jira"
                else "This plan can be applied locally as a markdown issue ledger for portable review."
            ),
            "Payloads are serialized and idempotent so a future publisher can replay safely.",
        ],
    }
    (publish_root / "publish-plan.json").write_text(json.dumps(publish_plan, indent=2) + "\n", encoding="utf-8")

    readme_lines = [
        f"# Issue Publish Plan: {bundle.get('pack_id', '')}",
        "",
        f"- Adapter: `{adapter}`",
        f"- Execution Mode: `{publish_plan['execution_mode']}`",
        f"- WorkUnit: `{work_unit.get('work_unit_id', '')}`",
        f"- Cloud ID: `{targets['cloud_id'] or 'unset'}`",
        f"- Site URL: `{targets['site_url'] or 'unset'}`",
        f"- Project Key: `{targets['project_key'] or 'unset'}`",
        "",
        "## Rate-Limit Policy",
        *[f"- {key}: `{value}`" for key, value in build_rate_limit_policy(adapter).items()],
        "",
        "## Payloads",
        *[f"- [{Path(item['payload_path']).name}](./payloads/{Path(item['payload_path']).name}) -> `{item['task_id']}`" for item in items],
        "",
    ]
    if adapter == "github":
        readme_lines.extend(
            [
                "## Local Apply",
                f"- Use `{cli} apply-publish-plan <publish-dir>` to materialize a local markdown issue set.",
                "",
            ]
        )
    (publish_root / "README.md").write_text("\n".join(readme_lines).rstrip() + "\n", encoding="utf-8")
    return publish_root


def publish_wiki(
    pack_dir: Path,
    registry: dict[str, Any],
    adapter: str = "confluence",
    output_dir: Path | None = None,
    space_key: str | None = None,
    cloud_id: str | None = None,
    site_url: str | None = None,
    space_id: str | None = None,
) -> Path:
    cli = cli_invocation()
    if adapter not in {"confluence", "markdown"}:
        raise ValueError("publish-wiki supports only 'confluence' and 'markdown'")

    export_dir = export_wiki(pack_dir, registry, adapter=adapter)
    markdown_export_dir = export_wiki(pack_dir, registry, adapter="markdown")
    bundle = load_json_file(export_dir / "pages.json")
    work_unit = bundle["work_unit"]
    targets = resolve_atlassian_targets(
        pack_dir,
        cloud_id=cloud_id,
        site_url=site_url,
        space_key=space_key,
        space_id=space_id,
    )
    publish_root = output_dir.expanduser() if output_dir else pack_dir / "exports" / "publish" / "wiki" / adapter
    publish_root.mkdir(parents=True, exist_ok=True)

    payload_dir = publish_root / "payloads"
    payload_dir.mkdir(parents=True, exist_ok=True)
    items: list[dict[str, Any]] = []

    for idx, page in enumerate(bundle.get("pages", []), start=1):
        page_key = stable_publish_key(
            adapter,
            bundle.get("pack_id", ""),
            work_unit.get("work_unit_id", ""),
            page.get("slug", ""),
            page.get("role", ""),
        )
        payload = {
            "idempotency_key": page_key,
            "cloud_id": targets["cloud_id"],
            "site_url": targets["site_url"],
            "space_key": targets["space_key"],
            "space_id": targets["space_id"],
            "title": page.get("title", ""),
            "slug": page.get("slug", ""),
            "role": page.get("role", ""),
            "body_markdown": page.get("body_markdown", ""),
        }
        payload_name = f"{idx:02d}-{slugify(page.get('slug', 'page'))}.json"
        payload_path = payload_dir / payload_name
        payload_path.write_text(json.dumps(payload, indent=2) + "\n", encoding="utf-8")
        items.append(
            {
                "sequence": idx,
                "slug": page.get("slug", ""),
                "title": page.get("title", ""),
                "idempotency_key": page_key,
                "payload_path": display_path(payload_path),
            }
        )

    fallback_active = adapter == "markdown" or not targets["space_key"]
    execution_mode = (
        "local-apply"
        if adapter == "markdown"
        else "markdown-fallback"
        if fallback_active
        else "connector-ready" if targets["cloud_id"] else "staged-only"
    )
    publish_plan = {
        "schema_version": "0.1.0",
        "publish_type": "JiniWikiPublishPlan",
        "generated_at": now_utc(),
        "adapter": adapter,
        "execution_mode": execution_mode,
        "target": {
            "cloud_id": targets["cloud_id"],
            "site_url": targets["site_url"],
            "space_key": targets["space_key"],
            "space_id": targets["space_id"],
            "configured": bool(targets["cloud_id"] and targets["space_key"]),
            "bound_config_path": targets["path"],
        },
        "work_unit": work_unit,
        "source_bundle": display_path(export_dir / "pages.json"),
        "fallback_bundle": display_path(markdown_export_dir / "pages.json"),
        "rate_limit_policy": build_rate_limit_policy(adapter),
        "items": items,
        "notes": [
            (
                "When Confluence is not configured, markdown export is the canonical fallback output."
                if adapter == "confluence"
                else "Markdown publish plans can be applied locally for portable documentation handoff."
            ),
            (
                "This plan is staged in serialized order so an external Confluence publisher can replay it safely."
                if adapter == "confluence"
                else "Payloads are serialized and idempotent so local markdown application stays replay-safe."
            ),
        ],
    }
    (publish_root / "publish-plan.json").write_text(json.dumps(publish_plan, indent=2) + "\n", encoding="utf-8")

    readme_lines = [
        f"# Wiki Publish Plan: {bundle.get('pack_id', '')}",
        "",
        f"- Adapter: `{adapter}`",
        f"- Execution Mode: `{execution_mode}`",
        f"- WorkUnit: `{work_unit.get('work_unit_id', '')}`",
        f"- Cloud ID: `{targets['cloud_id'] or 'unset'}`",
        f"- Site URL: `{targets['site_url'] or 'unset'}`",
        f"- Space Key: `{targets['space_key'] or 'unset'}`",
        f"- Space ID: `{targets['space_id'] or 'unset'}`",
        f"- Markdown Fallback: `{display_path(markdown_export_dir)}`",
        "",
        "## Rate-Limit Policy",
        *[f"- {key}: `{value}`" for key, value in build_rate_limit_policy(adapter).items()],
        "",
        "## Payloads",
        *[f"- [{Path(item['payload_path']).name}](./payloads/{Path(item['payload_path']).name}) -> `{item['slug']}`" for item in items],
        "",
    ]
    if execution_mode in {"local-apply", "markdown-fallback"}:
        readme_lines.extend(
            [
                "## Local Apply",
                f"- Use `{cli} apply-publish-plan <publish-dir>` to materialize local markdown pages.",
                "",
            ]
        )
    (publish_root / "README.md").write_text("\n".join(readme_lines).rstrip() + "\n", encoding="utf-8")
    return publish_root


def resolve_publish_plan_input(path: Path) -> Path:
    candidate = path.expanduser()
    if candidate.is_dir():
        candidate = candidate / "publish-plan.json"
    if not candidate.exists():
        raise ValueError(f"Publish plan not found at {display_path(candidate)}")
    return candidate


def next_publish_apply_receipt_path(plan_dir: Path, adapter: str) -> Path:
    receipt_dir = plan_dir / "applied"
    receipt_dir.mkdir(parents=True, exist_ok=True)
    stamp = datetime.now(timezone.utc).strftime("%Y%m%dT%H%M%S%fZ")
    return receipt_dir / f"apply-{slugify(adapter)}-{stamp}.json"


def render_local_issue_markdown(payload: dict[str, Any]) -> str:
    labels = ", ".join(str(item) for item in payload.get("labels", [])) or "none"
    lines = [
        f"# {payload.get('summary', '')}",
        "",
        f"- Idempotency Key: `{payload.get('idempotency_key', '')}`",
        f"- Source Task: `{payload.get('source_task_id', '')}`",
        f"- Issue Type Hint: `{payload.get('issue_type_hint', '')}`",
        f"- Assignee Hint: `{payload.get('assignee_hint', '') or 'unset'}`",
        f"- Status Hint: `{payload.get('status_hint', '') or 'unset'}`",
        f"- Labels: `{labels}`",
        "",
        "## Description",
        payload.get("description_markdown", "").rstrip(),
        "",
    ]
    return "\n".join(lines)


def render_local_wiki_markdown(payload: dict[str, Any]) -> str:
    title = str(payload.get("title", "")).strip()
    body = str(payload.get("body_markdown", "")).rstrip()
    if body.startswith(f"# {title}"):
        return body + ("\n" if not body.endswith("\n") else "")
    return f"# {title}\n\n{body}\n"


def apply_publish_plan(
    path: Path,
    *,
    output_dir: Path | None = None,
) -> tuple[dict[str, Any], Path]:
    plan_path = resolve_publish_plan_input(path)
    plan = load_json_file(plan_path)
    if not isinstance(plan, dict):
        raise ValueError("Publish plan must be a mapping")
    publish_type = str(plan.get("publish_type", "")).strip()
    adapter = str(plan.get("adapter", "")).strip()
    execution_mode = str(plan.get("execution_mode", "")).strip()
    plan_dir = plan_path.parent
    applied_root = output_dir.expanduser() if output_dir is not None else plan_dir / "applied"
    applied_root.mkdir(parents=True, exist_ok=True)
    applied_paths: list[str] = []
    status = "staged-external"
    notes: list[str] = []

    if publish_type == "JiniIssuePublishPlan" and adapter == "github":
        issues_dir = applied_root / "issues"
        issues_dir.mkdir(parents=True, exist_ok=True)
        for item in plan.get("items", []):
            if not isinstance(item, dict):
                continue
            payload_path = resolve_display_path(str(item.get("payload_path", "")))
            payload = load_json_file(payload_path)
            filename = f"{int(item.get('sequence', 0) or 0):02d}-{slugify(str(item.get('summary', 'issue')))}.md"
            destination = issues_dir / filename
            destination.write_text(render_local_issue_markdown(payload), encoding="utf-8")
            applied_paths.append(str(destination))
        status = "applied-local"
        notes.append("Applied GitHub-flavored issue plan to a local markdown issue ledger.")
    elif publish_type == "JiniWikiPublishPlan" and execution_mode in {"markdown-fallback", "local-apply"}:
        pages_dir = applied_root / "pages"
        pages_dir.mkdir(parents=True, exist_ok=True)
        for item in plan.get("items", []):
            if not isinstance(item, dict):
                continue
            payload_path = resolve_display_path(str(item.get("payload_path", "")))
            payload = load_json_file(payload_path)
            filename = f"{int(item.get('sequence', 0) or 0):02d}-{slugify(str(item.get('slug', 'page')))}.md"
            destination = pages_dir / filename
            destination.write_text(render_local_wiki_markdown(payload), encoding="utf-8")
            applied_paths.append(str(destination))
        status = "applied-local"
        notes.append("Applied wiki publish plan to a local markdown page set.")
    else:
        notes.append(
            f"Adapter `{adapter}` with execution mode `{execution_mode or 'unknown'}` remains staged; no local apply target is available."
        )

    receipt = {
        "schema_version": "0.1.0",
        "receipt_type": "JiniAppliedPublishPlan",
        "generated_at": now_utc(),
        "publish_type": publish_type,
        "adapter": adapter,
        "execution_mode": execution_mode,
        "status": status,
        "plan_path": display_path(plan_path),
        "output_root": str(applied_root),
        "applied_paths": applied_paths,
        "notes": notes,
    }
    receipt_path = next_publish_apply_receipt_path(plan_dir, adapter or "publish")
    receipt["receipt_path"] = display_path(receipt_path)
    receipt_path.write_text(json.dumps(receipt, indent=2) + "\n", encoding="utf-8")
    return receipt, receipt_path


def build_staged_publish_receipt(path: Path) -> dict[str, Any]:
    plan_path = resolve_publish_plan_input(path)
    plan = load_json_file(plan_path)
    if not isinstance(plan, dict):
        raise ValueError("Publish plan must be a mapping")
    adapter = str(plan.get("adapter", "")).strip()
    execution_mode = str(plan.get("execution_mode", "")).strip()
    return {
        "schema_version": "0.1.0",
        "receipt_type": "JiniStagedPublishPlan",
        "generated_at": now_utc(),
        "publish_type": str(plan.get("publish_type", "")).strip(),
        "adapter": adapter,
        "execution_mode": execution_mode,
        "status": "staged",
        "plan_path": display_path(plan_path),
        "output_root": display_path(plan_path.parent),
        "applied_paths": [],
        "notes": [
            (
                "Local apply is available for this adapter; rerun with --apply-local to materialize local outputs."
                if execution_mode in {"local-apply", "markdown-fallback"}
                else "This publish plan remains staged for an external connector or future replay."
            )
        ],
    }


def next_publish_execution_bundle_path(executed_dir: Path, adapter: str) -> Path:
    executed_dir.mkdir(parents=True, exist_ok=True)
    stamp = datetime.now(timezone.utc).strftime("%Y%m%dT%H%M%S%fZ")
    return executed_dir / f"publish-result-{slugify(adapter)}-{stamp}.json"


def next_publish_execution_receipt_path(executed_dir: Path, adapter: str) -> Path:
    executed_dir.mkdir(parents=True, exist_ok=True)
    stamp = datetime.now(timezone.utc).strftime("%Y%m%dT%H%M%S%fZ")
    return executed_dir / f"publish-execution-{slugify(adapter)}-{stamp}.json"


def publish_plan_target_kind(publish_type: str) -> str:
    if publish_type == "JiniIssuePublishPlan":
        return "issue"
    if publish_type == "JiniWikiPublishPlan":
        return "wiki-page"
    return "publication"


def publish_plan_source_ref(item: dict[str, Any], payload_path: Path) -> str:
    for key, prefix in (("task_id", "task"), ("slug", "page"), ("summary", "item"), ("title", "item")):
        value = str(item.get(key, "")).strip()
        if value:
            return f"{prefix}:{value}"
    return display_path(payload_path)


def execute_publish_plan(
    path: Path,
    *,
    runner: Path,
    output_dir: Path | None = None,
    timeout_seconds: int = 15,
) -> tuple[dict[str, Any], Path, Path]:
    plan_path = resolve_publish_plan_input(path)
    plan = load_json_file(plan_path)
    if not isinstance(plan, dict):
        raise ValueError("Publish plan must be a mapping")
    runner_path = runner.expanduser()
    if not runner_path.exists():
        raise ValueError(f"Bridge runner not found at {display_path(runner_path)}")

    publish_type = str(plan.get("publish_type", "")).strip()
    adapter = str(plan.get("adapter", "")).strip()
    execution_mode = str(plan.get("execution_mode", "")).strip()
    items = plan.get("items", [])
    if not isinstance(items, list) or not items:
        raise ValueError("Publish plan must contain at least one item")

    plan_dir = plan_path.parent
    executed_root = output_dir.expanduser() if output_dir is not None else plan_dir / "executed"
    executed_root.mkdir(parents=True, exist_ok=True)
    target_kind = publish_plan_target_kind(publish_type)
    records: list[dict[str, Any]] = []
    failures: list[dict[str, Any]] = []
    notes: list[str] = []

    for item in items:
        if not isinstance(item, dict):
            failures.append({"reason": "Plan item must be an object"})
            break
        payload_path = resolve_display_path(str(item.get("payload_path", "")))
        env = os.environ.copy()
        env["JINI_PLAN_PATH"] = str(plan_path)
        env["JINI_ADAPTER"] = adapter
        env["JINI_PUBLISH_TYPE"] = publish_type
        env["JINI_TARGET_KIND"] = target_kind
        env["JINI_PLAN_ITEM_JSON"] = json.dumps(item)
        try:
            command_result = subprocess.run(
                [str(runner_path), str(payload_path)],
                capture_output=True,
                text=True,
                timeout=timeout_seconds,
                env=env,
            )
        except OSError as exc:
            raise ValueError(f"Bridge runner failed to start: {exc}") from exc
        except subprocess.TimeoutExpired:
            failures.append(
                {
                    "source_ref": publish_plan_source_ref(item, payload_path),
                    "payload_path": display_path(payload_path),
                    "reason": f"Bridge runner timed out after {timeout_seconds}s",
                }
            )
            break
        if command_result.returncode != 0:
            failures.append(
                {
                    "source_ref": publish_plan_source_ref(item, payload_path),
                    "payload_path": display_path(payload_path),
                    "reason": f"Bridge runner exited with status {command_result.returncode}",
                    "stdout": command_result.stdout.strip(),
                    "stderr": command_result.stderr.strip(),
                }
            )
            break
        stdout = command_result.stdout.strip()
        try:
            response = json.loads(stdout)
        except json.JSONDecodeError:
            failures.append(
                {
                    "source_ref": publish_plan_source_ref(item, payload_path),
                    "payload_path": display_path(payload_path),
                    "reason": "Bridge runner output was not valid JSON",
                    "stdout": stdout,
                }
            )
            break
        if not isinstance(response, dict):
            failures.append(
                {
                    "source_ref": publish_plan_source_ref(item, payload_path),
                    "payload_path": display_path(payload_path),
                    "reason": "Bridge runner output must be a JSON object",
                }
            )
            break
        external_id = str(response.get("external_id", "")).strip()
        external_url = str(response.get("external_url", "")).strip()
        publication_status = str(response.get("publication_status", "published")).strip() or "published"
        published_at = str(response.get("published_at", "")).strip() or now_utc()
        item_notes = response.get("notes", [])
        if not external_id or not external_url:
            failures.append(
                {
                    "source_ref": publish_plan_source_ref(item, payload_path),
                    "payload_path": display_path(payload_path),
                    "reason": "Bridge runner output must include external_id and external_url",
                    "stdout": stdout,
                }
            )
            break
        if not isinstance(item_notes, list):
            failures.append(
                {
                    "source_ref": publish_plan_source_ref(item, payload_path),
                    "payload_path": display_path(payload_path),
                    "reason": "Bridge runner notes must be an array",
                    "stdout": stdout,
                }
            )
            break
        record = {
            "adapter": adapter,
            "target_kind": target_kind,
            "source_ref": publish_plan_source_ref(item, payload_path),
            "external_id": external_id,
            "external_url": external_url,
            "published_at": published_at,
            "publication_status": publication_status,
            "notes": [str(note) for note in item_notes],
        }
        records.append(record)

    status = "executed" if not failures else "failed"
    if status == "executed":
        notes.append("Executed staged publish plan through a bridge runner and emitted a portable publication result bundle.")
    else:
        notes.append("Bridge execution stopped before all publish-plan items completed.")

    result_bundle = {
        "schema_version": "0.1.0",
        "result_type": "JiniPublicationResult",
        "generated_at": now_utc(),
        "publish_type": publish_type,
        "adapter": adapter,
        "target_kind": target_kind,
        "plan_path": display_path(plan_path),
        "runner": display_path(runner_path),
        "records": records,
        "failures": failures,
    }
    result_path = next_publish_execution_bundle_path(executed_root, adapter)
    result_path.write_text(json.dumps(result_bundle, indent=2) + "\n", encoding="utf-8")

    receipt = {
        "schema_version": "0.1.0",
        "receipt_type": "JiniExecutedPublishPlan",
        "generated_at": now_utc(),
        "publish_type": publish_type,
        "adapter": adapter,
        "execution_mode": execution_mode,
        "status": status,
        "plan_path": display_path(plan_path),
        "runner": display_path(runner_path),
        "output_root": str(executed_root),
        "result_path": display_path(result_path),
        "records": records,
        "failures": failures,
        "notes": notes,
    }
    receipt_path = next_publish_execution_receipt_path(executed_root, adapter)
    receipt["receipt_path"] = display_path(receipt_path)
    receipt_path.write_text(json.dumps(receipt, indent=2) + "\n", encoding="utf-8")
    return receipt, receipt_path, result_path


def bump_execution_class(level: str) -> str:
    order = ["cheap", "standard", "deep"]
    if level not in order:
        return "standard"
    idx = min(order.index(level) + 1, len(order) - 1)
    return order[idx]


def recommend_execution(
    pack_dir: Path,
    registry: dict[str, Any],
    intent: str | None = None,
    repo_path: Path | None = None,
    home_path: Path | None = None,
    runtime_target: str | None = None,
) -> dict[str, Any]:
    summary = summarise_pack(pack_dir, registry)
    work_unit = summary["work_unit"]
    profile = str(work_unit.get("profile_id", "")).strip()
    state = str(work_unit.get("current_state", "")).strip()
    chosen_intent = (intent or summary.get("next_operation", "Make")).strip().lower()
    active_policy = load_active_policy_rollout(pack_dir, policy_id="runtime-routing")

    base_class_map = {
        "scope": "standard",
        "probe": "standard",
        "model": "standard",
        "decide": "standard",
        "make": "standard",
        "verify": "deep",
        "export": "cheap",
        "issues": "cheap",
        "wiki": "cheap",
        "publish": "standard",
        "research": "standard",
    }
    execution_class = base_class_map.get(chosen_intent, "standard")

    rationale: list[str] = [
        f"Intent `{chosen_intent}` defaults to `{execution_class}` execution",
    ]
    policy_runtime_target = ""

    if active_policy is not None:
        intent_overrides = active_policy.get("intent_overrides", {})
        if isinstance(intent_overrides, dict):
            override = str(intent_overrides.get(chosen_intent, "")).strip()
            if override in {"cheap", "standard", "deep"}:
                execution_class = override
                rationale.append(
                    f"Active routing rollout `{active_policy.get('candidate_id', '')}` overrides intent "
                    f"`{chosen_intent}` to `{execution_class}`"
                )
        policy_runtime_target = str(active_policy.get("recommended_runtime_target", "")).strip()
        if active_policy.get("compact_reload_preferred"):
            rationale.append("Active routing rollout keeps compact reloads as the preferred low-token resume path")
        if active_policy.get("memory_loop_required"):
            rationale.append("Active routing rollout keeps durable home-memory capture in the default execution loop")

    if profile in {"Critical", "Regulated"} and chosen_intent not in {"export", "issues", "wiki"}:
        execution_class = bump_execution_class(execution_class)
        rationale.append(f"Profile `{profile}` increases the execution class")

    if state in {"awaiting_verification", "incident", "operational"} and chosen_intent not in {"export", "issues", "wiki"}:
        execution_class = bump_execution_class(execution_class)
        rationale.append(f"State `{state}` requires stronger verification posture")

    if chosen_intent in {"export", "issues", "wiki"}:
        tool_order = [
            "local artifacts and rendered views",
            "structured JSON/markdown export",
            "external publish only if explicitly requested",
        ]
        context_policy = "summary-first"
        delegation = "no child delegation by default"
    elif chosen_intent in {"scope", "probe", "research"}:
        tool_order = [
            "memory indexes and local docs",
            "targeted text search",
            "bounded external text fetch only when needed",
        ]
        context_policy = "targeted"
        delegation = "at most one scoped helper; child must return if the task needs deeper reasoning"
    elif chosen_intent in {"model", "decide", "make"}:
        tool_order = [
            "local repo/docs and canonical artifacts",
            "structured parsers and export surfaces",
            "external systems only after local state is coherent",
        ]
        context_policy = "targeted"
        delegation = "one specialist or reviewer max; no recursive cheap-worker fan-out"
    else:
        tool_order = [
            "canonical evidence and approval artifacts",
            "local status and export surfaces",
            "external publish or verification endpoints one at a time",
        ]
        context_policy = "broad-but-bounded"
        delegation = "coordinator plus reviewer allowed; max spawn depth 2"

    rate_limit_strategy = [
        "Prefer local rendering and export over live API calls when both satisfy the need",
        "Serialize external publish calls; never burst Jira and Confluence writes in parallel",
        "If adapter availability or quota is uncertain, emit markdown/json bundles and stop before live publish",
        "Do not auto-escalate a child to a more expensive execution class; return to the parent for rerouting",
        "Compact and reload the smallest context slice that satisfies the intent",
    ]

    recommendation = {
        "schema_version": "0.1.0",
        "recommendation_type": "JiniExecutionRecommendation",
        "generated_at": now_utc(),
        "pack_id": summary.get("pack_id", ""),
        "work_unit_id": work_unit.get("work_unit_id", ""),
        "state": state,
        "profile_id": profile,
        "health": summary.get("health", ""),
        "intent": chosen_intent,
        "execution_class": execution_class,
        "context_policy": context_policy,
        "tool_order": tool_order,
        "delegation_policy": delegation,
        "rate_limit_strategy": rate_limit_strategy,
        "rationale": rationale,
        "active_policy": {
            "policy_id": active_policy.get("policy_id", ""),
            "candidate_id": active_policy.get("candidate_id", ""),
            "status": active_policy.get("status", ""),
            "intent_overrides": active_policy.get("intent_overrides", {}),
            "recommended_runtime_target": active_policy.get("recommended_runtime_target", ""),
        }
        if active_policy is not None
        else {},
    }
    repo_context = inspect_repo_context(pack_dir, repo_path=repo_path)
    apply_repo_guidance(recommendation, repo_context)
    home_binding = resolve_home_binding(pack_dir, explicit_home=home_path)
    if home_binding.get("bound"):
        rationale.append(
            f"Personal home is bound at `{display_path(Path(home_binding['home_root']))}` for reusable memory and routines"
        )
        if "bound personal home memory" not in tool_order:
            tool_order.insert(0, "bound personal home memory")
    recommendation["memory_context"] = build_memory_context(
        summary,
        chosen_intent,
        latest_harvest=latest_harvest_report_summary(pack_dir),
        latest_run=latest_run_report_summary(pack_dir),
        home_binding=home_binding,
        repo_context=repo_context,
    )
    runtime_guidance = build_adapter_resolution(
        capability="pack-guidance",
        layer="runtime-target",
        preferred=runtime_target or policy_runtime_target or None,
    )
    recommendation["runtime_guidance"] = runtime_guidance
    rationale.append(
        f"Runtime guidance prefers `{runtime_guidance['selected']['id']}` for portable pack guidance surfaces"
    )
    return recommendation


def materialize_compile_outputs(pack_dir: Path, registry: dict[str, Any]) -> list[str]:
    warnings: list[str] = []
    actions = [
        ("export-tasks", lambda: export_tasks(pack_dir, registry)),
        ("sync-tasks", lambda: sync_tasks(pack_dir, registry)),
        ("export-issues:github", lambda: export_issues(pack_dir, registry, adapter="github")),
        ("export-issues:jira", lambda: export_issues(pack_dir, registry, adapter="jira")),
        ("export-wiki:markdown", lambda: export_wiki(pack_dir, registry, adapter="markdown")),
        ("export-wiki:confluence", lambda: export_wiki(pack_dir, registry, adapter="confluence")),
    ]
    for label, action in actions:
        try:
            action()
        except ValueError as exc:
            warnings.append(f"{label} skipped: {exc}")
    return warnings


def run_pack(
    pack_dir: Path,
    registry: dict[str, Any],
    mode: str = "supervised",
    intent: str | None = None,
    repo_path: Path | None = None,
    home_path: Path | None = None,
    runtime_target: str | None = None,
    consent_grants: list[str] | None = None,
    issue_adapter: str = "jira",
    wiki_adapter: str = "confluence",
    project_key: str | None = None,
    space_key: str | None = None,
) -> tuple[dict[str, Any], Path]:
    summary_before = summarise_pack(pack_dir, registry)
    recommendation = recommend_execution(
        pack_dir,
        registry,
        intent=intent,
        repo_path=repo_path,
        home_path=home_path,
        runtime_target=runtime_target,
    )
    atlassian_targets = resolve_atlassian_targets(
        pack_dir,
        project_key=project_key,
        space_key=space_key,
    )
    home_binding = resolve_home_binding(pack_dir, explicit_home=home_path)
    consent_state = load_runtime_consent(pack_dir)
    newly_granted = sorted(set(consent_grants or []))
    consent_path: Path | None = None
    if newly_granted:
        for category in newly_granted:
            if category not in RUNTIME_CONSENT_CATEGORIES:
                raise ValueError(f"Unsupported consent category {category!r}")
            consent_state[category] = True
        consent_path = persist_runtime_consent(pack_dir, consent_state, granted_by=f"run-pack:{mode}")

    actions: list[dict[str, Any]] = []
    blockers: list[str] = []
    intent_key = recommendation["intent"]
    chosen_wiki_adapter = wiki_adapter
    if chosen_wiki_adapter == "confluence" and not atlassian_targets["space_key"]:
        chosen_wiki_adapter = "markdown"

    if consent_state.get("write", False):
        try:
            tasks_view = export_tasks(pack_dir, registry)
            append_run_action(
                actions,
                category="write",
                command="export-tasks",
                status="executed",
                message="Rendered the task board from canonical task state",
                output_path=tasks_view,
            )
        except ValueError as exc:
            append_run_action(
                actions,
                category="write",
                command="export-tasks",
                status="skipped",
                message=str(exc),
            )

        try:
            sync_path = sync_tasks(pack_dir, registry)
            append_run_action(
                actions,
                category="write",
                command="sync-tasks",
                status="executed",
                message="Exported the neutral task sync payload",
                output_path=sync_path,
            )
        except ValueError as exc:
            append_run_action(
                actions,
                category="write",
                command="sync-tasks",
                status="skipped",
                message=str(exc),
            )

        try:
            issues_path = export_issues(pack_dir, registry, adapter=issue_adapter)
            append_run_action(
                actions,
                category="write",
                command=f"export-issues:{issue_adapter}",
                status="executed",
                message="Rendered issue-tracker bundle from canonical task state",
                output_path=issues_path,
            )
        except ValueError as exc:
            append_run_action(
                actions,
                category="write",
                command=f"export-issues:{issue_adapter}",
                status="skipped",
                message=str(exc),
            )

        try:
            wiki_path = export_wiki(pack_dir, registry, adapter=chosen_wiki_adapter)
            wiki_message = "Rendered wiki bundle from canonical artifacts"
            if wiki_adapter == "confluence" and chosen_wiki_adapter == "markdown":
                wiki_message += " using markdown fallback because no Confluence space key was provided"
            append_run_action(
                actions,
                category="write",
                command=f"export-wiki:{chosen_wiki_adapter}",
                status="executed",
                message=wiki_message,
                output_path=wiki_path,
            )
        except ValueError as exc:
            append_run_action(
                actions,
                category="write",
                command=f"export-wiki:{chosen_wiki_adapter}",
                status="skipped",
                message=str(exc),
            )
    else:
        append_run_action(
            actions,
            category="write",
            command="local-exports",
            status="blocked",
            message="Write consent required before Jini can render or refresh local task/wiki/issue outputs",
        )

    if intent_key in {"publish", "issues", "wiki"}:
        if consent_state.get("publish", False):
            try:
                issues_publish_path = publish_issues(
                    pack_dir,
                    registry,
                    adapter=issue_adapter,
                    project_key=atlassian_targets["project_key"],
                    cloud_id=atlassian_targets["cloud_id"],
                    site_url=atlassian_targets["site_url"],
                )
                append_run_action(
                    actions,
                    category="publish",
                    command=f"publish-issues:{issue_adapter}",
                    status="executed",
                    message=(
                        "Staged a serialized Jira publish plan"
                        if issue_adapter == "jira"
                        else "Staged a portable local issue publish plan"
                    ),
                    output_path=issues_publish_path,
                )
            except ValueError as exc:
                append_run_action(
                    actions,
                    category="publish",
                    command=f"publish-issues:{issue_adapter}",
                    status="skipped",
                    message=str(exc),
                )

            try:
                wiki_publish_path = publish_wiki(
                    pack_dir,
                    registry,
                    adapter=chosen_wiki_adapter,
                    space_key=atlassian_targets["space_key"],
                    cloud_id=atlassian_targets["cloud_id"],
                    site_url=atlassian_targets["site_url"],
                    space_id=atlassian_targets["space_id"],
                )
                wiki_publish_message = "Staged a serialized wiki publish plan"
                if chosen_wiki_adapter == "markdown":
                    wiki_publish_message = "Staged a local markdown wiki publish plan"
                elif wiki_adapter == "confluence" and not atlassian_targets["space_key"]:
                    wiki_publish_message += " with markdown fallback because Confluence is not configured"
                append_run_action(
                    actions,
                    category="publish",
                    command=f"publish-wiki:{chosen_wiki_adapter}",
                    status="executed",
                    message=wiki_publish_message,
                    output_path=wiki_publish_path,
                )
            except ValueError as exc:
                append_run_action(
                    actions,
                    category="publish",
                    command=f"publish-wiki:{chosen_wiki_adapter}",
                    status="skipped",
                    message=str(exc),
                )
        else:
            append_run_action(
                actions,
                category="publish",
                command="publish-bundles",
                status="blocked",
                message="Publish consent required before Jini can stage external Jira or wiki publish plans",
            )

    if mode == "autonomous":
        if consent_state.get("command", False):
            try:
                current_state, target_state = advance_pack_state(pack_dir, registry)
                append_run_action(
                    actions,
                    category="command",
                    command="advance-pack",
                    status="executed",
                    message=f"Advanced the pack from {current_state} to {target_state}",
                )
            except ValueError as exc:
                blockers.append(str(exc))
                append_run_action(
                    actions,
                    category="command",
                    command="advance-pack",
                    status="blocked",
                    message=str(exc),
                )
        else:
            append_run_action(
                actions,
                category="command",
                command="advance-pack",
                status="blocked",
                message="Command consent required before autonomous state transitions are allowed",
            )
    else:
        append_run_action(
            actions,
            category="command",
            command="advance-pack",
            status="planned",
            message="Supervised mode does not auto-advance state; review outputs and run advance-pack explicitly or rerun in autonomous mode",
        )

    summary_after = summarise_pack(pack_dir, registry)
    runtime_dir = pack_dir / "runtime"
    runtime_dir.mkdir(parents=True, exist_ok=True)
    report_path = runtime_dir / "last-run.json"
    report = {
        "schema_version": "0.1.0",
        "report_type": "JiniRunReport",
        "generated_at": now_utc(),
        "mode": mode,
        "intent": intent_key,
        "pack_id": summary_after.get("pack_id", ""),
        "work_unit_id": summary_after["work_unit"].get("work_unit_id", ""),
        "state_before": summary_before["work_unit"].get("current_state", ""),
        "state_after": summary_after["work_unit"].get("current_state", ""),
        "health_before": summary_before.get("health", ""),
        "health_after": summary_after.get("health", ""),
        "recommendation": recommendation,
        "atlassian_targets": atlassian_targets,
        "home_binding": {
            "bound": bool(home_binding.get("bound")),
            "home_root": str(home_binding.get("home_root", "")) if home_binding.get("home_root") else "",
            "source": home_binding.get("source", ""),
            "binding_path": home_binding.get("binding_path", ""),
        },
        "consent": {
            "categories": consent_state,
            "persisted_path": display_path(consent_path) if consent_path else display_path(runtime_consent_path(pack_dir)),
            "newly_granted": newly_granted,
        },
        "actions": actions,
        "blockers": blockers,
        "next_operation_after": summary_after.get("next_operation", ""),
    }
    report["memory_append"] = append_home_observation(
        pack_dir,
        home_binding=home_binding,
        line=(
            f"run-pack for {summary_after['work_unit'].get('title', summary_after.get('pack_id', 'pack'))}: "
            f"intent={intent_key}, mode={mode}, state {summary_before['work_unit'].get('current_state', '')}"
            f"->{summary_after['work_unit'].get('current_state', '')}, actions={len(actions)}, blockers={len(blockers)}."
        ),
    )
    report_path.write_text(json.dumps(report, indent=2) + "\n", encoding="utf-8")
    append_learning_event(
        "run-pack",
        {
            "pack_id": summary_after.get("pack_id", ""),
            "work_unit_id": summary_after["work_unit"].get("work_unit_id", ""),
            "mode": mode,
            "intent": intent_key,
            "state_before": summary_before["work_unit"].get("current_state", ""),
            "state_after": summary_after["work_unit"].get("current_state", ""),
            "health_before": summary_before.get("health", ""),
            "health_after": summary_after.get("health", ""),
            "execution_class": recommendation["execution_class"],
            "action_count": len(actions),
            "blocker_count": len(blockers),
            "home_bound": bool(home_binding.get("bound")),
            "memory_appended": bool(report["memory_append"]["appended"]),
            "runtime_target": recommendation["runtime_guidance"]["selected"]["id"],
        },
        pack_dir=pack_dir,
    )
    return report, report_path


def next_execute_flow_report_path(pack_dir: Path) -> Path:
    flow_dir = pack_dir / "runtime" / "flows"
    flow_dir.mkdir(parents=True, exist_ok=True)
    stamp = datetime.now(timezone.utc).strftime("%Y%m%dT%H%M%S%fZ")
    return flow_dir / f"execute-flow-{stamp}.json"


def execute_flow(
    pack_dir: Path,
    registry: dict[str, Any],
    *,
    mode: str = "supervised",
    intent: str | None = None,
    repo_path: Path | None = None,
    home_path: Path | None = None,
    runtime_target: str | None = None,
    consent_grants: list[str] | None = None,
    issue_adapter: str = "jira",
    wiki_adapter: str = "confluence",
    project_key: str | None = None,
    space_key: str | None = None,
    activate_runtime: bool = False,
    activation_prefix: Path | None = None,
    author_actor_id: str = "jini-flow",
    max_items: int = 5,
    max_chars: int = 900,
) -> tuple[dict[str, Any], Path]:
    recommendation = recommend_execution(
        pack_dir,
        registry,
        intent=intent,
        repo_path=repo_path,
        home_path=home_path,
        runtime_target=runtime_target,
    )
    compact = build_compact_context(
        pack_dir,
        registry,
        intent=recommendation["intent"],
        repo_path=repo_path,
        home_path=home_path,
        runtime_target=runtime_target,
        max_items=max_items,
        max_chars=max_chars,
    )
    checklist = build_execution_checklist(
        pack_dir,
        registry,
        intent=recommendation["intent"],
        repo_path=repo_path,
        home_path=home_path,
        runtime_target=runtime_target,
    )
    repo_map = None
    if recommendation.get("repo_context", {}).get("discovered"):
        repo_map = build_repo_map(pack_dir, repo_path=repo_path or pack_dir)
    handoff, handoff_path = build_runtime_handoff(
        pack_dir,
        registry,
        intent=recommendation["intent"],
        repo_path=repo_path,
        home_path=home_path,
        runtime_target=runtime_target,
        max_items=max_items,
        max_chars=max_chars,
        recommendation=recommendation,
        compact=compact,
        checklist=checklist,
        repo_map=repo_map,
    )

    activation_receipt: dict[str, Any] | None = None
    activation_receipt_path: Path | None = None
    if activate_runtime:
        activation_receipt, activation_receipt_path = activate_runtime_target(
            pack_dir,
            registry,
            handoff_path=handoff_path,
            repo_path=repo_path,
            home_path=home_path,
            runtime_target=runtime_target,
            prefix=activation_prefix,
            max_items=max_items,
            max_chars=max_chars,
        )

    run_report, run_report_path = run_pack(
        pack_dir,
        registry,
        mode=mode,
        intent=recommendation["intent"],
        repo_path=repo_path,
        home_path=home_path,
        runtime_target=runtime_target,
        consent_grants=consent_grants,
        issue_adapter=issue_adapter,
        wiki_adapter=wiki_adapter,
        project_key=project_key,
        space_key=space_key,
    )

    harvest_report: dict[str, Any] | None = None
    harvest_report_path: Path | None = None
    evidence_artifact_path: Path | None = None
    harvest_error = ""
    if repo_path is not None and recommendation["intent"] == "verify":
        try:
            harvest_report, harvest_report_path, evidence_artifact_path = harvest_evidence(
                pack_dir,
                registry,
                author_actor_id=author_actor_id,
                repo_path=repo_path,
                home_path=home_path,
                categories=["test", "verify", "startup", "demo"],
                status="reviewed",
            )
        except ValueError as exc:
            harvest_error = str(exc)

    local_publish_receipts: list[dict[str, Any]] = []
    for action in run_report.get("actions", []):
        if action.get("category") != "publish" or action.get("status") != "executed" or not action.get("output_path"):
            continue
        try:
            receipt, receipt_path = apply_publish_plan(resolve_display_path(str(action["output_path"])))
        except ValueError as exc:
            local_publish_receipts.append(
                {
                    "command": action.get("command", ""),
                    "status": "skipped",
                    "message": str(exc),
                    "receipt_path": "",
                }
            )
        else:
            local_publish_receipts.append(
                {
                    "command": action.get("command", ""),
                    "status": receipt.get("status", ""),
                    "message": "; ".join(receipt.get("notes", [])),
                    "receipt_path": display_path(receipt_path),
                    "output_root": receipt.get("output_root", ""),
                    "applied_paths": receipt.get("applied_paths", []),
                }
            )

    flow_path = next_execute_flow_report_path(pack_dir)
    report = {
        "schema_version": "0.1.0",
        "report_type": "JiniExecuteFlowReport",
        "generated_at": now_utc(),
        "flow_path": display_path(flow_path),
        "pack_id": recommendation.get("pack_id", ""),
        "work_unit_id": recommendation.get("work_unit_id", ""),
        "intent": recommendation["intent"],
        "mode": mode,
        "state": recommendation["state"],
        "execution_class": recommendation["execution_class"],
        "recommendation": recommendation,
        "compact_context": compact,
        "execution_checklist": checklist,
        "repo_map": repo_map,
        "runtime_handoff_path": display_path(handoff_path),
        "runtime_activation": activation_receipt,
        "runtime_activation_path": display_path(activation_receipt_path) if activation_receipt_path is not None else "",
        "run_report_path": display_path(run_report_path),
        "run_report": run_report,
        "harvest_report": harvest_report,
        "harvest_report_path": display_path(harvest_report_path) if harvest_report_path is not None else "",
        "evidence_artifact_path": display_path(evidence_artifact_path) if evidence_artifact_path is not None else "",
        "harvest_error": harvest_error,
        "local_publish_receipts": local_publish_receipts,
        "flow_steps": [
            "recommend-execution",
            "compact-context",
            "execution-checklist",
            "stage-runtime-handoff",
            *(["activate-runtime-target"] if activate_runtime else []),
            "run-pack",
            *(["harvest-evidence"] if harvest_report is not None else []),
            *(["apply-publish-plan"] if local_publish_receipts else []),
        ],
        "token_strategy": {
            "compact_estimated_tokens": compact.get("token_budget", {}).get("estimated_tokens", 0),
            "compact_compression_ratio": compact.get("token_budget", {}).get("compression_ratio", 0.0),
            "reused_context_surfaces": [
                "recommendation",
                "compact-context",
                "execution-checklist",
                "runtime-handoff",
            ],
        },
    }
    flow_path.write_text(json.dumps(report, indent=2) + "\n", encoding="utf-8")
    append_learning_event(
        "execute-flow",
        {
            "pack_id": report["pack_id"],
            "work_unit_id": report["work_unit_id"],
            "intent": report["intent"],
            "mode": mode,
            "execution_class": report["execution_class"],
            "activate_runtime": activate_runtime,
            "harvest_ready": bool(harvest_report and harvest_report.get("readiness") == "ready"),
            "local_publish_receipt_count": len(local_publish_receipts),
            "estimated_tokens": report["token_strategy"]["compact_estimated_tokens"],
            "compression_ratio": report["token_strategy"]["compact_compression_ratio"],
            "runtime_target": recommendation["runtime_guidance"]["selected"]["id"],
        },
        pack_dir=pack_dir,
    )
    return report, flow_path


def print_apply_publish_plan(receipt: dict[str, Any]) -> None:
    print(f"TYPE   {receipt.get('publish_type', '')}")
    print(f"ADAPTER {receipt.get('adapter', '')}")
    print(f"MODE   {receipt.get('execution_mode', '')}")
    print(f"STATUS {receipt.get('status', '')}")
    if receipt.get("receipt_path"):
        print(f"RECEIPT {receipt['receipt_path']}")
    if receipt.get("output_root"):
        print(f"OUTPUT {receipt['output_root']}")
    for path in receipt.get("applied_paths", []):
        print(f"  - {path}")
    for note in receipt.get("notes", []):
        print(f"NOTE   {note}")


def print_execute_publish_plan(receipt: dict[str, Any]) -> None:
    print(f"TYPE   {receipt.get('publish_type', '')}")
    print(f"ADAPTER {receipt.get('adapter', '')}")
    print(f"MODE   {receipt.get('execution_mode', '')}")
    print(f"STATUS {receipt.get('status', '')}")
    if receipt.get("receipt_path"):
        print(f"RECEIPT {receipt['receipt_path']}")
    if receipt.get("result_path"):
        print(f"RESULT {receipt['result_path']}")
    if receipt.get("runner"):
        print(f"RUNNER {receipt['runner']}")
    for record in receipt.get("records", []):
        print(f"  - {record.get('source_ref', '')} -> {record.get('external_id', '')}")
    for failure in receipt.get("failures", []):
        print(f"FAIL   {failure.get('source_ref', '')}: {failure.get('reason', '')}")
    for note in receipt.get("notes", []):
        print(f"NOTE   {note}")


def print_execute_flow(report: dict[str, Any]) -> None:
    print(f"PACK   {report.get('pack_id', '')}")
    print(f"WORK   {report.get('work_unit_id', '')}")
    print(f"INTENT {report.get('intent', '')}")
    print(f"MODE   {report.get('mode', '')}")
    print(f"STATE  {report.get('state', '')}")
    print(f"CLASS  {report.get('execution_class', '')}")
    if report.get("flow_path"):
        print(f"FLOW   {report['flow_path']}")
    if report.get("runtime_handoff_path"):
        print(f"HANDOFF {report['runtime_handoff_path']}")
    if report.get("runtime_activation_path"):
        print(f"ACTIVE {report['runtime_activation_path']}")
    if report.get("run_report_path"):
        print(f"RUN    {report['run_report_path']}")
    token_strategy = report.get("token_strategy", {})
    if token_strategy:
        print(
            f"TOKENS est={token_strategy.get('compact_estimated_tokens', 0)} "
            f"ratio={token_strategy.get('compact_compression_ratio', 0.0)}"
        )
    print("STEPS")
    for step in report.get("flow_steps", []):
        print(f"  - {step}")
    if report.get("local_publish_receipts"):
        print("PUBLISH")
        for item in report["local_publish_receipts"]:
            print(f"  - {item.get('status', '')} {item.get('command', '')}")


def capture_output(
    pack_dir: Path,
    registry: dict[str, Any],
    author_actor_id: str,
    task_index: int,
    task_status: str,
    note: str,
    references: list[str],
    deliverable: str | None = None,
) -> Path:
    summary = summarise_pack(pack_dir, registry)
    latest_by_type = summary["latest_by_type"]
    if "Tasks" not in latest_by_type:
        raise ValueError("capture-output requires a Tasks artifact")

    tasks_path, tasks_doc = latest_by_type["Tasks"]
    tasks = list(tasks_doc.get("tasks", []))
    if not tasks:
        raise ValueError("capture-output requires at least one task")
    if task_index < 1 or task_index > len(tasks):
        raise ValueError(f"task-index must be between 1 and {len(tasks)}")
    if not note.strip():
        raise ValueError("capture-output requires a non-empty --note")

    updated = deepcopy(tasks_doc)
    idx = task_index - 1
    updated["revision"] = int(updated.get("revision", 0)) + 1
    updated["updated_at"] = now_utc()
    updated["author_actor_id"] = author_actor_id

    statuses = list(updated.get("status_per_task", []))
    while len(statuses) < len(tasks):
        statuses.append("pending")
    statuses[idx] = task_status
    updated["status_per_task"] = statuses

    output_notes = list(updated.get("output_notes", []))
    while len(output_notes) < len(tasks):
        output_notes.append("")
    output_notes[idx] = note
    updated["output_notes"] = output_notes

    output_refs = list(updated.get("output_refs", []))
    while len(output_refs) < len(tasks):
        output_refs.append("")
    output_refs[idx] = ", ".join(ref for ref in references if ref)
    updated["output_refs"] = output_refs

    if deliverable is not None:
        deliverables = list(updated.get("deliverables", []))
        while len(deliverables) < len(tasks):
            deliverables.append("")
        deliverables[idx] = deliverable
        updated["deliverables"] = deliverables

    dump_document(tasks_path, updated)
    export_tasks(pack_dir, registry)
    return tasks_path


def print_pack_status(summary: dict[str, Any]) -> None:
    work_unit = summary["work_unit"]
    print(f"PACK   {summary['pack_id'] or '(unresolved)'}")
    print(f"PATH   {display_path(summary['pack_dir'])}")
    if summary["manifest_path"] is not None:
        print(f"PACKY  {display_path(summary['manifest_path'])}")
    print(f"WORK   {work_unit.get('work_unit_id', '')}")
    print(f"TITLE  {work_unit.get('title', '')}")
    print(f"HEALTH {summary['health']}")
    print(f"STATE  {work_unit.get('current_state', '')}")
    print(f"NEXT   {summary['next_operation']}")
    print(f"PROF   {work_unit.get('profile_id', '')}")
    print(
        f"VALID  {len(summary['validation_errors'])} error(s), "
        f"{len(summary['validation_warnings'])} warning(s)"
    )
    if summary["compiled_flow"]:
        print(f"FLOW   {' -> '.join(summary['compiled_flow'])}")
    if summary["control_packs"]:
        print(f"CTRL   {', '.join(summary['control_packs'])}")

    stage_required_total = len(summary["stage_required_artifacts"])
    stage_required_present = stage_required_total - len(summary["missing_stage_required"])
    ready_stage_required = len(summary["ready_stage_required"])
    full_required_total = len(summary["full_required_artifacts"])
    full_required_present = full_required_total - len(summary["missing_full_required"])
    print("ARTS")
    if stage_required_total:
        print(
            f"  stage present:    {stage_required_present}/{stage_required_total}"
        )
        print(
            f"  stage ready:      {ready_stage_required}/{stage_required_total}"
        )
    if full_required_total:
        print(
            f"  full pack present:{full_required_present}/{full_required_total}"
        )
    if not stage_required_total and not full_required_total:
        print(f"  present types:    {len(summary['present_types'])}")

    for artifact_type in sorted(summary["present_types"]):
        _, artifact_doc = summary["latest_by_type"][artifact_type]
        print(
            f"  - {artifact_type}: status={artifact_doc.get('status')} "
            f"revision={artifact_doc.get('revision')}"
        )

    if summary["missing_stage_required"]:
        print("MISSING-NOW")
        for artifact_type in summary["missing_stage_required"]:
            print(f"  - {artifact_type}")
    future_missing = [
        artifact_type
        for artifact_type in summary["missing_full_required"]
        if artifact_type not in summary["missing_stage_required"]
    ]
    if future_missing:
        print("MISSING-LATER")
        for artifact_type in future_missing:
            print(f"  - {artifact_type}")

    task_summary = summary["task_summary"]
    if task_summary["total"]:
        print("TASKS")
        print(f"  done:       {task_summary['done']}/{task_summary['total']}")
        print(f"  unresolved: {task_summary['unresolved']}/{task_summary['total']}")
        if task_summary["counts"]:
            counts_text = ", ".join(
                f"{status}={count}" for status, count in sorted(task_summary["counts"].items())
            )
            print(f"  statuses:   {counts_text}")
        if task_summary["blocked_by"]:
            print("  blocked_by:")
            for blocker in task_summary["blocked_by"]:
                print(f"    - {blocker}")

    if summary["evidence_doc"]:
        evidence_doc = summary["evidence_doc"]
        print("EVIDENCE")
        print(
            f"  target: {evidence_doc.get('target_artifact_id')} "
            f"r{evidence_doc.get('target_revision')}"
        )
        print(f"  claims: {len(evidence_doc.get('claims_validated', []))}")
        print(f"  risks:  {len(evidence_doc.get('residual_risks', []))}")

    if summary["validation_warnings"]:
        print("WARNINGS")
        for warning in summary["validation_warnings"]:
            print(f"  - {warning}")


def build_outcome_view(
    pack_dir: Path,
    registry: dict[str, Any],
    *,
    repo_path: Path | None = None,
) -> dict[str, Any]:
    summary = summarise_pack(pack_dir, registry)
    work_unit = summary["work_unit"]
    task_summary = summary["task_summary"]
    missing_now = list(summary["missing_stage_required"])
    missing_later = [
        artifact_type
        for artifact_type in summary["missing_full_required"]
        if artifact_type not in summary["missing_stage_required"]
    ]
    next_command = f"{cli_invocation()} next {display_path(pack_dir)}"
    resume_command = f"{cli_invocation()} resume {display_path(pack_dir)}"
    if repo_path is not None:
        repo_display = display_path(repo_path)
        next_command = f"{next_command} --repo {repo_display}"
        resume_command = f"{resume_command} --repo {repo_display}"
    next_command = f"{next_command} --intent {summary['next_operation'].lower()}"
    resume_command = f"{resume_command} --intent {summary['next_operation'].lower()} --max-chars 900"
    return {
        "schema_version": "0.1.0",
        "view_type": "JiniOutcomeView",
        "pack_id": summary["pack_id"],
        "pack_dir": display_path(pack_dir),
        "work_unit_id": str(work_unit.get("work_unit_id", "")),
        "title": str(work_unit.get("title", "")),
        "health": summary["health"],
        "state": str(work_unit.get("current_state", "")),
        "next_operation": summary["next_operation"],
        "task_summary": {
            "done": int(task_summary.get("done", 0) or 0),
            "total": int(task_summary.get("total", 0) or 0),
            "unresolved": int(task_summary.get("unresolved", 0) or 0),
        },
        "questions": {
            "what_is_done": (
                f"{task_summary.get('done', 0)}/{task_summary.get('total', 0)} tasks completed"
                if task_summary.get("total", 0)
                else "No task list has been captured yet"
            ),
            "what_happens_next": summary["next_operation"],
            "what_is_still_missing_now": missing_now,
            "what_is_still_missing_later": missing_later,
        },
        "continue_with": [next_command, resume_command],
        "validation_errors": list(summary["validation_errors"]),
        "validation_warnings": list(summary["validation_warnings"]),
    }


def print_outcome_view(report: dict[str, Any]) -> None:
    print(f"WORK   {report.get('work_unit_id', '')}")
    print(f"TITLE  {report.get('title', '')}")
    print(f"HEALTH {report.get('health', '')}")
    print(f"STATE  {report.get('state', '')}")
    print()
    print("WHAT IS DONE?")
    print(f"  {report.get('questions', {}).get('what_is_done', '')}")
    print()
    print("WHAT HAPPENS NEXT?")
    print(f"  {report.get('questions', {}).get('what_happens_next', '')}")
    missing_now = report.get("questions", {}).get("what_is_still_missing_now", [])
    missing_later = report.get("questions", {}).get("what_is_still_missing_later", [])
    print()
    print("WHAT IS STILL MISSING NOW?")
    if missing_now:
        for item in missing_now:
            print(f"  - {item}")
    else:
        print("  Nothing stage-critical is missing right now.")
    print()
    print("WHAT IS STILL MISSING LATER?")
    if missing_later:
        for item in missing_later:
            print(f"  - {item}")
    else:
        print("  No future-required artifact gaps are visible right now.")
    if report.get("validation_errors"):
        print()
        print("FIX FIRST")
        for item in report.get("validation_errors", []):
            print(f"  - {item}")
    print()
    print("CONTINUE")
    for command in report.get("continue_with", []):
        print(f"  {command}")


def format_pack_surface_error(pack_path: Path, exc: Exception) -> str:
    if isinstance(exc, FileNotFoundError):
        return f"Pack path is missing required Jini files: {display_path(pack_path)}"
    return str(exc)

    if summary["blockers"]:
        print("BLOCKERS")
        for blocker in summary["blockers"]:
            print(f"  - {blocker}")


def artifact_schema_for_type(registry: dict[str, Any], artifact_type: str) -> Path:
    for canonical, meta in registry["artifacts"].items():
        if artifact_type == canonical:
            return SCHEMA_ROOT / meta["schema"]
    raise KeyError(f"Unknown artifact_type {artifact_type!r}")


def validate_file(path: Path, registry: dict[str, Any], explicit_schema: str | None = None) -> tuple[list[str], list[str]]:
    doc = load_document(path)
    warnings: list[str] = []

    if explicit_schema == "work-unit" or path.name == "work-unit.yaml":
        schema = json.loads((SCHEMA_ROOT / registry["work_unit"]["schema"]).read_text(encoding="utf-8"))
        return validate(doc, schema), warnings

    artifact_type = doc.get("artifact_type")
    if not artifact_type:
        return [f"{path}: missing artifact_type and no explicit schema supplied"], warnings

    schema_path = artifact_schema_for_type(registry, artifact_type)
    schema = json.loads(schema_path.read_text(encoding="utf-8"))
    return validate(doc, schema), warnings


def validate_pack(pack_dir: Path, registry: dict[str, Any]) -> tuple[list[str], list[str]]:
    errors: list[str] = []
    warnings: list[str] = []

    work_unit_path = pack_dir / "work-unit.yaml"
    if not work_unit_path.exists():
        return [f"{pack_dir}: missing work-unit.yaml"], warnings

    work_unit_doc = load_document(work_unit_path)
    work_unit_errors, work_unit_warnings = validate_file(work_unit_path, registry, explicit_schema="work-unit")
    errors.extend(work_unit_errors)
    warnings.extend(work_unit_warnings)

    artifact_dir = pack_dir / "artifacts"
    if not artifact_dir.exists():
        return errors + [f"{pack_dir}: missing artifacts/ directory"], warnings

    artifact_ids: set[str] = set()
    artifact_paths = sorted(artifact_dir.glob("*.y*ml")) + sorted(artifact_dir.glob("*.json"))
    if not artifact_paths:
        errors.append(f"{artifact_dir}: no artifacts found")
        return errors, warnings

    for artifact_path in artifact_paths:
        artifact_doc = load_document(artifact_path)
        artifact_errors, artifact_warnings = validate_file(artifact_path, registry)
        errors.extend(artifact_errors)
        warnings.extend(artifact_warnings)

        artifact_id = artifact_doc.get("artifact_id")
        if artifact_id:
            if artifact_id in artifact_ids:
                errors.append(f"{artifact_path}: duplicate artifact_id {artifact_id!r}")
            artifact_ids.add(artifact_id)

        if artifact_doc.get("work_unit_id") != work_unit_doc.get("work_unit_id"):
            errors.append(
                f"{artifact_path}: work_unit_id {artifact_doc.get('work_unit_id')!r} "
                f"does not match pack work_unit_id {work_unit_doc.get('work_unit_id')!r}"
            )

        if artifact_doc.get("branch_id") != work_unit_doc.get("branch_id"):
            errors.append(
                f"{artifact_path}: branch_id {artifact_doc.get('branch_id')!r} "
                f"does not match pack branch_id {work_unit_doc.get('branch_id')!r}"
            )

    return errors, warnings


def now_utc() -> str:
    return datetime.now(timezone.utc).replace(microsecond=0).strftime("%Y-%m-%dT%H:%M:%SZ")


def write_initial_pack(
    pack_id: str,
    output_dir: Path,
    work_unit_id: str,
    title: str,
    purpose: str,
    owner_actor_id: str,
    approver_actor_ids: list[str],
    stakeholder_actor_ids: list[str],
    branch_id: str,
) -> Path:
    pack_map_index = pack_map()
    if pack_id not in pack_map_index:
        raise KeyError(f"Unknown pack_id {pack_id!r}")

    pack_dir, manifest = pack_map_index[pack_id]
    target_dir = output_dir.expanduser()
    if target_dir.exists():
        raise FileExistsError(f"Output directory already exists: {target_dir}")

    target_artifact_dir = target_dir / "artifacts"
    target_artifact_dir.mkdir(parents=True, exist_ok=False)

    timestamp = now_utc()
    profile = manifest["target_profile"]
    extensions = manifest.get("extensions", [])

    work_unit = {
        "work_unit_id": work_unit_id,
        "title": title,
        "purpose": purpose,
        "current_state": "scoped",
        "profile_id": profile,
        "active_extensions": extensions,
        "branch_id": branch_id,
        "parent_work_unit_id": "",
        "owner_actor_id": owner_actor_id,
        "approver_actor_ids": approver_actor_ids,
        "operator_actor_id": "",
        "rollback_authority_actor_id": "",
        "service_owner_actor_id": "",
        "stakeholder_actor_ids": stakeholder_actor_ids,
        "created_at": timestamp,
        "updated_at": timestamp,
    }

    brief = {
        "artifact_id": f"brief-{work_unit_id}",
        "artifact_type": "Brief",
        "schema_version": "0.1.0",
        "work_unit_id": work_unit_id,
        "branch_id": branch_id,
        "revision": 1,
        "status": "draft",
        "author_actor_id": owner_actor_id,
        "approver_actor_ids": approver_actor_ids,
        "created_at": timestamp,
        "updated_at": timestamp,
        "supersedes": "",
        "references": [f"packs/{pack_id}/context/benchmark-context.md"],
        "change_class": "semantic",
        "objective": purpose,
        "stakeholders": stakeholder_actor_ids or [owner_actor_id],
        "constraints": [
            "Replace with concrete environment and runtime constraints",
            f"Pack baseline: {pack_id}"
        ],
        "success_criteria": [
            "Replace with measurable success criteria",
            "Validate the generated pack before execution"
        ],
        "non_goals": [],
        "scope_summary": title,
    }

    assumptions = {
        "artifact_id": f"assumptions-{work_unit_id}",
        "artifact_type": "Assumptions",
        "schema_version": "0.1.0",
        "work_unit_id": work_unit_id,
        "branch_id": branch_id,
        "revision": 1,
        "status": "draft",
        "author_actor_id": owner_actor_id,
        "approver_actor_ids": approver_actor_ids,
        "created_at": timestamp,
        "updated_at": timestamp,
        "supersedes": "",
        "references": [f"brief-{work_unit_id}"],
        "change_class": "semantic",
        "assumptions": [
            "Replace with the first validated assumption for this work unit"
        ],
        "known_unknowns": [
            "Replace with the first key unknown"
        ],
        "validation_plan": [
            "Replace with the first concrete validation step"
        ],
        "deferred_questions": []
    }

    pack_instance = {
        "pack_id": pack_id,
        "pack_version": manifest["version"],
        "compiled_flow": manifest.get("compiled_flow", []),
        "target_profile": profile,
        "extensions": extensions,
        "control_packs": manifest.get("control_packs", []),
        "work_unit_id": work_unit_id,
        "created_at": timestamp
    }

    dump_document(target_dir / "work-unit.yaml", work_unit)
    dump_document(target_dir / "pack-instance.yaml", pack_instance)
    dump_document(target_artifact_dir / "01-brief.yaml", brief)
    dump_document(target_artifact_dir / "02-assumptions.yaml", assumptions)
    return target_dir


def base_artifact(
    artifact_id: str,
    artifact_type: str,
    work_unit_id: str,
    branch_id: str,
    author_actor_id: str,
    approver_actor_ids: list[str],
    timestamp: str,
    status: str = "reviewed",
    references: list[str] | None = None,
) -> dict[str, Any]:
    return {
        "artifact_id": artifact_id,
        "artifact_type": artifact_type,
        "schema_version": "0.1.0",
        "work_unit_id": work_unit_id,
        "branch_id": branch_id,
        "revision": 1,
        "status": status,
        "author_actor_id": author_actor_id,
        "approver_actor_ids": approver_actor_ids,
        "created_at": timestamp,
        "updated_at": timestamp,
        "supersedes": "",
        "references": references or [],
        "change_class": "semantic",
    }


def write_prd_view(
    target_dir: Path,
    title: str,
    brief: dict[str, Any],
    literature: dict[str, Any],
    method: dict[str, Any],
    spec: dict[str, Any],
    plan: dict[str, Any],
) -> None:
    view_dir = target_dir / "views"
    view_dir.mkdir(parents=True, exist_ok=True)
    lines = [
        f"# PRD: {title}",
        "",
        "## Problem",
        brief["objective"],
        "",
        "## Scope Summary",
        brief["scope_summary"],
        "",
        "## Stakeholders",
        *[f"- {item}" for item in brief["stakeholders"]],
        "",
        "## Success Criteria",
        *[f"- {item}" for item in brief["success_criteria"]],
        "",
        "## Evidence Highlights",
        *[f"- {item}" for item in literature["key_findings"]],
        "",
        "## Research Method",
        method["design"],
        "",
        "## Requirements",
        *[f"- {item}" for item in spec["requirements"]],
        "",
        "## Acceptance Criteria",
        *[f"- {item}" for item in spec["acceptance_criteria"]],
        "",
        "## Delivery Slices",
        *[f"- {item}" for item in plan["slices"]],
        "",
    ]
    (view_dir / "prd.md").write_text("\n".join(lines), encoding="utf-8")


def write_itinerary_view(
    target_dir: Path,
    title: str,
    brief: dict[str, Any],
    spec: dict[str, Any],
    plan: dict[str, Any],
    tasks: dict[str, Any],
) -> None:
    view_dir = target_dir / "views"
    view_dir.mkdir(parents=True, exist_ok=True)
    lines = [
        f"# Itinerary: {title}",
        "",
        "## Objective",
        brief["objective"],
        "",
        "## Scope Summary",
        brief["scope_summary"],
        "",
        "## Constraints",
        *[f"- {item}" for item in brief["constraints"]],
        "",
        "## Requirements",
        *[f"- {item}" for item in spec["requirements"]],
        "",
        "## Delivery Slices",
        *[f"- {item}" for item in plan["slices"]],
        "",
        "## Checklist",
        *[f"- {item}" for item in tasks["tasks"]],
        "",
    ]
    (view_dir / "itinerary.md").write_text("\n".join(lines), encoding="utf-8")


def write_budget_view(
    target_dir: Path,
    title: str,
    brief: dict[str, Any],
    assumptions: dict[str, Any],
    spec: dict[str, Any],
    plan: dict[str, Any],
    tasks: dict[str, Any],
) -> None:
    view_dir = target_dir / "views"
    view_dir.mkdir(parents=True, exist_ok=True)
    lines = [
        f"# Budget Plan: {title}",
        "",
        "## Objective",
        brief["objective"],
        "",
        "## Scope Summary",
        brief["scope_summary"],
        "",
        "## Constraints",
        *[f"- {item}" for item in brief["constraints"]],
        "",
        "## Key Assumptions",
        *[f"- {item}" for item in assumptions["assumptions"]],
        "",
        "## Requirements",
        *[f"- {item}" for item in spec["requirements"]],
        "",
        "## Phases",
        *[f"- {item}" for item in plan["slices"]],
        "",
        "## Checklist",
        *[f"- {item}" for item in tasks["tasks"]],
        "",
    ]
    (view_dir / "budget.md").write_text("\n".join(lines), encoding="utf-8")


def write_response_view(
    target_dir: Path,
    title: str,
    brief: dict[str, Any],
    assumptions: dict[str, Any],
    spec: dict[str, Any],
    plan: dict[str, Any],
    tasks: dict[str, Any],
) -> None:
    view_dir = target_dir / "views"
    view_dir.mkdir(parents=True, exist_ok=True)
    lines = [
        f"# Incident Response Plan: {title}",
        "",
        "## Objective",
        brief["objective"],
        "",
        "## Scope Summary",
        brief["scope_summary"],
        "",
        "## Constraints",
        *[f"- {item}" for item in brief["constraints"]],
        "",
        "## Key Assumptions",
        *[f"- {item}" for item in assumptions["assumptions"]],
        "",
        "## Requirements",
        *[f"- {item}" for item in spec["requirements"]],
        "",
        "## Response Phases",
        *[f"- {item}" for item in plan["slices"]],
        "",
        "## Checklist",
        *[f"- {item}" for item in tasks["tasks"]],
        "",
    ]
    (view_dir / "response.md").write_text("\n".join(lines), encoding="utf-8")


def write_audit_view(
    target_dir: Path,
    title: str,
    brief: dict[str, Any],
    assumptions: dict[str, Any],
    spec: dict[str, Any],
    plan: dict[str, Any],
    tasks: dict[str, Any],
) -> None:
    view_dir = target_dir / "views"
    view_dir.mkdir(parents=True, exist_ok=True)
    lines = [
        f"# Compliance Audit Plan: {title}",
        "",
        "## Objective",
        brief["objective"],
        "",
        "## Scope Summary",
        brief["scope_summary"],
        "",
        "## Constraints",
        *[f"- {item}" for item in brief["constraints"]],
        "",
        "## Key Assumptions",
        *[f"- {item}" for item in assumptions["assumptions"]],
        "",
        "## Requirements",
        *[f"- {item}" for item in spec["requirements"]],
        "",
        "## Audit Phases",
        *[f"- {item}" for item in plan["slices"]],
        "",
        "## Checklist",
        *[f"- {item}" for item in tasks["tasks"]],
        "",
    ]
    (view_dir / "audit.md").write_text("\n".join(lines), encoding="utf-8")


def write_selection_view(
    target_dir: Path,
    title: str,
    brief: dict[str, Any],
    assumptions: dict[str, Any],
    spec: dict[str, Any],
    plan: dict[str, Any],
    tasks: dict[str, Any],
) -> None:
    view_dir = target_dir / "views"
    view_dir.mkdir(parents=True, exist_ok=True)
    lines = [
        f"# Vendor Selection Plan: {title}",
        "",
        "## Objective",
        brief["objective"],
        "",
        "## Scope Summary",
        brief["scope_summary"],
        "",
        "## Constraints",
        *[f"- {item}" for item in brief["constraints"]],
        "",
        "## Key Assumptions",
        *[f"- {item}" for item in assumptions["assumptions"]],
        "",
        "## Evaluation Criteria",
        *[f"- {item}" for item in spec["requirements"]],
        "",
        "## Decision Phases",
        *[f"- {item}" for item in plan["slices"]],
        "",
        "## Checklist",
        *[f"- {item}" for item in tasks["tasks"]],
        "",
    ]
    (view_dir / "selection.md").write_text("\n".join(lines), encoding="utf-8")


def write_followup_view(
    target_dir: Path,
    title: str,
    brief: dict[str, Any],
    assumptions: dict[str, Any],
    spec: dict[str, Any],
    plan: dict[str, Any],
    tasks: dict[str, Any],
) -> None:
    view_dir = target_dir / "views"
    view_dir.mkdir(parents=True, exist_ok=True)
    lines = [
        f"# Meeting Follow-up: {title}",
        "",
        "## Objective",
        brief["objective"],
        "",
        "## Scope Summary",
        brief["scope_summary"],
        "",
        "## Constraints",
        *[f"- {item}" for item in brief["constraints"]],
        "",
        "## Key Assumptions",
        *[f"- {item}" for item in assumptions["assumptions"]],
        "",
        "## Follow-up Requirements",
        *[f"- {item}" for item in spec["requirements"]],
        "",
        "## Follow-up Phases",
        *[f"- {item}" for item in plan["slices"]],
        "",
        "## Checklist",
        *[f"- {item}" for item in tasks["tasks"]],
        "",
    ]
    (view_dir / "followup.md").write_text("\n".join(lines), encoding="utf-8")


def write_compiled_research_prd_pack(
    manifest: dict[str, Any],
    pack_id: str,
    output_dir: Path,
    work_unit_id: str,
    title: str,
    purpose: str,
    owner_actor_id: str,
    approver_actor_ids: list[str],
    stakeholder_actor_ids: list[str],
    branch_id: str,
    context_reference: str,
) -> Path:
    target_dir = output_dir.expanduser()
    if target_dir.exists():
        raise FileExistsError(f"Output directory already exists: {target_dir}")

    artifact_dir = target_dir / "artifacts"
    artifact_dir.mkdir(parents=True, exist_ok=False)
    timestamp = now_utc()

    work_unit = {
        "work_unit_id": work_unit_id,
        "title": title,
        "purpose": purpose,
        "current_state": "decided",
        "profile_id": manifest["target_profile"],
        "active_extensions": manifest.get("extensions", []),
        "branch_id": branch_id,
        "parent_work_unit_id": "",
        "owner_actor_id": owner_actor_id,
        "approver_actor_ids": approver_actor_ids,
        "operator_actor_id": "",
        "rollback_authority_actor_id": "",
        "service_owner_actor_id": "",
        "stakeholder_actor_ids": stakeholder_actor_ids,
        "created_at": timestamp,
        "updated_at": timestamp,
    }

    pack_instance = {
        "pack_id": pack_id,
        "pack_version": manifest["version"],
        "compiled_flow": manifest.get("compiled_flow", []),
        "target_profile": manifest["target_profile"],
        "extensions": manifest.get("extensions", []),
        "control_packs": manifest.get("control_packs", []),
        "work_unit_id": work_unit_id,
        "compiled_at": timestamp,
        "compiled_from_context": context_reference,
    }

    brief = base_artifact(
        f"brief-{work_unit_id}",
        "Brief",
        work_unit_id,
        branch_id,
        owner_actor_id,
        approver_actor_ids,
        timestamp,
        status="approved",
        references=[context_reference],
    )
    brief.update(
        {
            "objective": purpose,
            "stakeholders": stakeholder_actor_ids,
            "constraints": [
                "Research must stay traceable to named sources",
                "The PRD view must be derived from canonical artifacts",
                "Spec and tasks must only claim what research actually supports",
            ],
            "success_criteria": list(manifest.get("success_checks", [])),
            "non_goals": [
                "Building production code in this pack",
                "Deciding beyond the validated research surface",
            ],
            "scope_summary": (
                "Turn validated research into a product-ready brief, a rendered PRD, and a build-ready "
                "spec/plan/task handoff."
            ),
        }
    )

    assumptions = base_artifact(
        f"assumptions-{work_unit_id}",
        "Assumptions",
        work_unit_id,
        branch_id,
        owner_actor_id,
        approver_actor_ids,
        timestamp,
        status="reviewed",
        references=[brief["artifact_id"]],
    )
    assumptions.update(
        {
            "assumptions": [
                "The target buyer prefers a guided workflow over raw flexibility",
                "Research findings are strong enough to justify a first product slice",
                "The initial PRD should optimize for clarity and execution speed, not full market coverage",
            ],
            "known_unknowns": [
                "What level of configurability buyers will demand after the first release",
                "Which proof point most strongly drives internal adoption",
            ],
            "validation_plan": [
                "Cross-check source consistency before finalizing the PRD view",
                "Verify that each build task traces back to a research-backed requirement",
            ],
            "deferred_questions": [
                "Whether a regulated profile should be the default in later iterations",
            ],
        }
    )

    sources = base_artifact(
        f"sources-{work_unit_id}",
        "Sources",
        work_unit_id,
        branch_id,
        "research-lead",
        approver_actor_ids,
        timestamp,
        status="reviewed",
        references=[context_reference],
    )
    sources.update(
        {
            "source_entries": [
                "12 semi-structured customer interviews",
                "Quarterly support ticket review",
                "Usage analytics export",
                "Competitive notes across 4 adjacent products",
            ],
            "credibility_notes": [
                "Interview sample spans admins, operators, and exec sponsors",
                "Support data is recent but biased toward active accounts",
                "Competitive notes are directional, not exhaustive",
            ],
            "coverage_gaps": [
                "Limited data from customers who churned early",
                "No in-person shadowing for this iteration",
            ],
            "refresh_expectations": [
                "Revisit source set before major pricing or packaging changes",
            ],
        }
    )

    literature = base_artifact(
        f"literature-{work_unit_id}",
        "Literature",
        work_unit_id,
        branch_id,
        "research-lead",
        approver_actor_ids,
        timestamp,
        status="reviewed",
        references=[sources["artifact_id"]],
    )
    literature.update(
        {
            "research_question": (
                "Which workflow should Jini prioritize so research reliably turns into product direction "
                "and build-ready execution?"
            ),
            "sources": sources["source_entries"],
            "key_findings": [
                "Teams lose momentum when research, PRD, and engineering planning live in separate systems",
                "Buyers want traceability from claims to supporting evidence before committing engineering time",
                "A compact, opinionated default flow is preferred over a large menu of optional process steps",
                "The highest-value bridge is research synthesis into a build-ready PRD and spec handoff",
            ],
            "gaps": [
                "No strong evidence yet on the best packaging for non-software teams",
                "Limited signal on how much customization enterprises need at launch",
            ],
            "relevance_map": [
                "Finding 1 -> product needs a single spine from research to execution",
                "Finding 2 -> evidence must be first-class",
                "Finding 3 -> Jini should keep the surface minimal",
                "Finding 4 -> research-prd is the right flagship pack",
            ],
        }
    )

    method = base_artifact(
        f"method-{work_unit_id}",
        "Method",
        work_unit_id,
        branch_id,
        "research-lead",
        approver_actor_ids,
        timestamp,
        status="reviewed",
        references=[sources["artifact_id"]],
    )
    method.update(
        {
            "design": "Mixed-method product discovery synthesis",
            "data_inputs": [
                "Interview notes",
                "Support ticket themes",
                "Usage metrics",
                "Competitive comparisons",
            ],
            "steps": [
                "Normalize source notes into a shared problem vocabulary",
                "Triangulate repeated pain points across source types",
                "Translate validated findings into product and delivery implications",
                "Render PRD and build handoff from the canonical artifacts",
            ],
            "validity_risks": [
                "Interview sample may over-represent current power users",
                "Analytics alone cannot explain user motivation",
            ],
            "reproducibility_notes": [
                "Each PRD section references the canonical artifacts used to derive it",
            ],
        }
    )

    spec = base_artifact(
        f"spec-{work_unit_id}",
        "Spec",
        work_unit_id,
        branch_id,
        "product-architect",
        approver_actor_ids,
        timestamp,
        status="approved",
        references=[brief["artifact_id"], literature["artifact_id"], method["artifact_id"]],
    )
    spec.update(
        {
            "requirements": [
                "Research findings must compile into a canonical Brief and PRD view",
                "The PRD view must hand off cleanly into Spec, Plan, and Tasks",
                "Every major requirement must trace to a named research finding or source cluster",
                "The workflow must stay minimal enough for repeated weekly use",
            ],
            "interfaces": [
                "Research context document",
                "Canonical artifact graph",
                "Rendered PRD markdown view",
                "Task handoff surface",
            ],
            "journeys": [
                "Research lead compiles findings into canonical artifacts",
                "Product lead reviews the rendered PRD and requirements",
                "Engineering lead accepts the build-ready handoff",
            ],
            "invariants": [
                "No PRD requirement exists without research support",
                "No task exists without a linked requirement or decision",
                "Derived views never become the source of truth",
            ],
            "dependencies": [
                "Research source set",
                "Canonical artifact schemas",
                "PRD rendering view",
            ],
            "acceptance_criteria": list(manifest.get("success_checks", [])),
        }
    )

    decision = base_artifact(
        f"decision-{work_unit_id}",
        "Decision",
        work_unit_id,
        branch_id,
        owner_actor_id,
        approver_actor_ids,
        timestamp,
        status="approved",
        references=[spec["artifact_id"], literature["artifact_id"]],
    )
    decision.update(
        {
            "decision_id": f"{work_unit_id}-prd-001",
            "decision_statement": (
                "Use the research-backed PRD bridge as the flagship Jini loop and keep the workflow compact."
            ),
            "options_considered": [
                "Generate PRDs directly from ad hoc prompts",
                "Use a research-backed canonical artifact chain that renders a PRD view",
            ],
            "selected_option": "Use a research-backed canonical artifact chain that renders a PRD view",
            "rationale": (
                "This gives product teams traceability and gives engineering teams a cleaner handoff."
            ),
            "tradeoffs": [
                "Slightly more structure up front",
                "Much stronger continuity and correctness downstream",
            ],
            "decision_owner": owner_actor_id,
            "effective_scope": "Research-to-PRD-to-build handoff workflow",
        }
    )

    plan = base_artifact(
        f"plan-{work_unit_id}",
        "Plan",
        work_unit_id,
        branch_id,
        owner_actor_id,
        approver_actor_ids,
        timestamp,
        status="approved",
        references=[decision["artifact_id"]],
    )
    plan.update(
        {
            "slices": [
                "Stabilize research artifacts and source traceability",
                "Render and review the PRD view",
                "Translate PRD into build-ready spec and tasks",
            ],
            "dependencies": [
                "Source and method artifacts must be ready before PRD rendering",
                "PRD review must complete before task handoff",
            ],
            "milestones": [
                "Research synthesis complete",
                "PRD reviewed",
                "Build handoff accepted",
            ],
            "rollback_intent": (
                "If PRD review exposes unsupported claims, return to research synthesis and regenerate the view."
            ),
            "acceptance_gates": [
                "Sources and findings are explicit",
                "PRD sections trace to canonical artifacts",
                "Tasks trace to validated requirements",
            ],
        }
    )

    tasks = base_artifact(
        f"tasks-{work_unit_id}",
        "Tasks",
        work_unit_id,
        branch_id,
        owner_actor_id,
        approver_actor_ids,
        timestamp,
        status="reviewed",
        references=[plan["artifact_id"]],
    )
    tasks.update(
        {
            "tasks": [
                "Validate source coverage and finalize research synthesis",
                "Review and approve the rendered PRD",
                "Confirm build-ready requirements and task ownership",
            ],
            "ownership": [
                "research-lead",
                "product-lead",
                "engineering-lead",
            ],
            "status_per_task": [
                "pending",
                "pending",
                "pending",
            ],
            "blocked_by": [],
            "deliverables": [
                "Reviewed research artifacts",
                "Approved PRD view",
                "Accepted build handoff",
            ],
        }
    )

    evidence = base_artifact(
        f"evidence-{work_unit_id}",
        "Evidence",
        work_unit_id,
        branch_id,
        "research-ops",
        approver_actor_ids,
        timestamp,
        status="reviewed",
        references=[sources["artifact_id"], literature["artifact_id"], method["artifact_id"], spec["artifact_id"]],
    )
    evidence.update(
        {
            "target_artifact_id": spec["artifact_id"],
            "target_revision": 1,
            "claims_validated": [
                "The workflow problem is repeated across multiple source types",
                "A PRD bridge is the highest-value next surface",
                "The proposed requirements map back to validated findings",
            ],
            "test_results": [
                "Source set completeness reviewed against the defined research question",
            ],
            "review_results": [
                "Product review confirms the PRD view matches the intended audience and scope",
            ],
            "operational_results": [
                "No operational side effects; workflow remains documentation- and planning-only at this stage",
            ],
            "residual_risks": [
                "Further buyer validation may change packaging or pricing emphasis",
            ],
        }
    )

    dump_document(target_dir / "work-unit.yaml", work_unit)
    dump_document(target_dir / "pack-instance.yaml", pack_instance)
    dump_document(artifact_dir / "01-brief.yaml", brief)
    dump_document(artifact_dir / "02-assumptions.yaml", assumptions)
    dump_document(artifact_dir / "03-sources.yaml", sources)
    dump_document(artifact_dir / "04-literature.yaml", literature)
    dump_document(artifact_dir / "05-method.yaml", method)
    dump_document(artifact_dir / "06-spec.yaml", spec)
    dump_document(artifact_dir / "07-decision.yaml", decision)
    dump_document(artifact_dir / "08-plan.yaml", plan)
    dump_document(artifact_dir / "09-tasks.yaml", tasks)
    dump_document(artifact_dir / "10-evidence.yaml", evidence)
    write_prd_view(target_dir, title, brief, literature, method, spec, plan)
    return target_dir


def write_compiled_travel_plan_pack(
    manifest: dict[str, Any],
    pack_id: str,
    output_dir: Path,
    work_unit_id: str,
    title: str,
    purpose: str,
    owner_actor_id: str,
    approver_actor_ids: list[str],
    stakeholder_actor_ids: list[str],
    branch_id: str,
    context_reference: str,
) -> Path:
    target_dir = output_dir.expanduser()
    if target_dir.exists():
        raise FileExistsError(f"Output directory already exists: {target_dir}")

    artifact_dir = target_dir / "artifacts"
    artifact_dir.mkdir(parents=True, exist_ok=False)
    timestamp = now_utc()

    work_unit = {
        "work_unit_id": work_unit_id,
        "title": title,
        "purpose": purpose,
        "current_state": "decided",
        "profile_id": manifest["target_profile"],
        "active_extensions": manifest.get("extensions", []),
        "branch_id": branch_id,
        "parent_work_unit_id": "",
        "owner_actor_id": owner_actor_id,
        "approver_actor_ids": approver_actor_ids,
        "operator_actor_id": "",
        "rollback_authority_actor_id": "",
        "service_owner_actor_id": "",
        "stakeholder_actor_ids": stakeholder_actor_ids,
        "created_at": timestamp,
        "updated_at": timestamp,
    }

    pack_instance = {
        "pack_id": pack_id,
        "pack_version": manifest["version"],
        "compiled_flow": manifest.get("compiled_flow", []),
        "target_profile": manifest["target_profile"],
        "extensions": manifest.get("extensions", []),
        "control_packs": manifest.get("control_packs", []),
        "work_unit_id": work_unit_id,
        "compiled_at": timestamp,
        "compiled_from_context": context_reference,
    }

    brief = base_artifact(
        f"brief-{work_unit_id}",
        "Brief",
        work_unit_id,
        branch_id,
        owner_actor_id,
        approver_actor_ids,
        timestamp,
        status="approved",
        references=[context_reference],
    )
    brief.update(
        {
            "objective": purpose,
            "stakeholders": stakeholder_actor_ids,
            "constraints": [
                "Dates and budget must be explicit before bookings are finalized",
                "The plan should be realistic for one traveler or small group",
                "Risk and contingency notes should be clear enough for handoff",
            ],
            "success_criteria": list(manifest.get("success_checks", [])),
            "non_goals": [
                "Real-time price scraping",
                "Live booking automation",
            ],
            "scope_summary": (
                "Turn a travel goal into a structured itinerary, decision record, logistics plan, "
                "and execution checklist using the same core Jini semantics."
            ),
        }
    )

    assumptions = base_artifact(
        f"assumptions-{work_unit_id}",
        "Assumptions",
        work_unit_id,
        branch_id,
        owner_actor_id,
        approver_actor_ids,
        timestamp,
        status="reviewed",
        references=[brief["artifact_id"]],
    )
    assumptions.update(
        {
            "assumptions": [
                "A compact itinerary with explicit constraints is more useful than a broad destination list",
                "A traveler needs visible budget and logistics tradeoffs before execution",
                "The same proof and task discipline used in software can improve personal planning reliability",
            ],
            "known_unknowns": [
                "Which activities are fixed-date versus optional",
                "Which budget constraint is truly hard",
            ],
            "validation_plan": [
                "Check that itinerary, budget, and logistics align before execution",
                "Refresh evidence if dates or location assumptions change materially",
            ],
            "deferred_questions": [
                "Whether to upgrade the trip into a multi-party coordination workflow later",
            ],
        }
    )

    spec = base_artifact(
        f"spec-{work_unit_id}",
        "Spec",
        work_unit_id,
        branch_id,
        owner_actor_id,
        approver_actor_ids,
        timestamp,
        status="approved",
        references=[brief["artifact_id"], assumptions["artifact_id"]],
    )
    spec.update(
        {
            "requirements": [
                "Capture destination, dates, and budget constraints explicitly",
                "Produce a day-by-day or phase-by-phase itinerary summary",
                "List logistics requirements such as transport, lodging, and documents",
                "Surface fallback or contingency steps when travel assumptions break",
            ],
            "interfaces": [
                "Itinerary view",
                "Task board",
                "Wiki export bundle",
            ],
            "journeys": [
                "Traveler aligns on dates and budget",
                "Traveler reviews itinerary and logistics",
                "Traveler verifies readiness before booking or departure",
            ],
            "invariants": [
                "Dates, budget, and logistics should not disagree across artifacts",
                "Evidence should bind to the active itinerary revision when details change",
            ],
            "dependencies": [
                "destination preferences",
                "date constraints",
                "budget guardrails",
            ],
            "acceptance_criteria": list(manifest.get("success_checks", [])),
        }
    )

    decision = base_artifact(
        f"decision-{work_unit_id}",
        "Decision",
        work_unit_id,
        branch_id,
        owner_actor_id,
        approver_actor_ids,
        timestamp,
        status="approved",
        references=[spec["artifact_id"]],
    )
    decision.update(
        {
            "decision_id": f"{work_unit_id}-travel-001",
            "decision_statement": "Use a compact itinerary-first planning workflow instead of ad hoc travel notes.",
            "options_considered": [
                "Keep travel planning as loose notes and links",
                "Use a structured Jini pack with itinerary, tasks, and verification surfaces",
            ],
            "selected_option": "Use a structured Jini pack with itinerary, tasks, and verification surfaces",
            "rationale": "Travel plans fail in the same predictable ways as delivery plans when state, tasks, and contingencies are implicit.",
            "tradeoffs": [
                "Slightly more structure up front",
                "Far better continuity and handoff",
            ],
            "decision_owner": owner_actor_id,
            "effective_scope": "End-to-end trip planning and execution readiness",
        }
    )

    plan = base_artifact(
        f"plan-{work_unit_id}",
        "Plan",
        work_unit_id,
        branch_id,
        owner_actor_id,
        approver_actor_ids,
        timestamp,
        status="approved",
        references=[decision["artifact_id"]],
    )
    plan.update(
        {
            "slices": [
                "Lock dates, destination, and budget",
                "Build itinerary and logistics sequence",
                "Verify readiness and contingencies before execution",
            ],
            "dependencies": [
                "Budget and date alignment before itinerary lock",
                "Itinerary lock before readiness verification",
            ],
            "milestones": [
                "Core trip constraints agreed",
                "Itinerary and logistics drafted",
                "Execution readiness verified",
            ],
            "rollback_intent": "If dates or budget break materially, return to constraint alignment before bookings proceed.",
            "acceptance_gates": [
                "Dates, budget, and logistics are explicit",
                "Itinerary is reviewable as one coherent plan",
                "Contingencies are visible before execution",
            ],
        }
    )

    tasks = base_artifact(
        f"tasks-{work_unit_id}",
        "Tasks",
        work_unit_id,
        branch_id,
        owner_actor_id,
        approver_actor_ids,
        timestamp,
        status="approved",
        references=[plan["artifact_id"]],
    )
    tasks.update(
        {
            "tasks": [
                "Confirm destination, dates, and budget guardrails",
                "Draft itinerary and major logistics checkpoints",
                "Review contingencies and verify trip readiness",
            ],
            "ownership": [
                owner_actor_id,
                owner_actor_id,
                owner_actor_id,
            ],
            "status_per_task": [
                "pending",
                "pending",
                "pending",
            ],
            "blocked_by": [],
            "deliverables": [
                "Locked core constraints",
                "Draft itinerary and logistics plan",
                "Ready-to-execute trip checklist",
            ],
        }
    )

    dump_document(target_dir / "work-unit.yaml", work_unit)
    dump_document(target_dir / "pack-instance.yaml", pack_instance)
    dump_document(artifact_dir / "01-brief.yaml", brief)
    dump_document(artifact_dir / "02-assumptions.yaml", assumptions)
    dump_document(artifact_dir / "03-spec.yaml", spec)
    dump_document(artifact_dir / "04-decision.yaml", decision)
    dump_document(artifact_dir / "05-plan.yaml", plan)
    dump_document(artifact_dir / "06-tasks.yaml", tasks)
    write_itinerary_view(target_dir, title, brief, spec, plan, tasks)
    return target_dir


def write_compiled_budget_cycle_pack(
    manifest: dict[str, Any],
    pack_id: str,
    output_dir: Path,
    work_unit_id: str,
    title: str,
    purpose: str,
    owner_actor_id: str,
    approver_actor_ids: list[str],
    stakeholder_actor_ids: list[str],
    branch_id: str,
    context_reference: str,
) -> Path:
    target_dir = output_dir.expanduser()
    if target_dir.exists():
        raise FileExistsError(f"Output directory already exists: {target_dir}")

    artifact_dir = target_dir / "artifacts"
    artifact_dir.mkdir(parents=True, exist_ok=False)
    timestamp = now_utc()

    work_unit = {
        "work_unit_id": work_unit_id,
        "title": title,
        "purpose": purpose,
        "current_state": "decided",
        "profile_id": manifest["target_profile"],
        "active_extensions": manifest.get("extensions", []),
        "branch_id": branch_id,
        "parent_work_unit_id": "",
        "owner_actor_id": owner_actor_id,
        "approver_actor_ids": approver_actor_ids,
        "operator_actor_id": "",
        "rollback_authority_actor_id": "",
        "service_owner_actor_id": "",
        "stakeholder_actor_ids": stakeholder_actor_ids,
        "created_at": timestamp,
        "updated_at": timestamp,
    }

    pack_instance = {
        "pack_id": pack_id,
        "pack_version": manifest["version"],
        "compiled_flow": manifest.get("compiled_flow", []),
        "target_profile": manifest["target_profile"],
        "extensions": manifest.get("extensions", []),
        "control_packs": manifest.get("control_packs", []),
        "work_unit_id": work_unit_id,
        "compiled_at": timestamp,
        "compiled_from_context": context_reference,
    }

    brief = base_artifact(
        f"brief-{work_unit_id}",
        "Brief",
        work_unit_id,
        branch_id,
        owner_actor_id,
        approver_actor_ids,
        timestamp,
        status="approved",
        references=[context_reference],
    )
    brief.update(
        {
            "objective": purpose,
            "stakeholders": stakeholder_actor_ids,
            "constraints": [
                "Income, fixed obligations, and target savings must be explicit",
                "The workflow should produce a budget simple enough to review monthly",
                "Tradeoffs and fallback cuts should be visible before execution",
            ],
            "success_criteria": list(manifest.get("success_checks", [])),
            "non_goals": [
                "Live bank synchronization",
                "Automated bill payment",
            ],
            "scope_summary": (
                "Turn a budgeting goal into a structured monthly plan, decision record, "
                "checklist, and verification surface using the same Jini core."
            ),
        }
    )

    assumptions = base_artifact(
        f"assumptions-{work_unit_id}",
        "Assumptions",
        work_unit_id,
        branch_id,
        owner_actor_id,
        approver_actor_ids,
        timestamp,
        status="reviewed",
        references=[brief["artifact_id"]],
    )
    assumptions.update(
        {
            "assumptions": [
                "A compact monthly budget with explicit categories is more useful than scattered spreadsheets",
                "People need visible tradeoffs between savings, fixed costs, and discretionary spend",
                "The same verification discipline used in delivery planning can improve financial follow-through",
            ],
            "known_unknowns": [
                "Which discretionary categories can absorb cuts without breaking the plan",
                "Whether income volatility requires a deeper contingency reserve",
            ],
            "validation_plan": [
                "Check that category totals and savings goals align before execution",
                "Refresh evidence when income or fixed obligations change materially",
            ],
            "deferred_questions": [
                "Whether to extend the budget into quarterly forecasting later",
            ],
        }
    )

    spec = base_artifact(
        f"spec-{work_unit_id}",
        "Spec",
        work_unit_id,
        branch_id,
        owner_actor_id,
        approver_actor_ids,
        timestamp,
        status="approved",
        references=[brief["artifact_id"], assumptions["artifact_id"]],
    )
    spec.update(
        {
            "requirements": [
                "Capture monthly income, fixed obligations, and savings targets explicitly",
                "Produce a category-based budget that is reviewable as one coherent plan",
                "List fallback cuts or deferrals if the budget goes out of bounds",
                "Surface verification steps for month-end review and adjustment",
            ],
            "interfaces": [
                "Budget view",
                "Task board",
                "Wiki export bundle",
            ],
            "journeys": [
                "Owner aligns on goals and current obligations",
                "Owner reviews the monthly category budget",
                "Owner verifies adherence and adjusts before the next cycle",
            ],
            "invariants": [
                "Income, expenses, and savings assumptions should not disagree across artifacts",
                "Evidence should bind to the active budget revision when assumptions change",
            ],
            "dependencies": [
                "income assumptions",
                "fixed cost list",
                "savings targets",
            ],
            "acceptance_criteria": list(manifest.get("success_checks", [])),
        }
    )

    decision = base_artifact(
        f"decision-{work_unit_id}",
        "Decision",
        work_unit_id,
        branch_id,
        owner_actor_id,
        approver_actor_ids,
        timestamp,
        status="approved",
        references=[spec["artifact_id"]],
    )
    decision.update(
        {
            "decision_id": f"{work_unit_id}-budget-001",
            "decision_statement": "Use a structured monthly budget workflow instead of scattered budgeting notes.",
            "options_considered": [
                "Keep budgeting in ad hoc notes or spreadsheet fragments",
                "Use a structured Jini pack with plan, tasks, and verification surfaces",
            ],
            "selected_option": "Use a structured Jini pack with plan, tasks, and verification surfaces",
            "rationale": "Budget plans fail predictably when assumptions, tasks, and fallback cuts stay implicit.",
            "tradeoffs": [
                "Slightly more structure every cycle",
                "Much better reviewability and continuity",
            ],
            "decision_owner": owner_actor_id,
            "effective_scope": "Monthly budget planning and verification readiness",
        }
    )

    plan = base_artifact(
        f"plan-{work_unit_id}",
        "Plan",
        work_unit_id,
        branch_id,
        owner_actor_id,
        approver_actor_ids,
        timestamp,
        status="approved",
        references=[decision["artifact_id"]],
    )
    plan.update(
        {
            "slices": [
                "Lock income, obligations, and savings targets",
                "Draft category budget and fallback cuts",
                "Verify budget adherence and adjustment points",
            ],
            "dependencies": [
                "Income and fixed obligations aligned before category planning",
                "Category plan drafted before verification and month-end review",
            ],
            "milestones": [
                "Financial constraints agreed",
                "Monthly budget drafted",
                "Verification and adjustment checklist ready",
            ],
            "rollback_intent": "If income or fixed obligations move materially, return to constraint alignment before locking the monthly budget.",
            "acceptance_gates": [
                "Income, obligations, and savings goals are explicit",
                "Category budget is reviewable as one coherent plan",
                "Fallback cuts are visible before execution",
            ],
        }
    )

    tasks = base_artifact(
        f"tasks-{work_unit_id}",
        "Tasks",
        work_unit_id,
        branch_id,
        owner_actor_id,
        approver_actor_ids,
        timestamp,
        status="approved",
        references=[plan["artifact_id"]],
    )
    tasks.update(
        {
            "tasks": [
                "Confirm income, obligations, and savings targets",
                "Draft the category budget and fallback cuts",
                "Review month-end verification and adjustment steps",
            ],
            "ownership": [owner_actor_id, owner_actor_id, owner_actor_id],
            "status_per_task": ["pending", "pending", "pending"],
            "blocked_by": [],
            "deliverables": [
                "Locked financial constraints",
                "Draft monthly budget",
                "Ready-to-run verification checklist",
            ],
        }
    )

    dump_document(target_dir / "work-unit.yaml", work_unit)
    dump_document(target_dir / "pack-instance.yaml", pack_instance)
    dump_document(artifact_dir / "01-brief.yaml", brief)
    dump_document(artifact_dir / "02-assumptions.yaml", assumptions)
    dump_document(artifact_dir / "03-spec.yaml", spec)
    dump_document(artifact_dir / "04-decision.yaml", decision)
    dump_document(artifact_dir / "05-plan.yaml", plan)
    dump_document(artifact_dir / "06-tasks.yaml", tasks)
    write_budget_view(target_dir, title, brief, assumptions, spec, plan, tasks)
    return target_dir


def write_compiled_incident_response_pack(
    manifest: dict[str, Any],
    pack_id: str,
    output_dir: Path,
    work_unit_id: str,
    title: str,
    purpose: str,
    owner_actor_id: str,
    approver_actor_ids: list[str],
    stakeholder_actor_ids: list[str],
    branch_id: str,
    context_reference: str,
) -> Path:
    target_dir = output_dir.expanduser()
    if target_dir.exists():
        raise FileExistsError(f"Output directory already exists: {target_dir}")

    artifact_dir = target_dir / "artifacts"
    artifact_dir.mkdir(parents=True, exist_ok=False)
    timestamp = now_utc()

    work_unit = {
        "work_unit_id": work_unit_id,
        "title": title,
        "purpose": purpose,
        "current_state": "decided",
        "profile_id": manifest["target_profile"],
        "active_extensions": manifest.get("extensions", []),
        "branch_id": branch_id,
        "parent_work_unit_id": "",
        "owner_actor_id": owner_actor_id,
        "approver_actor_ids": approver_actor_ids,
        "operator_actor_id": "",
        "rollback_authority_actor_id": "",
        "service_owner_actor_id": "",
        "stakeholder_actor_ids": stakeholder_actor_ids,
        "created_at": timestamp,
        "updated_at": timestamp,
    }

    pack_instance = {
        "pack_id": pack_id,
        "pack_version": manifest["version"],
        "compiled_flow": manifest.get("compiled_flow", []),
        "target_profile": manifest["target_profile"],
        "extensions": manifest.get("extensions", []),
        "control_packs": manifest.get("control_packs", []),
        "work_unit_id": work_unit_id,
        "compiled_at": timestamp,
        "compiled_from_context": context_reference,
    }

    brief = base_artifact(
        f"brief-{work_unit_id}",
        "Brief",
        work_unit_id,
        branch_id,
        owner_actor_id,
        approver_actor_ids,
        timestamp,
        status="approved",
        references=[context_reference],
    )
    brief.update(
        {
            "objective": purpose,
            "stakeholders": stakeholder_actor_ids,
            "constraints": [
                "Incident scope, severity, and current customer impact must be explicit",
                "Mitigation steps should prefer bounded, reversible actions over improvised changes",
                "Communication and verification cadence should stay visible until closure",
            ],
            "success_criteria": list(manifest.get("success_checks", [])),
            "non_goals": [
                "Unreviewed permanent architecture changes during stabilization",
                "Open-ended investigation without explicit owner or next verification point",
            ],
            "scope_summary": (
                "Turn an incident into a structured stabilization, communication, verification, "
                "and rollback-aware response plan using the same Jini core semantics."
            ),
        }
    )

    assumptions = base_artifact(
        f"assumptions-{work_unit_id}",
        "Assumptions",
        work_unit_id,
        branch_id,
        owner_actor_id,
        approver_actor_ids,
        timestamp,
        status="reviewed",
        references=[brief["artifact_id"]],
    )
    assumptions.update(
        {
            "assumptions": [
                "A concise incident plan with explicit owners is more useful than free-form war-room notes",
                "The same evidence and rollback discipline used in delivery should govern response actions",
                "Repo-local and runtime-local verification surfaces should be refreshed before closure",
            ],
            "known_unknowns": [
                "Whether the incident is isolated to one service boundary or reflects a wider regression",
                "Whether mitigation or rollback has a lower risk profile given the current blast radius",
            ],
            "validation_plan": [
                "Check service health, key workflows, and operator notes after each major mitigation",
                "Refresh evidence and re-open scope assessment if customer impact expands materially",
            ],
            "deferred_questions": [
                "Whether the incident should graduate into a longer postmortem or change-program work unit",
            ],
        }
    )

    spec = base_artifact(
        f"spec-{work_unit_id}",
        "Spec",
        work_unit_id,
        branch_id,
        owner_actor_id,
        approver_actor_ids,
        timestamp,
        status="approved",
        references=[brief["artifact_id"], assumptions["artifact_id"]],
    )
    spec.update(
        {
            "requirements": [
                "Capture severity, impacted surface, incident owner, and current mitigation status explicitly",
                "Produce a stabilization sequence with bounded commands and rollback-aware actions",
                "List communication checkpoints for stakeholders and operators during active response",
                "Surface verification, closure, and follow-up requirements before the incident is considered stable",
            ],
            "interfaces": [
                "Response view",
                "Task board",
                "Wiki export bundle",
            ],
            "journeys": [
                "Incident owner assesses scope and assigns the immediate response path",
                "Operators execute stabilization steps while keeping communication and evidence current",
                "Service owner verifies recovery, residual risk, and closure readiness",
            ],
            "invariants": [
                "Severity, impact scope, and mitigation state should not disagree across artifacts",
                "Evidence and rollback notes should bind to the active incident revision before closure",
            ],
            "dependencies": [
                "service health signals",
                "bounded repo or runtime verification checks",
                "stakeholder communication path",
            ],
            "acceptance_criteria": list(manifest.get("success_checks", [])),
        }
    )

    decision = base_artifact(
        f"decision-{work_unit_id}",
        "Decision",
        work_unit_id,
        branch_id,
        owner_actor_id,
        approver_actor_ids,
        timestamp,
        status="approved",
        references=[spec["artifact_id"]],
    )
    decision.update(
        {
            "decision_id": f"{work_unit_id}-incident-001",
            "decision_statement": "Use a structured incident response workflow instead of ad hoc mitigation notes.",
            "options_considered": [
                "Run the incident through chat logs and ephemeral war-room notes",
                "Use a structured Jini pack with response phases, tasks, and verification surfaces",
            ],
            "selected_option": "Use a structured Jini pack with response phases, tasks, and verification surfaces",
            "rationale": "Incident response degrades quickly when owners, rollback paths, and verification steps remain implicit.",
            "tradeoffs": [
                "Slightly more ceremony during the incident",
                "Far better continuity, auditability, and handoff under pressure",
            ],
            "decision_owner": owner_actor_id,
            "effective_scope": "Active incident response, verification, and closure readiness",
        }
    )

    plan = base_artifact(
        f"plan-{work_unit_id}",
        "Plan",
        work_unit_id,
        branch_id,
        owner_actor_id,
        approver_actor_ids,
        timestamp,
        status="approved",
        references=[decision["artifact_id"]],
    )
    plan.update(
        {
            "slices": [
                "Assess scope, assign ownership, and stabilize the highest-risk surface",
                "Execute bounded mitigation and stakeholder communication checkpoints",
                "Verify recovery, residual risk, and closure readiness",
            ],
            "dependencies": [
                "Scope assessment and owner assignment before deeper mitigation",
                "Stabilization and communication before closure verification",
            ],
            "milestones": [
                "Scope and severity are explicit",
                "Primary mitigation path is in flight",
                "Recovery verification and closure notes are ready",
            ],
            "rollback_intent": "If mitigation worsens impact or verification regresses, return to the previous stable state before widening changes.",
            "acceptance_gates": [
                "Incident scope, owners, and mitigation state are explicit",
                "Communication checkpoints are visible during response",
                "Verification and rollback paths are available before closure",
            ],
        }
    )

    tasks = base_artifact(
        f"tasks-{work_unit_id}",
        "Tasks",
        work_unit_id,
        branch_id,
        owner_actor_id,
        approver_actor_ids,
        timestamp,
        status="approved",
        references=[plan["artifact_id"]],
    )
    tasks.update(
        {
            "tasks": [
                "Confirm severity, scope, owner, and first stabilization action",
                "Record mitigation, communication, and rollback checkpoints",
                "Verify recovery state and capture closure readiness",
            ],
            "ownership": [owner_actor_id, owner_actor_id, owner_actor_id],
            "status_per_task": ["pending", "pending", "pending"],
            "blocked_by": [],
            "deliverables": [
                "Explicit incident scope and owner assignment",
                "Active response log with mitigation and communication checkpoints",
                "Ready-to-close verification and residual-risk checklist",
            ],
        }
    )

    dump_document(target_dir / "work-unit.yaml", work_unit)
    dump_document(target_dir / "pack-instance.yaml", pack_instance)
    dump_document(artifact_dir / "01-brief.yaml", brief)
    dump_document(artifact_dir / "02-assumptions.yaml", assumptions)
    dump_document(artifact_dir / "03-spec.yaml", spec)
    dump_document(artifact_dir / "04-decision.yaml", decision)
    dump_document(artifact_dir / "05-plan.yaml", plan)
    dump_document(artifact_dir / "06-tasks.yaml", tasks)
    write_response_view(target_dir, title, brief, assumptions, spec, plan, tasks)
    return target_dir


def write_compiled_compliance_audit_pack(
    manifest: dict[str, Any],
    pack_id: str,
    output_dir: Path,
    work_unit_id: str,
    title: str,
    purpose: str,
    owner_actor_id: str,
    approver_actor_ids: list[str],
    stakeholder_actor_ids: list[str],
    branch_id: str,
    context_reference: str,
) -> Path:
    target_dir = output_dir.expanduser()
    if target_dir.exists():
        raise FileExistsError(f"Output directory already exists: {target_dir}")

    artifact_dir = target_dir / "artifacts"
    artifact_dir.mkdir(parents=True, exist_ok=False)
    timestamp = now_utc()

    work_unit = {
        "work_unit_id": work_unit_id,
        "title": title,
        "purpose": purpose,
        "current_state": "decided",
        "profile_id": manifest["target_profile"],
        "active_extensions": manifest.get("extensions", []),
        "branch_id": branch_id,
        "parent_work_unit_id": "",
        "owner_actor_id": owner_actor_id,
        "approver_actor_ids": approver_actor_ids,
        "operator_actor_id": "",
        "rollback_authority_actor_id": "",
        "service_owner_actor_id": "",
        "stakeholder_actor_ids": stakeholder_actor_ids,
        "created_at": timestamp,
        "updated_at": timestamp,
    }

    pack_instance = {
        "pack_id": pack_id,
        "pack_version": manifest["version"],
        "compiled_flow": manifest.get("compiled_flow", []),
        "target_profile": manifest["target_profile"],
        "extensions": manifest.get("extensions", []),
        "control_packs": manifest.get("control_packs", []),
        "work_unit_id": work_unit_id,
        "compiled_at": timestamp,
        "compiled_from_context": context_reference,
    }

    brief = base_artifact(
        f"brief-{work_unit_id}",
        "Brief",
        work_unit_id,
        branch_id,
        owner_actor_id,
        approver_actor_ids,
        timestamp,
        status="approved",
        references=[context_reference],
    )
    brief.update(
        {
            "objective": purpose,
            "stakeholders": stakeholder_actor_ids,
            "constraints": [
                "Control obligations, evidence paths, and approval boundaries must be explicit",
                "The audit surface should stay concise enough for reviewers to inspect quickly",
                "Every closure claim should map back to a concrete artifact or verification result",
            ],
            "success_criteria": list(manifest.get("success_checks", [])),
            "non_goals": [
                "Inventing new regulatory semantics in the kernel",
                "Treating unverified narrative summaries as sufficient proof",
            ],
            "scope_summary": (
                "Turn a regulated or compliance-oriented review into a structured audit, "
                "evidence, and approval workflow using the same Jini core semantics."
            ),
        }
    )

    assumptions = base_artifact(
        f"assumptions-{work_unit_id}",
        "Assumptions",
        work_unit_id,
        branch_id,
        owner_actor_id,
        approver_actor_ids,
        timestamp,
        status="reviewed",
        references=[brief["artifact_id"]],
    )
    assumptions.update(
        {
            "assumptions": [
                "Regulated review becomes more reliable when control statements, evidence, and approval gates stay explicit",
                "A compact audit plan is more usable than a large free-form compliance memo",
                "The same lifecycle and evidence semantics should work without introducing domain-specific kernel rules",
            ],
            "known_unknowns": [
                "Which obligations demand fresh evidence versus reuse of existing proof",
                "Whether any residual risk requires waiver or phased remediation",
            ],
            "validation_plan": [
                "Check that every audit requirement maps to evidence, owner, and reviewer path",
                "Refresh evidence and approval if obligations or scope change materially",
            ],
            "deferred_questions": [
                "Whether the audit should expand into a broader remediation program after closure",
            ],
        }
    )

    spec = base_artifact(
        f"spec-{work_unit_id}",
        "Spec",
        work_unit_id,
        branch_id,
        owner_actor_id,
        approver_actor_ids,
        timestamp,
        status="approved",
        references=[brief["artifact_id"], assumptions["artifact_id"]],
    )
    spec.update(
        {
            "requirements": [
                "Capture control scope, reviewer roles, and approval boundaries explicitly",
                "Produce a reviewable audit checklist tied to evidence and residual-risk outcomes",
                "List remediation, waiver, or follow-up requirements before closure",
                "Surface publication or handoff outputs without weakening traceability",
            ],
            "interfaces": [
                "Audit view",
                "Task board",
                "Wiki export bundle",
            ],
            "journeys": [
                "Audit owner aligns scope and obligations with reviewers",
                "Reviewers inspect evidence and remediation status against the audit checklist",
                "Approvers determine whether closure, waiver, or follow-up is warranted",
            ],
            "invariants": [
                "Requirements, evidence, and approval scope should not disagree across artifacts",
                "Residual risks should remain visible until explicitly accepted or remediated",
            ],
            "dependencies": [
                "control inventory",
                "evidence references",
                "approval path",
            ],
            "acceptance_criteria": list(manifest.get("success_checks", [])),
        }
    )

    decision = base_artifact(
        f"decision-{work_unit_id}",
        "Decision",
        work_unit_id,
        branch_id,
        owner_actor_id,
        approver_actor_ids,
        timestamp,
        status="approved",
        references=[spec["artifact_id"]],
    )
    decision.update(
        {
            "decision_id": f"{work_unit_id}-audit-001",
            "decision_statement": "Use a structured compliance audit workflow instead of informal review notes.",
            "options_considered": [
                "Rely on spreadsheets, emails, and free-form review summaries",
                "Use a structured Jini pack with audit phases, tasks, and approval-ready evidence surfaces",
            ],
            "selected_option": "Use a structured Jini pack with audit phases, tasks, and approval-ready evidence surfaces",
            "rationale": "Regulated work breaks down when controls, proof, and signoff remain implicit or scattered.",
            "tradeoffs": [
                "Slightly more upfront structure",
                "Far better traceability and approval discipline",
            ],
            "decision_owner": owner_actor_id,
            "effective_scope": "Audit planning, evidence review, remediation tracking, and closure readiness",
        }
    )

    plan = base_artifact(
        f"plan-{work_unit_id}",
        "Plan",
        work_unit_id,
        branch_id,
        owner_actor_id,
        approver_actor_ids,
        timestamp,
        status="approved",
        references=[decision["artifact_id"]],
    )
    plan.update(
        {
            "slices": [
                "Lock audit scope, obligations, and reviewers",
                "Review evidence and track remediation or waiver paths",
                "Verify approval readiness and closure conditions",
            ],
            "dependencies": [
                "Scope and reviewers aligned before detailed evidence review",
                "Evidence review completed before closure approval",
            ],
            "milestones": [
                "Audit scope agreed",
                "Evidence and remediation status reviewed",
                "Approval-ready closure decision prepared",
            ],
            "rollback_intent": "If evidence or obligations shift materially, return to scope alignment before closure or publication proceeds.",
            "acceptance_gates": [
                "Control scope and reviewer path are explicit",
                "Evidence and residual risks are visible as one coherent audit",
                "Approval readiness is reviewable before closure",
            ],
        }
    )

    tasks = base_artifact(
        f"tasks-{work_unit_id}",
        "Tasks",
        work_unit_id,
        branch_id,
        owner_actor_id,
        approver_actor_ids,
        timestamp,
        status="approved",
        references=[plan["artifact_id"]],
    )
    tasks.update(
        {
            "tasks": [
                "Confirm audit scope, reviewers, and approval boundaries",
                "Review evidence and record remediation or waiver paths",
                "Verify closure readiness and capture approval conditions",
            ],
            "ownership": [owner_actor_id, owner_actor_id, owner_actor_id],
            "status_per_task": ["pending", "pending", "pending"],
            "blocked_by": [],
            "deliverables": [
                "Locked audit scope and reviewer list",
                "Reviewed evidence and remediation summary",
                "Approval-ready closure checklist",
            ],
        }
    )

    dump_document(target_dir / "work-unit.yaml", work_unit)
    dump_document(target_dir / "pack-instance.yaml", pack_instance)
    dump_document(artifact_dir / "01-brief.yaml", brief)
    dump_document(artifact_dir / "02-assumptions.yaml", assumptions)
    dump_document(artifact_dir / "03-spec.yaml", spec)
    dump_document(artifact_dir / "04-decision.yaml", decision)
    dump_document(artifact_dir / "05-plan.yaml", plan)
    dump_document(artifact_dir / "06-tasks.yaml", tasks)
    write_audit_view(target_dir, title, brief, assumptions, spec, plan, tasks)
    return target_dir


def write_compiled_vendor_selection_pack(
    manifest: dict[str, Any],
    pack_id: str,
    output_dir: Path,
    work_unit_id: str,
    title: str,
    purpose: str,
    owner_actor_id: str,
    approver_actor_ids: list[str],
    stakeholder_actor_ids: list[str],
    branch_id: str,
    context_reference: str,
) -> Path:
    target_dir = output_dir.expanduser()
    if target_dir.exists():
        raise FileExistsError(f"Output directory already exists: {target_dir}")

    artifact_dir = target_dir / "artifacts"
    artifact_dir.mkdir(parents=True, exist_ok=False)
    timestamp = now_utc()

    work_unit = {
        "work_unit_id": work_unit_id,
        "title": title,
        "purpose": purpose,
        "current_state": "decided",
        "profile_id": manifest["target_profile"],
        "active_extensions": manifest.get("extensions", []),
        "branch_id": branch_id,
        "parent_work_unit_id": "",
        "owner_actor_id": owner_actor_id,
        "approver_actor_ids": approver_actor_ids,
        "operator_actor_id": "",
        "rollback_authority_actor_id": "",
        "service_owner_actor_id": "",
        "stakeholder_actor_ids": stakeholder_actor_ids,
        "created_at": timestamp,
        "updated_at": timestamp,
    }

    pack_instance = {
        "pack_id": pack_id,
        "pack_version": manifest["version"],
        "compiled_flow": manifest.get("compiled_flow", []),
        "target_profile": manifest["target_profile"],
        "extensions": manifest.get("extensions", []),
        "control_packs": manifest.get("control_packs", []),
        "work_unit_id": work_unit_id,
        "compiled_at": timestamp,
        "compiled_from_context": context_reference,
    }

    brief = base_artifact(
        f"brief-{work_unit_id}",
        "Brief",
        work_unit_id,
        branch_id,
        owner_actor_id,
        approver_actor_ids,
        timestamp,
        status="approved",
        references=[context_reference],
    )
    brief.update(
        {
            "objective": purpose,
            "stakeholders": stakeholder_actor_ids,
            "constraints": [
                "Evaluation criteria, budget guardrails, and decision authority should stay explicit",
                "The recommendation should remain compact enough for commercial reviewers to inspect quickly",
                "Commitment should not proceed until tradeoffs and fallback options are visible",
            ],
            "success_criteria": list(manifest.get("success_checks", [])),
            "non_goals": [
                "Replacing commercial judgment with opaque scoring",
                "Treating unstructured vendor notes as sufficient decision rationale",
            ],
            "scope_summary": (
                "Turn a vendor or partner selection into a structured scoring, decision, "
                "approval, and follow-through workflow using the same Jini core semantics."
            ),
        }
    )

    assumptions = base_artifact(
        f"assumptions-{work_unit_id}",
        "Assumptions",
        work_unit_id,
        branch_id,
        owner_actor_id,
        approver_actor_ids,
        timestamp,
        status="reviewed",
        references=[brief["artifact_id"]],
    )
    assumptions.update(
        {
            "assumptions": [
                "Commercial decisions improve when criteria, tradeoffs, and approval boundaries are explicit",
                "A compact selection packet is more reusable than scattered notes and spreadsheets",
                "Implementation follow-through should stay attached to the same decision surface as the recommendation",
            ],
            "known_unknowns": [
                "Which short-listed option will produce the best tradeoff between cost, speed, and operating fit",
                "Whether negotiation or phased rollout is required before commitment",
            ],
            "validation_plan": [
                "Check that each candidate is scored against the same criteria and budget guardrails",
                "Refresh the packet if new constraints, risks, or approval requirements emerge",
            ],
            "deferred_questions": [
                "Whether the selected vendor should trigger a later implementation or migration work unit",
            ],
        }
    )

    spec = base_artifact(
        f"spec-{work_unit_id}",
        "Spec",
        work_unit_id,
        branch_id,
        owner_actor_id,
        approver_actor_ids,
        timestamp,
        status="approved",
        references=[brief["artifact_id"], assumptions["artifact_id"]],
    )
    spec.update(
        {
            "requirements": [
                "Capture candidate set, evaluation criteria, and budget guardrails explicitly",
                "Produce a reviewable recommendation with tradeoffs, fallback, and approval path",
                "List negotiation or implementation follow-through before commitment proceeds",
                "Surface publishable decision outputs without weakening traceability",
            ],
            "interfaces": [
                "Selection view",
                "Task board",
                "Wiki export bundle",
            ],
            "journeys": [
                "Owner aligns criteria and short-list with reviewers",
                "Reviewers inspect tradeoffs and challenge the recommendation against budget and operating fit",
                "Approvers decide whether to commit, negotiate, or re-open evaluation",
            ],
            "invariants": [
                "Criteria, recommendation, and approval scope should not disagree across artifacts",
                "Fallback options should remain visible until the decision is accepted",
            ],
            "dependencies": [
                "candidate inventory",
                "budget guardrails",
                "approval path",
            ],
            "acceptance_criteria": list(manifest.get("success_checks", [])),
        }
    )

    decision = base_artifact(
        f"decision-{work_unit_id}",
        "Decision",
        work_unit_id,
        branch_id,
        owner_actor_id,
        approver_actor_ids,
        timestamp,
        status="approved",
        references=[spec["artifact_id"]],
    )
    decision.update(
        {
            "decision_id": f"{work_unit_id}-vendor-001",
            "decision_statement": "Use a structured vendor-selection workflow instead of scattered comparison notes.",
            "options_considered": [
                "Rely on spreadsheets, emails, and ad hoc commercial summaries",
                "Use a structured Jini pack with selection phases, tasks, and approval-ready outputs",
            ],
            "selected_option": "Use a structured Jini pack with selection phases, tasks, and approval-ready outputs",
            "rationale": "Commercial evaluation degrades when criteria, fallback options, and approver expectations remain implicit.",
            "tradeoffs": [
                "Slightly more upfront structure",
                "Far better continuity, reviewability, and handoff under time pressure",
            ],
            "decision_owner": owner_actor_id,
            "effective_scope": "Vendor evaluation, recommendation, approval readiness, and follow-through planning",
        }
    )

    plan = base_artifact(
        f"plan-{work_unit_id}",
        "Plan",
        work_unit_id,
        branch_id,
        owner_actor_id,
        approver_actor_ids,
        timestamp,
        status="approved",
        references=[decision["artifact_id"]],
    )
    plan.update(
        {
            "slices": [
                "Lock candidate set, criteria, and budget guardrails",
                "Review tradeoffs and converge on a recommendation with fallback options",
                "Verify approval readiness plus negotiation or implementation follow-through",
            ],
            "dependencies": [
                "Criteria aligned before recommendation hardens",
                "Recommendation reviewed before commitment or negotiation begins",
            ],
            "milestones": [
                "Candidate set and criteria agreed",
                "Recommendation and fallback path reviewed",
                "Approval-ready decision packet prepared",
            ],
            "rollback_intent": "If material tradeoffs or constraints change, return to criteria and candidate review before commitment proceeds.",
            "acceptance_gates": [
                "Criteria and guardrails are explicit",
                "Tradeoffs and fallback options are visible as one coherent decision packet",
                "Approval readiness is reviewable before commitment",
            ],
        }
    )

    tasks = base_artifact(
        f"tasks-{work_unit_id}",
        "Tasks",
        work_unit_id,
        branch_id,
        owner_actor_id,
        approver_actor_ids,
        timestamp,
        status="approved",
        references=[plan["artifact_id"]],
    )
    tasks.update(
        {
            "tasks": [
                "Confirm candidate set, criteria, and approval boundaries",
                "Review tradeoffs and record the recommendation with fallback options",
                "Verify approval readiness and capture negotiation or implementation follow-through",
            ],
            "ownership": [owner_actor_id, owner_actor_id, owner_actor_id],
            "status_per_task": ["pending", "pending", "pending"],
            "blocked_by": [],
            "deliverables": [
                "Locked candidate set and criteria",
                "Reviewed recommendation and fallback summary",
                "Approval-ready decision packet with next-step plan",
            ],
        }
    )

    dump_document(target_dir / "work-unit.yaml", work_unit)
    dump_document(target_dir / "pack-instance.yaml", pack_instance)
    dump_document(artifact_dir / "01-brief.yaml", brief)
    dump_document(artifact_dir / "02-assumptions.yaml", assumptions)
    dump_document(artifact_dir / "03-spec.yaml", spec)
    dump_document(artifact_dir / "04-decision.yaml", decision)
    dump_document(artifact_dir / "05-plan.yaml", plan)
    dump_document(artifact_dir / "06-tasks.yaml", tasks)
    write_selection_view(target_dir, title, brief, assumptions, spec, plan, tasks)
    return target_dir


def write_compiled_meeting_followup_pack(
    manifest: dict[str, Any],
    pack_id: str,
    output_dir: Path,
    work_unit_id: str,
    title: str,
    purpose: str,
    owner_actor_id: str,
    approver_actor_ids: list[str],
    stakeholder_actor_ids: list[str],
    branch_id: str,
    context_reference: str,
) -> Path:
    target_dir = output_dir.expanduser()
    if target_dir.exists():
        raise FileExistsError(f"Output directory already exists: {target_dir}")

    artifact_dir = target_dir / "artifacts"
    artifact_dir.mkdir(parents=True, exist_ok=False)
    timestamp = now_utc()

    work_unit = {
        "work_unit_id": work_unit_id,
        "title": title,
        "purpose": purpose,
        "current_state": "decided",
        "profile_id": manifest["target_profile"],
        "active_extensions": manifest.get("extensions", []),
        "branch_id": branch_id,
        "parent_work_unit_id": "",
        "owner_actor_id": owner_actor_id,
        "approver_actor_ids": approver_actor_ids,
        "operator_actor_id": "",
        "rollback_authority_actor_id": "",
        "service_owner_actor_id": "",
        "stakeholder_actor_ids": stakeholder_actor_ids,
        "created_at": timestamp,
        "updated_at": timestamp,
    }

    pack_instance = {
        "pack_id": pack_id,
        "pack_version": manifest["version"],
        "compiled_flow": manifest.get("compiled_flow", []),
        "target_profile": manifest["target_profile"],
        "extensions": manifest.get("extensions", []),
        "control_packs": manifest.get("control_packs", []),
        "work_unit_id": work_unit_id,
        "compiled_at": timestamp,
        "compiled_from_context": context_reference,
    }

    brief = base_artifact(
        f"brief-{work_unit_id}",
        "Brief",
        work_unit_id,
        branch_id,
        owner_actor_id,
        approver_actor_ids,
        timestamp,
        status="approved",
        references=[context_reference],
    )
    brief.update(
        {
            "objective": purpose,
            "stakeholders": stakeholder_actor_ids,
            "constraints": [
                "Decisions, owners, and due dates must be explicit before the meeting is considered closed",
                "Open questions should remain visible instead of being buried in notes",
                "Approvals or escalations should remain attached to the follow-up record",
            ],
            "success_criteria": list(manifest.get("success_checks", [])),
            "non_goals": [
                "Automatic meeting transcription",
                "Real-time calendar or chat integration",
            ],
            "scope_summary": (
                "Turn one meeting into a structured follow-up workflow with canonical decisions, tasks, "
                "owners, deadlines, and approval-ready outputs."
            ),
        }
    )

    assumptions = base_artifact(
        f"assumptions-{work_unit_id}",
        "Assumptions",
        work_unit_id,
        branch_id,
        owner_actor_id,
        approver_actor_ids,
        timestamp,
        status="reviewed",
        references=[brief["artifact_id"]],
    )
    assumptions.update(
        {
            "assumptions": [
                "A short, canonical follow-up packet is more useful than scattered notes and chat fragments",
                "People act faster when owners and due dates are visible in the same surface as the decision",
                "Meetings are only complete when unresolved questions and approvals are still visible",
            ],
            "known_unknowns": [
                "Whether every action item has a single accountable owner",
                "Which decisions need formal approval before execution",
            ],
            "validation_plan": [
                "Check that each task has an owner and deliverable before advancing the work unit",
                "Refresh evidence if new decisions or blockers emerge after the meeting",
            ],
            "deferred_questions": [
                "Whether to expand this follow-up into a larger project work unit later",
            ],
        }
    )

    spec = base_artifact(
        f"spec-{work_unit_id}",
        "Spec",
        work_unit_id,
        branch_id,
        owner_actor_id,
        approver_actor_ids,
        timestamp,
        status="approved",
        references=[brief["artifact_id"], assumptions["artifact_id"]],
    )
    spec.update(
        {
            "requirements": [
                "Capture meeting objective, decisions, owners, and deadlines explicitly",
                "Render follow-up as one coherent summary rather than a loose note dump",
                "Keep approvals, escalations, and unresolved questions visible before execution",
                "Produce a task-ready checklist for the immediate next cycle of work",
            ],
            "interfaces": [
                "Follow-up view",
                "Task board",
                "Wiki export bundle",
            ],
            "journeys": [
                "Facilitator closes the meeting into a canonical follow-up packet",
                "Owners review assigned tasks and challenge missing assumptions",
                "Approvers confirm which decisions are ready to act on",
            ],
            "invariants": [
                "Decisions, owners, and deadlines should not disagree across artifacts",
                "Approval state should remain explicit until the relevant decision is cleared",
            ],
            "dependencies": [
                "meeting objective",
                "task owners",
                "approval path",
            ],
            "acceptance_criteria": list(manifest.get("success_checks", [])),
        }
    )

    decision = base_artifact(
        f"decision-{work_unit_id}",
        "Decision",
        work_unit_id,
        branch_id,
        owner_actor_id,
        approver_actor_ids,
        timestamp,
        status="approved",
        references=[spec["artifact_id"]],
    )
    decision.update(
        {
            "decision_id": f"{work_unit_id}-meeting-001",
            "decision_statement": "Use a structured meeting follow-up workflow instead of scattered notes and implied ownership.",
            "options_considered": [
                "Leave follow-up in chat, notes, and personal memory",
                "Use a structured Jini pack with decisions, owners, tasks, and verification surfaces",
            ],
            "selected_option": "Use a structured Jini pack with decisions, owners, tasks, and verification surfaces",
            "rationale": "Most meeting drift happens after the meeting when ownership, approvals, and next actions become implicit again.",
            "tradeoffs": [
                "Slightly more structure after each meeting",
                "Far better clarity, continuity, and accountability",
            ],
            "decision_owner": owner_actor_id,
            "effective_scope": "Meeting summary, follow-up tasks, approvals, and execution readiness",
        }
    )

    plan = base_artifact(
        f"plan-{work_unit_id}",
        "Plan",
        work_unit_id,
        branch_id,
        owner_actor_id,
        approver_actor_ids,
        timestamp,
        status="approved",
        references=[decision["artifact_id"]],
    )
    plan.update(
        {
            "slices": [
                "Capture decisions, owners, and explicit next steps",
                "Review unresolved questions and approval requirements",
                "Verify follow-up readiness before the work leaves the meeting boundary",
            ],
            "dependencies": [
                "Decisions and owners aligned before downstream execution begins",
                "Approval requirements visible before tasks are treated as ready",
            ],
            "milestones": [
                "Meeting summary and owners locked",
                "Open questions and approvals reviewed",
                "Follow-up packet ready for execution",
            ],
            "rollback_intent": "If new information invalidates decisions or ownership, reopen the follow-up packet before execution continues.",
            "acceptance_gates": [
                "Decisions, owners, and dates are explicit",
                "Open questions and approvals remain visible",
                "Follow-up is reviewable as one coherent packet",
            ],
        }
    )

    tasks = base_artifact(
        f"tasks-{work_unit_id}",
        "Tasks",
        work_unit_id,
        branch_id,
        owner_actor_id,
        approver_actor_ids,
        timestamp,
        status="approved",
        references=[plan["artifact_id"]],
    )
    tasks.update(
        {
            "tasks": [
                "Capture the meeting decisions, owners, and due dates",
                "Review unresolved questions and approval boundaries",
                "Verify the follow-up packet before execution leaves the meeting boundary",
            ],
            "ownership": [
                owner_actor_id,
                owner_actor_id,
                owner_actor_id,
            ],
            "status_per_task": [
                "pending",
                "pending",
                "pending",
            ],
            "blocked_by": [],
            "deliverables": [
                "Canonical meeting summary",
                "Reviewed approval and escalation state",
                "Ready-to-execute follow-up packet",
            ],
        }
    )

    dump_document(target_dir / "work-unit.yaml", work_unit)
    dump_document(target_dir / "pack-instance.yaml", pack_instance)
    dump_document(artifact_dir / "01-brief.yaml", brief)
    dump_document(artifact_dir / "02-assumptions.yaml", assumptions)
    dump_document(artifact_dir / "03-spec.yaml", spec)
    dump_document(artifact_dir / "04-decision.yaml", decision)
    dump_document(artifact_dir / "05-plan.yaml", plan)
    dump_document(artifact_dir / "06-tasks.yaml", tasks)
    write_followup_view(target_dir, title, brief, assumptions, spec, plan, tasks)
    return target_dir


def write_compiled_pack(
    pack_id: str,
    output_dir: Path,
    work_unit_id: str,
    title: str,
    purpose: str,
    owner_actor_id: str,
    approver_actor_ids: list[str],
    stakeholder_actor_ids: list[str],
    branch_id: str,
    operator_actor_id: str,
    rollback_authority_actor_id: str,
    service_owner_actor_id: str,
    context_path: Path | None,
) -> Path:
    pack_map_index = pack_map()
    if pack_id not in pack_map_index:
        raise KeyError(f"Unknown pack_id {pack_id!r}")

    pack_dir, manifest = pack_map_index[pack_id]
    if pack_id == "research-prd":
        if context_path is None:
            context_path = pack_dir / "context" / "benchmark-context.md"
        if not context_path.exists():
            raise FileNotFoundError(f"Context file not found: {context_path}")

        context_reference = display_path(context_path)

        return write_compiled_research_prd_pack(
            manifest=manifest,
            pack_id=pack_id,
            output_dir=output_dir,
            work_unit_id=work_unit_id,
            title=title,
            purpose=purpose,
            owner_actor_id=owner_actor_id,
            approver_actor_ids=approver_actor_ids,
            stakeholder_actor_ids=stakeholder_actor_ids,
            branch_id=branch_id,
            context_reference=context_reference,
        )

    if pack_id == "travel-plan":
        if context_path is None:
            context_path = pack_dir / "context" / "benchmark-context.md"
        if not context_path.exists():
            raise FileNotFoundError(f"Context file not found: {context_path}")

        context_reference = display_path(context_path)

        return write_compiled_travel_plan_pack(
            manifest=manifest,
            pack_id=pack_id,
            output_dir=output_dir,
            work_unit_id=work_unit_id,
            title=title,
            purpose=purpose,
            owner_actor_id=owner_actor_id,
            approver_actor_ids=approver_actor_ids,
            stakeholder_actor_ids=stakeholder_actor_ids,
            branch_id=branch_id,
            context_reference=context_reference,
        )

    if pack_id == "budget-cycle":
        if context_path is None:
            context_path = pack_dir / "context" / "benchmark-context.md"
        if not context_path.exists():
            raise FileNotFoundError(f"Context file not found: {context_path}")

        context_reference = display_path(context_path)

        return write_compiled_budget_cycle_pack(
            manifest=manifest,
            pack_id=pack_id,
            output_dir=output_dir,
            work_unit_id=work_unit_id,
            title=title,
            purpose=purpose,
            owner_actor_id=owner_actor_id,
            approver_actor_ids=approver_actor_ids,
            stakeholder_actor_ids=stakeholder_actor_ids,
            branch_id=branch_id,
            context_reference=context_reference,
        )

    if pack_id == "incident-response":
        if context_path is None:
            context_path = pack_dir / "context" / "benchmark-context.md"
        if not context_path.exists():
            raise FileNotFoundError(f"Context file not found: {context_path}")

        context_reference = display_path(context_path)

        return write_compiled_incident_response_pack(
            manifest=manifest,
            pack_id=pack_id,
            output_dir=output_dir,
            work_unit_id=work_unit_id,
            title=title,
            purpose=purpose,
            owner_actor_id=owner_actor_id,
            approver_actor_ids=approver_actor_ids,
            stakeholder_actor_ids=stakeholder_actor_ids,
            branch_id=branch_id,
            context_reference=context_reference,
        )

    if pack_id == "compliance-audit":
        if context_path is None:
            context_path = pack_dir / "context" / "benchmark-context.md"
        if not context_path.exists():
            raise FileNotFoundError(f"Context file not found: {context_path}")

        context_reference = display_path(context_path)

        return write_compiled_compliance_audit_pack(
            manifest=manifest,
            pack_id=pack_id,
            output_dir=output_dir,
            work_unit_id=work_unit_id,
            title=title,
            purpose=purpose,
            owner_actor_id=owner_actor_id,
            approver_actor_ids=approver_actor_ids,
            stakeholder_actor_ids=stakeholder_actor_ids,
            branch_id=branch_id,
            context_reference=context_reference,
        )

    if pack_id == "vendor-selection":
        if context_path is None:
            context_path = pack_dir / "context" / "benchmark-context.md"
        if not context_path.exists():
            raise FileNotFoundError(f"Context file not found: {context_path}")

        context_reference = display_path(context_path)

        return write_compiled_vendor_selection_pack(
            manifest=manifest,
            pack_id=pack_id,
            output_dir=output_dir,
            work_unit_id=work_unit_id,
            title=title,
            purpose=purpose,
            owner_actor_id=owner_actor_id,
            approver_actor_ids=approver_actor_ids,
            stakeholder_actor_ids=stakeholder_actor_ids,
            branch_id=branch_id,
            context_reference=context_reference,
        )

    if pack_id == "meeting-followup":
        if context_path is None:
            context_path = pack_dir / "context" / "benchmark-context.md"
        if not context_path.exists():
            raise FileNotFoundError(f"Context file not found: {context_path}")

        context_reference = display_path(context_path)

        return write_compiled_meeting_followup_pack(
            manifest=manifest,
            pack_id=pack_id,
            output_dir=output_dir,
            work_unit_id=work_unit_id,
            title=title,
            purpose=purpose,
            owner_actor_id=owner_actor_id,
            approver_actor_ids=approver_actor_ids,
            stakeholder_actor_ids=stakeholder_actor_ids,
            branch_id=branch_id,
            context_reference=context_reference,
        )

    raise KeyError(f"compile-pack is not implemented for pack_id {pack_id!r}")


def main() -> int:
    parser = argparse.ArgumentParser(
        prog=cli_invocation(),
        description="Jini CLI for clearer next steps, less rework, and stronger handoffs",
    )
    parser.add_argument("--version", action="version", version=f"jini {load_version()}")
    subparsers = parser.add_subparsers(dest="command", required=True)

    list_parser = subparsers.add_parser("list-schemas", help="List registered schemas")
    list_parser.set_defaults(command="list-schemas")

    packs_parser = subparsers.add_parser("list-packs", help="List available workflow packs")
    packs_parser.set_defaults(command="list-packs")

    catalog_parser = subparsers.add_parser(
        "catalog-packs",
        help="Show the current pack catalog, profiles, extensions, and bootstrap modes",
    )
    catalog_parser.add_argument(
        "--format",
        choices=["text", "json"],
        default="text",
        help="Output format for the pack catalog",
    )

    bootstrap_home_parser = subparsers.add_parser(
        "bootstrap-home",
        help="Create a Jini personal OS home with memory, tools, and routine scaffolding",
    )
    bootstrap_home_parser.add_argument("path", type=Path)
    bootstrap_home_parser.add_argument("--owner-name", default="", help="Optional user name for user.md")
    bootstrap_home_parser.add_argument("--assistant-name", default="Jini", help="Assistant name for soul.md")
    bootstrap_home_parser.add_argument(
        "--format",
        choices=["text", "json"],
        default="text",
        help="Output format for bootstrap-home",
    )

    bootstrap_steering_parser = subparsers.add_parser(
        "bootstrap-steering",
        help="Create canonical Jini workspace steering docs for product, tech, structure, and testing context",
    )
    bootstrap_steering_parser.add_argument("path", type=Path)
    bootstrap_steering_parser.add_argument(
        "--format",
        choices=["text", "json"],
        default="text",
        help="Output format for bootstrap-steering",
    )

    append_memory_parser = subparsers.add_parser(
        "append-memory",
        help="Append one durable line to the daily memory log",
    )
    append_memory_parser.add_argument("path", type=Path)
    append_memory_parser.add_argument("--line", required=True, help="One durable memory line")
    append_memory_parser.add_argument("--date", help="Optional YYYY-MM-DD override for the daily file")
    append_memory_parser.add_argument(
        "--format",
        choices=["text", "json"],
        default="text",
        help="Output format for append-memory",
    )

    dream_memory_parser = subparsers.add_parser(
        "dream-memory",
        help="Compress daily memory logs into long-term memory",
    )
    dream_memory_parser.add_argument("path", type=Path)
    dream_memory_parser.add_argument(
        "--format",
        choices=["text", "json"],
        default="text",
        help="Output format for dream-memory",
    )

    memory_status_parser = subparsers.add_parser(
        "memory-status",
        help="Show bounded-memory usage, drift, and dream-memory recommendations for a Jini home",
    )
    memory_status_parser.add_argument("path", type=Path)
    memory_status_parser.add_argument(
        "--format",
        choices=["text", "json"],
        default="text",
        help="Output format for memory-status",
    )

    show_steering_parser = subparsers.add_parser(
        "show-steering",
        help="Show detected workspace steering docs and which ones are active by default",
    )
    show_steering_parser.add_argument("path", type=Path)
    show_steering_parser.add_argument("--file", type=Path, help="Optional file path to evaluate fileMatch activation against")
    show_steering_parser.add_argument(
        "--format",
        choices=["text", "json"],
        default="text",
        help="Output format for show-steering",
    )

    repo_map_parser = subparsers.add_parser(
        "repo-map",
        help="Emit a compact repo map optimized for delivery, planning, and low-token handoff",
    )
    repo_map_parser.add_argument("path", type=Path)
    repo_map_parser.add_argument("--repo", type=Path, help="Optional repo or worktree path; defaults to the provided path")
    repo_map_parser.add_argument("--max-entries", type=int, default=8, help="Maximum top-level files or directories to retain")
    repo_map_parser.add_argument(
        "--format",
        choices=["text", "json"],
        default="text",
        help="Output format for repo-map",
    )

    list_tools_parser = subparsers.add_parser(
        "list-tools",
        help="Show the operator-facing personal tool inventory for a Jini home",
    )
    list_tools_parser.add_argument("path", type=Path)
    list_tools_parser.add_argument(
        "--format",
        choices=["text", "json"],
        default="text",
        help="Output format for list-tools",
    )

    list_routines_parser = subparsers.add_parser(
        "list-routines",
        help="Show the local and remote routine catalog for a Jini home",
    )
    list_routines_parser.add_argument("path", type=Path)
    list_routines_parser.add_argument("--mode", choices=["local", "remote"], help="Optional routine mode filter")
    list_routines_parser.add_argument(
        "--format",
        choices=["text", "json"],
        default="text",
        help="Output format for list-routines",
    )

    run_routine_parser = subparsers.add_parser(
        "run-routine",
        help="Run a local routine or stage a remote routine from a Jini home",
    )
    run_routine_parser.add_argument("path", type=Path)
    run_routine_parser.add_argument("routine_id", help="Routine id to run")
    run_routine_parser.add_argument("--mode", choices=["local", "remote"], help="Optional mode assertion")
    run_routine_parser.add_argument(
        "--format",
        choices=["text", "json"],
        default="text",
        help="Output format for run-routine",
    )

    bind_home_parser = subparsers.add_parser(
        "bind-home",
        help="Bind a Jini personal home to a workflow pack",
    )
    bind_home_parser.add_argument("path", type=Path)
    bind_home_parser.add_argument("--home", required=True, type=Path, help="Path to a bootstrapped Jini home")
    bind_home_parser.add_argument(
        "--format",
        choices=["text", "json"],
        default="text",
        help="Output format for bind-home",
    )

    validate_parser = subparsers.add_parser("validate", help="Validate a single YAML or JSON file")
    validate_parser.add_argument("path", type=Path)
    validate_parser.add_argument("--schema", choices=["work-unit"], help="Explicit schema key")

    pack_parser = subparsers.add_parser("validate-pack", help="Validate a pack example directory")
    pack_parser.add_argument("path", type=Path)

    status_parser = subparsers.add_parser(
        "status-pack",
        help="Summarize pack readiness, required artifacts, and next protocol step",
    )
    status_parser.add_argument("path", type=Path)

    outcome_parser = subparsers.add_parser(
        "outcome",
        help="Answer what is done, what happens next, and what is still missing before the work is truly across the line",
    )
    outcome_parser.add_argument("path", type=Path)
    outcome_parser.add_argument("--repo", type=Path, help="Optional repo or worktree path for follow-on commands")
    outcome_parser.add_argument(
        "--format",
        choices=["text", "json"],
        default="text",
        help="Output format for the outcome view",
    )

    recommend_execution_parser = subparsers.add_parser(
        "recommend-execution",
        help="Recommend the execution class, context posture, and rate-limit strategy for a pack",
    )
    recommend_execution_parser.add_argument("path", type=Path)
    recommend_execution_parser.add_argument(
        "--intent",
        choices=["scope", "probe", "model", "decide", "make", "verify", "export", "issues", "wiki", "publish", "research"],
        help="Optional intent override; defaults to the pack's next operation",
    )
    recommend_execution_parser.add_argument(
        "--format",
        choices=["text", "json"],
        default="text",
        help="Output format for the recommendation",
    )
    recommend_execution_parser.add_argument(
        "--repo",
        type=Path,
        help="Optional repo or worktree path for repo-aware guidance",
    )
    recommend_execution_parser.add_argument("--home", type=Path, help="Optional personal home path")
    recommend_execution_parser.add_argument(
        "--runtime-target",
        "--harness",
        dest="runtime_target",
        help="Optional preferred harness",
    )

    kpi_parser = subparsers.add_parser(
        "show-kpis",
        help="Show the competitive KPI scorecard and next build actions",
    )
    kpi_parser.add_argument(
        "--dimension",
        help="Optional KPI id or label filter for a detailed view",
    )
    kpi_parser.add_argument(
        "--limit",
        type=int,
        default=5,
        help="Number of KPI rows to show in summary mode",
    )
    kpi_parser.add_argument(
        "--format",
        choices=["text", "json"],
        default="text",
        help="Output format for the scorecard",
    )

    publish_readiness_parser = subparsers.add_parser(
        "publish-readiness",
        help="Summarize whether Jini is lean, documented, and productized enough for wider publication",
    )
    publish_readiness_parser.add_argument(
        "--format",
        choices=["text", "json"],
        default="text",
        help="Output format for the publishing-readiness report",
    )

    golden_benchmark_parser = subparsers.add_parser(
        "validate-golden-benchmark",
        help="Validate Jini against the golden product benchmark and compare it to major competitors",
    )
    golden_benchmark_parser.add_argument(
        "--format",
        choices=["text", "json"],
        default="text",
        help="Output format for the golden benchmark report",
    )

    get_started_parser = subparsers.add_parser(
        "get-started",
        help="Show the beginner and power-user paths through the same Jini system",
    )
    get_started_parser.add_argument(
        "--target",
        "--harness",
        dest="target",
        help="Optional preferred harness for onboarding commands",
    )
    get_started_parser.add_argument(
        "--audience",
        choices=["beginner", "power-user", "both"],
        default="both",
        help="Which onboarding path to show",
    )
    get_started_parser.add_argument(
        "--format",
        choices=["text", "json"],
        default="text",
        help="Output format for the getting-started guide",
    )

    try_example_parser = subparsers.add_parser(
        "try-example",
        help="Show Jini on a common workflow without requiring pack assembly first",
    )
    try_example_parser.add_argument(
        "example_id",
        choices=sorted(PUBLIC_EXAMPLE_SPECS.keys()),
        help="Public example to materialize or inspect",
    )
    try_example_parser.add_argument(
        "--output",
        type=Path,
        help="Optional output directory for generated examples; bundled examples ignore this",
    )
    try_example_parser.add_argument(
        "--format",
        choices=["text", "json"],
        default="text",
        help="Output format for the example proof",
    )

    framework_review_parser = subparsers.add_parser(
        "review-framework",
        help="Critique the Jini framework against adoption constraints and competitor gaps",
    )
    framework_review_parser.add_argument("--dimension", help="Optional dimension filter")
    framework_review_parser.add_argument("--limit", type=int, default=5, help="Maximum dimensions to prioritize")
    framework_review_parser.add_argument(
        "--format",
        choices=["text", "json"],
        default="text",
        help="Output format for the framework review",
    )

    stage_framework_experiment_parser = subparsers.add_parser(
        "stage-framework-experiment",
        help="Stage a governed framework-evolution experiment from the latest or selected review",
    )
    stage_framework_experiment_parser.add_argument("--review", type=Path, help="Optional specific framework review path")
    stage_framework_experiment_parser.add_argument("--dimension", help="Optional dimension to stage")
    stage_framework_experiment_parser.add_argument("--index", type=int, default=1, help="1-based recommended experiment index")
    stage_framework_experiment_parser.add_argument(
        "--format",
        choices=["text", "json"],
        default="text",
        help="Output format for the staged experiment",
    )

    record_framework_outcome_parser = subparsers.add_parser(
        "record-framework-outcome",
        help="Record the outcome of a framework-evolution experiment and compute its reward",
    )
    record_framework_outcome_parser.add_argument("path", type=Path, help="Framework experiment path")
    record_framework_outcome_parser.add_argument("--actor", required=True, help="Actor recording the outcome")
    record_framework_outcome_parser.add_argument(
        "--result",
        required=True,
        choices=["success", "mixed", "failed"],
        help="Outcome classification",
    )
    record_framework_outcome_parser.add_argument("--score-delta", type=float, required=True, help="Observed score delta")
    record_framework_outcome_parser.add_argument("--signal", action="append", default=[], help="Adoption or product signal; may be repeated")
    record_framework_outcome_parser.add_argument("--note", action="append", default=[], help="Outcome note; may be repeated")
    record_framework_outcome_parser.add_argument(
        "--format",
        choices=["text", "json"],
        default="text",
        help="Output format for the recorded outcome",
    )

    backtest_framework_parser = subparsers.add_parser(
        "backtest-framework-evolution",
        help="Summarize framework-evolution outcomes and recommend the next focus dimension",
    )
    backtest_framework_parser.add_argument("--limit", type=int, default=100, help="Maximum number of outcomes to read")
    backtest_framework_parser.add_argument(
        "--format",
        choices=["text", "json"],
        default="text",
        help="Output format for the framework backtest",
    )

    checklist_parser = subparsers.add_parser(
        "execution-checklist",
        help="Build a concrete next-step checklist from pack state, repo context, and evidence posture",
    )
    checklist_parser.add_argument("path", type=Path)
    checklist_parser.add_argument("--intent", help="Optional intent override; defaults to the pack's next operation")
    checklist_parser.add_argument("--repo", type=Path, help="Optional repo or worktree path for repo-aware targeting")
    checklist_parser.add_argument("--home", type=Path, help="Optional personal home path")
    checklist_parser.add_argument(
        "--runtime-target",
        "--harness",
        dest="runtime_target",
        help="Optional preferred harness",
    )
    checklist_parser.add_argument(
        "--format",
        choices=["text", "json"],
        default="text",
        help="Output format for the checklist",
    )

    compact_context_parser = subparsers.add_parser(
        "compact-context",
        help="Emit a compact resume context slice for low-token reloads and memory-aware resumptions",
    )
    compact_context_parser.add_argument("path", type=Path)
    compact_context_parser.add_argument("--intent", help="Optional intent override; defaults to the pack's next operation")
    compact_context_parser.add_argument("--repo", type=Path, help="Optional repo or worktree path for repo-aware targeting")
    compact_context_parser.add_argument("--home", type=Path, help="Optional personal home path")
    compact_context_parser.add_argument(
        "--runtime-target",
        "--harness",
        dest="runtime_target",
        help="Optional preferred harness",
    )
    compact_context_parser.add_argument(
        "--max-items",
        type=int,
        default=5,
        help="Maximum items to keep per list inside the compact context",
    )
    compact_context_parser.add_argument(
        "--max-chars",
        type=int,
        default=1200,
        help="Soft character budget for the compact context payload",
    )
    compact_context_parser.add_argument(
        "--format",
        choices=["text", "json"],
        default="text",
        help="Output format for the compact context",
    )

    adapters_parser = subparsers.add_parser(
        "show-adapters",
        help="Show the current adapter registry, layers, capabilities, and maturity levels",
    )
    adapters_parser.add_argument("--capability", help="Optional capability filter, for example issues-publish-plan")
    adapters_parser.add_argument(
        "--format",
        choices=["text", "json"],
        default="text",
        help="Output format for the adapter summary",
    )

    adapter_conformance_parser = subparsers.add_parser(
        "adapter-conformance",
        help="Check the adapter registry against install shims and wired export surfaces",
    )
    adapter_conformance_parser.add_argument(
        "--format",
        choices=["text", "json"],
        default="text",
        help="Output format for adapter conformance",
    )

    resolve_adapter_parser = subparsers.add_parser(
        "resolve-adapter",
        help="Resolve the preferred adapter and fallbacks for a capability",
    )
    resolve_adapter_parser.add_argument("--capability", required=True, help="Capability to resolve")
    resolve_adapter_parser.add_argument("--layer", help="Optional layer filter, for example runtime-target")
    resolve_adapter_parser.add_argument("--preferred", help="Optional preferred adapter id")
    resolve_adapter_parser.add_argument(
        "--format",
        choices=["text", "json"],
        default="text",
        help="Output format for adapter resolution",
    )

    adapter_matrix_parser = subparsers.add_parser(
        "adapter-matrix",
        help="Show adapter coverage by layer and capability",
    )
    adapter_matrix_parser.add_argument(
        "--format",
        choices=["text", "json"],
        default="text",
        help="Output format for the adapter matrix",
    )

    harness_parser = subparsers.add_parser(
        "harnesses",
        help="Show the coding harnesses Jini can work above",
    )
    harness_parser.add_argument(
        "--format",
        choices=["text", "json"],
        default="text",
        help="Output format for the harness catalog",
    )

    events_parser = subparsers.add_parser(
        "show-learning-events",
        help="Show recent learning/runtime events for RL instrumentation and policy review",
    )
    events_parser.add_argument("path", nargs="?", type=Path, help="Optional pack path to read local runtime events")
    events_parser.add_argument("--event-type", help="Optional event type filter")
    events_parser.add_argument("--work-unit-id", help="Optional work unit filter")
    events_parser.add_argument("--limit", type=int, default=20, help="Maximum number of events to return")
    events_parser.add_argument(
        "--format",
        choices=["text", "json"],
        default="text",
        help="Output format for learning events",
    )

    learning_snapshot_parser = subparsers.add_parser(
        "learning-snapshot",
        help="Summarize runtime learning events for offline evaluation and RL instrumentation review",
    )
    learning_snapshot_parser.add_argument("path", nargs="?", type=Path, help="Optional pack path to read local runtime events")
    learning_snapshot_parser.add_argument("--limit", type=int, default=200, help="Maximum number of events to summarize")
    learning_snapshot_parser.add_argument(
        "--format",
        choices=["text", "json"],
        default="text",
        help="Output format for the learning snapshot",
    )

    routing_backtest_parser = subparsers.add_parser(
        "routing-backtest",
        help="Summarize runtime learning events into offline routing and execution-class recommendations",
    )
    routing_backtest_parser.add_argument("path", nargs="?", type=Path, help="Optional pack path to read local runtime events")
    routing_backtest_parser.add_argument("--limit", type=int, default=200, help="Maximum number of events to backtest")
    routing_backtest_parser.add_argument("--min-samples", type=int, default=1, help="Minimum samples required per bucket")
    routing_backtest_parser.add_argument(
        "--format",
        choices=["text", "json"],
        default="text",
        help="Output format for the routing backtest",
    )

    handoff_parser = subparsers.add_parser(
        "stage-runtime-handoff",
        help="Persist a runtime-ready handoff bundle with compact context, checklist, adapter target, and install preview",
    )
    handoff_parser.add_argument("path", type=Path)
    handoff_parser.add_argument("--intent", help="Optional intent override; defaults to the pack's next operation")
    handoff_parser.add_argument("--repo", type=Path, help="Optional repo or worktree path for repo-aware targeting")
    handoff_parser.add_argument("--home", type=Path, help="Optional personal home path")
    handoff_parser.add_argument(
        "--runtime-target",
        "--harness",
        dest="runtime_target",
        help="Optional preferred harness",
    )
    handoff_parser.add_argument(
        "--max-items",
        type=int,
        default=5,
        help="Maximum items to keep per list inside the embedded compact context",
    )
    handoff_parser.add_argument(
        "--max-chars",
        type=int,
        default=1200,
        help="Soft character budget for the embedded compact context payload",
    )
    handoff_parser.add_argument(
        "--format",
        choices=["text", "json"],
        default="text",
        help="Output format for the runtime handoff bundle",
    )

    activate_runtime_parser = subparsers.add_parser(
        "activate-runtime-target",
        help="Install the selected runtime shim and materialize a real local activation bundle from a runtime handoff",
    )
    activate_runtime_parser.add_argument("path", type=Path)
    activate_runtime_parser.add_argument("--handoff", type=Path, help="Optional existing handoff bundle to activate")
    activate_runtime_parser.add_argument("--intent", help="Optional intent override when building a handoff on demand")
    activate_runtime_parser.add_argument("--repo", type=Path, help="Optional repo or worktree path for repo-aware activation")
    activate_runtime_parser.add_argument("--home", type=Path, help="Optional personal home path")
    activate_runtime_parser.add_argument(
        "--runtime-target",
        "--harness",
        dest="runtime_target",
        help="Optional preferred harness",
    )
    activate_runtime_parser.add_argument("--prefix", type=Path, help="Optional prefix for safe local activation testing")
    activate_runtime_parser.add_argument(
        "--max-items",
        type=int,
        default=5,
        help="Maximum items to keep per list inside an on-demand compact context",
    )
    activate_runtime_parser.add_argument(
        "--max-chars",
        type=int,
        default=1200,
        help="Soft character budget for an on-demand compact context",
    )
    activate_runtime_parser.add_argument(
        "--format",
        choices=["text", "json"],
        default="text",
        help="Output format for the runtime activation receipt",
    )

    execute_flow_parser = subparsers.add_parser(
        "execute-flow",
        help="Run the guided Jini execution loop with compact context reuse, optional runtime activation, and local adapter apply",
    )
    execute_flow_parser.add_argument("path", type=Path)
    execute_flow_parser.add_argument(
        "--mode",
        choices=["supervised", "autonomous"],
        default="supervised",
        help="Execution mode for the embedded run-pack step",
    )
    execute_flow_parser.add_argument("--intent", help="Optional runtime intent override")
    execute_flow_parser.add_argument("--repo", type=Path, help="Optional repo or worktree path for repo-aware execution")
    execute_flow_parser.add_argument("--home", type=Path, help="Optional personal home path")
    execute_flow_parser.add_argument(
        "--runtime-target",
        "--harness",
        dest="runtime_target",
        help="Optional preferred harness",
    )
    execute_flow_parser.add_argument(
        "--consent",
        action="append",
        default=[],
        choices=list(RUNTIME_CONSENT_CATEGORIES),
        help="Grant and persist consent for an action category; may be repeated",
    )
    execute_flow_parser.add_argument(
        "--issue-adapter",
        choices=["jira", "github"],
        default="jira",
        help="Issue adapter for the embedded publish step",
    )
    execute_flow_parser.add_argument(
        "--wiki-adapter",
        choices=["confluence", "markdown"],
        default="confluence",
        help="Wiki adapter for the embedded publish step",
    )
    execute_flow_parser.add_argument("--project-key", help="Optional Jira project key for the embedded publish step")
    execute_flow_parser.add_argument("--space-key", help="Optional Confluence space key for the embedded publish step")
    execute_flow_parser.add_argument(
        "--activate-runtime",
        action="store_true",
        help="Install and materialize the selected runtime target before running the flow",
    )
    execute_flow_parser.add_argument("--prefix", type=Path, help="Optional safe prefix for runtime activation")
    execute_flow_parser.add_argument("--author", default="jini-flow", help="Actor id used for auto-harvest events")
    execute_flow_parser.add_argument(
        "--max-items",
        type=int,
        default=5,
        help="Maximum items to keep per list inside the compact context",
    )
    execute_flow_parser.add_argument(
        "--max-chars",
        type=int,
        default=900,
        help="Soft character budget for the compact context payload",
    )
    execute_flow_parser.add_argument(
        "--format",
        choices=["text", "json"],
        default="text",
        help="Output format for the guided flow report",
    )

    policy_review_parser = subparsers.add_parser(
        "review-policy",
        help="Summarize learning traces into guarded, non-mutating policy candidates",
    )
    policy_review_parser.add_argument("path", nargs="?", type=Path, help="Optional pack path to review local runtime events")
    policy_review_parser.add_argument("--limit", type=int, default=200, help="Maximum number of events to review")
    policy_review_parser.add_argument("--min-samples", type=int, default=1, help="Minimum samples required per routing bucket")
    policy_review_parser.add_argument(
        "--format",
        choices=["text", "json"],
        default="text",
        help="Output format for the policy review",
    )

    stage_policy_parser = subparsers.add_parser(
        "stage-policy-candidate",
        help="Turn the latest pack-local policy review into a governed policy candidate artifact",
    )
    stage_policy_parser.add_argument("path", type=Path, help="Pack path used to locate review reports and store candidates")
    stage_policy_parser.add_argument("--review", type=Path, help="Optional specific review report path")
    stage_policy_parser.add_argument(
        "--format",
        choices=["text", "json"],
        default="text",
        help="Output format for the staged policy candidate",
    )

    approve_policy_parser = subparsers.add_parser(
        "approve-policy-candidate",
        help="Approve and activate a pack-local policy candidate with rollback-safe metadata",
    )
    approve_policy_parser.add_argument("path", type=Path, help="Pack path that owns the candidate and rollout")
    approve_policy_parser.add_argument("candidate", type=Path, help="Candidate path returned by stage-policy-candidate")
    approve_policy_parser.add_argument("--approver", required=True, help="Approver actor id")
    approve_policy_parser.add_argument(
        "--format",
        choices=["text", "json"],
        default="text",
        help="Output format for the activated rollout",
    )

    rollback_policy_parser = subparsers.add_parser(
        "rollback-policy-candidate",
        help="Roll back an active pack-local policy rollout and record the rollback receipt",
    )
    rollback_policy_parser.add_argument("path", type=Path, help="Pack path that owns the candidate and rollout")
    rollback_policy_parser.add_argument("candidate", type=Path, help="Candidate path returned by stage-policy-candidate")
    rollback_policy_parser.add_argument("--actor", required=True, help="Actor id performing the rollback")
    rollback_policy_parser.add_argument("--reason", required=True, help="Rollback reason")
    rollback_policy_parser.add_argument(
        "--format",
        choices=["text", "json"],
        default="text",
        help="Output format for the rollback receipt",
    )

    catalog_install_parser = subparsers.add_parser(
        "catalog-bundles",
        help="Show installable Jini bundles and curated kits from the install manifest",
    )
    catalog_install_parser.add_argument("--target", "--harness", dest="target", help="Optional harness filter")
    catalog_install_parser.add_argument("--kind", help="Optional bundle kind filter")
    catalog_install_parser.add_argument(
        "--format",
        choices=["text", "json"],
        default="text",
        help="Output format for the install catalog",
    )

    install_parser = subparsers.add_parser(
        "plan-install",
        help="Preview Jini bundle installation across supported targets without writing files",
    )
    install_parser.add_argument(
        "--kit",
        action="append",
        default=[],
        help="Curated install kit id to include; may be repeated.",
    )
    install_parser.add_argument(
        "--bundle",
        action="append",
        default=[],
        help="Bundle id to include; may be repeated. When omitted, the manifest default kit is used.",
    )
    install_parser.add_argument(
        "--target",
        action="append",
        default=[],
        help="Target id to include; may be repeated. Defaults to all targets.",
    )
    install_parser.add_argument(
        "--link-mode",
        choices=["auto", "copy", "symlink"],
        default="auto",
        help="Preview the install with the selected link strategy",
    )
    install_parser.add_argument(
        "--format",
        choices=["text", "json"],
        default="text",
        help="Output format for the install plan",
    )
    install_parser.add_argument(
        "--prefix",
        type=Path,
        help="Optional prefix to remap target destinations for testing or local staging",
    )

    install_exec_parser = subparsers.add_parser(
        "install-bundles",
        help="Install selected Jini bundles for selected targets from the manifest",
    )
    install_exec_parser.add_argument("--kit", action="append", default=[], help="Install kit id to include; may be repeated")
    install_exec_parser.add_argument("--bundle", action="append", default=[], help="Bundle id to install; may be repeated")
    install_exec_parser.add_argument(
        "--target",
        "--harness",
        dest="target",
        action="append",
        default=[],
        help="Harness id to install for; may be repeated",
    )
    install_exec_parser.add_argument(
        "--link-mode",
        choices=["auto", "copy", "symlink"],
        default="auto",
        help="Installation link strategy",
    )
    install_exec_parser.add_argument(
        "--prefix",
        type=Path,
        help="Optional prefix to remap target destinations for testing or local staging",
    )
    install_exec_parser.add_argument(
        "--format",
        choices=["text", "json"],
        default="text",
        help="Output format for the install result",
    )

    update_parser = subparsers.add_parser(
        "update-bundles",
        help="Refresh selected Jini bundle installations from source",
    )
    update_parser.add_argument("--kit", action="append", default=[], help="Install kit id to update; may be repeated")
    update_parser.add_argument("--bundle", action="append", default=[], help="Bundle id to update; may be repeated")
    update_parser.add_argument(
        "--target",
        "--harness",
        dest="target",
        action="append",
        default=[],
        help="Harness id to update; may be repeated",
    )
    update_parser.add_argument(
        "--link-mode",
        choices=["auto", "copy", "symlink"],
        default="auto",
        help="Installation link strategy",
    )
    update_parser.add_argument(
        "--prefix",
        type=Path,
        help="Optional prefix to remap target destinations for testing or local staging",
    )
    update_parser.add_argument(
        "--format",
        choices=["text", "json"],
        default="text",
        help="Output format for the update result",
    )

    uninstall_parser = subparsers.add_parser(
        "uninstall-bundles",
        help="Remove selected Jini bundle installations for selected targets",
    )
    uninstall_parser.add_argument("--kit", action="append", default=[], help="Install kit id to remove; may be repeated")
    uninstall_parser.add_argument("--bundle", action="append", default=[], help="Bundle id to remove; may be repeated")
    uninstall_parser.add_argument(
        "--target",
        "--harness",
        dest="target",
        action="append",
        default=[],
        help="Harness id to remove; may be repeated",
    )
    uninstall_parser.add_argument(
        "--prefix",
        type=Path,
        help="Optional prefix to remap target destinations for testing or local staging",
    )
    uninstall_parser.add_argument(
        "--format",
        choices=["text", "json"],
        default="text",
        help="Output format for the uninstall result",
    )

    doctor_parser = subparsers.add_parser(
        "doctor-install",
        help="Verify whether selected bundle installations exist and match the install manifest",
    )
    doctor_parser.add_argument("--kit", action="append", default=[], help="Install kit id to inspect; may be repeated")
    doctor_parser.add_argument("--bundle", action="append", default=[], help="Bundle id to inspect; may be repeated")
    doctor_parser.add_argument(
        "--target",
        "--harness",
        dest="target",
        action="append",
        default=[],
        help="Harness id to inspect; may be repeated",
    )
    doctor_parser.add_argument(
        "--prefix",
        type=Path,
        help="Optional prefix to remap target destinations for testing or local staging",
    )
    doctor_parser.add_argument(
        "--format",
        choices=["text", "json"],
        default="text",
        help="Output format for the doctor report",
    )

    run_pack_parser = subparsers.add_parser(
        "run-pack",
        help="Execute the next deterministic local Jini actions in supervised or autonomous mode",
    )
    run_pack_parser.add_argument("path", type=Path)
    run_pack_parser.add_argument(
        "--mode",
        choices=["supervised", "autonomous"],
        default="supervised",
        help="Execution mode; autonomous may advance state when consent and guards allow",
    )
    run_pack_parser.add_argument(
        "--intent",
        help="Optional runtime intent override; defaults to the pack's next operation",
    )
    run_pack_parser.add_argument(
        "--consent",
        action="append",
        default=[],
        choices=list(RUNTIME_CONSENT_CATEGORIES),
        help="Grant and persist consent for an action category; may be repeated",
    )
    run_pack_parser.add_argument(
        "--issue-adapter",
        choices=["jira", "github"],
        default="jira",
        help="Issue export adapter for local execution surfaces",
    )
    run_pack_parser.add_argument(
        "--wiki-adapter",
        choices=["confluence", "markdown"],
        default="confluence",
        help="Wiki export adapter; Confluence falls back to markdown when unconfigured",
    )
    run_pack_parser.add_argument("--project-key", help="Optional Jira project key for staged publish plans")
    run_pack_parser.add_argument("--space-key", help="Optional Confluence space key for staged wiki publish plans")
    run_pack_parser.add_argument(
        "--format",
        choices=["text", "json"],
        default="text",
        help="Output format for the run report",
    )
    run_pack_parser.add_argument(
        "--repo",
        type=Path,
        help="Optional repo or worktree path for repo-aware guidance",
    )
    run_pack_parser.add_argument("--home", type=Path, help="Optional personal home path")
    run_pack_parser.add_argument(
        "--runtime-target",
        "--harness",
        dest="runtime_target",
        help="Optional preferred harness",
    )

    bind_atlassian_parser = subparsers.add_parser(
        "bind-atlassian",
        help="Persist Atlassian site, Jira project, and Confluence space targets for a pack",
    )
    bind_atlassian_parser.add_argument("path", type=Path)
    bind_atlassian_parser.add_argument("--cloud-id", required=True, help="Atlassian cloud id")
    bind_atlassian_parser.add_argument("--site-url", required=True, help="Atlassian site url")
    bind_atlassian_parser.add_argument("--project-key", help="Default Jira project key")
    bind_atlassian_parser.add_argument("--space-key", help="Default Confluence space key")
    bind_atlassian_parser.add_argument("--space-id", help="Optional Confluence space id")

    show_atlassian_parser = subparsers.add_parser(
        "show-atlassian",
        help="Show the persisted Atlassian target binding for a pack",
    )
    show_atlassian_parser.add_argument("path", type=Path)
    show_atlassian_parser.add_argument(
        "--format",
        choices=["text", "json"],
        default="text",
        help="Output format for the binding",
    )

    export_tasks_parser = subparsers.add_parser(
        "export-tasks",
        help="Render a practical markdown task board from the canonical Tasks artifact",
    )
    export_tasks_parser.add_argument("path", type=Path)
    export_tasks_parser.add_argument("--output", type=Path, help="Optional output markdown path")

    sync_tasks_parser = subparsers.add_parser(
        "sync-tasks",
        help="Export a neutral structured task sync payload from the canonical Tasks artifact",
    )
    sync_tasks_parser.add_argument("path", type=Path)
    sync_tasks_parser.add_argument("--output", type=Path, help="Optional output JSON path")

    export_issues_parser = subparsers.add_parser(
        "export-issues",
        help="Export issue-system bundles from canonical Tasks via the sync payload",
    )
    export_issues_parser.add_argument("path", type=Path)
    export_issues_parser.add_argument(
        "--adapter",
        choices=["github", "jira"],
        default="jira",
        help="Issue-system adapter to render",
    )
    export_issues_parser.add_argument("--output", type=Path, help="Optional output directory")

    export_wiki_parser = subparsers.add_parser(
        "export-wiki",
        help="Export a wiki/documentation bundle from canonical artifacts and rendered views",
    )
    export_wiki_parser.add_argument("path", type=Path)
    export_wiki_parser.add_argument(
        "--adapter",
        choices=["confluence", "markdown"],
        default="confluence",
        help="Wiki-system adapter to render",
    )
    export_wiki_parser.add_argument("--output", type=Path, help="Optional output directory")

    publish_issues_parser = subparsers.add_parser(
        "publish-issues",
        help="Stage a serialized, idempotent issue publish plan from canonical task state",
    )
    publish_issues_parser.add_argument("path", type=Path)
    publish_issues_parser.add_argument(
        "--adapter",
        choices=["jira", "github"],
        default="jira",
        help="Issue-system target to stage for",
    )
    publish_issues_parser.add_argument("--project-key", help="Optional Jira project key for the staged publish plan")
    publish_issues_parser.add_argument("--cloud-id", help="Optional Atlassian cloud id override")
    publish_issues_parser.add_argument("--site-url", help="Optional Atlassian site url override")
    publish_issues_parser.add_argument("--output", type=Path, help="Optional output directory")
    publish_issues_parser.add_argument(
        "--apply-local",
        action="store_true",
        help="Immediately apply local-capable publish plans after staging",
    )
    publish_issues_parser.add_argument(
        "--bridge-runner",
        type=Path,
        help="Optional bridge command to execute the staged publish plan immediately",
    )
    publish_issues_parser.add_argument(
        "--bridge-timeout-seconds",
        type=int,
        default=15,
        help="Per-item timeout for bridge execution when --bridge-runner is provided",
    )
    publish_issues_parser.add_argument(
        "--format",
        choices=["text", "json"],
        default="text",
        help="Output format for the staged or applied publish receipt",
    )

    apply_publish_plan_parser = subparsers.add_parser(
        "apply-publish-plan",
        help="Apply a staged local publish plan when the adapter supports a local markdown target",
    )
    apply_publish_plan_parser.add_argument("path", type=Path)
    apply_publish_plan_parser.add_argument("--output", type=Path, help="Optional output directory override")
    apply_publish_plan_parser.add_argument(
        "--format",
        choices=["text", "json"],
        default="text",
        help="Output format for the apply receipt",
    )

    execute_publish_plan_parser = subparsers.add_parser(
        "execute-publish-plan",
        help="Execute a staged publish plan through a bridge command and emit a portable publication result bundle",
    )
    execute_publish_plan_parser.add_argument("path", type=Path)
    execute_publish_plan_parser.add_argument("--runner", type=Path, required=True, help="Executable bridge command")
    execute_publish_plan_parser.add_argument("--output", type=Path, help="Optional output directory override")
    execute_publish_plan_parser.add_argument(
        "--timeout-seconds",
        type=int,
        default=15,
        help="Per-item timeout for bridge execution",
    )
    execute_publish_plan_parser.add_argument(
        "--format",
        choices=["text", "json"],
        default="text",
        help="Output format for the bridge execution receipt",
    )

    publish_wiki_parser = subparsers.add_parser(
        "publish-wiki",
        help="Stage a serialized wiki publish plan and fall back to markdown when Confluence is unavailable",
    )
    publish_wiki_parser.add_argument("path", type=Path)
    publish_wiki_parser.add_argument(
        "--adapter",
        choices=["confluence", "markdown"],
        default="confluence",
        help="Wiki-system target to stage for",
    )
    publish_wiki_parser.add_argument("--space-key", help="Optional Confluence space key for the staged publish plan")
    publish_wiki_parser.add_argument("--space-id", help="Optional Confluence space id override")
    publish_wiki_parser.add_argument("--cloud-id", help="Optional Atlassian cloud id override")
    publish_wiki_parser.add_argument("--site-url", help="Optional Atlassian site url override")
    publish_wiki_parser.add_argument("--output", type=Path, help="Optional output directory")
    publish_wiki_parser.add_argument(
        "--apply-local",
        action="store_true",
        help="Immediately apply local-capable publish plans after staging",
    )
    publish_wiki_parser.add_argument(
        "--bridge-runner",
        type=Path,
        help="Optional bridge command to execute the staged publish plan immediately",
    )
    publish_wiki_parser.add_argument(
        "--bridge-timeout-seconds",
        type=int,
        default=15,
        help="Per-item timeout for bridge execution when --bridge-runner is provided",
    )
    publish_wiki_parser.add_argument(
        "--format",
        choices=["text", "json"],
        default="text",
        help="Output format for the staged or applied publish receipt",
    )

    evidence_parser = subparsers.add_parser(
        "capture-evidence",
        help="Create or update the canonical Evidence artifact for a pack",
    )
    evidence_parser.add_argument("path", type=Path)
    evidence_parser.add_argument("--author", required=True, help="Actor id creating the evidence")
    evidence_parser.add_argument("--approver", action="append", default=None, help="Approver actor id; may be repeated")
    evidence_parser.add_argument("--target-artifact", help="Artifact type to validate; defaults to Spec")
    evidence_parser.add_argument("--claim", action="append", default=[], help="Validated claim; may be repeated")
    evidence_parser.add_argument("--test-result", action="append", default=[], help="Test result; may be repeated")
    evidence_parser.add_argument("--review-result", action="append", default=[], help="Review result; may be repeated")
    evidence_parser.add_argument("--operational-result", action="append", default=[], help="Operational result; may be repeated")
    evidence_parser.add_argument("--risk", action="append", default=[], help="Residual risk; may be repeated")
    evidence_parser.add_argument("--reference", action="append", default=[], help="Additional reference id or path; may be repeated")
    evidence_parser.add_argument(
        "--status",
        default="reviewed",
        choices=["draft", "reviewed", "approved", "superseded", "invalidated", "merged", "archived"],
        help="Artifact status for the captured evidence",
    )

    harvest_parser = subparsers.add_parser(
        "harvest-evidence",
        help="Run bounded local repo verification checks and capture the results as canonical Evidence",
    )
    harvest_parser.add_argument("path", type=Path)
    harvest_parser.add_argument("--author", required=True, help="Actor id creating the evidence")
    harvest_parser.add_argument("--repo", type=Path, help="Optional repo or worktree path for bounded verification")
    harvest_parser.add_argument("--home", type=Path, help="Optional personal home path")
    harvest_parser.add_argument("--approver", action="append", default=None, help="Approver actor id; may be repeated")
    harvest_parser.add_argument("--target-artifact", help="Artifact type to validate; defaults to Spec")
    harvest_parser.add_argument(
        "--category",
        action="append",
        default=[],
        choices=list(HARVEST_CATEGORY_ORDER),
        help="Repo verification category to harvest; may be repeated",
    )
    harvest_parser.add_argument("--claim", action="append", default=[], help="Validated claim; may be repeated")
    harvest_parser.add_argument("--risk", action="append", default=[], help="Residual risk; may be repeated")
    harvest_parser.add_argument("--reference", action="append", default=[], help="Additional reference id or path; may be repeated")
    harvest_parser.add_argument(
        "--timeout-seconds",
        type=int,
        default=20,
        help="Per-command timeout for harvested verification checks",
    )
    harvest_parser.add_argument(
        "--max-targets",
        type=int,
        default=5,
        help="Maximum number of repo verification targets to harvest",
    )
    harvest_parser.add_argument(
        "--status",
        default="reviewed",
        choices=["draft", "reviewed", "approved", "superseded", "invalidated", "merged", "archived"],
        help="Requested artifact status; downgraded to draft when harvested proof is not ready",
    )
    harvest_parser.add_argument(
        "--format",
        choices=["text", "json"],
        default="text",
        help="Output format for the harvest report",
    )

    output_parser = subparsers.add_parser(
        "capture-output",
        help="Record execution output against a specific task and refresh the task board view",
    )
    output_parser.add_argument("path", type=Path)
    output_parser.add_argument("--author", required=True, help="Actor id capturing the output")
    output_parser.add_argument("--task-index", type=int, required=True, help="1-based task index from the Tasks artifact")
    output_parser.add_argument("--status", default="done", help="Task status to record")
    output_parser.add_argument("--note", required=True, help="Human-readable output note")
    output_parser.add_argument("--reference", action="append", default=[], help="Reference path or id; may be repeated")
    output_parser.add_argument("--deliverable", help="Optional replacement deliverable text for the task")

    approval_parser = subparsers.add_parser(
        "capture-approval",
        help="Create or update the canonical Approval artifact for a pack",
    )
    approval_parser.add_argument("path", type=Path)
    approval_parser.add_argument("--author", required=True, help="Actor id recording the approval artifact")
    approval_parser.add_argument("--approver-actor", required=True, help="Actor id granting approval")
    approval_parser.add_argument("--scope", required=True, help="Approval scope, for example operational-readiness")
    approval_parser.add_argument("--target-artifact", help="Artifact type to approve; defaults to Evidence")
    approval_parser.add_argument("--waiver", action="append", default=[], help="Waiver note; may be repeated")
    approval_parser.add_argument("--condition", action="append", default=[], help="Condition note; may be repeated")
    approval_parser.add_argument(
        "--status",
        default="approved",
        choices=["draft", "reviewed", "approved", "superseded", "invalidated", "merged", "archived"],
        help="Artifact status for the captured approval",
    )

    publication_parser = subparsers.add_parser(
        "capture-publication",
        help="Create or update the canonical Publication artifact from a publish-result bundle",
    )
    publication_parser.add_argument("path", type=Path)
    publication_parser.add_argument("--author", required=True, help="Actor id recording the publication result")
    publication_parser.add_argument("--input", type=Path, required=True, help="JSON bundle containing publication records")
    publication_parser.add_argument("--scope", required=True, help="Publication scope, for example jira-issues or confluence-pages")
    publication_parser.add_argument("--approver", action="append", default=None, help="Approver actor id; may be repeated")
    publication_parser.add_argument(
        "--status",
        default="reviewed",
        choices=["draft", "reviewed", "approved", "superseded", "invalidated", "merged", "archived"],
        help="Artifact status for the captured publication",
    )

    advance_parser = subparsers.add_parser(
        "advance-pack",
        help="Advance a pack to its next legal linear state when required artifacts are ready",
    )
    advance_parser.add_argument("path", type=Path)
    advance_parser.add_argument(
        "--to",
        choices=LINEAR_STATE_ORDER[1:],
        help="Optional explicit target state; must be the immediate next state",
    )

    init_parser = subparsers.add_parser("init-pack", help="Create a new WorkUnit scaffold from a pack")
    init_parser.add_argument("pack_id", help="Pack id, for example research-prd")
    init_parser.add_argument("--work-unit-id", required=True, help="Canonical work unit id")
    init_parser.add_argument("--title", required=True, help="Human-readable title")
    init_parser.add_argument("--purpose", required=True, help="Short purpose/objective")
    init_parser.add_argument("--owner", required=True, help="Owner actor id")
    init_parser.add_argument("--approver", action="append", default=[], help="Approver actor id; may be repeated")
    init_parser.add_argument("--stakeholder", action="append", default=[], help="Stakeholder actor id; may be repeated")
    init_parser.add_argument("--branch", default="main", help="Branch id to embed in generated artifacts")
    init_parser.add_argument("--output", type=Path, required=True, help="Target output directory")

    compile_parser = subparsers.add_parser(
        "compile-pack",
        help="Compile a fuller, stateful pack scaffold from a pack context",
    )
    compile_parser.add_argument("pack_id", help="Pack id, for example research-prd")
    compile_parser.add_argument("--work-unit-id", required=True, help="Canonical work unit id")
    compile_parser.add_argument("--title", required=True, help="Human-readable title")
    compile_parser.add_argument("--purpose", required=True, help="Short purpose/objective")
    compile_parser.add_argument("--owner", required=True, help="Owner actor id")
    compile_parser.add_argument("--approver", action="append", default=[], help="Approver actor id; may be repeated")
    compile_parser.add_argument("--stakeholder", action="append", default=[], help="Stakeholder actor id; may be repeated")
    compile_parser.add_argument("--branch", default="main", help="Branch id to embed in generated artifacts")
    compile_parser.add_argument("--operator", help="Operator actor id; defaults to owner")
    compile_parser.add_argument("--rollback-authority", help="Rollback authority actor id; defaults to first approver or owner")
    compile_parser.add_argument("--service-owner", help="Service owner actor id; defaults to owner")
    compile_parser.add_argument("--context", type=Path, help="Optional context file override")
    compile_parser.add_argument("--output", type=Path, required=True, help="Target output directory")

    bootstrap_parser = subparsers.add_parser(
        "bootstrap-pack",
        help="Use the learned bootstrap policy to choose init-pack or compile-pack",
    )
    bootstrap_parser.add_argument("pack_id", help="Pack id, for example research-prd")
    bootstrap_parser.add_argument("--work-unit-id", required=True, help="Canonical work unit id")
    bootstrap_parser.add_argument("--title", required=True, help="Human-readable title")
    bootstrap_parser.add_argument("--purpose", required=True, help="Short purpose/objective")
    bootstrap_parser.add_argument("--owner", required=True, help="Owner actor id")
    bootstrap_parser.add_argument("--approver", action="append", default=[], help="Approver actor id; may be repeated")
    bootstrap_parser.add_argument("--stakeholder", action="append", default=[], help="Stakeholder actor id; may be repeated")
    bootstrap_parser.add_argument("--branch", default="main", help="Branch id to embed in generated artifacts")
    bootstrap_parser.add_argument("--operator", help="Operator actor id; defaults to owner")
    bootstrap_parser.add_argument("--rollback-authority", help="Rollback authority actor id; defaults to first approver or owner")
    bootstrap_parser.add_argument("--service-owner", help="Service owner actor id; defaults to owner")
    bootstrap_parser.add_argument("--context", type=Path, help="Optional context file override")
    bootstrap_parser.add_argument(
        "--mode",
        choices=["auto", "init-pack", "compile-pack"],
        default="auto",
        help="Override the learned bootstrap mode",
    )
    bootstrap_parser.add_argument("--output", type=Path, required=True, help="Target output directory")

    argv = sys.argv[1:]
    if not argv:
        print_cli_overview()
        return 0
    if argv[0] in {"-h", "--help"}:
        if "--all" in argv[1:]:
            parser.print_help()
        else:
            print_cli_overview()
        return 0
    if argv[0] == "help":
        if "--all" in argv[1:]:
            parser.print_help()
        else:
            print_cli_overview()
        return 0

    args = parser.parse_args(normalize_cli_argv(argv))
    registry = load_registry()

    def merged_stakeholders(owner: str, stakeholders: list[str]) -> list[str]:
        ordered = [owner, *stakeholders]
        return list(dict.fromkeys(value for value in ordered if value))

    if args.command == "list-schemas":
        print("work-unit ->", registry["work_unit"]["schema"])
        for canonical, meta in registry["artifacts"].items():
            print(f"{canonical} -> {meta['schema']}")
        return 0

    if args.command == "list-packs":
        packs = list_packs()
        if not packs:
            print("No packs found")
            return 0
        for pack_id, pack_dir, manifest in packs:
            print(f"{pack_id} -> {pack_dir.relative_to(ROOT)}")
            print(f"  profile: {manifest.get('target_profile', '')}")
            print(f"  flow: {' -> '.join(manifest.get('compiled_flow', []))}")
        return 0

    if args.command == "catalog-packs":
        catalog = build_pack_catalog()
        if args.format == "json":
            print(json.dumps(catalog, indent=2))
        else:
            print_pack_catalog(catalog)
        return 0

    if args.command == "bootstrap-home":
        try:
            result = bootstrap_home(
                args.path,
                owner_name=args.owner_name,
                assistant_name=args.assistant_name,
            )
        except (FileExistsError, ValueError) as exc:
            print(f"ERROR {exc}")
            return 1
        if args.format == "json":
            print(json.dumps(result, indent=2))
        else:
            print(f"HOME   {result['home_root']}")
            for path in result["created_paths"]:
                print(f"  - {path}")
        return 0

    if args.command == "bootstrap-steering":
        try:
            result = bootstrap_workspace_steering(args.path)
        except (FileExistsError, ValueError, OSError) as exc:
            print(f"ERROR {exc}")
            return 1
        if args.format == "json":
            print(json.dumps(result, indent=2))
        else:
            print(f"WORKSPACE {result['workspace_root']}")
            print(f"STEERING  {result['steering_root']}")
            for path in result["created_paths"]:
                print(f"  - {path}")
        return 0

    if args.command == "append-memory":
        try:
            result = append_memory_line(args.path, line=args.line, date_text=args.date)
        except ValueError as exc:
            print(f"ERROR {exc}")
            return 1
        if args.format == "json":
            print(json.dumps(result, indent=2))
        else:
            print(f"DATE   {result['date']}")
            print(f"PATH   {result['daily_path']}")
            print(f"LINE   {result['line']}")
        return 0

    if args.command == "dream-memory":
        try:
            result = dream_memory(args.path)
        except ValueError as exc:
            print(f"ERROR {exc}")
            return 1
        if args.format == "json":
            print(json.dumps(result, indent=2))
        else:
            print(f"PATH   {result['long_term_path']}")
            print(f"LINES  {result['source_line_count']}")
        return 0

    if args.command == "memory-status":
        try:
            result = build_memory_status(args.path)
        except ValueError as exc:
            print(f"ERROR {exc}")
            return 1
        if args.format == "json":
            print(json.dumps(result, indent=2))
        else:
            print(f"HOME   {result['home_root']}")
            print(f"LONG   chars={result['long_term_chars']}/{result['long_term_char_limit']} ratio={result['long_term_char_ratio']}")
            print(f"DAILY  files={result['daily_file_count']} lines={result['daily_line_count']} since_dream={result['daily_since_dream']}")
            if result.get("recommended_action"):
                print(f"NEXT   {result['recommended_action']}")
            for item in result.get("stale_signals", []):
                print(f"  - {item}")
        return 0

    if args.command == "show-steering":
        try:
            result = build_steering_summary(args.path, target_path=args.file)
        except ValueError as exc:
            print(f"ERROR {exc}")
            return 1
        if args.format == "json":
            print(json.dumps(result, indent=2))
        else:
            print_steering_summary(result)
        return 0

    if args.command == "repo-map":
        try:
            result = build_repo_map(args.path, repo_path=args.repo or args.path, max_entries=args.max_entries)
        except ValueError as exc:
            print(f"ERROR {exc}")
            return 1
        if args.format == "json":
            print(json.dumps(result, indent=2))
        else:
            print_repo_map(result)
        append_learning_event(
            "repo-map",
            {
                "repo_root": result["repo_root"],
                "stack_count": len(result.get("detected_stack", [])),
                "steering_count": len(result.get("steering", {}).get("documents", [])),
                "verification_target_count": len(result.get("verification_targets", [])),
            },
        )
        return 0

    if args.command == "list-tools":
        try:
            result = list_personal_tools(args.path)
        except ValueError as exc:
            print(f"ERROR {exc}")
            return 1
        if args.format == "json":
            print(json.dumps(result, indent=2))
        else:
            print(f"HOME   {result['home_root']}")
            for tool in result["tools"]:
                print(f"- {tool['id']} | {tool['kind']} | {tool['scope']} | {tool['location']}")
        return 0

    if args.command == "list-routines":
        try:
            result = list_personal_routines(args.path, mode=args.mode)
        except ValueError as exc:
            print(f"ERROR {exc}")
            return 1
        if args.format == "json":
            print(json.dumps(result, indent=2))
        else:
            print(f"HOME   {result['home_root']}")
            for routine in result["routines"]:
                print(f"- {routine['routine_id']} | {routine['mode']} | {routine['runner']} | {routine['summary']}")
        return 0

    if args.command == "run-routine":
        try:
            result = run_personal_routine(args.path, args.routine_id, mode=args.mode)
        except ValueError as exc:
            print(f"ERROR {exc}")
            return 1
        if args.format == "json":
            print(json.dumps(result, indent=2))
        else:
            print(f"ROUTINE {result['routine_id']}")
            print(f"STATUS  {result['status']}")
            for output_path in result.get("output_paths", []):
                print(f"OUTPUT  {output_path}")
            if result.get("receipt_path"):
                print(f"RECEIPT {result['receipt_path']}")
        return 0

    if args.command == "bind-home":
        try:
            binding_path = bind_personal_home(args.path, args.home)
            binding = load_home_binding(args.path) or {}
        except ValueError as exc:
            print(f"ERROR {exc}")
            return 1
        payload = {
            "binding_path": str(binding_path),
            "home_root": binding.get("home_root", ""),
            "updated_at": binding.get("updated_at", ""),
        }
        if args.format == "json":
            print(json.dumps(payload, indent=2))
        else:
            print(f"BIND   {display_path(binding_path)}")
            print(f"HOME   {payload['home_root']}")
        return 0

    if args.command == "init-pack":
        stakeholders = merged_stakeholders(args.owner, args.stakeholder)
        try:
            target_dir = write_initial_pack(
                pack_id=args.pack_id,
                output_dir=args.output,
                work_unit_id=args.work_unit_id,
                title=args.title,
                purpose=args.purpose,
                owner_actor_id=args.owner,
                approver_actor_ids=args.approver,
                stakeholder_actor_ids=stakeholders,
                branch_id=args.branch,
            )
        except (FileExistsError, FileNotFoundError, KeyError, ValueError) as exc:
            print(f"ERROR {exc}")
            return 1
        errors, warnings = validate_pack(target_dir, registry)
        for warning in warnings:
            print(f"WARN  {warning}")
        if errors:
            for error in errors:
                print(f"ERROR {error}")
            return 1
        print(f"OK    initialized {args.pack_id} at {target_dir}")
        return 0

    if args.command == "compile-pack":
        stakeholders = merged_stakeholders(args.owner, args.stakeholder)
        operator = args.operator or args.owner
        rollback_authority = args.rollback_authority or (args.approver[0] if args.approver else args.owner)
        service_owner = args.service_owner or args.owner
        try:
            target_dir = write_compiled_pack(
                pack_id=args.pack_id,
                output_dir=args.output,
                work_unit_id=args.work_unit_id,
                title=args.title,
                purpose=args.purpose,
                owner_actor_id=args.owner,
                approver_actor_ids=args.approver,
                stakeholder_actor_ids=stakeholders,
                branch_id=args.branch,
                operator_actor_id=operator,
                rollback_authority_actor_id=rollback_authority,
                service_owner_actor_id=service_owner,
                context_path=args.context,
            )
        except (FileExistsError, FileNotFoundError, KeyError, ValueError) as exc:
            print(f"ERROR {exc}")
            return 1
        errors, warnings = validate_pack(target_dir, registry)
        for warning in warnings:
            print(f"WARN  {warning}")
        if errors:
            for error in errors:
                print(f"ERROR {error}")
            return 1
        for warning in materialize_compile_outputs(target_dir, registry):
            print(f"WARN  {warning}")
        print(f"OK    compiled {args.pack_id} at {target_dir}")
        return 0

    if args.command == "bootstrap-pack":
        pack_map_index = pack_map()
        if args.pack_id not in pack_map_index:
            print(f"ERROR Unknown pack_id {args.pack_id!r}")
            return 1

        _, manifest = pack_map_index[args.pack_id]
        selected_mode, rationale = recommended_bootstrap_mode(args.pack_id, manifest)
        if args.mode != "auto":
            selected_mode = args.mode
            rationale = [f"Explicit override selected {selected_mode}"]

        stakeholders = merged_stakeholders(args.owner, args.stakeholder)
        operator = args.operator or args.owner
        rollback_authority = args.rollback_authority or (args.approver[0] if args.approver else args.owner)
        service_owner = args.service_owner or args.owner

        try:
            if selected_mode == "compile-pack":
                target_dir = write_compiled_pack(
                    pack_id=args.pack_id,
                    output_dir=args.output,
                    work_unit_id=args.work_unit_id,
                    title=args.title,
                    purpose=args.purpose,
                    owner_actor_id=args.owner,
                    approver_actor_ids=args.approver,
                    stakeholder_actor_ids=stakeholders,
                    branch_id=args.branch,
                    operator_actor_id=operator,
                    rollback_authority_actor_id=rollback_authority,
                    service_owner_actor_id=service_owner,
                    context_path=args.context,
                )
            else:
                target_dir = write_initial_pack(
                    pack_id=args.pack_id,
                    output_dir=args.output,
                    work_unit_id=args.work_unit_id,
                    title=args.title,
                    purpose=args.purpose,
                    owner_actor_id=args.owner,
                    approver_actor_ids=args.approver,
                    stakeholder_actor_ids=stakeholders,
                    branch_id=args.branch,
                )
        except (FileExistsError, FileNotFoundError, KeyError, ValueError) as exc:
            print(f"ERROR {exc}")
            return 1

        errors, warnings = validate_pack(target_dir, registry)
        for warning in warnings:
            print(f"WARN  {warning}")
        if errors:
            for error in errors:
                print(f"ERROR {error}")
            return 1
        if selected_mode == "compile-pack":
            for warning in materialize_compile_outputs(target_dir, registry):
                print(f"WARN  {warning}")
        print(f"MODE  {selected_mode}")
        for reason in rationale:
            print(f"WHY   {reason}")
        print(f"OK    bootstrapped {args.pack_id} at {target_dir}")
        return 0

    if args.command == "status-pack":
        try:
            summary = summarise_pack(args.path, registry)
        except (FileNotFoundError, TypeError, ValueError) as exc:
            print(f"ERROR {format_pack_surface_error(args.path, exc)}")
            return 1
        print_pack_status(summary)
        return 1 if summary["validation_errors"] else 0

    if args.command == "outcome":
        try:
            report = build_outcome_view(args.path, registry, repo_path=args.repo)
        except (FileNotFoundError, TypeError, ValueError) as exc:
            print(f"ERROR {format_pack_surface_error(args.path, exc)}")
            return 1
        if args.format == "json":
            print(json.dumps(report, indent=2))
        else:
            print_outcome_view(report)
        return 1 if report["validation_errors"] else 0

    if args.command == "recommend-execution":
        recommendation = recommend_execution(
            args.path,
            registry,
            intent=args.intent,
            repo_path=args.repo,
            home_path=args.home,
            runtime_target=args.runtime_target,
        )
        if args.format == "json":
            print(json.dumps(recommendation, indent=2))
        else:
            print(f"PACK   {recommendation['pack_id']}")
            print(f"WORK   {recommendation['work_unit_id']}")
            print(f"STATE  {recommendation['state']}")
            print(f"PROF   {recommendation['profile_id']}")
            print(f"INTENT {recommendation['intent']}")
            print(f"CLASS  {recommendation['execution_class']}")
            print(f"CTX    {recommendation['context_policy']}")
            print(f"DELEG  {recommendation['delegation_policy']}")
            print("TOOLS")
            for item in recommendation["tool_order"]:
                print(f"  - {item}")
            print("LIMITS")
            for item in recommendation["rate_limit_strategy"]:
                print(f"  - {item}")
            runtime_guidance = recommendation.get("runtime_guidance", {})
            if runtime_guidance:
                print("RUNTIME")
                selected = runtime_guidance.get("selected", {})
                if selected:
                    print(
                        f"  SELECTED {selected.get('id', '')} | {selected.get('maturity', '')} | "
                        f"priority={selected.get('priority', 0)}"
                    )
                for item in runtime_guidance.get("fallbacks", []):
                    print(f"  FALLBACK {item}")
            active_policy = recommendation.get("active_policy", {})
            if active_policy:
                print("POLICY")
                print(
                    f"  ACTIVE {active_policy.get('policy_id', '')} "
                    f"candidate={active_policy.get('candidate_id', '')}"
                )
                if active_policy.get("intent_overrides"):
                    print(f"  OVERRIDES {json.dumps(active_policy.get('intent_overrides', {}), sort_keys=True)}")
            print("WHY")
            for item in recommendation["rationale"]:
                print(f"  - {item}")
            repo_context = recommendation.get("repo_context", {})
            print("REPO")
            if repo_context.get("discovered"):
                print(f"  ROOT {repo_context.get('repo_root', '')}")
                git_info = repo_context.get("git", {})
                if git_info.get("tracked"):
                    print(
                        f"  GIT branch={git_info.get('branch') or 'unknown'} "
                        f"dirty={git_info.get('dirty_files', 0)}"
                    )
                for item in repo_context.get("next_actions", []):
                    print(f"  - {item}")
            else:
                for item in repo_context.get("notes", []):
                    print(f"  - {item}")
            memory_context = recommendation.get("memory_context", {})
            print("MEM")
            for item in memory_context.get("resume_items", []):
                print(f"  - {item}")
            home_context = memory_context.get("home", {})
            if home_context.get("bound"):
                print(f"  - home: {home_context.get('home_root', '')}")
            for item in memory_context.get("stale_signals", []):
                print(f"  - stale: {item}")
        return 0

    if args.command == "show-kpis":
        try:
            scorecard = load_competitive_kpis()
            summary = build_competitive_kpi_summary(
                scorecard,
                dimension=args.dimension,
                limit=args.limit,
            )
        except (FileNotFoundError, KeyError, TypeError, ValueError) as exc:
            print(f"ERROR {exc}")
            return 1
        if args.format == "json":
            print(json.dumps(summary, indent=2))
        else:
            print_competitive_kpi_summary(summary)
        return 0

    if args.command == "publish-readiness":
        try:
            report = build_publish_readiness()
        except (FileNotFoundError, KeyError, TypeError, ValueError) as exc:
            print(f"ERROR {exc}")
            return 1
        if args.format == "json":
            print(json.dumps(report, indent=2))
        else:
            print_publish_readiness(report)
        return 0

    if args.command == "validate-golden-benchmark":
        try:
            report, _report_path = build_golden_benchmark_report()
        except (FileNotFoundError, KeyError, TypeError, ValueError, subprocess.TimeoutExpired) as exc:
            print(f"ERROR {exc}")
            return 1
        if args.format == "json":
            print(json.dumps(report, indent=2))
        else:
            print_golden_benchmark_report(report)
        return 0

    if args.command == "get-started":
        try:
            guide = build_get_started_guide(target=args.target, audience=args.audience)
        except (FileNotFoundError, KeyError, TypeError, ValueError) as exc:
            print(f"ERROR {exc}")
            return 1
        if args.format == "json":
            print(json.dumps(guide, indent=2))
        else:
            print_get_started_guide(guide)
        return 0

    if args.command == "try-example":
        try:
            report = build_public_example_proof(
                args.example_id,
                output_path=args.output,
                registry=registry,
            )
        except (FileExistsError, FileNotFoundError, KeyError, TypeError, ValueError) as exc:
            print(f"ERROR {exc}")
            return 1
        if args.format == "json":
            print(json.dumps(report, indent=2))
        else:
            print_public_example_proof(report)
        return 0

    if args.command == "review-framework":
        try:
            review, _review_path = build_framework_review(
                dimension=args.dimension,
                limit=args.limit,
            )
        except (FileNotFoundError, KeyError, TypeError, ValueError) as exc:
            print(f"ERROR {exc}")
            return 1
        if args.format == "json":
            print(json.dumps(review, indent=2))
        else:
            print_framework_review(review)
        return 0

    if args.command == "stage-framework-experiment":
        try:
            experiment, _experiment_path = stage_framework_experiment(
                review_path=args.review,
                dimension=args.dimension,
                index=args.index,
            )
        except (FileNotFoundError, KeyError, TypeError, ValueError) as exc:
            print(f"ERROR {exc}")
            return 1
        if args.format == "json":
            print(json.dumps(experiment, indent=2))
        else:
            print_framework_experiment(experiment)
        return 0

    if args.command == "record-framework-outcome":
        try:
            outcome, _outcome_path = record_framework_experiment_outcome(
                args.path,
                actor=args.actor,
                result=args.result,
                score_delta=args.score_delta,
                adoption_signals=args.signal,
                notes=args.note,
            )
        except (FileNotFoundError, KeyError, TypeError, ValueError) as exc:
            print(f"ERROR {exc}")
            return 1
        if args.format == "json":
            print(json.dumps(outcome, indent=2))
        else:
            print_framework_outcome(outcome)
        return 0

    if args.command == "backtest-framework-evolution":
        try:
            backtest = build_framework_evolution_backtest(limit=args.limit)
        except (FileNotFoundError, KeyError, TypeError, ValueError) as exc:
            print(f"ERROR {exc}")
            return 1
        if args.format == "json":
            print(json.dumps(backtest, indent=2))
        else:
            print_framework_evolution_backtest(backtest)
        return 0

    if args.command == "execution-checklist":
        checklist = build_execution_checklist(
            args.path,
            registry,
            intent=args.intent,
            repo_path=args.repo,
            home_path=args.home,
            runtime_target=args.runtime_target,
        )
        if args.format == "json":
            print(json.dumps(checklist, indent=2))
        else:
            print_execution_checklist(checklist)
        append_learning_event(
            "execution-checklist",
            {
                "pack_id": checklist["pack_id"],
                "work_unit_id": checklist["work_unit_id"],
                "intent": checklist["intent"],
                "state": checklist["state"],
                "execution_class": checklist["execution_class"],
                "item_count": len(checklist["items"]),
                "runtime_target": checklist.get("runtime_target", {}).get("selected", ""),
            },
            pack_dir=args.path,
        )
        return 0

    if args.command == "compact-context":
        compact = build_compact_context(
            args.path,
            registry,
            intent=args.intent,
            repo_path=args.repo,
            home_path=args.home,
            runtime_target=args.runtime_target,
            max_items=args.max_items,
            max_chars=args.max_chars,
        )
        if args.format == "json":
            print(json.dumps(compact, indent=2))
        else:
            print_compact_context(compact)
        append_learning_event(
            "compact-context",
            {
                "pack_id": compact["pack_id"],
                "work_unit_id": compact["work_unit_id"],
                "intent": compact["intent"],
                "state": compact["state"],
                "execution_class": compact["execution_class"],
                "resume_item_count": len(compact["resume_items"]),
                "stale_signal_count": len(compact["stale_signals"]),
                "estimated_tokens": compact.get("token_budget", {}).get("estimated_tokens", 0),
                "compression_ratio": compact.get("token_budget", {}).get("compression_ratio", 0.0),
                "home_bound": bool(compact.get("home_memory")),
                "runtime_target": compact.get("runtime_target", {}).get("selected", ""),
            },
            pack_dir=args.path,
        )
        return 0

    if args.command == "show-adapters":
        try:
            summary = build_adapter_summary(capability=args.capability)
        except (FileNotFoundError, TypeError, ValueError) as exc:
            print(f"ERROR {exc}")
            return 1
        if args.format == "json":
            print(json.dumps(summary, indent=2))
        else:
            print_adapter_summary(summary)
        return 0

    if args.command == "adapter-conformance":
        try:
            summary = build_adapter_conformance_summary()
        except (FileNotFoundError, TypeError, ValueError) as exc:
            print(f"ERROR {exc}")
            return 1
        if args.format == "json":
            print(json.dumps(summary, indent=2))
        else:
            print_adapter_conformance(summary)
        return 0

    if args.command == "resolve-adapter":
        try:
            resolution = build_adapter_resolution(
                capability=args.capability,
                layer=args.layer,
                preferred=args.preferred,
            )
        except (FileNotFoundError, TypeError, ValueError) as exc:
            print(f"ERROR {exc}")
            return 1
        if args.format == "json":
            print(json.dumps(resolution, indent=2))
        else:
            print_adapter_resolution(resolution)
        return 0

    if args.command == "adapter-matrix":
        try:
            matrix = build_adapter_matrix()
        except (FileNotFoundError, TypeError, ValueError) as exc:
            print(f"ERROR {exc}")
            return 1
        if args.format == "json":
            print(json.dumps(matrix, indent=2))
        else:
            print_adapter_matrix(matrix)
        return 0

    if args.command == "harnesses":
        report = build_harness_catalog()
        if args.format == "json":
            print(json.dumps(report, indent=2))
        else:
            print_harness_catalog(report)
        return 0

    if args.command == "show-learning-events":
        source_path = runtime_events_path(args.path) if args.path is not None else LEARNING_EVENTS_PATH
        events = read_learning_events(
            path=source_path,
            limit=args.limit,
            event_type=args.event_type,
            work_unit_id=args.work_unit_id,
        )
        if args.format == "json":
            print(json.dumps({"path": display_path(source_path), "events": events}, indent=2))
        else:
            print(f"PATH   {display_path(source_path)}")
            print_learning_events(events)
        return 0

    if args.command == "learning-snapshot":
        source_path = runtime_events_path(args.path) if args.path is not None else LEARNING_EVENTS_PATH
        snapshot = build_learning_snapshot(path=source_path, limit=args.limit)
        if args.format == "json":
            print(json.dumps(snapshot, indent=2))
        else:
            print_learning_snapshot(snapshot)
        return 0

    if args.command == "routing-backtest":
        source_path = runtime_events_path(args.path) if args.path is not None else LEARNING_EVENTS_PATH
        backtest = build_routing_backtest(
            path=source_path,
            limit=args.limit,
            min_samples=args.min_samples,
        )
        if args.format == "json":
            print(json.dumps(backtest, indent=2))
        else:
            print_routing_backtest(backtest)
        return 0

    if args.command == "stage-runtime-handoff":
        try:
            handoff, handoff_path = build_runtime_handoff(
                args.path,
                registry,
                intent=args.intent,
                repo_path=args.repo,
                home_path=args.home,
                runtime_target=args.runtime_target,
                max_items=args.max_items,
                max_chars=args.max_chars,
            )
        except (FileNotFoundError, TypeError, ValueError, KeyError) as exc:
            print(f"ERROR {exc}")
            return 1
        if args.format == "json":
            print(json.dumps(handoff, indent=2))
        else:
            print_runtime_handoff(handoff)
        return 0

    if args.command == "activate-runtime-target":
        try:
            activation, _receipt_path = activate_runtime_target(
                args.path,
                registry,
                handoff_path=args.handoff,
                intent=args.intent,
                repo_path=args.repo,
                home_path=args.home,
                runtime_target=args.runtime_target,
                prefix=args.prefix,
                max_items=args.max_items,
                max_chars=args.max_chars,
            )
        except (FileNotFoundError, TypeError, ValueError, KeyError) as exc:
            print(f"ERROR {exc}")
            return 1
        if args.format == "json":
            print(json.dumps(activation, indent=2))
        else:
            print_runtime_activation(activation)
        return 0

    if args.command == "execute-flow":
        try:
            report, _flow_path = execute_flow(
                args.path,
                registry,
                mode=args.mode,
                intent=args.intent,
                repo_path=args.repo,
                home_path=args.home,
                runtime_target=args.runtime_target,
                consent_grants=args.consent,
                issue_adapter=args.issue_adapter,
                wiki_adapter=args.wiki_adapter,
                project_key=args.project_key,
                space_key=args.space_key,
                activate_runtime=args.activate_runtime,
                activation_prefix=args.prefix,
                author_actor_id=args.author,
                max_items=args.max_items,
                max_chars=args.max_chars,
            )
        except (FileNotFoundError, TypeError, ValueError, KeyError) as exc:
            print(f"ERROR {exc}")
            return 1
        if args.format == "json":
            print(json.dumps(report, indent=2))
        else:
            print_execute_flow(report)
        return 0

    if args.command == "review-policy":
        try:
            review, _report_path = build_policy_review(
                pack_dir=args.path,
                limit=args.limit,
                min_samples=args.min_samples,
            )
        except (FileNotFoundError, TypeError, ValueError, KeyError) as exc:
            print(f"ERROR {exc}")
            return 1
        if args.format == "json":
            print(json.dumps(review, indent=2))
        else:
            print_policy_review(review)
        return 0

    if args.command == "stage-policy-candidate":
        try:
            candidate, _candidate_path = stage_policy_candidate(
                args.path,
                review_path=args.review,
            )
        except (FileNotFoundError, TypeError, ValueError, KeyError, json.JSONDecodeError) as exc:
            print(f"ERROR {exc}")
            return 1
        if args.format == "json":
            print(json.dumps(candidate, indent=2))
        else:
            print_policy_candidate(candidate)
        return 0

    if args.command == "approve-policy-candidate":
        try:
            rollout, _rollout_path = approve_policy_candidate(
                args.path,
                args.candidate,
                approver=args.approver,
            )
        except (FileNotFoundError, TypeError, ValueError, KeyError, json.JSONDecodeError) as exc:
            print(f"ERROR {exc}")
            return 1
        if args.format == "json":
            print(json.dumps(rollout, indent=2))
        else:
            print_policy_rollout(rollout)
        return 0

    if args.command == "rollback-policy-candidate":
        try:
            rollback, _rollback_path = rollback_policy_candidate(
                args.path,
                args.candidate,
                actor=args.actor,
                reason=args.reason,
            )
        except (FileNotFoundError, TypeError, ValueError, KeyError, json.JSONDecodeError) as exc:
            print(f"ERROR {exc}")
            return 1
        if args.format == "json":
            print(json.dumps(rollback, indent=2))
        else:
            print_policy_rollout(rollback)
        return 0

    if args.command == "plan-install":
        try:
            plan = plan_install(
                bundle_ids=args.bundle or None,
                kit_ids=args.kit or None,
                target_ids=args.target or None,
                link_mode=args.link_mode,
                prefix=args.prefix,
            )
        except (FileNotFoundError, KeyError, TypeError, ValueError) as exc:
            print(f"ERROR {exc}")
            return 1
        if args.format == "json":
            print(json.dumps(plan, indent=2))
        else:
            print_install_plan(plan)
        return 0

    if args.command == "install-bundles":
        try:
            result = install_bundles(
                bundle_ids=args.bundle or None,
                kit_ids=args.kit or None,
                target_ids=args.target or None,
                link_mode=args.link_mode,
                prefix=args.prefix,
            )
        except (FileNotFoundError, KeyError, TypeError, ValueError, OSError) as exc:
            print(f"ERROR {exc}")
            return 1
        if args.format == "json":
            print(json.dumps(result, indent=2))
        else:
            print(f"STATUS  {result['status']}")
            print(f"RECEIPT {result['receipt_path']}")
            for install in result["installs"]:
                print(
                    f"  - {install['bundle_label']} -> {install['target_label']} "
                    f"({install['link_mode']}) {install['universal_destination']}"
                )
        return 0

    if args.command == "update-bundles":
        try:
            result = update_bundles(
                bundle_ids=args.bundle or None,
                kit_ids=args.kit or None,
                target_ids=args.target or None,
                link_mode=args.link_mode,
                prefix=args.prefix,
            )
        except (FileNotFoundError, KeyError, TypeError, ValueError, OSError) as exc:
            print(f"ERROR {exc}")
            return 1
        if args.format == "json":
            print(json.dumps(result, indent=2))
        else:
            print(f"STATUS  {result['status']}")
            print(f"RECEIPT {result['receipt_path']}")
            for install in result["installs"]:
                print(
                    f"  - {install['bundle_label']} -> {install['target_label']} "
                    f"({install['link_mode']}) {install['universal_destination']}"
                )
        return 0

    if args.command == "uninstall-bundles":
        try:
            result = uninstall_bundles(
                bundle_ids=args.bundle or None,
                kit_ids=args.kit or None,
                target_ids=args.target or None,
                prefix=args.prefix,
            )
        except (FileNotFoundError, KeyError, TypeError, ValueError, OSError) as exc:
            print(f"ERROR {exc}")
            return 1
        if args.format == "json":
            print(json.dumps(result, indent=2))
        else:
            print(f"STATUS  {result['status']}")
            print(f"RECEIPT {result['receipt_path']}")
            for removal in result["removals"]:
                removed = ", ".join(removal["removed_paths"]) or "nothing to remove"
                print(f"  - {removal['bundle_id']} -> {removal['target_id']}: {removed}")
        return 0

    if args.command == "doctor-install":
        try:
            result = doctor_install(
                bundle_ids=args.bundle or None,
                kit_ids=args.kit or None,
                target_ids=args.target or None,
                prefix=args.prefix,
            )
        except (FileNotFoundError, KeyError, TypeError, ValueError, OSError) as exc:
            print(f"ERROR {exc}")
            return 1
        if args.format == "json":
            print(json.dumps(result, indent=2))
        else:
            print(f"STATUS  {result['status']}")
            for check in result["checks"]:
                missing = len(check["missing_paths"])
                print(
                    f"  - {check['bundle_id']} -> {check['target_id']}: "
                    f"status={check['status']} missing={missing} receipt={check['latest_receipt_status']} "
                    f"activation={check.get('activation_status', '')}"
                )
                for health in check.get("health_checks", []):
                    print(f"    health {health['id']}={health['status']} {health['detail']}")
                for step in check.get("activation_steps", []):
                    print(f"    activate {step}")
                for path in check["missing_paths"]:
                    print(f"    missing {path}")
        return 0

    if args.command == "catalog-bundles":
        try:
            catalog = build_install_catalog(target=args.target, kind=args.kind)
        except (FileNotFoundError, KeyError, TypeError, ValueError) as exc:
            print(f"ERROR {exc}")
            return 1
        if args.format == "json":
            print(json.dumps(catalog, indent=2))
        else:
            print_install_catalog(catalog)
        return 0

    if args.command == "bind-atlassian":
        try:
            output_path = bind_atlassian_targets(
                args.path,
                cloud_id=args.cloud_id,
                site_url=args.site_url,
                jira_project_key=args.project_key,
                confluence_space_key=args.space_key,
                confluence_space_id=args.space_id,
            )
        except ValueError as exc:
            print(f"ERROR {exc}")
            return 1
        print(f"OK    bound Atlassian targets in {display_path(output_path)}")
        return 0

    if args.command == "show-atlassian":
        targets = load_atlassian_targets(args.path)
        if targets is None:
            print("No Atlassian target binding found")
            return 1
        if args.format == "json":
            print(json.dumps(targets, indent=2))
        else:
            print(f"CLOUD {targets.get('cloud_id', '')}")
            print(f"SITE  {targets.get('site_url', '')}")
            jira = targets.get("jira", {}) if isinstance(targets.get("jira", {}), dict) else {}
            confluence = (
                targets.get("confluence", {}) if isinstance(targets.get("confluence", {}), dict) else {}
            )
            print(f"JIRA  {jira.get('project_key', '') or 'unset'}")
            print(f"SPACE {confluence.get('space_key', '') or 'unset'}")
            print(f"SID   {confluence.get('space_id', '') or 'unset'}")
        return 0

    if args.command == "run-pack":
        try:
            report, report_path = run_pack(
                args.path,
                registry,
                mode=args.mode,
                intent=args.intent,
                repo_path=args.repo,
                home_path=args.home,
                runtime_target=args.runtime_target,
                consent_grants=args.consent,
                issue_adapter=args.issue_adapter,
                wiki_adapter=args.wiki_adapter,
                project_key=args.project_key,
                space_key=args.space_key,
            )
        except ValueError as exc:
            print(f"ERROR {exc}")
            return 1

        if args.format == "json":
            print(json.dumps(report, indent=2))
        else:
            print(f"PACK   {report['pack_id']}")
            print(f"WORK   {report['work_unit_id']}")
            print(f"MODE   {report['mode']}")
            print(f"INTENT {report['intent']}")
            print(f"STATE  {report['state_before']} -> {report['state_after']}")
            print(f"HEALTH {report['health_before']} -> {report['health_after']}")
            print(f"CLASS  {report['recommendation']['execution_class']}")
            print(f"CONSENT {', '.join([k for k, v in report['consent']['categories'].items() if v]) or 'none'}")
            if report.get("home_binding", {}).get("bound"):
                print(f"HOME   {report['home_binding']['home_root']}")
            if report["atlassian_targets"]["bound"]:
                print(
                    f"ATLASSIAN cloud={report['atlassian_targets']['cloud_id'] or 'unset'} "
                    f"jira={report['atlassian_targets']['project_key'] or 'unset'} "
                    f"space={report['atlassian_targets']['space_key'] or 'unset'}"
                )
            print(f"REPORT {display_path(report_path)}")
            repo_context = report["recommendation"].get("repo_context", {})
            print("REPO")
            if repo_context.get("discovered"):
                print(f"  ROOT {repo_context.get('repo_root', '')}")
                for item in repo_context.get("next_actions", []):
                    print(f"  - {item}")
            else:
                for item in repo_context.get("notes", []):
                    print(f"  - {item}")
            memory_context = report["recommendation"].get("memory_context", {})
            print("MEM")
            for item in memory_context.get("resume_items", []):
                print(f"  - {item}")
            for item in memory_context.get("stale_signals", []):
                print(f"  - stale: {item}")
            if report.get("memory_append", {}).get("appended"):
                print(f"MEMORY {report['memory_append']['daily_path']}")
            print("ACTIONS")
            for action in report["actions"]:
                output_suffix = f" -> {action['output_path']}" if action.get("output_path") else ""
                print(
                    f"  - [{action['status']}] {action['category']} {action['command']}: "
                    f"{action['message']}{output_suffix}"
                )
            if report["blockers"]:
                print("BLOCKERS")
                for blocker in report["blockers"]:
                    print(f"  - {blocker}")
        return 0

    if args.command == "export-tasks":
        try:
            output_path = export_tasks(args.path, registry, output_path=args.output)
        except ValueError as exc:
            print(f"ERROR {exc}")
            return 1
        print(f"OK    exported tasks to {display_path(output_path)}")
        return 0

    if args.command == "sync-tasks":
        try:
            output_path = sync_tasks(args.path, registry, output_path=args.output)
        except ValueError as exc:
            print(f"ERROR {exc}")
            return 1
        print(f"OK    synced tasks to {display_path(output_path)}")
        return 0

    if args.command == "export-issues":
        try:
            output_path = export_issues(
                args.path,
                registry,
                adapter=args.adapter,
                output_dir=args.output,
            )
        except ValueError as exc:
            print(f"ERROR {exc}")
            return 1
        print(f"OK    exported issues to {display_path(output_path)}")
        return 0

    if args.command == "export-wiki":
        try:
            output_path = export_wiki(
                args.path,
                registry,
                adapter=args.adapter,
                output_dir=args.output,
            )
        except ValueError as exc:
            print(f"ERROR {exc}")
            return 1
        print(f"OK    exported wiki to {display_path(output_path)}")
        return 0

    if args.command == "publish-issues":
        if args.apply_local and args.bridge_runner:
            print("ERROR publish-issues accepts only one of --apply-local or --bridge-runner")
            return 1
        try:
            output_path = publish_issues(
                args.path,
                registry,
                adapter=args.adapter,
                output_dir=args.output,
                project_key=args.project_key,
                cloud_id=args.cloud_id,
                site_url=args.site_url,
            )
        except ValueError as exc:
            print(f"ERROR {exc}")
            return 1
        if args.apply_local:
            try:
                receipt, _receipt_path = apply_publish_plan(output_path)
            except ValueError as exc:
                print(f"ERROR {exc}")
                return 1
            if args.format == "json":
                print(json.dumps(receipt, indent=2))
            else:
                print_apply_publish_plan(receipt)
            return 0
        if args.bridge_runner:
            try:
                receipt, _receipt_path, _result_path = execute_publish_plan(
                    output_path,
                    runner=args.bridge_runner,
                    timeout_seconds=args.bridge_timeout_seconds,
                )
            except ValueError as exc:
                print(f"ERROR {exc}")
                return 1
            if args.format == "json":
                print(json.dumps(receipt, indent=2))
            else:
                print_execute_publish_plan(receipt)
            return 0
        staged_receipt = build_staged_publish_receipt(output_path)
        if args.format == "json":
            print(json.dumps(staged_receipt, indent=2))
        else:
            print(f"OK    staged issue publish plan at {display_path(output_path)}")
        return 0

    if args.command == "apply-publish-plan":
        try:
            receipt, _receipt_path = apply_publish_plan(
                args.path,
                output_dir=args.output,
            )
        except ValueError as exc:
            print(f"ERROR {exc}")
            return 1
        if args.format == "json":
            print(json.dumps(receipt, indent=2))
        else:
            print_apply_publish_plan(receipt)
        return 0

    if args.command == "execute-publish-plan":
        try:
            receipt, _receipt_path, _result_path = execute_publish_plan(
                args.path,
                runner=args.runner,
                output_dir=args.output,
                timeout_seconds=args.timeout_seconds,
            )
        except ValueError as exc:
            print(f"ERROR {exc}")
            return 1
        if args.format == "json":
            print(json.dumps(receipt, indent=2))
        else:
            print_execute_publish_plan(receipt)
        return 0

    if args.command == "publish-wiki":
        if args.apply_local and args.bridge_runner:
            print("ERROR publish-wiki accepts only one of --apply-local or --bridge-runner")
            return 1
        try:
            output_path = publish_wiki(
                args.path,
                registry,
                adapter=args.adapter,
                output_dir=args.output,
                space_key=args.space_key,
                cloud_id=args.cloud_id,
                site_url=args.site_url,
                space_id=args.space_id,
            )
        except ValueError as exc:
            print(f"ERROR {exc}")
            return 1
        if args.apply_local:
            try:
                receipt, _receipt_path = apply_publish_plan(output_path)
            except ValueError as exc:
                print(f"ERROR {exc}")
                return 1
            if args.format == "json":
                print(json.dumps(receipt, indent=2))
            else:
                print_apply_publish_plan(receipt)
            return 0
        if args.bridge_runner:
            try:
                receipt, _receipt_path, _result_path = execute_publish_plan(
                    output_path,
                    runner=args.bridge_runner,
                    timeout_seconds=args.bridge_timeout_seconds,
                )
            except ValueError as exc:
                print(f"ERROR {exc}")
                return 1
            if args.format == "json":
                print(json.dumps(receipt, indent=2))
            else:
                print_execute_publish_plan(receipt)
            return 0
        staged_receipt = build_staged_publish_receipt(output_path)
        if args.format == "json":
            print(json.dumps(staged_receipt, indent=2))
        else:
            print(f"OK    staged wiki publish plan at {display_path(output_path)}")
        return 0

    if args.command == "capture-evidence":
        try:
            artifact_path = capture_evidence(
                args.path,
                registry,
                author_actor_id=args.author,
                claims=args.claim,
                test_results=args.test_result,
                review_results=args.review_result,
                operational_results=args.operational_result,
                residual_risks=args.risk,
                approver_actor_ids=args.approver,
                target_artifact_type=args.target_artifact,
                references=args.reference,
                status=args.status,
            )
        except ValueError as exc:
            print(f"ERROR {exc}")
            return 1
        errors, warnings = validate_pack(args.path, registry)
        for warning in warnings:
            print(f"WARN  {warning}")
        if errors:
            for error in errors:
                print(f"ERROR {error}")
            return 1
        print(f"OK    captured evidence at {display_path(artifact_path)}")
        return 0

    if args.command == "harvest-evidence":
        try:
            report, report_path, evidence_path = harvest_evidence(
                args.path,
                registry,
                author_actor_id=args.author,
                repo_path=args.repo,
                home_path=args.home,
                categories=args.category or None,
                claims=args.claim,
                residual_risks=args.risk,
                approver_actor_ids=args.approver,
                target_artifact_type=args.target_artifact,
                references=args.reference,
                status=args.status,
                timeout_seconds=args.timeout_seconds,
                max_targets=args.max_targets,
            )
        except ValueError as exc:
            print(f"ERROR {exc}")
            return 1
        errors, warnings = validate_pack(args.path, registry)
        for warning in warnings:
            print(f"WARN  {warning}")
        if errors:
            for error in errors:
                print(f"ERROR {error}")
            return 1
        if args.format == "json":
            print(json.dumps(report, indent=2))
        else:
            print_harvest_report(report)
            print(f"OK    harvested evidence at {display_path(evidence_path)} from {display_path(report_path)}")
        return 0

    if args.command == "capture-output":
        try:
            artifact_path = capture_output(
                args.path,
                registry,
                author_actor_id=args.author,
                task_index=args.task_index,
                task_status=args.status,
                note=args.note,
                references=args.reference,
                deliverable=args.deliverable,
            )
        except ValueError as exc:
            print(f"ERROR {exc}")
            return 1
        errors, warnings = validate_pack(args.path, registry)
        for warning in warnings:
            print(f"WARN  {warning}")
        if errors:
            for error in errors:
                print(f"ERROR {error}")
            return 1
        print(f"OK    captured output in {display_path(artifact_path)}")
        return 0

    if args.command == "capture-approval":
        try:
            artifact_path = capture_approval(
                args.path,
                registry,
                author_actor_id=args.author,
                approver_actor_id=args.approver_actor,
                approval_scope=args.scope,
                waivers=args.waiver,
                conditions=args.condition,
                target_artifact_type=args.target_artifact,
                status=args.status,
            )
        except ValueError as exc:
            print(f"ERROR {exc}")
            return 1
        errors, warnings = validate_pack(args.path, registry)
        for warning in warnings:
            print(f"WARN  {warning}")
        if errors:
            for error in errors:
                print(f"ERROR {error}")
            return 1
        print(f"OK    captured approval at {display_path(artifact_path)}")
        return 0

    if args.command == "capture-publication":
        try:
            artifact_path = capture_publication(
                args.path,
                registry,
                author_actor_id=args.author,
                input_path=args.input,
                publication_scope=args.scope,
                approver_actor_ids=args.approver,
                status=args.status,
            )
        except ValueError as exc:
            print(f"ERROR {exc}")
            return 1
        errors, warnings = validate_pack(args.path, registry)
        for warning in warnings:
            print(f"WARN  {warning}")
        if errors:
            for error in errors:
                print(f"ERROR {error}")
            return 1
        print(f"OK    captured publication at {display_path(artifact_path)}")
        return 0

    if args.command == "advance-pack":
        try:
            current_state, target_state = advance_pack_state(args.path, registry, target_state=args.to)
        except ValueError as exc:
            print(f"ERROR {exc}")
            return 1
        summary = summarise_pack(args.path, registry)
        print(f"OK    advanced {display_path(args.path)} from {current_state} to {target_state}")
        print_pack_status(summary)
        return 0

    if args.command == "validate":
        errors, warnings = validate_file(args.path, registry, explicit_schema=args.schema)
    else:
        errors, warnings = validate_pack(args.path, registry)

    for warning in warnings:
        print(f"WARN  {warning}")

    if errors:
        for error in errors:
            print(f"ERROR {error}")
        return 1

    if args.command == "validate":
        print(f"OK    {args.path}")
    else:
        print(f"OK    {args.path} ({len(warnings)} warnings)")
    return 0


if __name__ == "__main__":
    sys.exit(main())
