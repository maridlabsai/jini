#!/usr/bin/env bash
set -euo pipefail

PROGRAM_NAME="jini"
DEFAULT_PREFIX="${HOME}/.local"
DEFAULT_INSTALL_DIR="${DEFAULT_PREFIX}/share/jini"
DEFAULT_REPO_URL="https://github.com/maridlabsai/jini.git"
DEFAULT_REPO_REF="main"
DEFAULT_RELEASE_BASE_URL="https://github.com/maridlabsai/jini/releases/latest/download"
DEFAULT_INSTALL_VERSION="0.1.0"

BIN_DIR="${JINI_BIN_DIR:-}"
INSTALL_DIR="${JINI_INSTALL_DIR:-$DEFAULT_INSTALL_DIR}"
SOURCE_DIR="${JINI_SOURCE_DIR:-}"
REPO_URL="${JINI_INSTALL_REPO:-$DEFAULT_REPO_URL}"
REPO_REF="${JINI_INSTALL_REF:-$DEFAULT_REPO_REF}"
RELEASE_BASE_URL="${JINI_RELEASE_BASE_URL:-$DEFAULT_RELEASE_BASE_URL}"
GO_BIN="${JINI_GO_BIN:-}"
FORCE_INSTALL=0
COPY_BINARY=0
TEMP_ROOT=""
SOURCE_REASON=""
INSTALL_MODE="source-build"
INSTALL_DETAIL="unknown"
RELEASE_VALIDATION="not-attempted"

usage() {
  cat <<'EOF'
Install Jini so the command is simply `jini`.

Usage:
  ./install.sh [options]

Options:
  --bin-dir PATH       Directory that should contain the `jini` command.
  --install-dir PATH   Directory that should hold the built binary and receipt.
  --source-dir PATH    Local Jini source directory to build from.
  --repo-url URL       Git repository to clone when no local source is provided.
  --repo-ref REF       Git ref to clone when no local source is provided. Default: main
  --copy               Copy the binary into bin-dir instead of symlinking.
  --force              Replace an existing install.
  --help               Show this help text.
EOF
}

say() {
  printf '%s\n' "$*"
}

fail() {
  printf 'install failed: %s\n' "$*" >&2
  exit 1
}

cleanup() {
  if [[ -n "${TEMP_ROOT}" && -d "${TEMP_ROOT}" ]]; then
    rm -rf "${TEMP_ROOT}"
  fi
}

trap cleanup EXIT

