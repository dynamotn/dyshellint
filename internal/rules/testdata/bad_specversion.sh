#!/usr/bin/env bash
# @file bad_specversion.sh
# @brief An entrypoint whose command line cannot say which build it is
# @description The root spec has `--help` but no `--version`; the spec of a
#   function answers to its caller and needs none.
# shellcheck source=/dev/null
. ./lib/dybatpho/init.sh --modules cli
dybatpho::register_common_handlers

#######################################
# @description Spec of bad_specversion.sh
# @noargs
#######################################
function _spec_main {
  dybatpho::opts::setup "Deploy the app" MAIN_ARGS action:"_main"
  dybatpho::opts::disp "Show help" --help action:"dybatpho::generate_help _spec_main"
}

#######################################
# @description Spec of the arguments of _greet
# @noargs
#######################################
function _spec_greet {
  dybatpho::opts::setup "Greet" GREET_ARGS action:"printf 'hi\n'"
  dybatpho::opts::disp "Show help" --help action:"dybatpho::generate_help _spec_greet"
}

#######################################
# @description Greet with options of its own
# @arg $@ string Options
#######################################
function _greet {
  dybatpho::generate_from_spec _spec_greet "$@"
}

#######################################
# @description Deploy
# @noargs
#######################################
function _main {
  _greet
}

dybatpho::generate_from_spec _spec_main "$@"
