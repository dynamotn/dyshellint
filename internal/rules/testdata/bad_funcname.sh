#!/usr/bin/env bash
# @file bad_funcname.sh
# @brief Error messages that name a function by its depth in the stack
# @description A helper that reads FUNCNAME two frames up names the wrong
#   function as soon as it is reached through one more call.
set -euo pipefail

#######################################
# @description Refuse a value, naming the function two frames up
# @arg $1 string Value to check
#######################################
function _refuse {
  printf '%s: bad value %s\n' "${FUNCNAME[2]}" "$1" >&2
  local caller="${FUNCNAME[1]}"
  local who="${3:-${FUNCNAME[3]:-main}}"
  printf '%s %s %s\n' "${FUNCNAME[0]}" "${caller}" "${who}" >&2
  return 1
}

_refuse "x" || true
