#!/usr/bin/env bash
# @file bad_dryrun.sh
# @brief A script that offers a dry run and then ignores it
# @description `DRY_RUN=true` promises nothing changes.
set -euo pipefail

#######################################
# @description Install and clean up without looking at DRY_RUN
# @arg $1 string Directory
#######################################
function _install {
  local dir="$1"
  rm -f -- "${dir}/old"
  ln -s -- "${dir}/new" "${dir}/current"
  systemctl restart app
  systemctl status app
  local staged="${dir}/.new"
  rm -f -- "${staged}"
}

#######################################
# @description Install through the dry-run helper
# @arg $1 string Directory
#######################################
function _install_safely {
  local dir="$1"
  if [[ "${DRY_RUN:-false}" == true ]]; then
    printf 'would remove %s\n' "${dir}/old"
    return 0
  fi
  rm -f -- "${dir}/old"
}
