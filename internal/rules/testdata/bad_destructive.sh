#!/usr/bin/env bash
# @file bad_destructive.sh
# @brief Recursive deletions of paths nothing checked
# @description An empty variable turns `rm -rf "${dir}/"` into `rm -rf /`.
set -euo pipefail

#######################################
# @description Clean a build directory, every way
# @arg $1 string Build directory
#######################################
function _clean {
  local dir="$1" scratch checked="$1"
  rm -rf "${dir}/"
  find "${dir}" -name '*.o' -delete
  chmod -R go-w "${dir}"
  rm -rf "${dir:?}/out"
  scratch="$(mktemp -d)"
  rm -rf "${scratch}"
  [[ -n "${checked}" ]] || return 1
  rm -rf -- "${checked}"
  rm -f "${dir}/one.o"
  local partial="${dir}/.partial"
  rm -rf -- "${partial}"
}
