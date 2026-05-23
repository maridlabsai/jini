from pathlib import Path
import unittest


REPO_ROOT = Path(__file__).resolve().parent.parent
TRACKER_PATH = REPO_ROOT / "docs" / "website-execution-tracker.md"


class WebsiteExecutionTrackerDocsTests(unittest.TestCase):
    def test_tracker_exists_and_has_required_sections(self) -> None:
        self.assertTrue(TRACKER_PATH.exists(), "website execution tracker must exist")
        text = TRACKER_PATH.read_text(encoding="utf-8")

        required_sections = [
            "# Website Execution Tracker",
            "## Current State",
            "## Website Done Criteria",
            "## Current Blocking Reality",
        ]
        for section in required_sections:
            self.assertIn(section, text)

    def test_tracker_covers_core_website_surfaces_and_next_cuts(self) -> None:
        text = TRACKER_PATH.read_text(encoding="utf-8")
        expected_rows = [
            "| Homepage | In progress |",
            "| Commercial page | In progress |",
            "| Proof page | In progress |",
            "| Install page | In progress |",
        ]
        for row in expected_rows:
            self.assertIn(row, text)

        required_next_cuts = [
            "Add tighter plan comparison and clearer subscription value framing to the commercial page",
            "Replace checked-in proof payloads with sanitized live proof ingestion",
            "Add live surface badges and release-packet links when app delivery becomes real",
            "Add release-packet links and surface badges when app delivery becomes real",
        ]
        for item in required_next_cuts:
            self.assertIn(item, text)


if __name__ == "__main__":
    unittest.main()
