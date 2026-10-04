#!/usr/bin/env bash
# @file bad_deprecated.sh
# @brief Commands current systems no longer ship
# @description Each one has a maintained replacement.
set -euo pipefail

#######################################
# @description Inspect the machine with deprecated tools
# @noargs
#######################################
function _inspect {
  which git
  egrep 'a|b' /etc/hosts
  netstat -tln
  command -v git
  grep -E 'a|b' /etc/hosts
}
