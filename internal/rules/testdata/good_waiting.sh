#!/usr/bin/env bash
# @file good_waiting.sh
# @brief Waits that give up
# @description Each loop stops at a deadline, after a number of tries, or with its input.
set -euo pipefail

#######################################
# @description Wait for a service, within limits
# @arg $1 string Socket path
# @arg $2 string Lock path
# @arg $3 number Process id
#######################################
function _wait_for_service {
  local socket="$1" lock="$2" pid="$3"
  local deadline=$((SECONDS + 60)) tries=0 line
  until [[ -S "${socket}" ]]; do
    ((SECONDS < deadline)) || return 1
    sleep 1
  done
  while [[ -e "${lock}" ]]; do
    tries=$((tries + 1))
    [[ "${tries}" -le 10 ]] || return 1
    sleep 5
  done
  while [[ ! -e "${lock}" ]]; do
    if [[ -S "${socket}" ]]; then
      break
    fi
    sleep 1
  done
  while kill -0 "${pid}" 2> /dev/null; do
    sleep 1
  done
  while IFS= read -r line || [[ -n "${line}" ]]; do
    printf '%s\n' "${line}"
    sleep 1
  done < "${lock}"
  while true; do
    printf 'tick\n'
    sleep 60
  done
}

_wait_for_service "$@"
