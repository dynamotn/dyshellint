#!/usr/bin/env bash
# @file good_download.sh
# @brief A download checked against a published digest before it runs
# @description `sha256sum -c` fails on a mismatch, and errexit stops the script.
set -euo pipefail

#######################################
# @description Install a tool from the network
# @arg $1 string URL of the installer
# @arg $2 string Expected SHA-256 digest
#######################################
function _install {
  local url="$1" digest="$2" installer=./install.sh
  curl --fail -sSL --max-time 60 "${url}" -o "${installer}"
  printf '%s  %s\n' "${digest}" "${installer}" | sha256sum -c -
  bash "${installer}"
}

_install "$@"
