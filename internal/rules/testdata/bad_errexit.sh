#!/usr/bin/env bash
# @file bad_errexit.sh
# @brief Command substitutions with errexit off inside
# @description Without inherit_errexit a failure inside $(...) is ignored.
set -euo pipefail

#######################################
# @description Read the version of the project
# @noargs
# @stdout Version
#######################################
function _version {
  git describe --tags
  printf 'unknown\n'
}

#######################################
# @description Print the version
# @noargs
#######################################
function _main {
  local version
  version="$(_version)"
  printf '%s\n' "${version}"
}

_main
