#!/usr/bin/env bash
# @file bad_jobs.sh
# @brief Handlers and waits that lose a status
# @description An EXIT handler with a fixed status, and waits that stop at the first failure.
set -euo pipefail

#######################################
# @description Clean up and report success, whatever happened
# @noargs
#######################################
function _finish {
  rm -f -- "${TEMP_FILE:?}"
  exit 0
}

#######################################
# @description Clean up and keep the status
# @noargs
#######################################
function _finish_right {
  local status=$?
  rm -f -- "${TEMP_FILE:?}"
  exit "${status}"
}

#######################################
# @description Run jobs and wait for them
# @noargs
#######################################
function _main {
  local -a pids=()
  local pid failed=0
  TEMP_FILE="$(mktemp)"
  trap 'rm -f -- "${TEMP_FILE}"; exit 0' EXIT
  trap _finish EXIT
  trap _finish_right EXIT
  trap 'exit 130' INT
  sleep 1 &
  pids+=("$!")
  for pid in "${pids[@]}"; do
    wait "${pid}"
  done
  for pid in "${pids[@]}"; do
    wait "${pid}" || failed=$((failed + 1))
  done
  wait
  printf '%s\n' "${failed}"
}

_main
