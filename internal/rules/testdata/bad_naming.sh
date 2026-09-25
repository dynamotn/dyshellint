#!/usr/bin/env bash
# @file bad_naming.sh
# @brief Function declarations the guide rejects
# @description Each declaration below breaks one rule of Function Names.
set -euo pipefail

#######################################
# @description Declared without the keyword
# @noargs
#######################################
verify_sha256() {
  printf 'no keyword\n'
}

#######################################
# @description Declared with redundant parentheses
# @noargs
#######################################
function _verify_sha512() {
  printf 'redundant parens\n'
}

#######################################
# @description Declared in camelCase
# @noargs
#######################################
function _verifySha1 {
  printf 'camelCase\n'
}

#######################################
# @description Public name in an entrypoint script
# @noargs
#######################################
function download_archive {
  printf 'no privacy prefix\n'
}
