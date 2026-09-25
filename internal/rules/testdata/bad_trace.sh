# A note that has no business being here
#!/usr/bin/env bash
# @file bad_trace.sh
# @brief Tracing that the caller cannot turn off
# @description The shebang is not the first line, and the trace is inline.
# shellcheck source=/dev/null
. ./lib/dybatpho/init.sh
dybatpho::register_common_handlers

#######################################
# @description Trace one command
# @noargs
#######################################
function _traced {
  set -x
  printf 'traced\n'
}

_traced
