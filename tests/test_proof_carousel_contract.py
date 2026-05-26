from __future__ import annotations

import json
import unittest
from collections import Counter
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

    def test_carousel_slide_ids_and_image_routes_are_unique_and_scoped(self) -> None:
        data = load_carousel_data()
        ids = [slide["id"] for slide in data["slides"]]
        images = [slide["image"] for slide in data["slides"]]
        self.assertFalse(
            [slide_id for slide_id, count in Counter(ids).items() if count > 1],
            "Proof carousel slide ids should be unique",
        )
        self.assertFalse(
            [image for image, count in Counter(images).items() if count > 1],
            "Proof carousel image routes should be unique",
        )
        for slide in data["slides"]:
            with self.subTest(slide=slide["id"]):
                self.assertTrue(slide["image"].startswith("/assets/proof/"))

    def test_carousel_eyebrow_sequence_matches_slide_order(self) -> None:
        data = load_carousel_data()
        for index, slide in enumerate(data["slides"], start=1):
            with self.subTest(slide=slide["id"]):
                self.assertEqual(slide["eyebrow"], f"Proof slide {index}")


if __name__ == "__main__":
    unittest.main()
