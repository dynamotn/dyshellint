#!/usr/bin/env bash
# @file bad_endofoptions.sh
# @brief Operands that may start with a dash
# @description A file named `-rf` or a pattern `-v` is read as an option.
set -euo pipefail

#######################################
# @description Handle a user-given path and pattern, every way
# @arg $1 string Path
# @arg $2 string Pattern
#######################################
function _handle {
  local path="$1" pattern="$2" mode="644"
  rm -f "${path}"
  grep "${pattern}" "${path}"
  chmod "${mode}" "${path}"
  rm -f -- "${path}"
  grep -e "${pattern}" -- "${path}"
  cat "./${path}"
  cp "/srv/${path}" "${path}.bak"
  mkdir -m 700 -- "${path}"
}
