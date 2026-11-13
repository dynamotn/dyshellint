#!/usr/bin/env bash
# @file good_nesting.sh
# @brief Guard clauses keep the nesting flat
# @description A guard clause that only leaves adds no level of its own.
set -euo pipefail

#######################################
# @description Print the matching lines of one file
# @arg $1 string File to read
#######################################
function _scan_file {
  local file="$1" line
  [[ -f "${file}" ]] || return 0
  while IFS= read -r line || [[ -n "${line}" ]]; do
    if [[ "${line}" != *match* ]]; then
      continue
    fi
    case "${line}" in
      *skip*) printf 'skip\n' ;;
      *)
        if [[ -n "${line}" ]]; then
          if [[ "${line}" == *stop* ]]; then
            break
          fi
          printf '%s\n' "${line}"
        elif [[ "${line}" == *empty* ]]; then
          printf 'empty\n'
        else
          printf 'other\n'
        fi
        ;;
    esac
  done < "${file}"
}

#######################################
# @description Scan every file
# @arg $@ string Files to read
#######################################
function _scan {
  local file
  for file in "$@"; do
    _scan_file "${file}"
  done
}

_scan "$@"
