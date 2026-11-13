#!/usr/bin/env bash
# @file good_menu.sh
# @brief A menu only on a terminal, an environment variable otherwise
# @description `DEPLOY_ENV` gives the same answer when no one is there.
set -euo pipefail

#######################################
# @description Pick an environment
# @noargs
#######################################
function _pick {
  local env="${DEPLOY_ENV-}"
  if [[ -z "${env}" ]]; then
    [[ -t 0 ]] || {
      printf 'Set DEPLOY_ENV when there is no terminal\n' >&2
      return 1
    }
    env="$(printf '%s\n' dev prod | fzf)"
  fi
  printf '%s\n' "${env}"
}

_pick
