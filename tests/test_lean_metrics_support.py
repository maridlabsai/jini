import tempfile
import unittest
from pathlib import Path

from tools import lean_metrics_support


class LeanMetricsSupportTests(unittest.TestCase):
    def test_build_latency_sample_uses_only_successful_durations(self) -> None:
        sample = lean_metrics_support.build_latency_sample(
            [
                {"duration_ms": 100.0, "exit_code": 0},
                {"duration_ms": 300.0, "exit_code": 0},
                {"duration_ms": 999.0, "exit_code": 1},
            ]
        )

        self.assertEqual(3, sample["sample_count"])
        self.assertEqual(2, sample["successful_sample_count"])
        self.assertEqual(100.0, sample["min_ms"])
        self.assertEqual(300.0, sample["max_ms"])
        self.assertEqual(200.0, sample["avg_ms"])

    def test_build_route_cost_prefers_lowest_cost_ready_adapter(self) -> None:
        report = lean_metrics_support.build_route_cost(
            {
                "adapters": [
                    {
                        "adapter_id": "fast",
                        "status": "ok",
                        "cold_start_cost_ms": 10,
                        "warm_latency_ms": 120,
                        "tokens_per_second": 40.0,
                        "quality_class": "strong",
                        "structured_reliability": "strong",
                    },
                    {
                        "adapter_id": "cheap",
                        "status": "degraded",
                        "cold_start_cost_ms": 5,
                        "warm_latency_ms": 150,
                        "tokens_per_second": 20.0,
                        "quality_class": "usable",
                        "structured_reliability": "usable",
                    },
                ]
            }
        )

        self.assertTrue(report["available"])
        self.assertEqual("measured", report["status"])
        self.assertEqual("cheap", report["cheapest_ready_adapter"]["adapter_id"])
        self.assertEqual(135.0, report["avg_ready_warm_latency_ms"])
        self.assertEqual(30.0, report["avg_ready_tokens_per_second"])

    def test_scan_public_aliases_counts_forbidden_matches(self) -> None:
        with tempfile.TemporaryDirectory() as tmp_dir:
            path = Path(tmp_dir) / "cli.md"
            path.write_text("Use jini check and jini check again.\n", encoding="utf-8")

            matches = lean_metrics_support.scan_public_aliases(
                (path,),
                ("jini check",),
                lambda item: str(item),
            )

        self.assertEqual(
            [
                {
                    "path": str(path),
                    "alias": "jini check",
                    "count": 2,
                }
            ],
            matches,
        )


if __name__ == "__main__":
    unittest.main()
