#!/usr/bin/env bash
# @file bad_arrayscalar.sh
# @brief Plain values assigned to arrays
# @description Only element 0 changes.
set -euo pipefail

#######################################
# @description Collect files, and reset the list the wrong way
# @arg $1 string File
#######################################
function _collect {
  local file="$1"
  local -a files=()
  local name=""
  files+=("${file}")
  files="${file}"
  files+="${file}"
  files=("${file}")
  files[1]="${file}"
  name="${file}"
  printf '%s\n' "${files[@]}" "${name}"
}

#######################################
# @description A scalar of the same name in another function
# @arg $1 string File
#######################################
function _other {
  local files="$1"
  printf '%s\n' "${files}"
}

_collect "$@"
_other "$@"
