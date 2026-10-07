#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
GO_BIN="${GO_BIN:-$(command -v go || true)}"

# Resolve Go cache directories portably (never the macOS-only /private/tmp,
# which fails on Linux CI). Precedence: explicit JINI_* override, else a standard
# GOCACHE/GOMODCACHE in the environment, else the platform default from `go env`.
resolve_go_cache_dir() { # $1=JINI override, $2=standard value, $3=go env key
  if [[ -n "$1" ]]; then
    printf '%s' "$1"
  elif [[ -n "$2" ]]; then
    printf '%s' "$2"
  elif [[ -x "${GO_BIN}" ]]; then
    "${GO_BIN}" env "$3" 2>/dev/null || true
  fi
}
GO_CACHE_DIR="$(resolve_go_cache_dir "${JINI_GOCACHE:-}" "${GOCACHE:-}" GOCACHE)"
GO_MOD_CACHE_DIR="$(resolve_go_cache_dir "${JINI_GOMODCACHE:-}" "${GOMODCACHE:-}" GOMODCACHE)"
if [[ -z "${GO_CACHE_DIR}" || -z "${GO_MOD_CACHE_DIR}" ]]; then
  printf 'Could not resolve Go cache directories (set JINI_GOCACHE/JINI_GOMODCACHE or GOCACHE/GOMODCACHE).\n' >&2
  exit 1
fi
mkdir -p "${GO_CACHE_DIR}" "${GO_MOD_CACHE_DIR}"
VERBOSE=0

FOCUSED_TEST_PATTERN='TestInteractiveLocalTextEditAppendsQuotedLineInsteadOfDrafting|TestCurrentWorkLocalTextEditExecutesWithoutStartPrompt|TestCurrentWorkSimpleFactualQuestionAnswersDirectly|TestCurrentWorkCapitalQuestionAcceptsNaturalPhrasing|TestDirectArgsSimpleFactualQuestionAnswersDirectly|TestInteractiveSimpleFactualQuestionAnswersDirectlyWithoutCurrentWork|TestInteractiveTypoCapitalQuestionAnswersDirectlyWithoutArtifactShell|TestCurrentWorkTypoCapitalQuestionAnswersDirectly|TestInteractiveMalformedCapitalQuestionCorrectsWithoutTravelFlow|TestInteractiveBareEntityAsksForIntentWithoutCreatingWork|TestInteractiveExplicitTripChoiceCanUseBareDestination|TestCurrentWorkMalformedCapitalQuestionCorrectsDirectly|TestCurrentWorkUnknownStandaloneQuestionStaysCompact|TestStandaloneQuestionUsesConfiguredCLIRouteWithoutCreatingWork|TestStandaloneQuestionTimeoutStaysCompact|TestStandaloneQuestionFailedCLIRouteStaysCompact|TestTaskShapedQuestionDoesNotUseStandaloneQuestionFallback|TestCurrentWorkQuestionClassifierDoesNotHijackStandaloneNextQuestion|TestShellOutputUsesPreciseProfessionalLanguage|TestShellOutputRejectsStaleWorkflowVocabulary|TestRouteCommandCanSetLocalSLMAliasToEligibleProfile|TestCommercialOnlyCommandFailsClosedInFreeCLI|TestCommercialOnlyInteractiveInputDoesNotCreateWork|TestInteractiveLauncherHandlesUnsureInputWithUsefulPass|TestLauncherStartsAsCompactShellWhenCurrentWorkExists|TestCurrentWorkInteractiveLauncherIsCompactByDefault|TestLauncherShowsOtherActiveWorkWhenMultipleProjectsExist|TestInteractiveLauncherCanResumeNamedActiveProject|TestPublicDocsUseCurrentFirstRunFlow|TestP1SimplicityPriorityCoversCommandsSkillsAndAgents'

usage() {
  cat <<'EOF'
Usage: bash tools/cli_ux_regression_gate.sh [--verbose]

Runs the direct CLI edit, simple-question, and intent-first UX regression gate.

This pins the incident scenarios where simple local edits, simple factual
questions, malformed questions, bare entities, generic fallback tasks, and
shell wording must not regress into template routing, draft/status frames,
Start/Keep choices, stale vocabulary, vague receipts, or verbose work-state output.
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
