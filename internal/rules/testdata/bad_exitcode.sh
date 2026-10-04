#!/usr/bin/env bash
# @file bad_exitcode.sh
# @brief Statuses the shell reserves or cannot carry
# @description Out of range or taken by the shell.
set -euo pipefail

#######################################
# @description Fail in every unhelpful way
# @arg $1 string Mode
#######################################
function _fail {
  case "$1" in
    range) exit 300 ;;
    negative) return -1 ;;
    missing) exit 127 ;;
    exec) return 126 ;;
    usage) return 2 ;;
    signal) exit 130 ;;
    *) exit 1 ;;
  esac
}

_fail "$@"
