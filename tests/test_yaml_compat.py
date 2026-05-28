from __future__ import annotations

import unittest

from tools.yaml_compat import safe_load


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


if __name__ == "__main__":
    unittest.main()
