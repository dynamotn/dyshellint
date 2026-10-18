#!/bin/bash
# @file bad_structure.sh
# @brief A layout the guide rejects
# @description Wrong shebang, no strict mode, `source` instead of `.`, and code
#   that runs between two declarations.
source ./lib/functions.sh

#######################################
# @description First declaration
# @noargs
#######################################
function _first {
  echo "Error: something went wrong"
  echo "error: x" >/dev/stderr
  { echo "error: x"; } >&2
  echo "Error: ${1}"
  printf 'error: %s\n' "${1}"
  echo -e "warning: ${1}"
  printf 'done: %s\n' "error"
}

rm -rf ./cache

#######################################
# @description Second declaration
# @noargs
#######################################
function _second {
  sudo tee /etc/hosts
}

_first
