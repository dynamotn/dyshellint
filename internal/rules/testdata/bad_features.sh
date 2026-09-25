#!/usr/bin/env bash
# @file bad_features.sh
# @brief Constructs the Features and Bugs chapter rejects
# @description eval, a pipe into while, a loop over command output, braces on a
#   positional parameter, a status read after the fact and a handmade temporary.
set -euo pipefail

#######################################
# @description Every rejected construct in one place
# @arg $1 string Root directory to walk
#######################################
function _walk {
  local root="${1}"
  eval "echo ${root}"

  local count=0
  find "${root}" -type f | while read -r file; do
    count=$((count + 1))
    printf '%s\n' "${file}"
  done

  for entry in $(ls "${root}"); do
    printf '%s\n' "${entry}"
  done

  grep -q pattern "${root}/file"
  if [[ $? -ne 0 ]]; then
    printf 'no match\n'
  fi

  printf 'data\n' > "/tmp/walk.$$"
}
