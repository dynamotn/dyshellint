#!/usr/bin/env bash
# @file bad_docker-entrypoint.sh
# @brief A container entrypoint that keeps Bash in front of the command
# @description The command never receives SIGTERM.
set -euo pipefail

mkdir -p /data
"$@"
