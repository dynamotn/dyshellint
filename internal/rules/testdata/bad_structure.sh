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
