#!/usr/bin/env bash
# @file bad_app_entrypoint.sh
# @brief A container entrypoint whose main function runs the server in front of Bash
# @description The server never receives SIGTERM.
set -euo pipefail

#######################################
# @description Prepare the data folder and start the server
# @arg $@ string Arguments of the server
#######################################
function _main {
  mkdir -p /data
  gunicorn --bind 0.0.0.0:8000 "$@"
}

_main "$@"
