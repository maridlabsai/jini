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

from tools.sync_public_proof import build_public_proof_snapshot


DATA_PATH = ROOT / "docs" / "_data" / "public_proof.json"
COMMERCIAL_BUNDLE_PATH = (
    ROOT.parent / "jini-commercial" / "apps" / "proof" / "public-proof-site-bundle.json"
)


class PublicProofIngestionTests(unittest.TestCase):
    def test_checked_in_public_proof_matches_sanitized_bundle(self) -> None:
        bundle = json.loads(COMMERCIAL_BUNDLE_PATH.read_text(encoding="utf-8"))
        expected = build_public_proof_snapshot(bundle)
        actual = json.loads(DATA_PATH.read_text(encoding="utf-8"))
        self.assertEqual(actual, expected)

    def test_sync_public_proof_script_writes_snapshot(self) -> None:
        with tempfile.TemporaryDirectory() as tempdir:
            output_path = Path(tempdir) / "public_proof.json"
            subprocess.run(
                [
                    sys.executable,
                    str(ROOT / "tools" / "sync_public_proof.py"),
                    "--input",
                    str(COMMERCIAL_BUNDLE_PATH),
                    "--output",
                    str(output_path),
                ],
                check=True,
                cwd="/tmp",
            )
            snapshot = json.loads(output_path.read_text(encoding="utf-8"))
            self.assertEqual(snapshot["source_contract"], "sanitized-commercial-proof-bundle")
            self.assertEqual(snapshot["proof_cards"][0]["value"], "5 surfaces")


if __name__ == "__main__":
    unittest.main()
