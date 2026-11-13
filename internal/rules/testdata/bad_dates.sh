#!/usr/bin/env bash
# @file bad_dates.sh
# @brief Clock reads that fork date
# @description Each one forks a process to read the time.
set -euo pipefail

#######################################
# @description Time a step with two forks of date
# @noargs
#######################################
function _time_step {
  local started finished now utc
  started="$(date +%s)"
  sleep 1
  finished="$(date '+%s')"
  now="$(date -u +%s)"
  utc="$(date --utc "+%s")"
  printf '%s %s %s\n' "$((finished - started))" "${now}" "${utc}"
}

_time_step
