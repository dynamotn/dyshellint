#!/usr/bin/env bash
# @file bad_hostkey.sh
# @brief Host key checking turned off in each spelling ssh reads
# @description Any of these lets a machine in the middle pose as the host.
set -euo pipefail

#######################################
# @description Reach a host without checking its key
# @arg $1 string Host
#######################################
function _deploy {
  local host="$1"
  local -a ssh_opts=(-o StrictHostKeyChecking=no -o ConnectTimeout=10)
  timeout 60 ssh "${ssh_opts[@]}" -- "${host}" true
  timeout 60 ssh -oStrictHostKeyChecking=no -o BatchMode=yes -- "${host}" true
  timeout 60 scp -o 'StrictHostKeyChecking No' ./app.tar "${host}:/srv/"
  timeout 60 sftp -o UserKnownHostsFile=/dev/null "${host}" < /dev/null
  rsync -a -e "ssh -o StrictHostKeyChecking=off" ./dist/ "${host}:/srv/"
  GIT_SSH_COMMAND="ssh -o UserKnownHostsFile=/dev/null" git fetch origin
  export RSYNC_RSH='ssh -o GlobalKnownHostsFile=/dev/null'
  local out
  out="$(timeout 60 ssh -o StrictHostKeyChecking=no -- "${host}" hostname)"
  printf '%s\n' "${out}"
  cat > ./ssh_config << CONFIG
Host *
  StrictHostKeyChecking no
CONFIG
}

_deploy "$@"
