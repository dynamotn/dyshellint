#!/usr/bin/env bash
# @file bad_seq.sh
# @brief Sequences from a process
# @description Brace expansion and an arithmetic for make them in the shell.
set -euo pipefail

#######################################
# @description Count up to a bound
# @arg $1 string Upper bound
#######################################
function _count {
  local count="$1" index
  for index in $(seq 1 "${count}"); do
    printf '%s\n' "${index}"
  done
  for ((index = 1; index <= count; index++)); do
    printf '%s\n' "${index}"
  done
  for index in {1..3}; do
    printf '%s\n' "${index}"
  done
}

_count "$@"
