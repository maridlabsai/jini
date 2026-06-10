#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

usage() {
  cat <<'EOF'
Usage: bash tools/macos_bundle_hygiene_gate.sh

Verifies that generated macOS app sidecar binaries stay out of git.
EOF
}

require_fragment() {
  local rel_path="$1"
  local fragment="$2"
  if ! grep -Fq -- "${fragment}" "${ROOT_DIR}/${rel_path}"; then
    printf 'macOS bundle hygiene gate missing %q in %s\n' "${fragment}" "${rel_path}" >&2
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

  require_fragment ".gitignore" "apps/macos/src-tauri/binaries/jini-sidecar-*"

  local tracked_sidecars
  tracked_sidecars="$(
    cd "${ROOT_DIR}"
    git ls-files 'apps/macos/src-tauri/binaries/jini-sidecar-*'
  )"
  if [[ -n "${tracked_sidecars}" ]]; then
    printf 'Generated macOS sidecar binaries must not be tracked:\n%s\n' "${tracked_sidecars}" >&2
    printf 'Keep apps/macos/src-tauri/binaries/.gitkeep tracked and build sidecars during packaging.\n' >&2
    exit 1
  fi
}

main "$@"
