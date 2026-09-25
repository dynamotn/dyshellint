#!/usr/bin/env bash
# @file good_entrypoint.sh
# @brief Show the shape the guide recommends for an entrypoint
# @description Constants, then declarations, then the one call that starts the
#   script. Nothing in between runs when the file is sourced.
readonly SCRIPT_DIR="$(realpath "$(dirname "${BASH_SOURCE[0]}")")"
# shellcheck source=/dev/null
. "${SCRIPT_DIR}/lib/dybatpho/init.sh" --modules cli
dybatpho::register_common_handlers

#######################################
# @description Spec of good_entrypoint.sh
# @noargs
#######################################
function _spec_main {
  dybatpho::opts::setup "Install the tools of a list" MAIN_ARGS action:"_main"
  dybatpho::opts::param "Log level" LOG_LEVEL --log-level init:="info" \
    validate:"dybatpho::validate_log_level \$OPTARG"
  dybatpho::opts::disp "Show help" --help action:"dybatpho::generate_help _spec_main"
}

#######################################
# @description Install every tool the caller asked for
# @arg $1 string Name of the list to install
# @stderr Progress messages
# @exitcode 0 If every tool is installed
#######################################
function _main {
  local list
  dybatpho::expect_args list -- "$@"

  local -a tools=()
  readarray -t tools < <(printf '%s\n' "${list}")
  for tool in "${tools[@]}"; do
    dybatpho::info "Installing ${tool}"
  done
}

dybatpho::generate_from_spec _spec_main "$@"
