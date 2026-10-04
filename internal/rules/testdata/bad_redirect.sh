#!/usr/bin/env bash
# @file bad_redirect.sh
# @brief Both streams sent to a file the long way
# @description And the order that leaves errors on the terminal.
set -euo pipefail

#######################################
# @description Run and log
# @arg $1 string Log file
#######################################
function _run {
  local log="$1"
  date > "${log}" 2>&1
  date >> "${log}" 2>&1
  date 2>&1 > "${log}"
  date &> "${log}"
  date 2>&1 | tee -- "${log}"
}

_run "$@"
