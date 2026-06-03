#!/usr/bin/env bash
set -euo pipefail

PROGRAM_NAME="jini"
DEFAULT_PREFIX="${HOME}/.local"
DEFAULT_BIN_DIR=""
DEFAULT_INSTALL_DIR="${DEFAULT_PREFIX}/share/jini"
DEFAULT_REPO_URL="https://github.com/maridlabsai/jini.git"
DEFAULT_REPO_REF="main"
DEFAULT_RELEASE_BASE_URL="https://github.com/maridlabsai/jini/releases/latest/download"

BIN_DIR="${JINI_BIN_DIR:-$DEFAULT_BIN_DIR}"
INSTALL_DIR="${JINI_INSTALL_DIR:-$DEFAULT_INSTALL_DIR}"
SOURCE_DIR="${JINI_SOURCE_DIR:-}"
REPO_URL="${JINI_INSTALL_REPO:-$DEFAULT_REPO_URL}"
REPO_REF="${JINI_INSTALL_REF:-$DEFAULT_REPO_REF}"
RELEASE_BASE_URL="${JINI_RELEASE_BASE_URL:-$DEFAULT_RELEASE_BASE_URL}"
FORCE_INSTALL=0
COPY_BINARY=0
TEMP_ROOT=""
SOURCE_DIR_EXPLICIT=0

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

if [[ -n "${SOURCE_DIR}" ]]; then
  SOURCE_DIR_EXPLICIT=1
fi

script_dir="$(pwd)"
script_source="${BASH_SOURCE[0]-}"
if [[ -n "${script_source}" && "${script_source}" != "bash" ]]; then
  script_dir="$(
    cd "$(dirname "${script_source}")" >/dev/null 2>&1
    pwd
  )"
fi

detect_local_source() {
  local candidate="$1"
  [[ -f "${candidate}/go.mod" && -f "${candidate}/cmd/jini/main.go" ]]
}

python_source_fallback_ready() {
  command -v python3 >/dev/null 2>&1 || return 1
}

ensure_python_yaml_runtime() {
  local app_dir="$1"
  local vendor_dir="${app_dir}/vendor"

  if python3 -c 'import yaml' >/dev/null 2>&1; then
    return 0
  fi

  python3 -m pip --version >/dev/null 2>&1 || return 1
  mkdir -p "${vendor_dir}"
  python3 -m pip install --disable-pip-version-check --no-input --target "${vendor_dir}" "PyYAML>=6,<7" >/dev/null 2>&1 || return 1
  JINI_VENDOR_DIR="${vendor_dir}" python3 - <<'PY' >/dev/null 2>&1 || return 1
import os
import sys

sys.path.insert(0, os.environ["JINI_VENDOR_DIR"])
import yaml  # noqa: F401
PY
}

go_source_fallback_ready() {
  command -v go >/dev/null 2>&1 || return 1
  go version >/dev/null 2>&1 || return 1
}

detect_go_command_path() {
  GO_COMMAND_PATH="$(command -v go 2>/dev/null || true)"
  [[ -n "${GO_COMMAND_PATH}" ]]
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
  local download_url=""
  local archive_path=""
  local unpack_dir=""
  local binary_path=""

  command -v curl >/dev/null 2>&1 || return 1
  command -v tar >/dev/null 2>&1 || return 1
  asset="$(detect_release_asset)" || return 1
  download_url="${RELEASE_BASE_URL}/${asset}"
  archive_path="${TEMP_ROOT}/${asset}"
  unpack_dir="${TEMP_ROOT}/release"

  if ! curl -fsSL "${download_url}" -o "${archive_path}" >/dev/null 2>&1; then
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
  return 0
}

release_binary_supports_public_contract() {
  local binary_path="$1"
  "${binary_path}" doctor >/dev/null 2>&1
}

install_provenance_summary() {
  if [[ "${INSTALL_MODE}" == "release-binary" ]]; then
    printf 'release binary'
    return 0
  fi

  case "${SOURCE_REASON}" in
    release-validation-failed)
      printf 'source fallback (release validation failed: %s)' "${RELEASE_VALIDATION}"
      ;;
    explicit-source-dir)
      printf 'source runtime (explicit source)'
      ;;
    local-repo-source)
      printf 'source runtime (local repo)'
      ;;
    cloned-repo-source)
      printf 'source runtime (cloned repo)'
      ;;
    *)
      printf 'source runtime (%s)' "${SOURCE_REASON}"
      ;;
  esac
}

install_provenance_next_step() {
  case "${SOURCE_REASON}" in
    explicit-source-dir)
      printf 'Keep using this checkout when you want local source changes to control Jini.'
      ;;
    release-unavailable)
      printf 'If this machine should have had a published release, file a release issue and include the receipt.'
      ;;
    release-validation-failed)
      printf 'Keep the source install, attach install-receipt.txt, and flag the stale release artifact.'
      ;;
    *)
      return 1
      ;;
  esac
}

