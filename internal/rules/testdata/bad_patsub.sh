#!/usr/bin/env bash
# @file bad_patsub.sh
# @brief Replacements that Bash 5.2 reads as the matched text
# @description With `patsub_replacement`, an unquoted `&` in the replacement of
#   `${var//pattern/replacement}` stands for the match.
set -euo pipefail

#######################################
# @description Mask a secret, putting it back through `&`
# @arg $1 string Text to mask
# @arg $2 string Secret
# @arg $3 string Placeholder
#######################################
function _mask_unquoted {
  local text="$1" secret="$2" mask="$3"
  printf '%s\n' "${text//"${secret}"/${mask}}"
  printf '%s\n' "${text//"${secret}"/$(printf '%s' "${mask}")}"
  printf '%s\n' "${text//and/&}"
}

#######################################
# @description Mask a secret with a quoted replacement
# @arg $1 string Text to mask
# @arg $2 string Secret
# @arg $3 string Placeholder
#######################################
function _mask_quoted {
  local text="$1" secret="$2" mask="$3"
  printf '%s\n' "${text//"${secret}"/"${mask}"}"
  printf '%s\n' "${text//and/\&}"
  printf '%s\n' "${text//$'\n'/ }"
  printf '%s\n' "${text//$'\n'/$'\\&.'}"
  printf '%s\n' "${text//[[:space:]]/_}"
}
