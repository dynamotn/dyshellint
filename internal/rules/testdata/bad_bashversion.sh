#!/usr/bin/env bash
# @file bad_bashversion.sh
# @brief Bash 4.4 features with no version check
# @description Fails on Bash 3.2 far from the cause.
set -euo pipefail

#######################################
# @description Split a list into the caller's array
# @arg $1 string Name of the array
#######################################
function _split_into {
  local -n __split_into_ref="$1"
  mapfile -d '' -t __split_into_ref < <(printf 'a\0b\0')
}

_split_into items
