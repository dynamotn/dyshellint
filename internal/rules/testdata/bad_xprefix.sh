#!/usr/bin/env bash
# @file bad_xprefix.sh
# @brief A letter put before both sides of a comparison
# @description `[[ ... ]]` never reads a value as an operator.
set -euo pipefail

#######################################
# @description Tell whether the user agreed
# @arg $1 string Answer
#######################################
function _agreed {
  local answer="$1" arch
  arch="$(uname -m)"
  if [[ "x${answer}" == "xyes" ]]; then
    printf '%s\n' "agreed"
  fi
  [[ "X$answer" != "Xno" ]] || return 1
  if [ "x${answer}" = "xmaybe" ]; then
    printf '%s\n' "undecided"
  fi
  [[ "${arch}" == "x86_64" ]] || printf '%s\n' "not x86_64"
  [[ "x${arch}" == "y86_64" ]] || printf '%s\n' "no prefix on both sides"
  [[ "xen" == "xen" ]] || printf '%s\n' "no variable"
  [[ "${answer}" == "yes" ]]
}

_agreed "$@"
