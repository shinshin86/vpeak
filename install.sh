#!/usr/bin/env bash

set -e

OWNER="shinshin86"
REPO="vpeak"
BIN_NAME="vpeak"

VERSION="${1:-latest}"
BIN_DIR="${BIN_DIR:-${HOME}/.local/bin}"

err() {
  echo "Error: $*" >&2
  exit 1
}

info() {
  echo "Info: $*"
}

need_cmd() {
  if ! command -v "$1" >/dev/null 2>&1; then
    err "Required command not found: $1"
  fi
}

detect_os() {
  local os
  os=$(uname -s | tr '[:upper:]' '[:lower:]')
  if [ "${os}" != "darwin" ]; then
    err "Unsupported OS: ${os} (only darwin supported)"
  fi
  echo "${os}"
}

detect_arch() {
  local arch
  arch=$(uname -m)
  case "${arch}" in
    x86_64) echo "amd64" ;;
    arm64) echo "arm64" ;;
    *)
      err "Unsupported architecture: ${arch}"
      ;;
  esac
}

main() {
  need_cmd curl
  need_cmd tar
  need_cmd install
  need_cmd shasum

  local os arch asset url tmp_dir checksum

  os=$(detect_os)
  arch=$(detect_arch)

  if [ "${VERSION}" = "latest" ]; then
    info "Fetching latest version..."
    VERSION=$(curl -fsSL "https://api.github.com/repos/${OWNER}/${REPO}/releases/latest" | \
      grep -m 1 '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    if [ -z "${VERSION}" ]; then
      err "Failed to fetch latest version"
    fi
  fi

  asset="${BIN_NAME}_${VERSION}_${os}_${arch}.tar.gz"
  url="https://github.com/${OWNER}/${REPO}/releases/download/${VERSION}/${asset}"

  tmp_dir=$(mktemp -d)
  trap 'rm -rf "${tmp_dir}"' EXIT

  info "Downloading ${url}..."
  curl -fsSL "${url}" -o "${tmp_dir}/${asset}"

  info "Downloading checksums..."
  curl -fsSL "https://github.com/${OWNER}/${REPO}/releases/download/${VERSION}/checksums.txt" \
    -o "${tmp_dir}/checksums.txt"

  checksum=$(grep " ${asset}$" "${tmp_dir}/checksums.txt" | awk '{print $1}')
  if [ -z "${checksum}" ]; then
    err "Checksum not found for ${asset}"
  fi

  info "Verifying checksum..."
  echo "${checksum}  ${tmp_dir}/${asset}" | shasum -a 256 -c -

  info "Installing to ${BIN_DIR}..."
  mkdir -p "${BIN_DIR}"
  tar -xzf "${tmp_dir}/${asset}" -C "${tmp_dir}"
  install -m 755 "${tmp_dir}/${BIN_NAME}" "${BIN_DIR}/${BIN_NAME}"

  info "Done! Run '${BIN_NAME} --version' to verify."
}

main "$@"
