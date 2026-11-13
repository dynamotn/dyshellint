#!/usr/bin/env bash
# @file good_currentuser.sh
# @brief Root checks and home paths read from the system
# @description EUID and the password database are the reliable sources.
set -euo pipefail

#######################################
# @description Check for root and find a user's home
# @arg $1 string Name of another user
#######################################
function _check {
  local user="$1"
  ((EUID == 0)) || return 1
  [[ "${USER:-}" == "${user}" ]] && return 0
  # shellcheck disable=SC2312 # asks about another user, which BSG136 leaves alone
  [[ "$(id -u "${user}")" == 0 ]] && return 0
  [[ "${USER:-}" == "deploy" ]] && return 0
  local home
  home="$(getent passwd "${user}" | cut -d: -f6)"
  dscl . -read "/Users/${user}" NFSHomeDirectory
  printf '%s\n' "${home}/.config" "${HOME}/.cache" "/home/shared/${user}" "${HOME}/home/${user}"
}

_check "$@"
