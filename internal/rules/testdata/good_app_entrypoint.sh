#!/usr/bin/env bash
# @file good_app_entrypoint.sh
# @brief A container entrypoint that hands over to its command
# @description The command replaces Bash and receives the signals.
set -euo pipefail

#######################################
# @description Prepare the data folder and run the container's command
# @arg $@ string Command of the container
#######################################
function _main {
  mkdir -p /data
  if (($# == 0)); then
    set -- gunicorn --bind 0.0.0.0:8000
  fi
  exec "$@"
}

_main "$@"
