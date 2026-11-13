#!/usr/bin/env bash
# @file bad_completion.sh
# @brief A completion written by hand next to a spec
# @description The completion and the parser list the options twice.
set -euo pipefail

#######################################
# @description Spec of bad_completion.sh
# @noargs
#######################################
function _spec_main {
  dybatpho::opts::setup "Deploy" MAIN_ARGS action:"_main"
  dybatpho::opts::disp "Show help" --help -h action:"dybatpho::generate_help _spec_main"
}

#######################################
# @description Complete the options by hand
# @noargs
#######################################
function _complete_main {
  mapfile -t COMPREPLY < <(compgen -W "--help --env" -- "${COMP_WORDS[COMP_CWORD]-}")
}

complete -F _complete_main bad_completion.sh
complete -W "--help --env" deploy
complete -p deploy
