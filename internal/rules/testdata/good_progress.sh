#!/usr/bin/env bash
# @file good_progress.sh
# @brief Progress drawn only on a terminal
# @description Without a terminal, one line per step instead.
set -euo pipefail

#######################################
# @description Count to a total, redrawing the line on a terminal
# @arg $1 number Total
#######################################
function _count {
  local total="$1" i
  for ((i = 1; i <= total; i++)); do
    if [[ -t 2 ]]; then
      printf '\r\033[K%d/%d' "${i}" "${total}" >&2
    else
      printf '%d/%d\n' "${i}" "${total}" >&2
    fi
  done
}

_count "$@"
