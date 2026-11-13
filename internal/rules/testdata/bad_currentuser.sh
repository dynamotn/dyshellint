#!/usr/bin/env bash
# @file bad_currentuser.sh
# @brief Root checks and home paths built by hand
# @description USER can be set to anything, and homes are not always in /home.
set -euo pipefail

#######################################
# @description Check for root in every fragile way
# @arg $1 string Name of another user
#######################################
function _check {
  local user="$1"
  [[ "${USER:-}" == "root" ]] || return 1
  [[ "$(whoami)" != root ]] && return 1
  [ "$(id -u)" -eq 0 ] || return 1
  test "${LOGNAME:-}" = root || return 1
  (($(id -u) == 0)) || return 1
  [[ 0 == "$(id -u)" ]] || return 1
  [[ "$(id -un)" == "root" ]] || return 1
  printf '%s\n' "/home/${user}/.config"
  printf '%s\n' "/Users/$(id -un)"
  local dir=/home/${user}
  printf '%s\n' "${dir}"
}

_check "$@"
