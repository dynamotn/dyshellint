#!/usr/bin/env bash
set -euo pipefail

# TODO(someone): explain why this exists
function _undocumented {
  printf '%s\n' "$1"
}

#######################################
# @description Take an argument without saying so
#######################################
function _partly_documented {
  printf '%s\n' "$1"
}
