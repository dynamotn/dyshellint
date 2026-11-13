#!/usr/bin/env bash
# @file good_unattended.sh
# @brief A script that sets its own environment
# @description PATH is set in the script, and only its own files are sourced.
set -euo pipefail

PATH=/usr/local/bin:/usr/bin:/bin
export PATH
# shellcheck source=/dev/null
. /etc/default/backup
printf '%s\n' ~/.bashrc
