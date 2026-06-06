#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

usage() {
  cat <<'EOF'
Usage: bash tools/security_configuration_gate.sh

Verifies that the repository keeps free security scanners wired into CI:
CodeQL SAST, govulncheck, OSV-Scanner, and Dependabot.
EOF
}

require_file() {
  local rel_path="$1"
  if [[ ! -f "${ROOT_DIR}/${rel_path}" ]]; then
    printf 'Security scanner gate missing required file: %s\n' "${rel_path}" >&2
    exit 1
  fi
}

require_fragment() {
  local rel_path="$1"
  local fragment="$2"
  if ! grep -Fq -- "${fragment}" "${ROOT_DIR}/${rel_path}"; then
    printf 'Security scanner gate missing %q in %s\n' "${fragment}" "${rel_path}" >&2
    exit 1
  fi
}

main() {
  if [[ $# -gt 0 ]]; then
    case "$1" in
      -h|--help)
        usage
        return 0
        ;;
      *)
        usage >&2
        return 1
        ;;
    esac
  fi

  local workflow=".github/workflows/security.yml"
  local dependabot=".github/dependabot.yml"

  require_file "${workflow}"
  require_file "${dependabot}"

  require_fragment "${workflow}" "github/codeql-action/init@v4"
  require_fragment "${workflow}" "github/codeql-action/analyze@v4"
  require_fragment "${workflow}" "golang/govulncheck-action@v1"
  require_fragment "${workflow}" "google/osv-scanner-action/.github/workflows/osv-scanner-reusable-pr.yml@v2.3.8"
  require_fragment "${workflow}" "google/osv-scanner-action/.github/workflows/osv-scanner-reusable.yml@v2.3.8"
  require_fragment "${workflow}" "security-events: write"
  require_fragment "${workflow}" "go-version-file: go.mod"

  require_fragment "${dependabot}" 'package-ecosystem: "gomod"'
  require_fragment "${dependabot}" 'package-ecosystem: "github-actions"'
  require_fragment "${dependabot}" 'directory: "/"'
  require_fragment "${dependabot}" 'interval: "weekly"'
}

main "$@"
