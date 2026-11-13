#!/usr/bin/env bash
# @file bad_descriptors.sh
# @brief Fixed descriptors and appends repeated in loops
# @description The descriptors may clash; each append reopens the file.
set -euo pipefail

#######################################
# @description Write a report through fixed descriptors
# @arg $1 string Report path
# @arg $2 string Input path
#######################################
function _report {
  local report="$1" input="$2"
  local host line
  exec 3> "${report}"
  exec 4< "${input}"
  exec 5>&1
  printf 'start\n' >&3
  exec 3>&-
  for host in alpha beta; do
    printf '%s\n' "${host}" >> "${report}"
    printf 'done %s\n' "${host}" >> "${report}"
  done
  while IFS= read -r line || [[ -n "${line}" ]]; do
    printf '%s\n' "${line}" &>> "${report}.log"
  done < "${input}"
}

_report "$@"
