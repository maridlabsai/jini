#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SETTLING_DOC="specs/product-settling-decisions.md"

usage() {
  cat <<'EOF'
Usage: bash tools/product_prd_drift_gate.sh

Blocks protected product and PRD surfaces from drifting unless the same change
updates specs/product-settling-decisions.md.
EOF
}

changed_files() {
  (
    cd "${ROOT_DIR}"
    git diff --name-only
    git diff --cached --name-only
    git ls-files --others --exclude-standard
  ) | sort -u
}

is_protected_product_surface() {
  case "$1" in
    README.md | \
    specs/number-one-platform-prd.md | \
    specs/macos-app-prd.md | \
    specs/macos-app-ux-design.md | \
    specs/macos-app-hld.md | \
    specs/macos-app-lld.md | \
    specs/product-settling-decisions.md | \
    specs/product-streamline-redline.md | \
    specs/agentic-development-operating-model.md | \
    specs/number-one-platform-hld.md | \
    specs/number-one-platform-lld.md | \
    specs/product-consensus-prd-and-plan.md | \
    specs/full-product-prd.md | \
    specs/full-product-prd-execution-plan.md | \
    specs/launcher-intake-design.md | \
    specs/lean-platform-doctrine.md | \
    specs/client-surfaces-and-free-tier.md | \
    specs/platform-offline-strategy.md | \
    specs/execution-routing-policy.md | \
    specs/cross-surface-session-platform-prd.md | \
    specs/skills-and-delegation-slice.md | \
    specs/competitive-release-plan.md | \
    specs/number-one-development-plan.md | \
    specs/number-one-product-research.md | \
    specs/travel-curated-experience-framework.md)
      return 0
      ;;
    *)
      return 1
      ;;
  esac
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

  local protected=()
  local settling_changed=0
  local path
  while IFS= read -r path; do
    [[ -n "${path}" ]] || continue
    if [[ "${path}" == "${SETTLING_DOC}" ]]; then
      settling_changed=1
    fi
    if is_protected_product_surface "${path}"; then
      protected+=("${path}")
    fi
  done < <(changed_files)

  if [[ ${#protected[@]} -eq 0 ]]; then
    return 0
  fi
  if [[ "${settling_changed}" -eq 1 ]]; then
    return 0
  fi

  printf 'Product PRD drift gate failed.\n' >&2
  printf 'Protected product surfaces changed without updating %s.\n' "${SETTLING_DOC}" >&2
  printf '\nChanged protected files:\n' >&2
  printf -- '- %s\n' "${protected[@]}" >&2
  printf '\nEither update %s in the same change or move the edit out of the protected product surface.\n' "${SETTLING_DOC}" >&2
  exit 1
}

main "$@"
