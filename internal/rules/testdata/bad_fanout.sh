#!/usr/bin/env bash
# @file bad_fanout.sh
# @brief Jobs started all at once
# @description One background job per item, with no limit.
set -euo pipefail

#######################################
# @description Compress every file at once
# @arg $@ string Files to compress
#######################################
function _compress_all {
  local file
  local -a pids=()
  for file in "$@"; do
    gzip -- "${file}" &
    pids+=("$!")
  done
  printf '%s\n' "${pids[@]}"
}

_compress_all "$@"
