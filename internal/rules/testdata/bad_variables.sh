#!/usr/bin/env bash
# @file bad_variables.sh
# @brief Variable declarations the guide rejects
# @description Each function below breaks one rule of Variable Names.
set -euo pipefail

#######################################
# @description Assign from a command substitution while declaring
# @noargs
#######################################
function _inline_assignment {
  local version="$(printf '1.2.3\n')"
  printf '%s\n' "${version}"
}

#######################################
# @description Assign a global by accident and walk it with a vague name
# @noargs
#######################################
function _accidental_global {
  version="1.2.3"
  local tools=()
  tools+=("${version}")
  for i in "${tools[@]}"; do
    printf '%s\n' "${i}"
  done
}
