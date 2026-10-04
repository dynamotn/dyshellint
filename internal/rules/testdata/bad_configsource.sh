#!/usr/bin/env bash
# @file bad_configsource.sh
# @brief Configuration files sourced as code
# @description Every line runs with the rights of the script.
set -euo pipefail

# shellcheck source=/dev/null # fixtures stand in for real files
. "${XDG_CONFIG_HOME:-${HOME}/.config}/app/config"
# shellcheck source=/dev/null # fixtures stand in for real files
. "${HOME}/.apprc"
# shellcheck source=/dev/null # fixtures stand in for real files
source "/tmp/app-${USER}.conf"
# shellcheck source=/dev/null # fixtures stand in for real files
. "${HOME}/lib/helpers.sh"
# shellcheck source=/dev/null # fixtures stand in for real files
. /etc/app/app.conf
