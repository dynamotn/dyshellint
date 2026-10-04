#!/usr/bin/env bash
# @file bad_parsels.sh
# @brief The output of ls read as data
# @description A glob hands each name to the script as it is.
set -euo pipefail

#######################################
# @description Walk the files of a directory
# @arg $1 string Directory
#######################################
function _walk {
  local dir="$1" file
  for file in $(ls -- "${dir}"); do
    printf '%s\n' "${file}"
  done
  ls -1 -- "${dir}" | sort
  ls -l -- "${dir}"
  git ls-files | sort
  for file in "${dir}"/*; do
    [[ -e "${file}" ]] || continue
    printf '%s\n' "${file}"
  done
}

_walk "$@"
