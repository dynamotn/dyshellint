#!/usr/bin/env bash
# @file bad_globloop.sh
# @brief Glob loops that run on a pattern matching nothing
# @description Without nullglob, an unmatched glob is the loop's only word.
set -euo pipefail

#######################################
# @description Walk the configuration files, every way
# @arg $1 string Directory
#######################################
function _walk {
  local dir="$1" conf
  for conf in "${dir}"/*.conf; do
    printf '%s\n' "${conf}"
  done
  for conf in "${dir}"/*.conf; do
    [[ -e "${conf}" ]] || continue
    printf '%s\n' "${conf}"
  done
  for conf in "${dir}/literal.conf" "${dir}"; do
    printf '%s\n' "${conf}"
  done
}
