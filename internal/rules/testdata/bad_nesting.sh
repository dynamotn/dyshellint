#!/usr/bin/env bash
# @file bad_nesting.sh
# @brief Control flow stacked too deep
# @description The fourth level is reported once, not every level below it.
set -euo pipefail

if [[ -n "${1-}" ]]; then
  for arg in "$@"; do
    while [[ -n "${arg}" ]]; do
      until [[ -z "${arg}" ]]; do
        arg=""
      done
    done
  done
fi

#######################################
# @description Print the matching lines of every file
# @arg $@ string Files to read
#######################################
function _scan {
  local file line
  for file in "$@"; do
    if [[ -f "${file}" ]]; then
      while IFS= read -r line || [[ -n "${line}" ]]; do
        if [[ "${line}" == *match* ]]; then
          case "${line}" in
            *skip*) printf 'skip\n' ;;
            *) printf '%s\n' "${line}" ;;
          esac
        elif [[ -z "${line}" ]]; then
          if [[ -n "${arg-}" ]]; then
            continue
          fi
          printf 'empty\n'
        fi
      done < "${file}"
    fi
  done
}

_scan "$@"
