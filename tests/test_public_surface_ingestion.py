from __future__ import annotations

import json
import subprocess
import sys
import tempfile
import unittest
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


if __name__ == "__main__":
    unittest.main()
