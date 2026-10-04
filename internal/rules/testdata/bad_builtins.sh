#!/usr/bin/env bash
# @file bad_builtins.sh
# @brief A process to read a file
# @description $(< file) reads it in the shell.
set -euo pipefail
# shellcheck disable=SC1091
. ./lib.sh
# shellcheck disable=SC2034,SC1091 # sourced by the caller
. ./other.sh
# shellcheck source=lib.sh
. ./lib.sh

#######################################
# @description Read a version file
# @arg $1 string File
#######################################
function _read {
  local file="$1" version
  version="$(cat "${file}")"
  version="$(< "${file}")"
  version="$(cat -- "${file}")"
  version="$(cat "${file}" | head -n 1)"
  printf '%s\n' "${version}"
}

_read "$@"
