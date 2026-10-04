#!/usr/bin/env bash
# @file bad_bashisms.sh
# @brief External commands for what Bash does itself
# @description Sequences, listings and field splits through a process.
set -euo pipefail

#######################################
# @description Walk files and fields
# @arg $1 string Upper bound
# @arg $2 string Colon-separated record
#######################################
function _walk {
  local count="$1" record="$2" index file user home
  for index in $(seq 1 "${count}"); do
    printf '%s\n' "${index}"
  done
  for ((index = 1; index <= count; index++)); do
    printf '%s\n' "${index}"
  done
  for file in $(ls); do
    printf '%s\n' "${file}"
  done
  ls -1 | sort
  ls -l -- "${HOME}"
  git ls-files | sort
  user="$(echo "${record}" | cut -d: -f1)"
  home="$(printf '%s\n' "${record}" | awk -F: '{print $6}')"
  user="$(cut -d : -f 1 <<< "${record}")"
  printf '%s\n' "${record}" | awk '{ total += $1 } END { print total }'
  printf '%s\n' "${record}" | cut -c1-4
  IFS=: read -r user _ _ _ _ home _ <<< "${record}"
  printf '%s %s\n' "${user}" "${home}"
}

_walk "$@"
