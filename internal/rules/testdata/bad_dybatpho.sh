#!/usr/bin/env bash
# @file bad_dybatpho.sh
# @brief A dybatpho script that ignores the library it sources
# @description No handler is installed, the command line is parsed by hand, the
#   spec has no help, and the temporary file cleans up after nobody.
# shellcheck source=/dev/null
. ./lib/dybatpho/init.sh --modules cli

#######################################
# @description Spec without a help option
# @noargs
#######################################
function _spec_main {
  dybatpho::opts::setup "Do something" MAIN_ARGS action:"_main"
}

#######################################
# @description Read the arguments by hand
# @arg $1 string Name of the tool
#######################################
function _main {
  local name="$1"
  local temp_file
  temp_file="$(mktemp)"
  dybatpho::info "${name} ${temp_file}"
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --name) shift ;;
    *) shift ;;
  esac
done
