#!/usr/bin/env bash
# @file bad_exportfunc.sh
# @brief Functions exported to child processes
# @description Every child process inherits them, without what they need.
set -euo pipefail

#######################################
# @description Compress one file
# @arg $1 string File to compress
#######################################
function _worker {
  gzip -- "$1"
}

#######################################
# @description Compress every log, four at a time
# @noargs
#######################################
function _main {
  export -f _worker
  declare -fx _worker
  typeset -xf _worker
  find . -name '*.log' -print0 | xargs -0 -P4 -n1 bash -c '_worker "$1"' _
}

_main
