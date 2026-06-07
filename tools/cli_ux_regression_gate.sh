#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
GO_BIN="${GO_BIN:-/Users/sharad.sharma/Developer/.local-go/bin/go}"
GO_CACHE_DIR="${JINI_GOCACHE:-/private/tmp/jini-go-cache}"
GO_MOD_CACHE_DIR="${JINI_GOMODCACHE:-/private/tmp/jini-go-mod}"
VERBOSE=0

FOCUSED_TEST_PATTERN='TestInteractiveLocalTextEditAppendsQuotedLineInsteadOfDrafting|TestCurrentWorkLocalTextEditExecutesWithoutStartPrompt|TestCurrentWorkSimpleFactualQuestionAnswersDirectly|TestCurrentWorkUnknownStandaloneQuestionStaysCompact|TestLauncherStartsAsCompactShellWhenCurrentWorkExists|TestCurrentWorkInteractiveLauncherIsCompactByDefault|TestLauncherShowsOtherActiveWorkWhenMultipleProjectsExist|TestInteractiveLauncherCanResumeNamedActiveProject|TestPublicDocsUseCurrentFirstRunFlow|TestP1SimplicityPriorityCoversCommandsSkillsAndAgents'

usage() {
  cat <<'EOF'
Usage: bash tools/cli_ux_regression_gate.sh [--verbose]

Runs the direct CLI edit and simple-question UX regression gate.

This pins the incident scenarios where simple local edits or simple factual
questions must not regress into draft/status frames, Start/Keep choices, or
verbose work-state output.
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
    printf 'Running CLI UX regression tests: %s\n' "${FOCUSED_TEST_PATTERN}"
  fi

  (
    cd "${ROOT_DIR}"
    GOCACHE="${GO_CACHE_DIR}" \
    GOMODCACHE="${GO_MOD_CACHE_DIR}" \
      "${GO_BIN}" test ./internal/app -run "${FOCUSED_TEST_PATTERN}" -count=1
  )
}

main "$@"
