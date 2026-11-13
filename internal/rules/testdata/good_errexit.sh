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
# @description Print the name given, or stop the script without one
# @arg $1 string Name
# @stdout The name
#######################################
function _name {
  [[ -n "${1:-}" ]] || exit 1
  printf '%s\n' "$1"
}

#######################################
# @description Print the version
# @noargs
#######################################
function _main {
  local version name
  version="$(_version)"
  name="$(_name "${1:-}")"
  printf '%s %s\n' "${version}" "${name}"
}

_main "$@"
