#!/usr/bin/env bash
# @file bad_curlfail.sh
# @brief Requests that take an error page for an answer
# @description Without --fail, curl exits 0 on any HTTP status.
set -euo pipefail

#######################################
# @description Fetch a document, every way
# @arg $1 string URL
#######################################
function _fetch {
  local url="$1" body code
  body="$(curl -sSL "${url}")"
  body="$(curl -fsSL "${url}")"
  body="$(curl --fail-with-body -sS "${url}")"
  code="$(curl -sS -o /dev/null -w '%{http_code}' "${url}")"
  printf '%s %s\n' "${body}" "${code}"
}
