#!/usr/bin/env bash
# @file good_dryrun_dybatpho.sh
# @brief A helper run through the dry-run helper of dybatpho
# @description dybatpho::dry_run prints the command instead under DRY_RUN.
# shellcheck source=/dev/null
. "${DYBATPHO_DIR}/init.sh" || exit 1
dybatpho::register_common_handlers

#######################################
# @description Remove the old copy
# @arg $1 string Directory
#######################################
function _remove_old {
  local dir
  dybatpho::expect_args dir -- "$@"
  rm -f -- "${dir}/old"
}

#######################################
# @description Remove the old copy, unless DRY_RUN says not to
# @arg $1 string Directory
#######################################
function _main {
  local dir
  dybatpho::expect_args dir -- "$@"
  dybatpho::dry_run _remove_old "${dir}"
}

_main "$@"
