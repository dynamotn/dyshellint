#!/usr/bin/env bash
# @file bad_subst.sh
# @brief Functions that can stop the script, called inside a substitution
# @description A refusal inside `$(...)` ends only the subshell, so each call
#   below either loses the refusal or checks for it on the spot.
set -euo pipefail

#######################################
# @description Print a number, or stop the script on anything else
# @arg $1 string Value to check
# @stdout The number
#######################################
function _parse {
  local value="$1"
  [[ "${value}" =~ ^[0-9]+$ ]] || exit 1
  printf '%s\n' "${value}"
}

#######################################
# @description Pass the value on to the parser
# @arg $1 string Value to check
# @stdout The number
#######################################
function _wrapper {
  _parse "$1"
}

#######################################
# @description Print a value, which never stops the script
# @arg $1 string Value to print
# @stdout The value
#######################################
function _echo {
  printf '%s\n' "$1"
}

#######################################
# @description Call the parser every way, checked and not
# @arg $1 string Value to check
#######################################
function _main {
  local number checked plain
  number=$(_parse "$1")
  checked=$(_parse "$1") || return 1
  if checked=$(_wrapper "$1"); then
    printf '%s\n' "${checked}"
  fi
  printf '%s\n' "$(_parse "$1")"
  number=$(_wrapper "$1")
  plain=$(_echo "$1")
  printf '%s %s\n' "${number}" "${plain}"
}

_main "$@"
