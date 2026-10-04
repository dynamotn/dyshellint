#!/usr/bin/env bash
# @file bad_flagcommand.sh
# @brief Flag variables run as commands
# @description Any value other than true or false runs.
set -euo pipefail

#######################################
# @description Act on flags
# @arg $1 string Force flag
#######################################
function _act {
  local force="$1" running=true check_cmd="true"
  if ${force}; then
    printf 'forced\n'
  fi
  while ${running}; do
    running=false
  done
  ${force} && printf 'again\n'
  if "${check_cmd}"; then
    printf 'checked\n'
  fi
  if [[ "${force}" == true ]]; then
    printf 'compared\n'
  fi
}

_act "$@"
