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

#######################################
# @description Remove the old copy, called only where DRY_RUN is decided
# @arg $1 string Directory
#######################################
function _apply {
  rm -f -- "$1/old"
}

#######################################
# @description Remove the old copy, called once without a guard
# @arg $1 string Directory
#######################################
function _apply_unguarded {
  rm -f -- "$1/unguarded"
}

#######################################
# @description Remove the old copy, called by a guarded helper
# @arg $1 string Directory
#######################################
function _apply_nested {
  rm -f -- "$1/nested"
}

#######################################
# @description Call the helper the guard already decided on
# @arg $1 string Directory
#######################################
function _apply_all {
  _apply_nested "$1"
}

#######################################
# @description Decide the dry run, then call the helpers
# @arg $1 string Directory
#######################################
function _main {
  if [[ "${DRY_RUN:-false}" == true ]]; then
    printf 'would remove %s\n' "$1/old"
  else
    _apply "$1"
    _apply_unguarded "$1"
  fi
  [[ "${DRY_RUN:-false}" == true ]] || _apply_all "$1"
  _apply_unguarded "$1"
}
