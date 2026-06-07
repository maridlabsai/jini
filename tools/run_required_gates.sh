#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
GO_BIN="${GO_BIN:-/Users/sharad.sharma/Developer/.local-go/bin/go}"
GO_CACHE_DIR="${JINI_GOCACHE:-/private/tmp/jini-go-cache}"
GO_MOD_CACHE_DIR="${JINI_GOMODCACHE:-/private/tmp/jini-go-mod}"
SECURITY_CONFIGURATION_GATE="${ROOT_DIR}/tools/security_configuration_gate.sh"
PRODUCT_PRD_DRIFT_GATE="${ROOT_DIR}/tools/product_prd_drift_gate.sh"

usage() {
  cat <<'EOF'
Usage: bash tools/run_required_gates.sh <commit|push|release>
EOF
}

require_go_bin() {
  if [[ ! -x "${GO_BIN}" ]]; then
    printf 'Required Go binary not found or not executable: %s\n' "${GO_BIN}" >&2
    exit 1
  fi
}

run_go_test() {
  local package_target="$1"
  (
    cd "${ROOT_DIR}"
    GOCACHE="${GO_CACHE_DIR}" \
    GOMODCACHE="${GO_MOD_CACHE_DIR}" \
      "${GO_BIN}" test "${package_target}"
  )
}

run_security_configuration_gate() {
  bash "${SECURITY_CONFIGURATION_GATE}"
}

run_product_prd_drift_gate() {
  bash "${PRODUCT_PRD_DRIFT_GATE}"
}

run_scorecard_gate() {
  (
    cd "${ROOT_DIR}"
    GOCACHE="${GO_CACHE_DIR}" \
    GOMODCACHE="${GO_MOD_CACHE_DIR}" \
      "${GO_BIN}" run ./cmd/jini scorecard-gate --format json
  )
}

run_commit_gate() {
  (
    cd "${ROOT_DIR}"
    git diff --check
    git diff --cached --check
  )
  run_security_configuration_gate
  run_product_prd_drift_gate
  run_scorecard_gate
  run_go_test "./..."
}

run_push_gate() {
  run_commit_gate
}

run_release_gate() {
  run_push_gate
  (
    cd "${ROOT_DIR}"
    GOCACHE="${GO_CACHE_DIR}" \
    GOMODCACHE="${GO_MOD_CACHE_DIR}" \
      "${GO_BIN}" run ./cmd/jini publish-readiness --format json
  )
}

main() {
  if [[ $# -ne 1 ]]; then
    usage
    exit 1
  fi

  require_go_bin

  case "$1" in
    commit)
      run_commit_gate
      ;;
    push)
      run_push_gate
      ;;
    release)
      run_release_gate
      ;;
    *)
      usage
      exit 1
      ;;
  esac
}

main "$@"
