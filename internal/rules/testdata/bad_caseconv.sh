#!/usr/bin/env bash
# @file bad_caseconv.sh
# @brief Case changed by hand, and case-blind matching left on
# @description `tr` on one value, and `nocasematch` that outlives its comparison.
set -euo pipefail

shopt -s nocaseglob

#######################################
# @description Normalise a name
# @arg $1 string Name
# @stdout The name in lower case
#######################################
function _lower {
  local name lower upper
  name="${1:?}"
  lower="$(printf '%s' "${name}" | tr '[:upper:]' '[:lower:]')"
  upper="$(echo "${name}" | tr a-z A-Z)"
  lower="$(tr '[A-Z]' '[a-z]' <<< "${name}")"
  tr '[:upper:]' '[:lower:]' < "${name}"
  printf '%s\n' "${lower}" "${upper}" | tr -d '[:upper:]'
}

#######################################
# @description Match yes in any case
# @arg $1 string Answer
#######################################
function _is_yes {
  local answer
  answer="${1:?}"
  shopt -s nocasematch
  [[ "${answer}" == yes ]]
}

#######################################
# @description Match no in any case, and turn the option off again
# @arg $1 string Answer
#######################################
function _is_no {
  local answer status=0
  answer="${1:?}"
  shopt -s nocasematch
  [[ "${answer}" == no ]] || status=1
  shopt -u nocasematch
  return "${status}"
}

#######################################
# @description Match maybe in any case: `local -` does not restore shopt
# @arg $1 string Answer
#######################################
function _is_maybe {
  local -
  local answer
  answer="${1:?}"
  shopt -s nocasematch
  [[ "${answer}" == maybe ]]
}

#######################################
# @description Match never in any case, and put the option back as it was
# @arg $1 string Answer
#######################################
function _is_never {
  local answer restore status=0
  answer="${1:?}"
  restore="$(shopt -p nocasematch)" || true
  shopt -s nocasematch
  [[ "${answer}" == never ]] || status=1
  eval "${restore}"
  return "${status}"
}

#######################################
# @description Match later in any case, in a subshell
# @arg $1 string Answer
#######################################
function _is_later {
  (
    shopt -s nocasematch
    [[ "${1:?}" == later ]]
  )
}

_lower "$@" && _is_yes "$@" && _is_no "$@" && _is_maybe "$@" && _is_never "$@" && _is_later "$@"
