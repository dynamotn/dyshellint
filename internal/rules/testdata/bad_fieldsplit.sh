#!/usr/bin/env bash
# @file bad_fieldsplit.sh
# @brief Fields of a string taken by a process
# @description read splits a record into named fields in the shell.
set -euo pipefail

#######################################
# @description Take a record apart
# @arg $1 string Colon-separated record
#######################################
function _fields {
  local record="$1" user home prefix
  user="$(printf '%s\n' "${record}" | cut -d: -f1)"
  home="$(printf '%s\n' "${record}" | awk -F: '{print $6}')"
  user="$(cut -d : -f 1 <<< "${record}")"
  printf '%s\n' "${record}" | awk '{ total += $1 } END { print total }'
  prefix="$(printf '%s' "${record}" | cut -c1-4)"
  IFS=: read -r user _ _ _ _ home _ <<< "${record}"
  printf '%s %s %s\n' "${user}" "${home}" "${prefix}"
}

_fields "$@"
