from __future__ import annotations

import sys
import unittest
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
if str(ROOT) not in sys.path:
    sys.path.insert(0, str(ROOT))

from tools.help_tail_contract import HELP_TAIL_EXAMPLE_REQUEST, help_tail_error_line, help_tail_message_lines

README_PATH = ROOT / "README.md"
CLI_PATH = ROOT / "docs" / "cli.md"
INSTALL_PATH = ROOT / "docs" / "install.md"
SIMPLE_PATH = ROOT / "docs" / "simple.md"


def read(path: Path) -> str:
    return path.read_text(encoding="utf-8")


class PublicCliDocsTests(unittest.TestCase):
    def test_readme_points_public_users_at_commands_and_admin_help(self) -> None:
        text = read(README_PATH)
        for marker in (
            "jini commands",
            "small public command catalog",
            "jini admin help",
            "routes, bundles, or release plumbing",
            "`jini status`",
            "`jini continue`",
            "`jini open`",
            "`jini doctor`",
            "`what is blocked?`",
            "`open the latest artifact`",
            "`plan this change`",
            "Jini should tell you what to connect instead of making you memorize route phrases",
        ):
            with self.subTest(marker=marker):
                self.assertIn(marker, text)

    def test_cli_guide_promotes_commands_as_public_catalog(self) -> None:
        text = read(CLI_PATH)
        for marker in (
            "title: Command Catalog",
            "Command catalog",
            "jini commands",
            "small product-facing command catalog",
            "jini admin help",
            "routes, bundles, or release plumbing",
            "Support commands when Jini points you there",
            "not because they are navigating a command tree by hand",
            "Help surfaces and <code>jini commands</code> are catalogs, not request entrypoints.",
            "jini help me edit notes.txt",
            "jini commands me edit notes.txt",
            "jini --help me edit notes.txt",
            "jini provider help me edit notes.txt",
            "jini review this repo",
            "jini fix failing tests",
            "jini review this branch",
            "That repo-aware start surface should stay light: one calm repo-context line, direct task suggestions, one or two useful commands Jini found in the repo, and at most one quiet adoption hint for existing Jini work.",
            "Internal diagnostics like <code>repo-map</code> and setup surfaces like <code>doctor</code> should stay off the first screen.",
            "A Claude Code, Codex, or GitHub CLI user should see task suggestions first, then at most a small <code>Already have Jini work?</code> note with <code>status /path/to/work</code> when an existing Jini work path actually needs adoption.",
            "The three starter suggestions should also stay brief and action-first",
            "Review the repo and suggest the next move.",
            "Fix the failing tests in this repo.",
            "Review the current branch and call out risks.",
            "show one calm repo-context line",
            "What do you want Jini to do?",
            "Type the task directly. Use `exit` to leave.",
            "Repo: sample-repo",
            "jini&gt; fix failing tests",
            "TASK    fix failing tests",
            "NEXT    make test",
            "jini&gt; doctor",
            "provider_id",
            "keep the <code>jini&gt;</code> prompt open",
            "keep the controls in the background instead of teaching them before the task starts",
            "<code>commands</code>, <code>doctor</code>, <code>help --admin</code>, and <code>exit</code> should still work inside the same live session",
            "should only print a <code>NEXT</code> line when it is genuinely steering you toward a concrete action",
            "Keep typing the next thing",
            "fix failing tests",
            "open the latest artifact",
            "what is blocked?",
            "switch to the other repo",
            "start from these notes",
            "You should not have to memorize product words like <code>Missing</code> or <code>Switch</code>",
            "plain asks like <code>plan this change</code> should still trigger the structured path",
            "The public catalog should stay narrow",
            "<code>status</code>, <code>continue</code>, <code>open</code>, and <code>doctor</code>",
            "<h3><code>jini continue</code></h3>",
            "<h3><code>jini open</code></h3>",
            "Commands like <code>try-example</code>, <code>get-started</code>, <code>show</code>, <code>expand</code>, <code>context</code>, <code>resume</code>, <code>metrics</code>, and <code>harnesses</code> should not sit in the public command story.",
        ):
            with self.subTest(marker=marker):
                self.assertIn(marker, text)

    def test_cli_guide_shows_help_request_tail_transcript(self) -> None:
        text = read(CLI_PATH)
        help_lines = help_tail_message_lines(
            "jini",
            "help",
            "CLI overview",
            HELP_TAIL_EXAMPLE_REQUEST.split(),
        )
        for marker in (
            f"$ jini help {HELP_TAIL_EXAMPLE_REQUEST}",
            *help_lines,
            "That corrective output is the expected terminal shape for a help request tail.",
        ):
            with self.subTest(marker=marker):
                self.assertIn(marker, text)

    def test_cli_guide_lists_help_variant_matrix(self) -> None:
        text = read(CLI_PATH)
        request_tokens = HELP_TAIL_EXAMPLE_REQUEST.split()
        for marker in (
            "The same redirect shape applies across the other help-entry variants too:",
            f"jini commands {HELP_TAIL_EXAMPLE_REQUEST}",
            help_tail_error_line("jini", "commands", "public command inventory", request_tokens),
            f"jini --help {HELP_TAIL_EXAMPLE_REQUEST}",
            help_tail_error_line("jini", "--help", "CLI overview", request_tokens),
            f"jini admin help {HELP_TAIL_EXAMPLE_REQUEST}",
            help_tail_error_line("jini", "admin help", "admin command inventory", request_tokens),
            f"jini provider help {HELP_TAIL_EXAMPLE_REQUEST}",
            help_tail_error_line("jini", "provider help", "admin command inventory", request_tokens),
            f"jini provider --help {HELP_TAIL_EXAMPLE_REQUEST}",
            help_tail_error_line("jini", "provider --help", "admin command inventory", request_tokens),
            "Only the first line changes.",
        ):
            with self.subTest(marker=marker):
                self.assertIn(marker, text)

    def test_cli_guide_keeps_support_commands_on_the_runtime_surface(self) -> None:
        text = read(CLI_PATH)
        for marker in (
            "jini status",
            "jini continue",
            "jini open",
            "jini doctor",
        ):
            with self.subTest(marker=marker):
                self.assertIn(marker, text)
        for marker in (
            "jini check",
            "jini setup",
            "jini metrics",
            "jini try-example research-prd",
            "jini get-started --harness codex",
            "jini harnesses",
        ):
            with self.subTest(marker=marker):
                self.assertNotIn(marker, text)
        self.assertIn(
            "If you already have work to adopt, use `jini status /path/to/work` once.",
            text,
        )

    def test_install_page_uses_same_public_vs_admin_boundary(self) -> None:
        text = read(INSTALL_PATH)
        for marker in (
            "jini commands",
            "small public command list",
            "jini admin help",
            "routes, bundles, or release plumbing",
            "That public list should stay deliberately short",
            "<code>status</code>, <code>continue</code>, <code>open</code>, and <code>doctor</code>",
            "Help surfaces and <code>jini commands</code> are catalogs, not request entrypoints.",
            "If you paste a work request after <code>help</code>, <code>--help</code>, or <code>commands</code>",
            "starting with <code>jini</code> for the start surface",
            "jini review this repo",
            "jini fix failing tests",
            "jini review this branch",
            "That repo-aware start surface should stay light: one calm repo-context line, direct task suggestions, one or two useful commands Jini found in the repo, and at most one quiet adoption hint for existing Jini work.",
            "Internal diagnostics like <code>repo-map</code> and setup surfaces like <code>doctor</code> should stay off the first screen.",
            "A Claude Code, Codex, or GitHub CLI user should see task suggestions first, then at most a small <code>Already have Jini work?</code> note with <code>status /path/to/work</code> when an existing Jini work path actually needs adoption.",
            "The three starter suggestions should stay brief and action-first",
            "Review the repo and suggest the next move.",
            "Fix the failing tests in this repo.",
            "Review the current branch and call out risks.",
            "show one calm repo-context line",
            "What do you want Jini to do?",
            "jini&gt;</code> prompt",
            "The task should stay primary, and the controls should stay in the background until you need them.",
            "<code>commands</code>, <code>doctor</code>, <code>help --admin</code>, and <code>exit</code> should still work as in-session escape hatches",
            "should only print a <code>NEXT</code> line when it is actually steering you to a concrete action",
            "keep typing plain follow-up asks like <code>fix failing tests</code>, <code>what is blocked?</code>, or <code>open the latest artifact</code> instead of learning Jini-specific action words",
        ):
            with self.subTest(marker=marker):
                self.assertIn(marker, text)

    def test_install_page_lists_help_variant_matrix(self) -> None:
        text = read(INSTALL_PATH)
        request_tokens = HELP_TAIL_EXAMPLE_REQUEST.split()
        for marker in (
            f"jini commands {HELP_TAIL_EXAMPLE_REQUEST}",
            help_tail_error_line("jini", "commands", "public command inventory", request_tokens),
            f"jini --help {HELP_TAIL_EXAMPLE_REQUEST}",
            help_tail_error_line("jini", "--help", "CLI overview", request_tokens),
            f"jini admin help {HELP_TAIL_EXAMPLE_REQUEST}",
            help_tail_error_line("jini", "admin help", "admin command inventory", request_tokens),
            f"jini provider help {HELP_TAIL_EXAMPLE_REQUEST}",
            help_tail_error_line("jini", "provider help", "admin command inventory", request_tokens),
            f"jini provider --help {HELP_TAIL_EXAMPLE_REQUEST}",
            help_tail_error_line("jini", "provider --help", "admin command inventory", request_tokens),
            "Only the first line changes. The redirect stays the same:",
        ):
            with self.subTest(marker=marker):
                self.assertIn(marker, text)

    def test_install_page_shows_provider_help_request_tail_transcript(self) -> None:
        text = read(INSTALL_PATH)
        help_lines = help_tail_message_lines(
            "jini",
            "provider help",
            "admin command inventory",
            HELP_TAIL_EXAMPLE_REQUEST.split(),
        )
        for marker in (
            f"$ jini provider help {HELP_TAIL_EXAMPLE_REQUEST}",
            *help_lines,
            "Use that provider-help example when someone drifts into the admin/provider tree during first run.",
        ):
            with self.subTest(marker=marker):
                self.assertIn(marker, text)

    def test_install_page_shows_dash_help_request_tail_transcript(self) -> None:
        text = read(INSTALL_PATH)
        help_lines = help_tail_message_lines(
            "jini",
            "--help",
            "CLI overview",
            HELP_TAIL_EXAMPLE_REQUEST.split(),
        )
        for marker in (
            f"$ jini --help {HELP_TAIL_EXAMPLE_REQUEST}",
            *help_lines,
            "Use that <code>--help</code> example when someone pastes work after the overview flag on first run.",
        ):
            with self.subTest(marker=marker):
                self.assertIn(marker, text)
        self.assertIn(
            "The contrast is simple: <code>--help</code> stays on the CLI-overview path, while <code>provider help</code> crosses into the admin-inventory path, even though both redirects send the user back to <code>jini</code>.",
            text,
        )

    def test_install_page_explains_installer_provenance_summary(self) -> None:
        text = read(INSTALL_PATH)
        for marker in (
            "- install source: release binary",
            "- install source: source runtime (explicit source)",
            "- install source: source runtime (release-unavailable)",
            "- install source: source fallback (release validation failed: unsupported-public-command-surface)",
        ):
            with self.subTest(marker=marker):
                self.assertIn(marker, text)

    def test_install_page_maps_provenance_lines_to_support_actions(self) -> None:
        text = read(INSTALL_PATH)
        for marker in (
            "Install source line",
            "What it usually means",
            "What to do next",
            "In short: <code>release binary</code> means no support receipt and no <code>next_step=</code>, while source-path follow-up output includes both.",
            "Matching published release asset was available, passed Jini's public command check, and was accepted without any source fallback.",
            "No action needed unless you expected a source install for local development.",
            "The absence of <code>- support receipt: ...</code> and <code>- next step: ...</code> is expected on this path, and a missing <code>next_step=</code> field is only meaningful on source-path installs.",
            "Treat <code>next_step=</code> as the actionable follow-up field for this source install.",
            "No matching release asset was available for this machine or release channel.",
            "If this machine should have had a published release, file a release issue and include the receipt. If support needs the install details, send <code>- support receipt: /path/to/install-receipt.txt (send version=, source_reason=, release_validation=, next_step=)</code>. Treat <code>next_step=</code> as the actionable follow-up field for this source install.",
            "Downloaded release binary did not support the current public command surface.",
            "Keep the source install, attach install-receipt.txt, and flag the stale release artifact. If support needs the install details, send <code>- support receipt: /path/to/install-receipt.txt (send version=, source_reason=, release_validation=, next_step=)</code>. Treat <code>next_step=</code> as the actionable follow-up field for this source install.",
        ):
            with self.subTest(marker=marker):
                self.assertIn(marker, text)

    def test_install_page_explains_receipt_next_step_field(self) -> None:
        text = read(INSTALL_PATH)
        for marker in (
            "install-receipt.txt",
            "source_reason=",
            "release_validation=",
            "next_step=",
            "In short: <code>release binary</code> means no support receipt and no <code>next_step=</code>, while source-path follow-up output includes both.",
            "On source-path installs that need follow-up, the terminal now prints this exact support handoff line: <code>- support receipt: /path/to/install-receipt.txt (send version=, source_reason=, release_validation=, next_step=)</code>.",
            "- support receipt: /path/to/install-receipt.txt (send version=, source_reason=, release_validation=, next_step=)",
        ):
            with self.subTest(marker=marker):
                self.assertIn(marker, text)
        self.assertEqual(
            text.count(
                "In short: <code>release binary</code> means no support receipt and no <code>next_step=</code>, while source-path follow-up output includes both."
            ),
            2,
        )
        self.assertEqual(
            text.count(
                "Treat <code>next_step=</code> as the actionable follow-up field for this source install."
            ),
            5,
        )
        self.assertEqual(
            text.count(
                "If support needs the install details, send <code>- support receipt: /path/to/install-receipt.txt (send version=, source_reason=, release_validation=, next_step=)</code>."
            ),
            3,
        )
        self.assertEqual(
            text.count(
                "On source-path installs that need follow-up, the terminal now prints this exact support handoff line: <code>- support receipt: /path/to/install-receipt.txt (send version=, source_reason=, release_validation=, next_step=)</code>."
            ),
            1,
        )
        self.assertEqual(
            text.count(
                "The absence of <code>- support receipt: ...</code> and <code>- next step: ...</code> is expected on this path, and a missing <code>next_step=</code> field is only meaningful on source-path installs."
            ),
            2,
        )
        self.assertEqual(
            text.count(
                "Unlike the source-path follow-up output below, <code>next_step=</code> appears only on source-path installs that need follow-up."
            ),
            1,
        )
        self.assertEqual(
            text.count(
                "- support receipt: /path/to/install-receipt.txt (send version=, source_reason=, release_validation=, next_step=)"
            ),
            4,
        )
        self.assertEqual(
            text.count(
                "- support receipt: /Users/you/.local/bin/install-receipt.txt (send version=, source_reason=, release_validation=, next_step=)"
            ),
            1,
        )

    def test_install_page_keeps_troubleshooting_block_shape(self) -> None:
        text = read(INSTALL_PATH)
        self.assertEqual(
            text.count(
                "In short: <code>release binary</code> means no support receipt and no <code>next_step=</code>, while source-path follow-up output includes both."
            ),
            2,
        )
        self.assertEqual(
            text.count(
                "On source-path installs that need follow-up, the terminal now prints this exact support handoff line: <code>- support receipt: /path/to/install-receipt.txt (send version=, source_reason=, release_validation=, next_step=)</code>."
            ),
            1,
        )
        self.assertEqual(
            text.count(
                "When support asks for install details on a source-path install, send the support receipt path plus only these receipt keys. Treat <code>next_step=</code> as the actionable follow-up field for this source install."
            ),
            1,
        )
        self.assertEqual(text.count("Install source line"), 1)
        self.assertEqual(text.count("What it usually means"), 1)
        self.assertEqual(text.count("What to do next"), 1)
        self.assertEqual(
            text.count(
                '<th scope="row"><code>- install source: release binary</code></th>'
            ),
            1,
        )
        self.assertEqual(
            text.count(
                '<th scope="row"><code>- install source: source runtime (explicit source)</code></th>'
            ),
            1,
        )
        self.assertEqual(
            text.count(
                '<th scope="row"><code>- install source: source runtime (release-unavailable)</code></th>'
            ),
            1,
        )
        self.assertEqual(
            text.count(
                '<th scope="row"><code>- install source: source fallback (release validation failed: unsupported-public-command-surface)</code></th>'
            ),
            1,
        )
        release_row = text.index(
            '<th scope="row"><code>- install source: release binary</code></th>'
        )
        explicit_source_row = text.index(
            '<th scope="row"><code>- install source: source runtime (explicit source)</code></th>'
        )
        release_unavailable_row = text.index(
            '<th scope="row"><code>- install source: source runtime (release-unavailable)</code></th>'
        )
        stale_release_row = text.index(
            '<th scope="row"><code>- install source: source fallback (release validation failed: unsupported-public-command-surface)</code></th>'
        )
        self.assertLess(release_row, explicit_source_row)
        self.assertLess(explicit_source_row, release_unavailable_row)
        self.assertLess(release_unavailable_row, stale_release_row)
        self.assertEqual(text.count("Example release-binary success output:"), 1)
        self.assertEqual(text.count("Example source-path follow-up output:"), 1)

    def test_install_page_shows_source_path_terminal_example(self) -> None:
        text = read(INSTALL_PATH)
        for marker in (
            "Example source-path follow-up output:",
            "- install source: source runtime (release-unavailable)",
            "- support receipt: /Users/you/.local/bin/install-receipt.txt (send version=, source_reason=, release_validation=, next_step=)",
            "- next step: If this machine should have had a published release, file a release issue and include the receipt.",
            "Treat <code>next_step=</code> as the actionable follow-up field for this source install.",
        ):
            with self.subTest(marker=marker):
                self.assertIn(marker, text)

    def test_install_page_shows_release_binary_terminal_example(self) -> None:
        text = read(INSTALL_PATH)
        for marker in (
            "Example release-binary success output:",
            "- install source: release binary",
            "jini",
            "The healthy release-binary path stops there.",
            "The absence of <code>- support receipt: ...</code> and <code>- next step: ...</code> is expected on this path, and a missing <code>next_step=</code> field is only meaningful on source-path installs.",
            "Unlike the source-path follow-up output below, <code>next_step=</code> appears only on source-path installs that need follow-up.",
        ):
            with self.subTest(marker=marker):
                self.assertIn(marker, text)

    def test_install_page_lists_minimal_support_receipt_keys(self) -> None:
        text = read(INSTALL_PATH)
        for marker in (
            "When support asks for install details on a source-path install, send the support receipt path plus only these receipt keys. Treat <code>next_step=</code> as the actionable follow-up field for this source install.",
            "install-receipt.txt</code> path from the printed <code>- support receipt: ...</code> line",
            "version=",
            "source_reason=",
            "release_validation=",
            "next_step=",
            "If the install output shows only <code>- install source: release binary</code> followed by <code>jini</code>, support can ignore this checklist.",
            "If <code>next_step=</code> is missing on a source-path install, that source path likely completed without extra follow-up. The healthy <code>release binary</code> path does not use this source-path handoff at all.",
        ):
            with self.subTest(marker=marker):
                self.assertIn(marker, text)

    def test_quickstart_uses_same_public_vs_admin_boundary(self) -> None:
        text = read(SIMPLE_PATH)
        for marker in (
            "jini commands",
            "small public command list",
            "jini admin help",
            "routes, bundles, or release plumbing",
            "<code>status</code>, <code>continue</code>, <code>open</code>, and <code>doctor</code>",
        ):
            with self.subTest(marker=marker):
                self.assertIn(marker, text)


if __name__ == "__main__":
    unittest.main()
