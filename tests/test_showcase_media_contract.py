from __future__ import annotations

import json
import re
import unittest
from collections import Counter, defaultdict
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
SHOWCASE_DATA_PATH = ROOT / "docs" / "_data" / "showcase_media.json"
DOCS_DIR = ROOT / "docs"
STYLE_PATH = DOCS_DIR / "assets" / "css" / "style.css"


def build_public_target_map() -> dict[str, Path]:
    mapping = {"/": DOCS_DIR / "index.md"}
    for path in DOCS_DIR.glob("*.md"):
        if path.stem == "index":
            continue
        mapping[f"/{path.stem}.html"] = path
    for path in DOCS_DIR.glob("*.html"):
        if path.name == "404.html":
            mapping[f"/{path.name}"] = path
    return mapping


PUBLIC_TARGETS = build_public_target_map()


def load_showcase_data() -> dict:
    return json.loads(SHOWCASE_DATA_PATH.read_text(encoding="utf-8"))


class ShowcaseMediaContractTests(unittest.TestCase):
    def test_every_card_declares_truth_safe_capture_metadata(self) -> None:
        data = load_showcase_data()
        for section_name in ("product_surface_cards", "output_cards"):
            for card in data[section_name]:
                with self.subTest(section=section_name, title=card["title"]):
                    self.assertIn("capture_id", card)
                    self.assertIn("capture_kind", card)
                    self.assertIn("capture_source", card)
                    self.assertIn("truth_note", card)
                    self.assertIn("reuse_note", card)
                    self.assertIn(
                        card["capture_kind"],
                        {"current-public-example", "checked-in-public-illustration"},
                    )

    def test_every_card_points_at_a_checked_in_example_asset(self) -> None:
        data = load_showcase_data()
        for section_name in ("product_surface_cards", "output_cards"):
            for card in data[section_name]:
                image = card["image"].lstrip("/")
                image_path = ROOT / "docs" / card["image"].lstrip("/")
                with self.subTest(section=section_name, image=image):
                    self.assertTrue(image_path.exists(), image_path)

    def test_reused_captures_keep_one_source_of_truth(self) -> None:
        data = load_showcase_data()
        grouped: dict[str, list[dict]] = defaultdict(list)
        for section_name in ("product_surface_cards", "output_cards"):
            for card in data[section_name]:
                grouped[card["capture_id"]].append(card)

        for capture_id, cards in grouped.items():
            with self.subTest(capture_id=capture_id):
                self.assertGreaterEqual(len(cards), 1)
                images = {card["image"] for card in cards}
                sources = {card["capture_source"] for card in cards}
                kinds = {card["capture_kind"] for card in cards}
                self.assertEqual(len(images), 1)
                self.assertEqual(len(sources), 1)
                self.assertEqual(len(kinds), 1)

    def test_reused_captures_are_explicitly_marked_as_reused(self) -> None:
        data = load_showcase_data()
        capture_counts = Counter(
            card["capture_id"]
            for section_name in ("product_surface_cards", "output_cards")
            for card in data[section_name]
        )

        for section_name in ("product_surface_cards", "output_cards"):
            for card in data[section_name]:
                if capture_counts[card["capture_id"]] > 1:
                    with self.subTest(section=section_name, capture_id=card["capture_id"]):
                        self.assertIn("reused", card["reuse_note"].lower())

    def test_every_card_href_resolves_to_a_public_page_and_known_anchor(self) -> None:
        data = load_showcase_data()
        id_re = re.compile(r'\bid="([^"]+)"')
        for section_name in ("product_surface_cards", "output_cards"):
            for card in data[section_name]:
                href = card["href"]
                route, _, fragment = href.partition("#")
                target_path = PUBLIC_TARGETS.get(route)
                with self.subTest(section=section_name, href=href):
                    self.assertIsNotNone(target_path, href)
                    if fragment:
                        text = target_path.read_text(encoding="utf-8")
                        ids = set(id_re.findall(text))
                        self.assertIn(fragment, ids, f"{href} should point at a declared id")

    def test_every_card_tone_has_a_matching_media_card_style(self) -> None:
        css = STYLE_PATH.read_text(encoding="utf-8")
        data = load_showcase_data()
        for section_name in ("product_surface_cards", "output_cards"):
            for card in data[section_name]:
                tone = card["tone"]
                with self.subTest(section=section_name, tone=tone):
                    self.assertIn(f".media-card-tone-{tone}", css)


if __name__ == "__main__":
    unittest.main()
