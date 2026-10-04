#!/usr/bin/env bash
# JINI-R0 claim-language gate (scoped — deliberately NOT a repo-wide grep).
#
# It FAILS CLOSED and inspects only current, user-visible surfaces:
#   1. public documentation: README.md + docs/**/*.md (recursive), and
#   2. live runtime command OUTPUT: jini share / streak / mcp list / savings
#      and the startup savings counter, exercised against a deterministic
#      TEMPORARY Jini home (never the developer's real home or ledger).
#
# It permits only narrowly qualified wording. A build failure, a command that
# does not exit as expected, missing expected output, or a scan error are all
# treated as a GATE FAILURE.
#
# Excluded from scanning (never added to the surface list): specs/** (incl.
# specs/archive/**), *_test.go and fixtures, docs/Jini_Trust_First_Recovery_Plan.md,
# docs/_site/** and docs/archive/**, and documents explicitly marked historical.
#
# Test overrides (CLAIM_GATE_MANIFEST / CLAIM_GATE_SKIP_RUNTIME /
# CLAIM_GATE_DOCROOT / CLAIM_GATE_BIN) are honored ONLY when CLAIM_GATE_TESTMODE=1
# is set by a test invoking this script directly. The production runner
# (tools/run_required_gates.sh) unsets all of them, so exported overrides can
# never neuter the committed gate.
set -uo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${ROOT}" || { echo "claim-language gate: cannot cd to repo root" >&2; exit 1; }

TESTMODE="${CLAIM_GATE_TESTMODE:-}"

fail=0
violation() { printf '::claim-gate:: %s\n' "$1" >&2; fail=1; }

# --- pattern vocabulary -------------------------------------------------------
# HARD: banned in any shipped surface; no qualifier (incl. "imputed") rescues
# these. "saved( me/you)? $" catches factual realized-savings wording tied to a
# dollar figure ("Saved ≈ $5", "saved you $5") while leaving ordinary
# non-financial uses like "saved a file" untouched (no adjacent $ figure).
HARD='un-?metered|walls avoided|[0-9]+ (throttle[- ])?walls|throttles dodged|total saved|saved( (me|you))? *≈? *\$|\$[0-9][0-9.,]* saved|\[ready\]'
# Any dollar figure (used only on savings surfaces, where every $ is a savings figure).
MONEY='\$[0-9]'
# Same-line qualifier that legitimizes a dollar FIGURE.
MONEY_OK='imputed|estimated'
# "free to run" is allowed only as a future/gated aspiration, and the qualifier
# must sit in the same sentence (no '.' between) so an unrelated far-away
# qualifier cannot suppress the violation.
FREE_OK='free to run[^.]*(future|v1\.5|aspiration|gated|target|goal)|(future|v1\.5|aspiration|gated|target|goal)[^.]*free to run'

matches() { printf '%s' "$2" | grep -qiE "$1"; }

# scan_doc: HARD phrases + the free-to-run qualifier rule. No bare-$ rule, so
# legitimate prices in docs are not flagged.
scan_doc() { # $1 label, lines on stdin
  local label="$1" line
  while IFS= read -r line; do
    [ -z "${line}" ] && continue
    if matches "${HARD}" "${line}"; then
      violation "${label}: unqualified claim: ${line}"
    fi
    if matches 'free to run' "${line}" && ! matches "${FREE_OK}" "${line}"; then
      violation "${label}: 'free to run' not qualified as a future/gated aspiration: ${line}"
    fi
  done
}

# scan_money: HARD phrases + every dollar figure must carry imputed/estimated on
# the same line. Used on savings surfaces only.
scan_money() { # $1 label, lines on stdin
  local label="$1" line
  while IFS= read -r line; do
    [ -z "${line}" ] && continue
    if matches "${HARD}" "${line}"; then
      violation "${label}: unqualified claim: ${line}"
    fi
    if matches "${MONEY}" "${line}" && ! matches "${MONEY_OK}" "${line}"; then
      violation "${label}: dollar figure not labeled imputed/estimated on the same line: ${line}"
    fi
  done
}

is_historical_doc() { # $1 path — skip files explicitly marked historical at the top
  head -n 8 "$1" 2>/dev/null | grep -qiE 'HISTORICAL|superseded by jini-r0|unverified historical snapshot'
}

