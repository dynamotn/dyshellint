#!/usr/bin/env bash
# @file bad_arithinput.sh
# @brief Arithmetic on numbers nobody checked
# @description Arithmetic evaluates what it is given: `08` is a bad octal
#   number, and `a[$(cmd)]` runs the command.
set -euo pipefail

#######################################
# @description Add up its arguments without checking them
# @arg $1 string First number
# @arg $2 string Second number
#######################################
function _add {
  local first="$1" second="$2" line
  read -r line < /dev/null || true
  printf '%s\n' "$((first + second))"
  printf '%s\n' "$(($1 * 2))"
  ((line > 0)) || true
}

#######################################
# @description Add up its arguments after checking them
# @arg $1 string First number
# @arg $2 string Second number
#######################################
function _add_checked {
  local first="$1" second="$2" total=0
  [[ "${first}" =~ ^[0-9]+$ ]] || return 1
  case "${second}" in *[!0-9]* | '') return 1 ;; esac
  printf '%s\n' "$((first + second + total))"
  printf '%s\n' "$((10#$1))"
}
