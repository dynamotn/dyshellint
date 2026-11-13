#!/usr/bin/env bash
# @file good_completion.sh
# @brief A completion for a command that has no spec
# @description With no spec, there is nothing to generate the completion from.
set -euo pipefail

#######################################
# @description Complete the hosts of the inventory
# @noargs
#######################################
function _complete_hosts {
  mapfile -t COMPREPLY < <(compgen -W "web db" -- "${COMP_WORDS[COMP_CWORD]-}")
}

complete -F _complete_hosts ssh-inventory
