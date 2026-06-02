import io
import json
import os
import shutil
import subprocess
import sys
import tempfile
import unittest
from contextlib import redirect_stderr, redirect_stdout
from pathlib import Path
from typing import Optional, Union
from unittest import mock

REPO_ROOT = Path(__file__).resolve().parents[1]
if str(REPO_ROOT) not in sys.path:
    sys.path.insert(0, str(REPO_ROOT))

from tools.yaml_compat import safe_load
from tools.help_tail_contract import HELP_TAIL_EXAMPLE_REQUEST, help_tail_message_lines
import tools.jini_validate as jini_validate


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
        return self.run_cli_in_cwd(REPO_ROOT, *args, env=env)

    def run_cli_in_cwd(
        self,
        cwd: Path,
        *args: object,
        env: Optional[dict[str, str]] = None,
        input_text: Optional[str] = None,
    ) -> subprocess.CompletedProcess[str]:
        run_env = dict(os.environ if env is None else env)
        run_env["JINI_STATE_DIR"] = str((self.tmp / ".jini").resolve())
        return subprocess.run(
            [*CLI, *[str(arg) for arg in args]],
            cwd=cwd,
            text=True,
            capture_output=True,
            env=run_env,
            input=input_text,
        )

    def assert_ok(self, result: subprocess.CompletedProcess[str]) -> None:
        if result.returncode != 0:
            self.fail(
                f"Expected command to succeed.\nSTDOUT:\n{result.stdout}\nSTDERR:\n{result.stderr}"
            )

    def assert_error(self, result: subprocess.CompletedProcess[str]) -> None:
        if result.returncode == 0:
            self.fail(f"Expected command to fail.\nSTDOUT:\n{result.stdout}")

    def assert_compact_interactive_escape_hatch_surface(self, stdout: str) -> None:
        self.assertIn("jini> ", stdout)
        self.assertIn("IN-SHELL  commands  doctor  help --admin  exit", stdout)
        self.assertIn("SUPPORT   run `jini status`, `jini continue`, or `jini open` outside the shell", stdout)
        self.assertIn("DOCTOR   Local preview [ok]", stdout)
        self.assertIn("ADMIN    validate  publish  route-feedback  skills  delegate", stdout)
        self.assertIn("MORE     run `jini help --admin` for the full inventory", stdout)
        self.assertIn("SETUP   run `jini setup --harness codex` if you need a fixed route", stdout)
        self.assertNotIn("Public command inventory", stdout)
        self.assertNotIn("Admin and developer command inventory", stdout)
        self.assertNotIn("START HERE", stdout)
        self.assertNotIn("Jini CLI 0.1.0", stdout)
        self.assertNotIn("Session closed.", stdout)

    def assert_interactive_single_token_recovery_hints(self, result: subprocess.CompletedProcess[str]) -> None:
        self.assert_ok(result)
        self.assertIn("Working on: fix failing tests", result.stdout)
        self.assertNotIn("Working on: status", result.stdout)
        self.assertNotIn("Working on: continue", result.stdout)
        self.assertNotIn("Working on: open", result.stdout)
        self.assertNotIn("Working on: stats", result.stdout)
        self.assertIn("Use `jini status` outside the live shell.", result.stderr)
        self.assertIn("Use `jini continue` outside the live shell.", result.stderr)
        self.assertIn("Use `jini open` outside the live shell.", result.stderr)
        self.assertIn('ERROR Unknown command "stats".', result.stderr)
        self.assertNotIn("usage: jini", result.stderr)

    def assert_interactive_multi_token_recovery_hints(self, result: subprocess.CompletedProcess[str]) -> None:
        self.assert_ok(result)
        self.assertIn("Working on: fix failing tests", result.stdout)
        self.assertNotIn("Working on: status /tmp/work", result.stdout)
        self.assertNotIn("Working on: continue --print-path", result.stdout)
        self.assertNotIn("Working on: open prd", result.stdout)
        self.assertNotIn("Working on: doctor --format json", result.stdout)
        self.assertNotIn("Working on: commands please", result.stdout)
        self.assertNotIn("Working on: help --admin please", result.stdout)
        self.assertIn("Use `jini status /tmp/work` outside the live shell.", result.stderr)
        self.assertIn("Use `jini continue --print-path` outside the live shell.", result.stderr)
        self.assertIn("Use `jini open prd` outside the live shell.", result.stderr)
        self.assertIn("Use `jini doctor --format json` outside the live shell.", result.stderr)
        self.assertIn("Use `commands` by itself inside the live shell.", result.stderr)
        self.assertIn("Use `help --admin` by itself inside the live shell.", result.stderr)
        self.assertNotIn("usage: jini", result.stderr)

    def assert_prefixed_interactive_escape_hatch_surface(self, result: subprocess.CompletedProcess[str]) -> None:
        self.assert_ok(result)
        self.assert_compact_interactive_escape_hatch_surface(result.stdout)
        self.assertEqual("", result.stderr)

    def assert_prefixed_interactive_support_recovery_hints(self, result: subprocess.CompletedProcess[str]) -> None:
        self.assert_ok(result)
        self.assertIn("Working on: fix failing tests", result.stdout)
        self.assertNotIn("Working on: jini status /tmp/work", result.stdout)
        self.assertNotIn("Working on: jini continue --print-path", result.stdout)
        self.assertNotIn("Working on: jini open prd", result.stdout)
        self.assertNotIn("Working on: jini doctor --format json", result.stdout)
        self.assertIn("Use `jini status /tmp/work` outside the live shell.", result.stderr)
        self.assertIn("Use `jini continue --print-path` outside the live shell.", result.stderr)
        self.assertIn("Use `jini open prd` outside the live shell.", result.stderr)
        self.assertIn("Use `jini doctor --format json` outside the live shell.", result.stderr)
        self.assertNotIn("usage: jini", result.stderr)

    def assert_interactive_artifact_and_path_recovery_hints(self, result: subprocess.CompletedProcess[str]) -> None:
        self.assert_ok(result)
        self.assertIn("Working on: fix failing tests", result.stdout)
        self.assertNotIn("Working on: tasks", result.stdout)
        self.assertNotIn("Working on: prd", result.stdout)
        self.assertNotIn("Working on: ./notes.md", result.stdout)
        self.assertIn("Use `jini open tasks` outside the live shell.", result.stderr)
        self.assertIn("Use `jini open prd` outside the live shell.", result.stderr)
        self.assertIn("Open `./notes.md` from your shell or editor, not as a live-shell task.", result.stderr)
        self.assertIn("Use `jini open` outside the live shell for Jini artifacts.", result.stderr)
        self.assertNotIn("usage: jini", result.stderr)

    def assert_interactive_numeric_and_filename_recovery_hints(self, result: subprocess.CompletedProcess[str]) -> None:
        self.assert_ok(result)
        self.assertIn("Working on: fix failing tests", result.stdout)
        self.assertNotIn("Working on: 1", result.stdout)
        self.assertNotIn("Working on: 2", result.stdout)
        self.assertNotIn("Working on: notes.md", result.stdout)
        self.assertIn("Use `jini open 1` outside the live shell.", result.stderr)
        self.assertIn("Use `jini open 2` outside the live shell.", result.stderr)
        self.assertIn("Open `notes.md` from your shell or editor, not as a live-shell task.", result.stderr)
        self.assertIn("Use `jini open` outside the live shell for Jini artifacts.", result.stderr)
        self.assertNotIn("usage: jini", result.stderr)

    def assert_interactive_existing_file_recovery_hints(self, result: subprocess.CompletedProcess[str]) -> None:
        self.assert_ok(result)
        self.assertIn("Working on: fix failing tests", result.stdout)
        self.assertNotIn("Working on: Makefile", result.stdout)
        self.assertNotIn("Working on: .env", result.stdout)
        self.assertIn("Open `Makefile` from your shell or editor, not as a live-shell task.", result.stderr)
        self.assertIn("Open `.env` from your shell or editor, not as a live-shell task.", result.stderr)
        self.assertIn("Use `jini open` outside the live shell for Jini artifacts.", result.stderr)
        self.assertNotIn("usage: jini", result.stderr)

    def assert_prefixed_interactive_file_shorthand_recovery_hints(self, result: subprocess.CompletedProcess[str]) -> None:
        self.assert_ok(result)
        self.assertIn("Working on: fix failing tests", result.stdout)
        self.assertNotIn("Working on: jini Makefile", result.stdout)
        self.assertIn("Open `Makefile` from your shell or editor, not as a live-shell task.", result.stderr)
        self.assertIn("Use `jini open` outside the live shell for Jini artifacts.", result.stderr)
        self.assertNotIn("usage: jini", result.stderr)

    def assert_prefixed_interactive_artifact_and_file_shorthand_recovery_hints(self, result: subprocess.CompletedProcess[str]) -> None:
        self.assert_prefixed_interactive_file_shorthand_recovery_hints(result)
        self.assertNotIn("Working on: jini tasks", result.stdout)
        self.assertNotIn("Working on: jini 1", result.stdout)
        self.assertIn("Use `jini open tasks` outside the live shell.", result.stderr)
        self.assertIn("Use `jini open 1` outside the live shell.", result.stderr)

    def assert_prefixed_interactive_path_and_filename_recovery_hints(self, result: subprocess.CompletedProcess[str]) -> None:
        self.assert_ok(result)
        self.assertIn("Working on: fix failing tests", result.stdout)
        self.assertNotIn("Working on: jini ./notes.md", result.stdout)
        self.assertNotIn("Working on: jini notes.md", result.stdout)
        self.assertNotIn("Working on: jini .env", result.stdout)
        self.assertIn("Open `./notes.md` from your shell or editor, not as a live-shell task.", result.stderr)
        self.assertIn("Open `notes.md` from your shell or editor, not as a live-shell task.", result.stderr)
        self.assertIn("Open `.env` from your shell or editor, not as a live-shell task.", result.stderr)
        self.assertIn("Use `jini open` outside the live shell for Jini artifacts.", result.stderr)
        self.assertNotIn("usage: jini", result.stderr)

    def assert_interactive_help_alias_surface(self, result: subprocess.CompletedProcess[str]) -> None:
        self.assert_ok(result)
        self.assertIn("jini> ", result.stdout)
        self.assertIn("IN-SHELL  commands  doctor  help --admin  exit", result.stdout)
        self.assertIn("SUPPORT   run `jini status`, `jini continue`, or `jini open` outside the shell", result.stdout)
        self.assertNotIn("DOCTOR   Local preview [ok]", result.stdout)
        self.assertNotIn("ADMIN    validate  publish  route-feedback  skills  delegate", result.stdout)
        self.assertNotIn("SETUP   run `jini setup --harness codex` if you need a fixed route", result.stdout)
        self.assertNotIn("Working on: help", result.stdout)
        self.assertNotIn("Working on: --help", result.stdout)
        self.assertNotIn("Working on: -h", result.stdout)
        self.assertNotIn("Working on: jini help", result.stdout)
        self.assertNotIn("Working on: jini --help", result.stdout)
        self.assertNotIn("Working on: jini -h", result.stdout)
        self.assertEqual("", result.stderr)

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

    def write_skill(
        self,
        root: Path,
        skill_id: str,
        *,
        label: Optional[str] = None,
        purpose: Optional[str] = None,
        enabled: bool = True,
        invalid_missing_output_contract: bool = False,
    ) -> Path:
        skill_dir = root / skill_id
        skill_dir.mkdir(parents=True, exist_ok=True)
        manifest = {
            "schema_version": "0.1.0",
            "skill_id": skill_id,
            "label": label or skill_id.replace("-", " ").title(),
            "purpose": purpose or f"Use {skill_id} inside the current environment.",
            "when_to_use": [f"Use {skill_id} when the current work benefits from that specialist lens."],
            "allowed_tools": ["read", "search", "write"],
            "input_contract": {"current_work": "required", "artifact_focus": "optional"},
            "prompt_file": "prompt.md",
            "enabled": enabled,
        }
        if not invalid_missing_output_contract:
            manifest["output_contract"] = {"status_values": ["staged"], "primary_artifact": "summary"}
        (skill_dir / "skill.yaml").write_text(json.dumps(manifest, indent=2) + "\n", encoding="utf-8")
        (skill_dir / "prompt.md").write_text(
            f"# {manifest['label']}\n\nStay attached to the same Jini work and improve the result.\n",
            encoding="utf-8",
        )
        return skill_dir

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

    def create_fake_gh_cli(self) -> Path:
        gh = self.tmp / "gh"
        gh.write_text(
            "\n".join(
                [
                    "#!/usr/bin/env python3",
                    "import json",
                    "import pathlib",
                    "import sys",
                    "",
                    "state_path = pathlib.Path(__file__).with_name('gh-state.json')",
                    "if state_path.exists():",
                    "    state = json.loads(state_path.read_text(encoding='utf-8'))",
                    "else:",
                    "    state = {'repos': {}}",
                    "",
                    "def save():",
                    "    state_path.write_text(json.dumps(state, indent=2) + '\\n', encoding='utf-8')",
                    "",
                    "def parse_flag(args, flag, default=''):",
                    "    if flag not in args:",
                    "        return default",
                    "    return args[args.index(flag) + 1]",
                    "",
                    "def parse_many(args, flag):",
                    "    values = []",
                    "    index = 0",
                    "    while index < len(args):",
                    "        if args[index] == flag and index + 1 < len(args):",
                    "            values.append(args[index + 1])",
                    "            index += 2",
                    "            continue",
                    "        index += 1",
                    "    return values",
                    "",
                    "args = sys.argv[1:]",
                    "def content_key(endpoint):",
                    "    if not endpoint.startswith('repos/') or '/contents/' not in endpoint:",
                    "        raise SystemExit(f'unsupported endpoint: {endpoint}')",
                    "    repo_part, path = endpoint.removeprefix('repos/').split('/contents/', 1)",
                    "    owner, repo, *_rest = repo_part.split('/')",
                    "    return f'{owner}/{repo}', path",
                    "",
                    "if args and args[0] == 'api':",
                    "    endpoint = args[1]",
                    "    repo, path = content_key(endpoint)",
                    "    method = parse_flag(args, '-X', 'GET')",
                    "    repo_state = state['repos'].setdefault(repo, [])",
                    "    docs = state.setdefault('docs', {}).setdefault(repo, {})",
                    "    if method == 'GET':",
                    "        if path not in docs:",
                    "            raise SystemExit('Not Found')",
                    "        record = docs[path]",
                    "        print(json.dumps({'path': path, 'sha': record['sha'], 'html_url': record['html_url']}))",
                    "        raise SystemExit(0)",
                    "    if method == 'PUT':",
                    "        input_path = pathlib.Path(parse_flag(args, '--input'))",
                    "        payload = json.loads(input_path.read_text(encoding='utf-8'))",
                    "        status = 'updated' if path in docs else 'created'",
                    "        sha = f'sha-{len(docs) + 1}' if status == 'created' else docs[path]['sha']",
                    "        docs[path] = {",
                    "            'sha': sha,",
                    "            'html_url': f'https://github.com/{repo}/blob/main/{path}',",
                    "            'message': payload.get('message', ''),",
                    "            'content': payload.get('content', ''),",
                    "        }",
                    "        save()",
                    "        print(json.dumps({'content': {'path': path, 'sha': sha, 'html_url': docs[path]['html_url']}, 'commit': {'sha': f'commit-{len(docs)}'}, 'status': status}))",
                    "        raise SystemExit(0)",
                    "    raise SystemExit(f'unsupported method: {method}')",
                    "",
                    "if args[:2] != ['issue', 'list'] and args[:2] != ['issue', 'create'] and args[:2] != ['issue', 'edit']:",
                    "    raise SystemExit(f'unsupported args: {args}')",
                    "",
                    "if args[:2] == ['issue', 'list']:",
                    "    repo = parse_flag(args, '--repo')",
                    "    label = parse_flag(args, '--label')",
                    "    repo_state = state['repos'].get(repo, [])",
                    "    rows = [",
                    "        {'number': issue['number'], 'url': issue['url'], 'title': issue['title']}",
                    "        for issue in repo_state",
                    "        if label in issue['labels']",
                    "    ]",
                    "    print(json.dumps(rows))",
                    "    raise SystemExit(0)",
                    "",
                    "if args[:2] == ['issue', 'create']:",
                    "    repo = parse_flag(args, '--repo')",
                    "    title = parse_flag(args, '--title')",
                    "    body_path = pathlib.Path(parse_flag(args, '--body-file'))",
                    "    labels = parse_many(args, '--label')",
                    "    repo_state = state['repos'].setdefault(repo, [])",
                    "    number = max((issue['number'] for issue in repo_state), default=0) + 1",
                    "    issue = {",
                    "        'number': number,",
                    "        'url': f'https://github.com/{repo}/issues/{number}',",
                    "        'title': title,",
                    "        'body': body_path.read_text(encoding='utf-8'),",
                    "        'labels': labels,",
                    "    }",
                    "    repo_state.append(issue)",
                    "    save()",
                    "    print(issue['url'])",
                    "    raise SystemExit(0)",
                    "",
                    "repo = parse_flag(args, '--repo')",
                    "number = int(args[2])",
                    "title = parse_flag(args, '--title')",
                    "body_path = pathlib.Path(parse_flag(args, '--body-file'))",
                    "labels = parse_many(args, '--add-label')",
                    "repo_state = state['repos'].setdefault(repo, [])",
                    "for issue in repo_state:",
                    "    if issue['number'] == number:",
                    "        issue['title'] = title",
                    "        issue['body'] = body_path.read_text(encoding='utf-8')",
                    "        issue['labels'] = sorted(set(issue['labels']) | set(labels))",
                    "        save()",
                    "        print(issue['url'])",
                    "        raise SystemExit(0)",
                    "raise SystemExit('unknown issue number')",
                ]
            )
            + "\n",
            encoding="utf-8",
        )
        gh.chmod(0o755)
        return gh

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
        self.assertIn("TYPE THE NEXT THING", result.stdout)
        self.assertIn("fix failing tests", result.stdout)
        self.assertIn("review this repo", result.stdout)
        self.assertIn("what is blocked?", result.stdout)
        self.assertIn("open the latest artifact", result.stdout)
        self.assertIn("SUPPORT COMMANDS", result.stdout)
        self.assertIn("jini status", result.stdout)
        self.assertIn("jini continue", result.stdout)
        self.assertIn("jini open", result.stdout)
        self.assertIn("jini doctor", result.stdout)
        self.assertIn("MORE", result.stdout)
        self.assertIn("jini commands", result.stdout)
        self.assertIn("jini help --admin", result.stdout)
        self.assertNotIn("IN THE SHELL", result.stdout)
        self.assertNotIn("\n  Continue\n", result.stdout)
        self.assertNotIn("\n  Missing\n", result.stdout)
        self.assertNotIn("\n  Plan\n", result.stdout)
        self.assertNotIn("jini setup --harness codex", result.stdout)
        self.assertNotIn("jini run --repo /path/to/repo --harness codex", result.stdout)
        self.assertNotIn("usage: jini", result.stdout)

    def test_zero_arg_cli_shows_repo_aware_start_surface_inside_repo(self) -> None:
        repo = self.create_repo_fixture()
        result = self.run_cli_in_cwd(repo)
        self.assert_ok(result)
        self.assertIn("Repo: sample-repo", result.stdout)
        self.assertNotIn("Nothing is in progress yet.", result.stdout)
        self.assertNotIn("Try one of these:", result.stdout)
        self.assertIn("jini review this repo", result.stdout)
        self.assertIn("jini fix failing tests", result.stdout)
        self.assertIn("jini review this branch", result.stdout)
        self.assertIn("Review the repo and suggest the next move.", result.stdout)
        self.assertIn("Fix the failing tests in this repo.", result.stdout)
        self.assertIn("Review the current branch and call out risks.", result.stdout)
        self.assertNotIn("jini plan this change", result.stdout)
        self.assertNotIn("Plan the change before editing.", result.stdout)
        self.assertIn("Useful here:", result.stdout)
        self.assertNotIn("Detected here:", result.stdout)
        self.assertIn("make test", result.stdout)
        self.assertIn("make start", result.stdout)
        self.assertEqual(2, result.stdout.count("  make "))
        self.assertNotIn("test    ", result.stdout)
        self.assertNotIn("startup ", result.stdout)
        self.assertNotIn("verify  ", result.stdout)
        self.assertIn("Already have Jini work?", result.stdout)
        self.assertIn("jini status /path/to/work", result.stdout)
        self.assertNotIn("Resume existing Jini work when this repo already has it.", result.stdout)
        self.assertNotIn("Only if needed:", result.stdout)
        self.assertNotIn("jini doctor", result.stdout)
        self.assertNotIn("Triage the current repo and surface the strongest next move.", result.stdout)
        self.assertNotIn("Focus Jini on the detected test and verification surfaces.", result.stdout)
        self.assertNotIn("Start with a repo-aware plan before editing.", result.stdout)
        self.assertNotIn("START WITH THE TASK", result.stdout)
        self.assertNotIn("DETECTED ENTRYPOINTS", result.stdout)
        self.assertNotIn("FALLBACKS", result.stdout)
        self.assertNotIn("jini repo-map .", result.stdout)
        self.assertNotIn("jini try-example research-prd", result.stdout)

    def test_zero_arg_cli_prompts_for_task_in_interactive_repo_mode(self) -> None:
        repo = self.create_repo_fixture()
        result = self.run_cli_in_cwd(
            repo,
            env={**os.environ, "JINI_FORCE_INTERACTIVE": "1"},
            input_text="fix failing tests\nexit\n",
        )
        self.assert_ok(result)
        self.assertIn("jini> ", result.stdout)
        self.assertIn("Working on: fix failing tests", result.stdout)
        self.assertIn("Start with `make test`.", result.stdout)
        self.assertNotIn("TASK    ", result.stdout)
        self.assertNotIn("NEXT    ", result.stdout)
        self.assertNotIn("Jini CLI 0.1.0", result.stdout)
        self.assertNotIn("ALSO    ", result.stdout)
        self.assertNotIn("What next?", result.stdout)
        self.assertNotIn("PATH   ", result.stdout)
        self.assertNotIn("GIT    ", result.stdout)
        self.assertNotIn("REPO   ", result.stdout)
        self.assertNotIn("Repo: sample-repo", result.stdout)
        self.assertNotIn("Working on: Jini Research To PRD", result.stdout)
        self.assertNotIn("Session closed.", result.stdout)
        self.assertNotIn("START WITH THE TASK", result.stdout)
        self.assertNotIn("REQUEST fix failing tests", result.stdout)
        self.assertNotIn("REPO    sample-repo", result.stdout)
        self.assertNotIn("INTENT  verify", result.stdout)

    def test_zero_arg_cli_interactive_repo_mode_explains_empty_exit(self) -> None:
        repo = self.create_repo_fixture()
        result = self.run_cli_in_cwd(
            repo,
            env={**os.environ, "JINI_FORCE_INTERACTIVE": "1"},
            input_text="",
        )
        self.assert_ok(result)
        self.assertIn("jini> ", result.stdout)
        self.assertNotIn("Type a task. Use `exit` to leave.", result.stdout)
        self.assertNotIn("Session closed.", result.stdout)

    def test_zero_arg_cli_interactive_repo_mode_keeps_prompt_open_for_follow_up_turns(self) -> None:
        repo = self.create_repo_fixture()
        result = self.run_cli_in_cwd(
            repo,
            env={**os.environ, "JINI_FORCE_INTERACTIVE": "1"},
            input_text="fix failing tests\nplan this change\nexit\n",
        )
        self.assert_ok(result)
        self.assertGreaterEqual(result.stdout.count("Working on: "), 2)
        self.assertIn("Working on: fix failing tests", result.stdout)
        self.assertIn("Working on: plan this change", result.stdout)
        self.assertIn("Start with `make test`.", result.stdout)
        self.assertEqual(1, result.stdout.count("Start with `"))
        self.assertNotIn("What next?", result.stdout)

    def test_zero_arg_cli_interactive_repo_mode_routes_escape_hatches(self) -> None:
        repo = self.create_repo_fixture()
        result = self.run_cli_in_cwd(
            repo,
            env={**os.environ, "JINI_FORCE_INTERACTIVE": "1"},
            input_text="commands\ndoctor\nhelp --admin\nsetup --harness codex\nexit\n",
        )
        self.assert_ok(result)
        self.assert_compact_interactive_escape_hatch_surface(result.stdout)
        self.assertNotIn("SETUP ESCAPE HATCH", result.stdout)
        self.assertNotIn("Provider\n", result.stdout)

    def test_zero_arg_cli_interactive_repo_mode_accepts_prefixed_escape_hatches(self) -> None:
        repo = self.create_repo_fixture()
        result = self.run_cli_in_cwd(
            repo,
            env={**os.environ, "JINI_FORCE_INTERACTIVE": "1"},
            input_text="jini commands\njini doctor\njini help --admin\njini setup --harness codex\njini exit\n",
        )
        self.assert_prefixed_interactive_escape_hatch_surface(result)

    def test_zero_arg_cli_interactive_repo_mode_routes_single_token_support_hints_after_initial_request(self) -> None:
        repo = self.create_repo_fixture()
        result = self.run_cli_in_cwd(
            repo,
            env={**os.environ, "JINI_FORCE_INTERACTIVE": "1"},
            input_text="fix failing tests\nstatus\ncontinue\nopen\nstats\nexit\n",
        )
        self.assert_interactive_single_token_recovery_hints(result)

    def test_zero_arg_cli_interactive_repo_mode_routes_multi_token_support_hints_after_initial_request(self) -> None:
        repo = self.create_repo_fixture()
        result = self.run_cli_in_cwd(
            repo,
            env={**os.environ, "JINI_FORCE_INTERACTIVE": "1"},
            input_text="fix failing tests\nstatus /tmp/work\ncontinue --print-path\nopen prd\ndoctor --format json\ncommands please\nhelp --admin please\nexit\n",
        )
        self.assert_interactive_multi_token_recovery_hints(result)

    def test_zero_arg_cli_interactive_repo_mode_routes_prefixed_support_hints_after_initial_request(self) -> None:
        repo = self.create_repo_fixture()
        result = self.run_cli_in_cwd(
            repo,
            env={**os.environ, "JINI_FORCE_INTERACTIVE": "1"},
            input_text="fix failing tests\njini status /tmp/work\njini continue --print-path\njini open prd\njini doctor --format json\njini exit\n",
        )
        self.assert_prefixed_interactive_support_recovery_hints(result)

    def test_zero_arg_cli_interactive_repo_mode_routes_path_like_follow_up_hints(self) -> None:
        repo = self.create_repo_fixture()
        result = self.run_cli_in_cwd(
            repo,
            env={**os.environ, "JINI_FORCE_INTERACTIVE": "1"},
            input_text="fix failing tests\n./notes.md\nexit\n",
        )
        self.assert_ok(result)
        self.assertIn("Working on: fix failing tests", result.stdout)
        self.assertNotIn("Working on: ./notes.md", result.stdout)
        self.assertIn("Open `./notes.md` from your shell or editor, not as a live-shell task.", result.stderr)
        self.assertIn("Use `jini open` outside the live shell for Jini artifacts.", result.stderr)

    def test_zero_arg_cli_interactive_repo_mode_routes_existing_file_follow_up_hints(self) -> None:
        repo = self.create_repo_fixture()
        (repo / ".env").write_text("FOO=bar\n", encoding="utf-8")
        result = self.run_cli_in_cwd(
            repo,
            env={**os.environ, "JINI_FORCE_INTERACTIVE": "1"},
            input_text="fix failing tests\nMakefile\n.env\nexit\n",
        )
        self.assert_interactive_existing_file_recovery_hints(result)

    def test_zero_arg_cli_interactive_repo_mode_routes_prefixed_shorthand_follow_up_hints(self) -> None:
        repo = self.create_repo_fixture()
        (repo / "Makefile").write_text("test:\n\t@echo ok\n", encoding="utf-8")
        result = self.run_cli_in_cwd(
            repo,
            env={**os.environ, "JINI_FORCE_INTERACTIVE": "1"},
            input_text="fix failing tests\njini tasks\njini 1\njini Makefile\nexit\n",
        )
        self.assert_prefixed_interactive_file_shorthand_recovery_hints(result)
        self.assertIn('ERROR Unknown command "tasks".', result.stderr)
        self.assertIn('ERROR Unknown command "1".', result.stderr)

    def test_zero_arg_cli_interactive_repo_mode_routes_prefixed_path_and_filename_follow_up_hints(self) -> None:
        repo = self.create_repo_fixture()
        (repo / ".env").write_text("FOO=bar\n", encoding="utf-8")
        result = self.run_cli_in_cwd(
            repo,
            env={**os.environ, "JINI_FORCE_INTERACTIVE": "1"},
            input_text="fix failing tests\njini ./notes.md\njini notes.md\njini .env\nexit\n",
        )
        self.assert_prefixed_interactive_path_and_filename_recovery_hints(result)

    def test_zero_arg_cli_interactive_repo_mode_routes_help_aliases_after_initial_request(self) -> None:
        repo = self.create_repo_fixture()
        result = self.run_cli_in_cwd(
            repo,
            env={**os.environ, "JINI_FORCE_INTERACTIVE": "1"},
            input_text="fix failing tests\nhelp\n--help\n-h\njini help\njini --help\njini -h\nexit\n",
        )
        self.assert_interactive_help_alias_surface(result)

    def test_zero_arg_cli_shows_generic_start_surface_without_repo_or_current_work(self) -> None:
        empty_dir = self.tmp / "empty"
        empty_dir.mkdir()
        result = self.run_cli_in_cwd(empty_dir)
        self.assert_ok(result)
        self.assertIn("Run `jini` from the repo or folder that needs work.", result.stdout)
        self.assertNotIn("Nothing is in progress yet.", result.stdout)
        self.assertNotIn("START HERE", result.stdout)
        self.assertNotIn("MORE", result.stdout)
        self.assertNotIn("Jini CLI 0.1.0", result.stdout)
        self.assertNotIn("jini commands", result.stdout)
        self.assertNotIn("jini doctor", result.stdout)
        self.assertNotIn("jini status /path/to/work", result.stdout)
        self.assertNotIn("jini try-example research-prd", result.stdout)
        self.assertNotIn("jini setup --harness codex", result.stdout)
        self.assertNotIn("jini help --admin", result.stdout)

    def test_zero_arg_cli_prompts_for_task_in_interactive_no_repo_mode(self) -> None:
        empty_dir = self.tmp / "empty-interactive"
        empty_dir.mkdir()
        result = self.run_cli_in_cwd(
            empty_dir,
            env={**os.environ, "JINI_FORCE_INTERACTIVE": "1"},
            input_text="fix failing tests\nexit\n",
        )
        self.assert_ok(result)
        self.assertIn("jini> ", result.stdout)
        self.assertIn("Working on: fix failing tests", result.stdout)
        self.assertIn("Run this from the repo or folder that needs work.", result.stdout)
        self.assertNotIn("Run `jini` from the repo or folder that needs work.", result.stdout)
        self.assertNotIn("Jini CLI 0.1.0", result.stdout)
        self.assertNotIn("START HERE", result.stdout)
        self.assertNotIn("Session closed.", result.stdout)

    def test_zero_arg_cli_interactive_no_repo_mode_routes_escape_hatches(self) -> None:
        empty_dir = self.tmp / "empty-interactive-escapes"
        empty_dir.mkdir()
        result = self.run_cli_in_cwd(
            empty_dir,
            env={**os.environ, "JINI_FORCE_INTERACTIVE": "1"},
            input_text="commands\ndoctor\nhelp --admin\nsetup --harness codex\nexit\n",
        )
        self.assert_ok(result)
        self.assert_compact_interactive_escape_hatch_surface(result.stdout)

    def test_zero_arg_cli_interactive_no_repo_mode_accepts_prefixed_escape_hatches(self) -> None:
        empty_dir = self.tmp / "empty-interactive-prefixed-escapes"
        empty_dir.mkdir()
        result = self.run_cli_in_cwd(
            empty_dir,
            env={**os.environ, "JINI_FORCE_INTERACTIVE": "1"},
            input_text="jini commands\njini doctor\njini help --admin\njini setup --harness codex\njini exit\n",
        )
        self.assert_prefixed_interactive_escape_hatch_surface(result)

    def test_zero_arg_cli_interactive_no_repo_mode_routes_single_token_support_hints_after_initial_request(self) -> None:
        empty_dir = self.tmp / "empty-interactive-support"
        empty_dir.mkdir()
        result = self.run_cli_in_cwd(
            empty_dir,
            env={**os.environ, "JINI_FORCE_INTERACTIVE": "1"},
            input_text="fix failing tests\nstatus\ncontinue\nopen\nstats\nexit\n",
        )
        self.assert_interactive_single_token_recovery_hints(result)

    def test_zero_arg_cli_interactive_no_repo_mode_routes_multi_token_support_hints_after_initial_request(self) -> None:
        empty_dir = self.tmp / "empty-interactive-multi-support"
        empty_dir.mkdir()
        result = self.run_cli_in_cwd(
            empty_dir,
            env={**os.environ, "JINI_FORCE_INTERACTIVE": "1"},
            input_text="fix failing tests\nstatus /tmp/work\ncontinue --print-path\nopen prd\ndoctor --format json\ncommands please\nhelp --admin please\nexit\n",
        )
        self.assert_interactive_multi_token_recovery_hints(result)

    def test_zero_arg_cli_interactive_no_repo_mode_routes_prefixed_support_hints_after_initial_request(self) -> None:
        empty_dir = self.tmp / "empty-interactive-prefixed-support"
        empty_dir.mkdir()
        result = self.run_cli_in_cwd(
            empty_dir,
            env={**os.environ, "JINI_FORCE_INTERACTIVE": "1"},
            input_text="fix failing tests\njini status /tmp/work\njini continue --print-path\njini open prd\njini doctor --format json\njini exit\n",
        )
        self.assert_prefixed_interactive_support_recovery_hints(result)

    def test_zero_arg_cli_interactive_no_repo_mode_routes_path_like_follow_up_hints(self) -> None:
        empty_dir = self.tmp / "empty-interactive-path-like"
        empty_dir.mkdir()
        result = self.run_cli_in_cwd(
            empty_dir,
            env={**os.environ, "JINI_FORCE_INTERACTIVE": "1"},
            input_text="fix failing tests\n./notes.md\nexit\n",
        )
        self.assert_ok(result)
        self.assertIn("Working on: fix failing tests", result.stdout)
        self.assertNotIn("Working on: ./notes.md", result.stdout)
        self.assertIn("Open `./notes.md` from your shell or editor, not as a live-shell task.", result.stderr)
        self.assertIn("Use `jini open` outside the live shell for Jini artifacts.", result.stderr)

    def test_zero_arg_cli_interactive_no_repo_mode_routes_existing_file_follow_up_hints(self) -> None:
        empty_dir = self.tmp / "empty-interactive-files"
        empty_dir.mkdir()
        (empty_dir / "Makefile").write_text("test:\n\t@echo ok\n", encoding="utf-8")
        (empty_dir / ".env").write_text("FOO=bar\n", encoding="utf-8")
        result = self.run_cli_in_cwd(
            empty_dir,
            env={**os.environ, "JINI_FORCE_INTERACTIVE": "1"},
            input_text="fix failing tests\nMakefile\n.env\nexit\n",
        )
        self.assert_interactive_existing_file_recovery_hints(result)

    def test_zero_arg_cli_interactive_no_repo_mode_routes_prefixed_shorthand_follow_up_hints(self) -> None:
        empty_dir = self.tmp / "empty-interactive-prefixed-shorthand"
        empty_dir.mkdir()
        (empty_dir / "Makefile").write_text("test:\n\t@echo ok\n", encoding="utf-8")
        result = self.run_cli_in_cwd(
            empty_dir,
            env={**os.environ, "JINI_FORCE_INTERACTIVE": "1"},
            input_text="fix failing tests\njini tasks\njini 1\njini Makefile\nexit\n",
        )
        self.assert_prefixed_interactive_file_shorthand_recovery_hints(result)
        self.assertIn('ERROR Unknown command "tasks".', result.stderr)
        self.assertIn('ERROR Unknown command "1".', result.stderr)

    def test_zero_arg_cli_interactive_no_repo_mode_routes_prefixed_path_and_filename_follow_up_hints(self) -> None:
        empty_dir = self.tmp / "empty-interactive-prefixed-paths"
        empty_dir.mkdir()
        (empty_dir / "notes.md").write_text("notes\n", encoding="utf-8")
        (empty_dir / ".env").write_text("FOO=bar\n", encoding="utf-8")
        result = self.run_cli_in_cwd(
            empty_dir,
            env={**os.environ, "JINI_FORCE_INTERACTIVE": "1"},
            input_text="fix failing tests\njini ./notes.md\njini notes.md\njini .env\nexit\n",
        )
        self.assert_prefixed_interactive_path_and_filename_recovery_hints(result)

    def test_zero_arg_cli_interactive_no_repo_mode_routes_help_aliases_after_initial_request(self) -> None:
        empty_dir = self.tmp / "empty-interactive-help-aliases"
        empty_dir.mkdir()
        result = self.run_cli_in_cwd(
            empty_dir,
            env={**os.environ, "JINI_FORCE_INTERACTIVE": "1"},
            input_text="fix failing tests\nhelp\n--help\n-h\njini help\njini --help\njini -h\nexit\n",
        )
        self.assert_interactive_help_alias_surface(result)

    def test_zero_arg_cli_resumes_current_work_surface(self) -> None:
        example_output = self.tmp / "research-example"
        seeded = self.run_cli("try-example", "research-prd", "--output", example_output)
        self.assert_ok(seeded)

        result = self.run_cli()
        self.assert_ok(result)
        self.assertIn("WORK   example-research-prd", result.stdout)
        self.assertIn("READY NOW", result.stdout)
        self.assertNotIn("START HERE", result.stdout)

    def test_zero_arg_cli_keeps_shell_open_when_current_work_exists(self) -> None:
        repo = self.create_repo_fixture()
        example_output = repo / ".jini-example"
        seeded = self.run_cli_in_cwd(repo, "try-example", "research-prd", "--output", example_output)
        self.assert_ok(seeded)

        result = self.run_cli_in_cwd(
            repo,
            env={**os.environ, "JINI_FORCE_INTERACTIVE": "1"},
            input_text="fix failing tests\nexit\n",
        )
        self.assert_ok(result)
        self.assertIn("jini> ", result.stdout)
        self.assertIn("Working on: fix failing tests", result.stdout)
        self.assertIn("Start with `make test`.", result.stdout)
        self.assertNotIn("Session closed.", result.stdout)
        self.assertNotIn("WORK   example-research-prd", result.stdout)
        self.assertNotIn("ARTIFACT SHELF", result.stdout)
        self.assertNotIn("Repo: sample-repo", result.stdout)
        self.assertNotIn("Working on: Jini Research To PRD", result.stdout)

    def test_zero_arg_cli_active_work_shell_uses_same_task_hint_as_repo_start(self) -> None:
        repo = self.create_repo_fixture()
        example_output = repo / ".jini-example"
        seeded = self.run_cli_in_cwd(repo, "try-example", "research-prd", "--output", example_output)
        self.assert_ok(seeded)

        result = self.run_cli_in_cwd(
            repo,
            env={**os.environ, "JINI_FORCE_INTERACTIVE": "1"},
            input_text="exit\n",
        )
        self.assert_ok(result)
        self.assertEqual(0, result.stdout.count("Type a task. Use `exit` to leave."))
        self.assertNotIn("Jini CLI 0.1.0", result.stdout)

    def test_zero_arg_cli_active_work_shell_skips_report_card_before_prompt(self) -> None:
        repo = self.create_repo_fixture()
        example_output = repo / ".jini-example"
        seeded = self.run_cli_in_cwd(repo, "try-example", "research-prd", "--output", example_output)
        self.assert_ok(seeded)

        result = self.run_cli_in_cwd(
            repo,
            env={**os.environ, "JINI_FORCE_INTERACTIVE": "1"},
            input_text="exit\n",
        )
        self.assert_ok(result)
        self.assertNotIn("WORK   example-research-prd", result.stdout)
        self.assertNotIn("TITLE  Jini Research To PRD", result.stdout)
        self.assertNotIn("HEALTH ready-to-verify", result.stdout)
        self.assertNotIn("STATE  awaiting_verification", result.stdout)
        self.assertNotIn("Repo: sample-repo", result.stdout)
        self.assertNotIn("Working on: Jini Research To PRD", result.stdout)

    def test_zero_arg_cli_saved_session_shell_skips_report_card_before_prompt(self) -> None:
        repo = self.create_repo_fixture()
        example_output = repo / ".jini-example"
        seeded = self.run_cli_in_cwd(repo, "try-example", "research-prd", "--output", example_output)
        self.assert_ok(seeded)

        current_work = self.tmp / ".jini" / "current-work.json"
        current_work.unlink()

        result = self.run_cli_in_cwd(
            repo,
            env={**os.environ, "JINI_FORCE_INTERACTIVE": "1"},
            input_text="exit\n",
        )
        self.assert_ok(result)
        self.assertIn("jini> ", result.stdout)
        self.assertNotIn("Type a task. Use `exit` to leave.", result.stdout)
        self.assertNotIn("Session closed.", result.stdout)
        self.assertNotIn("WORK   example-research-prd", result.stdout)
        self.assertNotIn("TITLE  Jini Research To PRD", result.stdout)
        self.assertNotIn("HEALTH ready-to-verify", result.stdout)
        self.assertNotIn("STATE  awaiting_verification", result.stdout)
        self.assertNotIn("Repo: sample-repo", result.stdout)
        self.assertNotIn("Working on: Jini Research To PRD", result.stdout)

    def test_zero_arg_cli_active_work_shell_routes_escape_hatches(self) -> None:
        repo = self.create_repo_fixture()
        example_output = repo / ".jini-example"
        seeded = self.run_cli_in_cwd(repo, "try-example", "research-prd", "--output", example_output)
        self.assert_ok(seeded)

        result = self.run_cli_in_cwd(
            repo,
            env={**os.environ, "JINI_FORCE_INTERACTIVE": "1"},
            input_text="commands\ndoctor\nhelp --admin\nsetup --harness codex\nexit\n",
        )
        self.assert_ok(result)
        self.assert_compact_interactive_escape_hatch_surface(result.stdout)
        self.assertNotIn("WORK   example-research-prd", result.stdout)
        self.assertNotIn("ARTIFACT SHELF", result.stdout)

    def test_zero_arg_cli_active_work_shell_keeps_prompt_open_for_follow_up_turns(self) -> None:
        repo = self.create_repo_fixture()
        example_output = repo / ".jini-example"
        seeded = self.run_cli_in_cwd(repo, "try-example", "research-prd", "--output", example_output)
        self.assert_ok(seeded)

        result = self.run_cli_in_cwd(
            repo,
            env={**os.environ, "JINI_FORCE_INTERACTIVE": "1"},
            input_text="fix failing tests\nplan this change\nexit\n",
        )
        self.assert_ok(result)
        self.assertIn("jini> ", result.stdout)
        self.assertIn("Working on: fix failing tests", result.stdout)
        self.assertIn("Working on: plan this change", result.stdout)
        self.assertIn("Start with `make test`.", result.stdout)
        self.assertEqual(1, result.stdout.count("Start with `"))
        self.assertNotIn("WORK   example-research-prd", result.stdout)
        self.assertNotIn("ARTIFACT SHELF", result.stdout)
        self.assertNotIn("Session closed.", result.stdout)

    def test_zero_arg_cli_active_work_shell_ignores_blank_input_after_initial_request(self) -> None:
        repo = self.create_repo_fixture()
        example_output = repo / ".jini-example"
        seeded = self.run_cli_in_cwd(repo, "try-example", "research-prd", "--output", example_output)
        self.assert_ok(seeded)

        result = self.run_cli_in_cwd(
            repo,
            env={**os.environ, "JINI_FORCE_INTERACTIVE": "1"},
            input_text="fix failing tests\n\nplan this change\nexit\n",
        )
        self.assert_ok(result)
        self.assertIn("Working on: fix failing tests", result.stdout)
        self.assertIn("Working on: plan this change", result.stdout)
        self.assertNotIn("Type the next task, or `exit` to leave.", result.stdout)
        self.assertNotIn("Session closed.", result.stdout)

    def test_zero_arg_cli_active_work_shell_routes_escape_hatches_after_initial_request(self) -> None:
        repo = self.create_repo_fixture()
        example_output = repo / ".jini-example"
        seeded = self.run_cli_in_cwd(repo, "try-example", "research-prd", "--output", example_output)
        self.assert_ok(seeded)

        result = self.run_cli_in_cwd(
            repo,
            env={**os.environ, "JINI_FORCE_INTERACTIVE": "1"},
            input_text="fix failing tests\ncommands\ndoctor\nhelp --admin\nsetup --harness codex\nexit\n",
        )
        self.assert_ok(result)
        self.assertIn("Working on: fix failing tests", result.stdout)
        self.assertIn("Start with `make test`.", result.stdout)
        self.assert_compact_interactive_escape_hatch_surface(result.stdout)
        self.assertNotIn("WORK   example-research-prd", result.stdout)
        self.assertNotIn("ARTIFACT SHELF", result.stdout)

    def test_zero_arg_cli_active_work_shell_accepts_prefixed_escape_hatches_after_initial_request(self) -> None:
        repo = self.create_repo_fixture()
        example_output = repo / ".jini-example"
        seeded = self.run_cli_in_cwd(repo, "try-example", "research-prd", "--output", example_output)
        self.assert_ok(seeded)

        result = self.run_cli_in_cwd(
            repo,
            env={**os.environ, "JINI_FORCE_INTERACTIVE": "1"},
            input_text="fix failing tests\njini commands\njini doctor\njini help --admin\njini setup --harness codex\njini exit\n",
        )
        self.assert_prefixed_interactive_escape_hatch_surface(result)
        self.assertNotIn("WORK   example-research-prd", result.stdout)
        self.assertNotIn("ARTIFACT SHELF", result.stdout)

    def test_zero_arg_cli_active_work_shell_routes_single_token_support_hints_after_initial_request(self) -> None:
        repo = self.create_repo_fixture()
        example_output = repo / ".jini-example"
        seeded = self.run_cli_in_cwd(repo, "try-example", "research-prd", "--output", example_output)
        self.assert_ok(seeded)

        result = self.run_cli_in_cwd(
            repo,
            env={**os.environ, "JINI_FORCE_INTERACTIVE": "1"},
            input_text="fix failing tests\nstatus\ncontinue\nopen\nstats\nexit\n",
        )
        self.assert_interactive_single_token_recovery_hints(result)
        self.assertNotIn("WORK   example-research-prd", result.stdout)
        self.assertNotIn("ARTIFACT SHELF", result.stdout)

    def test_zero_arg_cli_active_work_shell_routes_multi_token_support_hints_after_initial_request(self) -> None:
        repo = self.create_repo_fixture()
        example_output = repo / ".jini-example"
        seeded = self.run_cli_in_cwd(repo, "try-example", "research-prd", "--output", example_output)
        self.assert_ok(seeded)

        result = self.run_cli_in_cwd(
            repo,
            env={**os.environ, "JINI_FORCE_INTERACTIVE": "1"},
            input_text="fix failing tests\nstatus /tmp/work\ncontinue --print-path\nopen prd\ndoctor --format json\ncommands please\nhelp --admin please\nexit\n",
        )
        self.assert_interactive_multi_token_recovery_hints(result)
        self.assertNotIn("WORK   example-research-prd", result.stdout)
        self.assertNotIn("ARTIFACT SHELF", result.stdout)

    def test_zero_arg_cli_active_work_shell_routes_prefixed_support_hints_after_initial_request(self) -> None:
        repo = self.create_repo_fixture()
        example_output = repo / ".jini-example"
        seeded = self.run_cli_in_cwd(repo, "try-example", "research-prd", "--output", example_output)
        self.assert_ok(seeded)

        result = self.run_cli_in_cwd(
            repo,
            env={**os.environ, "JINI_FORCE_INTERACTIVE": "1"},
            input_text="fix failing tests\njini status /tmp/work\njini continue --print-path\njini open prd\njini doctor --format json\njini exit\n",
        )
        self.assert_prefixed_interactive_support_recovery_hints(result)
        self.assertNotIn("WORK   example-research-prd", result.stdout)
        self.assertNotIn("ARTIFACT SHELF", result.stdout)

    def test_zero_arg_cli_active_work_shell_routes_artifact_and_path_follow_up_hints(self) -> None:
        repo = self.create_repo_fixture()
        example_output = repo / ".jini-example"
        seeded = self.run_cli_in_cwd(repo, "try-example", "research-prd", "--output", example_output)
        self.assert_ok(seeded)

        result = self.run_cli_in_cwd(
            repo,
            env={**os.environ, "JINI_FORCE_INTERACTIVE": "1"},
            input_text="fix failing tests\ntasks\nprd\n./notes.md\nexit\n",
        )
        self.assert_interactive_artifact_and_path_recovery_hints(result)
        self.assertNotIn("WORK   example-research-prd", result.stdout)
        self.assertNotIn("ARTIFACT SHELF", result.stdout)

    def test_zero_arg_cli_active_work_shell_routes_numeric_and_filename_follow_up_hints(self) -> None:
        repo = self.create_repo_fixture()
        example_output = repo / ".jini-example"
        seeded = self.run_cli_in_cwd(repo, "try-example", "research-prd", "--output", example_output)
        self.assert_ok(seeded)

        result = self.run_cli_in_cwd(
            repo,
            env={**os.environ, "JINI_FORCE_INTERACTIVE": "1"},
            input_text="fix failing tests\n1\n2\nnotes.md\nexit\n",
        )
        self.assert_interactive_numeric_and_filename_recovery_hints(result)
        self.assertNotIn("WORK   example-research-prd", result.stdout)
        self.assertNotIn("ARTIFACT SHELF", result.stdout)

    def test_zero_arg_cli_active_work_shell_routes_existing_file_follow_up_hints(self) -> None:
        repo = self.create_repo_fixture()
        (repo / ".env").write_text("FOO=bar\n", encoding="utf-8")
        example_output = repo / ".jini-example"
        seeded = self.run_cli_in_cwd(repo, "try-example", "research-prd", "--output", example_output)
        self.assert_ok(seeded)

        result = self.run_cli_in_cwd(
            repo,
            env={**os.environ, "JINI_FORCE_INTERACTIVE": "1"},
            input_text="fix failing tests\nMakefile\n.env\nexit\n",
        )
        self.assert_interactive_existing_file_recovery_hints(result)
        self.assertNotIn("WORK   example-research-prd", result.stdout)
        self.assertNotIn("ARTIFACT SHELF", result.stdout)

    def test_zero_arg_cli_active_work_shell_routes_prefixed_shorthand_follow_up_hints(self) -> None:
        repo = self.create_repo_fixture()
        example_output = repo / ".jini-example"
        seeded = self.run_cli_in_cwd(repo, "try-example", "research-prd", "--output", example_output)
        self.assert_ok(seeded)

        result = self.run_cli_in_cwd(
            repo,
            env={**os.environ, "JINI_FORCE_INTERACTIVE": "1"},
            input_text="fix failing tests\njini tasks\njini 1\njini Makefile\nexit\n",
        )
        self.assert_prefixed_interactive_artifact_and_file_shorthand_recovery_hints(result)
        self.assertNotIn("WORK   example-research-prd", result.stdout)
        self.assertNotIn("ARTIFACT SHELF", result.stdout)

    def test_zero_arg_cli_active_work_shell_routes_prefixed_path_and_filename_follow_up_hints(self) -> None:
        repo = self.create_repo_fixture()
        (repo / ".env").write_text("FOO=bar\n", encoding="utf-8")
        example_output = repo / ".jini-example"
        seeded = self.run_cli_in_cwd(repo, "try-example", "research-prd", "--output", example_output)
        self.assert_ok(seeded)

        result = self.run_cli_in_cwd(
            repo,
            env={**os.environ, "JINI_FORCE_INTERACTIVE": "1"},
            input_text="fix failing tests\njini ./notes.md\njini notes.md\njini .env\nexit\n",
        )
        self.assert_prefixed_interactive_path_and_filename_recovery_hints(result)
        self.assertNotIn("WORK   example-research-prd", result.stdout)
        self.assertNotIn("ARTIFACT SHELF", result.stdout)

    def test_zero_arg_cli_active_work_shell_routes_help_aliases_after_initial_request(self) -> None:
        repo = self.create_repo_fixture()
        example_output = repo / ".jini-example"
        seeded = self.run_cli_in_cwd(repo, "try-example", "research-prd", "--output", example_output)
        self.assert_ok(seeded)

        result = self.run_cli_in_cwd(
            repo,
            env={**os.environ, "JINI_FORCE_INTERACTIVE": "1"},
            input_text="fix failing tests\nhelp\n--help\n-h\njini help\njini --help\njini -h\nexit\n",
        )
        self.assert_interactive_help_alias_surface(result)
        self.assertNotIn("WORK   example-research-prd", result.stdout)
        self.assertNotIn("ARTIFACT SHELF", result.stdout)

    def test_zero_arg_cli_saved_session_shell_routes_escape_hatches(self) -> None:
        repo = self.create_repo_fixture()
        example_output = repo / ".jini-example"
        seeded = self.run_cli_in_cwd(repo, "try-example", "research-prd", "--output", example_output)
        self.assert_ok(seeded)

        current_work = self.tmp / ".jini" / "current-work.json"
        current_work.unlink()

        result = self.run_cli_in_cwd(
            repo,
            env={**os.environ, "JINI_FORCE_INTERACTIVE": "1"},
            input_text="commands\ndoctor\nhelp --admin\nsetup --harness codex\nexit\n",
        )
        self.assert_ok(result)
        self.assert_compact_interactive_escape_hatch_surface(result.stdout)
        self.assertNotIn("WORK   example-research-prd", result.stdout)
        self.assertNotIn("ARTIFACT SHELF", result.stdout)

    def test_zero_arg_cli_saved_session_shell_keeps_prompt_open_for_follow_up_turns(self) -> None:
        repo = self.create_repo_fixture()
        example_output = repo / ".jini-example"
        seeded = self.run_cli_in_cwd(repo, "try-example", "research-prd", "--output", example_output)
        self.assert_ok(seeded)

        current_work = self.tmp / ".jini" / "current-work.json"
        current_work.unlink()

        result = self.run_cli_in_cwd(
            repo,
            env={**os.environ, "JINI_FORCE_INTERACTIVE": "1"},
            input_text="fix failing tests\nplan this change\nexit\n",
        )
        self.assert_ok(result)
        self.assertIn("jini> ", result.stdout)
        self.assertIn("Working on: fix failing tests", result.stdout)
        self.assertIn("Working on: plan this change", result.stdout)
        self.assertIn("Start with `make test`.", result.stdout)
        self.assertEqual(1, result.stdout.count("Start with `"))
        self.assertNotIn("WORK   example-research-prd", result.stdout)
        self.assertNotIn("ARTIFACT SHELF", result.stdout)
        self.assertNotIn("Session closed.", result.stdout)

    def test_zero_arg_cli_saved_session_shell_ignores_blank_input_after_initial_request(self) -> None:
        repo = self.create_repo_fixture()
        example_output = repo / ".jini-example"
        seeded = self.run_cli_in_cwd(repo, "try-example", "research-prd", "--output", example_output)
        self.assert_ok(seeded)

        current_work = self.tmp / ".jini" / "current-work.json"
        current_work.unlink()

        result = self.run_cli_in_cwd(
            repo,
            env={**os.environ, "JINI_FORCE_INTERACTIVE": "1"},
            input_text="fix failing tests\n\nplan this change\nexit\n",
        )
        self.assert_ok(result)
        self.assertIn("Working on: fix failing tests", result.stdout)
        self.assertIn("Working on: plan this change", result.stdout)
        self.assertNotIn("Type the next task, or `exit` to leave.", result.stdout)
        self.assertNotIn("Session closed.", result.stdout)

    def test_zero_arg_cli_saved_session_shell_routes_escape_hatches_after_initial_request(self) -> None:
        repo = self.create_repo_fixture()
        example_output = repo / ".jini-example"
        seeded = self.run_cli_in_cwd(repo, "try-example", "research-prd", "--output", example_output)
        self.assert_ok(seeded)

        current_work = self.tmp / ".jini" / "current-work.json"
        current_work.unlink()

        result = self.run_cli_in_cwd(
            repo,
            env={**os.environ, "JINI_FORCE_INTERACTIVE": "1"},
            input_text="fix failing tests\ncommands\ndoctor\nhelp --admin\nsetup --harness codex\nexit\n",
        )
        self.assert_ok(result)
        self.assertIn("Working on: fix failing tests", result.stdout)
        self.assertIn("Start with `make test`.", result.stdout)
        self.assert_compact_interactive_escape_hatch_surface(result.stdout)
        self.assertNotIn("WORK   example-research-prd", result.stdout)
        self.assertNotIn("ARTIFACT SHELF", result.stdout)

    def test_zero_arg_cli_saved_session_shell_accepts_prefixed_escape_hatches_after_initial_request(self) -> None:
        repo = self.create_repo_fixture()
        example_output = repo / ".jini-example"
        seeded = self.run_cli_in_cwd(repo, "try-example", "research-prd", "--output", example_output)
        self.assert_ok(seeded)

        current_work = self.tmp / ".jini" / "current-work.json"
        current_work.unlink()

        result = self.run_cli_in_cwd(
            repo,
            env={**os.environ, "JINI_FORCE_INTERACTIVE": "1"},
            input_text="fix failing tests\njini commands\njini doctor\njini help --admin\njini setup --harness codex\njini exit\n",
        )
        self.assert_prefixed_interactive_escape_hatch_surface(result)
        self.assertNotIn("WORK   example-research-prd", result.stdout)
        self.assertNotIn("ARTIFACT SHELF", result.stdout)

    def test_zero_arg_cli_saved_session_shell_routes_single_token_support_hints_after_initial_request(self) -> None:
        repo = self.create_repo_fixture()
        example_output = repo / ".jini-example"
        seeded = self.run_cli_in_cwd(repo, "try-example", "research-prd", "--output", example_output)
        self.assert_ok(seeded)

        current_work = self.tmp / ".jini" / "current-work.json"
        current_work.unlink()

        result = self.run_cli_in_cwd(
            repo,
            env={**os.environ, "JINI_FORCE_INTERACTIVE": "1"},
            input_text="fix failing tests\nstatus\ncontinue\nopen\nstats\nexit\n",
        )
        self.assert_interactive_single_token_recovery_hints(result)
        self.assertNotIn("WORK   example-research-prd", result.stdout)
        self.assertNotIn("ARTIFACT SHELF", result.stdout)

    def test_zero_arg_cli_saved_session_shell_routes_multi_token_support_hints_after_initial_request(self) -> None:
        repo = self.create_repo_fixture()
        example_output = repo / ".jini-example"
        seeded = self.run_cli_in_cwd(repo, "try-example", "research-prd", "--output", example_output)
        self.assert_ok(seeded)

        current_work = self.tmp / ".jini" / "current-work.json"
        current_work.unlink()

        result = self.run_cli_in_cwd(
            repo,
            env={**os.environ, "JINI_FORCE_INTERACTIVE": "1"},
            input_text="fix failing tests\nstatus /tmp/work\ncontinue --print-path\nopen prd\ndoctor --format json\ncommands please\nhelp --admin please\nexit\n",
        )
        self.assert_interactive_multi_token_recovery_hints(result)
        self.assertNotIn("WORK   example-research-prd", result.stdout)
        self.assertNotIn("ARTIFACT SHELF", result.stdout)

    def test_zero_arg_cli_saved_session_shell_routes_prefixed_support_hints_after_initial_request(self) -> None:
        repo = self.create_repo_fixture()
        example_output = repo / ".jini-example"
        seeded = self.run_cli_in_cwd(repo, "try-example", "research-prd", "--output", example_output)
        self.assert_ok(seeded)

        current_work = self.tmp / ".jini" / "current-work.json"
        current_work.unlink()

        result = self.run_cli_in_cwd(
            repo,
            env={**os.environ, "JINI_FORCE_INTERACTIVE": "1"},
            input_text="fix failing tests\njini status /tmp/work\njini continue --print-path\njini open prd\njini doctor --format json\njini exit\n",
        )
        self.assert_prefixed_interactive_support_recovery_hints(result)
        self.assertNotIn("WORK   example-research-prd", result.stdout)
        self.assertNotIn("ARTIFACT SHELF", result.stdout)

    def test_zero_arg_cli_saved_session_shell_routes_artifact_and_path_follow_up_hints(self) -> None:
        repo = self.create_repo_fixture()
        example_output = repo / ".jini-example"
        seeded = self.run_cli_in_cwd(repo, "try-example", "research-prd", "--output", example_output)
        self.assert_ok(seeded)

        current_work = self.tmp / ".jini" / "current-work.json"
        current_work.unlink()

        result = self.run_cli_in_cwd(
            repo,
            env={**os.environ, "JINI_FORCE_INTERACTIVE": "1"},
            input_text="fix failing tests\ntasks\nprd\n./notes.md\nexit\n",
        )
        self.assert_interactive_artifact_and_path_recovery_hints(result)
        self.assertNotIn("WORK   example-research-prd", result.stdout)
        self.assertNotIn("ARTIFACT SHELF", result.stdout)

    def test_zero_arg_cli_saved_session_shell_routes_numeric_and_filename_follow_up_hints(self) -> None:
        repo = self.create_repo_fixture()
        example_output = repo / ".jini-example"
        seeded = self.run_cli_in_cwd(repo, "try-example", "research-prd", "--output", example_output)
        self.assert_ok(seeded)

        current_work = self.tmp / ".jini" / "current-work.json"
        current_work.unlink()

        result = self.run_cli_in_cwd(
            repo,
            env={**os.environ, "JINI_FORCE_INTERACTIVE": "1"},
            input_text="fix failing tests\n1\n2\nnotes.md\nexit\n",
        )
        self.assert_interactive_numeric_and_filename_recovery_hints(result)
        self.assertNotIn("WORK   example-research-prd", result.stdout)
        self.assertNotIn("ARTIFACT SHELF", result.stdout)

    def test_zero_arg_cli_saved_session_shell_routes_existing_file_follow_up_hints(self) -> None:
        repo = self.create_repo_fixture()
        (repo / ".env").write_text("FOO=bar\n", encoding="utf-8")
        example_output = repo / ".jini-example"
        seeded = self.run_cli_in_cwd(repo, "try-example", "research-prd", "--output", example_output)
        self.assert_ok(seeded)

        current_work = self.tmp / ".jini" / "current-work.json"
        current_work.unlink()

        result = self.run_cli_in_cwd(
            repo,
            env={**os.environ, "JINI_FORCE_INTERACTIVE": "1"},
            input_text="fix failing tests\nMakefile\n.env\nexit\n",
        )
        self.assert_interactive_existing_file_recovery_hints(result)
        self.assertNotIn("WORK   example-research-prd", result.stdout)
        self.assertNotIn("ARTIFACT SHELF", result.stdout)

    def test_zero_arg_cli_saved_session_shell_routes_prefixed_shorthand_follow_up_hints(self) -> None:
        repo = self.create_repo_fixture()
        example_output = repo / ".jini-example"
        seeded = self.run_cli_in_cwd(repo, "try-example", "research-prd", "--output", example_output)
        self.assert_ok(seeded)

        current_work = self.tmp / ".jini" / "current-work.json"
        current_work.unlink()

        result = self.run_cli_in_cwd(
            repo,
            env={**os.environ, "JINI_FORCE_INTERACTIVE": "1"},
            input_text="fix failing tests\njini tasks\njini 1\njini Makefile\nexit\n",
        )
        self.assert_prefixed_interactive_artifact_and_file_shorthand_recovery_hints(result)
        self.assertNotIn("WORK   example-research-prd", result.stdout)
        self.assertNotIn("ARTIFACT SHELF", result.stdout)

    def test_zero_arg_cli_saved_session_shell_routes_prefixed_path_and_filename_follow_up_hints(self) -> None:
        repo = self.create_repo_fixture()
        (repo / ".env").write_text("FOO=bar\n", encoding="utf-8")
        example_output = repo / ".jini-example"
        seeded = self.run_cli_in_cwd(repo, "try-example", "research-prd", "--output", example_output)
        self.assert_ok(seeded)

        current_work = self.tmp / ".jini" / "current-work.json"
        current_work.unlink()

        result = self.run_cli_in_cwd(
            repo,
            env={**os.environ, "JINI_FORCE_INTERACTIVE": "1"},
            input_text="fix failing tests\njini ./notes.md\njini notes.md\njini .env\nexit\n",
        )
        self.assert_prefixed_interactive_path_and_filename_recovery_hints(result)
        self.assertNotIn("WORK   example-research-prd", result.stdout)
        self.assertNotIn("ARTIFACT SHELF", result.stdout)

    def test_zero_arg_cli_saved_session_shell_routes_help_aliases_after_initial_request(self) -> None:
        repo = self.create_repo_fixture()
        example_output = repo / ".jini-example"
        seeded = self.run_cli_in_cwd(repo, "try-example", "research-prd", "--output", example_output)
        self.assert_ok(seeded)

        current_work = self.tmp / ".jini" / "current-work.json"
        current_work.unlink()

        result = self.run_cli_in_cwd(
            repo,
            env={**os.environ, "JINI_FORCE_INTERACTIVE": "1"},
            input_text="fix failing tests\nhelp\n--help\n-h\njini help\njini --help\njini -h\nexit\n",
        )
        self.assert_interactive_help_alias_surface(result)
        self.assertNotIn("WORK   example-research-prd", result.stdout)
        self.assertNotIn("ARTIFACT SHELF", result.stdout)

    def test_help_with_request_tail_shows_corrective_hint(self) -> None:
        request_tokens = HELP_TAIL_EXAMPLE_REQUEST.split()
        result = self.run_cli("help", *request_tokens)
        self.assert_error(result)
        self.assertEqual(
            result.stderr.strip().splitlines(),
            list(help_tail_message_lines("jini", "help", "CLI overview", request_tokens)),
        )
        self.assertEqual("", result.stdout)
        self.assertNotIn("OPEN JINI", result.stderr)

    def test_dash_help_with_request_tail_shows_corrective_hint(self) -> None:
        request_tokens = HELP_TAIL_EXAMPLE_REQUEST.split()
        result = self.run_cli("--help", *request_tokens)
        self.assert_error(result)
        self.assertEqual(
            result.stderr.strip().splitlines(),
            list(help_tail_message_lines("jini", "--help", "CLI overview", request_tokens)),
        )
        self.assertEqual("", result.stdout)
        self.assertNotIn("OPEN JINI", result.stderr)

    def test_help_all_shows_public_command_inventory(self) -> None:
        result = self.run_cli("help", "--all")
        self.assert_ok(result)
        self.assertIn("Public command inventory", result.stdout)
        self.assertIn("SUPPORT THE CURRENT WORK", result.stdout)
        self.assertIn("jini status", result.stdout)
        self.assertIn("jini continue", result.stdout)
        self.assertIn("jini open", result.stdout)
        self.assertIn("jini help --admin", result.stdout)
        self.assertNotIn("jini try-example research-prd", result.stdout)
        self.assertNotIn("jini metrics", result.stdout)
        self.assertNotIn("usage: jini", result.stdout)
        self.assertNotIn("stage-framework-experiment", result.stdout)

    def test_commands_shows_public_command_inventory(self) -> None:
        result = self.run_cli("commands")
        self.assert_ok(result)
        self.assertIn("Public command inventory", result.stdout)
        self.assertIn("START WITH JINI", result.stdout)
        self.assertIn("SUPPORT THE CURRENT WORK", result.stdout)
        self.assertIn("jini continue", result.stdout)
        self.assertIn("jini open", result.stdout)
        self.assertIn("jini help --admin", result.stdout)
        self.assertNotIn("jini get-started --harness codex", result.stdout)
        self.assertNotIn("jini try-example research-prd", result.stdout)
        self.assertNotIn("jini metrics", result.stdout)

    def test_commands_with_request_tail_shows_corrective_hint(self) -> None:
        request_tokens = HELP_TAIL_EXAMPLE_REQUEST.split()
        result = self.run_cli("commands", *request_tokens)
        self.assert_error(result)
        self.assertEqual(
            result.stderr.strip().splitlines(),
            list(help_tail_message_lines("jini", "commands", "public command inventory", request_tokens)),
        )
        self.assertEqual("", result.stdout)
        self.assertNotIn("Public command inventory", result.stderr)

    def test_request_text_in_repo_surfaces_repo_intake_instead_of_argparse(self) -> None:
        repo = self.create_repo_fixture()
        result = self.run_cli_in_cwd(repo, "fix", "failing", "tests")
        self.assert_ok(result)
        self.assertIn("Working on: fix failing tests", result.stdout)
        self.assertIn("Start with `make test`.", result.stdout)
        self.assertNotIn("REQUEST fix failing tests", result.stdout)
        self.assertNotIn("REPO    sample-repo", result.stdout)
        self.assertNotIn("INTENT  verify", result.stdout)
        self.assertNotIn("BEST NEXT MOVE", result.stdout)
        self.assertNotIn("KEEP MOVING", result.stdout)
        self.assertNotIn("jini repo-map .", result.stdout)
        self.assertNotIn("jini doctor", result.stdout)
        self.assertNotIn("usage: jini", result.stdout)

    def test_request_text_outside_repo_returns_concise_recovery_hint(self) -> None:
        empty_dir = self.tmp / "generic-request"
        empty_dir.mkdir()
        result = self.run_cli_in_cwd(empty_dir, "fix", "failing", "tests")
        self.assert_ok(result)
        self.assertIn("Working on: fix failing tests", result.stdout)
        self.assertIn("Run this from the repo or folder that needs work.", result.stdout)
        self.assertNotIn("Jini CLI 0.1.0", result.stdout)
        self.assertNotIn("REQUEST fix failing tests", result.stdout)
        self.assertNotIn("START HERE", result.stdout)
        self.assertNotIn("jini try-example research-prd", result.stdout)
        self.assertNotIn("jini doctor", result.stdout)
        self.assertNotIn("jini status /path/to/work", result.stdout)
        self.assertNotIn("usage: jini", result.stdout)

    def test_request_text_in_interactive_no_repo_mode_keeps_live_session_open(self) -> None:
        empty_dir = self.tmp / "generic-request-interactive"
        empty_dir.mkdir()
        result = self.run_cli_in_cwd(
            empty_dir,
            "fix",
            "failing",
            "tests",
            env={**os.environ, "JINI_FORCE_INTERACTIVE": "1"},
            input_text="plan this change\nexit\n",
        )
        self.assert_ok(result)
        self.assertIn("Working on: fix failing tests", result.stdout)
        self.assertIn("Run this from the repo or folder that needs work.", result.stdout)
        self.assertIn("jini> ", result.stdout)
        self.assertIn("Working on: plan this change", result.stdout)
        self.assertEqual(2, result.stdout.count("Run this from the repo or folder that needs work."))
        self.assertNotIn("Jini CLI 0.1.0", result.stdout)
        self.assertNotIn("Session closed.", result.stdout)

    def test_request_text_in_interactive_no_repo_mode_routes_escape_hatches_after_initial_request(self) -> None:
        empty_dir = self.tmp / "generic-request-interactive-escapes"
        empty_dir.mkdir()
        result = self.run_cli_in_cwd(
            empty_dir,
            "fix",
            "failing",
            "tests",
            env={**os.environ, "JINI_FORCE_INTERACTIVE": "1"},
            input_text="commands\ndoctor\nhelp --admin\nsetup --harness codex\nexit\n",
        )
        self.assert_ok(result)
        self.assertIn("Working on: fix failing tests", result.stdout)
        self.assertIn("Run this from the repo or folder that needs work.", result.stdout)
        self.assert_compact_interactive_escape_hatch_surface(result.stdout)

    def test_request_text_in_interactive_repo_mode_keeps_live_session_open(self) -> None:
        repo = self.create_repo_fixture()
        result = self.run_cli_in_cwd(
            repo,
            "fix",
            "failing",
            "tests",
            env={**os.environ, "JINI_FORCE_INTERACTIVE": "1"},
            input_text="plan this change\nexit\n",
        )
        self.assert_ok(result)
        self.assertIn("Working on: fix failing tests", result.stdout)
        self.assertIn("Start with `make test`.", result.stdout)
        self.assertNotIn("What next?", result.stdout)
        self.assertIn("jini> ", result.stdout)
        self.assertIn("Working on: plan this change", result.stdout)
        self.assertEqual(1, result.stdout.count("Start with `"))
        self.assertNotIn("README.md", result.stdout)
        self.assertNotIn("Jini CLI 0.1.0", result.stdout)
        self.assertNotIn("Repo: sample-repo", result.stdout)
        self.assertNotIn("Type a task. Use `exit` to leave.", result.stdout)
        self.assertNotIn("Session closed.", result.stdout)
        self.assertNotIn("usage: jini", result.stdout)
        self.assertNotIn("REQUEST fix failing tests", result.stdout)

    def test_request_text_in_interactive_repo_mode_routes_escape_hatches_after_initial_request(self) -> None:
        repo = self.create_repo_fixture()
        result = self.run_cli_in_cwd(
            repo,
            "fix",
            "failing",
            "tests",
            env={**os.environ, "JINI_FORCE_INTERACTIVE": "1"},
            input_text="commands\ndoctor\nhelp --admin\nsetup --harness codex\nexit\n",
        )
        self.assert_ok(result)
        self.assertIn("Working on: fix failing tests", result.stdout)
        self.assertIn("Start with `make test`.", result.stdout)
        self.assert_compact_interactive_escape_hatch_surface(result.stdout)

    def test_zero_arg_cli_interactive_repo_mode_uses_calm_empty_input_prompt(self) -> None:
        repo = self.create_repo_fixture()
        result = self.run_cli_in_cwd(
            repo,
            env={**os.environ, "JINI_FORCE_INTERACTIVE": "1"},
            input_text="\nexit\n",
        )
        self.assert_ok(result)
        self.assertNotIn("Type the next task, or `exit` to leave.", result.stdout)

    def test_prompt_repo_task_keeps_shell_alive_after_keyboard_interrupt(self) -> None:
        repo = self.create_repo_fixture()
        repo_context = jini_validate.inspect_repo_context(repo)
        output = io.StringIO()
        with mock.patch("builtins.input", side_effect=[KeyboardInterrupt(), "exit"]), redirect_stdout(output):
            result = jini_validate.prompt_repo_task(repo_context)
        transcript = output.getvalue()
        self.assertEqual(0, result)
        self.assertNotIn("Type a task. Use `exit` to leave.", transcript)
        self.assertNotIn("Type the next task, or `exit` to leave.", transcript)
        self.assertEqual(0, transcript.count("Session closed."))

    def test_prompt_current_work_task_keeps_shell_alive_after_keyboard_interrupt(self) -> None:
        repo = self.create_repo_fixture()
        output = io.StringIO()
        with (
            mock.patch("tools.jini_validate.Path.cwd", return_value=repo),
            mock.patch("builtins.input", side_effect=[KeyboardInterrupt(), "exit"]),
            redirect_stdout(output),
        ):
            result = jini_validate.prompt_current_work_task(self.tmp / "pack", {})
        transcript = output.getvalue()
        self.assertEqual(0, result)
        self.assertNotIn("Type the next task, or `exit` to leave.", transcript)
        self.assertEqual(0, transcript.count("Session closed."))

    def test_prompt_generic_task_keeps_shell_alive_after_keyboard_interrupt(self) -> None:
        output = io.StringIO()
        with mock.patch("builtins.input", side_effect=[KeyboardInterrupt(), "exit"]), redirect_stdout(output):
            result = jini_validate.prompt_generic_task()
        transcript = output.getvalue()
        self.assertEqual(0, result)
        self.assertNotIn("Type a task. Use `exit` to leave.", transcript)
        self.assertNotIn("Type the next task, or `exit` to leave.", transcript)
        self.assertEqual(0, transcript.count("Session closed."))

    def test_run_cli_converts_top_level_keyboard_interrupt_into_clean_cancel(self) -> None:
        stderr = io.StringIO()
        with mock.patch("tools.jini_validate.main", side_effect=KeyboardInterrupt), redirect_stdout(io.StringIO()), redirect_stderr(stderr):
            result = jini_validate.run_cli()
        self.assertEqual(130, result)
        self.assertEqual("\nCancelled.\n", stderr.getvalue())

    def test_run_cli_converts_broken_pipe_into_quiet_exit(self) -> None:
        stdout = io.StringIO()
        stderr = io.StringIO()
        with (
            mock.patch("tools.jini_validate.main", side_effect=BrokenPipeError),
            mock.patch("tools.jini_validate.suppress_broken_pipe_stdout") as suppress_stdout,
            redirect_stdout(stdout),
            redirect_stderr(stderr),
        ):
            result = jini_validate.run_cli()
        self.assertEqual(141, result)
        suppress_stdout.assert_called_once_with()
        self.assertEqual("", stdout.getvalue())
        self.assertEqual("", stderr.getvalue())

    def test_suppress_broken_pipe_stdout_redirects_stdout_to_devnull(self) -> None:
        fake_stdout = mock.Mock()
        fake_stdout.fileno.return_value = 7
        with (
            mock.patch("tools.jini_validate.sys.stdout", fake_stdout),
            mock.patch("tools.jini_validate.os.open", return_value=11) as open_devnull,
            mock.patch("tools.jini_validate.os.dup2") as dup2,
            mock.patch("tools.jini_validate.os.close") as close_fd,
        ):
            jini_validate.suppress_broken_pipe_stdout()
        open_devnull.assert_called_once_with(os.devnull, os.O_WRONLY)
        dup2.assert_called_once_with(11, 7)
        close_fd.assert_called_once_with(11)

    def test_interactive_repo_mode_uses_compact_follow_up_cards(self) -> None:
        repo = self.create_repo_fixture()
        result = self.run_cli_in_cwd(
            repo,
            env={**os.environ, "JINI_FORCE_INTERACTIVE": "1"},
            input_text="fix failing tests\nexit\n",
        )
        self.assert_ok(result)
        self.assertIn("Working on: fix failing tests", result.stdout)
        self.assertIn("Start with `make test`.", result.stdout)
        self.assertNotIn("ALSO    ", result.stdout)
        self.assertEqual(1, result.stdout.count("Start with `"))
        self.assertNotIn("Jini CLI 0.1.0", result.stdout)
        self.assertNotIn("BEST NEXT MOVE", result.stdout)
        self.assertNotIn("KEEP MOVING", result.stdout)
        self.assertNotIn("PATH   ", result.stdout)
        self.assertNotIn("GIT    ", result.stdout)
        self.assertNotIn("PATH    /", result.stdout)

    def test_unknown_single_token_suggests_closest_command(self) -> None:
        result = self.run_cli("stats")
        self.assert_error(result)
        self.assertEqual("", result.stdout)
        self.assertIn('ERROR Unknown command "stats".', result.stderr)
        self.assertIn("Closest matches:", result.stderr)
        self.assertIn("jini status", result.stderr)
        self.assertNotIn("usage: jini", result.stderr)

    def test_repo_path_token_shows_repo_specific_hint(self) -> None:
        result = self.run_cli(".")
        self.assert_error(result)
        self.assertEqual("", result.stdout)
        self.assertIn("is not a direct repo entrypoint", result.stderr)
        self.assertIn("jini review this repo", result.stderr)
        self.assertIn("jini status /path/to/work", result.stderr)
        self.assertNotIn("usage: jini", result.stderr)

    def test_help_admin_shows_internal_inventory(self) -> None:
        result = self.run_cli("help", "--admin")
        self.assert_ok(result)
        self.assertIn("Admin and developer command inventory", result.stdout)
        self.assertIn("skills / delegate", result.stdout)
        self.assertIn("stage-framework-experiment", result.stdout)
        self.assertIn("capture-evidence", result.stdout)
        self.assertNotIn("usage: jini", result.stdout)

    def test_admin_help_alias_shows_internal_inventory(self) -> None:
        result = self.run_cli("admin", "help")
        self.assert_ok(result)
        self.assertIn("Admin and developer command inventory", result.stdout)
        self.assertIn("skills / delegate", result.stdout)
        self.assertIn("stage-framework-experiment", result.stdout)
        self.assertIn("capture-evidence", result.stdout)

    def test_admin_help_with_request_tail_shows_corrective_hint(self) -> None:
        request_tokens = HELP_TAIL_EXAMPLE_REQUEST.split()
        result = self.run_cli("admin", "help", *request_tokens)
        self.assert_error(result)
        self.assertEqual(
            result.stderr.strip().splitlines(),
            list(help_tail_message_lines("jini", "admin help", "admin command inventory", request_tokens)),
        )
        self.assertEqual("", result.stdout)
        self.assertNotIn("Admin and developer command inventory", result.stderr)

    def test_provider_help_with_request_tail_shows_corrective_hint(self) -> None:
        request_tokens = HELP_TAIL_EXAMPLE_REQUEST.split()
        result = self.run_cli("provider", "help", *request_tokens)
        self.assert_error(result)
        self.assertEqual(
            result.stderr.strip().splitlines(),
            list(help_tail_message_lines("jini", "provider help", "admin command inventory", request_tokens)),
        )
        self.assertEqual("", result.stdout)
        self.assertNotIn("Admin and developer command inventory", result.stderr)

    def test_provider_dash_help_with_request_tail_shows_corrective_hint(self) -> None:
        request_tokens = HELP_TAIL_EXAMPLE_REQUEST.split()
        result = self.run_cli("provider", "--help", *request_tokens)
        self.assert_error(result)
        self.assertEqual(
            result.stderr.strip().splitlines(),
            list(help_tail_message_lines("jini", "provider --help", "admin command inventory", request_tokens)),
        )
        self.assertEqual("", result.stdout)
        self.assertNotIn("Admin and developer command inventory", result.stderr)

    def test_provider_namespace_defaults_to_provider_doctor_surface(self) -> None:
        result = self.run_cli("provider")
        self.assert_ok(result)
        self.assertIn("Provider", result.stdout)
        self.assertIn("Status", result.stdout)
        self.assertNotIn("Admin and developer command inventory", result.stdout)

    def test_provider_namespace_accepts_doctor_flags_directly(self) -> None:
        env = {
            "JINI_PROVIDER": "claude",
            "ANTHROPIC_API_KEY": "sk-live-secret",
            "JINI_MODEL": "sonnet",
        }
        result = self.run_cli("provider", "--format", "json", env=env)
        self.assert_ok(result)
        report = json.loads(result.stdout)
        self.assertEqual("anthropic", report["provider_id"])

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

    def test_skills_lists_builtin_catalog(self) -> None:
        result = self.run_cli("skills", "--format", "json")
        self.assert_ok(result)

        payload = json.loads(result.stdout)
        by_id = {item["skill_id"]: item for item in payload["skills"]}
        self.assertIn("reviewer", by_id)
        self.assertIn("debugger", by_id)
        self.assertIn("research", by_id)
        self.assertEqual("builtin", by_id["reviewer"]["scope"])
        self.assertTrue(by_id["reviewer"]["enabled"])

    def test_skills_filter_prefers_project_override(self) -> None:
        repo = self.create_repo_fixture()
        self.write_skill(
            repo / ".jini" / "skills",
            "reviewer",
            label="Team Reviewer",
            purpose="Review using the repo's local standards.",
        )

        result = self.run_cli_in_cwd(repo, "skills", "reviewer", "--format", "json")
        self.assert_ok(result)

        payload = json.loads(result.stdout)
        self.assertEqual(1, len(payload["skills"]))
        self.assertEqual("reviewer", payload["skills"][0]["skill_id"])
        self.assertEqual("Team Reviewer", payload["skills"][0]["label"])
        self.assertEqual("project", payload["skills"][0]["scope"])

    def test_delegate_requires_current_work(self) -> None:
        result = self.run_cli("delegate", "reviewer")
        self.assert_error(result)
        self.assertEqual("", result.stdout)
        self.assertIn("Current work is required for delegate reviewer.", result.stderr)

    def test_delegate_unknown_skill_returns_clean_error(self) -> None:
        pack_dir = self.compile_research_pack()

        result = self.run_cli("delegate", "reviewerx", pack_dir)
        self.assert_error(result)
        self.assertEqual("", result.stdout)
        self.assertIn("Unknown skill 'reviewerx'.", result.stderr)

    def test_delegate_invalid_project_skill_fails_cleanly(self) -> None:
        repo = self.create_repo_fixture()
        self.write_skill(
            repo / ".jini" / "skills",
            "reviewer",
            invalid_missing_output_contract=True,
        )
        pack_dir = self.compile_research_pack()

        result = self.run_cli_in_cwd(repo, "delegate", "reviewer", pack_dir)
        self.assert_error(result)
        self.assertEqual("", result.stdout)
        self.assertIn("Skill 'reviewer' is invalid: missing output_contract.", result.stderr)

    def test_delegate_succeeds_with_current_work_and_records_state(self) -> None:
        pack_dir = self.compile_research_pack()
        self.assert_ok(self.run_cli("status", pack_dir))

        result = self.run_cli("delegate", "reviewer", "--instruction", "Focus on the biggest risks.", "--format", "json")
        self.assert_ok(result)

        payload = json.loads(result.stdout)
        self.assertEqual("reviewer", payload["skill_id"])
        self.assertEqual("staged", payload["status"])
        request_path = Path(payload["request_path"])
        result_path = Path(payload["result_path"])
        summary_path = Path(payload["summary_path"])
        thread_state_path = Path(payload["thread_state_path"])
        for path in (request_path, result_path, summary_path, thread_state_path):
            self.assertTrue(path.exists(), str(path))
        request_payload = json.loads(request_path.read_text(encoding="utf-8"))
        self.assertEqual(payload["delegation_id"], request_payload["delegation_id"])
        self.assertEqual("reviewer", request_payload["skill_id"])
        self.assertEqual("Focus on the biggest risks.", request_payload["instruction"])
        result_payload = json.loads(result_path.read_text(encoding="utf-8"))
        self.assertEqual(payload["delegation_id"], result_payload["delegation_id"])
        self.assertEqual(str(summary_path), result_payload["summary_path"])
        thread_state = json.loads(thread_state_path.read_text(encoding="utf-8"))
        self.assertEqual(payload["delegation_id"], thread_state["active_delegation_id"])
        self.assertEqual("reviewer", thread_state["active_skill_id"])
        self.assertEqual("staged", thread_state["active_delegation_status"])
        current_work = json.loads((self.tmp / ".jini" / "current-work.json").read_text(encoding="utf-8"))
        self.assertEqual(str(pack_dir.resolve()), current_work["pack_dir"])

    def test_delegate_updates_status_continue_show_and_resume_focus(self) -> None:
        pack_dir = self.compile_research_pack()
        self.assert_ok(self.run_cli("status", pack_dir))

        delegate_result = self.run_cli("delegate", "reviewer", "--instruction", "Focus on risks.", "--format", "json")
        self.assert_ok(delegate_result)
        delegation = json.loads(delegate_result.stdout)
        summary_path = Path(delegation["summary_path"])

        status_result = self.run_cli("status", "--format", "json")
        self.assert_ok(status_result)
        status_payload = json.loads(status_result.stdout)
        self.assertEqual("reviewer-brief", status_payload["current_focus"]["id"])
        self.assertEqual("Reviewer delegation brief", status_payload["current_focus"]["label"])
        self.assertEqual("jini continue", status_payload["current_focus"]["show_command"])

        continue_result = self.run_cli("continue", "--print-path")
        self.assert_ok(continue_result)
        self.assertEqual(display := str(summary_path.relative_to(REPO_ROOT)) if summary_path.is_relative_to(REPO_ROOT) else str(summary_path), continue_result.stdout.strip())

        show_result = self.run_cli("show")
        self.assert_ok(show_result)
        self.assertIn("# Delegation: Reviewer", show_result.stdout)
        self.assertIn("## Instruction", show_result.stdout)

        open_result = self.run_cli("open", "--print-path")
        self.assert_ok(open_result)
        self.assertEqual(display, open_result.stdout.strip())

        open_json_result = self.run_cli("open", "--print-path", "--format", "json")
        self.assert_ok(open_json_result)
        open_payload = json.loads(open_json_result.stdout)
        self.assertEqual("reviewer-brief", open_payload["artifact"]["id"])
        self.assertEqual("Reviewer delegation brief", open_payload["artifact"]["label"])
        self.assertEqual(display, open_payload["path"])

        resume_result = self.run_cli("resume", "--format", "json")
        self.assert_ok(resume_result)
        resume_payload = json.loads(resume_result.stdout)
        self.assertEqual("reviewer-brief", resume_payload["current_focus"]["id"])
        self.assertEqual("Reviewer delegation brief", resume_payload["current_focus"]["label"])
        self.assertEqual("Reviewer delegation brief", resume_payload["recent_artifacts"][0]["artifact_type"])
        self.assertTrue(any("Focused delegation brief: `Reviewer delegation brief`" in item for item in resume_payload["resume_items"]))

        skills_result = self.run_cli("skills", "--format", "json")
        self.assert_ok(skills_result)
        skills_payload = json.loads(skills_result.stdout)
        self.assertEqual("reviewer", skills_payload["skills"][0]["skill_id"])
        feedback = skills_payload["skills"][0]["feedback"]
        self.assertEqual(1, feedback["repo_signals"]["delegated"])
        self.assertEqual(1, feedback["repo_signals"]["continued"])
        self.assertEqual(1, feedback["repo_signals"]["shown"])
        self.assertEqual(2, feedback["repo_signals"]["opened"])
        self.assertGreater(feedback["repo_score"], 0)

    def test_skills_feedback_stays_repo_specific(self) -> None:
        repo_a = self.create_repo_fixture()
        repo_b = self.tmp / "other-repo"
        shutil.copytree(repo_a, repo_b)
        pack_dir = self.compile_research_pack()

        self.assert_ok(self.run_cli_in_cwd(repo_a, "status", pack_dir))
        delegate_result = self.run_cli_in_cwd(repo_a, "delegate", "reviewer", "--instruction", "Focus on risks.", "--format", "json")
        self.assert_ok(delegate_result)

        repo_a_skills = self.run_cli_in_cwd(repo_a, "skills", "reviewer", "--format", "json")
        self.assert_ok(repo_a_skills)
        repo_a_payload = json.loads(repo_a_skills.stdout)
        repo_a_feedback = repo_a_payload["skills"][0]["feedback"]
        self.assertEqual(1, repo_a_feedback["repo_signals"]["delegated"])
        self.assertGreater(repo_a_feedback["repo_score"], 0)

        repo_b_skills = self.run_cli_in_cwd(repo_b, "skills", "reviewer", "--format", "json")
        self.assert_ok(repo_b_skills)
        repo_b_payload = json.loads(repo_b_skills.stdout)
        repo_b_feedback = repo_b_payload["skills"][0]["feedback"]
        self.assertEqual({}, repo_b_feedback["repo_signals"])
        self.assertEqual(0, repo_b_feedback["repo_score"])
        self.assertEqual(1, repo_b_feedback["global_signals"]["delegated"])
        self.assertGreater(repo_b_feedback["global_score"], 0)

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

        resume_result = self.run_cli("resume", pack_dir, "--format", "json", "--max-chars", "900")
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

    def test_status_and_resume_surface_efficiency_posture(self) -> None:
        pack_dir = self.compile_research_pack()

        status_text = self.run_cli("status", pack_dir)
        self.assert_ok(status_text)
        self.assertIn("EFFICIENCY", status_text.stdout)
        self.assertIn("class=standard", status_text.stdout)
        self.assertIn("context=targeted", status_text.stdout)
        self.assertIn("RUNTIME", status_text.stdout)
        self.assertIn("route=", status_text.stdout)
        self.assertIn("model=", status_text.stdout)
        self.assertIn("effort=standard", status_text.stdout)
        self.assertIn("reason=", status_text.stdout)

        status_json = self.run_cli("status", pack_dir, "--format", "json")
        self.assert_ok(status_json)
        report = json.loads(status_json.stdout)
        posture = report["efficiency_posture"]
        self.assertEqual("make", posture["intent"])
        self.assertEqual("standard", posture["execution_class"])
        self.assertEqual("targeted", posture["context_policy"])
        self.assertTrue(posture["selected_runtime"])
        self.assertTrue(posture["rationale"])
        runtime = report["runtime_readout"]
        self.assertEqual("auto", runtime["selection_mode"])
        self.assertEqual(posture["selected_runtime"], runtime["route"])
        self.assertEqual(runtime["route"], runtime["runtime_target"])
        self.assertTrue(runtime["model"])
        self.assertEqual("standard", runtime["effort"])
        self.assertEqual("targeted", runtime["context_policy"])
        self.assertIn("make", runtime["reason"])

        resume_json = self.run_cli(
            "resume",
            pack_dir,
            "--intent",
            "export",
            "--format",
            "json",
            "--max-chars",
            "900",
        )
        self.assert_ok(resume_json)
        compact = json.loads(resume_json.stdout)
        resume_posture = compact["efficiency_posture"]
        self.assertEqual("cheap", resume_posture["execution_class"])
        self.assertTrue(resume_posture["cheap_path"])
        self.assertTrue(any("export" in item for item in resume_posture["rationale"]))
        resume_runtime = compact["runtime_readout"]
        self.assertEqual("auto", resume_runtime["selection_mode"])
        self.assertEqual(resume_posture["selected_runtime"], resume_runtime["route"])
        self.assertEqual("cheap", resume_runtime["effort"])
        self.assertIn("export", resume_runtime["reason"])
        self.assertLessEqual(compact["token_budget"]["estimated_chars"], 900)

    def test_runtime_readouts_surface_measured_local_route_evidence(self) -> None:
        pack_dir = self.compile_research_pack()
        state_root = self.tmp / ".jini"
        state_root.mkdir(parents=True, exist_ok=True)
        (state_root / "local-runtime-capabilities.json").write_text(
            json.dumps(
                {
                    "schema_version": "0.4.0",
                    "context_type": "JiniLocalRuntimeCapabilities",
                    "captured_at": "2026-05-21T19:00:00Z",
                    "local_runtime_class": "local-ollama",
                    "adapters": {
                        "local-workhorse": {
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
                }
            ),
            encoding="utf-8",
        )

        recommendation_result = self.run_cli("recommend-execution", pack_dir, "--format", "json")
        self.assert_ok(recommendation_result)
        recommendation = json.loads(recommendation_result.stdout)
        self.assertTrue(recommendation["route_evidence"]["available"])
        self.assertEqual("local-ollama", recommendation["route_evidence"]["local_runtime_class"])
        self.assertEqual("measured", recommendation["route_cost"]["status"])
        self.assertEqual("local-fast", recommendation["route_cost"]["cheapest_ready_adapter"]["adapter_id"])
        execution_route = recommendation["runtime_guidance"]["execution_route"]
        selected_runtime_target = recommendation["runtime_guidance"]["selected"]["id"]
        self.assertEqual("local-workhorse", execution_route["selected"]["id"])
        self.assertEqual("measured-local-runtime", execution_route["selection_basis"])
        self.assertEqual("local-fast", execution_route["cheapest_ready_adapter"])
        self.assertNotIn("local-fast", execution_route["fallbacks"])
        self.assertIn("codex", execution_route["fallbacks"])

        status_result = self.run_cli("status", pack_dir, "--format", "json")
        self.assert_ok(status_result)
        status_payload = json.loads(status_result.stdout)
        self.assertEqual("local-workhorse", status_payload["efficiency_posture"]["selected_runtime"])
        status_readout = status_payload["runtime_readout"]
        self.assertEqual("local-workhorse", status_readout["route"])
        self.assertEqual(selected_runtime_target, status_readout["runtime_target"])
        self.assertEqual("local-ollama", status_readout["route_evidence"]["local_runtime_class"])
        self.assertEqual(2, status_readout["route_evidence"]["ready_adapter_count"])
        self.assertEqual("measured", status_readout["route_evidence"]["cost_status"])
        self.assertEqual("zero-external-api-spend", status_readout["route_evidence"]["cost_posture"])
        self.assertEqual("local-fast", status_readout["route_evidence"]["cheapest_ready_adapter"])
        self.assertEqual("local-workhorse", status_readout["route_evidence"]["selected_ready_adapter"])
        self.assertIn("local-workhorse", status_readout["reason"])

        resume_result = self.run_cli(
            "resume",
            pack_dir,
            "--intent",
            "export",
            "--format",
            "json",
            "--max-chars",
            "1100",
        )
        self.assert_ok(resume_result)
        resume_payload = json.loads(resume_result.stdout)
        self.assertEqual("local-fast", resume_payload["efficiency_posture"]["selected_runtime"])
        resume_readout = resume_payload["runtime_readout"]
        self.assertEqual("local-fast", resume_readout["route"])
        self.assertEqual("local-ollama", resume_readout["route_evidence"]["local_runtime_class"])
        self.assertEqual("local-fast", resume_readout["route_evidence"]["cheapest_ready_adapter"])
        self.assertEqual("local-fast", resume_readout["route_evidence"]["selected_ready_adapter"])

    def test_runtime_readouts_explain_when_no_ready_local_route_is_available(self) -> None:
        pack_dir = self.compile_research_pack()
        state_root = self.tmp / ".jini"
        state_root.mkdir(parents=True, exist_ok=True)
        (state_root / "local-runtime-capabilities.json").write_text(
            json.dumps(
                {
                    "schema_version": "0.4.0",
                    "context_type": "JiniLocalRuntimeCapabilities",
                    "captured_at": "2026-05-21T19:00:00Z",
                    "local_runtime_class": "local-ollama",
                    "adapters": {
                        "local-workhorse": {
                            "status": "offline",
                            "latency_ms": 0,
                            "warm_latency_ms": 0,
                            "cold_start_cost_ms": 0,
                            "tokens_per_second": 0,
                            "quality_class": "strong",
                            "structured_reliability": "strong",
                            "benchmarked_at": "2026-05-21T19:00:00Z",
                        },
                        "local-fast": {
                            "status": "blocked",
                            "latency_ms": 0,
                            "warm_latency_ms": 0,
                            "cold_start_cost_ms": 0,
                            "tokens_per_second": 0,
                            "quality_class": "usable",
                            "structured_reliability": "usable",
                            "benchmarked_at": "2026-05-21T19:00:00Z",
                        },
                    },
                }
            ),
            encoding="utf-8",
        )

        recommendation_result = self.run_cli("recommend-execution", pack_dir, "--format", "json")
        self.assert_ok(recommendation_result)
        recommendation = json.loads(recommendation_result.stdout)
        self.assertTrue(recommendation["route_evidence"]["available"])
        self.assertEqual("local-ollama", recommendation["route_evidence"]["local_runtime_class"])
        self.assertEqual("unavailable", recommendation["route_cost"]["status"])
        self.assertEqual("unknown", recommendation["route_cost"]["posture"])
        self.assertEqual({}, recommendation["runtime_guidance"].get("execution_route", {}))
        self.assertTrue(
            any(
                "no ready local route is available for this task, so Jini must leave the device-local path."
                in note
                for note in recommendation["rationale"]
            )
        )

        status_result = self.run_cli("status", pack_dir, "--format", "json")
        self.assert_ok(status_result)
        status_payload = json.loads(status_result.stdout)
        status_readout = status_payload["runtime_readout"]
        self.assertEqual("local-ollama", status_readout["route_evidence"]["local_runtime_class"])
        self.assertEqual(0, status_readout["route_evidence"]["ready_adapter_count"])
        self.assertEqual("unavailable", status_readout["route_evidence"]["cost_status"])
        self.assertIn(
            "no ready local route is available for this task, so Jini must leave the device-local path.",
            status_readout["reason"],
        )

        resume_result = self.run_cli(
            "resume",
            pack_dir,
            "--intent",
            "export",
            "--format",
            "json",
            "--max-chars",
            "1100",
        )
        self.assert_ok(resume_result)
        resume_payload = json.loads(resume_result.stdout)
        resume_readout = resume_payload["runtime_readout"]
        self.assertEqual("local-ollama", resume_readout["route_evidence"]["local_runtime_class"])
        self.assertEqual(0, resume_readout["route_evidence"]["ready_adapter_count"])
        self.assertEqual("unavailable", resume_readout["route_evidence"]["cost_status"])
        self.assertTrue(
            resume_readout["reason"].startswith(
                "Measured local runtime `local-ollama` has 0/2 ready adapter(s); no ready local route is avail"
            )
        )

    def test_runtime_readouts_explain_when_ready_local_adapters_miss_threshold(self) -> None:
        pack_dir = self.compile_research_pack()
        state_root = self.tmp / ".jini"
        state_root.mkdir(parents=True, exist_ok=True)
        (state_root / "local-runtime-capabilities.json").write_text(
            json.dumps(
                {
                    "schema_version": "0.4.0",
                    "context_type": "JiniLocalRuntimeCapabilities",
                    "captured_at": "2026-05-21T19:00:00Z",
                    "local_runtime_class": "local-ollama",
                    "adapters": {
                        "local-workhorse": {
                            "status": "ok",
                            "latency_ms": 180,
                            "warm_latency_ms": 150,
                            "cold_start_cost_ms": 30,
                            "tokens_per_second": 22.5,
                            "quality_class": "usable",
                            "structured_reliability": "usable",
                            "benchmarked_at": "2026-05-21T19:00:00Z",
                        },
                        "local-fast": {
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
                }
            ),
            encoding="utf-8",
        )

        recommendation_result = self.run_cli("recommend-execution", pack_dir, "--intent", "make", "--format", "json")
        self.assert_ok(recommendation_result)
        recommendation = json.loads(recommendation_result.stdout)
        self.assertTrue(recommendation["route_evidence"]["available"])
        self.assertEqual("local-ollama", recommendation["route_evidence"]["local_runtime_class"])
        self.assertEqual("measured", recommendation["route_cost"]["status"])
        self.assertEqual({}, recommendation["runtime_guidance"].get("execution_route", {}))
        self.assertTrue(
            any(
                "but none meet the `strong` quality threshold for this task, so Jini must leave the device-local path."
                in note
                for note in recommendation["rationale"]
            )
        )

        status_result = self.run_cli("status", pack_dir, "--format", "json")
        self.assert_ok(status_result)
        status_payload = json.loads(status_result.stdout)
        status_readout = status_payload["runtime_readout"]
        self.assertEqual("local-ollama", status_readout["route_evidence"]["local_runtime_class"])
        self.assertEqual(2, status_readout["route_evidence"]["ready_adapter_count"])
        self.assertEqual("measured", status_readout["route_evidence"]["cost_status"])
        self.assertIn(
            "but none meet the `strong` quality threshold for this task, so Jini must leave the device-local path.",
            status_readout["reason"],
        )

        resume_result = self.run_cli(
            "resume",
            pack_dir,
            "--intent",
            "make",
            "--format",
            "json",
            "--max-chars",
            "1100",
        )
        self.assert_ok(resume_result)
        resume_payload = json.loads(resume_result.stdout)
        resume_readout = resume_payload["runtime_readout"]
        self.assertEqual("local-ollama", resume_readout["route_evidence"]["local_runtime_class"])
        self.assertEqual(2, resume_readout["route_evidence"]["ready_adapter_count"])
        self.assertEqual("measured", resume_readout["route_evidence"]["cost_status"])
        self.assertTrue(
            resume_readout["reason"].startswith(
                "Measured local runtime `local-ollama` has 2/2 ready adapter(s), but none meet the `strong` qu"
            )
        )

    def test_runtime_readouts_keep_execution_escalation_reason_with_local_threshold_fallback(self) -> None:
        pack_dir = self.compile_research_pack()
        work_unit_path = pack_dir / "work-unit.yaml"
        work_unit_text = work_unit_path.read_text(encoding="utf-8")
        self.assertIn("current_state: decided", work_unit_text)
        work_unit_path.write_text(
            work_unit_text.replace("current_state: decided", "current_state: operational"),
            encoding="utf-8",
        )
        state_root = self.tmp / ".jini"
        state_root.mkdir(parents=True, exist_ok=True)
        (state_root / "local-runtime-capabilities.json").write_text(
            json.dumps(
                {
                    "schema_version": "0.4.0",
                    "context_type": "JiniLocalRuntimeCapabilities",
                    "captured_at": "2026-05-21T19:00:00Z",
                    "local_runtime_class": "local-ollama",
                    "adapters": {
                        "local-workhorse": {
                            "status": "ok",
                            "latency_ms": 180,
                            "warm_latency_ms": 150,
                            "cold_start_cost_ms": 30,
                            "tokens_per_second": 22.5,
                            "quality_class": "usable",
                            "structured_reliability": "usable",
                            "benchmarked_at": "2026-05-21T19:00:00Z",
                        },
                        "local-fast": {
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
                }
            ),
            encoding="utf-8",
        )

        status_result = self.run_cli("status", pack_dir, "--format", "json")
        self.assert_ok(status_result)
        status_payload = json.loads(status_result.stdout)
        self.assertEqual("deep", status_payload["efficiency_posture"]["execution_class"])
        self.assertEqual(
            "State `operational` requires stronger verification posture; "
            "Measured local runtime `local-ollama` has 2/2 ready adapter(s), but none meet the `strong` "
            "quality threshold for this task, so Jini must leave the device-local path.",
            status_payload["runtime_readout"]["reason"],
        )

        resume_result = self.run_cli(
            "resume",
            pack_dir,
            "--format",
            "json",
            "--max-chars",
            "1200",
        )
        self.assert_ok(resume_result)
        resume_payload = json.loads(resume_result.stdout)
        self.assertEqual("deep", resume_payload["efficiency_posture"]["execution_class"])
        self.assertTrue(
            resume_payload["runtime_readout"]["reason"].startswith(
                "State `operational` requires stronger verification posture; Measured local runtime `local-oll..."
            )
        )

    def test_runtime_readouts_surface_managed_runtime_target_fallback_reason(self) -> None:
        pack_dir = self.compile_research_pack()
        work_unit_path = pack_dir / "work-unit.yaml"
        work_unit_text = work_unit_path.read_text(encoding="utf-8")
        self.assertIn("current_state: decided", work_unit_text)
        work_unit_path.write_text(
            work_unit_text.replace("current_state: decided", "current_state: operational"),
            encoding="utf-8",
        )
        rollout_dir = pack_dir / "runtime" / "policy-rollouts"
        rollout_dir.mkdir(parents=True, exist_ok=True)
        (rollout_dir / "runtime-routing-active.json").write_text(
            json.dumps(
                {
                    "schema_version": "0.1.0",
                    "policy_type": "JiniPolicyRollout",
                    "policy_id": "runtime-routing",
                    "candidate_id": "runtime-readout-managed-fallback-test",
                    "status": "active",
                    "recommended_runtime_target": "kiro-cli",
                    "intent_overrides": {},
                    "route_feedback_drivers": {},
                },
                indent=2,
            )
            + "\n",
            encoding="utf-8",
        )
        env = os.environ.copy()
        env["JINI_RUNTIME_TARGET_HEALTH"] = "kiro-cli=throttled"

        status_result = self.run_cli("status", pack_dir, "--format", "json", env=env)
        self.assert_ok(status_result)
        status_payload = json.loads(status_result.stdout)
        self.assertEqual("deep", status_payload["efficiency_posture"]["execution_class"])
        self.assertEqual(
            "State `operational` requires stronger verification posture; "
            "Policy-preferred adapter `kiro-cli` is `throttled`; switched to `codex` for a healthier runtime target from env.",
            status_payload["runtime_readout"]["reason"],
        )

        resume_result = self.run_cli(
            "resume",
            pack_dir,
            "--format",
            "json",
            "--max-chars",
            "1200",
            env=env,
        )
        self.assert_ok(resume_result)
        resume_payload = json.loads(resume_result.stdout)
        self.assertEqual("deep", resume_payload["efficiency_posture"]["execution_class"])
        self.assertEqual(
            "State `operational` requires stronger verification posture; "
            "Policy-preferred adapter `kiro-cli` is `throttled`; switched to `codex` for a healthier runtime target from env.",
            resume_payload["runtime_readout"]["reason"],
        )

    def test_runtime_readouts_surface_no_healthier_managed_target_reason(self) -> None:
        pack_dir = self.compile_research_pack()
        work_unit_path = pack_dir / "work-unit.yaml"
        work_unit_text = work_unit_path.read_text(encoding="utf-8")
        self.assertIn("current_state: decided", work_unit_text)
        work_unit_path.write_text(
            work_unit_text.replace("current_state: decided", "current_state: operational"),
            encoding="utf-8",
        )
        rollout_dir = pack_dir / "runtime" / "policy-rollouts"
        rollout_dir.mkdir(parents=True, exist_ok=True)
        (rollout_dir / "runtime-routing-active.json").write_text(
            json.dumps(
                {
                    "schema_version": "0.1.0",
                    "policy_type": "JiniPolicyRollout",
                    "policy_id": "runtime-routing",
                    "candidate_id": "runtime-readout-no-healthier-target-test",
                    "status": "active",
                    "recommended_runtime_target": "kiro-cli",
                    "intent_overrides": {},
                    "route_feedback_drivers": {},
                },
                indent=2,
            )
            + "\n",
            encoding="utf-8",
        )
        env = os.environ.copy()
        env["JINI_RUNTIME_TARGET_HEALTH"] = ",".join(
            [
                "kiro-cli=degraded",
                "codex=unavailable",
                "claude-code=unavailable",
                "github-copilot=unavailable",
                "junie=unavailable",
                "augment=unavailable",
            ]
        )

        expected_reason = (
            "State `operational` requires stronger verification posture; "
            "Policy-preferred adapter `kiro-cli` remained `degraded`; kept it as the selected runtime target.; "
            "Selected runtime target `kiro-cli` despite `degraded` status because no healthier higher-priority target was available."
        )

        status_result = self.run_cli("status", pack_dir, "--format", "json", env=env)
        self.assert_ok(status_result)
        status_payload = json.loads(status_result.stdout)
        self.assertEqual("deep", status_payload["efficiency_posture"]["execution_class"])
        self.assertEqual(expected_reason, status_payload["runtime_readout"]["reason"])

        resume_result = self.run_cli(
            "resume",
            pack_dir,
            "--format",
            "json",
            "--max-chars",
            "1200",
            env=env,
        )
        self.assert_ok(resume_result)
        resume_payload = json.loads(resume_result.stdout)
        self.assertEqual("deep", resume_payload["efficiency_posture"]["execution_class"])
        self.assertTrue(
            resume_payload["runtime_readout"]["reason"].startswith(
                "State `operational` requires stronger verification posture; Policy-preferred adapter `kiro-cl..."
            )
        )

    def test_runtime_readouts_surface_explicit_unhealthy_runtime_target_pin(self) -> None:
        pack_dir = self.compile_research_pack()
        work_unit_path = pack_dir / "work-unit.yaml"
        work_unit_text = work_unit_path.read_text(encoding="utf-8")
        self.assertIn("current_state: decided", work_unit_text)
        work_unit_path.write_text(
            work_unit_text.replace("current_state: decided", "current_state: operational"),
            encoding="utf-8",
        )
        env = os.environ.copy()
        env["JINI_RUNTIME_TARGET_HEALTH"] = "kiro-cli=throttled"
        expected_reason = (
            "State `operational` requires stronger verification posture; "
            "Preferred adapter `kiro-cli` was selected explicitly despite `throttled` status."
        )

        status_result = self.run_cli(
            "status",
            pack_dir,
            "--runtime-target",
            "kiro-cli",
            "--format",
            "json",
            env=env,
        )
        self.assert_ok(status_result)
        status_payload = json.loads(status_result.stdout)
        self.assertEqual("deep", status_payload["efficiency_posture"]["execution_class"])
        self.assertEqual(expected_reason, status_payload["runtime_readout"]["reason"])

        resume_result = self.run_cli(
            "resume",
            pack_dir,
            "--runtime-target",
            "kiro-cli",
            "--format",
            "json",
            "--max-chars",
            "1200",
            env=env,
        )
        self.assert_ok(resume_result)
        resume_payload = json.loads(resume_result.stdout)
        self.assertEqual("deep", resume_payload["efficiency_posture"]["execution_class"])
        self.assertEqual(expected_reason, resume_payload["runtime_readout"]["reason"])

    def test_runtime_readouts_keep_unhealthy_explicit_pin_with_local_route_evidence(self) -> None:
        pack_dir = self.compile_research_pack()
        work_unit_path = pack_dir / "work-unit.yaml"
        work_unit_text = work_unit_path.read_text(encoding="utf-8")
        self.assertIn("current_state: decided", work_unit_text)
        work_unit_path.write_text(
            work_unit_text.replace("current_state: decided", "current_state: operational"),
            encoding="utf-8",
        )
        state_root = self.tmp / ".jini"
        state_root.mkdir(parents=True, exist_ok=True)
        (state_root / "local-runtime-capabilities.json").write_text(
            json.dumps(
                {
                    "schema_version": "0.4.0",
                    "context_type": "JiniLocalRuntimeCapabilities",
                    "captured_at": "2026-05-21T19:00:00Z",
                    "local_runtime_class": "local-ollama",
                    "adapters": {
                        "local-workhorse": {
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
                            "status": "ok",
                            "latency_ms": 95,
                            "warm_latency_ms": 90,
                            "cold_start_cost_ms": 5,
                            "tokens_per_second": 38.2,
                            "quality_class": "usable",
                            "structured_reliability": "usable",
                            "benchmarked_at": "2026-05-21T19:00:00Z",
                        },
                    },
                }
            ),
            encoding="utf-8",
        )
        env = os.environ.copy()
        env["JINI_HOME"] = str(state_root)
        env["JINI_RUNTIME_TARGET_HEALTH"] = "kiro-cli=throttled"
        expected_reason = (
            "State `operational` requires stronger verification posture; "
            "Preferred adapter `kiro-cli` was selected explicitly despite `throttled` status.; "
            "Measured local runtime `local-ollama` selects `local-workhorse` with `strong` quality threshold; "
            "fallbacks: kiro-cli, codex, claude-code."
        )

        status_result = self.run_cli(
            "status",
            pack_dir,
            "--runtime-target",
            "kiro-cli",
            "--format",
            "json",
            env=env,
        )
        self.assert_ok(status_result)
        status_payload = json.loads(status_result.stdout)
        self.assertEqual("local-workhorse", status_payload["runtime_readout"]["route"])
        self.assertEqual("kiro-cli", status_payload["runtime_readout"]["runtime_target"])
        self.assertEqual(expected_reason, status_payload["runtime_readout"]["reason"])

        resume_result = self.run_cli(
            "resume",
            pack_dir,
            "--runtime-target",
            "kiro-cli",
            "--format",
            "json",
            "--max-chars",
            "1200",
            env=env,
        )
        self.assert_ok(resume_result)
        resume_payload = json.loads(resume_result.stdout)
        self.assertEqual("local-workhorse", resume_payload["runtime_readout"]["route"])
        self.assertEqual("kiro-cli", resume_payload["runtime_target"]["selected"])
        self.assertTrue(
            resume_payload["runtime_readout"]["reason"].startswith(
                "State `operational` requires stronger verification posture; Preferred adapter `kiro-cli` was ..."
            )
        )

        tight_resume_result = self.run_cli(
            "resume",
            pack_dir,
            "--runtime-target",
            "kiro-cli",
            "--format",
            "json",
            "--max-chars",
            "900",
            env=env,
        )
        self.assert_ok(tight_resume_result)
        tight_resume_payload = json.loads(tight_resume_result.stdout)
        self.assertEqual("local-workhorse", tight_resume_payload["runtime_readout"]["route"])
        self.assertEqual("kiro-cli", tight_resume_payload["runtime_target"]["selected"])
        self.assertEqual("local-workhorse", tight_resume_payload["efficiency_posture"]["selected_runtime"])
        self.assertTrue(
            tight_resume_payload["runtime_readout"]["reason"].startswith(
                "State `operational` requires stronger verification posture; Preferred adapter `kiro-cli` was ..."
            )
        )
        self.assertNotIn("route_evidence", tight_resume_payload["runtime_readout"])
        self.assertLessEqual(tight_resume_payload["token_budget"]["estimated_chars"], 900)

        tighter_resume_result = self.run_cli(
            "resume",
            pack_dir,
            "--runtime-target",
            "kiro-cli",
            "--format",
            "json",
            "--max-chars",
            "750",
            env=env,
        )
        self.assert_ok(tighter_resume_result)
        tighter_resume_payload = json.loads(tighter_resume_result.stdout)
        self.assertEqual("local-workhorse", tighter_resume_payload["runtime_readout"]["route"])
        self.assertEqual("kiro-cli", tighter_resume_payload["runtime_target"]["selected"])
        self.assertTrue(
            tighter_resume_payload["runtime_readout"]["reason"].startswith(
                "State `operational` requires stronger verification posture; Preferred adapter `kiro-cli` was ..."
            )
        )
        self.assertLessEqual(tighter_resume_payload["token_budget"]["estimated_chars"], 750)

        minimal_resume_result = self.run_cli(
            "resume",
            pack_dir,
            "--runtime-target",
            "kiro-cli",
            "--format",
            "json",
            "--max-chars",
            "600",
            env=env,
        )
        self.assert_ok(minimal_resume_result)
        minimal_resume_payload = json.loads(minimal_resume_result.stdout)
        self.assertEqual("local-workhorse", minimal_resume_payload["runtime_readout"]["route"])
        self.assertEqual("kiro-cli", minimal_resume_payload["runtime_target"]["selected"])
        self.assertTrue(
            minimal_resume_payload["runtime_readout"]["reason"].startswith(
                "State `operational` requires stronger verification posture; Preferred adapter `kiro-cli` was ..."
            )
        )
        self.assertLessEqual(minimal_resume_payload["token_budget"]["estimated_chars"], 600)

        micro_resume_result = self.run_cli(
            "resume",
            pack_dir,
            "--runtime-target",
            "kiro-cli",
            "--format",
            "json",
            "--max-chars",
            "475",
            env=env,
        )
        self.assert_ok(micro_resume_result)
        micro_resume_payload = json.loads(micro_resume_result.stdout)
        self.assertEqual("local-workhorse", micro_resume_payload["runtime_readout"]["route"])
        self.assertEqual("kiro-cli", micro_resume_payload["runtime_target"]["selected"])
        self.assertTrue(
            micro_resume_payload["runtime_readout"]["reason"].startswith(
                "State `operational` requires stronger verification posture; Preferred adapter `kiro-cli` was ..."
            )
        )
        self.assertLessEqual(micro_resume_payload["token_budget"]["estimated_chars"], 475)

        nano_resume_result = self.run_cli(
            "resume",
            pack_dir,
            "--runtime-target",
            "kiro-cli",
            "--format",
            "json",
            "--max-chars",
            "375",
            env=env,
        )
        self.assert_ok(nano_resume_result)
        nano_resume_payload = json.loads(nano_resume_result.stdout)
        self.assertEqual("local-workhorse", nano_resume_payload["runtime_readout"]["route"])
        self.assertEqual("kiro-cli", nano_resume_payload["runtime_target"]["selected"])
        self.assertTrue(
            nano_resume_payload["runtime_readout"]["reason"].startswith(
                "State `operational` requires stronger verification posture; Preferred adapter `kiro-cli` was ..."
            )
        )
        self.assertNotIn("context_type", nano_resume_payload)
        self.assertNotIn("schema_version", nano_resume_payload)
        self.assertNotIn("resume_items", nano_resume_payload)
        self.assertLessEqual(nano_resume_payload["token_budget"]["estimated_chars"], 375)
        self.assertLessEqual(len(nano_resume_result.stdout.strip()), 375)

        pico_resume_result = self.run_cli(
            "resume",
            pack_dir,
            "--runtime-target",
            "kiro-cli",
            "--format",
            "json",
            "--max-chars",
            "300",
            env=env,
        )
        self.assert_ok(pico_resume_result)
        pico_resume_payload = json.loads(pico_resume_result.stdout)
        self.assertEqual("local-workhorse", pico_resume_payload["runtime_readout"]["route"])
        self.assertEqual("kiro-cli", pico_resume_payload["runtime_target"]["selected"])
        self.assertTrue(
            pico_resume_payload["runtime_readout"]["reason"].startswith(
                "State `operational` requires stronger verification posture; Preferred adapter `kiro-cli` was ..."
            )
        )
        self.assertNotIn("pack_id", pico_resume_payload)
        self.assertNotIn("work_unit_id", pico_resume_payload)
        self.assertNotIn("efficiency_posture", pico_resume_payload)
        self.assertLessEqual(pico_resume_payload["token_budget"]["estimated_chars"], 300)
        self.assertLessEqual(len(pico_resume_result.stdout.strip()), 300)

        femto_resume_result = self.run_cli(
            "resume",
            pack_dir,
            "--runtime-target",
            "kiro-cli",
            "--format",
            "json",
            "--max-chars",
            "225",
            env=env,
        )
        self.assert_ok(femto_resume_result)
        femto_resume_payload = json.loads(femto_resume_result.stdout)
        self.assertEqual("local-workhorse", femto_resume_payload["runtime_readout"]["route"])
        self.assertEqual("kiro-cli", femto_resume_payload["runtime_target"]["selected"])
        self.assertTrue(
            femto_resume_payload["runtime_readout"]["reason"].startswith(
                "State `operational` requires stronger verification posture; Preferred adapter `kiro-cli` was ..."
            )
        )
        self.assertNotIn("state", femto_resume_payload)
        self.assertNotIn("token_budget", femto_resume_payload)
        self.assertLessEqual(len(femto_resume_result.stdout.strip()), 225)

        atto_resume_result = self.run_cli(
            "resume",
            pack_dir,
            "--runtime-target",
            "kiro-cli",
            "--format",
            "json",
            "--max-chars",
            "170",
            env=env,
        )
        self.assert_ok(atto_resume_result)
        atto_resume_payload = json.loads(atto_resume_result.stdout)
        self.assertEqual("local-workhorse", atto_resume_payload["runtime_readout"]["route"])
        self.assertEqual("kiro-cli", atto_resume_payload["runtime_target"]["selected"])
        self.assertTrue(
            atto_resume_payload["runtime_readout"]["reason"].startswith(
                "State `operational` requires stronger verification postur"
            )
        )
        self.assertLessEqual(len(atto_resume_result.stdout.strip()), 170)

    def test_route_outcome_feedback_self_corrects_measured_local_selection(self) -> None:
        pack_dir = self.compile_research_pack()
        state_root = self.tmp / ".jini"
        state_root.mkdir(parents=True, exist_ok=True)
        capabilities_path = state_root / "local-runtime-capabilities.json"
        capabilities_path.write_text(
            json.dumps(
                {
                    "schema_version": "0.4.0",
                    "context_type": "JiniLocalRuntimeCapabilities",
                    "captured_at": "2026-05-21T19:00:00Z",
                    "local_runtime_class": "local-ollama",
                    "adapters": {
                        "local-workhorse": {
                            "status": "ok",
                            "latency_ms": 190,
                            "warm_latency_ms": 160,
                            "cold_start_cost_ms": 30,
                            "tokens_per_second": 22.5,
                            "quality_class": "usable",
                            "structured_reliability": "usable",
                            "benchmarked_at": "2026-05-21T19:00:00Z",
                        },
                        "local-fast": {
                            "status": "ok",
                            "latency_ms": 85,
                            "warm_latency_ms": 70,
                            "cold_start_cost_ms": 5,
                            "tokens_per_second": 40.0,
                            "quality_class": "usable",
                            "structured_reliability": "usable",
                            "benchmarked_at": "2026-05-21T19:00:00Z",
                        },
                    },
                }
            ),
            encoding="utf-8",
        )

        baseline = self.run_cli("recommend-execution", pack_dir, "--intent", "export", "--format", "json")
        self.assert_ok(baseline)
        self.assertEqual(
            "local-fast",
            json.loads(baseline.stdout)["runtime_guidance"]["execution_route"]["selected"]["id"],
        )

        feedback = self.run_cli(
            "record-route-outcome",
            pack_dir,
            "--adapter-id",
            "local-fast",
            "--intent",
            "export",
            "--outcome",
            "replaced-this",
            "--reason",
            "Needed a full rewrite before use.",
            "--format",
            "json",
        )
        self.assert_ok(feedback)
        feedback_payload = json.loads(feedback.stdout)
        self.assertEqual("JiniRouteOutcomeFeedback", feedback_payload["feedback_type"])
        self.assertEqual("local-fast", feedback_payload["adapter_id"])
        self.assertEqual("export", feedback_payload["cohort_key"])
        self.assertEqual(-7, feedback_payload["feedback_bias"])

        saved = self.read_json(capabilities_path)
        feedback_row = saved["cohort_feedback"]["local-fast"]["export"]
        self.assertEqual(1, feedback_row["outcome_replaced"])
        self.assertIn("outcome_replaced", feedback_row["counter_last_observed_at"])

        corrected = self.run_cli("recommend-execution", pack_dir, "--intent", "export", "--format", "json")
        self.assert_ok(corrected)
        corrected_payload = json.loads(corrected.stdout)
        execution_route = corrected_payload["runtime_guidance"]["execution_route"]
        self.assertEqual("local-workhorse", execution_route["selected"]["id"])
        self.assertEqual("export", execution_route["feedback_cohort"])
        self.assertEqual(1, execution_route["outcome_signal_count"])
        self.assertIn("local-fast", execution_route["feedback_adjusted_adapters"])
        feedback_evidence = execution_route["feedback_evidence"]
        self.assertEqual("export", feedback_evidence["cohort_key"])
        self.assertEqual(1, feedback_evidence["total_signal_count"])
        self.assertEqual(["local-fast"], feedback_evidence["adjusted_adapters"])
        adapter_evidence = {item["adapter_id"]: item for item in feedback_evidence["adapters"]}
        self.assertEqual(-7, adapter_evidence["local-fast"]["bias"])
        self.assertEqual(1, adapter_evidence["local-fast"]["signal_count"])
        self.assertEqual({"outcome_replaced": 1}, adapter_evidence["local-fast"]["counters"])
        self.assertEqual("penalized", adapter_evidence["local-fast"]["routing_effect"])
        selection_delta = feedback_evidence["selection_delta"]
        self.assertTrue(selection_delta["selected_changed"])
        self.assertEqual("local-fast", selection_delta["baseline_selected_adapter"])
        self.assertEqual("local-workhorse", selection_delta["feedback_selected_adapter"])
        self.assertEqual(["local-fast", "local-workhorse"], selection_delta["baseline_rank"])
        self.assertEqual(["local-workhorse", "local-fast"], selection_delta["feedback_rank"])
        self.assertEqual(["local-fast"], selection_delta["demoted_adapters"])
        self.assertEqual(["local-workhorse"], selection_delta["promoted_adapters"])
        self.assertTrue(
            any(
                "Measured local runtime `local-ollama` reranked `local-fast` to `local-workhorse` from `export` feedback"
                in note
                for note in corrected_payload["runtime_guidance"]["notes"]
            )
        )

        corrected_text = self.run_cli("recommend-execution", pack_dir, "--intent", "export")
        self.assert_ok(corrected_text)
        self.assertIn("IMPACT baseline=local-fast feedback=local-workhorse changed=yes", corrected_text.stdout)

        resume_result = self.run_cli(
            "resume",
            pack_dir,
            "--intent",
            "export",
            "--format",
            "json",
            "--max-chars",
            "1100",
        )
        self.assert_ok(resume_result)
        resume_payload = json.loads(resume_result.stdout)
        self.assertTrue(
            resume_payload["runtime_readout"]["reason"].startswith(
                "Measured local runtime `local-ollama` reranked `local-fast` to `local-workhorse` from `export"
            )
        )

        events = self.run_cli("show-learning-events", pack_dir, "--event-type", "route-outcome-feedback", "--format", "json")
        self.assert_ok(events)
        event = json.loads(events.stdout)["events"][-1]
        self.assertEqual("local-fast", event["adapter_id"])
        self.assertEqual("replaced-this", event["outcome"])

        backtest_result = self.run_cli("routing-backtest", pack_dir, "--format", "json")
        self.assert_ok(backtest_result)
        route_feedback = json.loads(backtest_result.stdout)["route_feedback"]
        self.assertEqual(1, route_feedback["event_count"])
        cohort = route_feedback["cohorts"][0]
        self.assertEqual("export", cohort["cohort_key"])
        self.assertEqual(1, cohort["signal_count"])
        self.assertEqual(["local-fast"], cohort["adjusted_adapters"])
        backtest_adapter = cohort["adapters"][0]
        self.assertEqual("local-fast", backtest_adapter["adapter_id"])
        self.assertEqual({"outcome_replaced": 1}, backtest_adapter["counters"])
        self.assertEqual(-7, backtest_adapter["latest_feedback_bias"])

    def test_route_feedback_notes_when_feedback_adjusts_but_keeps_local_winner(self) -> None:
        pack_dir = self.compile_research_pack()
        state_root = self.tmp / ".jini"
        state_root.mkdir(parents=True, exist_ok=True)
        capabilities_path = state_root / "local-runtime-capabilities.json"
        capabilities_path.write_text(
            json.dumps(
                {
                    "schema_version": "0.4.0",
                    "context_type": "JiniLocalRuntimeCapabilities",
                    "captured_at": "2026-05-21T19:00:00Z",
                    "local_runtime_class": "local-ollama",
                    "adapters": {
                        "local-workhorse": {
                            "status": "ok",
                            "latency_ms": 190,
                            "warm_latency_ms": 160,
                            "cold_start_cost_ms": 30,
                            "tokens_per_second": 22.5,
                            "quality_class": "strong",
                            "structured_reliability": "strong",
                            "benchmarked_at": "2026-05-21T19:00:00Z",
                        },
                        "local-fast": {
                            "status": "ok",
                            "latency_ms": 85,
                            "warm_latency_ms": 70,
                            "cold_start_cost_ms": 5,
                            "tokens_per_second": 40.0,
                            "quality_class": "usable",
                            "structured_reliability": "usable",
                            "benchmarked_at": "2026-05-21T19:00:00Z",
                        },
                    },
                }
            ),
            encoding="utf-8",
        )

        feedback = self.run_cli(
            "record-route-outcome",
            pack_dir,
            "--adapter-id",
            "local-workhorse",
            "--intent",
            "make",
            "--outcome",
            "replaced-this",
            "--reason",
            "Needed cleanup before use.",
            "--format",
            "json",
        )
        self.assert_ok(feedback)

        result = self.run_cli("recommend-execution", pack_dir, "--intent", "make", "--format", "json")
        self.assert_ok(result)
        payload = json.loads(result.stdout)
        execution_route = payload["runtime_guidance"]["execution_route"]
        self.assertEqual("local-workhorse", execution_route["selected"]["id"])
        self.assertEqual(["local-workhorse"], execution_route["feedback_adjusted_adapters"])
        selection_delta = execution_route["feedback_evidence"]["selection_delta"]
        self.assertFalse(selection_delta["selected_changed"])
        self.assertEqual("local-workhorse", selection_delta["baseline_selected_adapter"])
        self.assertEqual("local-workhorse", selection_delta["feedback_selected_adapter"])
        self.assertTrue(
            any(
                "Measured local runtime `local-ollama` kept `local-workhorse` after `make` feedback adjusted local-workhorse"
                in note
                for note in payload["runtime_guidance"]["notes"]
            )
        )

        status_result = self.run_cli("status", pack_dir, "--format", "json")
        self.assert_ok(status_result)
        status_payload = json.loads(status_result.stdout)
        self.assertIn(
            "Measured local runtime `local-ollama` kept `local-workhorse` after `make` feedback adjusted local-workhorse",
            status_payload["runtime_readout"]["reason"],
        )

    def test_stale_route_outcome_feedback_does_not_permanently_steer_selection(self) -> None:
        pack_dir = self.compile_research_pack()
        state_root = self.tmp / ".jini"
        state_root.mkdir(parents=True, exist_ok=True)
        capabilities_path = state_root / "local-runtime-capabilities.json"
        capabilities_path.write_text(
            json.dumps(
                {
                    "schema_version": "0.4.0",
                    "context_type": "JiniLocalRuntimeCapabilities",
                    "captured_at": "2026-05-21T19:00:00Z",
                    "local_runtime_class": "local-ollama",
                    "adapters": {
                        "local-workhorse": {
                            "status": "ok",
                            "latency_ms": 190,
                            "warm_latency_ms": 160,
                            "cold_start_cost_ms": 30,
                            "tokens_per_second": 22.5,
                            "quality_class": "usable",
                            "structured_reliability": "usable",
                            "benchmarked_at": "2026-05-21T19:00:00Z",
                        },
                        "local-fast": {
                            "status": "ok",
                            "latency_ms": 85,
                            "warm_latency_ms": 70,
                            "cold_start_cost_ms": 5,
                            "tokens_per_second": 40.0,
                            "quality_class": "usable",
                            "structured_reliability": "usable",
                            "benchmarked_at": "2026-05-21T19:00:00Z",
                        },
                    },
                    "cohort_feedback": {
                        "local-fast": {
                            "export": {
                                "outcome_replaced": 1,
                                "counter_last_observed_at": {
                                    "outcome_replaced": "2026-01-01T00:00:00Z",
                                },
                            },
                        },
                    },
                }
            ),
            encoding="utf-8",
        )

        result = self.run_cli("recommend-execution", pack_dir, "--intent", "export", "--format", "json")
        self.assert_ok(result)
        execution_route = json.loads(result.stdout)["runtime_guidance"]["execution_route"]
        self.assertEqual("local-fast", execution_route["selected"]["id"])
        self.assertEqual(0, execution_route["outcome_signal_count"])
        self.assertEqual([], execution_route["feedback_adjusted_adapters"])
        feedback_evidence = execution_route["feedback_evidence"]
        self.assertEqual(0, feedback_evidence["total_signal_count"])
        self.assertEqual(1, feedback_evidence["total_expired_signal_count"])
        adapter_evidence = {item["adapter_id"]: item for item in feedback_evidence["adapters"]}
        self.assertEqual(0, adapter_evidence["local-fast"]["bias"])
        self.assertEqual(0, adapter_evidence["local-fast"]["signal_count"])
        self.assertEqual(1, adapter_evidence["local-fast"]["expired_signal_count"])
        self.assertEqual({"outcome_replaced": 1}, adapter_evidence["local-fast"]["expired_counters"])
        self.assertEqual("neutral", adapter_evidence["local-fast"]["routing_effect"])

    def test_route_feedback_readout_can_prune_expired_state_without_deleting_history(self) -> None:
        pack_dir = self.compile_research_pack()
        state_root = self.tmp / ".jini"
        state_root.mkdir(parents=True, exist_ok=True)
        capabilities_path = state_root / "local-runtime-capabilities.json"
        capabilities_path.write_text(
            json.dumps(
                {
                    "schema_version": "0.4.0",
                    "context_type": "JiniLocalRuntimeCapabilities",
                    "captured_at": "2026-05-21T19:00:00Z",
                    "local_runtime_class": "local-ollama",
                    "adapters": {
                        "local-fast": {
                            "status": "ok",
                            "latency_ms": 85,
                            "warm_latency_ms": 70,
                            "cold_start_cost_ms": 5,
                            "tokens_per_second": 40.0,
                            "quality_class": "usable",
                            "structured_reliability": "usable",
                            "benchmarked_at": "2026-05-21T19:00:00Z",
                        },
                    },
                }
            ),
            encoding="utf-8",
        )
        feedback = self.run_cli(
            "record-route-outcome",
            pack_dir,
            "--adapter-id",
            "local-fast",
            "--intent",
            "export",
            "--outcome",
            "replaced-this",
            "--reason",
            "No longer useful for this export.",
            "--format",
            "json",
        )
        self.assert_ok(feedback)
        saved = self.read_json(capabilities_path)
        saved["cohort_feedback"]["local-fast"]["export"]["counter_last_observed_at"]["outcome_replaced"] = (
            "2026-01-01T00:00:00Z"
        )
        capabilities_path.write_text(json.dumps(saved, indent=2, sort_keys=True) + "\n", encoding="utf-8")

        readout = self.run_cli("route-feedback", "--format", "json")
        self.assert_ok(readout)
        report = json.loads(readout.stdout)
        self.assertEqual("JiniRouteFeedbackMaintenance", report["report_type"])
        self.assertFalse(report["pruned"])
        self.assertEqual(0, report["total_active_signal_count"])
        self.assertEqual(1, report["total_expired_signal_count"])
        adapter = report["adapters"][0]
        cohort = adapter["cohorts"][0]
        self.assertEqual("local-fast", adapter["adapter_id"])
        self.assertEqual("export", cohort["cohort_key"])
        self.assertEqual({"outcome_replaced": 1}, cohort["expired_counters"])

        before_events = self.run_cli(
            "show-learning-events",
            pack_dir,
            "--event-type",
            "route-outcome-feedback",
            "--format",
            "json",
        )
        self.assert_ok(before_events)
        self.assertEqual(1, len(json.loads(before_events.stdout)["events"]))

        pruned = self.run_cli("route-feedback", "--prune-expired", "--format", "json")
        self.assert_ok(pruned)
        pruned_report = json.loads(pruned.stdout)
        self.assertTrue(pruned_report["pruned"])
        self.assertEqual(1, pruned_report["pruned_signal_count"])
        self.assertEqual(0, pruned_report["total_expired_signal_count"])
        pruned_payload = self.read_json(capabilities_path)
        self.assertNotIn("local-fast", pruned_payload.get("cohort_feedback", {}))

        after_events = self.run_cli(
            "show-learning-events",
            pack_dir,
            "--event-type",
            "route-outcome-feedback",
            "--format",
            "json",
        )
        self.assert_ok(after_events)
        self.assertEqual(1, len(json.loads(after_events.stdout)["events"]))

    def test_route_feedback_readout_surfaces_multi_cohort_routing_impact_summary(self) -> None:
        pack_dir = self.compile_research_pack()
        state_root = self.tmp / ".jini"
        state_root.mkdir(parents=True, exist_ok=True)
        capabilities_path = state_root / "local-runtime-capabilities.json"
        capabilities_path.write_text(
            json.dumps(
                {
                    "schema_version": "0.4.0",
                    "context_type": "JiniLocalRuntimeCapabilities",
                    "captured_at": "2026-05-21T19:00:00Z",
                    "local_runtime_class": "local-ollama",
                    "adapters": {
                        "local-fast": {
                            "status": "ok",
                            "latency_ms": 85,
                            "warm_latency_ms": 70,
                            "cold_start_cost_ms": 5,
                            "tokens_per_second": 40.0,
                            "quality_class": "strong",
                            "structured_reliability": "strong",
                            "benchmarked_at": "2026-05-21T19:00:00Z",
                        },
                        "local-workhorse": {
                            "status": "ok",
                            "latency_ms": 190,
                            "warm_latency_ms": 160,
                            "cold_start_cost_ms": 30,
                            "tokens_per_second": 22.5,
                            "quality_class": "strong",
                            "structured_reliability": "strong",
                            "benchmarked_at": "2026-05-21T19:00:00Z",
                        },
                    },
                }
            ),
            encoding="utf-8",
        )
        feedback = self.run_cli(
            "record-route-outcome",
            pack_dir,
            "--adapter-id",
            "local-fast",
            "--intent",
            "export",
            "--outcome",
            "replaced-this",
            "--reason",
            "Needed a stronger export route.",
            "--format",
            "json",
        )
        self.assert_ok(feedback)
        second_feedback = self.run_cli(
            "record-route-outcome",
            pack_dir,
            "--adapter-id",
            "local-fast",
            "--intent",
            "wiki",
            "--outcome",
            "replaced-this",
            "--reason",
            "Needed a stronger wiki route.",
            "--format",
            "json",
        )
        self.assert_ok(second_feedback)

        readout = self.run_cli("route-feedback", "--format", "json")
        self.assert_ok(readout)
        report = json.loads(readout.stdout)
        impact = report["route_feedback_impact"]
        self.assertEqual("changed", impact["status"])
        self.assertEqual(2, impact["active_cohort_count"])
        self.assertEqual(2, impact["changed_selection_count"])
        self.assertEqual("Review routing policy", impact["recommended_action"]["label"])
        self.assertEqual("jini review-policy", impact["recommended_action"]["command"])
        self.assertEqual(["export", "wiki"], impact["changed_cohort_keys"])
        self.assertEqual(
            [
                "export:local-fast->local-workhorse",
                "wiki:local-fast->local-workhorse",
            ],
            impact["cohort_preview"]["entries"],
        )
        self.assertEqual(
            "export:local-fast->local-workhorse,wiki:local-fast->local-workhorse",
            impact["cohort_preview"]["text"],
        )
        self.assertEqual("export", impact["cohorts"][0]["cohort_key"])
        self.assertEqual("local-fast", impact["cohorts"][0]["baseline_selected_adapter"])
        self.assertEqual("local-workhorse", impact["cohorts"][0]["feedback_selected_adapter"])
        self.assertEqual("wiki", impact["cohorts"][1]["cohort_key"])

        readout_text = self.run_cli("route-feedback")
        self.assert_ok(readout_text)
        self.assertIn(
            "IMPACT changed=2/2 cohorts=export:local-fast->local-workhorse,wiki:local-fast->local-workhorse action=jini review-policy",
            readout_text.stdout,
        )

    def test_status_and_metrics_surface_route_feedback_freshness(self) -> None:
        pack_dir = self.compile_research_pack()
        state_root = self.tmp / ".jini"
        state_root.mkdir(parents=True, exist_ok=True)
        (state_root / "local-runtime-capabilities.json").write_text(
            json.dumps(
                {
                    "schema_version": "0.4.0",
                    "context_type": "JiniLocalRuntimeCapabilities",
                    "captured_at": "2026-05-21T19:00:00Z",
                    "local_runtime_class": "local-ollama",
                    "adapters": {
                        "local-fast": {
                            "status": "ok",
                            "latency_ms": 85,
                            "warm_latency_ms": 70,
                            "cold_start_cost_ms": 5,
                            "tokens_per_second": 40.0,
                            "quality_class": "usable",
                            "structured_reliability": "usable",
                            "benchmarked_at": "2026-05-21T19:00:00Z",
                        },
                        "local-workhorse": {
                            "status": "ok",
                            "latency_ms": 190,
                            "warm_latency_ms": 160,
                            "cold_start_cost_ms": 30,
                            "tokens_per_second": 22.5,
                            "quality_class": "usable",
                            "structured_reliability": "usable",
                            "benchmarked_at": "2026-05-21T19:00:00Z",
                        },
                    },
                    "cohort_feedback": {
                        "local-fast": {
                            "export": {
                                "outcome_replaced": 1,
                                "not_useful": 1,
                                "counter_last_observed_at": {
                                    "outcome_replaced": "2026-01-01T00:00:00Z",
                                    "not_useful": "2026-05-21T19:10:00Z",
                                },
                            },
                            "wiki": {
                                "outcome_replaced": 1,
                                "counter_last_observed_at": {
                                    "outcome_replaced": "2026-05-21T19:11:00Z",
                                },
                            },
                        }
                    },
                }
            ),
            encoding="utf-8",
        )

        status_json = self.run_cli("status", pack_dir, "--format", "json")
        self.assert_ok(status_json)
        status_payload = json.loads(status_json.stdout)
        status_health = status_payload["route_feedback_health"]
        self.assertEqual("stale", status_health["status"])
        self.assertEqual(2, status_health["active_signal_count"])
        self.assertEqual(1, status_health["expired_signal_count"])
        self.assertEqual("jini route-feedback --prune-expired", status_health["recommended_action"]["command"])
        self.assertEqual(status_health, status_payload["runtime_readout"]["route_feedback_health"])
        status_impact = status_payload["route_feedback_impact"]
        self.assertEqual("changed", status_impact["status"])
        self.assertEqual(2, status_impact["changed_selection_count"])
        self.assertEqual("Review routing policy", status_impact["recommended_action"]["label"])
        self.assertEqual("jini review-policy", status_impact["recommended_action"]["command"])
        self.assertEqual(["export", "wiki"], status_impact["changed_cohort_keys"])
        self.assertEqual(
            "export:local-fast->local-workhorse,wiki:local-fast->local-workhorse",
            status_impact["cohort_preview"]["text"],
        )
        self.assertEqual(status_impact, status_payload["runtime_readout"]["route_feedback_impact"])
        self.assertEqual("local-fast", status_impact["cohorts"][0]["baseline_selected_adapter"])
        self.assertEqual("local-workhorse", status_impact["cohorts"][0]["feedback_selected_adapter"])

        status_text = self.run_cli("status", pack_dir)
        self.assert_ok(status_text)
        self.assertIn("LEARNING", status_text.stdout)
        self.assertIn("route-feedback status=stale active=2 expired=1", status_text.stdout)
        self.assertIn(
            "impact changed=2/2 cohorts=export:local-fast->local-workhorse,wiki:local-fast->local-workhorse action=jini review-policy",
            status_text.stdout,
        )
        self.assertIn("jini route-feedback --prune-expired", status_text.stdout)

        metrics_json = self.run_cli("metrics", "--format", "json")
        self.assert_ok(metrics_json)
        metrics_payload = json.loads(metrics_json.stdout)
        metrics_health = metrics_payload["route_feedback_health"]
        self.assertEqual("stale", metrics_health["status"])
        self.assertEqual(1, metrics_health["expired_signal_count"])
        self.assertEqual("jini route-feedback --prune-expired", metrics_health["recommended_action"]["command"])
        metrics_impact = metrics_payload["route_feedback_impact"]
        self.assertEqual("changed", metrics_impact["status"])
        self.assertEqual(2, metrics_impact["changed_selection_count"])
        self.assertEqual("Review routing policy", metrics_impact["recommended_action"]["label"])
        self.assertEqual("jini review-policy", metrics_impact["recommended_action"]["command"])
        self.assertEqual(["export", "wiki"], metrics_impact["changed_cohort_keys"])
        self.assertEqual(
            "export:local-fast->local-workhorse,wiki:local-fast->local-workhorse",
            metrics_impact["cohort_preview"]["text"],
        )
        self.assertEqual("local-fast", metrics_impact["cohorts"][0]["baseline_selected_adapter"])
        self.assertEqual("local-workhorse", metrics_impact["cohorts"][0]["feedback_selected_adapter"])

        metrics_text = self.run_cli("metrics")
        self.assert_ok(metrics_text)
        self.assertIn(
            "ROUTEFEEDBACK status=stale active=2 expired=1 adapters=1 action=jini route-feedback --prune-expired",
            metrics_text.stdout,
        )
        self.assertIn(
            "ROUTEIMPACT status=changed changed=2/2 cohorts=export:local-fast->local-workhorse,wiki:local-fast->local-workhorse action=jini review-policy",
            metrics_text.stdout,
        )

    def test_passive_route_outcome_feedback_is_captured_from_user_actions(self) -> None:
        pack_dir = self.compile_research_pack()
        state_root = self.tmp / ".jini"
        state_root.mkdir(parents=True, exist_ok=True)
        capabilities_path = state_root / "local-runtime-capabilities.json"
        capabilities_path.write_text(
            json.dumps(
                {
                    "schema_version": "0.4.0",
                    "context_type": "JiniLocalRuntimeCapabilities",
                    "captured_at": "2026-05-21T19:00:00Z",
                    "local_runtime_class": "local-ollama",
                    "adapters": {
                        "local-fast": {
                            "status": "ok",
                            "latency_ms": 85,
                            "warm_latency_ms": 70,
                            "cold_start_cost_ms": 5,
                            "tokens_per_second": 40.0,
                            "quality_class": "strong",
                            "structured_reliability": "strong",
                            "benchmarked_at": "2026-05-21T19:00:00Z",
                        },
                    },
                }
            ),
            encoding="utf-8",
        )

        opened = self.run_cli("open", "tasks", "--from", pack_dir, "--print-path")
        self.assert_ok(opened)
        exported = self.run_cli("export-tasks", pack_dir)
        self.assert_ok(exported)
        rewritten = self.run_cli("rewrite", "checklist", "tasks", "--from", pack_dir, "--format", "json")
        self.assert_ok(rewritten)

        saved = self.read_json(capabilities_path)
        feedback = saved["cohort_feedback"]["local-fast"]
        self.assertEqual(1, feedback["open"]["passive_reopened"])
        self.assertEqual(1, feedback["export"]["passive_export_opened"])
        self.assertEqual(1, feedback["rewrite"]["passive_needed_light_edits"])

        events = self.run_cli("show-learning-events", pack_dir, "--event-type", "route-outcome-feedback", "--format", "json")
        self.assert_ok(events)
        passive_events = [event for event in json.loads(events.stdout)["events"] if event["passive"]]
        self.assertEqual(["open", "export", "rewrite"], [event["intent"] for event in passive_events[-3:]])
        self.assertEqual(["used-this", "shared-this", "needed-light-edits"], [event["outcome"] for event in passive_events[-3:]])

    def test_passive_route_outcome_feedback_debounces_repeated_user_actions(self) -> None:
        pack_dir = self.compile_research_pack()
        state_root = self.tmp / ".jini"
        state_root.mkdir(parents=True, exist_ok=True)
        capabilities_path = state_root / "local-runtime-capabilities.json"
        capabilities_path.write_text(
            json.dumps(
                {
                    "schema_version": "0.4.0",
                    "context_type": "JiniLocalRuntimeCapabilities",
                    "captured_at": "2026-05-21T19:00:00Z",
                    "local_runtime_class": "local-ollama",
                    "adapters": {
                        "local-fast": {
                            "status": "ok",
                            "latency_ms": 85,
                            "warm_latency_ms": 70,
                            "cold_start_cost_ms": 5,
                            "tokens_per_second": 40.0,
                            "quality_class": "strong",
                            "structured_reliability": "strong",
                            "benchmarked_at": "2026-05-21T19:00:00Z",
                        },
                    },
                }
            ),
            encoding="utf-8",
        )

        for _ in range(2):
            self.assert_ok(self.run_cli("open", "tasks", "--from", pack_dir, "--print-path"))
            self.assert_ok(self.run_cli("export-tasks", pack_dir))
            self.assert_ok(self.run_cli("rewrite", "checklist", "tasks", "--from", pack_dir, "--format", "json"))

        saved = self.read_json(capabilities_path)
        feedback = saved["cohort_feedback"]["local-fast"]
        self.assertEqual(1, feedback["open"]["passive_reopened"])
        self.assertEqual(1, feedback["export"]["passive_export_opened"])
        self.assertEqual(1, feedback["rewrite"]["passive_needed_light_edits"])
        observations = saved["passive_feedback_observations"]["local-fast"]
        self.assertEqual(3, len(observations))

        events = self.run_cli("show-learning-events", pack_dir, "--event-type", "route-outcome-feedback", "--format", "json")
        self.assert_ok(events)
        passive_events = [event for event in json.loads(events.stdout)["events"] if event["passive"]]
        self.assertEqual(3, len(passive_events))

    def test_status_surfaces_turn_record_and_progress_frame(self) -> None:
        pack_dir = self.compile_research_pack()

        status_json = self.run_cli("status", pack_dir, "--format", "json")
        self.assert_ok(status_json)
        report = json.loads(status_json.stdout)

        progress = report["progress_snapshot"]
        self.assertEqual("Test Research Pack", progress["goal"])
        self.assertIn("ready artifact", progress["working_with_summary"])
        self.assertEqual("Make", progress["next"])
        self.assertTrue(progress["safe_to_do"])

        turn = report["turn_record"]
        self.assertEqual("test-research-pack", turn["thread_id"])
        self.assertIn("initial-request", turn["user_input_ids"])
        self.assertTrue(turn["artifacts_created"])
        self.assertEqual([], turn["artifacts_updated"])
        self.assertTrue(turn["state_changes"])
        self.assertLessEqual(len([ask for ask in turn["asks_opened"] if ask.get("blocking")]), 1)
        self.assertEqual("session-kernel", turn["route_decision"]["provider_id"])

        status_text = self.run_cli("status", pack_dir)
        self.assert_ok(status_text)
        self.assertIn("JUST FINISHED", status_text.stdout)
        self.assertIn("DOING NOW", status_text.stdout)
        self.assertIn("UP NEXT", status_text.stdout)

    def test_status_surfaces_working_with_input_strip(self) -> None:
        pack_dir = self.compile_research_pack()

        status_json = self.run_cli("status", pack_dir, "--format", "json")
        self.assert_ok(status_json)
        report = json.loads(status_json.stdout)

        input_items = report["working_with"]["input_items"]
        self.assertEqual(1, len(input_items))
        self.assertEqual(input_items, report["input_items"])
        self.assertEqual("initial-request", input_items[0]["input_id"])
        self.assertEqual("test-research-pack", input_items[0]["thread_id"])
        self.assertEqual("text", input_items[0]["kind"])
        self.assertEqual("processed", input_items[0]["status"])
        self.assertEqual("product-lead", input_items[0]["source_actor"])
        self.assertIn("Exercise the research pack lifecycle", input_items[0]["preview"])
        self.assertEqual("extracted", input_items[0]["extraction_status"])
        self.assertIn("Captured text request", input_items[0]["extraction_summary"])
        self.assertIn("prd", input_items[0]["derived_artifact_ids"])
        self.assertIn("initial-request", report["turn_record"]["user_input_ids"])

        status_text = self.run_cli("status", pack_dir)
        self.assert_ok(status_text)
        self.assertIn("WORKING WITH", status_text.stdout)
        self.assertIn("Test Research Pack", status_text.stdout)
        self.assertIn("Exercise the research pack lifecycle", status_text.stdout)
        self.assertIn("extracted:", status_text.stdout)

    def test_status_refresh_preserves_saved_non_initial_input_items(self) -> None:
        pack_dir = self.compile_research_pack()
        projection_path = self.tmp / ".jini" / "sessions" / "test-research-pack" / "projection.json"
        projection_doc = json.loads(projection_path.read_text(encoding="utf-8"))
        projection_doc["input_items"].append(
            {
                "input_id": "uploaded-brief",
                "thread_id": "test-research-pack",
                "kind": "file",
                "title": "Uploaded Brief",
                "source_actor": "product-lead",
                "status": "processed",
                "preview": "Imported customer evidence brief",
                "origin_ref": "brief.pdf",
                "derived_artifact_ids": ["prd"],
                "created_at": "2026-05-01T00:00:00Z",
                "updated_at": "2026-05-01T00:00:00Z",
            }
        )
        projection_path.write_text(json.dumps(projection_doc, indent=2), encoding="utf-8")

        status_json = self.run_cli("status", pack_dir, "--format", "json")
        self.assert_ok(status_json)
        report = json.loads(status_json.stdout)

        input_ids = [item["input_id"] for item in report["input_items"]]
        self.assertIn("initial-request", input_ids)
        self.assertIn("uploaded-brief", input_ids)
        self.assertIn("uploaded-brief", report["turn_record"]["user_input_ids"])

    def test_status_surfaces_input_extraction_summaries_and_failures(self) -> None:
        pack_dir = self.compile_research_pack()
        projection_path = self.tmp / ".jini" / "sessions" / "test-research-pack" / "projection.json"
        projection_doc = json.loads(projection_path.read_text(encoding="utf-8"))
        projection_doc["input_items"].extend(
            [
                {
                    "input_id": "hotel-photo",
                    "thread_id": "test-research-pack",
                    "kind": "image",
                    "title": "Hotel Photo",
                    "source_actor": "product-lead",
                    "status": "processed",
                    "preview": "Read whiteboard budget totals and dates",
                    "origin_ref": "hotel-whiteboard.png",
                    "derived_artifact_ids": ["prd"],
                    "created_at": "2026-05-01T00:00:00Z",
                    "updated_at": "2026-05-01T00:00:00Z",
                },
                {
                    "input_id": "voice-note",
                    "thread_id": "test-research-pack",
                    "kind": "audio",
                    "title": "Voice Note",
                    "source_actor": "product-lead",
                    "status": "failed",
                    "preview": "Voice note could not be transcribed",
                    "origin_ref": "voice-note.m4a",
                    "error_message": "transcription confidence below threshold",
                    "derived_artifact_ids": [],
                    "created_at": "2026-05-01T00:00:00Z",
                    "updated_at": "2026-05-01T00:00:00Z",
                },
            ]
        )
        projection_path.write_text(json.dumps(projection_doc, indent=2), encoding="utf-8")

        status_json = self.run_cli("status", pack_dir, "--format", "json")
        self.assert_ok(status_json)
        report = json.loads(status_json.stdout)

        by_id = {item["input_id"]: item for item in report["input_items"]}
        self.assertEqual("extracted", by_id["hotel-photo"]["extraction_status"])
        self.assertIn("Observed image input", by_id["hotel-photo"]["extraction_summary"])
        self.assertIn("hotel-whiteboard.png", by_id["hotel-photo"]["extraction_summary"])
        self.assertEqual("failed", by_id["voice-note"]["extraction_status"])
        self.assertEqual(
            "transcription confidence below threshold",
            by_id["voice-note"]["failure_reason"],
        )
        self.assertIn("Could not process voice-note.m4a", by_id["voice-note"]["extraction_summary"])

        status_text = self.run_cli("status", pack_dir)
        self.assert_ok(status_text)
        self.assertIn("Hotel Photo [image/processed]", status_text.stdout)
        self.assertIn("extracted: Observed image input", status_text.stdout)
        self.assertIn("Voice Note [audio/failed]", status_text.stdout)
        self.assertIn("failed: Could not process voice-note.m4a", status_text.stdout)

    def test_status_surfaces_grouped_artifact_shelf(self) -> None:
        pack_dir = self.compile_research_pack()

        status_json = self.run_cli("status", pack_dir, "--format", "json")
        self.assert_ok(status_json)
        report = json.loads(status_json.stdout)

        shelf = report["artifact_shelf"]
        self.assertEqual("Ready now", shelf["ready_now"]["label"])
        self.assertEqual("Needs input", shelf["needs_input"]["label"])
        self.assertEqual("Blocked", shelf["blocked"]["label"])
        ready_cards = shelf["ready_now"]["cards"]
        self.assertTrue(ready_cards)
        ready_ids = [card["artifact_id"] for card in ready_cards]
        self.assertIn("tasks", ready_ids)
        tasks_card = next(card for card in ready_cards if card["artifact_id"] == "tasks")
        self.assertEqual("JiniArtifactCard", tasks_card["card_type"])
        self.assertEqual("test-research-pack", tasks_card["thread_id"])
        self.assertEqual("Tasks", tasks_card["title"])
        self.assertEqual("ready", tasks_card["status"])
        self.assertTrue(tasks_card["summary"])
        self.assertTrue(tasks_card["preview"])
        self.assertEqual("jini open tasks", tasks_card["open_action"]["command"])
        self.assertTrue(tasks_card["export_actions"])
        self.assertIn("initial-request", tasks_card["source_input_ids"])
        self.assertEqual(ready_cards, report["artifact_cards"])

        status_text = self.run_cli("status", pack_dir)
        self.assert_ok(status_text)
        self.assertIn("ARTIFACT SHELF", status_text.stdout)
        self.assertIn("Ready now", status_text.stdout)
        self.assertIn("Tasks [ready]", status_text.stdout)
        self.assertIn("jini open tasks", status_text.stdout)

    def test_status_shelf_groups_missing_required_artifacts_as_needs_input(self) -> None:
        pack_dir = self.compile_research_pack()
        work_unit_path = pack_dir / "work-unit.yaml"
        work_unit_text = work_unit_path.read_text(encoding="utf-8")
        self.assertIn("current_state: decided", work_unit_text)
        work_unit_path.write_text(
            work_unit_text.replace("current_state: decided", "current_state: operational"),
            encoding="utf-8",
        )

        status_json = self.run_cli("status", pack_dir, "--format", "json")
        self.assert_ok(status_json)
        report = json.loads(status_json.stdout)

        needs_cards = report["artifact_shelf"]["needs_input"]["cards"]
        needs_ids = [card["artifact_id"] for card in needs_cards]
        self.assertIn("approval", needs_ids)
        approval_card = next(card for card in needs_cards if card["artifact_id"] == "approval")
        self.assertEqual("Approval", approval_card["artifact_type"])
        self.assertEqual("needs_input", approval_card["status"])
        self.assertIn("Missing required artifact", approval_card["summary"])

        status_text = self.run_cli("status", pack_dir)
        self.assert_ok(status_text)
        self.assertIn("Needs input", status_text.stdout)
        self.assertIn("Approval [needs_input]", status_text.stdout)

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
        self.assertTrue(projection_doc["ready"])
        self.assertTrue(projection_doc["ready"][0]["snapshot_markdown"])
        self.assertIn("snapshot_trimmed", projection_doc["ready"][0])
        self.assertIn("current_focus", projection_doc)
        self.assertEqual("artifact", projection_doc["current_focus"]["kind"])
        self.assertEqual("JiniTurnRecord", projection_doc["turn_record"]["record_type"])
        self.assertEqual("test-research-pack", projection_doc["turn_record"]["thread_id"])
        self.assertIn("progress_snapshot", projection_doc)
        self.assertEqual("Test Research Pack", projection_doc["progress_snapshot"]["goal"])
        self.assertIn("input_items", projection_doc)
        self.assertEqual("initial-request", projection_doc["input_items"][0]["input_id"])
        self.assertEqual("test-research-pack", projection_doc["input_items"][0]["thread_id"])
        self.assertEqual("text", projection_doc["input_items"][0]["kind"])
        self.assertEqual("processed", projection_doc["input_items"][0]["status"])
        self.assertIn("tasks", projection_doc["input_items"][0]["derived_artifact_ids"])
        self.assertIn("artifact_shelf", projection_doc)
        self.assertEqual("Ready now", projection_doc["artifact_shelf"]["ready_now"]["label"])
        self.assertTrue(projection_doc["artifact_shelf"]["ready_now"]["cards"])
        self.assertIn("artifact_cards", projection_doc)
        self.assertEqual(
            projection_doc["artifact_shelf"]["ready_now"]["cards"],
            projection_doc["artifact_cards"],
        )

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

    def test_show_without_artifact_uses_focused_current_work(self) -> None:
        pack_dir = self.compile_travel_pack()

        focused = self.run_cli("show", "itinerary", "--from", pack_dir)
        self.assert_ok(focused)

        show = self.run_cli("show")
        self.assert_ok(show)
        self.assertIn("# Itinerary: Test Travel Pack", show.stdout)

        open_result = self.run_cli("open", "--print-path")
        self.assert_ok(open_result)
        self.assertIn(str((pack_dir / "views" / "itinerary.md").resolve()), open_result.stdout.strip())

    def test_open_without_artifact_shows_actionable_shelf(self) -> None:
        pack_dir = self.compile_research_pack()

        result = self.run_cli("open", "--from", pack_dir)
        self.assert_ok(result)

        self.assertIn("OPEN SHELF", result.stdout)
        self.assertIn("READY NOW", result.stdout)
        self.assertIn("SHAREABLE EXPORTS", result.stdout)
        self.assertIn("DETAILS", result.stdout)
        self.assertIn("1.", result.stdout)
        self.assertIn("prd", result.stdout)
        self.assertIn("tasks", result.stdout)
        self.assertIn("github", result.stdout)
        self.assertIn("jini open 1", result.stdout)
        self.assertNotIn("--from", result.stdout)
        self.assertNotIn(str(pack_dir.resolve()), result.stdout)
        self.assertNotIn("/views/", result.stdout)

    def test_open_shelf_number_selects_artifact_and_records_observation(self) -> None:
        pack_dir = self.compile_research_pack()

        result = self.run_cli("open", "2", "--from", pack_dir, "--print-path")
        self.assert_ok(result)

        self.assertIn(str((pack_dir / "views" / "tasks.md").resolve()), result.stdout.strip())
        events_path = pack_dir / "runtime" / "events.jsonl"
        self.assertTrue(events_path.exists())
        events = [
            json.loads(line)
            for line in events_path.read_text(encoding="utf-8").splitlines()
            if line.strip()
        ]
        artifact_open_events = [event for event in events if event["event_type"] == "artifact-open"]
        self.assertTrue(artifact_open_events)
        self.assertEqual("tasks", artifact_open_events[-1]["artifact_id"])
        self.assertEqual("number", artifact_open_events[-1]["selection_mode"])
        self.assertEqual("print-path", artifact_open_events[-1]["open_mode"])

    def test_expand_opens_next_ready_artifact_from_current_focus(self) -> None:
        pack_dir = self.compile_research_pack()
        self.assert_ok(self.run_cli("show", "prd", "--from", pack_dir))

        result = self.run_cli("expand", "--print-path")
        self.assert_ok(result)

        self.assertIn(str((pack_dir / "views" / "tasks.md").resolve()), result.stdout.strip())
        events_path = pack_dir / "runtime" / "events.jsonl"
        events = [
            json.loads(line)
            for line in events_path.read_text(encoding="utf-8").splitlines()
            if line.strip()
        ]
        expand_events = [event for event in events if event["event_type"] == "artifact-expand"]
        self.assertTrue(expand_events)
        self.assertEqual("prd", expand_events[-1]["source_artifact_id"])
        self.assertEqual("tasks", expand_events[-1]["artifact_id"])
        self.assertEqual("next-ready", expand_events[-1]["selection_reason"])

    def test_show_more_alias_falls_back_to_export_after_last_ready_artifact(self) -> None:
        pack_dir = self.compile_research_pack()
        self.assert_ok(self.run_cli("continue", "--from", pack_dir, "--print-path"))

        result = self.run_cli("show", "more", "--from", pack_dir, "--print-path", "--format", "json")
        self.assert_ok(result)

        payload = json.loads(result.stdout)
        self.assertEqual("JiniExpandArtifact", payload["result_type"])
        self.assertEqual("tasks", payload["source_artifact"]["id"])
        self.assertEqual("github", payload["artifact"]["id"])
        self.assertEqual("fallback-export", payload["selection_reason"])
        self.assertEqual("print-path", payload["open_mode"])
        self.assertTrue(payload["path"].endswith("exports/issues/github/README.md"))

    def test_rewrite_versions_and_undo_restore_current_artifact(self) -> None:
        pack_dir = self.compile_research_pack()
        artifact_path = pack_dir / "views" / "tasks.md"
        original_text = artifact_path.read_text(encoding="utf-8")
        versions_root = pack_dir / "runtime" / "artifact-versions"

        initial_versions = self.run_cli("versions", "tasks", "--from", pack_dir, "--format", "json")
        self.assert_ok(initial_versions)
        self.assertEqual(0, json.loads(initial_versions.stdout)["version_count"])
        self.assertFalse(versions_root.exists())

        rewrite = self.run_cli(
            "rewrite",
            "checklist",
            "tasks",
            "--from",
            pack_dir,
            "--format",
            "json",
        )
        self.assert_ok(rewrite)

        rewrite_report = json.loads(rewrite.stdout)
        self.assertEqual("JiniArtifactRewrite", rewrite_report["result_type"])
        self.assertEqual("tasks", rewrite_report["artifact_id"])
        self.assertEqual("checklist", rewrite_report["shortcut"])
        self.assertTrue(Path(rewrite_report["snapshot_path"]).exists())
        rewritten_text = artifact_path.read_text(encoding="utf-8")
        self.assertNotEqual(original_text, rewritten_text)
        self.assertIn("- [ ]", rewritten_text)

        versions = self.run_cli("versions", "tasks", "--from", pack_dir, "--format", "json")
        self.assert_ok(versions)
        versions_report = json.loads(versions.stdout)
        self.assertEqual("JiniArtifactVersions", versions_report["result_type"])
        self.assertEqual("tasks", versions_report["artifact_id"])
        self.assertEqual(1, versions_report["version_count"])
        self.assertEqual(rewrite_report["snapshot_id"], versions_report["versions"][0]["snapshot_id"])

        versions_text = self.run_cli("versions", "tasks", "--from", pack_dir)
        self.assert_ok(versions_text)
        self.assertIn("VERSIONS", versions_text.stdout)
        self.assertIn("rewrite:checklist", versions_text.stdout)

        undo = self.run_cli("undo", "tasks", "--from", pack_dir, "--format", "json")
        self.assert_ok(undo)
        undo_report = json.loads(undo.stdout)
        self.assertEqual("JiniArtifactUndo", undo_report["result_type"])
        self.assertEqual(rewrite_report["snapshot_id"], undo_report["restored_snapshot_id"])
        self.assertEqual(original_text, artifact_path.read_text(encoding="utf-8"))

    def test_context_capsule_explains_what_shaped_current_artifact(self) -> None:
        pack_dir = self.compile_research_pack()
        self.assert_ok(self.run_cli("show", "tasks", "--from", pack_dir))

        result = self.run_cli("context", "tasks", "--from", pack_dir, "--format", "json")
        self.assert_ok(result)

        capsule = json.loads(result.stdout)
        self.assertEqual("JiniContextCapsule", capsule["result_type"])
        self.assertEqual("tasks", capsule["artifact"]["id"])
        self.assertEqual("test-research-pack", capsule["work_unit_id"])
        self.assertTrue(any("Test Research Pack" in item["value"] for item in capsule["direct_user_inputs"]))
        self.assertTrue(any("Exercise the research pack lifecycle" in item["value"] for item in capsule["direct_user_inputs"]))
        self.assertTrue(any("Research context document" in item["label"] for item in capsule["source_references"]))
        self.assertTrue(any("12 semi-structured customer interviews" in item["label"] for item in capsule["source_references"]))
        self.assertTrue(any("known_unknown" == item["kind"] for item in capsule["missing_or_uncertain"]))
        self.assertTrue(any("coverage_gap" == item["kind"] for item in capsule["missing_or_uncertain"]))
        self.assertEqual("session-kernel", capsule["route_and_continuity"]["route"]["provider_id"])
        self.assertEqual("continue-existing-work", capsule["route_and_continuity"]["cost_posture"]["current_path"])
        self.assertFalse(capsule["side_effects"]["mutates_artifact"])

        text_result = self.run_cli("context", "tasks", "--from", pack_dir)
        self.assert_ok(text_result)
        self.assertIn("WHAT JINI USED", text_result.stdout)
        self.assertIn("DIRECT INPUTS", text_result.stdout)
        self.assertIn("SOURCES", text_result.stdout)
        self.assertIn("MISSING OR UNCERTAIN", text_result.stdout)
        self.assertNotIn(str(pack_dir.resolve()), text_result.stdout)

    def test_context_capsule_uses_current_work_without_path(self) -> None:
        pack_dir = self.compile_research_pack()
        self.assert_ok(self.run_cli("show", "tasks", "--from", pack_dir))

        result = self.run_cli("context", "--format", "json")
        self.assert_ok(result)

        capsule = json.loads(result.stdout)
        self.assertEqual("JiniContextCapsule", capsule["result_type"])
        self.assertEqual("tasks", capsule["artifact"]["id"])
        self.assertEqual("test-research-pack", capsule["work_unit_id"])
        self.assertEqual("session-kernel", capsule["route_and_continuity"]["route"]["provider_id"])

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
        focused = self.run_cli("show", "itinerary", "--from", pack_dir)
        self.assert_ok(focused)
        shutil.rmtree(pack_dir)
        current_work = self.tmp / ".jini" / "current-work.json"
        current_work.unlink()

        result = self.run_cli("status", "--format", "json")
        self.assert_ok(result)
        report = json.loads(result.stdout)
        self.assertEqual("session-only", report["health"])
        self.assertEqual("test-travel-pack", report["work_unit_id"])
        self.assertTrue(report["ready_now"])
        self.assertEqual("itinerary", report["current_focus"]["id"])
        self.assertEqual("itinerary", report["ready_now"][0]["id"])
        self.assertIn("saved session projection", report["validation_warnings"][0])
        self.assertEqual(["jini continue"], report["continue_with"])

    def test_pathless_continue_falls_back_to_saved_session_projection_when_pack_is_missing(self) -> None:
        pack_dir = self.compile_travel_pack()
        shutil.rmtree(pack_dir)
        current_work = self.tmp / ".jini" / "current-work.json"
        current_work.unlink()

        result = self.run_cli("continue")
        self.assert_ok(result)
        self.assertIn("NEXT", result.stdout)
        self.assertIn("HEALTH session-only", result.stdout)
        self.assertIn("CONTINUE", result.stdout)
        self.assertIn("saved artifact snapshot", result.stdout)
        self.assertIn("jini open", result.stdout)

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

    def test_pathless_resume_prefers_saved_focus_when_pack_is_missing(self) -> None:
        pack_dir = self.compile_travel_pack()

        focused = self.run_cli("show", "itinerary", "--from", pack_dir)
        self.assert_ok(focused)

        refresh = self.run_cli("status", pack_dir, "--format", "json")
        self.assert_ok(refresh)

        shutil.rmtree(pack_dir)
        current_work = self.tmp / ".jini" / "current-work.json"
        current_work.unlink()

        result = self.run_cli("resume", "--format", "json")
        self.assert_ok(result)
        compact = json.loads(result.stdout)
        self.assertEqual("itinerary", compact["current_focus"]["id"])
        self.assertEqual("Itinerary", compact["current_focus"]["label"])
        self.assertEqual("Itinerary", compact["recent_artifacts"][0]["artifact_type"])
        self.assertTrue(any("Focused artifact: `Itinerary`" in item for item in compact["resume_items"]))

    def test_pathless_resume_text_surfaces_saved_focus_when_pack_is_missing(self) -> None:
        pack_dir = self.compile_travel_pack()

        focused = self.run_cli("show", "itinerary", "--from", pack_dir)
        self.assert_ok(focused)

        refresh = self.run_cli("status", pack_dir, "--format", "json")
        self.assert_ok(refresh)

        shutil.rmtree(pack_dir)
        current_work = self.tmp / ".jini" / "current-work.json"
        current_work.unlink()

        result = self.run_cli("resume")
        self.assert_ok(result)
        self.assertIn("FOCUS  itinerary", result.stdout)
        self.assertIn("Focused artifact: `Itinerary` via `jini show itinerary`.", result.stdout)

    def test_pathless_open_falls_back_to_saved_session_snapshot_when_pack_is_missing(self) -> None:
        pack_dir = self.compile_travel_pack()
        shutil.rmtree(pack_dir)
        current_work = self.tmp / ".jini" / "current-work.json"
        current_work.unlink()

        result = self.run_cli("open", "itinerary", "--print-path")
        self.assert_ok(result)
        snapshot_path = Path(result.stdout.strip())
        self.assertTrue(snapshot_path.exists())
        snapshot_text = snapshot_path.read_text(encoding="utf-8")
        self.assertIn("Saved artifact snapshot", snapshot_text)
        self.assertIn("# Itinerary:", snapshot_text)

    def test_pathless_open_without_artifact_prefers_saved_focus_snapshot_when_pack_is_missing(self) -> None:
        pack_dir = self.compile_travel_pack()

        focused = self.run_cli("show", "itinerary", "--from", pack_dir)
        self.assert_ok(focused)

        refresh = self.run_cli("status", pack_dir, "--format", "json")
        self.assert_ok(refresh)

        shutil.rmtree(pack_dir)
        current_work = self.tmp / ".jini" / "current-work.json"
        current_work.unlink()

        result = self.run_cli("open", "--print-path")
        self.assert_ok(result)
        snapshot_path = Path(result.stdout.strip())
        self.assertTrue(snapshot_path.exists())
        snapshot_text = snapshot_path.read_text(encoding="utf-8")
        self.assertIn("Saved artifact snapshot", snapshot_text)
        self.assertIn("# Itinerary:", snapshot_text)
        self.assertNotIn("# Tasks:", snapshot_text)

    def test_pathless_show_falls_back_to_saved_session_snapshot_when_pack_is_missing(self) -> None:
        pack_dir = self.compile_travel_pack()
        shutil.rmtree(pack_dir)
        current_work = self.tmp / ".jini" / "current-work.json"
        current_work.unlink()

        result = self.run_cli("show", "itinerary")
        self.assert_ok(result)
        self.assertIn("Saved artifact snapshot", result.stdout)
        self.assertIn("# Itinerary:", result.stdout)

    def test_pathless_continue_prefers_saved_focus_snapshot_when_pack_is_missing(self) -> None:
        pack_dir = self.compile_travel_pack()

        focused = self.run_cli("show", "itinerary", "--from", pack_dir)
        self.assert_ok(focused)

        refresh = self.run_cli("status", pack_dir, "--format", "json")
        self.assert_ok(refresh)

        shutil.rmtree(pack_dir)
        current_work = self.tmp / ".jini" / "current-work.json"
        current_work.unlink()

        result = self.run_cli("continue")
        self.assert_ok(result)
        self.assertIn("LABEL  Itinerary", result.stdout)
        self.assertIn("# Itinerary:", result.stdout)
        self.assertNotIn("# Tasks:", result.stdout)

    def test_status_without_current_work_fails_cleanly(self) -> None:
        result = self.run_cli("status")
        self.assert_error(result)
        self.assertEqual("", result.stdout)
        self.assertIn("Nothing is in progress yet. Run `jini` to start from the current repo or directory", result.stderr)
        self.assertIn("jini status /path/to/work", result.stderr)

    def test_status_missing_path_returns_friendly_error(self) -> None:
        result = self.run_cli("status", self.tmp / "missing-pack")
        self.assert_error(result)
        self.assertEqual("", result.stdout)
        self.assertIn("ERROR Pack path is missing required Jini files", result.stderr)

    def test_next_missing_path_returns_friendly_error_on_stderr(self) -> None:
        result = self.run_cli("next", self.tmp / "missing-pack")
        self.assert_error(result)
        self.assertEqual("", result.stdout)
        self.assertIn("ERROR Pack path is missing required Jini files", result.stderr)

    def test_resume_missing_path_returns_friendly_error_on_stderr(self) -> None:
        result = self.run_cli("resume", self.tmp / "missing-pack")
        self.assert_error(result)
        self.assertEqual("", result.stdout)
        self.assertIn("ERROR Pack path is missing required Jini files", result.stderr)

    def test_continue_without_current_work_fails_cleanly_on_stderr(self) -> None:
        result = self.run_cli("continue")
        self.assert_error(result)
        self.assertEqual("", result.stdout)
        self.assertIn("Nothing is in progress yet. Run `jini` to start from the current repo or directory", result.stderr)

    def test_open_missing_path_returns_friendly_error_on_stderr(self) -> None:
        result = self.run_cli("open", "itinerary", "--from", self.tmp / "missing-pack", "--print-path")
        self.assert_error(result)
        self.assertEqual("", result.stdout)
        self.assertIn("ERROR Pack path is missing required Jini files", result.stderr)

    def test_recommend_execution_missing_path_returns_friendly_error_on_stderr(self) -> None:
        result = self.run_cli("recommend-execution", self.tmp / "missing-pack")
        self.assert_error(result)
        self.assertEqual("", result.stdout)
        self.assertIn("ERROR Pack path is missing required Jini files", result.stderr)

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
        self.assertIn("competitive-watch", routine_ids)
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

    def test_run_routine_competitive_watch_emits_watch_brief(self) -> None:
        home = self.personal_home()
        self.assert_ok(self.run_cli("bootstrap-home", home, "--owner-name", "Sharad"))

        result = self.run_cli(
            "run-routine",
            home,
            "competitive-watch",
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
        self.assertIn("Competitive Watch", text)
        self.assertIn("Latest Execute Flow", text)
        self.assertIn("Coverage Gaps", text)

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
        self.assertIn("Latest Execute Flow", text)
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

    def test_resolve_adapter_demotes_throttled_runtime_target_when_unpinned(self) -> None:
        env = os.environ.copy()
        env["JINI_RUNTIME_TARGET_HEALTH"] = "codex=throttled"

        result = self.run_cli(
            "resolve-adapter",
            "--capability",
            "pack-guidance",
            "--layer",
            "runtime-target",
            "--format",
            "json",
            env=env,
        )
        self.assert_ok(result)

        resolution = json.loads(result.stdout)
        self.assertEqual("claude-code", resolution["selected"]["id"])
        self.assertEqual("throttled", next(item for item in resolution["matches"] if item["id"] == "codex")["health_status"])
        self.assertIn("codex", resolution["fallbacks"])

    def test_resolve_adapter_keeps_explicit_runtime_target_pin_under_throttle(self) -> None:
        env = os.environ.copy()
        env["JINI_RUNTIME_TARGET_HEALTH"] = "codex=throttled"

        result = self.run_cli(
            "resolve-adapter",
            "--capability",
            "pack-guidance",
            "--layer",
            "runtime-target",
            "--preferred",
            "codex",
            "--format",
            "json",
            env=env,
        )
        self.assert_ok(result)

        resolution = json.loads(result.stdout)
        self.assertEqual("codex", resolution["selected"]["id"])
        self.assertIn(
            "Preferred adapter `codex` was selected explicitly despite `throttled` status.",
            resolution["notes"],
        )
        self.assertFalse(
            any("no healthier higher-priority target was available" in note for note in resolution["notes"])
        )

    def test_recommend_execution_falls_back_from_throttled_policy_runtime_target(self) -> None:
        pack_dir = self.compile_research_pack()
        rollout_dir = pack_dir / "runtime" / "policy-rollouts"
        rollout_dir.mkdir(parents=True, exist_ok=True)
        rollout_path = rollout_dir / "runtime-routing-active.json"
        rollout_path.write_text(
            json.dumps(
                {
                    "schema_version": "0.1.0",
                    "policy_type": "JiniPolicyRollout",
                    "policy_id": "runtime-routing",
                    "candidate_id": "runtime-routing-throttle-test",
                    "status": "active",
                    "recommended_runtime_target": "kiro-cli",
                    "intent_overrides": {},
                    "route_feedback_drivers": {},
                },
                indent=2,
            )
            + "\n",
            encoding="utf-8",
        )
        env = os.environ.copy()
        env["JINI_RUNTIME_TARGET_HEALTH"] = "kiro-cli=throttled"

        result = self.run_cli("recommend-execution", pack_dir, "--intent", "make", "--format", "json", env=env)
        self.assert_ok(result)

        recommendation = json.loads(result.stdout)
        self.assertEqual("kiro-cli", recommendation["active_policy"]["recommended_runtime_target"])
        self.assertEqual("codex", recommendation["runtime_guidance"]["selected"]["id"])
        self.assertTrue(
            any("Policy-preferred adapter `kiro-cli` is `throttled`; switched to `codex`" in note for note in recommendation["runtime_guidance"]["notes"])
        )

    def test_recommend_execution_falls_back_from_persisted_throttled_runtime_target(self) -> None:
        pack_dir = self.compile_research_pack()
        rollout_dir = pack_dir / "runtime" / "policy-rollouts"
        rollout_dir.mkdir(parents=True, exist_ok=True)
        rollout_path = rollout_dir / "runtime-routing-active.json"
        rollout_path.write_text(
            json.dumps(
                {
                    "schema_version": "0.1.0",
                    "policy_type": "JiniPolicyRollout",
                    "policy_id": "runtime-routing",
                    "candidate_id": "runtime-routing-persisted-throttle-test",
                    "status": "active",
                    "recommended_runtime_target": "kiro-cli",
                    "intent_overrides": {},
                    "route_feedback_drivers": {},
                },
                indent=2,
            )
            + "\n",
            encoding="utf-8",
        )
        recorded = self.run_cli(
            "record-runtime-target-outcome",
            pack_dir,
            "--runtime-target",
            "kiro-cli",
            "--status",
            "throttled",
            "--reason",
            "Observed CLI throttling during execution.",
            "--format",
            "json",
        )
        self.assert_ok(recorded)
        recorded_payload = json.loads(recorded.stdout)
        self.assertEqual("kiro-cli", recorded_payload["runtime_target"])
        self.assertEqual("throttled", recorded_payload["status"])
        self.assertEqual("throttled", recorded_payload["health_status"])
        self.assertEqual(1, recorded_payload["health_signal_count"])

        result = self.run_cli("recommend-execution", pack_dir, "--intent", "make", "--format", "json")
        self.assert_ok(result)

        recommendation = json.loads(result.stdout)
        self.assertEqual("kiro-cli", recommendation["active_policy"]["recommended_runtime_target"])
        self.assertEqual("codex", recommendation["runtime_guidance"]["selected"]["id"])
        persisted = next(item for item in recommendation["runtime_guidance"]["matches"] if item["id"] == "kiro-cli")
        self.assertEqual("throttled", persisted["health_status"])
        self.assertEqual("learning-events", persisted["health_source"])
        self.assertEqual(1, persisted["health_signal_count"])
        self.assertTrue(
            any(
                "Policy-preferred adapter `kiro-cli` is `throttled`; switched to `codex`" in note
                for note in recommendation["runtime_guidance"]["notes"]
            )
        )

    def test_env_runtime_target_override_preserves_persisted_health_evidence(self) -> None:
        pack_dir = self.compile_research_pack()
        self.assert_ok(
            self.run_cli(
                "record-runtime-target-outcome",
                pack_dir,
                "--runtime-target",
                "kiro-cli",
                "--status",
                "throttled",
                "--reason",
                "Observed CLI throttling during execution.",
                "--format",
                "json",
            )
        )
        env = os.environ.copy()
        env["JINI_RUNTIME_TARGET_HEALTH"] = "kiro-cli=unavailable"
        merged = jini_validate.load_runtime_target_health(
            path=jini_validate.runtime_events_path(pack_dir),
            env=env,
        )
        merged_row = merged["kiro-cli"]
        self.assertEqual("unavailable", merged_row["status"])
        self.assertEqual("env", merged_row["source"])
        self.assertEqual(1, merged_row["signal_count"])
        self.assertEqual("throttled", merged_row["last_status"])
        self.assertEqual("Observed CLI throttling during execution.", merged_row["last_reason"])
        self.assertTrue(merged_row["last_recorded_at"])

        with mock.patch.dict(os.environ, env, clear=False):
            resolution = jini_validate.build_adapter_resolution(
                capability="pack-guidance",
                layer="runtime-target",
                preferred="kiro-cli",
                learning_events_path=jini_validate.runtime_events_path(pack_dir),
            )
        selected = resolution["selected"]
        self.assertEqual("kiro-cli", selected["id"])
        self.assertEqual("unavailable", selected["health_status"])
        self.assertEqual("env", selected["health_source"])
        self.assertEqual(1, selected["health_signal_count"])
        self.assertEqual("throttled", selected["health_last_status"])
        self.assertEqual("Observed CLI throttling during execution.", selected["health_last_reason"])
        self.assertTrue(selected["health_last_recorded_at"])
        self.assertIn(
            "Preferred adapter `kiro-cli` was selected explicitly despite `unavailable` status.",
            resolution["notes"],
        )

    def test_env_runtime_target_override_normalizes_recovered_to_ready(self) -> None:
        env = os.environ.copy()
        env["JINI_RUNTIME_TARGET_HEALTH"] = "kiro-cli=recovered"

        parsed = jini_validate.parse_runtime_target_health_overrides(env=env)
        self.assertEqual({"kiro-cli": "ready"}, parsed)

        with mock.patch.dict(os.environ, env, clear=False):
            resolution = jini_validate.build_adapter_resolution(
                capability="pack-guidance",
                layer="runtime-target",
                preferred="kiro-cli",
            )
        self.assertEqual("kiro-cli", resolution["selected"]["id"])
        self.assertEqual("ready", resolution["selected"]["health_status"])

    def test_env_runtime_target_override_ignores_unknown_status(self) -> None:
        pack_dir = self.compile_research_pack()
        self.assert_ok(
            self.run_cli(
                "record-runtime-target-outcome",
                pack_dir,
                "--runtime-target",
                "kiro-cli",
                "--status",
                "throttled",
                "--reason",
                "Observed CLI throttling during execution.",
                "--format",
                "json",
            )
        )
        env = os.environ.copy()
        env["JINI_RUNTIME_TARGET_HEALTH"] = "kiro-cli=throttle"

        parsed = jini_validate.parse_runtime_target_health_overrides(env=env)
        self.assertEqual({}, parsed)

        merged = jini_validate.load_runtime_target_health(
            path=jini_validate.runtime_events_path(pack_dir),
            env=env,
        )
        merged_row = merged["kiro-cli"]
        self.assertEqual("throttled", merged_row["status"])
        self.assertEqual("learning-events", merged_row["source"])
        self.assertEqual("throttled", merged_row["last_status"])

    def test_recommend_execution_keeps_policy_runtime_target_when_no_healthier_target_is_available(self) -> None:
        pack_dir = self.compile_research_pack()
        rollout_dir = pack_dir / "runtime" / "policy-rollouts"
        rollout_dir.mkdir(parents=True, exist_ok=True)
        rollout_path = rollout_dir / "runtime-routing-active.json"
        rollout_path.write_text(
            json.dumps(
                {
                    "schema_version": "0.1.0",
                    "policy_type": "JiniPolicyRollout",
                    "policy_id": "runtime-routing",
                    "candidate_id": "runtime-routing-no-healthier-target-test",
                    "status": "active",
                    "recommended_runtime_target": "kiro-cli",
                    "intent_overrides": {},
                    "route_feedback_drivers": {},
                },
                indent=2,
            )
            + "\n",
            encoding="utf-8",
        )
        env = os.environ.copy()
        env["JINI_RUNTIME_TARGET_HEALTH"] = ",".join(
            [
                "kiro-cli=throttled",
                "codex=unavailable",
                "claude-code=unavailable",
                "github-copilot=unavailable",
                "junie=unavailable",
                "augment=unavailable",
            ]
        )

        result = self.run_cli("recommend-execution", pack_dir, "--intent", "make", "--format", "json", env=env)
        self.assert_ok(result)

        recommendation = json.loads(result.stdout)
        self.assertEqual("kiro-cli", recommendation["active_policy"]["recommended_runtime_target"])
        self.assertEqual("kiro-cli", recommendation["runtime_guidance"]["selected"]["id"])
        self.assertTrue(
            any(
                "Policy-preferred adapter `kiro-cli` remained `throttled`; kept it as the selected runtime target." in note
                for note in recommendation["runtime_guidance"]["notes"]
            )
        )
        self.assertTrue(
            any(
                "Selected runtime target `kiro-cli` despite `throttled` status because no healthier higher-priority target was available."
                in note
                for note in recommendation["runtime_guidance"]["notes"]
            )
        )

    def test_recommend_execution_keeps_explicit_runtime_target_pin_under_throttle(self) -> None:
        pack_dir = self.compile_research_pack()
        env = os.environ.copy()
        env["JINI_RUNTIME_TARGET_HEALTH"] = "kiro-cli=throttled"

        result = self.run_cli(
            "recommend-execution",
            pack_dir,
            "--intent",
            "make",
            "--runtime-target",
            "kiro-cli",
            "--format",
            "json",
            env=env,
        )
        self.assert_ok(result)

        recommendation = json.loads(result.stdout)
        self.assertEqual("kiro-cli", recommendation["runtime_guidance"]["selected"]["id"])
        self.assertIn(
            "Preferred adapter `kiro-cli` was selected explicitly despite `throttled` status.",
            recommendation["runtime_guidance"]["notes"],
        )
        self.assertFalse(
            any(
                "no healthier higher-priority target was available" in note
                for note in recommendation["runtime_guidance"]["notes"]
            )
        )

    def test_recommend_execution_keeps_explicit_runtime_target_pin_under_persisted_throttle(self) -> None:
        pack_dir = self.compile_research_pack()
        self.assert_ok(
            self.run_cli(
                "record-runtime-target-outcome",
                pack_dir,
                "--runtime-target",
                "kiro-cli",
                "--status",
                "throttled",
                "--reason",
                "Observed CLI throttling during execution.",
                "--format",
                "json",
            )
        )

        result = self.run_cli(
            "recommend-execution",
            pack_dir,
            "--intent",
            "make",
            "--runtime-target",
            "kiro-cli",
            "--format",
            "json",
        )
        self.assert_ok(result)

        recommendation = json.loads(result.stdout)
        self.assertEqual("kiro-cli", recommendation["runtime_guidance"]["selected"]["id"])
        self.assertIn(
            "Preferred adapter `kiro-cli` was selected explicitly despite `throttled` status.",
            recommendation["runtime_guidance"]["notes"],
        )

    def test_recommend_execution_restores_policy_runtime_target_after_recovered_signal(self) -> None:
        pack_dir = self.compile_research_pack()
        rollout_dir = pack_dir / "runtime" / "policy-rollouts"
        rollout_dir.mkdir(parents=True, exist_ok=True)
        rollout_path = rollout_dir / "runtime-routing-active.json"
        rollout_path.write_text(
            json.dumps(
                {
                    "schema_version": "0.1.0",
                    "policy_type": "JiniPolicyRollout",
                    "policy_id": "runtime-routing",
                    "candidate_id": "runtime-routing-recovered-target-test",
                    "status": "active",
                    "recommended_runtime_target": "kiro-cli",
                    "intent_overrides": {},
                    "route_feedback_drivers": {},
                },
                indent=2,
            )
            + "\n",
            encoding="utf-8",
        )
        self.assert_ok(
            self.run_cli(
                "record-runtime-target-outcome",
                pack_dir,
                "--runtime-target",
                "kiro-cli",
                "--status",
                "throttled",
                "--reason",
                "Observed CLI throttling during execution.",
                "--format",
                "json",
            )
        )
        self.assert_ok(
            self.run_cli(
                "record-runtime-target-outcome",
                pack_dir,
                "--runtime-target",
                "kiro-cli",
                "--status",
                "recovered",
                "--reason",
                "Observed runtime recovery after retry.",
                "--format",
                "json",
            )
        )

        result = self.run_cli("recommend-execution", pack_dir, "--intent", "make", "--format", "json")
        self.assert_ok(result)

        recommendation = json.loads(result.stdout)
        self.assertEqual("kiro-cli", recommendation["active_policy"]["recommended_runtime_target"])
        self.assertEqual("kiro-cli", recommendation["runtime_guidance"]["selected"]["id"])
        recovered = next(item for item in recommendation["runtime_guidance"]["matches"] if item["id"] == "kiro-cli")
        self.assertEqual("ready", recovered["health_status"])
        self.assertEqual("recovered", recovered["health_last_status"])
        self.assertEqual("learning-events", recovered["health_source"])
        self.assertEqual(2, recovered["health_signal_count"])
        self.assertTrue(
            any(
                "Policy-preferred adapter `kiro-cli` recovered to `ready`; kept it as the selected runtime target."
                for note in recommendation["runtime_guidance"]["notes"]
            )
        )
        self.assertFalse(
            any(
                "Policy-preferred adapter `kiro-cli` remained `ready`; kept it as the selected runtime target." in note
                for note in recommendation["runtime_guidance"]["notes"]
            )
        )

    def test_recommend_execution_treats_post_recovery_degraded_target_as_degraded_not_throttled(self) -> None:
        pack_dir = self.compile_research_pack()
        rollout_dir = pack_dir / "runtime" / "policy-rollouts"
        rollout_dir.mkdir(parents=True, exist_ok=True)
        rollout_path = rollout_dir / "runtime-routing-active.json"
        rollout_path.write_text(
            json.dumps(
                {
                    "schema_version": "0.1.0",
                    "policy_type": "JiniPolicyRollout",
                    "policy_id": "runtime-routing",
                    "candidate_id": "runtime-routing-post-recovery-degraded-test",
                    "status": "active",
                    "recommended_runtime_target": "kiro-cli",
                    "intent_overrides": {},
                    "route_feedback_drivers": {},
                },
                indent=2,
            )
            + "\n",
            encoding="utf-8",
        )
        for status, reason in [
            ("throttled", "Observed CLI throttling during execution."),
            ("recovered", "Observed runtime recovery after retry."),
            ("degraded", "Observed slower but usable runtime after recovery."),
        ]:
            self.assert_ok(
                self.run_cli(
                    "record-runtime-target-outcome",
                    pack_dir,
                    "--runtime-target",
                    "kiro-cli",
                    "--status",
                    status,
                    "--reason",
                    reason,
                    "--format",
                    "json",
                )
            )

        env = os.environ.copy()
        env["JINI_RUNTIME_TARGET_HEALTH"] = ",".join(
            [
                "codex=unavailable",
                "claude-code=unavailable",
                "github-copilot=unavailable",
                "junie=unavailable",
                "augment=unavailable",
            ]
        )

        result = self.run_cli("recommend-execution", pack_dir, "--intent", "make", "--format", "json", env=env)
        self.assert_ok(result)

        recommendation = json.loads(result.stdout)
        self.assertEqual("kiro-cli", recommendation["runtime_guidance"]["selected"]["id"])
        selected = recommendation["runtime_guidance"]["selected"]
        self.assertEqual("degraded", selected["health_status"])
        self.assertEqual("degraded", selected["health_last_status"])
        self.assertEqual(-12, selected["health_adjustment"])
        self.assertTrue(
            any(
                "Policy-preferred adapter `kiro-cli` remained `degraded`; kept it as the selected runtime target." in note
                for note in recommendation["runtime_guidance"]["notes"]
            )
        )
        self.assertTrue(
            any(
                "Selected runtime target `kiro-cli` despite `degraded` status because no healthier higher-priority target was available."
                in note
                for note in recommendation["runtime_guidance"]["notes"]
            )
        )

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
        self.assertEqual("deterministic-reapply", receipt["replay_contract"]["mode"])
        self.assertTrue(receipt["replay_contract"]["stable_output_names"])
        self.assertTrue(receipt["replay_contract"]["identity_keys"])
        self.assertTrue(Path(receipt["receipt_path"]).exists())
        self.assertTrue(receipt["applied_paths"])
        first_issue = Path(receipt["applied_paths"][0])
        self.assertTrue(first_issue.exists())
        self.assertIn("## Description", first_issue.read_text(encoding="utf-8"))

    def test_publish_issues_github_reapply_is_stable(self) -> None:
        pack_dir = self.compile_research_pack()

        self.assert_ok(self.run_cli("publish-issues", pack_dir, "--adapter", "github"))
        first = self.run_cli(
            "apply-publish-plan",
            pack_dir / "exports" / "publish" / "issues" / "github",
            "--format",
            "json",
        )
        second = self.run_cli(
            "apply-publish-plan",
            pack_dir / "exports" / "publish" / "issues" / "github",
            "--format",
            "json",
        )
        self.assert_ok(first)
        self.assert_ok(second)

        first_receipt = json.loads(first.stdout)
        second_receipt = json.loads(second.stdout)
        self.assertEqual(first_receipt["applied_paths"], second_receipt["applied_paths"])
        self.assertEqual(first_receipt["replay_contract"], second_receipt["replay_contract"])
        first_issue = Path(first_receipt["applied_paths"][0])
        second_issue = Path(second_receipt["applied_paths"][0])
        self.assertEqual(
            first_issue.read_text(encoding="utf-8"),
            second_issue.read_text(encoding="utf-8"),
        )

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

    def test_publish_issues_github_can_execute_natively(self) -> None:
        pack_dir = self.compile_research_pack()
        gh = self.create_fake_gh_cli()
        env = dict(os.environ)
        env["PATH"] = f"{gh.parent}:{env.get('PATH', '')}"

        result = self.run_cli(
            "publish-issues",
            pack_dir,
            "--adapter",
            "github",
            "--repository",
            "acme/demo",
            "--execute-native",
            "--format",
            "json",
            env=env,
        )
        self.assert_ok(result)

        receipt = json.loads(result.stdout)
        self.assertEqual("executed", receipt["status"])
        self.assertEqual("github", receipt["adapter"])
        self.assertEqual("idempotent-upsert", receipt["replay_contract"]["mode"])
        self.assertEqual("acme/demo", receipt["repository"])
        self.assertTrue(Path(receipt["result_path"]).exists())
        bundle = self.read_json(Path(receipt["result_path"]))
        self.assertEqual(3, len(bundle["records"]))
        self.assertEqual("issue", bundle["records"][0]["target_kind"])
        self.assertEqual("acme/demo", bundle["repository"])
        self.assertEqual(receipt["replay_contract"], bundle["replay_contract"])
        self.assertTrue(Path(receipt["publication_artifact_path"]).exists())
        publication_files = sorted((pack_dir / "artifacts").glob("*-publication.yaml"))
        self.assertEqual(1, len(publication_files))
        publication = self.read_yaml(publication_files[0])
        self.assertEqual("github-issues", publication["publication_scope"])
        self.assertEqual(3, len(publication["records"]))
        self.assertTrue(publication["records"][0]["external_url"].startswith("https://github.com/acme/demo/issues/"))

    def test_recovery_surfaces_latest_native_publication_links(self) -> None:
        pack_dir = self.compile_research_pack()
        gh = self.create_fake_gh_cli()
        env = dict(os.environ)
        env["PATH"] = f"{gh.parent}:{env.get('PATH', '')}"

        publish = self.run_cli(
            "publish-issues",
            pack_dir,
            "--adapter",
            "github",
            "--repository",
            "acme/demo",
            "--execute-native",
            "--format",
            "json",
            env=env,
        )
        self.assert_ok(publish)
        url = "https://github.com/acme/demo/issues/1"

        status_text = self.run_cli("status", pack_dir)
        self.assert_ok(status_text)
        self.assertIn("PUBLISHED", status_text.stdout)
        self.assertIn(url, status_text.stdout)

        status_json = self.run_cli("status", pack_dir, "--format", "json")
        self.assert_ok(status_json)
        status_report = json.loads(status_json.stdout)
        self.assertEqual(url, status_report["published_links"][0]["external_url"])
        self.assertEqual("github-issues", status_report["published_links"][0]["scope"])

        resume_json = self.run_cli("resume", pack_dir, "--format", "json", "--max-chars", "1600")
        self.assert_ok(resume_json)
        compact = json.loads(resume_json.stdout)
        self.assertEqual(url, compact["publication_links"][0]["external_url"])

        resume_text = self.run_cli("resume", pack_dir, "--max-chars", "1600")
        self.assert_ok(resume_text)
        self.assertIn("PUBLISHED", resume_text.stdout)
        self.assertIn(url, resume_text.stdout)

        continue_result = self.run_cli("continue", "--from", pack_dir)
        self.assert_ok(continue_result)
        self.assertIn("PUBLISHED", continue_result.stdout)
        self.assertIn(url, continue_result.stdout)

    def test_execute_publish_plan_native_github_replay_is_upsert_safe(self) -> None:
        pack_dir = self.compile_research_pack()
        gh = self.create_fake_gh_cli()
        env = dict(os.environ)
        env["PATH"] = f"{gh.parent}:{env.get('PATH', '')}"

        staged = self.run_cli(
            "publish-issues",
            pack_dir,
            "--adapter",
            "github",
            "--repository",
            "acme/demo",
            "--format",
            "json",
            env=env,
        )
        self.assert_ok(staged)
        self.assertEqual("native-ready", json.loads(staged.stdout)["execution_mode"])

        first = self.run_cli(
            "execute-publish-plan",
            pack_dir / "exports" / "publish" / "issues" / "github",
            "--native-github",
            "--format",
            "json",
            env=env,
        )
        second = self.run_cli(
            "execute-publish-plan",
            pack_dir / "exports" / "publish" / "issues" / "github",
            "--native-github",
            "--format",
            "json",
            env=env,
        )
        self.assert_ok(first)
        self.assert_ok(second)

        first_receipt = json.loads(first.stdout)
        second_receipt = json.loads(second.stdout)
        self.assertEqual(first_receipt["replay_contract"], second_receipt["replay_contract"])
        self.assertEqual("idempotent-upsert", first_receipt["replay_contract"]["mode"])
        stable_fields = ("source_ref", "external_id", "external_url", "target_kind")
        first_projection = [{key: record[key] for key in stable_fields} for record in first_receipt["records"]]
        second_projection = [{key: record[key] for key in stable_fields} for record in second_receipt["records"]]
        self.assertEqual(first_projection, second_projection)
        self.assertTrue(all(record["publication_status"] == "created" for record in first_receipt["records"]))
        self.assertTrue(all(record["publication_status"] == "updated" for record in second_receipt["records"]))
        publication_files = sorted((pack_dir / "artifacts").glob("*-publication.yaml"))
        self.assertEqual(1, len(publication_files))
        publication = self.read_yaml(publication_files[0])
        self.assertEqual(2, publication["revision"])
        self.assertEqual("github-issues", publication["publication_scope"])
        self.assertTrue(all(record["publication_status"] == "updated" for record in publication["records"]))

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
        self.assertEqual("replay-safe", receipt["replay_contract"]["mode"])
        self.assertTrue(receipt["replay_contract"]["serialized_execution"])
        self.assertTrue(receipt["replay_contract"]["identity_keys"])
        self.assertTrue(Path(receipt["receipt_path"]).exists())
        self.assertTrue(Path(receipt["result_path"]).exists())
        bundle = self.read_json(Path(receipt["result_path"]))
        self.assertEqual(3, len(bundle["records"]))
        self.assertEqual("issue", bundle["records"][0]["target_kind"])
        self.assertEqual(receipt["replay_contract"], bundle["replay_contract"])

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

    def test_execute_publish_plan_replay_is_stable(self) -> None:
        pack_dir = self.compile_research_pack()
        runner = self.create_publish_bridge_runner()
        self.assert_ok(self.run_cli("publish-wiki", pack_dir, "--adapter", "confluence"))

        first = self.run_cli(
            "execute-publish-plan",
            pack_dir / "exports" / "publish" / "wiki" / "confluence",
            "--runner",
            runner,
            "--format",
            "json",
        )
        second = self.run_cli(
            "execute-publish-plan",
            pack_dir / "exports" / "publish" / "wiki" / "confluence",
            "--runner",
            runner,
            "--format",
            "json",
        )
        self.assert_ok(first)
        self.assert_ok(second)

        first_receipt = json.loads(first.stdout)
        second_receipt = json.loads(second.stdout)
        self.assertEqual(first_receipt["replay_contract"], second_receipt["replay_contract"])
        self.assertEqual("replay-safe", first_receipt["replay_contract"]["mode"])
        stable_fields = ("source_ref", "external_id", "external_url", "publication_status", "target_kind")
        first_projection = [{key: record[key] for key in stable_fields} for record in first_receipt["records"]]
        second_projection = [{key: record[key] for key in stable_fields} for record in second_receipt["records"]]
        self.assertEqual(first_projection, second_projection)

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

    def test_publish_wiki_github_docs_can_execute_natively(self) -> None:
        pack_dir = self.compile_research_pack()
        gh = self.create_fake_gh_cli()
        env = dict(os.environ)
        env["PATH"] = f"{gh.parent}:{env.get('PATH', '')}"

        result = self.run_cli(
            "publish-wiki",
            pack_dir,
            "--adapter",
            "github-docs",
            "--repository",
            "acme/demo",
            "--docs-path",
            "docs/jini",
            "--execute-native",
            "--format",
            "json",
            env=env,
        )
        self.assert_ok(result)

        receipt = json.loads(result.stdout)
        self.assertEqual("executed", receipt["status"])
        self.assertEqual("github-docs", receipt["adapter"])
        self.assertEqual("idempotent-upsert", receipt["replay_contract"]["mode"])
        self.assertEqual("acme/demo", receipt["repository"])
        self.assertEqual("docs/jini", receipt["docs_path"])
        self.assertTrue(Path(receipt["result_path"]).exists())
        bundle = self.read_json(Path(receipt["result_path"]))
        self.assertEqual(3, len(bundle["records"]))
        self.assertEqual("wiki-page", bundle["records"][0]["target_kind"])
        self.assertEqual("acme/demo", bundle["repository"])
        self.assertEqual("docs/jini", bundle["docs_path"])
        self.assertEqual(receipt["replay_contract"], bundle["replay_contract"])
        self.assertTrue(Path(receipt["publication_artifact_path"]).exists())
        publication_files = sorted((pack_dir / "artifacts").glob("*-publication.yaml"))
        self.assertEqual(1, len(publication_files))
        publication = self.read_yaml(publication_files[0])
        self.assertEqual("github-docs", publication["publication_scope"])
        self.assertEqual(3, len(publication["records"]))
        self.assertTrue(publication["records"][0]["external_url"].startswith("https://github.com/acme/demo/blob/main/docs/jini/"))

    def test_execute_publish_plan_native_github_docs_replay_is_upsert_safe(self) -> None:
        pack_dir = self.compile_research_pack()
        gh = self.create_fake_gh_cli()
        env = dict(os.environ)
        env["PATH"] = f"{gh.parent}:{env.get('PATH', '')}"

        staged = self.run_cli(
            "publish-wiki",
            pack_dir,
            "--adapter",
            "github-docs",
            "--repository",
            "acme/demo",
            "--docs-path",
            "docs/jini",
            "--format",
            "json",
            env=env,
        )
        self.assert_ok(staged)
        self.assertEqual("native-ready", json.loads(staged.stdout)["execution_mode"])

        first = self.run_cli(
            "execute-publish-plan",
            pack_dir / "exports" / "publish" / "wiki" / "github-docs",
            "--native-github",
            "--format",
            "json",
            env=env,
        )
        second = self.run_cli(
            "execute-publish-plan",
            pack_dir / "exports" / "publish" / "wiki" / "github-docs",
            "--native-github",
            "--format",
            "json",
            env=env,
        )
        self.assert_ok(first)
        self.assert_ok(second)

        first_receipt = json.loads(first.stdout)
        second_receipt = json.loads(second.stdout)
        self.assertEqual(first_receipt["replay_contract"], second_receipt["replay_contract"])
        self.assertEqual("idempotent-upsert", first_receipt["replay_contract"]["mode"])
        stable_fields = ("source_ref", "external_id", "external_url", "target_kind")
        first_projection = [{key: record[key] for key in stable_fields} for record in first_receipt["records"]]
        second_projection = [{key: record[key] for key in stable_fields} for record in second_receipt["records"]]
        self.assertEqual(first_projection, second_projection)
        self.assertTrue(all(record["publication_status"] == "created" for record in first_receipt["records"]))
        self.assertTrue(all(record["publication_status"] == "updated" for record in second_receipt["records"]))
        publication_files = sorted((pack_dir / "artifacts").glob("*-publication.yaml"))
        self.assertEqual(1, len(publication_files))
        publication = self.read_yaml(publication_files[0])
        self.assertEqual(2, publication["revision"])
        self.assertEqual("github-docs", publication["publication_scope"])
        self.assertTrue(all(record["publication_status"] == "updated" for record in publication["records"]))

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
        self.assertFalse(any(" rev " in item for item in compact["resume_items"]))
        self.assertLessEqual(len(compact["recent_artifacts"]), 3)
        self.assertLessEqual(result.stdout.count("\n"), 1)

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
        github = next(item for item in summary["checks"] if item["id"] == "github")
        github_docs = next(item for item in summary["checks"] if item["id"] == "github-docs")
        jira = next(item for item in summary["checks"] if item["id"] == "jira")
        markdown = next(item for item in summary["checks"] if item["id"] == "markdown")
        confluence = next(item for item in summary["checks"] if item["id"] == "confluence")
        self.assertEqual("deterministic-reapply", github["publish_semantics"]["local_apply"])
        self.assertEqual("idempotent-upsert", github["publish_semantics"]["native_execution"])
        self.assertEqual("portable-json", github["publish_semantics"]["receipt_contract"])
        self.assertEqual("deterministic-reapply", github_docs["publish_semantics"]["local_apply"])
        self.assertEqual("idempotent-upsert", github_docs["publish_semantics"]["native_execution"])
        self.assertEqual("portable-json", github_docs["publish_semantics"]["receipt_contract"])
        self.assertEqual("replay-safe", jira["publish_semantics"]["bridge_execution"])
        self.assertEqual("replay-safe", confluence["publish_semantics"]["bridge_execution"])
        self.assertEqual("deterministic-reapply", markdown["publish_semantics"]["local_apply"])
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

    def test_append_learning_event_ignores_global_write_failures(self) -> None:
        from tools import jini_validate

        with mock.patch("tools.jini_validate.Path.mkdir", side_effect=PermissionError("blocked")):
            paths = jini_validate.append_learning_event("compact-context", {"pack_id": "demo"})

        self.assertEqual("", paths["global_path"])
        self.assertIn("PermissionError", paths["global_error"])

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
        self.assertEqual({}, snapshot["latest_execute_flow"])

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
        self.assertEqual({}, backtest["latest_execute_flow"])
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

    def test_activate_runtime_target_records_unavailable_runtime_target_outcome_on_install_failure(self) -> None:
        pack_dir = self.compile_research_pack()
        repo = self.create_repo_fixture()
        self.assert_ok(self.run_cli("bootstrap-steering", repo))
        home = self.personal_home()
        prefix = self.install_prefix()
        self.assert_ok(self.run_cli("bootstrap-home", home))
        self.assert_ok(self.run_cli("bind-home", pack_dir, "--home", home))

        registry = jini_validate.load_registry()
        with mock.patch("tools.jini_validate.install_bundles", side_effect=ValueError("throttled by harness")):
            with self.assertRaisesRegex(ValueError, "throttled by harness"):
                jini_validate.activate_runtime_target(
                    pack_dir,
                    registry,
                    repo_path=repo,
                    home_path=home,
                    runtime_target="codex",
                    prefix=prefix,
                )

        events = jini_validate.read_learning_events(
            path=jini_validate.runtime_events_path(pack_dir),
            event_type="runtime-target-outcome",
        )
        codex_events = [item for item in events if item.get("runtime_target") == "codex"]
        self.assertEqual(1, len(codex_events))
        self.assertEqual("unavailable", codex_events[0]["status"])
        self.assertIn("Runtime activation failed: ValueError: throttled by harness", codex_events[0]["reason"])

        snapshot = jini_validate.build_runtime_target_health_snapshot(path=jini_validate.runtime_events_path(pack_dir))
        codex_health = snapshot["targets"]["codex"]
        self.assertEqual("unavailable", codex_health["status"])
        self.assertEqual("unavailable", codex_health["last_status"])
        self.assertEqual(1, codex_health["signal_count"])

    def test_activate_runtime_target_records_unavailable_runtime_target_outcome_on_subprocess_failure(self) -> None:
        pack_dir = self.compile_research_pack()
        repo = self.create_repo_fixture()
        self.assert_ok(self.run_cli("bootstrap-steering", repo))
        home = self.personal_home()
        prefix = self.install_prefix()
        self.assert_ok(self.run_cli("bootstrap-home", home))
        self.assert_ok(self.run_cli("bind-home", pack_dir, "--home", home))

        registry = jini_validate.load_registry()
        failure = subprocess.CalledProcessError(17, ["codex", "install"], stderr="quota limited")
        with mock.patch("tools.jini_validate.install_bundles", side_effect=failure):
            with self.assertRaises(subprocess.CalledProcessError):
                jini_validate.activate_runtime_target(
                    pack_dir,
                    registry,
                    repo_path=repo,
                    home_path=home,
                    runtime_target="codex",
                    prefix=prefix,
                )

        events = jini_validate.read_learning_events(
            path=jini_validate.runtime_events_path(pack_dir),
            event_type="runtime-target-outcome",
        )
        codex_events = [item for item in events if item.get("runtime_target") == "codex"]
        self.assertEqual(1, len(codex_events))
        self.assertEqual("unavailable", codex_events[0]["status"])
        self.assertIn("Runtime activation failed: CalledProcessError:", codex_events[0]["reason"])

        snapshot = jini_validate.build_runtime_target_health_snapshot(path=jini_validate.runtime_events_path(pack_dir))
        codex_health = snapshot["targets"]["codex"]
        self.assertEqual("unavailable", codex_health["status"])
        self.assertEqual("unavailable", codex_health["last_status"])
        self.assertEqual(1, codex_health["signal_count"])

    def test_activate_runtime_target_records_recovered_runtime_target_outcome_after_prior_throttle(self) -> None:
        pack_dir = self.compile_research_pack()
        repo = self.create_repo_fixture()
        self.assert_ok(self.run_cli("bootstrap-steering", repo))
        home = self.personal_home()
        prefix = self.install_prefix()
        self.assert_ok(self.run_cli("bootstrap-home", home))
        self.assert_ok(self.run_cli("bind-home", pack_dir, "--home", home))

        registry = jini_validate.load_registry()
        initial = jini_validate.record_runtime_target_outcome(
            pack_dir,
            registry,
            runtime_target="codex",
            status="throttled",
            reason="Observed CLI throttling during activation warmup.",
            intent="make",
        )
        self.assertEqual("throttled", initial["health_status"])
        self.assertEqual(1, initial["health_signal_count"])

        activation, _receipt_path = jini_validate.activate_runtime_target(
            pack_dir,
            registry,
            repo_path=repo,
            home_path=home,
            runtime_target="codex",
            prefix=prefix,
        )
        self.assertEqual("codex", activation["runtime_target"])

        events = jini_validate.read_learning_events(
            path=jini_validate.runtime_events_path(pack_dir),
            event_type="runtime-target-outcome",
        )
        codex_statuses = [item["status"] for item in events if item.get("runtime_target") == "codex"]
        self.assertIn("throttled", codex_statuses)
        self.assertIn("recovered", codex_statuses)
        recovered = next(item for item in events if item.get("runtime_target") == "codex" and item.get("status") == "recovered")
        self.assertIn("Runtime activation succeeded after prior `throttled` status.", recovered["reason"])

        snapshot = jini_validate.build_runtime_target_health_snapshot(path=jini_validate.runtime_events_path(pack_dir))
        codex_health = snapshot["targets"]["codex"]
        self.assertEqual("ready", codex_health["status"])
        self.assertEqual("recovered", codex_health["last_status"])
        self.assertEqual(2, codex_health["signal_count"])

    def test_recommend_execution_falls_back_after_activation_failure_records_unavailable_target(self) -> None:
        pack_dir = self.compile_research_pack()
        repo = self.create_repo_fixture()
        self.assert_ok(self.run_cli("bootstrap-steering", repo))
        home = self.personal_home()
        prefix = self.install_prefix()
        self.assert_ok(self.run_cli("bootstrap-home", home))
        self.assert_ok(self.run_cli("bind-home", pack_dir, "--home", home))

        rollout_dir = pack_dir / "runtime" / "policy-rollouts"
        rollout_dir.mkdir(parents=True, exist_ok=True)
        rollout_path = rollout_dir / "runtime-routing-active.json"
        rollout_path.write_text(
            json.dumps(
                {
                    "schema_version": "0.1.0",
                    "policy_type": "JiniPolicyRollout",
                    "policy_id": "runtime-routing",
                    "candidate_id": "runtime-routing-activation-failure-test",
                    "status": "active",
                    "recommended_runtime_target": "codex",
                    "intent_overrides": {},
                    "route_feedback_drivers": {},
                },
                indent=2,
            )
            + "\n",
            encoding="utf-8",
        )

        registry = jini_validate.load_registry()
        with mock.patch("tools.jini_validate.install_bundles", side_effect=ValueError("runtime target unavailable")):
            with self.assertRaisesRegex(ValueError, "runtime target unavailable"):
                jini_validate.activate_runtime_target(
                    pack_dir,
                    registry,
                    repo_path=repo,
                    home_path=home,
                    runtime_target="codex",
                    prefix=prefix,
                )

        result = self.run_cli("recommend-execution", pack_dir, "--intent", "make", "--format", "json")
        self.assert_ok(result)

        recommendation = json.loads(result.stdout)
        self.assertEqual("codex", recommendation["active_policy"]["recommended_runtime_target"])
        self.assertEqual("claude-code", recommendation["runtime_guidance"]["selected"]["id"])
        codex = next(item for item in recommendation["runtime_guidance"]["matches"] if item["id"] == "codex")
        self.assertEqual("unavailable", codex["health_status"])
        self.assertEqual("learning-events", codex["health_source"])
        self.assertTrue(
            any(
                "Preferred adapter `codex` is `unavailable`; switched to `claude-code`" in note
                for note in recommendation["runtime_guidance"]["notes"]
            )
        )

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
        self.assertEqual({}, report["runtime_handoff_summary"])
        self.assertTrue(Path(report["runtime_activation_path"]).exists())
        self.assertEqual({}, report["runtime_activation_summary"])
        self.assertTrue(self.resolve_repo_path(report["run_report_path"]).exists())
        self.assertTrue(report["local_publish_receipts"])
        self.assertTrue(any(item["status"] == "applied-local" for item in report["local_publish_receipts"]))
        self.assertGreater(report["token_strategy"]["compact_estimated_tokens"], 0)
        self.assertIn("compact-context", report["token_strategy"]["reused_context_surfaces"])
        self.assertIn("runtime-handoff", report["token_strategy"]["reused_context_surfaces"])

        no_activation = self.run_cli(
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
        self.assert_ok(no_activation)
        no_activation_doc = json.loads(no_activation.stdout)
        self.assertEqual({}, no_activation_doc["runtime_handoff_summary"])
        self.assertIsNone(no_activation_doc["runtime_activation"])
        self.assertEqual({}, no_activation_doc["runtime_activation_summary"])
        self.assertEqual("", no_activation_doc["runtime_activation_path"])

        no_activation_text = self.run_cli(
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
            "--consent",
            "write",
            "--consent",
            "publish",
            "--issue-adapter",
            "github",
            "--wiki-adapter",
            "markdown",
        )
        self.assert_ok(no_activation_text)
        self.assertNotIn("ACTIVE ", no_activation_text.stdout)
        self.assertNotIn("ACTIVE-DRIVERS ", no_activation_text.stdout)

    def test_review_policy_summarizes_learning_without_mutating_policy(self) -> None:
        pack_dir = self.compile_research_pack()
        repo = self.create_repo_fixture()
        home = self.personal_home()
        state_root = self.tmp / ".jini"
        state_root.mkdir(parents=True, exist_ok=True)
        capabilities_path = state_root / "local-runtime-capabilities.json"
        capabilities_path.write_text(
            json.dumps(
                {
                    "schema_version": "0.4.0",
                    "context_type": "JiniLocalRuntimeCapabilities",
                    "captured_at": "2026-05-21T19:00:00Z",
                    "local_runtime_class": "local-ollama",
                    "adapters": {
                        "local-fast": {
                            "status": "ok",
                            "latency_ms": 85,
                            "warm_latency_ms": 70,
                            "cold_start_cost_ms": 5,
                            "tokens_per_second": 40.0,
                            "quality_class": "usable",
                            "structured_reliability": "usable",
                            "benchmarked_at": "2026-05-21T19:00:00Z",
                        },
                        "local-workhorse": {
                            "status": "ok",
                            "latency_ms": 190,
                            "warm_latency_ms": 160,
                            "cold_start_cost_ms": 30,
                            "tokens_per_second": 22.5,
                            "quality_class": "usable",
                            "structured_reliability": "usable",
                            "benchmarked_at": "2026-05-21T19:00:00Z",
                        },
                    },
                }
            ),
            encoding="utf-8",
        )
        self.assert_ok(self.run_cli("bootstrap-home", home))
        self.assert_ok(self.run_cli("bind-home", pack_dir, "--home", home))
        self.assert_ok(
            self.run_cli(
                "record-route-outcome",
                pack_dir,
                "--adapter-id",
                "local-fast",
                "--intent",
                "export",
                "--outcome",
                "replaced-this",
                "--reason",
                "Needed a stronger export route.",
                "--format",
                "json",
            )
        )
        self.assert_ok(
            self.run_cli(
                "record-route-outcome",
                pack_dir,
                "--adapter-id",
                "local-fast",
                "--intent",
                "wiki",
                "--outcome",
                "replaced-this",
                "--reason",
                "Needed a stronger wiki route.",
                "--format",
                "json",
            )
        )
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
        routing_defaults = [item for item in review["policy_candidates"] if item["kind"] == "routing-default"]
        self.assertTrue(routing_defaults)
        for item in routing_defaults:
            self.assertNotIn("route_feedback_drivers", item)
        self.assertEqual("changed", review["route_feedback_impact"]["status"])
        self.assertEqual(["export", "wiki"], review["route_feedback_impact"]["changed_cohort_keys"])
        self.assertEqual(
            "export:local-fast->local-workhorse,wiki:local-fast->local-workhorse",
            review["route_feedback_impact"]["cohort_preview"]["text"],
        )
        review_text = self.run_cli("review-policy", pack_dir)
        self.assert_ok(review_text)
        self.assertIn(
            "IMPACT changed=2/2 cohorts=export:local-fast->local-workhorse,wiki:local-fast->local-workhorse action=jini review-policy",
            review_text.stdout,
        )
        events = self.run_cli("show-learning-events", pack_dir, "--format", "json")
        self.assert_ok(events)
        event_types = [item["event_type"] for item in json.loads(events.stdout)["events"]]
        self.assertIn("policy-review", event_types)

    def test_review_policy_keeps_unrelated_route_feedback_off_multi_candidate_defaults(self) -> None:
        pack_dir = self.compile_research_pack()
        home = self.personal_home()
        state_root = self.tmp / ".jini"
        state_root.mkdir(parents=True, exist_ok=True)
        capabilities_path = state_root / "local-runtime-capabilities.json"
        capabilities_path.write_text(
            json.dumps(
                {
                    "schema_version": "0.4.0",
                    "context_type": "JiniLocalRuntimeCapabilities",
                    "captured_at": "2026-05-21T19:00:00Z",
                    "local_runtime_class": "local-ollama",
                    "adapters": {
                        "local-fast": {
                            "status": "ok",
                            "latency_ms": 85,
                            "warm_latency_ms": 70,
                            "cold_start_cost_ms": 5,
                            "tokens_per_second": 40.0,
                            "quality_class": "usable",
                            "structured_reliability": "usable",
                            "benchmarked_at": "2026-05-21T19:00:00Z",
                        },
                        "local-workhorse": {
                            "status": "ok",
                            "latency_ms": 190,
                            "warm_latency_ms": 160,
                            "cold_start_cost_ms": 30,
                            "tokens_per_second": 22.5,
                            "quality_class": "usable",
                            "structured_reliability": "usable",
                            "benchmarked_at": "2026-05-21T19:00:00Z",
                        },
                    },
                }
            ),
            encoding="utf-8",
        )
        self.assert_ok(self.run_cli("bootstrap-home", home))
        self.assert_ok(self.run_cli("bind-home", pack_dir, "--home", home))
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
                "event_type": "run-pack",
                "recorded_at": "2026-05-11T10:05:00Z",
                "pack_id": "research-prd",
                "work_unit_id": "test-research-pack",
                "intent": "verify",
                "execution_class": "cheap",
                "state_before": "in_make",
                "state_after": "ready_for_review",
                "blocker_count": 0,
            },
        ]
        events_path.write_text("".join(json.dumps(item, sort_keys=True) + "\n" for item in events), encoding="utf-8")
        self.assert_ok(
            self.run_cli(
                "record-route-outcome",
                pack_dir,
                "--adapter-id",
                "local-fast",
                "--intent",
                "export",
                "--outcome",
                "replaced-this",
                "--reason",
                "Needed a stronger export route.",
                "--format",
                "json",
            )
        )

        review = self.run_cli("review-policy", pack_dir, "--format", "json")
        self.assert_ok(review)
        review_doc = json.loads(review.stdout)
        self.assertEqual(["export"], review_doc["route_feedback_impact"]["changed_cohort_keys"])
        routing_defaults = [
            item
            for item in review_doc["policy_candidates"]
            if item["kind"] == "routing-default"
        ]
        self.assertEqual(["make", "verify"], sorted(item["intent"] for item in routing_defaults))
        for item in routing_defaults:
            self.assertNotIn("route_feedback_drivers", item)

    def test_stage_policy_candidate_merges_multiple_route_feedback_driver_blocks(self) -> None:
        pack_dir = self.compile_research_pack()
        review_path = pack_dir / "runtime" / "policy-reviews" / "synthetic-review.json"
        review_path.parent.mkdir(parents=True, exist_ok=True)
        review_path.write_text(
            json.dumps(
                {
                    "pack_id": "research-prd",
                    "work_unit_id": "test-research-pack",
                    "policy_candidates": [
                        {
                            "kind": "routing-default",
                            "intent": "make",
                            "proposed_execution_class": "cheap",
                            "route_feedback_drivers": {
                                "status": "changed",
                                "changed_cohort_keys": ["export"],
                                "cohort_preview": {
                                    "entries": ["export:local-fast->local-workhorse"],
                                    "remaining_count": 0,
                                    "text": "export:local-fast->local-workhorse",
                                },
                            },
                        },
                        {
                            "kind": "routing-default",
                            "intent": "verify",
                            "proposed_execution_class": "standard",
                            "route_feedback_drivers": {
                                "status": "changed",
                                "changed_cohort_keys": ["wiki"],
                                "cohort_preview": {
                                    "entries": ["wiki:local-fast->local-workhorse"],
                                    "remaining_count": 0,
                                    "text": "wiki:local-fast->local-workhorse",
                                },
                            },
                        },
                    ],
                }
            ),
            encoding="utf-8",
        )

        staged = self.run_cli(
            "stage-policy-candidate",
            pack_dir,
            "--review",
            review_path,
            "--format",
            "json",
        )
        self.assert_ok(staged)
        candidate = json.loads(staged.stdout)
        self.assertEqual(
            ["export", "wiki"],
            candidate["route_feedback_drivers"]["changed_cohort_keys"],
        )
        self.assertEqual(
            "export:local-fast->local-workhorse,wiki:local-fast->local-workhorse",
            candidate["route_feedback_drivers"]["cohort_preview"]["text"],
        )

    def test_policy_candidate_lifecycle_can_activate_and_rollback_routing_override(self) -> None:
        pack_dir = self.compile_research_pack()
        home = self.personal_home()
        state_root = self.tmp / ".jini"
        state_root.mkdir(parents=True, exist_ok=True)
        capabilities_path = state_root / "local-runtime-capabilities.json"
        capabilities_path.write_text(
            json.dumps(
                {
                    "schema_version": "0.4.0",
                    "context_type": "JiniLocalRuntimeCapabilities",
                    "captured_at": "2026-05-21T19:00:00Z",
                    "local_runtime_class": "local-ollama",
                    "adapters": {
                        "local-fast": {
                            "status": "ok",
                            "latency_ms": 85,
                            "warm_latency_ms": 70,
                            "cold_start_cost_ms": 5,
                            "tokens_per_second": 40.0,
                            "quality_class": "usable",
                            "structured_reliability": "usable",
                            "benchmarked_at": "2026-05-21T19:00:00Z",
                        },
                        "local-workhorse": {
                            "status": "ok",
                            "latency_ms": 190,
                            "warm_latency_ms": 160,
                            "cold_start_cost_ms": 30,
                            "tokens_per_second": 22.5,
                            "quality_class": "usable",
                            "structured_reliability": "usable",
                            "benchmarked_at": "2026-05-21T19:00:00Z",
                        },
                    },
                }
            ),
            encoding="utf-8",
        )
        self.assert_ok(self.run_cli("bootstrap-home", home))
        self.assert_ok(self.run_cli("bind-home", pack_dir, "--home", home))
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
        self.assert_ok(
            self.run_cli(
                "record-route-outcome",
                pack_dir,
                "--adapter-id",
                "local-fast",
                "--intent",
                "export",
                "--outcome",
                "replaced-this",
                "--reason",
                "Needed a stronger export route.",
                "--format",
                "json",
            )
        )

        self.assert_ok(self.run_cli("review-policy", pack_dir))
        stage = self.run_cli("stage-policy-candidate", pack_dir, "--format", "json")
        self.assert_ok(stage)
        candidate = json.loads(stage.stdout)
        candidate_path = Path(pack_dir) / "runtime" / "policy-candidates" / f"{candidate['candidate_id']}.json"
        self.assertTrue(candidate_path.exists())
        self.assertEqual("cheap", candidate["intent_overrides"]["make"])
        self.assertEqual(["export"], candidate["route_feedback_drivers"]["changed_cohort_keys"])
        self.assertEqual(
            "export:local-fast->local-workhorse",
            candidate["route_feedback_drivers"]["cohort_preview"]["text"],
        )

        approve = self.run_cli(
            "approve-policy-candidate",
            pack_dir,
            candidate_path,
            "--approver",
            "policy-lead",
        )
        self.assert_ok(approve)
        self.assertIn("DRIVERS export:local-fast->local-workhorse", approve.stdout)
        rollout_path = Path(pack_dir) / "runtime" / "policy-rollouts" / "runtime-routing-active.json"
        rollout = json.loads(rollout_path.read_text(encoding="utf-8"))
        self.assertEqual("active", rollout["status"])
        self.assertEqual("cheap", rollout["intent_overrides"]["make"])
        self.assertEqual(["export"], rollout["route_feedback_drivers"]["changed_cohort_keys"])
        self.assertEqual(
            "export:local-fast->local-workhorse",
            rollout["route_feedback_drivers"]["cohort_preview"]["text"],
        )

        recommendation = self.run_cli("recommend-execution", pack_dir, "--intent", "make", "--format", "json")
        self.assert_ok(recommendation)
        recommendation_doc = json.loads(recommendation.stdout)
        self.assertEqual("cheap", recommendation_doc["execution_class"])
        self.assertEqual(candidate["candidate_id"], recommendation_doc["active_policy"]["candidate_id"])
        self.assertEqual(
            ["export"],
            recommendation_doc["active_policy"]["route_feedback_drivers"]["changed_cohort_keys"],
        )
        self.assertEqual(
            "export:local-fast->local-workhorse",
            recommendation_doc["active_policy"]["route_feedback_drivers"]["cohort_preview"]["text"],
        )
        self.assertEqual("kiro-cli", recommendation_doc["runtime_guidance"]["selected"]["id"])
        recommendation_text = self.run_cli("recommend-execution", pack_dir, "--intent", "make")
        self.assert_ok(recommendation_text)
        self.assertIn("DRIVERS export:local-fast->local-workhorse", recommendation_text.stdout)
        checklist = self.run_cli("next", pack_dir, "--intent", "make", "--format", "json")
        self.assert_ok(checklist)
        checklist_doc = json.loads(checklist.stdout)
        active_rollout = next(item for item in checklist_doc["items"] if item["kind"] == "active-rollout")
        self.assertEqual(
            ["export"],
            active_rollout["route_feedback_drivers"]["changed_cohort_keys"],
        )
        self.assertEqual(
            "export:local-fast->local-workhorse",
            active_rollout["route_feedback_drivers"]["cohort_preview"]["text"],
        )
        checklist_text = self.run_cli("next", pack_dir, "--intent", "make")
        self.assert_ok(checklist_text)
        self.assertIn("drivers=export:local-fast->local-workhorse", checklist_text.stdout)
        handoff = self.run_cli("stage-runtime-handoff", pack_dir, "--intent", "make", "--format", "json")
        self.assert_ok(handoff)
        handoff_doc = json.loads(handoff.stdout)
        self.assertEqual(
            ["export"],
            handoff_doc["active_policy"]["route_feedback_drivers"]["changed_cohort_keys"],
        )
        self.assertEqual(
            "export:local-fast->local-workhorse",
            handoff_doc["active_policy"]["route_feedback_drivers"]["cohort_preview"]["text"],
        )
        self.assertEqual(
            "export:local-fast->local-workhorse",
            handoff_doc["route_feedback_driver_preview"],
        )
        self.assertTrue(
            any("drivers=export:local-fast->local-workhorse" in item for item in handoff_doc["handoff_steps"])
        )
        handoff_text = self.run_cli("stage-runtime-handoff", pack_dir, "--intent", "make")
        self.assert_ok(handoff_text)
        self.assertIn("DRIVERS export:local-fast->local-workhorse", handoff_text.stdout)
        activation_prefix = self.install_prefix()
        activation = self.run_cli(
            "activate-runtime-target",
            pack_dir,
            "--intent",
            "make",
            "--prefix",
            activation_prefix,
            "--format",
            "json",
        )
        self.assert_ok(activation)
        activation_doc = json.loads(activation.stdout)
        self.assertEqual(["export"], activation_doc["route_feedback_drivers"]["changed_cohort_keys"])
        self.assertEqual(
            "export:local-fast->local-workhorse",
            activation_doc["route_feedback_drivers"]["cohort_preview"]["text"],
        )
        self.assertEqual(
            "export:local-fast->local-workhorse",
            activation_doc["route_feedback_driver_preview"],
        )
        activation_markdown_path = next(
            Path(path)
            for path in activation_doc["activation_files"]
            if path.endswith("Jini-RUNTIME.md")
        )
        activation_markdown = activation_markdown_path.read_text(encoding="utf-8")
        self.assertIn("## Active Policy", activation_markdown)
        self.assertIn("- Drivers: `export:local-fast->local-workhorse`", activation_markdown)
        self.assertIn(
            "- [recommended] Active policy rollout", activation_markdown
        )
        self.assertIn("drivers=export:local-fast->local-workhorse", activation_markdown)
        activation_text = self.run_cli(
            "activate-runtime-target",
            pack_dir,
            "--intent",
            "make",
            "--prefix",
            activation_prefix,
        )
        self.assert_ok(activation_text)
        self.assertIn("DRIVERS export:local-fast->local-workhorse", activation_text.stdout)
        flow = self.run_cli(
            "run",
            pack_dir,
            "--mode",
            "supervised",
            "--intent",
            "make",
            "--activate-runtime",
            "--prefix",
            activation_prefix,
            "--format",
            "json",
        )
        self.assert_ok(flow)
        flow_doc = json.loads(flow.stdout)
        self.assertEqual(
            {"route_feedback_driver_preview": "export:local-fast->local-workhorse"},
            flow_doc["runtime_handoff_summary"],
        )
        self.assertEqual(
            ["export"],
            flow_doc["runtime_activation"]["route_feedback_drivers"]["changed_cohort_keys"],
        )
        self.assertEqual(
            "export:local-fast->local-workhorse",
            flow_doc["runtime_activation"]["route_feedback_drivers"]["cohort_preview"]["text"],
        )
        self.assertEqual(
            "export:local-fast->local-workhorse",
            flow_doc["runtime_activation"]["route_feedback_driver_preview"],
        )
        self.assertEqual(
            {"route_feedback_driver_preview": "export:local-fast->local-workhorse"},
            flow_doc["runtime_activation_summary"],
        )
        flow_text = self.run_cli(
            "run",
            pack_dir,
            "--mode",
            "supervised",
            "--intent",
            "make",
            "--activate-runtime",
            "--prefix",
            activation_prefix,
        )
        self.assert_ok(flow_text)
        self.assertIn("ACTIVE-DRIVERS export:local-fast->local-workhorse", flow_text.stdout)
        self.assertNotIn("HANDOFF-DRIVERS ", flow_text.stdout)

        handoff_flow = self.run_cli(
            "run",
            pack_dir,
            "--mode",
            "supervised",
            "--intent",
            "make",
            "--format",
            "json",
        )
        self.assert_ok(handoff_flow)
        handoff_flow_doc = json.loads(handoff_flow.stdout)
        self.assertEqual(
            {"route_feedback_driver_preview": "export:local-fast->local-workhorse"},
            handoff_flow_doc["runtime_handoff_summary"],
        )
        self.assertIsNone(handoff_flow_doc["runtime_activation"])
        self.assertEqual({}, handoff_flow_doc["runtime_activation_summary"])
        self.assertEqual("", handoff_flow_doc["runtime_activation_path"])
        handoff_flow_text = self.run_cli(
            "run",
            pack_dir,
            "--mode",
            "supervised",
            "--intent",
            "make",
        )
        self.assert_ok(handoff_flow_text)
        self.assertIn("HANDOFF-DRIVERS export:local-fast->local-workhorse", handoff_flow_text.stdout)
        self.assertNotIn("ACTIVE-DRIVERS ", handoff_flow_text.stdout)

        learning_snapshot = self.run_cli("learning-snapshot", pack_dir, "--format", "json")
        self.assert_ok(learning_snapshot)
        learning_snapshot_doc = json.loads(learning_snapshot.stdout)
        self.assertEqual(
            "export:local-fast->local-workhorse",
            learning_snapshot_doc["latest_execute_flow"]["route_feedback_driver_preview"],
        )
        self.assertEqual("kiro-cli", learning_snapshot_doc["latest_execute_flow"]["runtime_target"])
        learning_snapshot_text = self.run_cli("learning-snapshot", pack_dir)
        self.assert_ok(learning_snapshot_text)
        self.assertIn("FLOW   research-prd make cheap", learning_snapshot_text.stdout)
        self.assertIn("TARGET kiro-cli", learning_snapshot_text.stdout)
        self.assertIn("DRIVERS export:local-fast->local-workhorse", learning_snapshot_text.stdout)

        routing_backtest = self.run_cli("routing-backtest", pack_dir, "--format", "json")
        self.assert_ok(routing_backtest)
        routing_backtest_doc = json.loads(routing_backtest.stdout)
        self.assertEqual(
            "export:local-fast->local-workhorse",
            routing_backtest_doc["latest_execute_flow"]["route_feedback_driver_preview"],
        )
        self.assertEqual("kiro-cli", routing_backtest_doc["latest_execute_flow"]["runtime_target"])
        routing_backtest_text = self.run_cli("routing-backtest", pack_dir)
        self.assert_ok(routing_backtest_text)
        self.assertIn("FLOW   research-prd make cheap", routing_backtest_text.stdout)
        self.assertIn("TARGET kiro-cli", routing_backtest_text.stdout)
        self.assertIn("DRIVERS export:local-fast->local-workhorse", routing_backtest_text.stdout)

        route_feedback = self.run_cli("route-feedback", "--format", "json")
        self.assert_ok(route_feedback)
        route_feedback_doc = json.loads(route_feedback.stdout)
        self.assertEqual("research-prd", route_feedback_doc["latest_execute_flow"]["pack_id"])
        self.assertEqual("make", route_feedback_doc["latest_execute_flow"]["intent"])
        self.assertEqual("cheap", route_feedback_doc["latest_execute_flow"]["execution_class"])
        self.assertEqual("kiro-cli", route_feedback_doc["latest_execute_flow"]["runtime_target"])
        self.assertEqual(
            "export:local-fast->local-workhorse",
            route_feedback_doc["latest_execute_flow"]["route_feedback_driver_preview"],
        )
        route_feedback_text = self.run_cli("route-feedback")
        self.assert_ok(route_feedback_text)
        self.assertIn("FLOW   research-prd make cheap", route_feedback_text.stdout)
        self.assertIn("TARGET kiro-cli", route_feedback_text.stdout)
        self.assertIn("DRIVERS export:local-fast->local-workhorse", route_feedback_text.stdout)

        review_after_flow = self.run_cli("review-policy", pack_dir, "--format", "json")
        self.assert_ok(review_after_flow)
        review_after_flow_doc = json.loads(review_after_flow.stdout)
        self.assertEqual(
            "export:local-fast->local-workhorse",
            review_after_flow_doc["learning_snapshot"]["latest_execute_flow"]["route_feedback_driver_preview"],
        )
        self.assertEqual("kiro-cli", review_after_flow_doc["learning_snapshot"]["latest_execute_flow"]["runtime_target"])
        review_after_flow_text = self.run_cli("review-policy", pack_dir)
        self.assert_ok(review_after_flow_text)
        self.assertIn("FLOW   research-prd make cheap", review_after_flow_text.stdout)
        self.assertIn("TARGET kiro-cli", review_after_flow_text.stdout)
        self.assertIn("DRIVERS export:local-fast->local-workhorse", review_after_flow_text.stdout)

        readiness = self.run_cli("publish-readiness", "--format", "json")
        self.assert_ok(readiness)
        readiness_doc = json.loads(readiness.stdout)
        self.assertEqual("research-prd", readiness_doc["latest_execute_flow"]["pack_id"])
        self.assertEqual("test-research-pack", readiness_doc["latest_execute_flow"]["work_unit_id"])
        self.assertEqual("make", readiness_doc["latest_execute_flow"]["intent"])
        self.assertEqual("cheap", readiness_doc["latest_execute_flow"]["execution_class"])
        self.assertEqual("kiro-cli", readiness_doc["latest_execute_flow"]["runtime_target"])
        self.assertFalse(readiness_doc["latest_execute_flow"]["activate_runtime"])
        self.assertEqual(
            "export:local-fast->local-workhorse",
            readiness_doc["latest_execute_flow"]["route_feedback_driver_preview"],
        )
        readiness_text = self.run_cli("publish-readiness")
        self.assert_ok(readiness_text)
        self.assertIn("FLOW   research-prd make cheap", readiness_text.stdout)
        self.assertIn("TARGET kiro-cli", readiness_text.stdout)
        self.assertIn("DRIVERS export:local-fast->local-workhorse", readiness_text.stdout)

        status_after_flow = self.run_cli("status", pack_dir, "--format", "json")
        self.assert_ok(status_after_flow)
        status_after_flow_doc = json.loads(status_after_flow.stdout)
        self.assertEqual("research-prd", status_after_flow_doc["latest_execute_flow"]["pack_id"])
        self.assertEqual("make", status_after_flow_doc["latest_execute_flow"]["intent"])
        self.assertEqual("cheap", status_after_flow_doc["latest_execute_flow"]["execution_class"])
        self.assertEqual("kiro-cli", status_after_flow_doc["latest_execute_flow"]["runtime_target"])
        self.assertEqual(
            "export:local-fast->local-workhorse",
            status_after_flow_doc["latest_execute_flow"]["route_feedback_driver_preview"],
        )
        status_after_flow_text = self.run_cli("status", pack_dir)
        self.assert_ok(status_after_flow_text)
        self.assertIn("flow=research-prd make cheap", status_after_flow_text.stdout)
        self.assertIn("target=kiro-cli", status_after_flow_text.stdout)
        self.assertIn("drivers=export:local-fast->local-workhorse", status_after_flow_text.stdout)

        metrics_after_flow = self.run_cli("metrics", "--format", "json")
        self.assert_ok(metrics_after_flow)
        metrics_after_flow_doc = json.loads(metrics_after_flow.stdout)
        self.assertEqual("research-prd", metrics_after_flow_doc["latest_execute_flow"]["pack_id"])
        self.assertEqual("make", metrics_after_flow_doc["latest_execute_flow"]["intent"])
        self.assertEqual("cheap", metrics_after_flow_doc["latest_execute_flow"]["execution_class"])
        self.assertEqual("kiro-cli", metrics_after_flow_doc["latest_execute_flow"]["runtime_target"])
        self.assertEqual(
            "export:local-fast->local-workhorse",
            metrics_after_flow_doc["latest_execute_flow"]["route_feedback_driver_preview"],
        )
        metrics_after_flow_text = self.run_cli("metrics")
        self.assert_ok(metrics_after_flow_text)
        self.assertIn(
            "ROUTEFLOW pack=research-prd intent=make class=cheap target=kiro-cli drivers=export:local-fast->local-workhorse",
            metrics_after_flow_text.stdout,
        )

        publish_routine = self.run_cli(
            "run-routine",
            home,
            "publish-readiness",
            "--mode",
            "local",
            "--format",
            "json",
        )
        self.assert_ok(publish_routine)
        publish_routine_doc = json.loads(publish_routine.stdout)
        publish_brief_path = Path(publish_routine_doc["output_paths"][0])
        publish_brief = publish_brief_path.read_text(encoding="utf-8")
        self.assertIn("## Latest Execute Flow", publish_brief)
        self.assertIn("- Runtime Target: `kiro-cli`", publish_brief)
        self.assertIn("- Drivers: `export:local-fast->local-workhorse`", publish_brief)

        rollback = self.run_cli(
            "rollback-policy-candidate",
            pack_dir,
            candidate_path,
            "--actor",
            "policy-lead",
            "--reason",
            "Restore baseline routing after evaluation",
        )
        self.assert_ok(rollback)
        self.assertIn("DRIVERS export:local-fast->local-workhorse", rollback.stdout)
        rollback_path = sorted((pack_dir / "runtime" / "policy-rollouts").glob("*rollback*.json"))[-1]
        rollback_doc = json.loads(rollback_path.read_text(encoding="utf-8"))
        self.assertEqual("rolled-back", rollback_doc["status"])
        self.assertEqual(["export"], rollback_doc["route_feedback_drivers"]["changed_cohort_keys"])
        self.assertEqual(
            "export:local-fast->local-workhorse",
            rollback_doc["route_feedback_drivers"]["cohort_preview"]["text"],
        )

        restored = self.run_cli("recommend-execution", pack_dir, "--intent", "make", "--format", "json")
        self.assert_ok(restored)
        restored_doc = json.loads(restored.stdout)
        self.assertEqual("standard", restored_doc["execution_class"])
        self.assertEqual({}, restored_doc["active_policy"])

    def test_show_kpis_json_reports_ranked_dimensions(self) -> None:
        result = self.run_cli("show-kpis", "--format", "json", "--limit", "3")
        self.assert_ok(result)

        summary = json.loads(result.stdout)
        self.assertEqual("2026-05-28", summary["updated_at"])
        self.assertIn("Claude Code", summary["core_benchmark_set"])
        self.assertIn("Windsurf", summary["watchlist"])
        self.assertEqual(3, len(summary["dimensions"]))
        self.assertEqual("workflow-rigor", summary["dimensions"][0]["id"])
        self.assertEqual("Kiro", summary["dimensions"][0]["strongest_competitor"]["name"])

    def test_publish_readiness_reports_publishable_surface(self) -> None:
        result = self.run_cli("publish-readiness", "--format", "json")
        self.assert_ok(result)

        report = json.loads(result.stdout)
        self.assertEqual("JiniPublishReadiness", report["result_type"])
        self.assertEqual("ok", report["status"])
        self.assertEqual("starter-kit", report["default_kit_id"])
        self.assertIsInstance(report["latest_execute_flow"], dict)
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
        self.assertEqual("ok", rewrite_momentum["status"])
        self.assertTrue(any(item["id"] == "overall-score-floor" and item["delta"] > 0 for item in rewrite_momentum["checks"]))
        self.assertTrue(
            any(
                item["id"] == "overall-lead-margin" and item["margin"] >= item["minimum_margin"] and item["status"] == "ok"
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

    def test_provider_doctor_reports_local_slm_with_device_runtime_evidence(self) -> None:
        state_root = self.tmp / ".jini"
        state_root.mkdir(parents=True, exist_ok=True)
        (state_root / "device-profile.json").write_text(
            json.dumps(
                {
                    "schema_version": "0.2.0",
                    "context_type": "JiniDeviceProfile",
                    "captured_at": "2026-05-30T14:00:00Z",
                    "os": "darwin",
                    "os_version": "15.5",
                    "arch": "arm64",
                    "accelerator_class": "apple-gpu",
                    "local_runtime_class": "local-ollama",
                    "device_class": "laptop-strong",
                    "local_profile_states": {
                        "local-fast": "ready",
                        "local-workhorse": "ready",
                        "local-deep": "degraded",
                        "local-multimodal": "ready",
                    },
                }
            ),
            encoding="utf-8",
        )
        env = os.environ.copy()
        env.update(
            {
                "JINI_PROVIDER": "local-slm",
                "JINI_LOCAL_SLM_ENDPOINT": "http://127.0.0.1:11434/v1",
                "JINI_LOCAL_SLM_MODEL": "qwen3:8b",
                "JINI_LOCAL_SLM_FAST_MODEL": "phi4-mini",
                "JINI_LOCAL_SLM_API_KEY": "top-secret-local-key",
            }
        )
        result = self.run_cli("doctor", "--format", "json", env=env)
        self.assert_ok(result)

        report = json.loads(result.stdout)
        self.assertEqual("local-slm", report["provider_id"])
        self.assertEqual("ok", report["status"])
        self.assertEqual("Local SLM", report["label"])
        settings = {item["name"]: item["presence"] for item in report["settings"]}
        self.assertEqual("set", settings["JINI_LOCAL_SLM_ENDPOINT"])
        self.assertEqual("laptop-strong", settings["DEVICE_CLASS"])
        self.assertEqual("darwin 15.5", settings["DEVICE_OS"])
        self.assertEqual("apple-gpu", settings["LOCAL_ACCELERATOR"])
        self.assertEqual("local-ollama", settings["LOCAL_RUNTIME_CLASS"])
        self.assertEqual("set", settings["JINI_LOCAL_SLM_MODEL"])
        self.assertEqual("set (ready on this device)", settings["JINI_LOCAL_SLM_FAST_MODEL"])
        self.assertEqual("missing (ready on this device)", settings["JINI_LOCAL_SLM_WORKHORSE_MODEL"])
        self.assertEqual("missing (degraded on this device)", settings["JINI_LOCAL_SLM_DEEP_MODEL"])
        rendered = json.dumps(report)
        self.assertIn("JINI_LOCAL_SLM_API_KEY", rendered)
        self.assertNotIn("top-secret-local-key", rendered)

    def test_provider_doctor_local_slm_requires_endpoint_and_model(self) -> None:
        env = os.environ.copy()
        env["JINI_PROVIDER"] = "local-slm"
        result = self.run_cli("doctor", "--format", "json", env=env)
        self.assertNotEqual(0, result.returncode)

        report = json.loads(result.stdout)
        self.assertEqual("local-slm", report["provider_id"])
        self.assertEqual("needs setup", report["status"])
        self.assertIn("JINI_LOCAL_SLM_ENDPOINT", report["missing"])
        self.assertIn("JINI_LOCAL_SLM_MODEL", report["missing"])

    def test_provider_doctor_auto_prefers_local_slm_when_ready(self) -> None:
        state_root = self.tmp / ".jini"
        state_root.mkdir(parents=True, exist_ok=True)
        (state_root / "device-profile.json").write_text(
            json.dumps(
                {
                    "schema_version": "0.2.0",
                    "context_type": "JiniDeviceProfile",
                    "captured_at": "2026-05-30T14:00:00Z",
                    "os": "darwin",
                    "os_version": "15.5",
                    "accelerator_class": "apple-gpu",
                    "local_runtime_class": "local-ollama",
                    "device_class": "laptop-strong",
                }
            ),
            encoding="utf-8",
        )
        env = os.environ.copy()
        env.update(
            {
                "JINI_PROVIDER": "auto",
                "JINI_LOCAL_SLM_ENDPOINT": "http://127.0.0.1:11434/v1",
                "JINI_LOCAL_SLM_MODEL": "qwen3:8b",
            }
        )
        result = self.run_cli("doctor", "--format", "json", env=env)
        self.assert_ok(result)

        report = json.loads(result.stdout)
        self.assertEqual("local-slm", report["provider_id"])
        self.assertEqual("ok", report["status"])
        provider_setting = next(item for item in report["settings"] if item["name"] == "JINI_PROVIDER")
        self.assertEqual("auto -> Local SLM", provider_setting["presence"])

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
        self.assertEqual("leading", report["overall"]["tracks"]["adoption-truth"]["status"])
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

    def test_competitive_watch_surfaces_freshness_coverage_and_blockers(self) -> None:
        result = self.run_cli("competitive-watch", "--format", "json")
        self.assert_ok(result)

        report = json.loads(result.stdout)
        self.assertEqual("JiniCompetitiveWatch", report["result_type"])
        self.assertIn(report["status"], {"ok", "warning"})
        self.assertEqual("ok", report["coverage"]["status"])
        self.assertEqual([], report["coverage"]["missing_from_benchmark_core"])
        self.assertEqual([], report["coverage"]["missing_from_scorecard_core"])
        self.assertIn("Windsurf", report["coverage"]["watchlist_only"])
        self.assertIn("Claude Code", report["scorecard"]["core_benchmark_set"])
        self.assertEqual("ok", report["score_truth"]["status"])
        self.assertGreater(report["benchmark"]["competitor_count"], 0)
        self.assertGreater(report["benchmark"]["scenario_count"], 0)
        self.assertTrue(all(item["source_urls"] for item in report["benchmark"]["competitors"]))
        self.assertTrue(report["freshness"]["checks"])
        self.assertEqual([], report["replacement_critical"]["blocked_dimensions"])
        self.assertEqual([], report["next_actions"])
        self.assertIsInstance(report["latest_execute_flow"], dict)

        text_result = self.run_cli("competitive-watch")
        self.assert_ok(text_result)
        self.assertIn("FLOW   ", text_result.stdout)

    def test_framework_review_surfaces_latest_execute_flow(self) -> None:
        result = self.run_cli("review-framework", "--format", "json")
        self.assert_ok(result)

        review = json.loads(result.stdout)
        self.assertEqual("JiniFrameworkEvolutionReview", review["review_type"])
        self.assertIsInstance(review["latest_execute_flow"], dict)

        text_result = self.run_cli("review-framework")
        self.assert_ok(text_result)
        self.assertIn("FLOW   ", text_result.stdout)

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
        self.assertIsInstance(experiment["latest_execute_flow"], dict)

        text_result = self.run_cli(
            "stage-framework-experiment",
            "--dimension",
            "delivery-maturity",
        )
        self.assert_ok(text_result)
        if experiment["latest_execute_flow"]:
            self.assertIn("FLOW     ", text_result.stdout)

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
        )
        self.assert_ok(outcome)

        updated_experiment = self.read_json(experiment_path)
        self.assertEqual("completed", updated_experiment["status"])
        outcome_doc = self.read_json(self.resolve_repo_path(updated_experiment["latest_outcome_path"]))
        self.assertEqual("JiniFrameworkEvolutionOutcome", outcome_doc["outcome_type"])
        self.assertEqual("success", outcome_doc["result"])
        self.assertTrue(self.resolve_repo_path(outcome_doc["outcome_path"]).exists())
        self.assertGreater(outcome_doc["computed_reward"], 0)
        self.assertIsInstance(outcome_doc["latest_execute_flow"], dict)
        if outcome_doc["latest_execute_flow"]:
            self.assertIn("FLOW     ", outcome.stdout)

        backtest = self.run_cli("backtest-framework-evolution", "--format", "json")
        self.assert_ok(backtest)
        backtest_doc = json.loads(backtest.stdout)
        self.assertGreaterEqual(backtest_doc["outcome_count"], 1)
        self.assertIsInstance(backtest_doc["latest_execute_flow"], dict)
        adapter_summary = next(
            item for item in backtest_doc["dimension_summaries"] if item["dimension_id"] == "adapter-portability"
        )
        self.assertGreaterEqual(adapter_summary["experiments"], 1)
        self.assertGreaterEqual(adapter_summary["successes"], 1)
        self.assertGreater(adapter_summary["average_score_delta"], 0)

        backtest_text = self.run_cli("backtest-framework-evolution")
        self.assert_ok(backtest_text)
        if backtest_doc["latest_execute_flow"]:
            self.assertIn("FLOW    ", backtest_text.stdout)

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
        self.assertIn("CMDS     4", result.stdout)
        self.assertIn("  - continue", result.stdout)
        self.assertIn("  - doctor", result.stdout)
        self.assertIn("  - open", result.stdout)
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
        self.assertEqual(4, payload["command_surface_count"])
        self.assertEqual(0, payload["compatibility_alias_count"])
        self.assertEqual(["continue", "doctor", "open", "status"], payload["taught_commands"])
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

    def test_metrics_json_surfaces_local_slm_provider_when_configured(self) -> None:
        state_root = self.tmp / ".jini"
        state_root.mkdir(parents=True, exist_ok=True)
        (state_root / "device-profile.json").write_text(
            json.dumps(
                {
                    "schema_version": "0.2.0",
                    "context_type": "JiniDeviceProfile",
                    "captured_at": "2026-05-30T14:00:00Z",
                    "os": "darwin",
                    "os_version": "15.5",
                    "accelerator_class": "apple-gpu",
                    "local_runtime_class": "local-ollama",
                    "device_class": "laptop-strong",
                }
            ),
            encoding="utf-8",
        )
        env = os.environ.copy()
        env.update(
            {
                "JINI_PROVIDER": "local-slm",
                "JINI_LOCAL_SLM_ENDPOINT": "http://127.0.0.1:11434/v1",
                "JINI_LOCAL_SLM_MODEL": "qwen3:8b",
            }
        )
        result = self.run_cli("metrics", "--format", "json", env=env)
        self.assert_ok(result)
        payload = json.loads(result.stdout)
        self.assertTrue(payload["provider_evidence"]["available"])
        self.assertEqual("local-slm", payload["provider_evidence"]["provider_id"])
        self.assertEqual("Local SLM", payload["provider_evidence"]["label"])
        self.assertEqual("ok", payload["provider_evidence"]["status"])


if __name__ == "__main__":
    unittest.main()
