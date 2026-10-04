#!/usr/bin/env bash
# @file bad_portability.sh
# @brief Options only the GNU tools understand
# @description Each call below fails on macOS, the BSDs or BusyBox.
set -euo pipefail

#######################################
# @description Use GNU options without checking for them
# @arg $1 string File to inspect
#######################################
function _inspect {
  local file="$1"
  date -d "yesterday" +%F
  sed -i 's/a/b/' "${file}"
  readlink -f "${file}"
  stat -c %s "${file}"
  grep -oP '\d+' "${file}"
  sort -rV "${file}"
  sed -n 's/a/b/p' "${file}"
  grep -o 'x' "${file}"
}

#######################################
# @description Pick the stat option after probing which stat this is
# @arg $1 string File to inspect
#######################################
function _probed_size {
  local file="$1"
  if stat --version > /dev/null 2>&1; then
    stat -c %s "${file}"
  else
    stat -f %z "${file}"
  fi
}
