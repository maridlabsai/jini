from __future__ import annotations

import argparse
import json
from pathlib import Path
from typing import Any


PACKET_IDS = ("mac", "windows", "ios", "android")


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="Sanitize commercial release packets for the public website."
    )
    parser.add_argument(
        "--packets-dir",
        required=True,
        help="Path to the commercial release packets directory.",
    )
    parser.add_argument(
        "--output",
        required=True,
        help="Path to the public data JSON file to write.",
    )
    return parser.parse_args()


def _surface_badge(packet: dict[str, Any]) -> str:
    if packet["availability"]["public_download_available_today"]:
        return "Available now"
    activation = str(packet["activation"]["subscription_activation"])
    if activation.startswith("planned-"):
        return "Preview only"
    return "Planned"


def _surface_name(packet_id: str) -> str:
    names = {
        "mac": "macOS app shell",
        "windows": "Windows app shell",
        "ios": "iOS companion app",
        "android": "Android companion app",
    }
    return names[packet_id]


def _surface_next_step(packet_id: str) -> str:
    next_steps = {
        "mac": "Free direct download once the signed wrapper and notarization lane is real",
        "windows": "Free direct download once the signed installer lane is real",
        "ios": "Free app once the real host container and submission lane exist",
        "android": "Free direct download first where policy allows, with Play Store secondary",
    }
    return next_steps[packet_id]


def _surface_activation(packet_id: str) -> str:
    activation = {
        "mac": "Start with a 30-day free trial, then buy on the website and sign in",
        "windows": "Start with a 30-day free trial, then buy on the website and sign in",
        "ios": "Sign in with an existing paid account",
        "android": "Sign in with an existing paid account",
    }
    return activation[packet_id]


def build_public_surfaces_snapshot(packet_map: dict[str, dict[str, Any]]) -> dict[str, Any]:
    surfaces = [
        {
            "id": "cli",
            "name": "CLI",
            "badge": "Available now",
            "current_state": "Installable now on macOS and Linux",
            "next_step": "This remains the primary shipped surface until desktop and mobile move out of preview",
            "activation": "None",
            "distribution_policy": "direct-shell-install",
            "release_readiness_status": "live",
            "artifact_name": "jini",
            "packet_reference": None,
        }
    ]

    for packet_id in PACKET_IDS:
        packet = packet_map[packet_id]
        surfaces.append(
            {
                "id": packet_id,
                "name": _surface_name(packet_id),
                "badge": _surface_badge(packet),
                "current_state": str(packet["availability"]["availability_summary"]),
                "next_step": _surface_next_step(packet_id),
                "activation": _surface_activation(packet_id),
                "distribution_policy": str(packet["availability"]["distribution_policy"]),
                "release_readiness_status": str(packet["release_readiness"]["status"]),
                "artifact_name": str(packet["artifacts"]["artifact_name"]),
                "packet_reference": f"{packet_id}.json",
            }
        )

    surfaces.append(
        {
            "id": "commercial-license",
            "name": "Commercial License",
            "badge": "Planned",
            "current_state": "Planned. Not live yet",
            "next_step": "30-day free trial first, then $1/month website checkout plus account entitlement once checkout is real",
            "activation": "Start with a 30-day free trial, then website checkout + account entitlement",
            "distribution_policy": "website-checkout",
            "release_readiness_status": "planned",
            "artifact_name": "web-checkout",
            "packet_reference": None,
        }
    )

    return {
        "product": "Jini public surface availability",
        "source_contract": "sanitized-commercial-release-packets",
        "surfaces": surfaces,
    }


def main() -> int:
    args = parse_args()
    packets_dir = Path(args.packets_dir)
    output_path = Path(args.output)
    packet_map = {
        packet_id: json.loads((packets_dir / f"{packet_id}.json").read_text(encoding="utf-8"))
        for packet_id in PACKET_IDS
    }
    snapshot = build_public_surfaces_snapshot(packet_map)
    output_path.parent.mkdir(parents=True, exist_ok=True)
    output_path.write_text(json.dumps(snapshot, indent=2) + "\n", encoding="utf-8")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
