from __future__ import annotations

import json
import unittest
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
CAROUSEL_DATA_PATH = ROOT / "docs" / "_data" / "proof_carousel.json"


def load_carousel_data() -> dict:
    return json.loads(CAROUSEL_DATA_PATH.read_text(encoding="utf-8"))


class ProofCarouselContractTests(unittest.TestCase):
    def test_carousel_exposes_checked_in_truth_safe_slides(self) -> None:
        data = load_carousel_data()
        self.assertEqual(len(data["slides"]), 5)
        for slide in data["slides"]:
            with self.subTest(slide=slide["id"]):
                self.assertIn("image", slide)
                self.assertIn("alt", slide)
                self.assertIn("title", slide)
                self.assertIn("truth_note", slide)
                self.assertIn("Checked-in proof carousel slide.", slide["truth_note"])

    def test_carousel_assets_exist_under_docs_assets(self) -> None:
        data = load_carousel_data()
        for slide in data["slides"]:
            asset_path = ROOT / "docs" / slide["image"].lstrip("/")
            with self.subTest(slide=slide["id"], asset=str(asset_path)):
                self.assertTrue(asset_path.exists(), asset_path)


if __name__ == "__main__":
    unittest.main()
