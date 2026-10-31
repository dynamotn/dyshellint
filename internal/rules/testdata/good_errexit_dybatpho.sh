#!/usr/bin/env bash
# @file good_errexit_dybatpho.sh
# @brief Command substitutions in a script that loads dybatpho
# @description dybatpho turns on inherit_errexit along with the strict mode.
# shellcheck source=/dev/null
. "${DYBATPHO_DIR}/init.sh" || exit 1
dybatpho::register_common_handlers

#######################################
# @description Read the version of the project
# @noargs
# @stdout Version
#######################################
function _version {
  git describe --tags
  printf 'unknown\n'
}

#######################################
# @description Read the version, or stop the script when there is none
# @noargs
# @stdout Version
#######################################
function _required_version {
  local version
  version="$(_version)"
  [[ -n "${version}" ]] || dybatpho::die "no version"
  printf '%s\n' "${version}"
}

#######################################
# @description Print the version
# @noargs
#######################################
function _main {
  local version required
  version="$(_version)"
  # The strict mode dybatpho turns on stops the script on this failure.
  required="$(_required_version)"
  dybatpho::info "${version} ${required}"
}

_main
