#!/usr/bin/env bash
# @file good_descriptors.sh
# @brief Descriptors picked by Bash and loops redirected once
# @description Each file is opened once.
set -euo pipefail

#######################################
# @description Write a report through a descriptor Bash picks
# @arg $1 string Report path
# @arg $2 string Output directory
#######################################
function _report {
  local report="$1" dir="$2"
  local host fd
  exec {fd}> "${report}"
  printf 'start\n' >&"${fd}"
  exec {fd}>&-
  for host in alpha beta; do
    printf '%s\n' "${host}"
  done >> "${report}"
  for host in alpha beta; do
    printf '%s\n' "${host}" >> "${dir}/${host}.log"
    printf '%s\n' "${host}" >> /dev/stderr
  done
  local -a seen=()
  for host in alpha beta; do
    mapfile -t seen < "${report}"
    printf '%s %d\n' "${host}" "${#seen[@]}" >> "${report}"
  done
  {
    printf 'one\n'
    printf 'two\n'
  } >> "${report}"
  { _swap 3>&1 1>&2 2>&3; } 2> /dev/null
}

#######################################
# @description Print to both streams
# @noargs
#######################################
function _swap {
  printf 'out\n'
  printf 'err\n' >&2
}

_report "$@"
