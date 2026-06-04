import tempfile
import unittest
from pathlib import Path
from unittest import mock
import sys

import tools.jini as go_launcher


class GoFirstLauncherTests(unittest.TestCase):
    def test_go_command_handles_native_single_token_inventory(self) -> None:
        for command in ["help", "--help", "-h", "commands", "init", "memory", "new", "permissions", "route", "doctor"]:
            with self.subTest(command=command):
                self.assertTrue(go_launcher._should_use_go([command]))
                self.assertFalse(go_launcher._should_use_go([command, "extra"]))

    def test_go_command_handles_doctor_json_surface(self) -> None:
        self.assertTrue(go_launcher._should_use_go(["doctor", "--format", "json"]))
        self.assertTrue(go_launcher._should_use_go(["provider", "--format", "json"]))
        self.assertTrue(go_launcher._should_use_go(["provider", "doctor", "--format", "json"]))
        self.assertTrue(go_launcher._should_use_go(["doctor", "--format=json"]))
        self.assertTrue(go_launcher._should_use_go(["provider", "--format=json"]))
        self.assertTrue(go_launcher._should_use_go(["provider", "doctor", "--format=json"]))

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

    def test_go_command_handles_check_surface(self) -> None:
        self.assertTrue(go_launcher._should_use_go(["check"]))
        self.assertTrue(go_launcher._should_use_go(["check", "/tmp/example-pack"]))
        self.assertFalse(go_launcher._should_use_go(["check", "/tmp/example-pack", "extra"]))

    def test_go_command_handles_native_utility_surfaces(self) -> None:
        self.assertTrue(go_launcher._should_use_go(["init"]))
        self.assertTrue(go_launcher._should_use_go(["memory"]))
        self.assertTrue(go_launcher._should_use_go(["new"]))
        self.assertTrue(go_launcher._should_use_go(["permissions"]))
        self.assertTrue(go_launcher._should_use_go(["route"]))

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