while [[ $# -gt 0 ]]; do
  case "$1" in
    --bin-dir)
      [[ $# -ge 2 ]] || fail "--bin-dir needs a value"
      BIN_DIR="$2"
      shift 2
      ;;
    --install-dir)
      [[ $# -ge 2 ]] || fail "--install-dir needs a value"
      INSTALL_DIR="$2"
      shift 2
      ;;
    --source-dir)
      [[ $# -ge 2 ]] || fail "--source-dir needs a value"
      SOURCE_DIR="$2"
      SOURCE_REASON="explicit-source-dir"
      shift 2
      ;;
    --repo-url)
      [[ $# -ge 2 ]] || fail "--repo-url needs a value"
      REPO_URL="$2"
      shift 2
      ;;
    --repo-ref)
      [[ $# -ge 2 ]] || fail "--repo-ref needs a value"
      REPO_REF="$2"
      shift 2
      ;;
    --copy)
      COPY_BINARY=1
      shift
      ;;
    --force)
      FORCE_INSTALL=1
      shift
      ;;
    --help|-h)
      usage
      exit 0
      ;;
    *)
      fail "unknown option: $1"
      ;;
  esac
done

script_dir=""
script_source="${BASH_SOURCE[0]-}"
case "${script_source}" in
  ""|"bash"|/dev/fd/*|/proc/self/fd/*)
    ;;
  *)
    if [[ -f "${script_source}" ]]; then
      script_dir="$(
        cd "$(dirname "${script_source}")" >/dev/null 2>&1
        pwd
      )"
    fi
    ;;
esac

detect_local_source() {
  local candidate="$1"
  [[ -f "${candidate}/go.mod" && -f "${candidate}/cmd/jini/main.go" ]]
}

require_go() {
  if [[ -n "${GO_BIN}" ]]; then
    [[ -x "${GO_BIN}" ]] || fail "Configured JINI_GO_BIN is not executable: ${GO_BIN}"
  else
    GO_BIN="$(command -v go 2>/dev/null || true)"
  fi
  if [[ -z "${GO_BIN}" ]]; then
    fail "Go is required for source installs. Install Go or use a published release asset."
  fi
  "${GO_BIN}" version >/dev/null 2>&1 || fail "Go is present but not runnable."
}

ensure_dir_ready() {
  local path="$1"
  local label="$2"
  mkdir -p "${path}" 2>/dev/null || fail "${label} directory could not be created: ${path}"
  [[ -d "${path}" ]] || fail "${label} directory is not a directory: ${path}"
  [[ -w "${path}" ]] || fail "${label} directory is not writable: ${path}"
}

pick_bin_dir() {
  local path_entry=""
  if [[ -n "${BIN_DIR}" ]]; then
    printf '%s\n' "${BIN_DIR}"
    return
  fi

  local local_bin="${DEFAULT_PREFIX}/bin"
  if [[ -d "${local_bin}" && -w "${local_bin}" ]]; then
    printf '%s\n' "${local_bin}"
    return
  fi

  IFS=':' read -r -a path_entries <<<"${PATH}"
  for path_entry in "${path_entries[@]}"; do
    [[ -n "${path_entry}" ]] || continue
    [[ "${path_entry}" == "." ]] && continue
    [[ "${path_entry}" == "${HOME}"* ]] || continue
    if [[ -d "${path_entry}" && -w "${path_entry}" ]]; then
      printf '%s\n' "${path_entry}"
      return
    fi
  done

  printf '%s\n' "${local_bin}"
}

should_try_release_install() {
  [[ -z "${SOURCE_DIR}" && "${REPO_URL}" == "${DEFAULT_REPO_URL}" ]]
}

detect_release_asset() {
  local os_name=""
  local arch_name=""
  case "$(uname -s)" in
    Darwin) os_name="darwin" ;;
    Linux) os_name="linux" ;;
    *) return 1 ;;
  esac
  case "$(uname -m)" in
    x86_64|amd64) arch_name="amd64" ;;
    arm64|aarch64) arch_name="arm64" ;;
    *) return 1 ;;
  esac
  printf 'jini-%s-%s.tar.gz\n' "${os_name}" "${arch_name}"
}

try_install_prebuilt_release() {
  local asset=""
  local archive_path=""
  local unpack_dir=""
  local binary_path=""

  command -v curl >/dev/null 2>&1 || return 1
  command -v tar >/dev/null 2>&1 || return 1
  asset="$(detect_release_asset)" || return 1
  archive_path="${TEMP_ROOT}/${asset}"
  unpack_dir="${TEMP_ROOT}/release"

  if ! curl -fsSL "${RELEASE_BASE_URL}/${asset}" -o "${archive_path}" >/dev/null 2>&1; then
    return 1
  fi
  mkdir -p "${unpack_dir}"
  if ! tar -xzf "${archive_path}" -C "${unpack_dir}" >/dev/null 2>&1; then
    return 1
  fi
  binary_path="${unpack_dir}/${PROGRAM_NAME}"
  [[ -f "${binary_path}" ]] || return 1
  mv "${binary_path}" "${TARGET_BINARY}"
  chmod 0755 "${TARGET_BINARY}"
}

release_binary_supports_public_contract() {
  local binary_path="$1"
  "${binary_path}" publish-readiness --format json >/dev/null 2>&1
}

fetch_source() {
  local destination="${TEMP_ROOT}/source"
  command -v git >/dev/null 2>&1 || fail "Git is required when installing from a repository."
  git clone --depth 1 --branch "${REPO_REF}" "${REPO_URL}" "${destination}" >/dev/null 2>&1 ||
    fail "Could not clone ${REPO_URL} at ref ${REPO_REF}."
  SOURCE_DIR="${destination}"
  SOURCE_REASON="${SOURCE_REASON:-cloned-repo-source}"
}

install_provenance_summary() {
  case "${INSTALL_MODE}" in
    release-binary)
      printf 'release binary'
      ;;
    source-build)
      printf 'Go source build (%s)' "${SOURCE_REASON}"
      ;;
    *)
      printf '%s' "${INSTALL_MODE}"
      ;;
  esac
}

BIN_DIR="$(pick_bin_dir)"
TEMP_ROOT="${TEMP_ROOT:-$(mktemp -d)}"
export GOCACHE="${GOCACHE:-${TEMP_ROOT}/go-cache}"

ensure_dir_ready "${GOCACHE}" "Go build cache"
ensure_dir_ready "${BIN_DIR}" "Bin"
ensure_dir_ready "${INSTALL_DIR}" "Install"

if [[ -z "${SOURCE_DIR}" && -n "${script_dir}" ]] && detect_local_source "${script_dir}"; then
  SOURCE_DIR="${script_dir}"
  SOURCE_REASON="${SOURCE_REASON:-local-repo-source}"
fi

TARGET_BINARY="${INSTALL_DIR}/${PROGRAM_NAME}"
COMMAND_PATH="${BIN_DIR}/${PROGRAM_NAME}"

if [[ ${FORCE_INSTALL} -ne 1 ]]; then
  if [[ -e "${COMMAND_PATH}" && ! -L "${COMMAND_PATH}" ]]; then
    fail "${COMMAND_PATH} already exists. Rerun with --force to replace it."
  fi
  if [[ -e "${TARGET_BINARY}" && ! -w "${TARGET_BINARY}" ]]; then
    fail "${TARGET_BINARY} already exists and is not writable."
  fi
else
  rm -f "${COMMAND_PATH}" "${TARGET_BINARY}"
fi

if should_try_release_install; then
  if try_install_prebuilt_release && release_binary_supports_public_contract "${TARGET_BINARY}"; then
    INSTALL_MODE="release-binary"
    INSTALL_DETAIL="prebuilt-release-binary"
    SOURCE_REASON="release-binary"
    RELEASE_VALIDATION="passed"
  else
    rm -f "${TARGET_BINARY}"
    RELEASE_VALIDATION="release-unavailable-or-invalid"
    SOURCE_REASON="${SOURCE_REASON:-release-unavailable}"
  fi
fi

if [[ ! -f "${TARGET_BINARY}" ]]; then
  if [[ -z "${SOURCE_DIR}" ]]; then
    fetch_source
  fi
  detect_local_source "${SOURCE_DIR}" || fail "source directory does not look like the Jini repo: ${SOURCE_DIR}"
  require_go
  (
    cd "${SOURCE_DIR}"
    "${GO_BIN}" build -o "${TARGET_BINARY}" ./cmd/jini
  )
  chmod 0755 "${TARGET_BINARY}"
  INSTALL_MODE="source-build"
  INSTALL_DETAIL="go-source-build"
fi

if [[ ${COPY_BINARY} -eq 1 ]]; then
  cp "${TARGET_BINARY}" "${COMMAND_PATH}"
else
  rm -f "${COMMAND_PATH}"
  ln -sfn "${TARGET_BINARY}" "${COMMAND_PATH}"
fi
chmod 0755 "${COMMAND_PATH}" 2>/dev/null || true

VERSION="${DEFAULT_INSTALL_VERSION}"
if [[ -n "${SOURCE_DIR}" && -f "${SOURCE_DIR}/VERSION" ]]; then
  VERSION="$(tr -d '[:space:]' <"${SOURCE_DIR}/VERSION")"
fi

RECEIPT_PATH="${INSTALL_DIR}/install-receipt.txt"
cat >"${RECEIPT_PATH}" <<EOF
program=${PROGRAM_NAME}
version=${VERSION}
installed_at=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
source_dir=${SOURCE_DIR}
binary_path=${TARGET_BINARY}
command_path=${COMMAND_PATH}
copy_mode=$([[ ${COPY_BINARY} -eq 1 ]] && printf copy || printf symlink)
install_mode=${INSTALL_MODE}
install_detail=${INSTALL_DETAIL}
source_reason=${SOURCE_REASON}
release_validation=${RELEASE_VALIDATION}
EOF

"${COMMAND_PATH}" publish-readiness --format json >/dev/null

say "Installed Jini"
say "- version: ${VERSION}"
say "- install source: $(install_provenance_summary)"
say "- binary: ${TARGET_BINARY}"
say "- command: ${COMMAND_PATH}"
say "- receipt: ${RECEIPT_PATH}"
say

case ":${PATH}:" in
  *":${BIN_DIR}:"*)
    say "Run:"
    say "  jini"
    ;;
  *)
    say "Add Jini to your PATH:"
    say "  export PATH=\"${BIN_DIR}:\$PATH\""
    say
    say "Then run:"
    say "  jini"
    ;;
esac
