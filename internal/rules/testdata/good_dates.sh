#!/usr/bin/env bash
# @file good_dates.sh
# @brief Clock reads that Bash answers itself
# @description Only formatting and parsing still call date.
set -euo pipefail

#######################################
# @description Time a step without forking
# @arg $1 string Path whose time is printed
#######################################
function _time_step {
  local path="$1"
  local started=${SECONDS} now stamp mtime
  sleep 1
  printf -v now '%(%s)T' -1
  stamp="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
  mtime="$(date -r "${path}" +%s)"
  printf '%s %s %s %s\n' "$((SECONDS - started))" "${now}" "${stamp}" "${mtime}"
}

_time_step "$0"
