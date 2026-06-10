#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
GO_BIN="${GO_BIN:-/Users/sharad.sharma/Developer/.local-go/bin/go}"
GO_CACHE_DIR="${JINI_GOCACHE:-/private/tmp/jini-go-cache}"
GO_MOD_CACHE_DIR="${JINI_GOMODCACHE:-/private/tmp/jini-go-mod}"
VERBOSE=0

FOCUSED_TEST_PATTERN='TestInteractiveLocalTextEditAppendsQuotedLineInsteadOfDrafting|TestDirectRepoReviewPrintsAndSavesModelFreeSnapshot|TestGenerateWithConfiguredProviderHandsOffToConfiguredCLI|TestGenerateWithConfiguredProviderHandsOffToCodexCLI|TestGenerateWithConfiguredProviderPreservesQuotedCustomCLIArgs|TestRunInteractiveKeepsCurrentWorkAfterFailedCLIHandoff|TestProviderDoctorFailsClosedForReservedCLIHandoffToolMode|TestDetectRouteAutoCanChooseReadyCLIHandoffCandidate|TestDetectRouteAutoAvoidsCLIHandoffWhenConnectivityIsOffline|TestDetectRouteAutoAvoidsPinnedRemoteProviderWhenConnectivityIsOffline|TestDetectRouteForLocalSLMAliasChoosesEligibleProfile|TestDetectRouteAutoPinnedLocalSLMWithoutRuntimeFallsBackToLocalPreview|TestDetectRouteAutoDoesNotSelectUnavailableLocalProfileInOfflineMode|TestStandaloneQuestionUsesConfiguredCLIRouteWithoutCreatingWork|TestStandaloneQuestionTimeoutStaysCompact|TestStandaloneQuestionFailedCLIRouteStaysCompact|TestCurrentWorkSimpleFactualQuestionAnswersDirectly|TestInteractiveTypoCapitalQuestionAnswersDirectlyWithoutArtifactShell|TestCurrentWorkTypoCapitalQuestionAnswersDirectly|TestClaudeCodexPromptBankCoversDiverseFirstMinuteUseCases'

usage() {
  cat <<'EOF'
Usage: bash tools/claude_codex_usecase_gate.sh [--verbose]

Runs Claude/Codex commit-gate use cases.

This keeps every commit honest against concrete Claude Code and Codex-style
expectations: local file edits, repo review, strict CLI handoff, custom Claude
args, Codex handoff, failed handoff recovery, compact questions, and the
100-prompt Aryan-derived first-minute prompt bank.
EOF
}

main() {
  while [[ $# -gt 0 ]]; do
    case "$1" in
      -h|--help)
        usage
        return 0
        ;;
      -v|--verbose)
        VERBOSE=1
        shift
        ;;
      *)
        usage >&2
        return 1
        ;;
    esac
  done

  if [[ ! -x "${GO_BIN}" ]]; then
    printf 'Required Go binary not found or not executable: %s\n' "${GO_BIN}" >&2
    return 1
  fi

  if [[ "${VERBOSE}" -eq 1 ]]; then
    printf 'Running Claude/Codex commit-gate use cases: %s\n' "${FOCUSED_TEST_PATTERN}"
  fi

  (
    cd "${ROOT_DIR}"
    GOCACHE="${GO_CACHE_DIR}" \
    GOMODCACHE="${GO_MOD_CACHE_DIR}" \
      "${GO_BIN}" test ./internal/app -run "${FOCUSED_TEST_PATTERN}" -count=1
  )
}

main "$@"
