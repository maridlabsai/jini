import json
import tempfile
import unittest
from pathlib import Path
from unittest import mock
import sys

import tools.jini as go_launcher


BOUNDARY_FIXTURE = Path(__file__).resolve().parent / "fixtures" / "go_boundary_contract.json"


def _load_boundary_cases() -> list[dict[str, object]]:
    return json.loads(BOUNDARY_FIXTURE.read_text(encoding="utf-8"))


class GoFirstLauncherTests(unittest.TestCase):
    def test_go_command_matches_shared_boundary_fixture(self) -> None:
        for case in _load_boundary_cases():
            with self.subTest(case=case["name"]):
                self.assertEqual(case["launcher_go"], go_launcher._should_use_go(case["args"]))

    def test_go_command_handles_doctor_json_surface(self) -> None:
        self.assertTrue(go_launcher._should_use_go(["doctor", "--format", "json"]))
        self.assertTrue(go_launcher._should_use_go(["provider", "--format", "json"]))
        self.assertTrue(go_launcher._should_use_go(["provider", "doctor", "--format", "json"]))
        self.assertTrue(go_launcher._should_use_go(["doctor", "--format=json"]))
        self.assertTrue(go_launcher._should_use_go(["provider", "--format=json"]))
        self.assertTrue(go_launcher._should_use_go(["provider", "doctor", "--format=json"]))
        self.assertTrue(go_launcher._should_use_go(["doctor", "--format=text"]))
        self.assertTrue(go_launcher._should_use_go(["provider", "--format=text"]))
        self.assertTrue(go_launcher._should_use_go(["provider", "doctor", "--format=text"]))

    def test_publish_readiness_json_stays_on_legacy_surface(self) -> None:
        self.assertFalse(go_launcher._should_use_go(["publish-readiness", "--format", "json"]))

    def test_unsupported_doctor_json_shapes_stay_on_legacy_surface(self) -> None:
        self.assertFalse(go_launcher._should_use_go(["doctor", "extra", "--format", "json"]))
        self.assertFalse(go_launcher._should_use_go(["doctor", "extra", "--format=json"]))
        self.assertFalse(go_launcher._should_use_go(["provider", "extra", "--format", "json"]))
        self.assertFalse(go_launcher._should_use_go(["provider", "extra", "--format=json"]))
        self.assertFalse(go_launcher._should_use_go(["provider", "doctor", "extra", "--format", "json"]))
        self.assertFalse(go_launcher._should_use_go(["provider", "doctor", "extra", "--format=json"]))

    def test_go_command_handles_direct_provider_surface(self) -> None:
        self.assertTrue(go_launcher._should_use_go(["provider"]))
        self.assertTrue(go_launcher._should_use_go(["provider", "--format", "text"]))
        self.assertTrue(go_launcher._should_use_go(["provider", "--format", "json"]))
        self.assertTrue(go_launcher._should_use_go(["provider", "--format=text"]))
        self.assertTrue(go_launcher._should_use_go(["provider", "--format=json"]))

    def test_go_command_handles_observe_surface(self) -> None:
        self.assertTrue(go_launcher._should_use_go(["observe"]))
        self.assertTrue(go_launcher._should_use_go(["observe", "status"]))
        self.assertTrue(go_launcher._should_use_go(["observe", "scan"]))
        self.assertTrue(go_launcher._should_use_go(["observe", "add", "/tmp/outside-notes.md"]))
        self.assertTrue(
            go_launcher._should_use_go(
                ["observe", "add", "--connector", "markdown", "/tmp/outside-notes.md"]
            )
        )
        self.assertTrue(
            go_launcher._should_use_go(
                ["observe", "add", "/tmp/outside-notes.md", "--connector=markdown"]
            )
        )
        self.assertFalse(go_launcher._should_use_go(["observe", "help"]))
        self.assertFalse(go_launcher._should_use_go(["observe", "extra"]))
        self.assertFalse(go_launcher._should_use_go(["observe", "status", "extra"]))
        self.assertFalse(go_launcher._should_use_go(["observe", "scan", "extra"]))
        self.assertFalse(go_launcher._should_use_go(["observe", "add"]))
        self.assertFalse(go_launcher._should_use_go(["observe", "add", "one", "two"]))
        self.assertFalse(go_launcher._should_use_go(["observe", "add", "--connector"]))

    def test_go_command_handles_open_surface(self) -> None:
        self.assertTrue(go_launcher._should_use_go(["open", "prd"]))
        self.assertFalse(go_launcher._should_use_go(["open", "prd", "--print-path"]))

    def test_go_command_prefers_local_go_over_repo_binary(self) -> None:
        with tempfile.TemporaryDirectory(prefix="jini-go-launcher-") as tempdir:
            root = Path(tempdir)
            built_binary = root / "jini"
            built_binary.write_text("#!/bin/sh\nexit 0\n", encoding="utf-8")
            built_binary.chmod(0o755)

            local_go = root / "go"
            local_go.write_text("#!/bin/sh\nexit 0\n", encoding="utf-8")
            local_go.chmod(0o755)

            with mock.patch.object(go_launcher, "ROOT", root), mock.patch.object(go_launcher, "LOCAL_GO", local_go):
                command = go_launcher._go_command(["doctor"])

            self.assertEqual([str(local_go), "run", "./cmd/jini", "doctor"], command)

    def test_go_command_uses_repo_binary_when_go_is_unavailable(self) -> None:
        with tempfile.TemporaryDirectory(prefix="jini-go-launcher-") as tempdir:
            root = Path(tempdir)
            built_binary = root / "jini"
            built_binary.write_text("#!/bin/sh\nexit 0\n", encoding="utf-8")
            built_binary.chmod(0o755)

            missing_go = root / "missing-go"

            with (
                mock.patch.object(go_launcher, "ROOT", root),
                mock.patch.object(go_launcher, "LOCAL_GO", missing_go),
                mock.patch("tools.jini.shutil.which", return_value=None),
            ):
                command = go_launcher._go_command(["doctor"])

            self.assertEqual([str(built_binary), "doctor"], command)

    def test_main_exports_legacy_python_for_go_boundary(self) -> None:
        with mock.patch.object(go_launcher, "_should_use_go", return_value=True), mock.patch.object(
            go_launcher, "_go_command", return_value=["go", "run", "./cmd/jini", "doctor"]
        ), mock.patch("tools.jini.subprocess.run") as run_mock:
            run_mock.return_value.returncode = 0
            exit_code = go_launcher.main(["doctor"])

        self.assertEqual(0, exit_code)
        _, kwargs = run_mock.call_args
        self.assertEqual(str(Path(sys.executable).resolve()), kwargs["env"]["JINI_LEGACY_PYTHON"])


if __name__ == "__main__":
    unittest.main()
