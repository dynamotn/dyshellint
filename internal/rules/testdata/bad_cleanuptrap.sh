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
  trap 'stty sane' INT
  trap 'tput cnorm >&2' TERM
  trap 'stty sane; trap - INT; kill -s INT "${BASHPID}"' INT
  trap _cleanup_then_bail INT
  trap '_remove_and_exit "${SCRATCH}"' TERM
  trap '_cleanup "${SCRATCH}"' INT
  trap '_cleanup; trap - INT; kill -s INT "$$"' INT
  trap "_cleanup; exit 130" TERM
}

#######################################
# @description Stop with the status of an interrupt
# @noargs
#######################################
function _bail {
  exit 130
}

#######################################
# @description Remove the scratch directory, then stop through a helper
# @noargs
#######################################
function _cleanup_then_bail {
  rm -rf -- "${SCRATCH:?}"
  _bail
}

#######################################
# @description Remove the directory it is given and stop
# @arg $1 string Directory to remove
#######################################
function _remove_and_exit {
  rm -rf -- "${1:?}"
  exit 130
}

_main
