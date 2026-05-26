from __future__ import annotations

import json
import subprocess
import sys
import tempfile
import unittest
from collections import Counter
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
if str(ROOT) not in sys.path:
    sys.path.insert(0, str(ROOT))

from tools.sync_public_surfaces import build_public_surfaces_snapshot


DATA_PATH = ROOT / "docs" / "_data" / "public_surfaces.json"
COMMERCIAL_PACKETS_DIR = ROOT.parent / "jini-commercial" / "apps" / "releases" / "packets"


class PublicSurfaceIngestionTests(unittest.TestCase):
    def test_checked_in_public_surfaces_match_sanitized_packets(self) -> None:
        packet_map = {
            packet_id: json.loads(
                (COMMERCIAL_PACKETS_DIR / f"{packet_id}.json").read_text(encoding="utf-8")
            )
            for packet_id in ("mac", "windows", "ios", "android")
        }
        expected = build_public_surfaces_snapshot(packet_map)
        actual = json.loads(DATA_PATH.read_text(encoding="utf-8"))
        self.assertEqual(actual, expected)
        activation_by_id = {surface["id"]: surface["activation"] for surface in actual["surfaces"]}
        self.assertEqual(
            activation_by_id["mac"],
            "When checkout is live, start with the planned 30-day free trial, then buy on the website and sign in",
        )
        self.assertEqual(
            activation_by_id["windows"],
            "When checkout is live, start with the planned 30-day free trial, then buy on the website and sign in",
        )
        self.assertEqual(
            activation_by_id["commercial-license"],
            "When checkout is live, start with the planned 30-day free trial, then activate website entitlement",
        )

    def test_sync_public_surfaces_script_writes_snapshot(self) -> None:
        with tempfile.TemporaryDirectory() as tempdir:
            output_path = Path(tempdir) / "public_surfaces.json"
            subprocess.run(
                [
                    sys.executable,
                    str(ROOT / "tools" / "sync_public_surfaces.py"),
                    "--packets-dir",
                    str(COMMERCIAL_PACKETS_DIR),
                    "--output",
                    str(output_path),
                ],
                check=True,
                cwd="/tmp",
            )
            snapshot = json.loads(output_path.read_text(encoding="utf-8"))
            self.assertEqual(snapshot["source_contract"], "sanitized-commercial-release-packets")
            self.assertEqual(snapshot["surfaces"][1]["artifact_name"], "jini-mac.dmg")
            self.assertEqual(
                snapshot["surfaces"][3]["next_step"],
                "Free app once the signed submission and store-delivery lanes are real",
            )

    def test_checked_in_public_surfaces_keep_unique_ids_and_valid_packet_refs(self) -> None:
        snapshot = json.loads(DATA_PATH.read_text(encoding="utf-8"))
        ids = [surface["id"] for surface in snapshot["surfaces"]]
        self.assertFalse(
            [surface_id for surface_id, count in Counter(ids).items() if count > 1],
            "Public surface ids should be unique",
        )

        known_packets = {f"{packet_id}.json" for packet_id in ("mac", "windows", "ios", "android")}
        for surface in snapshot["surfaces"]:
            packet_reference = surface["packet_reference"]
            with self.subTest(surface=surface["id"]):
                if packet_reference is None:
                    self.assertIn(surface["id"], {"cli", "commercial-license"})
                else:
                    self.assertIn(packet_reference, known_packets)

    def test_checked_in_public_surfaces_keep_status_and_badge_pairs_consistent(self) -> None:
        snapshot = json.loads(DATA_PATH.read_text(encoding="utf-8"))
        expected_pairs = {
            "live": "Available now",
            "blocked": "Preview only",
            "planned": "Planned",
        }
        for surface in snapshot["surfaces"]:
            with self.subTest(surface=surface["id"]):
                self.assertEqual(
                    surface["badge"],
                    expected_pairs[surface["release_readiness_status"]],
                )
                self.assertIn(
                    surface["distribution_policy"],
                    {
                        "direct-shell-install",
                        "direct-first-when-allowed",
                        "store-required-by-platform",
                        "website-checkout",
                    },
                )


if __name__ == "__main__":
    unittest.main()
