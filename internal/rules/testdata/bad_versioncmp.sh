#!/usr/bin/env bash
# @file bad_versioncmp.sh
# @brief Versions compared as text or as numbers
# @description `1.10` sorts before `1.9` as text, and arithmetic stops at a dot.
set -euo pipefail

#######################################
# @description Check a tool's version, every way
# @arg $1 string Installed version
#######################################
function _check {
  local version="$1" count=3
  [[ "${version}" < "1.10" ]] && printf 'old\n'
  ((version > 2)) && printf 'new\n'
  ((BASH_VERSINFO[0] < 4)) && printf 'bash 3\n'
  [[ "${count}" -lt 4 ]] && printf 'few\n'
  [[ "${version}" == "1.10" ]] && printf 'exact\n'
}