scan_doc_tree() { # $1 docs dir, $2 README path
  local docs="$1" readme="$2" f
  [ -f "${readme}" ] && scan_doc "${readme}" < "${readme}"
  [ -d "${docs}" ] || return 0
  while IFS= read -r f; do
    case "${f}" in
      *Jini_Trust_First_Recovery_Plan.md) continue ;;
      */_site/*|*/archive/*) continue ;;
      *historical*) continue ;;
    esac
    is_historical_doc "${f}" && continue
    scan_doc "${f}" < "${f}"
  done < <(find "${docs}" -type f -name '*.md' 2>/dev/null | sort)
}

# --- test-mode manifest override (honored only under CLAIM_GATE_TESTMODE=1) ----
if [ "${TESTMODE}" = "1" ] && [ -n "${CLAIM_GATE_MANIFEST:-}" ]; then
  for f in ${CLAIM_GATE_MANIFEST}; do
    [ -f "${f}" ] || continue
    scan_doc "${f}" < "${f}"
  done
  if [ "${fail}" -eq 0 ]; then echo "claim-language gate: clean (test manifest)"; exit 0; fi
  echo "claim-language gate: FAILED (test manifest)" >&2; exit 1
fi

# --- documentation surfaces ---------------------------------------------------
if [ "${TESTMODE}" = "1" ] && [ -n "${CLAIM_GATE_DOCROOT:-}" ]; then
  scan_doc_tree "${CLAIM_GATE_DOCROOT}" "${CLAIM_GATE_DOCROOT}/README.md"
else
  scan_doc_tree "docs" "README.md"
fi

# --- runtime surfaces ---------------------------------------------------------
if [ "${TESTMODE}" = "1" ] && [ -n "${CLAIM_GATE_SKIP_RUNTIME:-}" ]; then
  : # test opt-out of the runtime scan
else
  workdir="$(mktemp -d)" || { echo "claim-language gate: mktemp failed" >&2; exit 1; }
  trap 'rm -rf "${workdir}"' EXIT
  bin="${workdir}/jini"
  home="${workdir}/home"
  mkdir -p "${home}/.jini"

  # Deterministic non-zero ledger so savings + startup counter render figures.
  cat > "${home}/.jini/savings-ledger.json" <<'LEDGER'
{"schema_version":"0.1.0","context_type":"JiniSavingsLedger","entries":[{"schema_version":"0.1.0","context_type":"JiniSavingsEntry","route_class":"subscription","usd_saved":3.25,"imputed":true,"throttle_dodged":true}],"folded":{"tasks":0,"usd_saved":0,"dodges":0},"totals":{"tasks":1,"usd_saved":3.25,"dodges":1}}
LEDGER
  # One found command (sh) and one missing command for mcp list.
  cat > "${home}/.jini/mcp.json" <<'MCP'
{"mcpServers": {"found": {"command": "sh", "args": ["-c","true"], "transport": "stdio"}, "missing": {"command": "jini-not-a-real-command-xyzzy", "args": [], "transport": "stdio"}}}
MCP

  build_ok=1
  if [ "${TESTMODE}" = "1" ] && [ -n "${CLAIM_GATE_BIN:-}" ]; then
    bin="${CLAIM_GATE_BIN}" # test-injected fake binary; skip the real build
  elif ! command -v go >/dev/null 2>&1; then
    violation "runtime: go toolchain unavailable — cannot verify live command output"
    build_ok=0
  elif ! go build -o "${bin}" ./cmd/jini >/dev/null 2>&1; then
    violation "runtime: 'go build ./cmd/jini' failed — cannot verify live output"
    build_ok=0
  fi

  if [ "${build_ok}" -eq 1 ]; then
    run() { JINI_HOME="${home}" "${bin}" "$@" 2>&1; }

    # share / streak: must fail closed (exit 1), emit no figures or HARD claims.
    out="$(run share)"; code=$?
    [ "${code}" -eq 1 ] || violation "runtime jini share: expected exit 1, got ${code}"
    matches "${MONEY}" "${out}" && violation "runtime jini share: must not emit a dollar figure: ${out}"
    scan_doc "runtime jini share" <<< "${out}"

    out="$(run streak)"; code=$?
    [ "${code}" -eq 1 ] || violation "runtime jini streak: expected exit 1, got ${code}"
    matches "${MONEY}" "${out}" && violation "runtime jini streak: must not emit a dollar figure: ${out}"
    scan_doc "runtime jini streak" <<< "${out}"

    # mcp list: listing-only wording for both found and missing; never '[ready]'.
    out="$(run mcp list)"; code=$?
    [ "${code}" -eq 0 ] || violation "runtime jini mcp list: expected exit 0, got ${code}"
    matches 'command found — configuration listing only; invocation not supported' "${out}" \
      || violation "runtime jini mcp list: missing listing-only wording for a found command: ${out}"
    matches 'command not found — configuration listing only; invocation not supported' "${out}" \
      || violation "runtime jini mcp list: missing listing-only wording for a missing command: ${out}"
    scan_doc "runtime jini mcp list" <<< "${out}"

    # savings: exit 0, neutral throttle wording, every $ figure qualified.
    out="$(run savings)"; code=$?
    [ "${code}" -eq 0 ] || violation "runtime jini savings: expected exit 0, got ${code}"
    matches "${MONEY}" "${out}" || violation "runtime jini savings: expected a dollar figure from the seeded ledger (missing output): ${out}"
    matches 'imputed' "${out}" || violation "runtime jini savings: expected an 'imputed' label (missing output): ${out}"
    scan_money "runtime jini savings" <<< "${out}"

    # startup savings counter: launcher prints it (stdin closed so it exits).
    out="$(JINI_HOME="${home}" "${bin}" </dev/null 2>&1)"; code=$?
    [ "${code}" -eq 0 ] || violation "runtime startup counter: launcher expected exit 0, got ${code}"
    counter="$(printf '%s\n' "${out}" | grep -iE '≈ *\$|avoided api spend' | head -n 1)"
    [ -n "${counter}" ] || violation "runtime startup counter: expected a savings line from the seeded ledger (missing output)"
    [ -n "${counter}" ] && scan_money "runtime startup counter" <<< "${counter}"
  fi
fi

if [ "${fail}" -eq 0 ]; then echo "claim-language gate: clean"; exit 0; fi
echo "claim-language gate: FAILED" >&2
exit 1
