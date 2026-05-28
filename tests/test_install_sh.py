import os
import json
import shutil
import stat
import subprocess
import sys
import tempfile
import unittest
from pathlib import Path
from typing import Optional


REPO_ROOT = Path(__file__).resolve().parents[1]
INSTALLER = REPO_ROOT / "install.sh"
LOCAL_GO_BIN = REPO_ROOT.parent / ".local-go" / "bin"
CLI_DOC = REPO_ROOT / "docs" / "cli.md"
RESEARCH_EXAMPLE = REPO_ROOT / "packs" / "research-prd" / "examples" / "research-prd-v1"


class InstallScriptTests(unittest.TestCase):
    maxDiff = None

    def setUp(self) -> None:
        self.temp_dir = tempfile.TemporaryDirectory(prefix="jini-install-tests-")
        self.tmp = Path(self.temp_dir.name)

    def tearDown(self) -> None:
        for path in self.tmp.rglob("*"):
            try:
                if path.is_dir():
                    path.chmod(0o755)
                else:
                    path.chmod(0o644)
            except OSError:
                pass
        self.temp_dir.cleanup()

    def run_installer(
        self,
        *args: str,
        env: Optional[dict[str, str]] = None,
        cwd: Optional[Path] = None,
        installer_path: Optional[Path] = None,
    ) -> subprocess.CompletedProcess[str]:
        run_env = dict(os.environ)
        if env:
            run_env.update(env)
        return subprocess.run(
            ["/bin/bash", str(installer_path or INSTALLER), *args],
            cwd=str(cwd or REPO_ROOT),
            text=True,
            capture_output=True,
            env=run_env,
        )

    def run_installer_via_stdin(
        self,
        *args: str,
        env: Optional[dict[str, str]] = None,
        cwd: Optional[Path] = None,
        installer_path: Optional[Path] = None,
    ) -> subprocess.CompletedProcess[str]:
        run_env = dict(os.environ)
        if env:
            run_env.update(env)
        script_text = (installer_path or INSTALLER).read_text(encoding="utf-8")
        return subprocess.run(
            ["/bin/bash", "-s", "--", *args],
            cwd=str(cwd or REPO_ROOT),
            text=True,
            input=script_text,
            capture_output=True,
            env=run_env,
        )

    def go_ready_env(self, extra_path: str = "") -> dict[str, str]:
        path_parts = [str(LOCAL_GO_BIN)]
        if extra_path:
            path_parts.append(extra_path)
        path_parts.append(os.environ.get("PATH", ""))
        return {"PATH": ":".join(part for part in path_parts if part)}

    def assert_ok(self, result: subprocess.CompletedProcess[str]) -> None:
        if result.returncode != 0:
            self.fail(
                f"Expected installer to succeed.\nSTDOUT:\n{result.stdout}\nSTDERR:\n{result.stderr}"
            )

    def assert_error(self, result: subprocess.CompletedProcess[str]) -> None:
        if result.returncode == 0:
            self.fail(f"Expected installer to fail.\nSTDOUT:\n{result.stdout}")

    def run_installed_jini(
        self,
        binary_path: Path,
        *args: str,
        env: Optional[dict[str, str]] = None,
        cwd: Optional[Path] = None,
    ) -> subprocess.CompletedProcess[str]:
        run_env = dict(os.environ)
        if env:
            run_env.update(env)
        return subprocess.run(
            [str(binary_path), *args],
            cwd=str(cwd or REPO_ROOT),
            text=True,
            capture_output=True,
            env=run_env,
        )

    def write_current_work(self, state_dir: Path, pack_dir: Path) -> None:
        state_dir.mkdir(parents=True, exist_ok=True)
        payload = {
            "pack_dir": str(pack_dir),
            "pack_id": "research-prd",
            "work_unit_id": "research-prd-v1",
            "title": "Research PRD Example",
            "state": "awaiting_verification",
            "health": "ready-to-verify",
        }
        (state_dir / "current-work.json").write_text(json.dumps(payload), encoding="utf-8")

    def assert_go_public_command_contract(self, binary_path: Path, *, env: Optional[dict[str, str]] = None) -> None:
        state_dir = self.tmp / ".jini-go-artifact-state"
        self.write_current_work(state_dir, RESEARCH_EXAMPLE)
        command_env = {"JINI_STATE_DIR": str(state_dir), "JINI_PROVIDER": "local-preview"}
        if env:
            command_env.update(env)

        smoke_cases = [
            (["--help"], 0, "Actions"),
            (["commands"], 0, "Public command inventory"),
            (["help", "--all"], 0, "Public command inventory"),
            (["admin", "help"], 0, "Admin and developer command inventory"),
            (["doctor"], 0, "Provider"),
            (["status"], 0, "Goal"),
        ]
        for args, expected_code, marker in smoke_cases:
            with self.subTest(args=args):
                result = self.run_installed_jini(binary_path, *args, env=command_env)
                self.assertEqual(
                    expected_code,
                    result.returncode,
                    msg=f"Expected {' '.join(args)} to exit {expected_code}.\nSTDOUT:\n{result.stdout}\nSTDERR:\n{result.stderr}",
                )
                self.assertIn(
                    marker,
                    result.stdout,
                    msg=f"Expected {' '.join(args)} to include {marker!r}.\nSTDOUT:\n{result.stdout}\nSTDERR:\n{result.stderr}",
                )
                self.assertNotIn(
                    "Unknown command",
                    result.stderr,
                    msg=f"Installed artifact rejected {' '.join(args)}.\nSTDOUT:\n{result.stdout}\nSTDERR:\n{result.stderr}",
                )

    def assert_python_public_command_contract(self, binary_path: Path, *, env: Optional[dict[str, str]] = None) -> None:
        command_env = {"JINI_PROVIDER": "local-preview"}
        if env:
            command_env.update(env)

        smoke_cases = [
            (["--help"], 0, "Jini CLI"),
            (["commands"], 0, "Public command inventory"),
            (["help", "--all"], 0, "Public command inventory"),
            (["admin", "help"], 0, "Admin and developer command inventory"),
            (["doctor"], 0, "Provider"),
            (["status", str(RESEARCH_EXAMPLE)], 0, "WORK   research-prd-v1"),
        ]
        for args, expected_code, marker in smoke_cases:
            with self.subTest(args=args):
                result = self.run_installed_jini(binary_path, *args, env=command_env)
                self.assertEqual(
                    expected_code,
                    result.returncode,
                    msg=f"Expected {' '.join(args)} to exit {expected_code}.\nSTDOUT:\n{result.stdout}\nSTDERR:\n{result.stderr}",
                )
                self.assertIn(
                    marker,
                    result.stdout,
                    msg=f"Expected {' '.join(args)} to include {marker!r}.\nSTDOUT:\n{result.stdout}\nSTDERR:\n{result.stderr}",
                )
                self.assertNotIn(
                    "Unknown command",
                    result.stderr,
                    msg=f"Installed artifact rejected {' '.join(args)}.\nSTDOUT:\n{result.stdout}\nSTDERR:\n{result.stderr}",
                )

    def create_remote_snapshot(self) -> Path:
        snapshot = self.tmp / "snapshot-repo"
        shutil.copytree(
            REPO_ROOT,
            snapshot,
            ignore=shutil.ignore_patterns(".git", ".gocache", ".jini", "__pycache__"),
        )
        subprocess.run(["git", "init", "-b", "main"], cwd=snapshot, check=True, capture_output=True, text=True)
        subprocess.run(["git", "add", "."], cwd=snapshot, check=True, capture_output=True, text=True)
        subprocess.run(
            ["git", "-c", "user.name=Jini Tests", "-c", "user.email=jini-tests@example.com", "commit", "-m", "snapshot"],
            cwd=snapshot,
            check=True,
            capture_output=True,
            text=True,
        )
        return snapshot

    def test_local_install_creates_jini_command_and_receipt(self) -> None:
        bin_dir = self.tmp / "bin"
        install_dir = self.tmp / "share" / "jini"
        result = self.run_installer(
            "--source-dir",
            str(REPO_ROOT),
            "--bin-dir",
            str(bin_dir),
            "--install-dir",
            str(install_dir),
            "--force",
            env=self.go_ready_env(),
        )
        self.assert_ok(result)
        self.assertIn("Installed Jini", result.stdout)
        self.assertTrue((bin_dir / "jini").exists())
        self.assertTrue((install_dir / "install-receipt.txt").exists())

        launch = subprocess.run(
            [str(bin_dir / "jini"), "doctor"],
            cwd=REPO_ROOT,
            text=True,
            capture_output=True,
            env=self.go_ready_env(),
        )
        self.assertEqual(0, launch.returncode, msg=launch.stderr)
        self.assertIn("Provider", launch.stdout)

    def test_local_go_install_supports_taught_public_command_contract(self) -> None:
        bin_dir = self.tmp / "go-bin"
        install_dir = self.tmp / "go-share" / "jini"
        result = self.run_installer(
            "--source-dir",
            str(REPO_ROOT),
            "--bin-dir",
            str(bin_dir),
            "--install-dir",
            str(install_dir),
            "--force",
            env=self.go_ready_env(),
        )
        self.assert_ok(result)
        self.assert_go_public_command_contract(bin_dir / "jini")

    def test_local_go_install_keeps_provider_doctor_compatibility_alias(self) -> None:
        bin_dir = self.tmp / "go-provider-bin"
        install_dir = self.tmp / "go-provider-share" / "jini"
        result = self.run_installer(
            "--source-dir",
            str(REPO_ROOT),
            "--bin-dir",
            str(bin_dir),
            "--install-dir",
            str(install_dir),
            "--force",
            env=self.go_ready_env(),
        )
        self.assert_ok(result)
        provider = self.run_installed_jini(
            bin_dir / "jini",
            "provider",
            "doctor",
            env={"JINI_PROVIDER": "local-preview"},
        )
        self.assertEqual(0, provider.returncode, msg=provider.stderr)
        self.assertIn("Provider", provider.stdout)

    def test_python_fallback_install_supports_taught_public_command_contract(self) -> None:
        bin_dir = self.tmp / "python-bin"
        install_dir = self.tmp / "python-share" / "jini"
        result = self.run_installer(
            "--source-dir",
            str(REPO_ROOT),
            "--bin-dir",
            str(bin_dir),
            "--install-dir",
            str(install_dir),
            "--force",
            env={"PATH": "/usr/bin:/bin"},
        )
        self.assert_ok(result)
        self.assert_python_public_command_contract(bin_dir / "jini")

    def test_cli_guide_taught_support_commands_match_installed_smoke_surface(self) -> None:
        text = CLI_DOC.read_text(encoding="utf-8")
        taught = set()
        for line in text.splitlines():
            if "<h3><code>jini " in line:
                taught.add(line.split("<h3><code>jini ", 1)[1].split("</code></h3>", 1)[0].strip())
        self.assertEqual({"status", "doctor"}, taught)

    def test_remote_style_install_works_from_outside_repo(self) -> None:
        remote_snapshot = self.create_remote_snapshot()
        remote_root = self.tmp / "remote"
        remote_root.mkdir()
        remote_installer = remote_root / "install.sh"
        shutil.copy2(INSTALLER, remote_installer)

        result = self.run_installer(
            "--repo-url",
            f"file://{remote_snapshot}",
            "--repo-ref",
            "main",
            "--bin-dir",
            str(remote_root / "bin"),
            "--install-dir",
            str(remote_root / "share" / "jini"),
            "--force",
            env=self.go_ready_env(),
            cwd=remote_root,
            installer_path=remote_installer,
        )
        self.assert_ok(result)
        self.assertIn("Installed Jini", result.stdout)

    def test_remote_style_install_without_go_uses_python_fallback(self) -> None:
        remote_snapshot = self.create_remote_snapshot()
        remote_root = self.tmp / "remote-python"
        remote_root.mkdir()
        remote_installer = remote_root / "install.sh"
        shutil.copy2(INSTALLER, remote_installer)

        result = self.run_installer(
            "--repo-url",
            f"file://{remote_snapshot}",
            "--repo-ref",
            "main",
            "--bin-dir",
            str(remote_root / "bin"),
            "--install-dir",
            str(remote_root / "share" / "jini"),
            "--force",
            env={"PATH": "/usr/bin:/bin"},
            cwd=remote_root,
            installer_path=remote_installer,
        )
        self.assert_ok(result)
        self.assertIn("Installed Jini", result.stdout)

    def test_piped_install_from_stdin_works_from_outside_repo(self) -> None:
        remote_snapshot = self.create_remote_snapshot()
        remote_root = self.tmp / "remote-pipe"
        remote_root.mkdir()

        result = self.run_installer_via_stdin(
            "--repo-url",
            f"file://{remote_snapshot}",
            "--repo-ref",
            "main",
            "--bin-dir",
            str(remote_root / "bin"),
            "--install-dir",
            str(remote_root / "share" / "jini"),
            "--force",
            env=self.go_ready_env(),
            cwd=remote_root,
        )
        self.assert_ok(result)
        self.assertIn("Installed Jini", result.stdout)

    def test_piped_install_without_go_fails_cleanly_not_with_bash_source(self) -> None:
        remote_root = self.tmp / "remote-pipe-no-go"
        remote_root.mkdir()

        result = self.run_installer_via_stdin(
            "--repo-url",
            "file:///does/not/exist",
            "--repo-ref",
            "main",
            "--bin-dir",
            str(remote_root / "bin"),
            "--install-dir",
            str(remote_root / "share" / "jini"),
            "--force",
            env={"PATH": "/usr/bin:/bin"},
            cwd=remote_root,
        )
        self.assert_error(result)
        self.assertNotIn("BASH_SOURCE[0]: unbound variable", result.stderr)

    def test_missing_go_uses_python_source_fallback(self) -> None:
        result = self.run_installer(
            "--source-dir",
            str(REPO_ROOT),
            "--bin-dir",
            str(self.tmp / "bin"),
            "--install-dir",
            str(self.tmp / "share" / "jini"),
            "--force",
            env={"PATH": "/usr/bin:/bin"},
        )
        self.assert_ok(result)
        self.assertIn("Installed Jini", result.stdout)

    def test_missing_python_and_go_reports_clear_error(self) -> None:
        fake_bin = self.tmp / "fake-no-runtime"
        fake_bin.mkdir()
        for name in ("python3", "go"):
            fake = fake_bin / name
            fake.write_text("#!/bin/sh\nexit 127\n", encoding="utf-8")
            fake.chmod(0o755)
        result = self.run_installer(
            "--source-dir",
            str(REPO_ROOT),
            "--bin-dir",
            str(self.tmp / "bin"),
            "--install-dir",
            str(self.tmp / "share" / "jini"),
            "--force",
            env={"PATH": f"{fake_bin}:/usr/bin:/bin"},
        )
        self.assert_error(result)
        self.assertIn(
            "source fallback needs either Python 3 (Jini will try to add PyYAML automatically) or Go",
            result.stderr,
        )

    def test_missing_git_reports_clear_error_for_remote_install(self) -> None:
        remote_root = self.tmp / "remote-no-git"
        remote_root.mkdir()
        remote_installer = remote_root / "install.sh"
        shutil.copy2(INSTALLER, remote_installer)
        fake_bin = self.tmp / "fake-bin"
        fake_bin.mkdir()
        fake_git = fake_bin / "git"
        fake_git.write_text("#!/bin/sh\nexit 127\n", encoding="utf-8")
        fake_git.chmod(0o755)

        result = self.run_installer(
            "--repo-url",
            "file:///does/not/matter",
            "--repo-ref",
            "main",
            "--bin-dir",
            str(remote_root / "bin"),
            "--install-dir",
            str(remote_root / "share" / "jini"),
            "--force",
            env={"PATH": f"{fake_bin}:{LOCAL_GO_BIN}:/usr/bin:/bin"},
            cwd=remote_root,
            installer_path=remote_installer,
        )
        self.assert_error(result)
        self.assertIn("Git is required when installing outside the Jini repo.", result.stderr)

    def test_existing_regular_command_requires_force(self) -> None:
        bin_dir = self.tmp / "bin"
        install_dir = self.tmp / "share" / "jini"
        bin_dir.mkdir(parents=True)
        command_path = bin_dir / "jini"
        command_path.write_text("old command\n", encoding="utf-8")

        result = self.run_installer(
            "--source-dir",
            str(REPO_ROOT),
            "--bin-dir",
            str(bin_dir),
            "--install-dir",
            str(install_dir),
            env=self.go_ready_env(),
        )
        self.assert_error(result)
        self.assertIn("already exists. Rerun with --force", result.stderr)

    def test_force_replaces_existing_command(self) -> None:
        bin_dir = self.tmp / "bin"
        install_dir = self.tmp / "share" / "jini"
        bin_dir.mkdir(parents=True)
        command_path = bin_dir / "jini"
        command_path.write_text("old command\n", encoding="utf-8")

        result = self.run_installer(
            "--source-dir",
            str(REPO_ROOT),
            "--bin-dir",
            str(bin_dir),
            "--install-dir",
            str(install_dir),
            "--force",
            env=self.go_ready_env(),
        )
        self.assert_ok(result)
        self.assertTrue(command_path.exists())
        self.assertNotEqual("old command\n", command_path.read_text(encoding="utf-8", errors="ignore"))

    def test_copy_mode_writes_plain_binary(self) -> None:
        bin_dir = self.tmp / "bin"
        install_dir = self.tmp / "share" / "jini"
        result = self.run_installer(
            "--source-dir",
            str(REPO_ROOT),
            "--bin-dir",
            str(bin_dir),
            "--install-dir",
            str(install_dir),
            "--copy",
            "--force",
            env=self.go_ready_env(),
        )
        self.assert_ok(result)
        self.assertFalse((bin_dir / "jini").is_symlink())

    def test_install_with_spaces_in_paths_succeeds(self) -> None:
        root = self.tmp / "path with spaces"
        bin_dir = root / "bin folder"
        install_dir = root / "share folder" / "jini"
        result = self.run_installer(
            "--source-dir",
            str(REPO_ROOT),
            "--bin-dir",
            str(bin_dir),
            "--install-dir",
            str(install_dir),
            "--force",
            env=self.go_ready_env(),
        )
        self.assert_ok(result)
        self.assertTrue((bin_dir / "jini").exists())

    def test_non_writable_bin_dir_reports_clear_error(self) -> None:
        bin_dir = self.tmp / "bin"
        install_dir = self.tmp / "share" / "jini"
        bin_dir.mkdir(parents=True)
        bin_dir.chmod(0o555)

        result = self.run_installer(
            "--source-dir",
            str(REPO_ROOT),
            "--bin-dir",
            str(bin_dir),
            "--install-dir",
            str(install_dir),
            "--force",
            env=self.go_ready_env(),
        )
        self.assert_error(result)
        self.assertIn("Bin directory is not writable", result.stderr)

    def test_non_writable_install_dir_reports_clear_error(self) -> None:
        bin_dir = self.tmp / "bin"
        install_dir = self.tmp / "share" / "jini"
        install_dir.mkdir(parents=True)
        install_dir.chmod(0o555)

        result = self.run_installer(
            "--source-dir",
            str(REPO_ROOT),
            "--bin-dir",
            str(bin_dir),
            "--install-dir",
            str(install_dir),
            "--force",
            env=self.go_ready_env(),
        )
        self.assert_error(result)
        self.assertIn("Install directory is not writable", result.stderr)

    def test_path_guidance_is_printed_when_bin_dir_not_on_path(self) -> None:
        bin_dir = self.tmp / "bin"
        install_dir = self.tmp / "share" / "jini"
        result = self.run_installer(
            "--source-dir",
            str(REPO_ROOT),
            "--bin-dir",
            str(bin_dir),
            "--install-dir",
            str(install_dir),
            "--force",
            env=self.go_ready_env(),
        )
        self.assert_ok(result)
        self.assertIn('export PATH="', result.stdout)

    def test_run_hint_is_printed_when_bin_dir_is_already_on_path(self) -> None:
        bin_dir = self.tmp / "bin"
        install_dir = self.tmp / "share" / "jini"
        env = self.go_ready_env(extra_path=str(bin_dir))
        result = self.run_installer(
            "--source-dir",
            str(REPO_ROOT),
            "--bin-dir",
            str(bin_dir),
            "--install-dir",
            str(install_dir),
            "--force",
            env=env,
        )
        self.assert_ok(result)
        self.assertIn("Run:\n  jini", result.stdout)


if __name__ == "__main__":
    unittest.main()
