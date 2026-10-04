#!/usr/bin/env bash
# @file bad_retry.sh
# @brief Network calls retried and waited on badly
# @description No pause between attempts, and ssh with no timeout.
set -euo pipefail

#######################################
# @description Fetch, retrying in every way
# @arg $1 string URL
#######################################
function _fetch {
  local url="$1" attempt=1
  until curl --fail -sS --max-time 30 "${url}"; do :; done
  while ! curl --fail -sS --max-time 30 "${url}"; do
    ((attempt < 5)) || return 1
    sleep $((2 ** attempt))
    attempt=$((attempt + 1))
  done
  ssh host 'uptime'
  ssh -o ConnectTimeout=10 -o BatchMode=yes host 'uptime'
  timeout 60 ssh host 'uptime'
  scp -- file host:
}

_fetch "$@"
