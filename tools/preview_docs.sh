#!/usr/bin/env bash
set -euo pipefail

mode="serve"
dry_run=0
host="127.0.0.1"
port="4000"

usage() {
  cat <<'EOF'
Usage: tools/preview_docs.sh [serve|build] [--host HOST] [--port PORT] [--dry-run]

Build or serve the public docs from docs/ using a reproducible local path.

Backends:
  1. bundle exec jekyll (using docs/Gemfile) when gems are installed
  2. jekyll from PATH when available
  3. docker with the official jekyll image when available

Environment:
  JINI_DOCS_EXTRA_CA_CERT  Optional PEM/CRT file to trust inside the docker
                           fallback. Useful behind corporate TLS interception.

Examples:
  tools/preview_docs.sh serve
  tools/preview_docs.sh build --dry-run
  tools/preview_docs.sh serve --host 0.0.0.0 --port 4001
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    serve|build)
      mode="$1"
      shift
      ;;
    --host)
      host="${2:-}"
      shift 2
      ;;
    --port)
      port="${2:-}"
      shift 2
      ;;
    --dry-run)
      dry_run=1
      shift
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      printf 'Unknown argument: %s\n\n' "$1" >&2
      usage >&2
      exit 2
      ;;
  esac
done

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
docs_dir="${repo_root}/docs"
config_path="${docs_dir}/_config.yml"
gemfile_path="${docs_dir}/Gemfile"
destination_path="${docs_dir}/_site"
preview_config_path="${docs_dir}/.preview-config.local.yml"
preview_config_name="$(basename "${preview_config_path}")"
docker_docs_dir="/srv/jekyll/docs"
docker_config_path="${docker_docs_dir}/_config.yml"
docker_preview_config_path="${docker_docs_dir}/${preview_config_name}"
docker_destination_path="${docker_docs_dir}/_site"
docker_bundle_path="${docker_docs_dir}/vendor/bundle"
docker_extra_ca_path="/usr/local/share/ca-certificates/jini-preview-extra-ca.crt"
docs_url="http://${host}:${port}"
extra_ca_cert="${JINI_DOCS_EXTRA_CA_CERT:-}"

trap 'rm -f "${preview_config_path}"' EXIT

cat > "${preview_config_path}" <<EOF
url: ${docs_url}
baseurl: ""
exclude:
  - vendor
  - .bundle
EOF

if [[ -z "${extra_ca_cert}" ]]; then
  for candidate in \
    "${HOME}/ca-cert/ZscalerRootCertificate-2048-SHA256.crt" \
    "${HOME}/ca-cert/ZscalerRootCertificate.crt"
  do
    if [[ -r "${candidate}" ]]; then
      extra_ca_cert="${candidate}"
      break
    fi
  done
fi

local_common_args=(
  "--source" "${docs_dir}"
  "--config" "${config_path},${preview_config_path}"
)

if [[ "${mode}" == "serve" ]]; then
  local_mode_args=(
    "serve"
    "--host" "${host}"
    "--port" "${port}"
    "--livereload"
    "--force_polling"
  )
  docker_mode_args=(
    "serve"
    "--host" "0.0.0.0"
    "--port" "4000"
    "--livereload"
    "--force_polling"
  )
else
  local_mode_args=(
    "build"
    "--destination" "${destination_path}"
  )
  docker_mode_args=(
    "build"
    "--destination" "${docker_destination_path}"
  )
fi

backend=""
declare -a command=()

if [[ -f "${gemfile_path}" ]] && BUNDLE_GEMFILE="${gemfile_path}" bundle exec ruby -e 'require "jekyll"' >/dev/null 2>&1; then
  backend="bundle"
  command=(env "BUNDLE_GEMFILE=${gemfile_path}" bundle exec jekyll "${local_mode_args[@]}" "${local_common_args[@]}")
elif command -v jekyll >/dev/null 2>&1; then
  backend="jekyll"
  command=(jekyll "${local_mode_args[@]}" "${local_common_args[@]}")
elif command -v docker >/dev/null 2>&1; then
  backend="docker"
  command=(
    docker run --rm
    -v "${repo_root}:/srv/jekyll"
    -w "${docker_docs_dir}"
  )
  if [[ -n "${extra_ca_cert}" ]] && [[ -r "${extra_ca_cert}" ]]; then
    command+=(-v "${extra_ca_cert}:/tmp/jini-preview-extra-ca.crt:ro")
  fi
  if [[ "${mode}" == "serve" ]]; then
    command+=(-p "${port}:4000")
  fi
  docker_jekyll_cmd="apk add --no-cache ca-certificates >/dev/null && "
  if [[ -n "${extra_ca_cert}" ]] && [[ -r "${extra_ca_cert}" ]]; then
    docker_jekyll_cmd+="cp /tmp/jini-preview-extra-ca.crt '${docker_extra_ca_path}' && "
  fi
  docker_jekyll_cmd+="update-ca-certificates >/dev/null 2>&1 && bundle config set --local path '${docker_bundle_path}' >/dev/null 2>&1 && bundle install && bundle exec jekyll"
  if [[ "${mode}" == "serve" ]]; then
    docker_jekyll_cmd+=" serve --host 0.0.0.0 --port 4000 --livereload --force_polling"
  else
    docker_jekyll_cmd+=" build --destination '${docker_destination_path}'"
  fi
  docker_jekyll_cmd+=" --source '${docker_docs_dir}' --config '${docker_config_path},${docker_preview_config_path}'"
  command+=(
    jekyll/jekyll:4
    bash -lc "${docker_jekyll_cmd}"
  )
else
  cat >&2 <<EOF
No docs preview backend is available.

Install one of:
  - bundle dependencies: BUNDLE_GEMFILE=docs/Gemfile bundle install
  - jekyll on PATH
  - docker for the jekyll/jekyll:4 fallback
EOF
  exit 1
fi

if [[ "${dry_run}" -eq 1 ]]; then
  printf 'backend=%s\n' "${backend}"
  printf 'command='
  printf '%q ' "${command[@]}"
  printf '\n'
  exit 0
fi

printf 'Using %s backend for docs %s at %s\n' "${backend}" "${mode}" "${docs_url}"
"${command[@]}"
