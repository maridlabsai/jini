import tempfile
import unittest
from pathlib import Path
from unittest import mock

import tools.jini as go_launcher


class GoFirstLauncherTests(unittest.TestCase):
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


if __name__ == "__main__":
    unittest.main()
