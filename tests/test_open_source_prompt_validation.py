import json
import os
import subprocess
import sys
import tempfile
import unittest
from pathlib import Path

REPO_ROOT = Path(__file__).resolve().parents[1]
if str(REPO_ROOT) not in sys.path:
    sys.path.insert(0, str(REPO_ROOT))

from tools.yaml_compat import safe_load


CLI = [sys.executable, str(REPO_ROOT / "tools" / "jini.py")]
CORPUS_PATH = REPO_ROOT / "specs" / "open-source-prompt-validation.yaml"


class OpenSourcePromptValidationTests(unittest.TestCase):
    maxDiff = None

    def setUp(self) -> None:
        self.temp_dir = tempfile.TemporaryDirectory(prefix="jini-open-source-prompts-")
        self.tmp = Path(self.temp_dir.name)

    def tearDown(self) -> None:
        self.temp_dir.cleanup()

    def run_cli(self, *args: object) -> subprocess.CompletedProcess[str]:
        env = dict(os.environ)
        env["JINI_STATE_DIR"] = str((self.tmp / ".jini").resolve())
        return subprocess.run(
            [*CLI, *[str(arg) for arg in args]],
            cwd=REPO_ROOT,
            text=True,
            capture_output=True,
            env=env,
        )

    def assert_ok(self, result: subprocess.CompletedProcess[str]) -> None:
        if result.returncode != 0:
            self.fail(
                f"Expected command to succeed.\nSTDOUT:\n{result.stdout}\nSTDERR:\n{result.stderr}"
            )

    def load_corpus(self) -> dict:
        self.assertTrue(CORPUS_PATH.exists(), "Open-source prompt validation corpus is missing")
        with CORPUS_PATH.open(encoding="utf-8") as handle:
            return safe_load(handle.read())

    def test_corpus_tracks_permissive_github_sources(self) -> None:
        corpus = self.load_corpus()
        sources = corpus.get("sources", [])
        self.assertGreaterEqual(len(sources), 3)

        source_ids = {source["id"] for source in sources}
        self.assertIn("prompts-chat", source_ids)
        self.assertIn("promptsource", source_ids)
        self.assertIn("promptfoo", source_ids)

        for source in sources:
            self.assertRegex(source.get("repository", ""), r"^https://github\.com/")
            self.assertIn(source.get("license"), {"Apache-2.0", "CC0-1.0", "MIT"})
            self.assertTrue(source.get("consumed_as"))
            self.assertTrue(source.get("adaptation_rule"))

    def test_prompt_classes_are_derived_and_assertable(self) -> None:
        corpus = self.load_corpus()
        source_ids = {source["id"] for source in corpus.get("sources", [])}
        prompt_classes = corpus.get("prompt_classes", [])
        self.assertGreaterEqual(len(prompt_classes), 6)

        for prompt_class in prompt_classes:
            with self.subTest(prompt_class=prompt_class.get("id")):
                self.assertTrue(prompt_class.get("id"))
                self.assertTrue(prompt_class.get("source_ids"))
                self.assertTrue(set(prompt_class["source_ids"]).issubset(source_ids))
                self.assertIs(prompt_class.get("derived_prompt"), True)
                self.assertIs(prompt_class.get("verbatim_source_prompt"), False)
                self.assertTrue(prompt_class.get("prompt"))
                self.assertTrue(prompt_class.get("expected_behavior"))
                self.assertTrue(prompt_class.get("comparison_signals"))
                self.assertNotIn("expected_output", prompt_class)

    def test_validation_cli_reports_source_and_prompt_gate(self) -> None:
        result = self.run_cli("validate-open-source-prompts", "--format", "json")
        self.assert_ok(result)
        report = json.loads(result.stdout)
        self.assertEqual(report["result_type"], "JiniOpenSourcePromptValidation")
        self.assertEqual(report["status"], "ok")
        self.assertGreaterEqual(report["source_count"], 3)
        self.assertGreaterEqual(report["prompt_class_count"], 6)
        self.assertEqual(report["failed_checks"], [])
        self.assertEqual(report["policy"]["use_derived_prompts_only"], True)
        self.assertIn("prompts-chat", report["source_ids"])
        self.assertIn("promptsource", report["source_ids"])
        self.assertIn("promptfoo", report["source_ids"])
