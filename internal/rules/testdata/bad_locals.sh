#!/usr/bin/env bash
# @file bad_locals.sh
# @brief Variables a function sets without declaring them local
# @description The loop variable, the arithmetic counter and the targets of
#   read, mapfile, printf -v and getopts all leak out of the function.
set -euo pipefail

#######################################
# @description Set variables without declaring them
# @noargs
#######################################
function _leaky {
  for name in alpha beta; do
    printf '%s\n' "${name}"
  done
  ((total += 1))
  read -r first rest <<< "a b c"
  read -r -a parts <<< "x y"
  mapfile -t lines < /dev/null
  printf -v stamp '%s' "now"
  getopts "ab" option || true
  printf '%s %s %s %s %s %s %s\n' "${total}" "${first}" "${rest}" "${parts[0]}" "${#lines[@]}" "${stamp}" "${option}"
}

#######################################
# @description Set variables that are declared, shared on purpose or documented
# @noargs
# @set result
#######################################
function _tidy {
  local name count=0 line
  for name in alpha beta; do
    ((count++))
  done
  read -r line <<< "a"
  declare -g shared
  read -r shared <<< "b"
  read -r SETTING <<< "c"
  printf -v result '%s' "${name}${line}${count}"
}

