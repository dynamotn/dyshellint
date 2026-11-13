#!/usr/bin/env bash
# @file bad_returnvalue.sh
# @brief Data passed through the status of a function
# @description `return` takes a status; data belongs on STDOUT.
set -euo pipefail

#######################################
# @description Count the lines of a file, in every wrong way
# @arg $1 string File
#######################################
function _count {
  local file
  file="${1:?}"
  [[ -f "${file}" ]] || return false
  [[ -r "${file}" ]] || return -1
  [[ -s "${file}" ]] || return 300
  if [[ "${file}" == *.gz ]]; then
    return "$(zcat -- "${file}" | wc -l)"
  fi
  return $(wc -l < "${file}")
}

#######################################
# @description Count the lines of a file, and return a status
# @arg $1 string File
# @stdout Number of lines
#######################################
function _count_good {
  local file count
  file="${1:?}"
  [[ -f "${file}" ]] || return 1
  count="$(wc -l < "${file}")" || return
  printf '%s\n' "${count}"
  return "$((count > 0 ? 0 : 1))"
}

_count "$@" || _count_good "$@"
