from __future__ import annotations

import unittest

from tools.yaml_compat import safe_dump, safe_load


class YamlCompatTests(unittest.TestCase):
    def test_safe_load_keeps_quoted_list_items_with_colons_as_scalars(self) -> None:
        document = safe_load(
            """
tasks:
  - "Sarah: draft the pricing update by Thursday."
  - "Amir: land the pricing page review comments by Wednesday."
  - "Priya: confirm the launch metric decision by Friday."
""".strip()
        )
        self.assertEqual(
            {
                "tasks": [
                    "Sarah: draft the pricing update by Thursday.",
                    "Amir: land the pricing page review comments by Wednesday.",
                    "Priya: confirm the launch metric decision by Friday.",
                ]
            },
            document,
        )

    def test_safe_dump_uses_plain_scalars_when_yaml_can_read_them_unambiguously(self) -> None:
        document = safe_dump(
            {
                "current_state": "decided",
                "task": "Sarah: draft the pricing update by Thursday.",
                "enabled": True,
            },
            sort_keys=False,
        )
        self.assertIn("current_state: decided", document)
        self.assertNotIn("task: Sarah: draft the pricing update by Thursday.", document)
        self.assertIn("enabled: true", document)
        self.assertEqual(
            {
                "current_state": "decided",
                "task": "Sarah: draft the pricing update by Thursday.",
                "enabled": True,
            },
            safe_load(document),
        )


if __name__ == "__main__":
    unittest.main()
