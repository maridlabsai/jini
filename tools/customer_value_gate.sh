#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

usage() {
  cat <<'EOF'
Usage: bash tools/customer_value_gate.sh

Blocks product drift where Jini looks polished but no longer proves real
customer value: token frugality, throttle resilience, configured-tool routing,
direct action, safety, and session continuity.
EOF
}

require_fragment() {
  local path="$1"
  local fragment="$2"
  if ! grep -Fq -- "${fragment}" "${ROOT_DIR}/${path}"; then
    printf 'Customer value gate failed.\n' >&2
    printf 'Missing required fragment in %s:\n%s\n' "${path}" "${fragment}" >&2
    return 1
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

  require_fragment "specs/product-settling-decisions.md" "## Customer Value Bar"
  require_fragment "specs/product-settling-decisions.md" "Jini is viable only when it makes the user's existing AI tools easier to use"
  require_fragment "specs/product-settling-decisions.md" "Every product, routing, CLI UX, or docs cut must map to at least one customer"
  require_fragment "specs/product-settling-decisions.md" "less token spend through compact context reuse instead of transcript replay"
  require_fragment "specs/product-settling-decisions.md" "fewer throttle stalls through route choice, fallback, or explicit setup"
  require_fragment "specs/product-settling-decisions.md" "lower switching cost across Claude Code, Codex, Gemini CLI, Aider, OpenCode,"
  require_fragment "specs/product-settling-decisions.md" "Anti-amateur constraints:"
  require_fragment "specs/product-settling-decisions.md" "Do not claim support for a framework, model, or CLI unless Jini can detect it,"
  require_fragment "specs/product-settling-decisions.md" "Do not answer with generic templates, dashboards, task snapshots, or workflow"

  require_fragment "specs/number-one-platform-prd.md" "Customer value bar: every shipped cut must improve token frugality, throttle"
  require_fragment "specs/number-one-platform-prd.md" "Preserve customer-value viability: reduce token waste, throttle friction,"
  require_fragment "specs/number-one-platform-prd.md" "bash tools/customer_value_gate.sh"
  require_fragment "specs/number-one-platform-prd.md" "No commit ships unless the customer-value gate can still prove the solution is"

  require_fragment "specs/golden-competitive-benchmark.yaml" "customer-value-viability-fixture"
  require_fragment "specs/golden-competitive-benchmark.yaml" "product-viability-customer-value"
  require_fragment "specs/golden-competitive-benchmark.yaml" "A green product cut must map to customer value, not just implementation activity"

  require_fragment "tools/run_required_gates.sh" "CUSTOMER_VALUE_GATE="
  require_fragment "tools/run_required_gates.sh" "run_customer_value_gate"
}

main "$@"
