#!/usr/bin/env bash
# @file bad_assocarray.sh
# @brief String keys on arrays never declared as maps
# @description Each key lands on index 0.
set -euo pipefail

#######################################
# @description Record versions, with and without -A
# @arg $1 string Tool
#######################################
function _versions {
  local tool="$1" index=0
  local -A known=()
  local -a list=() versions=()
  known["jq"]="1.7.1"
  versions["jq"]="1.7.1"
  versions['yq']="2"
  versions["${tool}"]="3"
  list["${index}"]="first again"
  local -n alias="${tool}"
  alias["jq"]="4"
  list[index]="first"
  list[0]="zero"
  printf '%s\n' "${known[@]}" "${list[@]}"
}

_versions "$@"
