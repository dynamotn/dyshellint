#!/usr/bin/env bash
# @file good_fanout.sh
# @brief Jobs started a few at a time
# @description The loop waits for a job before it starts another.
if ((BASH_VERSINFO[0] < 4 || (BASH_VERSINFO[0] == 4 && BASH_VERSINFO[1] < 4))); then
  printf 'This script needs Bash 4.4 or newer, found %s\n' "${BASH_VERSION}" >&2
  exit 1
fi
set -euo pipefail

#######################################
# @description Compress files, four at a time
# @arg $@ string Files to compress
#######################################
function _compress_all {
  local file failed=0 running=0
  for file in "$@"; do
    if ((running >= 4)); then
      wait -n || failed=$((failed + 1))
      running=$((running - 1))
    fi
    gzip -- "${file}" &
    running=$((running + 1))
  done
  while ((running > 0)); do
    wait -n || failed=$((failed + 1))
    running=$((running - 1))
  done
  ((failed == 0))
}

_compress_all "$@"
