import json
import os
import shutil
import subprocess
import sys
import tempfile
import unittest
from pathlib import Path
from typing import Optional, Union

REPO_ROOT = Path(__file__).resolve().parents[1]
if str(REPO_ROOT) not in sys.path:
    sys.path.insert(0, str(REPO_ROOT))

from tools.yaml_compat import safe_load


CLI = [sys.executable, str(REPO_ROOT / "tools" / "jini.py")]
RESEARCH_EXAMPLE = REPO_ROOT / "packs" / "research-prd" / "examples" / "research-prd-v1"


class JiniCliConformanceTests(unittest.TestCase):
    maxDiff = None

    def setUp(self) -> None:
        self.temp_dir = tempfile.TemporaryDirectory(prefix="jini-cli-tests-")
        self.tmp = Path(self.temp_dir.name)
        self.framework_review_existing = set((REPO_ROOT / "learning" / "framework-evolution" / "reviews").glob("*.json"))
        self.framework_experiment_existing = set((REPO_ROOT / "learning" / "framework-evolution" / "experiments").glob("*.json"))
        self.framework_outcome_existing = set((REPO_ROOT / "learning" / "framework-evolution" / "outcomes").glob("*.json"))

    def tearDown(self) -> None:
        for path in set((REPO_ROOT / "learning" / "framework-evolution" / "reviews").glob("*.json")) - self.framework_review_existing:
            path.unlink(missing_ok=True)
        for path in set((REPO_ROOT / "learning" / "framework-evolution" / "experiments").glob("*.json")) - self.framework_experiment_existing:
            path.unlink(missing_ok=True)
        for path in set((REPO_ROOT / "learning" / "framework-evolution" / "outcomes").glob("*.json")) - self.framework_outcome_existing:
            path.unlink(missing_ok=True)
        self.temp_dir.cleanup()

    def run_cli(self, *args: object, env: Optional[dict[str, str]] = None) -> subprocess.CompletedProcess[str]:
        run_env = dict(os.environ if env is None else env)
        run_env["JINI_STATE_DIR"] = str((self.tmp / ".jini").resolve())
        return subprocess.run(
            [*CLI, *[str(arg) for arg in args]],
            cwd=REPO_ROOT,
            text=True,
            capture_output=True,
            env=run_env,
        )

    def assert_ok(self, result: subprocess.CompletedProcess[str]) -> None:
        if result.returncode != 0:
            self.fail(
                f"Expected command to succeed.\nSTDOUT:\n{result.stdout}\nSTDERR:\n{result.stderr}"
            )

    def assert_error(self, result: subprocess.CompletedProcess[str]) -> None:
        if result.returncode == 0:
            self.fail(f"Expected command to fail.\nSTDOUT:\n{result.stdout}")

    def compile_research_pack(self) -> Path:
        output = self.tmp / "research-pack"
        result = self.run_cli(
            "compile-pack",
            "research-prd",
            "--work-unit-id",
            "test-research-pack",
            "--title",
            "Test Research Pack",
            "--purpose",
            "Exercise the research pack lifecycle",
            "--owner",
            "product-lead",
            "--approver",
            "eng-manager",
            "--output",
            output,
        )
        self.assert_ok(result)
        return output

    def compile_travel_pack(self) -> Path:
        output = self.tmp / "travel-pack"
        result = self.run_cli(
            "compile-pack",
            "travel-plan",
            "--work-unit-id",
            "test-travel-pack",
            "--title",
            "Test Travel Pack",
            "--purpose",
            "Exercise the travel planning pack lifecycle",
            "--owner",
            "traveler",
            "--output",
            output,
        )
        self.assert_ok(result)
        return output

    def compile_budget_pack(self) -> Path:
        output = self.tmp / "budget-pack"
        result = self.run_cli(
            "compile-pack",
            "budget-cycle",
            "--work-unit-id",
            "test-budget-pack",
            "--title",
            "Test Budget Pack",
            "--purpose",
            "Exercise the budget planning pack lifecycle",
            "--owner",
            "finance-owner",
            "--output",
            output,
        )
        self.assert_ok(result)
        return output

    def compile_meeting_pack(self) -> Path:
        output = self.tmp / "meeting-pack"
        result = self.run_cli(
            "compile-pack",
            "meeting-followup",
            "--work-unit-id",
            "test-meeting-pack",
            "--title",
            "Test Meeting Pack",
            "--purpose",
            "Exercise the meeting follow-up pack lifecycle",
            "--owner",
            "meeting-owner",
            "--approver",
            "team-lead",
            "--output",
            output,
        )
        self.assert_ok(result)
        return output

    def compile_incident_pack(self) -> Path:
        output = self.tmp / "incident-pack"
        result = self.run_cli(
            "compile-pack",
            "incident-response",
            "--work-unit-id",
            "test-incident-pack",
            "--title",
            "Test Incident Pack",
            "--purpose",
            "Exercise the incident response pack lifecycle",
            "--owner",
            "incident-commander",
            "--approver",
            "service-owner",
            "--output",
            output,
        )
        self.assert_ok(result)
        return output

    def compile_compliance_pack(self) -> Path:
        output = self.tmp / "compliance-pack"
        result = self.run_cli(
            "compile-pack",
            "compliance-audit",
            "--work-unit-id",
            "test-compliance-pack",
            "--title",
            "Test Compliance Pack",
            "--purpose",
            "Exercise the compliance audit pack lifecycle",
            "--owner",
            "compliance-lead",
            "--approver",
            "risk-officer",
            "--output",
            output,
        )
        self.assert_ok(result)
        return output

    def compile_vendor_pack(self) -> Path:
        output = self.tmp / "vendor-pack"
        result = self.run_cli(
            "compile-pack",
            "vendor-selection",
            "--work-unit-id",
            "test-vendor-pack",
            "--title",
            "Test Vendor Pack",
            "--purpose",
            "Exercise the vendor selection pack lifecycle",
            "--owner",
            "procurement-lead",
            "--approver",
            "finance-approver",
            "--output",
            output,
        )
        self.assert_ok(result)
        return output

    def copy_research_example(self) -> Path:
        target = self.tmp / "research-example"
        shutil.copytree(RESEARCH_EXAMPLE, target)
        return target

    def create_repo_fixture(self) -> Path:
        repo = self.tmp / "sample-repo"
        (repo / "scripts").mkdir(parents=True)
        (repo / "tests").mkdir()
        (repo / "README.md").write_text("# Sample Repo\n", encoding="utf-8")
        (repo / "docker-compose.yml").write_text("services:\n  app:\n    image: sample\n", encoding="utf-8")
        (repo / "Makefile").write_text(
            (
                "start:\n\tsh ./scripts/start-demo.sh\n"
                "verify:\n\tsh ./scripts/verify.sh\n"
                "demo:\n\tsh ./scripts/demo-check.sh\n"
                "test:\n\tsh ./scripts/test.sh\n"
                "build:\n\tsh ./scripts/build.sh\n"
            ),
            encoding="utf-8",
        )
        (repo / "package.json").write_text(
            json.dumps(
                {
                    "name": "sample-repo",
                    "scripts": {
                        "dev": "sh ./scripts/start-demo.sh",
                        "build": "sh ./scripts/build.sh",
                        "test": "sh ./scripts/test.sh",
                        "demo": "sh ./scripts/demo-check.sh",
                    },
                },
                indent=2,
            )
            + "\n",
            encoding="utf-8",
        )
        (repo / "scripts" / "start-demo.sh").write_text("#!/bin/sh\necho start\n", encoding="utf-8")
        (repo / "scripts" / "verify.sh").write_text("#!/bin/sh\necho verify\n", encoding="utf-8")
        (repo / "scripts" / "demo-check.sh").write_text("#!/bin/sh\necho demo\n", encoding="utf-8")
        (repo / "scripts" / "test.sh").write_text("#!/bin/sh\necho test\n", encoding="utf-8")
        (repo / "scripts" / "build.sh").write_text("#!/bin/sh\necho build\n", encoding="utf-8")
        return repo

    def create_malicious_verification_repo(self) -> Path:
        repo = self.tmp / "malicious-repo"
        (repo / "scripts").mkdir(parents=True)
        (repo / "README.md").write_text("# Malicious Repo\n", encoding="utf-8")
        (repo / "scripts" / "verify;touch PWN.sh").write_text("#!/bin/sh\necho verify\n", encoding="utf-8")
        return repo

    def create_rust_repo_fixture(self) -> Path:
        repo = self.tmp / "rust-repo"
        repo.mkdir()
        (repo / "README.md").write_text("# Rust Repo\n", encoding="utf-8")
        (repo / "Cargo.toml").write_text(
            (
                "[package]\n"
                "name = \"demo\"\n"
                "version = \"0.1.0\"\n"
                "edition = \"2021\"\n"
            ),
            encoding="utf-8",
        )
        (repo / "Justfile").write_text(
            (
                "verify:\n"
                "    cargo test\n"
                "run:\n"
                "    cargo run\n"
            ),
            encoding="utf-8",
        )
        return repo

    def install_prefix(self) -> Path:
        prefix = self.tmp / "install-root"
        prefix.mkdir()
        return prefix

    def personal_home(self) -> Path:
        return self.tmp / "personal-home"

    def create_publish_bridge_runner(self) -> Path:
        runner = self.tmp / "publish-bridge.py"
        runner.write_text(
            "\n".join(
                [
                    "#!/usr/bin/env python3",
                    "import json",
                    "import pathlib",
                    "import sys",
                    "",
                    "payload = json.loads(pathlib.Path(sys.argv[1]).read_text(encoding='utf-8'))",
                    "key = payload.get('source_task_id') or payload.get('slug') or payload.get('idempotency_key') or 'item'",
                    "result = {",
                    "    'external_id': f'ext-{key}',",
                    "    'external_url': f'https://example.test/{key}',",
                    "    'publication_status': 'published',",
                    "    'notes': [f'published {key} via bridge'],",
                    "}",
                    "print(json.dumps(result))",
                ]
            )
            + "\n",
            encoding="utf-8",
        )
        runner.chmod(0o755)
        return runner

    def read_json(self, path: Path) -> dict[str, object]:
        return json.loads(path.read_text(encoding="utf-8"))

    def read_yaml(self, path: Path) -> dict[str, object]:
        return safe_load(path.read_text(encoding="utf-8"))

    def resolve_repo_path(self, path: Union[Path, str]) -> Path:
        candidate = Path(path)
        return candidate if candidate.is_absolute() else REPO_ROOT / candidate

    def test_compile_pack_existing_output_dir_returns_friendly_error(self) -> None:
        output = self.tmp / "existing-output"
        output.mkdir()
        result = self.run_cli(
            "compile-pack",
            "research-prd",
            "--work-unit-id",
            "existing-output-pack",
            "--title",
            "Existing Output Pack",
            "--purpose",
            "Validate CLI error handling",
            "--owner",
            "product-lead",
            "--approver",
            "eng-manager",
            "--output",
            output,
        )
        self.assert_error(result)
        self.assertIn("ERROR Output directory already exists", result.stdout)
        self.assertEqual("", result.stderr)

    def test_version_flag_reports_current_version(self) -> None:
        result = self.run_cli("--version")
        self.assert_ok(result)
        self.assertEqual("jini 0.1.0", result.stdout.strip())

    def test_help_shows_curated_beginner_surface(self) -> None:
        result = self.run_cli("help")
        self.assert_ok(result)
        self.assertIn("OPEN JINI", result.stdout)
        self.assertIn("jini\n", result.stdout)
        self.assertIn("IN THE SHELL", result.stdout)
        self.assertIn("Open", result.stdout)
        self.assertIn("Plan", result.stdout)
        self.assertIn("POWER PATHS", result.stdout)
        self.assertIn("jini setup --harness codex", result.stdout)
        self.assertIn("jini doctor", result.stdout)
        self.assertIn("COMMAND CATALOG", result.stdout)
        self.assertIn("jini commands", result.stdout)
        self.assertIn("ADMIN TOOLS", result.stdout)
        self.assertNotIn("jini run --repo /path/to/repo --harness codex", result.stdout)
        self.assertNotIn("usage: jini", result.stdout)

    def test_help_all_shows_public_command_inventory(self) -> None:
        result = self.run_cli("help", "--all")
        self.assert_ok(result)
        self.assertIn("Public command inventory", result.stdout)
        self.assertIn("WORK WITH JINI", result.stdout)
        self.assertIn("jini status", result.stdout)
        self.assertIn("jini help --admin", result.stdout)
        self.assertNotIn("usage: jini", result.stdout)
        self.assertNotIn("stage-framework-experiment", result.stdout)

    def test_commands_shows_public_command_inventory(self) -> None:
        result = self.run_cli("commands")
        self.assert_ok(result)
        self.assertIn("Public command inventory", result.stdout)
        self.assertIn("GET STARTED", result.stdout)
        self.assertIn("jini try-example research-prd", result.stdout)
        self.assertIn("jini help --admin", result.stdout)

    def test_help_admin_shows_internal_inventory(self) -> None:
        result = self.run_cli("help", "--admin")
        self.assert_ok(result)
        self.assertIn("Admin and developer command inventory", result.stdout)
        self.assertIn("stage-framework-experiment", result.stdout)
        self.assertIn("capture-evidence", result.stdout)
        self.assertNotIn("usage: jini", result.stdout)

    def test_admin_help_alias_shows_internal_inventory(self) -> None:
        result = self.run_cli("admin", "help")
        self.assert_ok(result)
        self.assertIn("Admin and developer command inventory", result.stdout)
        self.assertIn("stage-framework-experiment", result.stdout)
        self.assertIn("capture-evidence", result.stdout)

    def test_admin_namespace_routes_internal_command(self) -> None:
        result = self.run_cli("admin", "list-schemas")
        self.assert_ok(result)
        self.assertIn("work-unit ->", result.stdout)

    def test_setup_command_runs_setup_surface(self) -> None:
        prefix = self.tmp / "setup-codex"
        result = self.run_cli("setup", "--harness", "codex", "--prefix", prefix)
        self.assert_ok(result)
        self.assertIn("HARNESS codex", result.stdout)
        self.assertIn("STATUS  ok", result.stdout)
        self.assertIn(f"PREFIX  {prefix.resolve()}", result.stdout)
        self.assertIn("READY", result.stdout)
        self.assertIn("jini\n", result.stdout)
        self.assertNotIn("jini status", result.stdout)
        self.assertNotIn("jini open", result.stdout)

    def test_start_alias_is_rejected(self) -> None:
        prefix = self.tmp / "setup-start-alias"
        result = self.run_cli("start", "--harness", "codex", "--prefix", prefix)
        self.assertNotEqual(0, result.returncode)

    def test_get_started_command_reports_beginner_guide(self) -> None:
        result = self.run_cli("get-started", "--harness", "codex")
        self.assert_ok(result)
        self.assertIn("BEGINNER", result.stdout)
        self.assertIn("jini", result.stdout)
        self.assertIn("entering the Jini shell", result.stdout)

    def test_try_example_command_reports_research_prd(self) -> None:
        result = self.run_cli("try-example", "research-prd", "--format", "json")
        self.assert_ok(result)
        report = json.loads(result.stdout)
        self.assertEqual("research-prd", report["example_id"])
        self.assertTrue(any(item == "jini continue" for item in report["continue_with"]))
        self.assertTrue(any(item == "jini" for item in report["continue_with"]))
        self.assertTrue(any(item == "Inside Jini: Continue" for item in report["continue_with"]))
        self.assertTrue(any(item == "Inside Jini: Open" for item in report["continue_with"]))

    def test_next_and_resume_commands_resolve(self) -> None:
        pack_dir = self.compile_research_pack()

        next_result = self.run_cli("next", pack_dir, "--format", "json")
        self.assert_ok(next_result)
        checklist = json.loads(next_result.stdout)
        self.assertEqual("research-prd", checklist["pack_id"])

        resume_result = self.run_cli("resume", pack_dir, "--format", "json")
        self.assert_ok(resume_result)
        compact = json.loads(resume_result.stdout)
        self.assertEqual("research-prd", compact["pack_id"])

    def test_continue_command_resolves_next_useful_artifact(self) -> None:
        pack_dir = self.compile_research_pack()

        result = self.run_cli("continue", "--from", pack_dir, "--print-path")
        self.assert_ok(result)
        self.assertIn(str((pack_dir / "views" / "tasks.md").resolve()), result.stdout.strip())

    def test_pathless_continue_uses_current_work(self) -> None:
        pack_dir = self.compile_meeting_pack()
        self.assert_ok(self.run_cli("status", pack_dir))

        result = self.run_cli("continue", "--print-path")
        self.assert_ok(result)
        self.assertIn(str((pack_dir / "views" / "tasks.md").resolve()), result.stdout.strip())

    def test_continue_command_uses_compact_preview_surface(self) -> None:
        pack_dir = self.compile_research_pack()

        continue_result = self.run_cli("continue", "--from", pack_dir)
        self.assert_ok(continue_result)
        self.assertIn("NEXT   tasks", continue_result.stdout)
        self.assertIn("# Tasks:", continue_result.stdout)
        self.assertIn("jini show tasks --from", continue_result.stdout)
        self.assertIn("jini open tasks --from", continue_result.stdout)

        resume_result = self.run_cli("resume", pack_dir, "--format", "json", "--max-chars", "700")
        self.assert_ok(resume_result)
        self.assertLess(len(continue_result.stdout), len(resume_result.stdout))

    def test_status_view_reports_plain_questions_and_follow_on_commands(self) -> None:
        pack_dir = self.compile_research_pack()
        result = self.run_cli("status", pack_dir)
        self.assert_ok(result)
        self.assertIn("WHAT IS DONE?", result.stdout)
        self.assertIn("WHAT HAPPENS NEXT?", result.stdout)
        self.assertIn("READY NOW", result.stdout)
        self.assertIn("CONTINUE", result.stdout)
        self.assertIn("jini continue", result.stdout)
        self.assertNotIn("jini resume ", result.stdout)

    def test_example_sets_current_work_for_pathless_status_and_artifacts(self) -> None:
        example_output = self.tmp / "research-example"
        result = self.run_cli("try-example", "research-prd", "--output", example_output)
        self.assert_ok(result)

        status = self.run_cli("status")
        self.assert_ok(status)
        self.assertIn("WORK   example-research-prd", status.stdout)
        self.assertIn("READY NOW", status.stdout)

        artifacts = self.run_cli("artifacts")
        self.assert_ok(artifacts)
        self.assertIn("READY NOW", artifacts.stdout)
        self.assertIn("prd", artifacts.stdout)
        self.assertIn("tasks", artifacts.stdout)

    def test_compile_pack_creates_canonical_session_files(self) -> None:
        pack_dir = self.compile_research_pack()
        session_dir = self.tmp / ".jini" / "sessions" / "test-research-pack"
        self.assertTrue((session_dir / "session.yaml").exists())
        self.assertTrue((session_dir / "projection.json").exists())
        self.assertTrue((session_dir / "events.ndjson").exists())

        session_doc = safe_load((session_dir / "session.yaml").read_text(encoding="utf-8"))
        self.assertEqual("test-research-pack", session_doc["session_id"])
        self.assertEqual(str(pack_dir.resolve()), session_doc["pack_dir"])
        projection_doc = json.loads((session_dir / "projection.json").read_text(encoding="utf-8"))
        self.assertEqual("test-research-pack", projection_doc["session_id"])
        self.assertIn("ready", projection_doc)

    def test_show_artifact_uses_current_work_without_path(self) -> None:
        pack_dir = self.compile_travel_pack()

        status = self.run_cli("status", pack_dir)
        self.assert_ok(status)

        show = self.run_cli("show", "itinerary")
        self.assert_ok(show)
        self.assertIn("# Itinerary: Test Travel Pack", show.stdout)

        open_result = self.run_cli("open", "itinerary", "--print-path")
        self.assert_ok(open_result)
        self.assertIn(str((pack_dir / "views" / "itinerary.md").resolve()), open_result.stdout.strip())

    def test_pathless_status_and_open_can_resume_from_canonical_session_pointer(self) -> None:
        pack_dir = self.compile_travel_pack()
        current_work = self.tmp / ".jini" / "current-work.json"
        current_work.unlink()

        status = self.run_cli("status")
        self.assert_ok(status)
        self.assertIn("WORK   test-travel-pack", status.stdout)

        open_result = self.run_cli("open", "itinerary", "--print-path")
        self.assert_ok(open_result)
        self.assertIn(str((pack_dir / "views" / "itinerary.md").resolve()), open_result.stdout.strip())

    def test_pathless_status_falls_back_to_saved_session_projection_when_pack_is_missing(self) -> None:
        pack_dir = self.compile_travel_pack()
        shutil.rmtree(pack_dir)
        current_work = self.tmp / ".jini" / "current-work.json"
        current_work.unlink()

        result = self.run_cli("status", "--format", "json")
        self.assert_ok(result)
        report = json.loads(result.stdout)
        self.assertEqual("session-only", report["health"])
        self.assertEqual("test-travel-pack", report["work_unit_id"])
        self.assertTrue(report["ready_now"])
        self.assertIn("saved session projection", report["validation_warnings"][0])
        self.assertEqual(["jini continue"], report["continue_with"])

    def test_pathless_continue_falls_back_to_saved_session_projection_when_pack_is_missing(self) -> None:
        pack_dir = self.compile_travel_pack()
        shutil.rmtree(pack_dir)
        current_work = self.tmp / ".jini" / "current-work.json"
        current_work.unlink()

        result = self.run_cli("continue")
        self.assert_ok(result)
        self.assertIn("WORK   test-travel-pack", result.stdout)
        self.assertIn("CLASS  projection-continue", result.stdout)
        self.assertIn("CONTINUE", result.stdout)
        self.assertIn("saved session projection", result.stdout)

    def test_pathless_resume_falls_back_to_saved_session_projection_when_pack_is_missing(self) -> None:
        pack_dir = self.compile_travel_pack()
        shutil.rmtree(pack_dir)
        current_work = self.tmp / ".jini" / "current-work.json"
        current_work.unlink()

        result = self.run_cli("resume", "--format", "json")
        self.assert_ok(result)
        compact = json.loads(result.stdout)
        self.assertEqual("projection-resume", compact["execution_class"])
        self.assertEqual("test-travel-pack", compact["work_unit_id"])
        self.assertTrue(compact["resume_items"])
        self.assertIn("saved session projection", compact["stale_signals"][0])

    def test_status_without_current_work_fails_cleanly(self) -> None:
        result = self.run_cli("status")
        self.assert_error(result)
        self.assertIn("Nothing is in progress yet. Run `jini` to start something", result.stdout)

    def test_status_missing_path_returns_friendly_error(self) -> None:
        result = self.run_cli("status", self.tmp / "missing-pack")
        self.assert_error(result)
        self.assertIn("ERROR Pack path is missing required Jini files", result.stdout)
        self.assertEqual("", result.stderr)

    def test_harnesses_command_lists_supported_harnesses(self) -> None:
        result = self.run_cli("harnesses", "--format", "json")
        self.assert_ok(result)
        report = json.loads(result.stdout)
        harness_ids = {item["id"] for item in report["harnesses"]}
        self.assertIn("codex", harness_ids)
        self.assertIn("claude-code", harness_ids)
        codex = next(item for item in report["harnesses"] if item["id"] == "codex")
        self.assertEqual("jini", codex["start_command"])

    def test_recommend_execution_command_reports_pack(self) -> None:
        pack_dir = self.compile_research_pack()
        result = self.run_cli("recommend-execution", pack_dir, "--format", "json")
        self.assert_ok(result)
        report = json.loads(result.stdout)
        self.assertEqual("research-prd", report["pack_id"])

    def test_bootstrap_home_materializes_personal_os_scaffold(self) -> None:
        home = self.personal_home()
        result = self.run_cli(
            "bootstrap-home",
            home,
            "--owner-name",
            "Sharad",
            "--assistant-name",
            "Jini",
            "--format",
            "json",
        )
        self.assert_ok(result)

        payload = json.loads(result.stdout)
        self.assertEqual(str(home.resolve()), payload["home_root"])
        self.assertTrue((home / "home.yaml").exists())
        self.assertTrue((home / "soul.md").exists())
        self.assertTrue((home / "user.md").exists())
        self.assertTrue((home / "tools.md").exists())
        self.assertTrue((home / "memory" / "daily").is_dir())
        self.assertTrue((home / "memory" / "long-term.md").exists())
        self.assertTrue((home / "routines" / "local" / "dream-memory.yaml").exists())
        self.assertTrue((home / "routines" / "remote" / "weekly-planning.yaml").exists())

    def test_bootstrap_steering_materializes_foundational_docs(self) -> None:
        repo = self.create_repo_fixture()
        result = self.run_cli("bootstrap-steering", repo, "--format", "json")
        self.assert_ok(result)

        payload = json.loads(result.stdout)
        steering_root = Path(payload["steering_root"])
        self.assertTrue((steering_root / "product.md").exists())
        self.assertTrue((steering_root / "tech.md").exists())
        self.assertTrue((steering_root / "structure.md").exists())
        self.assertTrue((steering_root / "testing.md").exists())

    def test_show_steering_reports_active_bootstrapped_docs(self) -> None:
        repo = self.create_repo_fixture()
        self.assert_ok(self.run_cli("bootstrap-steering", repo))

        result = self.run_cli("show-steering", repo, "--format", "json")
        self.assert_ok(result)

        summary = json.loads(result.stdout)
        self.assertTrue(summary["found"])
        ids = [item["id"] for item in summary["documents"]]
        self.assertIn("product", ids)
        self.assertIn("tech", ids)
        active_ids = [item["id"] for item in summary["documents"] if item["active"]]
        self.assertIn("product", active_ids)
        self.assertIn("tech", active_ids)

    def test_append_memory_and_dream_memory_create_long_term_summary(self) -> None:
        home = self.personal_home()
        self.assert_ok(self.run_cli("bootstrap-home", home, "--owner-name", "Sharad"))

        self.assert_ok(
            self.run_cli(
                "append-memory",
                home,
                "--line",
                "Prefers morning planning and short execution checklists.",
                "--date",
                "2026-05-11",
            )
        )
        self.assert_ok(
            self.run_cli(
                "append-memory",
                home,
                "--line",
                "Repeatedly asks for token-efficient resumptions.",
                "--date",
                "2026-05-11",
            )
        )

        result = self.run_cli("dream-memory", home, "--format", "json")
        self.assert_ok(result)

        payload = json.loads(result.stdout)
        self.assertGreaterEqual(payload["source_line_count"], 2)
        long_term = (home / "memory" / "long-term.md").read_text(encoding="utf-8")
        self.assertIn("Prefers morning planning", long_term)
        self.assertIn("token-efficient resumptions", long_term)

    def test_memory_status_reports_budget_and_compaction_recommendation(self) -> None:
        home = self.personal_home()
        self.assert_ok(self.run_cli("bootstrap-home", home))
        for index in range(45):
            self.assert_ok(
                self.run_cli(
                    "append-memory",
                    home,
                    "--line",
                    f"Durable note {index} about delivery and repo habits.",
                )
            )

        before = self.run_cli("memory-status", home, "--format", "json")
        self.assert_ok(before)
        status_before = json.loads(before.stdout)
        self.assertEqual("dream-memory", status_before["recommended_action"])

        result = self.run_cli("dream-memory", home, "--format", "json")
        self.assert_ok(result)
        dream = json.loads(result.stdout)
        self.assertLessEqual(dream["char_count"], dream["char_limit"])

        after = self.run_cli("memory-status", home, "--format", "json")
        self.assert_ok(after)
        status_after = json.loads(after.stdout)
        self.assertLessEqual(status_after["long_term_chars"], status_after["long_term_char_limit"])

    def test_list_tools_reports_seeded_inventory(self) -> None:
        home = self.personal_home()
        self.assert_ok(self.run_cli("bootstrap-home", home))

        result = self.run_cli("list-tools", home, "--format", "json")
        self.assert_ok(result)

        payload = json.loads(result.stdout)
        ids = [item["id"] for item in payload["tools"]]
        self.assertIn("jini-cli", ids)
        self.assertIn("adapter-registry", ids)
        self.assertIn("golden-benchmark", ids)
        self.assertIn("framework-evolution", ids)

    def test_list_routines_reports_local_and_remote_routines(self) -> None:
        home = self.personal_home()
        self.assert_ok(self.run_cli("bootstrap-home", home))

        result = self.run_cli("list-routines", home, "--format", "json")
        self.assert_ok(result)

        payload = json.loads(result.stdout)
        routine_ids = [item["routine_id"] for item in payload["routines"]]
        self.assertIn("dream-memory", routine_ids)
        self.assertIn("daily-brief", routine_ids)
        self.assertIn("golden-benchmark", routine_ids)
        self.assertIn("framework-review", routine_ids)
        self.assertIn("publish-readiness", routine_ids)
        self.assertIn("weekly-planning", routine_ids)

    def test_run_routine_local_builtin_emits_brief(self) -> None:
        home = self.personal_home()
        self.assert_ok(self.run_cli("bootstrap-home", home, "--owner-name", "Sharad"))
        self.assert_ok(
            self.run_cli(
                "append-memory",
                home,
                "--line",
                "Needs a concise daily planning brief.",
            )
        )

        result = self.run_cli(
            "run-routine",
            home,
            "daily-brief",
            "--mode",
            "local",
            "--format",
            "json",
        )
        self.assert_ok(result)

        payload = json.loads(result.stdout)
        self.assertEqual("executed", payload["status"])
        output_path = Path(payload["output_paths"][0])
        self.assertTrue(output_path.exists())
        self.assertIn("Daily Brief", output_path.read_text(encoding="utf-8"))

    def test_run_routine_publish_readiness_emits_release_brief(self) -> None:
        home = self.personal_home()
        self.assert_ok(self.run_cli("bootstrap-home", home, "--owner-name", "Sharad"))

        result = self.run_cli(
            "run-routine",
            home,
            "publish-readiness",
            "--mode",
            "local",
            "--format",
            "json",
        )
        self.assert_ok(result)

        payload = json.loads(result.stdout)
        self.assertEqual("executed", payload["status"])
        output_path = Path(payload["output_paths"][0])
        self.assertTrue(output_path.exists())
        self.assertIn("Publish Readiness", output_path.read_text(encoding="utf-8"))

    def test_run_routine_golden_benchmark_emits_benchmark_brief(self) -> None:
        home = self.personal_home()
        self.assert_ok(self.run_cli("bootstrap-home", home, "--owner-name", "Sharad"))

        result = self.run_cli(
            "run-routine",
            home,
            "golden-benchmark",
            "--mode",
            "local",
            "--format",
            "json",
        )
        self.assert_ok(result)

        payload = json.loads(result.stdout)
        self.assertEqual("executed", payload["status"])
        output_path = Path(payload["output_paths"][0])
        self.assertTrue(output_path.exists())
        text = output_path.read_text(encoding="utf-8")
        self.assertIn("Golden Benchmark", text)
        self.assertIn("Kiro", text)
        self.assertIn("Hermes", text)

    def test_run_routine_framework_review_emits_review_brief(self) -> None:
        home = self.personal_home()
        self.assert_ok(self.run_cli("bootstrap-home", home, "--owner-name", "Sharad"))

        result = self.run_cli(
            "run-routine",
            home,
            "framework-review",
            "--mode",
            "local",
            "--format",
            "json",
        )
        self.assert_ok(result)

        payload = json.loads(result.stdout)
        self.assertEqual("executed", payload["status"])
        output_path = Path(payload["output_paths"][0])
        self.assertTrue(output_path.exists())
        text = output_path.read_text(encoding="utf-8")
        self.assertIn("Framework Review", text)
        self.assertIn("Best Next Dimension", text)

    def test_run_routine_remote_stages_receipt(self) -> None:
        home = self.personal_home()
        self.assert_ok(self.run_cli("bootstrap-home", home))

        result = self.run_cli(
            "run-routine",
            home,
            "weekly-planning",
            "--mode",
            "remote",
            "--format",
            "json",
        )
        self.assert_ok(result)

        payload = json.loads(result.stdout)
        self.assertEqual("staged-remote", payload["status"])
        receipt_path = Path(payload["receipt_path"])
        self.assertTrue(receipt_path.exists())
        receipt = self.read_json(receipt_path)
        self.assertEqual("weekly-planning", receipt["routine_id"])

    def test_run_routine_shell_requires_trusted_local_flag(self) -> None:
        home = self.personal_home()
        self.assert_ok(self.run_cli("bootstrap-home", home, "--owner-name", "Sharad"))
        routine_path = home / "routines" / "local" / "unsafe-shell.yaml"
        routine_path.write_text(
            "\n".join(
                [
                    "schema_version: 0.1.0",
                    "routine_id: unsafe-shell",
                    "title: Unsafe Shell",
                    "mode: local",
                    "runner: shell",
                    "enabled: true",
                    "summary: test shell routine gate",
                    "command: echo unsafe",
                ]
            )
            + "\n",
            encoding="utf-8",
        )

        result = self.run_cli("run-routine", home, "unsafe-shell", "--mode", "local")
        self.assert_error(result)
        self.assertIn("trusted_local: true", result.stdout)

    def test_bind_home_surfaces_long_term_memory_in_recommendation(self) -> None:
        pack_dir = self.compile_research_pack()
        home = self.personal_home()
        self.assert_ok(self.run_cli("bootstrap-home", home, "--owner-name", "Sharad"))
        self.assert_ok(
            self.run_cli(
                "append-memory",
                home,
                "--line",
                "Prefers low-token resumptions and home-bound execution guidance.",
                "--date",
                "2026-05-11",
            )
        )
        self.assert_ok(self.run_cli("dream-memory", home))
        self.assert_ok(self.run_cli("bind-home", pack_dir, "--home", home))

        result = self.run_cli("recommend-execution", pack_dir, "--format", "json")
        self.assert_ok(result)

        recommendation = json.loads(result.stdout)
        home_context = recommendation["memory_context"]["home"]
        self.assertTrue(home_context["bound"])
        self.assertEqual(str(home.resolve()), home_context["home_root"])
        self.assertTrue(
            any("low-token resumptions" in item for item in home_context["long_term_memory"])
        )
        self.assertTrue(
            any("low-token resumptions" in item for item in recommendation["memory_context"]["resume_items"])
        )

    def test_recommend_execution_can_prefer_runtime_target(self) -> None:
        pack_dir = self.compile_research_pack()

        result = self.run_cli(
            "recommend-execution",
            pack_dir,
            "--runtime-target",
            "kiro-cli",
            "--format",
            "json",
        )
        self.assert_ok(result)

        recommendation = json.loads(result.stdout)
        runtime_guidance = recommendation["runtime_guidance"]
        self.assertEqual("kiro-cli", runtime_guidance["selected"]["id"])
        self.assertIn("codex", runtime_guidance["fallbacks"])

    def test_compact_context_includes_home_memory_and_compression_ratio(self) -> None:
        pack_dir = self.compile_research_pack()
        home = self.personal_home()
        self.assert_ok(self.run_cli("bootstrap-home", home))
        self.assert_ok(
            self.run_cli(
                "append-memory",
                home,
                "--line",
                "Needs compact memory reloads for verification work.",
            )
        )
        self.assert_ok(self.run_cli("dream-memory", home))
        self.assert_ok(self.run_cli("bind-home", pack_dir, "--home", home))

        result = self.run_cli(
            "resume",
            pack_dir,
            "--max-chars",
            "700",
            "--format",
            "json",
        )
        self.assert_ok(result)

        compact = json.loads(result.stdout)
        self.assertTrue(any("compact memory reloads" in item for item in compact["home_memory"]))
        self.assertIn("compression_ratio", compact["token_budget"])
        self.assertGreater(compact["token_budget"]["compression_ratio"], 0)
        self.assertLessEqual(compact["token_budget"]["compression_ratio"], 1)

    def test_run_pack_appends_daily_memory_when_home_is_bound(self) -> None:
        pack_dir = self.compile_research_pack()
        home = self.personal_home()
        self.assert_ok(self.run_cli("bootstrap-home", home))
        self.assert_ok(self.run_cli("bind-home", pack_dir, "--home", home))

        result = self.run_cli("run-pack", pack_dir, "--mode", "supervised", "--format", "json")
        self.assert_ok(result)

        report = json.loads(result.stdout)
        self.assertTrue(report["home_binding"]["bound"])
        self.assertTrue(report["memory_append"]["appended"])
        daily_files = sorted((home / "memory" / "daily").glob("*.md"))
        self.assertTrue(daily_files)
        latest_daily = daily_files[-1].read_text(encoding="utf-8")
        self.assertIn("run-pack", latest_daily)
        self.assertIn("Test Research Pack", latest_daily)

    def test_harvest_evidence_appends_daily_memory_when_home_is_bound(self) -> None:
        pack_dir = self.compile_research_pack()
        repo = self.create_repo_fixture()
        home = self.personal_home()
        self.assert_ok(self.run_cli("bootstrap-home", home))
        self.assert_ok(self.run_cli("bind-home", pack_dir, "--home", home))

        result = self.run_cli(
            "harvest-evidence",
            pack_dir,
            "--author",
            "release-bot",
            "--repo",
            repo,
            "--category",
            "verify",
            "--format",
            "json",
        )
        self.assert_ok(result)

        report = json.loads(result.stdout)
        self.assertTrue(report["memory_append"]["appended"])
        daily_files = sorted((home / "memory" / "daily").glob("*.md"))
        self.assertTrue(daily_files)
        latest_daily = daily_files[-1].read_text(encoding="utf-8")
        self.assertIn("harvest", latest_daily.lower())
        self.assertIn("ready", latest_daily)

    def test_learning_snapshot_reports_memory_writes_and_avg_compaction(self) -> None:
        pack_dir = self.compile_research_pack()
        repo = self.create_repo_fixture()
        home = self.personal_home()
        self.assert_ok(self.run_cli("bootstrap-home", home))
        self.assert_ok(self.run_cli("bind-home", pack_dir, "--home", home))
        self.assert_ok(self.run_cli("resume", pack_dir, "--repo", repo))
        self.assert_ok(self.run_cli("run-pack", pack_dir, "--mode", "supervised"))

        result = self.run_cli("learning-snapshot", pack_dir, "--format", "json")
        self.assert_ok(result)

        snapshot = json.loads(result.stdout)
        self.assertGreaterEqual(snapshot["memory_write_count"], 1)
        self.assertGreater(snapshot["average_compaction_ratio"], 0)

    def test_compiled_research_pack_reports_expected_status_and_recommendation(self) -> None:
        pack_dir = self.compile_research_pack()

        status_result = self.run_cli("status-pack", pack_dir)
        self.assert_ok(status_result)
        self.assertIn("PACK   research-prd", status_result.stdout)
        self.assertIn("STATE  decided", status_result.stdout)
        self.assertIn("HEALTH ready-to-make", status_result.stdout)

        recommendation_result = self.run_cli("recommend-execution", pack_dir, "--format", "json")
        self.assert_ok(recommendation_result)
        recommendation = json.loads(recommendation_result.stdout)
        self.assertEqual("research-prd", recommendation["pack_id"])
        self.assertEqual("make", recommendation["intent"])
        self.assertEqual("standard", recommendation["execution_class"])

    def test_compile_pack_materializes_local_execution_surfaces(self) -> None:
        pack_dir = self.compile_research_pack()

        self.assertTrue((pack_dir / "views" / "prd.md").exists())
        self.assertTrue((pack_dir / "views" / "tasks.md").exists())
        self.assertTrue((pack_dir / "exports" / "tasks-sync.json").exists())
        self.assertTrue((pack_dir / "exports" / "issues" / "jira" / "issues.json").exists())
        self.assertTrue((pack_dir / "exports" / "issues" / "github" / "issues.json").exists())
        self.assertTrue((pack_dir / "exports" / "wiki" / "markdown" / "pages.json").exists())
        self.assertTrue((pack_dir / "exports" / "wiki" / "confluence" / "pages.json").exists())

    def test_compile_meeting_pack_materializes_followup_and_task_surfaces(self) -> None:
        pack_dir = self.compile_meeting_pack()

        self.assertTrue((pack_dir / "views" / "followup.md").exists())
        self.assertTrue((pack_dir / "views" / "tasks.md").exists())
        self.assertTrue((pack_dir / "artifacts" / "06-tasks.yaml").exists())
        self.assertTrue((pack_dir / "exports" / "tasks-sync.json").exists())

    def test_compile_travel_pack_materializes_itinerary_and_exports(self) -> None:
        pack_dir = self.compile_travel_pack()

        self.assertTrue((pack_dir / "views" / "itinerary.md").exists())
        self.assertTrue((pack_dir / "views" / "budget-sketch.md").exists())
        self.assertTrue((pack_dir / "views" / "travel-logistics.md").exists())
        self.assertTrue((pack_dir / "views" / "tasks.md").exists())
        self.assertTrue((pack_dir / "exports" / "tasks-sync.json").exists())
        self.assertTrue((pack_dir / "exports" / "wiki" / "markdown" / "pages.json").exists())
        itinerary = (pack_dir / "views" / "itinerary.md").read_text(encoding="utf-8")
        self.assertIn("## Day-by-day draft", itinerary)
        self.assertIn("### Day 1", itinerary)
        self.assertIn("## Budget sketch", itinerary)
        self.assertIn("## Logistics to lock", itinerary)
        self.assertIn("## If something changes", itinerary)

    def test_compile_budget_pack_materializes_budget_view_and_exports(self) -> None:
        pack_dir = self.compile_budget_pack()

        self.assertTrue((pack_dir / "views" / "budget.md").exists())
        self.assertTrue((pack_dir / "views" / "tasks.md").exists())
        self.assertTrue((pack_dir / "exports" / "tasks-sync.json").exists())
        self.assertTrue((pack_dir / "exports" / "wiki" / "markdown" / "pages.json").exists())

    def test_compile_meeting_pack_materializes_followup_view_and_exports(self) -> None:
        pack_dir = self.compile_meeting_pack()

        self.assertTrue((pack_dir / "views" / "followup.md").exists())
        self.assertTrue((pack_dir / "views" / "tasks.md").exists())
        self.assertTrue((pack_dir / "exports" / "tasks-sync.json").exists())
        self.assertTrue((pack_dir / "exports" / "wiki" / "markdown" / "pages.json").exists())
        followup = (pack_dir / "views" / "followup.md").read_text(encoding="utf-8")
        self.assertIn("# Sendable Follow-Up:", followup)
        self.assertIn("## Send this", followup)
        self.assertIn("## Owners and due points", followup)
        self.assertIn("## Open questions", followup)
        self.assertIn("## What happens next", followup)
        self.assertIn("Pricing draft moves to Sarah by Thursday", followup)
        self.assertIn("Amir: land the pricing page review comments by Wednesday", followup)
        self.assertIn("Priya: confirm the launch metric decision by Friday", followup)
        self.assertIn("Should launch success be measured by sign-ups or paid conversion", followup)
        self.assertIn("Do we need legal review before publishing pricing changes", followup)

        work_unit = safe_load((pack_dir / "work-unit.yaml").read_text(encoding="utf-8"))
        self.assertEqual("Delivery", work_unit["profile_id"])
        self.assertIn("Business:team-operations", work_unit["active_extensions"])

    def test_compile_incident_pack_materializes_response_view_and_exports(self) -> None:
        pack_dir = self.compile_incident_pack()

        self.assertTrue((pack_dir / "views" / "response.md").exists())
        self.assertTrue((pack_dir / "views" / "tasks.md").exists())
        self.assertTrue((pack_dir / "exports" / "tasks-sync.json").exists())
        self.assertTrue((pack_dir / "exports" / "wiki" / "markdown" / "pages.json").exists())

    def test_compile_compliance_pack_materializes_audit_view_and_exports(self) -> None:
        pack_dir = self.compile_compliance_pack()

        self.assertTrue((pack_dir / "views" / "audit.md").exists())
        self.assertTrue((pack_dir / "views" / "tasks.md").exists())
        self.assertTrue((pack_dir / "exports" / "tasks-sync.json").exists())
        self.assertTrue((pack_dir / "exports" / "wiki" / "markdown" / "pages.json").exists())

    def test_compile_vendor_pack_materializes_selection_view_and_exports(self) -> None:
        pack_dir = self.compile_vendor_pack()

        self.assertTrue((pack_dir / "views" / "selection.md").exists())
        self.assertTrue((pack_dir / "views" / "tasks.md").exists())
        self.assertTrue((pack_dir / "exports" / "tasks-sync.json").exists())
        self.assertTrue((pack_dir / "exports" / "wiki" / "markdown" / "pages.json").exists())

    def test_run_pack_supervised_blocks_local_exports_without_write_consent(self) -> None:
        pack_dir = self.compile_research_pack()

        result = self.run_cli("run-pack", pack_dir, "--mode", "supervised", "--format", "json")
        self.assert_ok(result)
        report = json.loads(result.stdout)

        self.assertEqual("decided", report["state_before"])
        self.assertEqual("decided", report["state_after"])
        self.assertEqual([], report["consent"]["newly_granted"])
        self.assertTrue(any(action["command"] == "local-exports" and action["status"] == "blocked" for action in report["actions"]))
        self.assertTrue(any(action["command"] == "advance-pack" and action["status"] == "planned" for action in report["actions"]))
        self.assertTrue((pack_dir / "runtime" / "last-run.json").exists())

    def test_run_pack_autonomous_advances_one_state_with_write_and_command_consent(self) -> None:
        pack_dir = self.compile_research_pack()

        result = self.run_cli(
            "run-pack",
            pack_dir,
            "--mode",
            "autonomous",
            "--consent",
            "write",
            "--consent",
            "command",
            "--format",
            "json",
        )
        self.assert_ok(result)
        report = json.loads(result.stdout)
        consent = self.read_json(pack_dir / "runtime" / "consent.json")

        self.assertEqual("decided", report["state_before"])
        self.assertEqual("in_make", report["state_after"])
        self.assertTrue(any(action["command"] == "advance-pack" and action["status"] == "executed" for action in report["actions"]))
        self.assertEqual({"write": True, "command": True, "publish": False}, consent["categories"])

    def test_advance_pack_blocks_verify_transition_when_tasks_are_unresolved(self) -> None:
        pack_dir = self.compile_research_pack()

        first = self.run_cli("advance-pack", pack_dir)
        self.assert_ok(first)

        second = self.run_cli("advance-pack", pack_dir)
        self.assert_error(second)
        self.assertIn("requires all tasks done", second.stdout)

    def test_capture_output_allows_verify_transition_after_tasks_complete(self) -> None:
        pack_dir = self.compile_research_pack()

        for task_index in range(1, 4):
            result = self.run_cli(
                "capture-output",
                pack_dir,
                "--author",
                "delivery-bot",
                "--task-index",
                task_index,
                "--status",
                "done",
                "--note",
                f"Completed task {task_index}",
                "--reference",
                f"artifacts/task-{task_index}.md",
            )
            self.assert_ok(result)

        self.assert_ok(self.run_cli("advance-pack", pack_dir))
        self.assert_ok(self.run_cli("advance-pack", pack_dir))

        status_result = self.run_cli("status-pack", pack_dir)
        self.assert_ok(status_result)
        self.assertIn("STATE  awaiting_verification", status_result.stdout)
        self.assertIn("HEALTH ready-to-verify", status_result.stdout)
        self.assertIn("done:       3/3", status_result.stdout)

    def test_capture_evidence_creates_artifact_for_compiled_research_pack(self) -> None:
        pack_dir = self.compile_research_pack()

        result = self.run_cli(
            "capture-evidence",
            pack_dir,
            "--author",
            "release-coordinator",
            "--claim",
            "Startup flow is documented",
            "--test-result",
            "Smoke tests passed on the active revision",
            "--review-result",
            "Runbook matches the benchmark topology",
            "--operational-result",
            "Health endpoints are usable for verification",
            "--risk",
            "Local validation does not prove production scale",
        )
        self.assert_ok(result)

        evidence_files = sorted((pack_dir / "artifacts").glob("*-evidence.yaml"))
        self.assertEqual(1, len(evidence_files))

        validate_result = self.run_cli("validate-pack", pack_dir)
        self.assert_ok(validate_result)

    def test_publish_wiki_falls_back_without_binding(self) -> None:
        pack_dir = self.compile_research_pack()

        result = self.run_cli("publish-wiki", pack_dir, "--adapter", "confluence")
        self.assert_ok(result)

        publish_plan = self.read_json(pack_dir / "exports" / "publish" / "wiki" / "confluence" / "publish-plan.json")
        self.assertEqual("markdown-fallback", publish_plan["execution_mode"])
        self.assertFalse(publish_plan["target"]["configured"])

    def test_binding_enables_connector_ready_publish_plans(self) -> None:
        pack_dir = self.compile_research_pack()

        bind_result = self.run_cli(
            "bind-atlassian",
            pack_dir,
            "--cloud-id",
            "cloud-123",
            "--site-url",
            "https://example.atlassian.net",
            "--project-key",
            "DEMO",
            "--space-key",
            "DOCS",
            "--space-id",
            "456",
        )
        self.assert_ok(bind_result)

        self.assert_ok(self.run_cli("publish-issues", pack_dir, "--adapter", "jira"))
        self.assert_ok(self.run_cli("publish-wiki", pack_dir, "--adapter", "confluence"))

        issue_plan = self.read_json(pack_dir / "exports" / "publish" / "issues" / "jira" / "publish-plan.json")
        wiki_plan = self.read_json(pack_dir / "exports" / "publish" / "wiki" / "confluence" / "publish-plan.json")
        self.assertEqual("connector-ready", issue_plan["execution_mode"])
        self.assertTrue(issue_plan["target"]["configured"])
        self.assertEqual("connector-ready", wiki_plan["execution_mode"])
        self.assertTrue(wiki_plan["target"]["configured"])

    def test_publish_issues_github_can_be_applied_locally(self) -> None:
        pack_dir = self.compile_research_pack()

        self.assert_ok(self.run_cli("publish-issues", pack_dir, "--adapter", "github"))

        result = self.run_cli(
            "apply-publish-plan",
            pack_dir / "exports" / "publish" / "issues" / "github",
            "--format",
            "json",
        )
        self.assert_ok(result)

        receipt = json.loads(result.stdout)
        self.assertEqual("applied-local", receipt["status"])
        self.assertTrue(Path(receipt["receipt_path"]).exists())
        self.assertTrue(receipt["applied_paths"])
        first_issue = Path(receipt["applied_paths"][0])
        self.assertTrue(first_issue.exists())
        self.assertIn("## Description", first_issue.read_text(encoding="utf-8"))

    def test_publish_issues_github_can_apply_locally_in_one_command(self) -> None:
        pack_dir = self.compile_research_pack()

        result = self.run_cli(
            "publish-issues",
            pack_dir,
            "--adapter",
            "github",
            "--apply-local",
            "--format",
            "json",
        )
        self.assert_ok(result)

        receipt = json.loads(result.stdout)
        self.assertEqual("applied-local", receipt["status"])
        self.assertTrue(receipt["applied_paths"])
        self.assertTrue(Path(receipt["receipt_path"]).exists())

    def test_publish_issues_jira_can_execute_via_bridge_and_capture_publication(self) -> None:
        pack_dir = self.compile_research_pack()
        runner = self.create_publish_bridge_runner()

        result = self.run_cli(
            "publish-issues",
            pack_dir,
            "--adapter",
            "jira",
            "--bridge-runner",
            runner,
            "--format",
            "json",
        )
        self.assert_ok(result)

        receipt = json.loads(result.stdout)
        self.assertEqual("executed", receipt["status"])
        self.assertEqual("jira", receipt["adapter"])
        self.assertTrue(Path(receipt["receipt_path"]).exists())
        self.assertTrue(Path(receipt["result_path"]).exists())
        bundle = self.read_json(Path(receipt["result_path"]))
        self.assertEqual(3, len(bundle["records"]))
        self.assertEqual("issue", bundle["records"][0]["target_kind"])

        capture = self.run_cli(
            "capture-publication",
            pack_dir,
            "--author",
            "publish-bot",
            "--input",
            receipt["result_path"],
            "--scope",
            "jira-issues",
        )
        self.assert_ok(capture)
        publication_files = sorted((pack_dir / "artifacts").glob("*-publication.yaml"))
        self.assertEqual(1, len(publication_files))

    def test_execute_publish_plan_runs_bridge_runner_directly(self) -> None:
        pack_dir = self.compile_research_pack()
        runner = self.create_publish_bridge_runner()
        self.assert_ok(self.run_cli("publish-wiki", pack_dir, "--adapter", "confluence"))

        result = self.run_cli(
            "execute-publish-plan",
            pack_dir / "exports" / "publish" / "wiki" / "confluence",
            "--runner",
            runner,
            "--format",
            "json",
        )
        self.assert_ok(result)

        receipt = json.loads(result.stdout)
        self.assertEqual("executed", receipt["status"])
        self.assertEqual("confluence", receipt["adapter"])
        self.assertTrue(Path(receipt["result_path"]).exists())
        bundle = self.read_json(Path(receipt["result_path"]))
        self.assertEqual(3, len(bundle["records"]))
        self.assertEqual("wiki-page", bundle["records"][0]["target_kind"])

    def test_publish_wiki_markdown_can_be_applied_locally(self) -> None:
        pack_dir = self.compile_research_pack()

        self.assert_ok(self.run_cli("publish-wiki", pack_dir, "--adapter", "markdown"))

        result = self.run_cli(
            "apply-publish-plan",
            pack_dir / "exports" / "publish" / "wiki" / "markdown",
            "--format",
            "json",
        )
        self.assert_ok(result)

        receipt = json.loads(result.stdout)
        self.assertEqual("applied-local", receipt["status"])
        self.assertTrue(Path(receipt["receipt_path"]).exists())
        self.assertTrue(receipt["applied_paths"])
        first_page = Path(receipt["applied_paths"][0])
        self.assertTrue(first_page.exists())
        self.assertTrue(first_page.read_text(encoding="utf-8").startswith("# "))

    def test_publish_wiki_markdown_can_apply_locally_in_one_command(self) -> None:
        pack_dir = self.compile_research_pack()

        result = self.run_cli(
            "publish-wiki",
            pack_dir,
            "--adapter",
            "markdown",
            "--apply-local",
            "--format",
            "json",
        )
        self.assert_ok(result)

        receipt = json.loads(result.stdout)
        self.assertEqual("applied-local", receipt["status"])
        self.assertTrue(receipt["applied_paths"])
        self.assertTrue(Path(receipt["receipt_path"]).exists())

    def test_recommend_execution_can_use_repo_context_for_guidance(self) -> None:
        pack_dir = self.compile_research_pack()
        repo = self.create_repo_fixture()

        result = self.run_cli("recommend-execution", pack_dir, "--repo", repo, "--format", "json")
        self.assert_ok(result)

        recommendation = json.loads(result.stdout)
        repo_context = recommendation["repo_context"]
        self.assertTrue(repo_context["discovered"])
        self.assertEqual(str(repo.resolve()), repo_context["repo_root"])
        self.assertEqual("repo-targeted", recommendation["context_policy"])
        self.assertEqual("repo entrypoints and worktree state", recommendation["tool_order"][0])
        self.assertTrue(any(entry["source"] == "package.json" for entry in repo_context["entrypoints"]["test"]))
        self.assertTrue(any(entry["source"] == "docker-compose.yml" for entry in repo_context["entrypoints"]["startup"]))
        self.assertTrue(any("startup surface" in item for item in repo_context["next_actions"]))

    def test_run_pack_report_includes_repo_context_when_requested(self) -> None:
        pack_dir = self.compile_research_pack()
        repo = self.create_repo_fixture()

        result = self.run_cli("run-pack", pack_dir, "--repo", repo, "--mode", "supervised", "--format", "json")
        self.assert_ok(result)

        report = json.loads(result.stdout)
        repo_context = report["recommendation"]["repo_context"]
        self.assertTrue(repo_context["discovered"])
        self.assertEqual(str(repo.resolve()), repo_context["repo_root"])
        self.assertTrue(any(target["artifact"] == "Evidence" for target in repo_context["verification_targets"]))

    def test_recommend_execution_surfaces_memory_context_for_resume(self) -> None:
        pack_dir = self.compile_research_pack()

        result = self.run_cli("recommend-execution", pack_dir, "--format", "json")
        self.assert_ok(result)

        recommendation = json.loads(result.stdout)
        memory_context = recommendation["memory_context"]
        artifact_types = [item["artifact_type"] for item in memory_context["recent_artifacts"]]
        self.assertIn("Spec", artifact_types)
        self.assertIn("Tasks", artifact_types)
        self.assertTrue(any(item["label"] == "knowledge" for item in memory_context["indexes"]))
        self.assertTrue(any("Task status:" in item for item in memory_context["resume_items"]))

    def test_verify_intent_surfaces_stale_memory_signals_when_evidence_is_missing(self) -> None:
        pack_dir = self.compile_research_pack()

        result = self.run_cli("recommend-execution", pack_dir, "--intent", "verify", "--format", "json")
        self.assert_ok(result)

        recommendation = json.loads(result.stdout)
        stale_signals = recommendation["memory_context"]["stale_signals"]
        self.assertTrue(any("unresolved task" in item for item in stale_signals))

    def test_example_research_pack_requires_approval_before_operational_transition(self) -> None:
        pack_dir = self.copy_research_example()

        blocked = self.run_cli("advance-pack", pack_dir)
        self.assert_error(blocked)
        self.assertIn("Operational transition requires Approval", blocked.stdout)

        approval = self.run_cli(
            "capture-approval",
            pack_dir,
            "--author",
            "release-coordinator",
            "--approver-actor",
            "eng-manager",
            "--scope",
            "operational-readiness",
        )
        self.assert_ok(approval)

        promoted = self.run_cli("advance-pack", pack_dir)
        self.assert_error(promoted)
        self.assertNotIn("Operational transition requires Approval", promoted.stdout)
        self.assertIn("Operational transition requires Runbook", promoted.stdout)
        self.assertIn("Operational transition requires Signals", promoted.stdout)
        self.assertIn("Operational transition requires Rollback", promoted.stdout)

    def test_harvest_evidence_captures_runtime_report_and_reviewed_status(self) -> None:
        pack_dir = self.compile_research_pack()
        repo = self.create_repo_fixture()

        result = self.run_cli(
            "harvest-evidence",
            pack_dir,
            "--author",
            "release-bot",
            "--repo",
            repo,
            "--category",
            "verify",
            "--category",
            "demo",
            "--category",
            "startup",
            "--max-targets",
            "3",
            "--format",
            "json",
        )
        self.assert_ok(result)

        report = json.loads(result.stdout)
        self.assertEqual("ready", report["readiness"])
        self.assertEqual("reviewed", report["evidence_status"])
        self.assertGreaterEqual(report["summary"]["passed"], 3)
        self.assertTrue(self.resolve_repo_path(report["report_path"]).exists())
        self.assertTrue(Path(report["evidence_artifact_path"]).exists())

        evidence = self.read_yaml(Path(report["evidence_artifact_path"]))
        self.assertEqual("reviewed", evidence["status"])
        self.assertIn(report["report_path"], evidence["references"])

    def test_harvest_evidence_downgrades_to_draft_when_verification_fails(self) -> None:
        pack_dir = self.compile_research_pack()
        repo = self.create_repo_fixture()
        (repo / "scripts" / "demo-check.sh").write_text("#!/bin/sh\nexit 1\n", encoding="utf-8")

        result = self.run_cli(
            "harvest-evidence",
            pack_dir,
            "--author",
            "release-bot",
            "--repo",
            repo,
            "--category",
            "demo",
            "--format",
            "json",
        )
        self.assert_ok(result)

        report = json.loads(result.stdout)
        self.assertEqual("bounded", report["readiness"])
        self.assertEqual("draft", report["evidence_status"])
        self.assertTrue(report["auto_downgraded"])
        self.assertGreaterEqual(report["summary"]["failed"], 1)
        self.assertTrue(any("downgraded to draft" in risk for risk in report["residual_risks"]))

        evidence = self.read_yaml(Path(report["evidence_artifact_path"]))
        self.assertEqual("draft", evidence["status"])

    def test_harvest_evidence_does_not_inject_script_filename(self) -> None:
        pack_dir = self.compile_research_pack()
        repo = self.create_malicious_verification_repo()

        result = self.run_cli(
            "harvest-evidence",
            pack_dir,
            "--author",
            "release-bot",
            "--repo",
            repo,
            "--category",
            "verify",
            "--max-targets",
            "1",
            "--format",
            "json",
        )
        self.assert_ok(result)

        report = json.loads(result.stdout)
        self.assertEqual("ready", report["readiness"])
        self.assertEqual("reviewed", report["evidence_status"])
        self.assertGreaterEqual(report["summary"]["passed"], 1)
        self.assertFalse((repo / "PWN.sh").exists())

    def test_next_surfaces_repo_targets_and_harvest_step(self) -> None:
        pack_dir = self.compile_research_pack()
        repo = self.create_repo_fixture()

        result = self.run_cli(
            "next",
            pack_dir,
            "--repo",
            repo,
            "--intent",
            "verify",
            "--format",
            "json",
        )
        self.assert_ok(result)

        checklist = json.loads(result.stdout)
        self.assertEqual("verify", checklist["intent"])
        self.assertEqual(str(repo.resolve()), checklist["repo_root"])
        self.assertTrue(any(item["kind"] == "repo-target" for item in checklist["items"]))
        self.assertTrue(
            any(
                item["kind"] == "command" and "harvest-evidence" in item.get("command", "")
                for item in checklist["items"]
            )
        )

    def test_resume_surfaces_latest_harvest_and_unresolved_tasks(self) -> None:
        pack_dir = self.compile_research_pack()
        repo = self.create_repo_fixture()

        self.assert_ok(
            self.run_cli(
                "harvest-evidence",
                pack_dir,
                "--author",
                "release-bot",
                "--repo",
                repo,
                "--category",
                "verify",
                "--format",
                "json",
            )
        )

        result = self.run_cli(
            "resume",
            pack_dir,
            "--repo",
            repo,
            "--intent",
            "verify",
            "--format",
            "json",
        )
        self.assert_ok(result)

        compact = json.loads(result.stdout)
        self.assertEqual("verify", compact["intent"])
        self.assertTrue(compact["unresolved_tasks"])
        self.assertEqual("ready", compact["latest_harvest"]["readiness"])
        self.assertTrue(any("State anchor:" in item for item in compact["resume_items"]))

    def test_resume_applies_token_budget_and_trim(self) -> None:
        pack_dir = self.compile_research_pack()
        repo = self.create_repo_fixture()

        result = self.run_cli(
            "resume",
            pack_dir,
            "--repo",
            repo,
            "--intent",
            "verify",
            "--max-items",
            "2",
            "--max-chars",
            "600",
            "--format",
            "json",
        )
        self.assert_ok(result)

        compact = json.loads(result.stdout)
        self.assertLessEqual(compact["token_budget"]["estimated_chars"], 600)
        self.assertIn("estimated_tokens", compact["token_budget"])
        self.assertLessEqual(len(compact["resume_items"]), 3)

    def test_catalog_packs_includes_budget_cycle(self) -> None:
        result = self.run_cli("catalog-packs", "--format", "json")
        self.assert_ok(result)

        catalog = json.loads(result.stdout)
        pack_ids = [item["pack_id"] for item in catalog["packs"]]
        self.assertIn("meeting-followup", pack_ids)
        self.assertIn("budget-cycle", pack_ids)
        self.assertIn("compliance-audit", pack_ids)
        self.assertIn("vendor-selection", pack_ids)
        meeting_entry = next(item for item in catalog["packs"] if item["pack_id"] == "meeting-followup")
        self.assertEqual("Delivery", meeting_entry["target_profile"])
        self.assertEqual("compile-pack", meeting_entry["bootstrap_mode"])
        budget_entry = next(item for item in catalog["packs"] if item["pack_id"] == "budget-cycle")
        self.assertEqual("Explore", budget_entry["target_profile"])
        self.assertEqual("compile-pack", budget_entry["bootstrap_mode"])
        compliance_entry = next(item for item in catalog["packs"] if item["pack_id"] == "compliance-audit")
        self.assertEqual("Regulated", compliance_entry["target_profile"])
        self.assertEqual("compile-pack", compliance_entry["bootstrap_mode"])
        vendor_entry = next(item for item in catalog["packs"] if item["pack_id"] == "vendor-selection")
        self.assertEqual("Delivery", vendor_entry["target_profile"])
        self.assertEqual("compile-pack", vendor_entry["bootstrap_mode"])

    def test_recommend_execution_detects_rust_and_just_repo_targets(self) -> None:
        pack_dir = self.compile_research_pack()
        repo = self.create_rust_repo_fixture()

        result = self.run_cli("recommend-execution", pack_dir, "--repo", repo, "--format", "json")
        self.assert_ok(result)

        recommendation = json.loads(result.stdout)
        next_actions = recommendation["repo_context"]["next_actions"]
        verification_targets = recommendation["repo_context"]["verification_targets"]
        self.assertTrue(any("cargo test" in item for item in next_actions))
        self.assertTrue(any(target.get("command") == "just verify" for target in verification_targets))

    def test_repo_map_reports_topology_stack_and_steering(self) -> None:
        repo = self.create_repo_fixture()
        self.assert_ok(self.run_cli("bootstrap-steering", repo))

        result = self.run_cli("repo-map", repo, "--format", "json")
        self.assert_ok(result)

        repo_map = json.loads(result.stdout)
        self.assertIn("node", repo_map["detected_stack"])
        self.assertIn("docker-compose", repo_map["detected_stack"])
        self.assertTrue(repo_map["steering"]["found"])
        self.assertIn("Makefile", repo_map["top_level_files"])
        self.assertTrue(any("make verify" in item for item in repo_map["entrypoints"]["verify"]))

    def test_show_adapters_filters_by_capability(self) -> None:
        result = self.run_cli("show-adapters", "--capability", "issues-publish-plan", "--format", "json")
        self.assert_ok(result)

        summary = json.loads(result.stdout)
        ids = [item["id"] for item in summary["adapters"]]
        self.assertEqual({"jira", "github"}, set(ids))

    def test_adapter_conformance_reports_ok_for_current_registry(self) -> None:
        result = self.run_cli("adapter-conformance", "--format", "json")
        self.assert_ok(result)

        summary = json.loads(result.stdout)
        self.assertEqual("ok", summary["status"])
        self.assertTrue(any(item["id"] == "codex" for item in summary["checks"]))
        self.assertTrue(all(item["status"] == "ok" for item in summary["checks"]))

    def test_resolve_adapter_prefers_jira_and_lists_fallbacks(self) -> None:
        result = self.run_cli(
            "resolve-adapter",
            "--capability",
            "issues-export",
            "--layer",
            "issue-system",
            "--format",
            "json",
        )
        self.assert_ok(result)

        resolution = json.loads(result.stdout)
        self.assertEqual("jira", resolution["selected"]["id"])
        self.assertIn("github", resolution["fallbacks"])

    def test_adapter_matrix_reports_runtime_bundle_install_coverage(self) -> None:
        result = self.run_cli("adapter-matrix", "--format", "json")
        self.assert_ok(result)

        matrix = json.loads(result.stdout)
        self.assertIn("runtime-target", matrix["layers"])
        self.assertIn("bundle-install", matrix["capabilities"])
        self.assertIn("codex", matrix["capabilities"]["bundle-install"])

    def test_show_learning_events_reads_pack_local_runtime_log(self) -> None:
        pack_dir = self.compile_research_pack()
        repo = self.create_repo_fixture()

        self.assert_ok(self.run_cli("resume", pack_dir, "--repo", repo))
        self.assert_ok(self.run_cli("next", pack_dir, "--repo", repo, "--intent", "verify"))
        self.assert_ok(
            self.run_cli(
                "harvest-evidence",
                pack_dir,
                "--author",
                "release-bot",
                "--repo",
                repo,
                "--category",
                "verify",
            )
        )

        result = self.run_cli(
            "show-learning-events",
            pack_dir,
            "--format",
            "json",
        )
        self.assert_ok(result)

        payload = json.loads(result.stdout)
        event_types = [item["event_type"] for item in payload["events"]]
        self.assertIn("compact-context", event_types)
        self.assertIn("execution-checklist", event_types)
        self.assertIn("harvest-evidence", event_types)

    def test_learning_snapshot_summarizes_recent_runtime_events(self) -> None:
        pack_dir = self.compile_research_pack()
        repo = self.create_repo_fixture()

        self.assert_ok(self.run_cli("resume", pack_dir, "--repo", repo, "--intent", "verify"))
        self.assert_ok(
            self.run_cli(
                "harvest-evidence",
                pack_dir,
                "--author",
                "release-bot",
                "--repo",
                repo,
                "--category",
                "verify",
            )
        )

        result = self.run_cli("learning-snapshot", pack_dir, "--format", "json")
        self.assert_ok(result)

        snapshot = json.loads(result.stdout)
        self.assertGreaterEqual(snapshot["event_count"], 2)
        self.assertIn("compact-context", snapshot["event_types"])
        self.assertIn("harvest-evidence", snapshot["event_types"])
        self.assertIn("ready", snapshot["harvest_readiness"])

    def test_routing_backtest_recommends_deep_for_verify_after_harvest(self) -> None:
        pack_dir = self.compile_research_pack()
        repo = self.create_repo_fixture()

        self.assert_ok(self.run_cli("resume", pack_dir, "--repo", repo, "--intent", "verify"))
        self.assert_ok(self.run_cli("next", pack_dir, "--repo", repo, "--intent", "verify"))
        self.assert_ok(
            self.run_cli(
                "harvest-evidence",
                pack_dir,
                "--author",
                "release-bot",
                "--repo",
                repo,
                "--category",
                "verify",
            )
        )

        result = self.run_cli("routing-backtest", pack_dir, "--format", "json")
        self.assert_ok(result)

        backtest = json.loads(result.stdout)
        verify_policy = next(item for item in backtest["policy_recommendations"] if item["intent"] == "verify")
        self.assertEqual("deep", verify_policy["recommended_execution_class"])
        self.assertTrue(any(bucket["execution_class"] == "deep" for bucket in backtest["buckets"]))

    def test_stage_runtime_handoff_materializes_target_ready_bundle(self) -> None:
        pack_dir = self.compile_research_pack()
        repo = self.create_repo_fixture()
        self.assert_ok(self.run_cli("bootstrap-steering", repo))
        home = self.personal_home()
        self.assert_ok(self.run_cli("bootstrap-home", home))
        self.assert_ok(
            self.run_cli(
                "append-memory",
                home,
                "--line",
                "Prefers runtime-ready handoff bundles with compact resume context.",
            )
        )
        self.assert_ok(self.run_cli("dream-memory", home))
        self.assert_ok(self.run_cli("bind-home", pack_dir, "--home", home))

        result = self.run_cli(
            "stage-runtime-handoff",
            pack_dir,
            "--repo",
            repo,
            "--runtime-target",
            "kiro-cli",
            "--format",
            "json",
        )
        self.assert_ok(result)

        handoff = json.loads(result.stdout)
        self.assertEqual("kiro-cli", handoff["runtime_target"]["selected"]["id"])
        self.assertTrue(handoff["home_binding"]["bound"])
        self.assertTrue(any("runtime-ready handoff bundles" in item for item in handoff["compact_context"]["home_memory"]))
        self.assertEqual("kiro-cli", handoff["compact_context"]["runtime_target"]["selected"])
        self.assertTrue(handoff["repo_map"]["steering"]["found"])
        self.assertTrue(handoff["install_plan"]["targets"])
        self.assertEqual("kiro-cli", handoff["install_plan"]["targets"][0]["id"])
        handoff_path = Path(handoff["handoff_path"])
        self.assertTrue(handoff_path.exists())
        saved = self.read_json(handoff_path)
        self.assertEqual("JiniRuntimeHandoff", saved["handoff_type"])
        self.assertEqual("kiro-cli", saved["runtime_target"]["selected"]["id"])
        events = self.run_cli("show-learning-events", pack_dir, "--format", "json")
        self.assert_ok(events)
        event_types = [item["event_type"] for item in json.loads(events.stdout)["events"]]
        self.assertIn("stage-runtime-handoff", event_types)

    def test_activate_runtime_target_materializes_local_runtime_bundle(self) -> None:
        pack_dir = self.compile_research_pack()
        repo = self.create_repo_fixture()
        self.assert_ok(self.run_cli("bootstrap-steering", repo))
        home = self.personal_home()
        prefix = self.install_prefix()
        self.assert_ok(self.run_cli("bootstrap-home", home))
        self.assert_ok(self.run_cli("bind-home", pack_dir, "--home", home))

        result = self.run_cli(
            "activate-runtime-target",
            pack_dir,
            "--repo",
            repo,
            "--runtime-target",
            "codex",
            "--prefix",
            prefix,
            "--format",
            "json",
        )
        self.assert_ok(result)

        activation = json.loads(result.stdout)
        self.assertEqual("codex", activation["runtime_target"])
        self.assertTrue(Path(activation["activation_root"]).exists())
        self.assertTrue(Path(activation["install_receipt_path"]).exists())
        self.assertTrue(any(path.endswith("Jini-RUNTIME.md") for path in activation["activation_files"]))
        self.assertTrue(activation["home_observation"]["appended"])
        handoff_copy = Path(activation["activation_root"]) / "handoff.json"
        self.assertTrue(handoff_copy.exists())
        saved_handoff = self.read_json(handoff_copy)
        self.assertEqual("JiniRuntimeHandoff", saved_handoff["handoff_type"])
        events = self.run_cli("show-learning-events", pack_dir, "--format", "json")
        self.assert_ok(events)
        event_types = [item["event_type"] for item in json.loads(events.stdout)["events"]]
        self.assertIn("activate-runtime-target", event_types)

    def test_run_collapses_activation_run_and_local_publish(self) -> None:
        pack_dir = self.compile_research_pack()
        repo = self.create_repo_fixture()
        home = self.personal_home()
        prefix = self.install_prefix()
        self.assert_ok(self.run_cli("bootstrap-home", home))

        result = self.run_cli(
            "run",
            pack_dir,
            "--mode",
            "supervised",
            "--intent",
            "publish",
            "--repo",
            repo,
            "--home",
            home,
            "--runtime-target",
            "codex",
            "--activate-runtime",
            "--prefix",
            prefix,
            "--consent",
            "write",
            "--consent",
            "publish",
            "--issue-adapter",
            "github",
            "--wiki-adapter",
            "markdown",
            "--format",
            "json",
        )
        self.assert_ok(result)

        report = json.loads(result.stdout)
        self.assertTrue(Path(report["flow_path"]).exists())
        self.assertEqual("publish", report["intent"])
        self.assertEqual("codex", report["recommendation"]["runtime_guidance"]["selected"]["id"])
        self.assertTrue(report["runtime_activation"])
        self.assertTrue(Path(report["runtime_activation_path"]).exists())
        self.assertTrue(self.resolve_repo_path(report["run_report_path"]).exists())
        self.assertTrue(report["local_publish_receipts"])
        self.assertTrue(any(item["status"] == "applied-local" for item in report["local_publish_receipts"]))
        self.assertGreater(report["token_strategy"]["compact_estimated_tokens"], 0)
        self.assertIn("compact-context", report["token_strategy"]["reused_context_surfaces"])
        self.assertIn("runtime-handoff", report["token_strategy"]["reused_context_surfaces"])

    def test_review_policy_summarizes_learning_without_mutating_policy(self) -> None:
        pack_dir = self.compile_research_pack()
        repo = self.create_repo_fixture()
        home = self.personal_home()
        self.assert_ok(self.run_cli("bootstrap-home", home))
        self.assert_ok(self.run_cli("bind-home", pack_dir, "--home", home))
        self.assert_ok(
            self.run_cli(
                "stage-runtime-handoff",
                pack_dir,
                "--repo",
                repo,
                "--runtime-target",
                "codex",
            )
        )
        self.assert_ok(self.run_cli("resume", pack_dir, "--repo", repo, "--intent", "verify"))
        self.assert_ok(self.run_cli("next", pack_dir, "--repo", repo, "--intent", "verify"))
        self.assert_ok(
            self.run_cli(
                "harvest-evidence",
                pack_dir,
                "--author",
                "release-bot",
                "--repo",
                repo,
                "--category",
                "verify",
            )
        )
        self.assert_ok(
            self.run_cli(
                "run-pack",
                pack_dir,
                "--mode",
                "supervised",
                "--repo",
                repo,
                "--runtime-target",
                "codex",
            )
        )

        result = self.run_cli("review-policy", pack_dir, "--format", "json")
        self.assert_ok(result)

        review = json.loads(result.stdout)
        self.assertFalse(review["guardrails"]["mutation_allowed"])
        self.assertTrue(review["report_path"])
        self.assertTrue(self.resolve_repo_path(review["report_path"]).exists())
        self.assertGreaterEqual(review["learning_snapshot"]["event_count"], 4)
        self.assertIn("codex", review["runtime_targets"])
        self.assertTrue(review["policy_candidates"])
        self.assertTrue(any(item["kind"] == "routing-default" for item in review["policy_candidates"]))
        self.assertTrue(any(item["kind"] == "promotion-gate" for item in review["policy_candidates"]))
        events = self.run_cli("show-learning-events", pack_dir, "--format", "json")
        self.assert_ok(events)
        event_types = [item["event_type"] for item in json.loads(events.stdout)["events"]]
        self.assertIn("policy-review", event_types)

    def test_policy_candidate_lifecycle_can_activate_and_rollback_routing_override(self) -> None:
        pack_dir = self.compile_research_pack()
        runtime_dir = pack_dir / "runtime"
        runtime_dir.mkdir(parents=True, exist_ok=True)
        events_path = runtime_dir / "events.jsonl"
        events = [
            {
                "schema_version": "0.1.0",
                "event_type": "run-pack",
                "recorded_at": "2026-05-11T10:00:00Z",
                "pack_id": "research-prd",
                "work_unit_id": "test-research-pack",
                "intent": "make",
                "execution_class": "cheap",
                "state_before": "decided",
                "state_after": "in_make",
                "blocker_count": 0,
            },
            {
                "schema_version": "0.1.0",
                "event_type": "stage-runtime-handoff",
                "recorded_at": "2026-05-11T10:01:00Z",
                "pack_id": "research-prd",
                "work_unit_id": "test-research-pack",
                "intent": "make",
                "execution_class": "cheap",
                "runtime_target": "kiro-cli",
                "estimated_tokens": 140,
                "compression_ratio": 0.42,
            },
        ]
        events_path.write_text("".join(json.dumps(item, sort_keys=True) + "\n" for item in events), encoding="utf-8")

        self.assert_ok(self.run_cli("review-policy", pack_dir))
        stage = self.run_cli("stage-policy-candidate", pack_dir, "--format", "json")
        self.assert_ok(stage)
        candidate = json.loads(stage.stdout)
        candidate_path = Path(pack_dir) / "runtime" / "policy-candidates" / f"{candidate['candidate_id']}.json"
        self.assertTrue(candidate_path.exists())
        self.assertEqual("cheap", candidate["intent_overrides"]["make"])

        approve = self.run_cli(
            "approve-policy-candidate",
            pack_dir,
            candidate_path,
            "--approver",
            "policy-lead",
            "--format",
            "json",
        )
        self.assert_ok(approve)
        rollout = json.loads(approve.stdout)
        self.assertEqual("active", rollout["status"])
        self.assertEqual("cheap", rollout["intent_overrides"]["make"])

        recommendation = self.run_cli("recommend-execution", pack_dir, "--intent", "make", "--format", "json")
        self.assert_ok(recommendation)
        recommendation_doc = json.loads(recommendation.stdout)
        self.assertEqual("cheap", recommendation_doc["execution_class"])
        self.assertEqual(candidate["candidate_id"], recommendation_doc["active_policy"]["candidate_id"])
        self.assertEqual("kiro-cli", recommendation_doc["runtime_guidance"]["selected"]["id"])

        rollback = self.run_cli(
            "rollback-policy-candidate",
            pack_dir,
            candidate_path,
            "--actor",
            "policy-lead",
            "--reason",
            "Restore baseline routing after evaluation",
            "--format",
            "json",
        )
        self.assert_ok(rollback)
        rollback_doc = json.loads(rollback.stdout)
        self.assertEqual("rolled-back", rollback_doc["status"])

        restored = self.run_cli("recommend-execution", pack_dir, "--intent", "make", "--format", "json")
        self.assert_ok(restored)
        restored_doc = json.loads(restored.stdout)
        self.assertEqual("standard", restored_doc["execution_class"])
        self.assertEqual({}, restored_doc["active_policy"])

    def test_show_kpis_json_reports_ranked_dimensions(self) -> None:
        result = self.run_cli("show-kpis", "--format", "json", "--limit", "3")
        self.assert_ok(result)

        summary = json.loads(result.stdout)
        self.assertEqual("2026-05-27", summary["updated_at"])
        self.assertEqual(3, len(summary["dimensions"]))
        self.assertEqual("workflow-rigor", summary["dimensions"][0]["id"])
        self.assertEqual("Kiro", summary["dimensions"][0]["strongest_competitor"]["name"])

    def test_publish_readiness_reports_publishable_surface(self) -> None:
        result = self.run_cli("publish-readiness", "--format", "json")
        self.assert_ok(result)

        report = json.loads(result.stdout)
        self.assertEqual("JiniPublishReadiness", report["result_type"])
        self.assertEqual("warning", report["status"])
        self.assertEqual("starter-kit", report["default_kit_id"])
        self.assertGreaterEqual(report["pack_count"], 6)
        self.assertGreaterEqual(report["kit_count"], 6)
        self.assertTrue(any(section["id"] == "install" and section["status"] == "ok" for section in report["sections"]))
        novice = next(section for section in report["sections"] if section["id"] == "novice")
        self.assertEqual("ok", novice["status"])
        self.assertTrue(any(item.get("id") == "beginner-command-count" and item["command_count"] <= 4 for item in novice["checks"]))
        self.assertTrue(any(item.get("id") == "beginner-command-prefix" and item["all_jini_commands"] for item in novice["checks"]))
        self.assertTrue(any(item.get("id") == "beginner-single-shell-command" and item["present"] for item in novice["checks"]))
        self.assertTrue(any(item.get("id") == "cli-guide-no-python-requirement" and item["present"] for item in novice["checks"]))
        self.assertTrue(any(item.get("id") == "readme-plain-words-entry" and item["present"] for item in novice["checks"]))
        self.assertTrue(any(item.get("id") == "simple-guide-exists" and item["present"] for item in novice["checks"]))
        self.assertTrue(any(item.get("id") == "homepage-plain-words-entry" and item["present"] and item["status"] == "ok" for item in novice["checks"]))
        self.assertTrue(any(item.get("id") == "simple-guide-core-questions" and item["present"] and item["status"] == "ok" for item in novice["checks"]))
        self.assertTrue(any(section["id"] == "breadth" and section["status"] == "ok" for section in report["sections"]))
        leadership = next(section for section in report["sections"] if section["id"] == "leadership")
        self.assertEqual("ok", leadership["status"])
        guarded_dimensions = {item["dimension_id"] for item in leadership["checks"]}
        self.assertIn("workflow-rigor", guarded_dimensions)
        self.assertIn("governance", guarded_dimensions)
        self.assertTrue(all(item["status"] == "ok" for item in leadership["checks"]))
        self.assertTrue(
            any(item["dimension_id"] == "learning-maturity" and item["position"] == "ahead" for item in leadership["checks"])
        )
        rewrite_momentum = next(section for section in report["sections"] if section["id"] == "rewrite-momentum")
        self.assertEqual("warning", rewrite_momentum["status"])
        self.assertTrue(any(item["id"] == "overall-score-floor" and item["delta"] > 0 for item in rewrite_momentum["checks"]))
        self.assertTrue(
            any(
                item["id"] == "overall-lead-margin" and item["margin"] < item["minimum_margin"] and item["status"] == "warning"
                for item in rewrite_momentum["checks"]
            )
        )
        consensus_gates = report["consensus_gates"]
        self.assertEqual("ok", consensus_gates["status"])
        expected_gate_ids = {
            "product-review-role-docs",
            "cross-critic-convergence",
            "two-flagship-replacement-flows",
            "useful-result-first",
            "i-am-not-sure-fallback",
            "real-continuation-actions",
            "parity-or-shared-generation",
        }
        self.assertEqual(expected_gate_ids, set(consensus_gates["check_ids"]))
        self.assertEqual(expected_gate_ids, set(consensus_gates["passed_gate_ids"]))
        self.assertTrue(any(section["id"] == "consensus-gates" and section["status"] == "ok" for section in report["sections"]))

    def test_provider_doctor_reports_azure_openai_safely(self) -> None:
        env = os.environ.copy()
        env.update(
            {
                "JINI_PROVIDER": "azure-openai",
                "AZURE_OPENAI_ENDPOINT": "https://example.openai.azure.com",
                "AZURE_OPENAI_API_KEY": "super-secret-key",
                "AZURE_OPENAI_DEPLOYMENT": "gpt-4o-prod",
                "AZURE_OPENAI_API_VERSION": "2024-10-21",
            }
        )
        result = self.run_cli("doctor", "--format", "json", env=env)
        self.assert_ok(result)

        report = json.loads(result.stdout)
        self.assertEqual("JiniProviderDoctor", report["result_type"])
        self.assertEqual("azure-openai", report["provider_id"])
        self.assertEqual("ok", report["status"])
        rendered = json.dumps(report)
        self.assertIn("AZURE_OPENAI_API_KEY", rendered)
        self.assertNotIn("super-secret-key", rendered)

    def test_provider_doctor_reports_bedrock_safely(self) -> None:
        env = os.environ.copy()
        env.update(
            {
                "JINI_PROVIDER": "bedrock",
                "AWS_REGION": "us-east-1",
                "AWS_PROFILE": "work-profile",
                "BEDROCK_MODEL_ID": "anthropic.claude-3-5-sonnet-20240620-v1:0",
            }
        )
        result = self.run_cli("doctor", "--format", "json", env=env)
        self.assert_ok(result)

        report = json.loads(result.stdout)
        self.assertEqual("bedrock", report["provider_id"])
        self.assertEqual("ok", report["status"])
        rendered = json.dumps(report)
        self.assertIn("AWS_PROFILE", rendered)
        self.assertNotIn("work-profile", rendered)
        self.assertNotIn("anthropic.claude-3-5-sonnet-20240620-v1:0", rendered)

    def test_provider_doctor_reports_claude_direct_safely(self) -> None:
        env = os.environ.copy()
        env.update(
            {
                "JINI_PROVIDER": "claude",
                "ANTHROPIC_API_KEY": "super-secret-key",
                "JINI_MODEL": "sonnet",
            }
        )
        result = self.run_cli("doctor", "--format", "json", env=env)
        self.assert_ok(result)

        report = json.loads(result.stdout)
        self.assertEqual("anthropic", report["provider_id"])
        self.assertEqual("ok", report["status"])
        self.assertIn("Claude API", report["label"])
        rendered = json.dumps(report)
        self.assertIn("ANTHROPIC_API_KEY", rendered)
        self.assertNotIn("super-secret-key", rendered)

    def test_provider_doctor_auto_prefers_bedrock_for_sonnet_46(self) -> None:
        env = os.environ.copy()
        env.update(
            {
                "JINI_PROVIDER": "auto",
                "JINI_MODEL": "sonnet-4.6",
                "AWS_REGION": "us-east-1",
                "AWS_PROFILE": "work-profile",
            }
        )
        result = self.run_cli("doctor", "--format", "json", env=env)
        self.assert_ok(result)

        report = json.loads(result.stdout)
        self.assertEqual("bedrock", report["provider_id"])
        self.assertEqual("ok", report["status"])
        provider_setting = next(item for item in report["settings"] if item["name"] == "JINI_PROVIDER")
        self.assertIn("auto -> Amazon Bedrock", provider_setting["presence"])
        model_setting = next(item for item in report["settings"] if item["name"] == "JINI_MODEL")
        self.assertEqual("sonnet-4.6 -> Claude Sonnet 4.6", model_setting["presence"])
        rendered = json.dumps(report)
        self.assertNotIn("work-profile", rendered)
        self.assertNotIn("anthropic.claude-sonnet-4-6", rendered)

    def test_provider_doctor_rejects_sonnet_46_shortcut_for_direct_claude(self) -> None:
        env = os.environ.copy()
        env.update(
            {
                "JINI_PROVIDER": "claude",
                "JINI_MODEL": "sonnet-4.6",
                "ANTHROPIC_API_KEY": "super-secret-key",
            }
        )
        result = self.run_cli("doctor", "--format", "json", env=env)
        self.assertNotEqual(0, result.returncode)

        report = json.loads(result.stdout)
        self.assertEqual("anthropic", report["provider_id"])
        self.assertEqual("needs setup", report["status"])
        self.assertTrue(any("supported only on Bedrock" in item for item in report["missing"]))

    def test_provider_nested_command_keeps_doctor_compatibility_alias(self) -> None:
        env = {
            "JINI_PROVIDER": "claude",
            "ANTHROPIC_API_KEY": "sk-live-secret",
            "JINI_MODEL": "sonnet",
        }
        result = self.run_cli("provider", "doctor", "--format", "json", env=env)
        self.assert_ok(result)
        report = json.loads(result.stdout)
        self.assertEqual("anthropic", report["provider_id"])

    def test_validate_golden_benchmark_reports_jini_against_expanded_competitor_field(self) -> None:
        result = self.run_cli("validate-golden-benchmark", "--format", "json")
        self.assert_ok(result)

        report = json.loads(result.stdout)
        self.assertEqual("JiniGoldenBenchmarkValidation", report["report_type"])
        self.assertTrue(self.resolve_repo_path(report["report_path"]).exists())
        self.assertEqual("2026-05-20", report["last_verified_at"])
        self.assertTrue(report["dataset_digest"])
        competitor_ids = {item["id"] for item in report["competitors"]}
        self.assertEqual({"claude-code", "codex", "chatgpt", "kiro", "hermes", "agentfield", "ai-hero"}, competitor_ids)
        self.assertTrue(all(item.get("source_urls") for item in report["competitors"]))
        self.assertEqual(8, report["scenario_count"])
        self.assertEqual("leading", report["overall"]["status"])
        self.assertIn("tracks", report["overall"])
        self.assertEqual("leading", report["overall"]["tracks"]["architecture-strength"]["status"])
        self.assertEqual("trailing", report["overall"]["tracks"]["adoption-truth"]["status"])
        for competitor_id, competitor_score in report["overall"]["competitor_scores"].items():
            self.assertGreater(report["overall"]["jini_score"], competitor_score, competitor_id)
        scenario_ids = {item["id"] for item in report["scenarios"]}
        self.assertIn("install-trust", scenario_ids)
        self.assertIn("portable-edges", scenario_ids)
        self.assertIn("guided-product-loop", scenario_ids)
        self.assertIn("first-minute-adoption", scenario_ids)
        self.assertIn("operational-breadth", scenario_ids)
        self.assertIn("product-consensus-gates", scenario_ids)
        install_scenario = next(item for item in report["scenarios"] if item["id"] == "install-trust")
        self.assertTrue(
            any(
                check["id"] == "doctor-install-runs-resolve-adapter-probe" and check["status"] == "ok"
                for check in install_scenario["checks"]
            )
        )
        guided_loop = next(item for item in report["scenarios"] if item["id"] == "guided-product-loop")
        self.assertTrue(
            any(
                check["id"] == "guided-loop-applies-local-publish" and check["status"] == "ok"
                for check in guided_loop["checks"]
            )
        )
        self.assertTrue(
            any(
                check["id"] == "guided-loop-reuses-handoff-context" and check["status"] == "ok"
                for check in guided_loop["checks"]
            )
        )
        adoption_scenario = next(item for item in report["scenarios"] if item["id"] == "first-minute-adoption")
        self.assertEqual(["adoption-truth"], adoption_scenario["tracks"])
        self.assertTrue(
            any(
                check["id"] == "beginner-path-is-single-command" and check["status"] == "ok"
                for check in adoption_scenario["checks"]
            )
        )
        breadth_scenario = next(item for item in report["scenarios"] if item["id"] == "operational-breadth")
        self.assertTrue(
            any(
                check["id"] == "golden-benchmark-routine-executes" and check["status"] == "ok"
                for check in breadth_scenario["checks"]
            )
        )
        self.assertTrue(
            any(
                check["id"] == "framework-review-routine-executes" and check["status"] == "ok"
                for check in breadth_scenario["checks"]
            )
        )
        consensus_scenario = next(item for item in report["scenarios"] if item["id"] == "product-consensus-gates")
        expected_consensus_checks = {
            "product-review-role-docs-gate",
            "cross-critic-convergence-gate",
            "two-flagship-replacement-flows-gate",
            "useful-result-first-gate",
            "i-am-not-sure-fallback-gate",
            "real-continuation-actions-gate",
            "parity-or-shared-generation-gate",
        }
        self.assertEqual(
            expected_consensus_checks,
            {check["id"] for check in consensus_scenario["checks"] if check["status"] == "ok"},
        )

    def test_get_started_reports_beginner_and_power_paths(self) -> None:
        result = self.run_cli("get-started", "--harness", "codex", "--format", "json")
        self.assert_ok(result)

        guide = json.loads(result.stdout)
        self.assertEqual("JiniGettingStartedGuide", guide["guide_type"])
        self.assertEqual("codex", guide["target"])
        self.assertEqual("starter-kit", guide["beginner_path"]["kit_id"])
        self.assertEqual("benchmark-delivery-kit", guide["power_user_path"]["kit_id"])
        self.assertEqual("preview", guide["beginner_path"]["trust_path"][0]["id"])
        self.assertIn("Bundle-level detail is intentionally hidden", guide["beginner_path"]["notes"][0])
        self.assertEqual(["jini"], guide["beginner_path"]["commands"])
        self.assertFalse(any("catalog-bundles" in item for item in guide["beginner_path"]["commands"]))
        self.assertFalse(any("show-kpis" in item for item in guide["beginner_path"]["commands"]))
        self.assertTrue(any("harnesses" in item for item in guide["power_user_path"]["commands"]))
        self.assertTrue(any(item == "jini status" for item in guide["power_user_path"]["commands"]))
        self.assertTrue(any(item == "jini open" for item in guide["power_user_path"]["commands"]))
        self.assertTrue(guide["shared_model"])

    def test_try_example_reports_bundled_research_prd_value(self) -> None:
        result = self.run_cli("try-example", "research-prd", "--format", "json")
        self.assert_ok(result)

        report = json.loads(result.stdout)
        self.assertEqual("JiniPublicExampleProof", report["example_type"])
        self.assertEqual("research-prd", report["example_id"])
        self.assertFalse(report["generated"])
        self.assertEqual("awaiting_verification", report["state"])
        self.assertEqual("ready-to-verify", report["health"])
        self.assertIn("Approval", report["missing_later"])
        self.assertEqual(3, report["task_summary"]["done"])
        self.assertEqual(0, report["task_summary"]["unresolved"])
        self.assertTrue(report["evidence_summary"]["present"])
        self.assertTrue(any("verification becomes a visible stage" in item.lower() for item in report["daily_value"]))
        self.assertTrue(any(item == "jini continue" for item in report["continue_with"]))
        self.assertTrue(any(item == "Inside Jini: Continue" for item in report["continue_with"]))
        self.assertTrue(any(item == "Inside Jini: Open" for item in report["continue_with"]))

    def test_try_example_generates_meeting_followup_into_requested_output(self) -> None:
        output = self.tmp / "example-meeting-followup"
        result = self.run_cli("try-example", "meeting-followup", "--output", output, "--format", "json")
        self.assert_ok(result)

        report = json.loads(result.stdout)
        self.assertEqual("meeting-followup", report["example_id"])
        self.assertTrue(report["generated"])
        self.assertEqual(str(output), report["path"])
        self.assertTrue(output.exists())
        self.assertEqual("decided", report["state"])
        self.assertEqual("ready-to-make", report["health"])
        self.assertIn("Approval", report["missing_later"])
        self.assertIn("Evidence", report["missing_later"])
        self.assertEqual(0, report["task_summary"]["done"])
        self.assertEqual(3, report["task_summary"]["unresolved"])
        self.assertTrue(any("action items stop living only in chat threads" in item.lower() for item in report["daily_value"]))
        followup = (output / "views" / "followup.md").read_text(encoding="utf-8")
        self.assertIn("# Sendable Follow-Up:", followup)
        self.assertIn("## Send this", followup)
        self.assertIn("## Owners and due points", followup)

    def test_meeting_example_sets_current_work_for_pathless_status_and_open(self) -> None:
        output = self.tmp / "example-meeting-followup"
        result = self.run_cli("try-example", "meeting-followup", "--output", output, "--format", "json")
        self.assert_ok(result)
        followup = (output / "views" / "followup.md").read_text(encoding="utf-8")

        status = self.run_cli("status")
        self.assert_ok(status)
        self.assertIn("WORK   example-meeting-followup", status.stdout)
        self.assertIn("READY NOW", status.stdout)
        self.assertIn("followup", status.stdout)

        open_result = self.run_cli("open", "followup", "--print-path")
        self.assert_ok(open_result)
        self.assertIn(str((output / "views" / "followup.md").resolve()), open_result.stdout.strip())
        self.assertIn("Pricing draft moves to Sarah by Thursday", followup)
        self.assertIn("Priya: confirm the launch metric decision by Friday", followup)

    def test_review_framework_persists_prioritized_adoption_report(self) -> None:
        result = self.run_cli("review-framework", "--format", "json", "--limit", "3")
        self.assert_ok(result)

        review = json.loads(result.stdout)
        self.assertEqual("JiniFrameworkEvolutionReview", review["review_type"])
        self.assertTrue(self.resolve_repo_path(review["review_path"]).exists())
        self.assertEqual(3, len(review["prioritized_dimensions"]))
        self.assertTrue(review["adoption_constraints"])
        self.assertTrue(review["best_next_dimension"])
        self.assertTrue(
            any("removed, demoted, or consolidated" in item for item in review["adoption_constraints"])
        )
        top_ids = {item["id"] for item in review["prioritized_dimensions"]}
        self.assertTrue(top_ids & {"packaging-install", "delivery-maturity", "adapter-portability", "token-efficiency"})
        self.assertTrue(review["prioritized_dimensions"][0]["cleanup_candidates"])
        self.assertTrue(review["best_next_experiment"])
        self.assertEqual("subtractive", review["best_next_experiment"]["change_type"])

    def test_stage_framework_experiment_can_target_dimension(self) -> None:
        self.assert_ok(self.run_cli("review-framework"))

        result = self.run_cli(
            "stage-framework-experiment",
            "--dimension",
            "delivery-maturity",
            "--format",
            "json",
        )
        self.assert_ok(result)

        experiment = json.loads(result.stdout)
        self.assertEqual("JiniFrameworkEvolutionExperiment", experiment["experiment_type"])
        self.assertEqual("proposed", experiment["status"])
        self.assertEqual("delivery-maturity", experiment["dimension_id"])
        self.assertTrue(self.resolve_repo_path(experiment["experiment_path"]).exists())
        self.assertEqual("subtractive", experiment["change_type"])
        self.assertTrue(experiment["change_plan"])
        self.assertIn("adoption_weight", experiment["reward_model"])

    def test_record_framework_outcome_and_backtest_reflect_reward(self) -> None:
        self.assert_ok(self.run_cli("review-framework"))
        stage = self.run_cli(
            "stage-framework-experiment",
            "--dimension",
            "adapter-portability",
            "--format",
            "json",
        )
        self.assert_ok(stage)
        experiment = json.loads(stage.stdout)
        experiment_path = self.resolve_repo_path(experiment["experiment_path"])

        outcome = self.run_cli(
            "record-framework-outcome",
            experiment_path,
            "--actor",
            "platform-lead",
            "--result",
            "success",
            "--score-delta",
            "0.3",
            "--signal",
            "portable edge validated",
            "--signal",
            "fewer manual handoff steps",
            "--note",
            "Local-apply loop reduced staging-only friction",
            "--format",
            "json",
        )
        self.assert_ok(outcome)

        outcome_doc = json.loads(outcome.stdout)
        self.assertEqual("JiniFrameworkEvolutionOutcome", outcome_doc["outcome_type"])
        self.assertEqual("success", outcome_doc["result"])
        self.assertTrue(self.resolve_repo_path(outcome_doc["outcome_path"]).exists())
        self.assertGreater(outcome_doc["computed_reward"], 0)

        updated_experiment = self.read_json(experiment_path)
        self.assertEqual("completed", updated_experiment["status"])
        self.assertEqual(outcome_doc["outcome_path"], updated_experiment["latest_outcome_path"])

        backtest = self.run_cli("backtest-framework-evolution", "--format", "json")
        self.assert_ok(backtest)
        backtest_doc = json.loads(backtest.stdout)
        self.assertGreaterEqual(backtest_doc["outcome_count"], 1)
        adapter_summary = next(
            item for item in backtest_doc["dimension_summaries"] if item["dimension_id"] == "adapter-portability"
        )
        self.assertGreaterEqual(adapter_summary["experiments"], 1)
        self.assertGreaterEqual(adapter_summary["successes"], 1)
        self.assertGreater(adapter_summary["average_score_delta"], 0)

    def test_plan_install_json_reports_provenance_targets_and_receipt(self) -> None:
        result = self.run_cli(
            "plan-install",
            "--bundle",
            "jini-core",
            "--target",
            "codex",
            "--target",
            "kiro-cli",
            "--format",
            "json",
        )
        self.assert_ok(result)

        plan = json.loads(result.stdout)
        self.assertEqual("Jini", plan["source"]["name"])
        self.assertEqual(2, len(plan["selected_targets"]))
        self.assertEqual(1, len(plan["selected_bundles"]))
        self.assertEqual("jini-core", plan["selected_bundles"][0]["id"])
        self.assertTrue(plan["selected_bundles"][0]["review_before_use"])
        self.assertEqual("install", plan["recommended_next_steps"][0]["id"])
        self.assertIn("install-bundles --bundle jini-core", plan["recommended_next_steps"][0]["command"])
        self.assertEqual("verify", plan["recommended_next_steps"][1]["id"])
        self.assertIn("doctor-install --bundle jini-core", plan["recommended_next_steps"][1]["command"])
        self.assertTrue(plan["receipt"]["receipt_id"])
        self.assertTrue(any("permission risk" in notice for notice in plan["risk_notices"]))

    def test_catalog_bundles_reports_kits_and_bundle_metadata(self) -> None:
        result = self.run_cli("catalog-bundles", "--format", "json")
        self.assert_ok(result)

        catalog = json.loads(result.stdout)
        self.assertEqual("JiniInstallCatalog", catalog["catalog_type"])
        bundle_ids = {item["id"] for item in catalog["bundles"]}
        kit_ids = {item["id"] for item in catalog["kits"]}
        self.assertIn("framework-evolution-surface", bundle_ids)
        self.assertIn("runtime-portability-surface", bundle_ids)
        self.assertIn("meeting-followup-pack", bundle_ids)
        self.assertIn("incident-response-pack", bundle_ids)
        self.assertIn("compliance-audit-pack", bundle_ids)
        self.assertIn("vendor-selection-pack", bundle_ids)
        self.assertIn("starter-kit", kit_ids)
        self.assertIn("benchmark-delivery-kit", kit_ids)
        self.assertIn("operations-response-kit", kit_ids)
        self.assertIn("regulated-readiness-kit", kit_ids)
        self.assertIn("vendor-decision-kit", kit_ids)
        self.assertEqual("starter-kit", catalog["recommended_paths"]["beginner"]["kit_id"])
        self.assertEqual(["jini"], catalog["recommended_paths"]["beginner"]["commands"])
        self.assertEqual("benchmark-delivery-kit", catalog["recommended_paths"]["power_user"]["kit_id"])
        starter_kit = next(item for item in catalog["kits"] if item["id"] == "starter-kit")
        self.assertEqual("beginner", starter_kit["audience"])
        benchmark_kit = next(item for item in catalog["kits"] if item["id"] == "benchmark-delivery-kit")
        self.assertIn("meeting-followup-pack", benchmark_kit["bundle_ids"])
        framework_bundle = next(item for item in catalog["bundles"] if item["id"] == "framework-evolution-surface")
        self.assertTrue(framework_bundle["activation_steps"])
        self.assertIn("specs/learning-system.md", framework_bundle["universal_paths"])

    def test_plan_install_kit_expands_bundle_set_and_activation_steps(self) -> None:
        result = self.run_cli(
            "plan-install",
            "--kit",
            "starter-kit",
            "--target",
            "codex",
            "--format",
            "json",
        )
        self.assert_ok(result)

        plan = json.loads(result.stdout)
        self.assertEqual("starter-kit", plan["selected_kits"][0]["id"])
        bundle_ids = {item["id"] for item in plan["selected_bundles"]}
        self.assertIn("jini-core", bundle_ids)
        self.assertIn("framework-evolution-surface", bundle_ids)
        self.assertTrue(plan["selected_targets"][0]["activation_steps"])
        self.assertTrue(any(bundle["activation_steps"] for bundle in plan["selected_bundles"]))
        self.assertIn("install-bundles --kit starter-kit", plan["recommended_next_steps"][0]["command"])
        self.assertIn("doctor-install --kit starter-kit", plan["recommended_next_steps"][1]["command"])
        self.assertIn("manifest_digest", plan["receipt"])

    def test_plan_install_regulated_kit_expands_bundle_set(self) -> None:
        result = self.run_cli(
            "plan-install",
            "--kit",
            "regulated-readiness-kit",
            "--target",
            "codex",
            "--format",
            "json",
        )
        self.assert_ok(result)

        plan = json.loads(result.stdout)
        self.assertEqual("regulated-readiness-kit", plan["selected_kits"][0]["id"])
        bundle_ids = {item["id"] for item in plan["selected_bundles"]}
        self.assertIn("jini-core", bundle_ids)
        self.assertIn("compliance-audit-pack", bundle_ids)
        self.assertIn("runtime-portability-surface", bundle_ids)

    def test_plan_install_vendor_kit_expands_bundle_set(self) -> None:
        result = self.run_cli(
            "plan-install",
            "--kit",
            "vendor-decision-kit",
            "--target",
            "codex",
            "--format",
            "json",
        )
        self.assert_ok(result)

        plan = json.loads(result.stdout)
        self.assertEqual("vendor-decision-kit", plan["selected_kits"][0]["id"])
        bundle_ids = {item["id"] for item in plan["selected_bundles"]}
        self.assertIn("jini-core", bundle_ids)
        self.assertIn("vendor-selection-pack", bundle_ids)
        self.assertIn("runtime-portability-surface", bundle_ids)

    def test_plan_install_without_selection_uses_default_starter_kit(self) -> None:
        result = self.run_cli("plan-install", "--target", "codex", "--format", "json")
        self.assert_ok(result)

        plan = json.loads(result.stdout)
        self.assertEqual("starter-kit", plan["default_kit_id"])
        self.assertEqual(["starter-kit"], [item["id"] for item in plan["selected_kits"]])
        bundle_ids = {item["id"] for item in plan["selected_bundles"]}
        self.assertIn("jini-core", bundle_ids)
        self.assertNotIn("travel-plan-pack", bundle_ids)

    def test_plan_install_text_surfaces_shim_strategy(self) -> None:
        result = self.run_cli(
            "plan-install",
            "--bundle",
            "research-prd-pack",
            "--target",
            "github-copilot",
            "--link-mode",
            "copy",
        )
        self.assert_ok(result)
        self.assertIn("SOURCE  Jini", result.stdout)
        self.assertIn("GitHub Copilot: copy", result.stdout)
        self.assertIn("RECEIPT", result.stdout)

    def test_install_bundles_materializes_receipt_and_paths_under_prefix(self) -> None:
        prefix = self.install_prefix()
        result = self.run_cli(
            "install-bundles",
            "--bundle",
            "jini-core",
            "--target",
            "codex",
            "--prefix",
            prefix,
            "--format",
            "json",
        )
        self.assert_ok(result)

        report = json.loads(result.stdout)
        self.assertEqual("installed", report["status"])
        install = report["installs"][0]
        universal_root = Path(install["universal_destination"])
        shim_root = Path(install["shim_destination"])
        self.assertTrue(universal_root.exists())
        self.assertTrue(shim_root.exists())
        self.assertTrue((universal_root / "tools" / "jini.py").is_symlink())
        self.assertTrue(Path(report["receipt_path"]).exists())

    def test_doctor_install_reports_ok_after_install(self) -> None:
        prefix = self.install_prefix()
        self.assert_ok(
            self.run_cli(
                "install-bundles",
                "--bundle",
                "research-prd-pack",
                "--target",
                "github-copilot",
                "--prefix",
                prefix,
                "--link-mode",
                "copy",
            )
        )

        result = self.run_cli(
            "doctor-install",
            "--bundle",
            "research-prd-pack",
            "--target",
            "github-copilot",
            "--prefix",
            prefix,
            "--format",
            "json",
        )
        self.assert_ok(result)
        report = json.loads(result.stdout)
        self.assertEqual("ok", report["status"])
        self.assertEqual("github-copilot", report["checks"][0]["target_id"])
        self.assertEqual([], report["checks"][0]["missing_paths"])
        self.assertEqual("installed", report["checks"][0]["latest_receipt_status"])
        self.assertTrue(report["checks"][0]["activation_steps"])
        self.assertTrue(any(item["id"] == "receipt-present" and item["status"] == "ok" for item in report["checks"][0]["health_checks"]))
        self.assertTrue(any(item["id"] == "receipt-manifest-current" and item["status"] == "ok" for item in report["checks"][0]["health_checks"]))
        self.assertTrue(any(item["id"] == "install-link-mode" and item["status"] == "ok" for item in report["checks"][0]["health_checks"]))

    def test_doctor_install_reports_ready_activation_after_runtime_activation(self) -> None:
        pack_dir = self.compile_research_pack()
        repo = self.create_repo_fixture()
        prefix = self.install_prefix()

        activate = self.run_cli(
            "activate-runtime-target",
            pack_dir,
            "--repo",
            repo,
            "--runtime-target",
            "codex",
            "--prefix",
            prefix,
            "--format",
            "json",
        )
        self.assert_ok(activate)

        result = self.run_cli(
            "doctor-install",
            "--bundle",
            "jini-core",
            "--target",
            "codex",
            "--prefix",
            prefix,
            "--format",
            "json",
        )
        self.assert_ok(result)
        report = json.loads(result.stdout)
        check = report["checks"][0]
        self.assertEqual("ok", check["status"])
        self.assertEqual("ready", check["activation_status"])
        self.assertTrue(check["activation_roots"])
        self.assertTrue(check["healthy_activation_roots"])
        self.assertTrue(any(item["id"] == "runtime-activation" and item["status"] == "ok" for item in check["health_checks"]))
        self.assertTrue(any(item["id"] == "install-link-mode" and item["status"] == "ok" for item in check["health_checks"]))
        self.assertTrue(any(item["id"] == "shim-target-behavior" and item["status"] == "ok" for item in check["health_checks"]))
        self.assertTrue(any(item["id"] == "smoke-catalog-bundles" and item["status"] == "ok" for item in check["health_checks"]))
        self.assertTrue(any(item["id"] == "smoke-get-started" and item["status"] == "ok" for item in check["health_checks"]))
        self.assertTrue(any(item["id"] == "smoke-plan-install" and item["status"] == "ok" for item in check["health_checks"]))
        self.assertTrue(any(item["id"] == "smoke-resolve-adapter" and item["status"] == "ok" for item in check["health_checks"]))
        self.assertTrue(any(item["id"] == "runtime-activation-target-match" and item["status"] == "ok" for item in check["health_checks"]))

    def test_uninstall_bundles_removes_materialized_install(self) -> None:
        prefix = self.install_prefix()
        self.assert_ok(
            self.run_cli(
                "install-bundles",
                "--bundle",
                "jini-core",
                "--target",
                "kiro-cli",
                "--prefix",
                prefix,
            )
        )

        uninstall = self.run_cli(
            "uninstall-bundles",
            "--bundle",
            "jini-core",
            "--target",
            "kiro-cli",
            "--prefix",
            prefix,
            "--format",
            "json",
        )
        self.assert_ok(uninstall)
        report = json.loads(uninstall.stdout)
        self.assertEqual("uninstalled", report["status"])

        doctor = self.run_cli(
            "doctor-install",
            "--bundle",
            "jini-core",
            "--target",
            "kiro-cli",
            "--prefix",
            prefix,
            "--format",
            "json",
        )
        self.assert_ok(doctor)
        doctor_report = json.loads(doctor.stdout)
        self.assertEqual("missing", doctor_report["status"])
        self.assertTrue(doctor_report["checks"][0]["missing_paths"])

    def test_update_bundles_refreshes_copied_install_from_source(self) -> None:
        prefix = self.install_prefix()
        self.assert_ok(
            self.run_cli(
                "install-bundles",
                "--bundle",
                "research-prd-pack",
                "--target",
                "github-copilot",
                "--prefix",
                prefix,
                "--link-mode",
                "copy",
            )
        )

        installed_readme = (
            prefix
            / ".copilot"
            / "jini"
            / "bundles"
            / "research-prd-pack"
            / "packs"
            / "research-prd"
            / "README.md"
        )
        installed_readme.write_text("stale copy\n", encoding="utf-8")

        update = self.run_cli(
            "update-bundles",
            "--bundle",
            "research-prd-pack",
            "--target",
            "github-copilot",
            "--prefix",
            prefix,
            "--link-mode",
            "copy",
            "--format",
            "json",
        )
        self.assert_ok(update)
        report = json.loads(update.stdout)
        self.assertEqual("updated", report["status"])
        self.assertEqual(
            (REPO_ROOT / "packs" / "research-prd" / "README.md").read_text(encoding="utf-8"),
            installed_readme.read_text(encoding="utf-8"),
        )

    def test_show_kpis_dimension_filter_surfaces_competitor_and_conversion(self) -> None:
        result = self.run_cli("show-kpis", "--dimension", "token")
        self.assert_ok(result)
        self.assertIn("DIM    Token efficiency (token-efficiency)", result.stdout)
        self.assertIn("BEST   Aider (9.0)", result.stdout)
        self.assertIn("low-token reloads", result.stdout)

    def test_metrics_reports_canonical_surface_without_alias_debt(self) -> None:
        result = self.run_cli("metrics")
        self.assert_ok(result)
        self.assertIn("STATUS   ok", result.stdout)
        self.assertIn("CMDS     2", result.stdout)
        self.assertIn("  - doctor", result.stdout)
        self.assertIn("  - status", result.stdout)
        self.assertIn("ALIASES  0", result.stdout)
        self.assertIn("SAMPLES  count=5 ok=5", result.stdout)
        self.assertIn("  - jini commands", result.stdout)
        self.assertIn("  - jini doctor --format json", result.stdout)
        self.assertIn("  - jini status packs/research-prd/examples/research-prd-v1", result.stdout)
        self.assertIn("  - jini continue --from packs/research-prd/examples/research-prd-v1", result.stdout)
        self.assertIn("  - jini resume packs/research-prd/examples/research-prd-v1 --format json --max-chars 700", result.stdout)
        self.assertIn("RESUME   available=yes status=measured", result.stdout)
        self.assertIn("PROVIDER available=yes id=local-preview status=ok", result.stdout)
        self.assertIn("  label=Local preview", result.stdout)
        self.assertIn("ROUTETREND available=no status=unavailable improving=0 stable=0 regressing=0", result.stdout)
        self.assertIn("ROUTECOST available=no status=unavailable basis=none posture=unknown", result.stdout)
        self.assertIn("COST     token-efficiency", result.stdout)
        self.assertIn("LATENCY  delivery-maturity", result.stdout)

    def test_metrics_json_reports_counts_and_zero_aliases(self) -> None:
        result = self.run_cli("metrics", "--format", "json")
        self.assert_ok(result)
        payload = json.loads(result.stdout)
        self.assertEqual("ok", payload["status"])
        self.assertEqual(2, payload["command_surface_count"])
        self.assertEqual(0, payload["compatibility_alias_count"])
        self.assertEqual(["doctor", "status"], payload["taught_commands"])
        self.assertEqual(5, payload["latency_sample"]["sample_count"])
        self.assertEqual(5, payload["latency_sample"]["successful_sample_count"])
        self.assertEqual("jini commands", payload["command_samples"][0]["command"])
        self.assertEqual("jini doctor --format json", payload["command_samples"][1]["command"])
        self.assertEqual("jini status packs/research-prd/examples/research-prd-v1", payload["command_samples"][2]["command"])
        self.assertEqual("jini continue --from packs/research-prd/examples/research-prd-v1", payload["command_samples"][3]["command"])
        self.assertEqual(
            "jini resume packs/research-prd/examples/research-prd-v1 --format json --max-chars 700",
            payload["command_samples"][4]["command"],
        )
        self.assertTrue(all(item["exit_code"] == 0 for item in payload["command_samples"]))
        self.assertTrue(all("stdout_preview" not in item for item in payload["command_samples"]))
        self.assertTrue(all("stderr_preview" not in item for item in payload["command_samples"]))
        self.assertTrue(payload["resume_cost"]["available"])
        self.assertEqual("measured", payload["resume_cost"]["status"])
        self.assertEqual("continue", payload["resume_cost"]["cheaper_surface"])
        self.assertLess(payload["resume_cost"]["continue_output_chars"], payload["resume_cost"]["resume_output_chars"])
        self.assertTrue(payload["provider_evidence"]["available"])
        self.assertEqual("local-preview", payload["provider_evidence"]["provider_id"])
        self.assertEqual("Local preview", payload["provider_evidence"]["label"])
        self.assertEqual("ok", payload["provider_evidence"]["status"])
        self.assertFalse(payload["route_trend"]["available"])
        self.assertEqual("unavailable", payload["route_trend"]["status"])
        self.assertEqual(0, payload["route_trend"]["improving_count"])
        self.assertEqual(0, payload["route_trend"]["stable_count"])
        self.assertEqual(0, payload["route_trend"]["regressing_count"])
        self.assertFalse(payload["route_cost"]["available"])
        self.assertEqual("unavailable", payload["route_cost"]["status"])
        self.assertEqual("none", payload["route_cost"]["basis"])
        self.assertEqual("unknown", payload["route_cost"]["posture"])
        self.assertFalse(payload["route_evidence"]["available"])
        self.assertEqual(0, payload["route_evidence"]["adapter_count"])
        self.assertEqual("token-efficiency", payload["cost_proxy"]["dimension"])
        self.assertEqual("delivery-maturity", payload["latency_proxy"]["dimension"])

    def test_metrics_json_surfaces_local_route_evidence_when_present(self) -> None:
        state_root = self.tmp / ".jini"
        state_root.mkdir(parents=True, exist_ok=True)
        report = {
            "schema_version": "0.4.0",
            "context_type": "JiniLocalRuntimeCapabilities",
            "captured_at": "2026-05-21T19:00:00Z",
            "jini_version": "0.1.0",
            "capability_registry_version": "test",
            "device_probe_fingerprint": "test-fingerprint",
            "local_endpoint_signature": "local-endpoint",
            "local_runtime_class": "local-ollama",
            "adapters": {
                "local-workhorse": {
                    "adapter_id": "local-workhorse",
                    "model_id": "qwen3:8b",
                    "status": "ok",
                    "latency_ms": 180,
                    "warm_latency_ms": 150,
                    "cold_start_cost_ms": 30,
                    "tokens_per_second": 22.5,
                    "quality_class": "strong",
                    "structured_reliability": "strong",
                    "benchmarked_at": "2026-05-21T19:00:00Z",
                },
                "local-fast": {
                    "adapter_id": "local-fast",
                    "model_id": "phi4-mini",
                    "status": "degraded",
                    "latency_ms": 95,
                    "warm_latency_ms": 90,
                    "cold_start_cost_ms": 5,
                    "tokens_per_second": 38.2,
                    "quality_class": "usable",
                    "structured_reliability": "usable",
                    "benchmarked_at": "2026-05-21T19:00:00Z",
                },
            },
            "history": {
                "local-workhorse": [
                    {
                        "model_id": "qwen3:8b",
                        "status": "failed",
                        "latency_ms": 320,
                        "tokens_per_second": 16.0,
                        "quality_class": "weak",
                        "structured_reliability": "usable",
                        "benchmarked_at": "2026-05-21T18:40:00Z",
                    },
                    {
                        "model_id": "qwen3:8b",
                        "status": "degraded",
                        "latency_ms": 260,
                        "tokens_per_second": 19.0,
                        "quality_class": "usable",
                        "structured_reliability": "usable",
                        "benchmarked_at": "2026-05-21T18:50:00Z",
                    },
                    {
                        "model_id": "qwen3:8b",
                        "status": "ok",
                        "latency_ms": 180,
                        "tokens_per_second": 22.5,
                        "quality_class": "strong",
                        "structured_reliability": "strong",
                        "benchmarked_at": "2026-05-21T19:00:00Z",
                    },
                ],
                "local-fast": [
                    {
                        "model_id": "phi4-mini",
                        "status": "ok",
                        "latency_ms": 40,
                        "tokens_per_second": 41.0,
                        "quality_class": "usable",
                        "structured_reliability": "usable",
                        "benchmarked_at": "2026-05-21T18:40:00Z",
                    },
                    {
                        "model_id": "phi4-mini",
                        "status": "ok",
                        "latency_ms": 48,
                        "tokens_per_second": 39.0,
                        "quality_class": "usable",
                        "structured_reliability": "usable",
                        "benchmarked_at": "2026-05-21T18:50:00Z",
                    },
                    {
                        "model_id": "phi4-mini",
                        "status": "degraded",
                        "latency_ms": 95,
                        "tokens_per_second": 38.2,
                        "quality_class": "usable",
                        "structured_reliability": "usable",
                        "benchmarked_at": "2026-05-21T19:00:00Z",
                    },
                ],
            },
        }
        (state_root / "local-runtime-capabilities.json").write_text(
            json.dumps(report),
            encoding="utf-8",
        )

        result = self.run_cli("metrics", "--format", "json")
        self.assert_ok(result)
        payload = json.loads(result.stdout)
        self.assertTrue(payload["route_evidence"]["available"])
        self.assertEqual("local-ollama", payload["route_evidence"]["local_runtime_class"])
        self.assertEqual(2, payload["route_evidence"]["adapter_count"])
        self.assertEqual(2, payload["route_evidence"]["ready_adapter_count"])
        self.assertTrue(payload["route_cost"]["available"])
        self.assertEqual("measured", payload["route_cost"]["status"])
        self.assertEqual("local-runtime-benchmark", payload["route_cost"]["basis"])
        self.assertEqual("zero-external-api-spend", payload["route_cost"]["posture"])
        self.assertEqual(2, payload["route_cost"]["ready_adapter_count"])
        self.assertEqual(120.0, payload["route_cost"]["avg_ready_warm_latency_ms"])
        self.assertEqual(30.4, payload["route_cost"]["avg_ready_tokens_per_second"])
        self.assertEqual("local-fast", payload["route_cost"]["cheapest_ready_adapter"]["adapter_id"])
        self.assertEqual(5, payload["route_cost"]["cheapest_ready_adapter"]["cold_start_cost_ms"])
        self.assertTrue(payload["route_trend"]["available"])
        self.assertEqual("measured", payload["route_trend"]["status"])
        self.assertEqual(1, payload["route_trend"]["improving_count"])
        self.assertEqual(0, payload["route_trend"]["stable_count"])
        self.assertEqual(1, payload["route_trend"]["regressing_count"])
        trend_rows = {
            item["adapter_id"]: item
            for item in payload["route_trend"]["adapters"]
        }
        self.assertEqual("recovered", trend_rows["local-workhorse"]["trend"])
        self.assertEqual("slower", trend_rows["local-fast"]["trend"])
        adapters = {
            item["adapter_id"]: item
            for item in payload["route_evidence"]["adapters"]
        }
        self.assertEqual(2, len(adapters))
        self.assertEqual(180, adapters["local-workhorse"]["latency_ms"])
        self.assertEqual("strong", adapters["local-workhorse"]["structured_reliability"])
        self.assertEqual(95, adapters["local-fast"]["latency_ms"])
        self.assertEqual("usable", adapters["local-fast"]["structured_reliability"])


if __name__ == "__main__":
    unittest.main()
