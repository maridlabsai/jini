from __future__ import annotations

import argparse
import json
from pathlib import Path
from typing import Any


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Sanitize a commercial proof bundle for the public website.")
    parser.add_argument("--input", required=True, help="Path to the source proof bundle JSON.")
    parser.add_argument("--output", required=True, help="Path to the public data JSON file to write.")
    return parser.parse_args()


def build_public_proof_snapshot(bundle: dict[str, Any]) -> dict[str, Any]:
    proof_cards = [
        {
            "label": str(card["label"]),
            "value": str(card["value"]),
        }
        for card in bundle["proof_cards"]
    ]
    sections = [
        {
            "id": str(section["id"]),
            "headline": str(section["headline"]),
            "bullets": [str(bullet) for bullet in section["bullets"]],
        }
        for section in bundle["sections"]
    ]
    return {
        "product": "Jini public proof snapshot",
        "proof_posture": str(bundle["proof_posture"]),
        "hero": {
            "eyebrow": str(bundle["hero"]["eyebrow"]),
            "headline": str(bundle["hero"]["headline"]),
            "body": str(bundle["hero"]["body"]),
        },
        "proof_cards": proof_cards,
        "sections": sections,
        "trust_rules": [str(rule) for rule in bundle["trust_rules"]],
        "source_contract": "sanitized-commercial-proof-bundle",
    }


def main() -> int:
    args = parse_args()
    input_path = Path(args.input)
    output_path = Path(args.output)
    bundle = json.loads(input_path.read_text(encoding="utf-8"))
    snapshot = build_public_proof_snapshot(bundle)
    output_path.parent.mkdir(parents=True, exist_ok=True)
    output_path.write_text(json.dumps(snapshot, indent=2) + "\n", encoding="utf-8")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
