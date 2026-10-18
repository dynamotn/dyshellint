#!/usr/bin/env bash
# @file bad_printing.sh
# @brief Data printed with echo
# @description echo reads options and backslashes from the data.
set -euo pipefail

#######################################
# @description Print a value, every wrong way
# @arg $1 string Value
#######################################
function _show {
  local value="$1"
  local -a items=("${value}")
  echo "${#items[@]}"
  echo "${#value}"
  echo "${value}"
  echo $value
  echo "$(date +%F)"
  echo -e "line\t${value}"
  echo -n "${value}"
  echo "value: ${value}"
  echo "Installing tools"
  echo $#
  echo "$?, done"
  printf '%s\n' "${value}"
}

_show "$@"
