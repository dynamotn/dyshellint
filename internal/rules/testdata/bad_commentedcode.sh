#!/usr/bin/env bash
# @file bad_commentedcode.sh
# @brief Code switched off with comments
# @description Old code left behind as comments, next to prose that only looks like it.
#   tar -czf "${backup}" -- "${dir}"
set -euo pipefail

#######################################
# @description Copy a directory
# @arg $1 string Directory to copy
# @example
#   _copy "${HOME}/src" && ls -- "${HOME}/src.bak"
#######################################
function _copy {
  local dir
  dir="${1:?}"
  # shellcheck disable=SC2034 # kept for the callers that read it
  local copied=0
  # cp -a -- "${dir}" "${dir}.old"
  # if [[ -d "${dir}.bak" ]]; then
  #   rm -rf -- "${dir}.bak"
  # fi
  cp -a -- "${dir}" "${dir}.bak"

  # rsync is faster, but not installed everywhere; run it as:
  #   rsync -a -- "${dir}/" "${dir}.bak/"
  # TODO: Bring this back once every host has rsync 3.
  # rsync -a -- "${dir}/" "${dir}.bak/"

  # Copy with `cp -a` so the modes and times stay; a plain cp | tar loses them
  # here and $dir is never empty.
  # The copy is done; log::info "done" would say so, but this script stays quiet
  # and the caller checks $? && moves on
  printf 'copied %s\n' "${dir}"

  # log::info "copied ${dir} to ${dir}.bak"
  # copied=$((copied + 1))
}

_copy "$@"
