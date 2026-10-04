#!/usr/bin/env bash
# @file bad_procsubst.sh
# @brief Functions that can fail, feeding a process substitution
# @description Nothing reads the status of `<(...)`, so a producer that fails
#   looks exactly like one that had nothing to say.
set -euo pipefail

#######################################
# @description List the entries of a directory, failing when it is missing
# @arg $1 string Directory to list
# @stdout One entry per line
#######################################
function _entries {
  [[ -d "$1" ]] || return 1
  printf '%s\n' "$1"/*
}

#######################################
# @description Print a fixed list, which cannot fail
# @noargs
# @stdout One name per line
#######################################
function _names {
  printf '%s\n' alpha beta
}

#######################################
# @description Read each list through a process substitution
# @arg $1 string Directory to list
#######################################
function _main {
  local -a entries=() names=() found=()
  mapfile -t entries < <(_entries "$1")
  mapfile -t names < <(_names)
  mapfile -t found < <(_entries "$1" | sort)
  printf '%s %s %s\n' "${#entries[@]}" "${#names[@]}" "${#found[@]}"
}

_main "$@"
