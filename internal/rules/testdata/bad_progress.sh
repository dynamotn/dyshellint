#!/usr/bin/env bash
# @file bad_progress.sh
# @brief Progress drawn whatever the output is
# @description In a log, the redraws pile up as control characters.
set -euo pipefail

#######################################
# @description Count to a total, redrawing the line
# @arg $1 number Total
#######################################
function _count {
  local total="$1" i
  for ((i = 1; i <= total; i++)); do
    printf '\r%d/%d' "${i}" "${total}" >&2
    printf '\033[K%s' "step ${i}" >&2
    printf '%s%d' $'\e[2K' "${i}" >&2
  done
}

_count "$@"
