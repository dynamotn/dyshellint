#!/usr/bin/env bash
# @file bad_readloop.sh
# @brief Read loops that drop a last line without a newline
# @description `read` reads the last line but fails, so the body never sees it.
set -euo pipefail

#######################################
# @description Print every line of a file, every way
# @arg $1 string File
#######################################
function _lines {
  local file="$1" line
  while read -r line; do
    printf '%s\n' "${line}"
  done < "${file}"
  while IFS= read -r line; do
    printf '%s\n' "${line}"
  done < <(cat "${file}")
  while IFS= read -r line; do
    printf '%s\n' "${line}"
  done < <(git ls-files)
  while read -r line || [[ -n "${line}" ]]; do
    printf '%s\n' "${line}"
  done < "${file}"
  while read -r line; do
    printf '%s\n' "${line}"
  done <<< "a"
  while IFS= read -r -d '' line; do
    printf '%s\n' "${line}"
  done < <(find . -print0)
}
