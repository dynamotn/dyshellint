#!/usr/bin/env bash
# @file bad_curltimeout.sh
# @brief Requests with no time limit
# @description curl waits as long as the server does.
set -euo pipefail

#######################################
# @description Fetch a document, with and without a limit
# @arg $1 string URL
#######################################
function _fetch {
  local url="$1"
  local -a options=(--fail -sS --max-time 30)
  curl --fail -sS "${url}"
  curl --fail -sS --max-time 30 "${url}"
  curl --fail -sSm 30 "${url}"
  curl --fail -sS --max-time=30 "${url}"
  curl "${options[@]}" "${url}"
  curl --config "${HOME}/.curlrc" "${url}"
  timeout 30 curl --fail -sS "${url}"
  curl --version
}

_fetch "$@"
