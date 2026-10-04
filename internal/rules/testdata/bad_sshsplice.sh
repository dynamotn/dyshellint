#!/usr/bin/env bash
# @file bad_sshsplice.sh
# @brief Local values that reach the remote shell of ssh unquoted
# @description ssh joins its command words, and the remote shell splits them.
set -euo pipefail

#######################################
# @description Print the disk usage of a directory
# @arg $1 string Directory
#######################################
function report::disk_usage {
  local dir="$1"
  du -sh -- "${dir}"
}

#######################################
# @description Report a directory of a host
# @arg $1 string Host
# @arg $2 string Directory
#######################################
function _report {
  local host="$1" dir="$2" quoted_dir
  local -a ssh_opts=(-o ConnectTimeout=10 -o BatchMode=yes)
  printf -v quoted_dir '%q' "${dir}"
  timeout 300 ssh -o ConnectTimeout=10 -o BatchMode=yes -- "${host}" du -sh -- "${dir}"
  timeout 300 ssh -o ConnectTimeout=10 -qo BatchMode=yes "${host}" "du -sh -- ${dir}"
  timeout 300 ssh -o ConnectTimeout=10 -p22 -t "${host}" ls -- "${dir}"/
  timeout 300 ssh -o ConnectTimeout=10 "${host}" "du -sh -- ${dir@Q}"
  timeout 300 ssh -o ConnectTimeout=10 "${host}" du -sh -- "${dir@Q}"
  timeout 300 ssh -o ConnectTimeout=10 "${host}" "du -sh -- ${quoted_dir}"
  timeout 300 ssh -o ConnectTimeout=10 "${host}" 'du -sh -- "${HOME}"'
  timeout 300 ssh "${ssh_opts[@]}" "${host}" du -sh -- "${dir}"
  {
    declare -f report::disk_usage
    printf 'report::disk_usage %s\n' "${dir@Q}"
  } | timeout 300 ssh -o ConnectTimeout=10 -o BatchMode=yes -- "${host}" bash -s
}

_report "$@"
