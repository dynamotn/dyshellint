#!/usr/bin/env bash
# @file bad_waiting.sh
# @brief Waits with no deadline
# @description Each loop sleeps until a condition that may never hold.
set -euo pipefail

#######################################
# @description Wait for a service in several unbounded ways
# @arg $1 string Socket path
# @arg $2 string Lock path
#######################################
function _wait_for_service {
  local socket="$1" lock="$2"
  until [[ -S "${socket}" ]]; do
    sleep 1
  done
  while [[ -e "${lock}" ]]; do
    printf 'waiting for %s\n' "${lock}" >&2
    sleep 5
  done
}

_wait_for_service "$@"
