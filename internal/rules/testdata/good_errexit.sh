#!/usr/bin/env bash
# @file good_errexit.sh
# @brief Command substitutions that keep errexit
# @description inherit_errexit passes set -e down.
if ((BASH_VERSINFO[0] < 4 || (BASH_VERSINFO[0] == 4 && BASH_VERSINFO[1] < 4))); then
  printf 'This script needs Bash 4.4 or newer, found %s\n' "${BASH_VERSION}" >&2
  exit 1
fi
set -euo pipefail
shopt -s inherit_errexit

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
