#!/usr/bin/env bash
# @file bad_temppath.sh
# @brief Scratch files at paths another user can guess
# @description A path built from the process id or $RANDOM can be planted in
#   advance, and `>`, `touch` and `mkdir -p` all go along with what is there.
set -euo pipefail

#######################################
# @description Make scratch files, every way
# @arg $1 string Directory to work in
#######################################
function _scratch {
  local dir="$1"
  printf 'x\n' > "${dir}/.work.$$"
  touch "${TMPDIR:-/var/tmp}/lock.${BASHPID}"
  mkdir -p "${dir}/run.${RANDOM}"
  mkdir "${dir}/run.${RANDOM}"
  printf 'x\n' > "${dir}/plain.out"
}

#######################################
# @description Make a scratch file exclusively
# @arg $1 string Directory to work in
#######################################
function _exclusive {
  local dir="$1"
  set -C
  printf 'x\n' > "${dir}/.work.$$"
}
