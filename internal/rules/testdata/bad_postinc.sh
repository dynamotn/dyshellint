#!/usr/bin/env bash
# @file bad_postinc.sh
# @brief Post-increments that end a set -e script
# @description `((x++))` returns the value before the step, so it fails when
#   that value is 0.
set -euo pipefail

#######################################
# @description Count things, every way
# @noargs
#######################################
function _count {
  local count=0 total=0
  ((count++))
  ((total--, count++))
  ((count++)) || true
  if ((count++)); then total=1; fi
  ((++count))
  count=$((count + 1))
  printf '%s %s\n' "${count}" "${total}"
}
