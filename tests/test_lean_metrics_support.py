import re
import sys
import tempfile
import unittest
from pathlib import Path

REPO_ROOT = Path(__file__).resolve().parents[1]
if str(REPO_ROOT) not in sys.path:
    sys.path.insert(0, str(REPO_ROOT))

from tools import lean_metrics_support


class LeanMetricsSupportTests(unittest.TestCase):
    def test_render_lean_platform_metrics_includes_key_sections(self) -> None:
        lines = lean_metrics_support.render_lean_platform_metrics(
            {
                "status": "ok",
                "command_surface_count": 5,
                "taught_commands": ["doctor", "metrics"],
                "compatibility_alias_count": 0,
                "compatibility_alias_matches": [],
                "latency_sample": {
                    "sample_count": 2,
                    "successful_sample_count": 2,
                    "avg_ms": 120.0,
                    "max_ms": 140.0,
                },
                "command_samples": [
                    {"command": "jini doctor --format json", "duration_ms": 100.0, "exit_code": 0}
                ],
                "provider_evidence": {
                    "available": True,
                    "provider_id": "local-preview",
                    "label": "Local preview",
                    "status": "ok",
                },
                "route_evidence": {
                    "available": True,
                    "adapter_count": 1,
                    "ready_adapter_count": 1,
                    "local_runtime_class": "local",
                    "captured_at": "2026-05-21T00:00:00Z",
                    "adapters": [
                        {
                            "adapter_id": "local-fast",
                            "status": "ok",
                            "latency_ms": 80,
                            "warm_latency_ms": 40,
                            "cold_start_cost_ms": 5,
                            "quality_class": "strong",
                            "structured_reliability": "strong",
                            "tokens_per_second": 22.5,
                        }
                    ],
                },
                "route_trend": {
                    "available": True,
                    "status": "measured",
                    "improving_count": 1,
                    "stable_count": 0,
                    "regressing_count": 0,
                    "adapters": [
                        {
                            "adapter_id": "local-fast",
                            "trend": "recovered",
                            "sample_count": 3,
                            "latest_latency_ms": 80,
                            "latest_tokens_per_second": 22.5,
                        }
                    ],
                },
                "route_cost": {
                    "available": True,
                    "status": "measured",
                    "basis": "local-runtime-benchmark",
                    "posture": "zero-external-api-spend",
                    "ready_adapter_count": 1,
                    "avg_ready_warm_latency_ms": 40.0,
                    "avg_ready_tokens_per_second": 22.5,
                    "cheapest_ready_adapter": {
                        "adapter_id": "local-fast",
                        "cold_start_cost_ms": 5,
                        "warm_latency_ms": 40,
                        "tokens_per_second": 22.5,
                    },
                },
                "cost_proxy": {
                    "dimension": "token-efficiency",
                    "current_score": 8.8,
                    "target_score": 9.0,
                    "competitive_gap": 0.2,
                    "strength_status": "trailing",
                },
                "latency_proxy": {
                    "dimension": "delivery-maturity",
                    "current_score": 8.9,
                    "target_score": 9.0,
                    "competitive_gap": 0.0,
                    "strength_status": "leading",
                },
            }
        )

        self.assertIn("STATUS   ok", lines)
        self.assertIn("TAUGHT", lines)
        self.assertIn("PROVIDER available=yes id=local-preview status=ok", lines)
        self.assertIn("ROUTECOST available=yes status=measured basis=local-runtime-benchmark posture=zero-external-api-spend", lines)
        self.assertTrue(any("local-fast | recovered" in line for line in lines))
        self.assertTrue(any("cheapest=local-fast" in line for line in lines))

    def test_build_lean_platform_report_assembles_surface_and_status(self) -> None:
        with tempfile.TemporaryDirectory() as tmp_dir:
            root = Path(tmp_dir)
            cli_doc = root / "cli.md"
            cli_doc.write_text("<h3><code>jini doctor</code></h3>\n<h3><code>jini status</code></h3>\n", encoding="utf-8")
            cli_entry = root / "jini.py"
            cli_entry.write_text("# placeholder\n", encoding="utf-8")

            report = lean_metrics_support.build_lean_platform_report(
                generated_at="2026-05-21T00:00:00Z",
                root=root,
                cli_path=cli_entry,
                cli_doc_path=cli_doc,
                taught_command_pattern=re.compile(r"<h3><code>jini ([^<]+)</code></h3>"),
                scan_paths=(cli_doc,),
                forbidden_aliases=("jini check",),
                session_state_root=root,
                display_path=lambda path: str(path),
                token_efficiency={"current_score": 8.8, "target_score": 9.0, "competitive_gap": 0.2, "strength_status": "trailing"},
                delivery_maturity={"current_score": 8.9, "target_score": 9.0, "competitive_gap": 0.0, "strength_status": "leading"},
            )

        self.assertEqual("2026-05-21T00:00:00Z", report["generated_at"])
        self.assertEqual(["doctor", "status"], report["taught_commands"])
        self.assertEqual(0, report["compatibility_alias_count"])
        self.assertIn("latency_sample", report)
        self.assertIn("provider_evidence", report)
        self.assertIn("route_cost", report)
        self.assertEqual("ok", report["status"])

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
