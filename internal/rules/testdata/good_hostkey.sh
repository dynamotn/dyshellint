#!/usr/bin/env bash
# @file good_hostkey.sh
# @brief Host keys pinned in a known_hosts file shipped with the script
# @description `StrictHostKeyChecking=yes` refuses a host whose key is not in
#   the file, and `accept-new` trusts only a first contact.
set -euo pipefail

#######################################
# @description Reach a host whose key is pinned
# @arg $1 string Host
#######################################
function _deploy {
  local host="$1"
  local -a ssh_opts=(-o UserKnownHostsFile=./known_hosts -o StrictHostKeyChecking=yes)
  timeout 60 ssh "${ssh_opts[@]}" -o BatchMode=yes -- "${host}" true
  timeout 60 ssh -o StrictHostKeyChecking=accept-new -o BatchMode=yes -- "${host}" true
  ssh-keygen -lf ./known_hosts
}

_deploy "$@"