try_install_python_source_runtime() {
  local app_dir="${INSTALL_DIR}/source-runtime"
  local wrapper_path="${TARGET_BINARY}"

  python_source_fallback_ready || return 1
  rm -rf "${app_dir}"
  mkdir -p "${app_dir}"
  cp -R "${SOURCE_DIR}/." "${app_dir}"
  rm -rf "${app_dir}/.git" "${app_dir}/.gocache" "${app_dir}/.jini"
  ensure_python_yaml_runtime "${app_dir}" || return 1

  cat >"${wrapper_path}" <<EOF
#!/usr/bin/env bash
set -euo pipefail
APP_DIR="${app_dir}"
if [[ -d "\${APP_DIR}/vendor" ]]; then
  export PYTHONPATH="\${APP_DIR}/vendor\${PYTHONPATH:+:\${PYTHONPATH}}"
fi
exec python3 "\${APP_DIR}/tools/jini.py" "\$@"
EOF
  chmod 0755 "${wrapper_path}"
  return 0
}

ensure_dir_ready() {
  local path="$1"
  local label="$2"
  if ! mkdir -p "${path}" 2>/dev/null; then
    fail "${label} directory could not be created: ${path}"
  fi
  if [[ ! -d "${path}" ]]; then
    fail "${label} directory is not a directory: ${path}"
  fi
  if [[ ! -w "${path}" ]]; then
    fail "${label} directory is not writable: ${path}"
  fi
}

