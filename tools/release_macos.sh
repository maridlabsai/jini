#!/usr/bin/env bash
# Build, sign, notarize, and package a macOS Jini release asset.
#
# The signing identity and notary profile are NEVER stored in the repo — they
# come from the environment, so this runs on a maintainer's machine (or CI with
# secrets), not in a normal build. Without them the script stops before it would
# produce an unsigned asset that install.sh will later reject.
#
# Required env:
#   JINI_SIGNING_IDENTITY   Developer ID Application identity (e.g. "Developer ID
#                           Application: Your Name (TEAMID)").
#   JINI_NOTARY_PROFILE     `xcrun notarytool` keychain profile name created with
#                           `xcrun notarytool store-credentials`.
# Optional env:
#   JINI_RELEASE_ARCH       darwin arch to build (arm64 or amd64). Default: host.
#   JINI_RELEASE_OUT        output directory. Default: ./dist.
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
OUT_DIR="${JINI_RELEASE_OUT:-${ROOT_DIR}/dist}"

fail() { printf 'release failed: %s\n' "$*" >&2; exit 1; }

[[ "$(uname -s)" == "Darwin" ]] || fail "macOS signing/notarization must run on macOS."
command -v codesign >/dev/null 2>&1 || fail "codesign not found (install Xcode command line tools)."
command -v xcrun >/dev/null 2>&1 || fail "xcrun not found (install Xcode command line tools)."
command -v go >/dev/null 2>&1 || fail "go not found."

[[ -n "${JINI_SIGNING_IDENTITY:-}" ]] || fail "set JINI_SIGNING_IDENTITY to a Developer ID Application identity."
[[ -n "${JINI_NOTARY_PROFILE:-}" ]] || fail "set JINI_NOTARY_PROFILE to a notarytool keychain profile."

case "${JINI_RELEASE_ARCH:-}" in
  arm64|amd64) arch="${JINI_RELEASE_ARCH}" ;;
  "") case "$(uname -m)" in
        arm64|aarch64) arch="arm64" ;;
        x86_64|amd64)  arch="amd64" ;;
        *) fail "unsupported host arch $(uname -m); set JINI_RELEASE_ARCH." ;;
      esac ;;
  *) fail "JINI_RELEASE_ARCH must be arm64 or amd64." ;;
esac

version="$(tr -d '[:space:]' < "${ROOT_DIR}/VERSION" 2>/dev/null || echo "0.0.0")"
mkdir -p "${OUT_DIR}"
work="$(mktemp -d)"; trap 'rm -rf "${work}"' EXIT
binary="${work}/jini"

printf 'Building jini %s (darwin/%s)\n' "${version}" "${arch}"
( cd "${ROOT_DIR}" && GOOS=darwin GOARCH="${arch}" CGO_ENABLED=0 go build -trimpath -o "${binary}" ./cmd/jini )

printf 'Signing with hardened runtime\n'
codesign --force --options runtime --timestamp \
  --sign "${JINI_SIGNING_IDENTITY}" "${binary}" \
  || fail "codesign failed."
codesign --verify --strict --verbose=2 "${binary}" || fail "codesign verification failed."

# Notarize the binary inside a zip (notarytool accepts zip/pkg/dmg).
notarize_zip="${work}/jini-notarize.zip"
( cd "${work}" && /usr/bin/zip -q "${notarize_zip}" jini )
printf 'Submitting to notary service (waits for result)\n'
xcrun notarytool submit "${notarize_zip}" --keychain-profile "${JINI_NOTARY_PROFILE}" --wait \
  || fail "notarization submission failed."
# CLI binaries can't be stapled directly (no bundle); Gatekeeper verifies the
# notarized signature online. Confirm the ticket resolves for this binary.
xcrun stapler validate "${binary}" 2>/dev/null \
  || printf 'note: stapler validate not applicable to a bare binary; Gatekeeper checks the notarized signature online.\n'

asset="jini-darwin-${arch}.tar.gz"
( cd "${work}" && tar -czf "${OUT_DIR}/${asset}" jini )
( cd "${OUT_DIR}" && shasum -a 256 "${asset}" > "${asset}.sha256" )

printf 'Done:\n- %s\n- %s\n' "${OUT_DIR}/${asset}" "${OUT_DIR}/${asset}.sha256"
printf 'Verify a downloaded copy with: shasum -a 256 -c %s.sha256\n' "${asset}"
