#!/usr/bin/env bash
# @file good_bashversion.sh
# @brief Bash 4.4 features behind a version check
# @description The check stops Bash 3.2 with a message.
if ((BASH_VERSINFO[0] < 4 || (BASH_VERSINFO[0] == 4 && BASH_VERSINFO[1] < 4))); then
  printf 'This script needs Bash 4.4 or newer, found %s\n' "${BASH_VERSION}" >&2
  exit 1
fi
set -euo pipefail

#######################################
# @description Split a list into the caller's array
# @arg $1 string Name of the array
#######################################
function _split_into {
  local -n __split_into_ref="$1"
  mapfile -d '' -t __split_into_ref < <(printf 'a\0b\0')
}

_split_into items
