#!/usr/bin/env bash
# @file bad_menu.sh
# @brief Menus drawn with no look for a terminal
# @description Each waits, or fails with an unclear error, in CI or a pipe.
set -euo pipefail

#######################################
# @description Pick an environment
# @noargs
#######################################
function _pick {
  local env
  select env in dev staging prod; do
    break
  done
  printf '%s\n' "${env}"
  printf '%s\n' dev prod | fzf
  whiptail --menu "Environment" 15 40 2 dev "" prod ""
  dialog --yesno "Deploy?" 7 40
  gum choose dev prod
}

#######################################
# @description Tools that do not ask
# @noargs
#######################################
function _quiet {
  printf '%s\n' dev prod | fzf --filter dev
  gum style "Done"
}

_pick
_quiet
