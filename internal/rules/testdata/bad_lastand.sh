#!/usr/bin/env bash
# @file bad_lastand.sh
# @brief Functions that end on a test
# @description A false test becomes the status.
set -euo pipefail

#######################################
# @description Log when verbose, the wrong way
# @arg $1 string Verbose flag
#######################################
function _log_wrong {
  local verbose="$1"
  [[ "${verbose}" == true ]] && printf 'done\n'
}

#######################################
# @description Log when verbose, the right way
# @arg $1 string Verbose flag
#######################################
function _log_right {
  local verbose="$1"
  if [[ "${verbose}" == true ]]; then
    printf 'done\n'
  fi
}

#######################################
# @description Fail on purpose when the copy fails
# @arg $1 string Source
#######################################
function _copy {
  local source="$1"
  mkdir -p -- out && cp -- "${source}" out/
}

#######################################
# @description Silence a whole block, and one command
# @arg $1 string Directory
#######################################
function _tidy {
  local dir="$1"
  {
    rm -f -- "${dir}/a"
    rmdir -- "${dir}"
  } 2> /dev/null
  rmdir -- "${dir}" 2> /dev/null || true # still in use by another job
  (
    set -C
    : > "${dir}/lock"
  ) 2> /dev/null || true # taken by another run
  for dir in "$@"; do
    rmdir -- "${dir}"
  done &> /dev/null
}

_log_wrong "$@"
_log_right "$@"
_copy "$@"
_tidy "$@"
(($# > 0)) && printf 'args\n'
