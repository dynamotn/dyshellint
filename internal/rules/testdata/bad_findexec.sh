#!/usr/bin/env bash
# @file bad_findexec.sh
# @brief find -exec that forks per file
# @description One process per match.
set -euo pipefail

#######################################
# @description Compress logs
# @arg $1 string Root
#######################################
function _compress {
  local root="$1"
  find "${root}" -name '*.log' -exec gzip -- {} \;
  find "${root}" -name '*.log' -exec gzip -- {} +
  find "${root}" -name '*.bak' -exec mv -- {} {}.old \;
  find "${root}" -type d -exec chmod 0755 {} ';'
}

_compress "$@"
