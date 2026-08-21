#!/usr/bin/env bash
#
# functional_smoke.sh — confirm Jini is functional across every core scenario.
#
# Two tiers:
#   core (default) — the hermetic functional harness (TestFunctionalHarness):
#     offline via local-preview, no network, deterministic. Safe to run any
#     time, in CI, or before a release. This is the always-on health check.
#   live (--live)  — real installed-CLI route validation (jini route smoke +
#     dogfood) for a claimed route. Makes real hand-off calls; needs an
#     installed, authenticated CLI. Produces the .jini/ evidence the push gate
#     requires. Route defaults to claude-code; override with JINI_SMOKE_ROUTE.
#
# Usage:
#   tools/functional_smoke.sh            # core harness only
#   tools/functional_smoke.sh --live     # core harness + live route smoke

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
GO_BIN="${GO_BIN:-$(command -v go || true)}"
GO_CACHE_DIR="${JINI_GOCACHE:-/private/tmp/jini-go-cache}"
GO_MOD_CACHE_DIR="${JINI_GOMODCACHE:-/private/tmp/jini-go-mod}"
SMOKE_ROUTE="${JINI_SMOKE_ROUTE:-claude-code}"
LIVE=0
[ "${1:-}" = "--live" ] && LIVE=1

cd "$ROOT_DIR"
export GOCACHE="$GO_CACHE_DIR" GOMODCACHE="$GO_MOD_CACHE_DIR"

echo "== Jini functional smoke =="

echo "-- core: hermetic functional harness --"
if "$GO_BIN" test ./internal/app/ -run 'TestFunctionalHarness' -count=1 -v >/tmp/jini-functional-core.log 2>&1; then
  passed="$(grep -c -- '--- PASS: TestFunctionalHarness/' /tmp/jini-functional-core.log || true)"
  echo "   core: PASS (${passed} scenarios)"
else
  echo "   core: FAIL"
  grep -E -- '--- FAIL:|missing|exit=' /tmp/jini-functional-core.log | head -20
  exit 1
fi

if [ "$LIVE" -eq 0 ]; then
  echo "   (the same scenarios ship as the runtime command: jini check functional)"
  echo "== functional smoke: core PASS (run with --live for real route validation) =="
  exit 0
fi

echo "-- live: route smoke for '$SMOKE_ROUTE' --"
JINI_BIN="${JINI_BIN:-$(command -v jini || true)}"
if [ -z "$JINI_BIN" ]; then
  echo "   live: SKIP — no 'jini' on PATH (build: go build -o /usr/local/bin/jini ./cmd/jini)"
  exit 1
fi
if ! "$JINI_BIN" route smoke "$SMOKE_ROUTE"; then
  echo "   live: route smoke FAILED for $SMOKE_ROUTE"
  exit 1
fi
JINI_CLI_RELEASE_ROUTES="$SMOKE_ROUTE" "$JINI_BIN" route dogfood run --claimed
echo "== functional smoke: core + live PASS =="
