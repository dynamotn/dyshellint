#!/usr/bin/env bash
# @file bad_cleanuptrap.sh
# @brief Cleanup handlers that keep the script running
# @description A handler on INT or TERM replaces the exit.
set -euo pipefail

#######################################
# @description Remove the scratch directory
# @noargs
#######################################
function _cleanup {
  rm -rf -- "${SCRATCH:?}"
}

#######################################
# @description Remove the scratch directory and stop
# @noargs
#######################################
function _cleanup_and_exit {
  rm -rf -- "${SCRATCH:?}"
  exit 130
}

#######################################
# @description Install the handlers, right and wrong
# @noargs
#######################################
function _main {
  SCRATCH="$(mktemp -d)"
  local caught=""
  trap 'rm -rf -- "${SCRATCH}"' EXIT INT TERM
  trap _cleanup INT
  trap _cleanup EXIT
  trap _cleanup_and_exit INT
  trap 'rm -f -- "${SCRATCH}/lock"; exit 143' TERM
  trap '' INT
  trap 'caught=INT' INT
}

_main
