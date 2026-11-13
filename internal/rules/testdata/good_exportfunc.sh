#!/usr/bin/env bash
# @file good_exportfunc.sh
# @brief The script runs itself for each file
# @description A subcommand gives the child the whole script.
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
# @arg $@ string Subcommand and its arguments, when the script calls itself
#######################################
function _main {
  if [[ "${1-}" == "_worker" ]]; then
    _worker "$2"
    return
  fi
  declare -f _worker > /dev/null
  find . -name '*.log' -print0 | xargs -0 -P4 -n1 "$0" _worker
}

_main "$@"