pick_bin_dir() {
  local path_entry=""
  if [[ -n "${BIN_DIR}" ]]; then
    printf '%s\n' "${BIN_DIR}"
    return
  fi

  local home_bin="${HOME}/bin"
  if [[ -d "${home_bin}" && -w "${home_bin}" ]]; then
    printf '%s\n' "${home_bin}"
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

fetch_source() {
  local destination="${TEMP_ROOT}/source"
  local status=0
  if ! command -v git >/dev/null 2>&1; then
    fail "Git is required when installing outside the Jini repo."
  fi
  set +e
  git clone --depth 1 --branch "${REPO_REF}" "${REPO_URL}" "${destination}" >/dev/null 2>&1
  status=$?
  set -e
  if [[ ${status} -eq 0 ]]; then
    SOURCE_DIR="${destination}"
    return
  fi
  if [[ ${status} -eq 127 ]]; then
    fail "Git is required when installing outside the Jini repo."
  fi
  fail "Could not clone ${REPO_URL} at ref ${REPO_REF}."
}

SOURCE_REASON=""
if [[ ${SOURCE_DIR_EXPLICIT} -eq 1 ]]; then
  SOURCE_REASON="explicit-source-dir"
elif [[ -z "${SOURCE_DIR}" ]] && detect_local_source "${script_dir}"; then
  SOURCE_DIR="${script_dir}"
  SOURCE_REASON="local-repo-source"
fi

BIN_DIR="$(pick_bin_dir)"
TEMP_ROOT="${TEMP_ROOT:-$(mktemp -d)}"
export GOCACHE="${GOCACHE:-${TEMP_ROOT}/go-cache}"
ensure_dir_ready "${GOCACHE}" "Go build cache"
ensure_dir_ready "${BIN_DIR}" "Bin"
ensure_dir_ready "${INSTALL_DIR}" "Install"

VERSION="dev"
if [[ -f "${SOURCE_DIR}/VERSION" ]]; then
  VERSION="$(tr -d '[:space:]' <"${SOURCE_DIR}/VERSION")"
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

BUILD_OUTPUT="${TEMP_ROOT}/${PROGRAM_NAME}"
GO_SOURCE_BINARY="${INSTALL_DIR}/${PROGRAM_NAME}-go"
LEGACY_SOURCE_DIR=""
LEGACY_PYTHONPATH=""
GO_COMMAND_PATH=""
INSTALLED_FROM_RELEASE=0
INSTALL_MODE="source-runtime"
INSTALL_DETAIL="unknown"
RELEASE_VALIDATION="not-attempted"

if should_try_release_install; then
  if try_install_prebuilt_release; then
    if release_binary_supports_public_contract "${TARGET_BINARY}"; then
      INSTALLED_FROM_RELEASE=1
      INSTALL_MODE="release-binary"
      INSTALL_DETAIL="prebuilt-release-binary"
      SOURCE_REASON="release-binary"
      RELEASE_VALIDATION="passed"
    else
      rm -f "${TARGET_BINARY}"
      SOURCE_REASON="release-validation-failed"
      RELEASE_VALIDATION="unsupported-public-command-surface"
      say "Downloaded release asset did not support the current public command surface. Falling back to source install."
    fi
  else
    SOURCE_REASON="${SOURCE_REASON:-release-unavailable}"
    RELEASE_VALIDATION="release-unavailable"
  fi
fi

if [[ ${INSTALLED_FROM_RELEASE} -ne 1 ]]; then
  if [[ -z "${SOURCE_DIR}" ]]; then
    fetch_source
    SOURCE_REASON="${SOURCE_REASON:-cloned-repo-source}"
  fi
  detect_local_source "${SOURCE_DIR}" || fail "source directory does not look like the Jini repo: ${SOURCE_DIR}"
  if go_source_fallback_ready; then
    detect_go_command_path || fail "Go executable path could not be resolved."
    LEGACY_SOURCE_DIR="${SOURCE_DIR}"
    if [[ ${SOURCE_DIR_EXPLICIT} -eq 0 ]]; then
      LEGACY_SOURCE_DIR="${INSTALL_DIR}/source-runtime"
      rm -rf "${LEGACY_SOURCE_DIR}"
      mkdir -p "${LEGACY_SOURCE_DIR}"
      cp -R "${SOURCE_DIR}/." "${LEGACY_SOURCE_DIR}"
      rm -rf "${LEGACY_SOURCE_DIR}/.git" "${LEGACY_SOURCE_DIR}/.gocache" "${LEGACY_SOURCE_DIR}/.jini"
      ensure_python_yaml_runtime "${LEGACY_SOURCE_DIR}" || fail "Python fallback runtime could not be prepared."
      if [[ -d "${LEGACY_SOURCE_DIR}/vendor" ]]; then
        LEGACY_PYTHONPATH="${LEGACY_SOURCE_DIR}/vendor"
      fi
    else
      local_legacy_vendor="${INSTALL_DIR}/legacy-python"
      rm -rf "${local_legacy_vendor}"
      mkdir -p "${local_legacy_vendor}"
      ensure_python_yaml_runtime "${local_legacy_vendor}" || fail "Python fallback runtime could not be prepared."
      if [[ -d "${local_legacy_vendor}/vendor" ]]; then
        LEGACY_PYTHONPATH="${local_legacy_vendor}/vendor"
      fi
    fi
    if [[ ${SOURCE_DIR_EXPLICIT} -eq 1 ]]; then
      cat >"${TARGET_BINARY}" <<EOF
#!/usr/bin/env bash
set -euo pipefail
export JINI_SOURCE_DIR="${LEGACY_SOURCE_DIR}"
if [[ -n "${LEGACY_PYTHONPATH}" ]]; then
  export JINI_LEGACY_PYTHONPATH="${LEGACY_PYTHONPATH}"
fi
export GOCACHE="\${GOCACHE:-/private/tmp/jini-go-cache}"
GO_RUNNER="${GO_COMMAND_PATH}"
if [[ ! -x "\${GO_RUNNER}" ]]; then
  GO_RUNNER="\$(command -v go 2>/dev/null || true)"
fi
if [[ -z "\${GO_RUNNER}" ]]; then
  printf 'jini install is missing a working go executable; reinstall or restore Go on PATH.\n' >&2
  exit 1
fi
cd "${SOURCE_DIR}"
exec "\${GO_RUNNER}" run ./cmd/jini "\$@"
EOF
      chmod 0755 "${TARGET_BINARY}"
      INSTALL_DETAIL="go-source-live-run"
    else
      rm -f "${GO_SOURCE_BINARY}"
      if ! (
        cd "${SOURCE_DIR}"
        go build -o "${BUILD_OUTPUT}" ./cmd/jini
      ); then
        fail "Go build failed."
      fi
      mv "${BUILD_OUTPUT}" "${GO_SOURCE_BINARY}"
      chmod 0755 "${GO_SOURCE_BINARY}"
      cat >"${TARGET_BINARY}" <<EOF
#!/usr/bin/env bash
set -euo pipefail
export JINI_SOURCE_DIR="${LEGACY_SOURCE_DIR}"
if [[ -n "${LEGACY_PYTHONPATH}" ]]; then
  export JINI_LEGACY_PYTHONPATH="${LEGACY_PYTHONPATH}"
fi
exec "${GO_SOURCE_BINARY}" "\$@"
EOF
      chmod 0755 "${TARGET_BINARY}"
      INSTALL_DETAIL="go-source-build"
    fi
  elif ! try_install_python_source_runtime; then
    go_source_fallback_ready || fail "Jini could not install a release binary, and source fallback needs either Python 3 (Jini will try to add PyYAML automatically) or Go."
  else
    INSTALL_DETAIL="python-source-runtime"
  fi
fi

if [[ ${COPY_BINARY} -eq 1 ]]; then
  cp "${TARGET_BINARY}" "${COMMAND_PATH}"
else
  rm -f "${COMMAND_PATH}"
  ln -sfn "${TARGET_BINARY}" "${COMMAND_PATH}"
fi
chmod 0755 "${COMMAND_PATH}" 2>/dev/null || true

NEXT_STEP=""
if next_step="$(install_provenance_next_step)"; then
  NEXT_STEP="${next_step}"
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
if [[ -n "${NEXT_STEP}" ]]; then
  printf 'next_step=%s\n' "${NEXT_STEP}" >>"${RECEIPT_PATH}"
fi

"${COMMAND_PATH}" doctor >/dev/null

say "Installed Jini"
say "- version: ${VERSION}"
say "- install source: $(install_provenance_summary)"
if [[ -n "${NEXT_STEP}" ]]; then
  say "- next step: ${NEXT_STEP}"
  say "- support receipt: ${RECEIPT_PATH} (send version=, source_reason=, release_validation=, next_step=)"
fi
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
