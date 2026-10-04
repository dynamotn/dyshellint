#!/usr/bin/env bash
# @file bad_sideeffects.sh
# @brief Functions whose side effects a command substitution throws away
# @description A function run inside `$(...)` changes only the subshell.
set -euo pipefail

CACHE=""

#######################################
# @description Look a value up once, remembering it in a global
# @noargs
# @set CACHE
# @stdout The value
#######################################
function _lookup {
  [[ -n "${CACHE}" ]] || CACHE="computed"
  printf '%s\n' "${CACHE}"
}

#######################################
# @description Print a value, with locals only
# @noargs
# @stdout The value
#######################################
function _pure {
  local value="computed" line
  IFS=, read -r line <<< "a,b"
  TZ=UTC date +%s > /dev/null
  printf '%s %s\n' "${value}" "${line}"
}

#######################################
# @description Use both
# @noargs
#######################################
function _main {
  local first second
  first="$(_lookup)"
  second="$(_pure)"
  _lookup > /dev/null
  printf '%s %s\n' "${first}" "${second}"
}
