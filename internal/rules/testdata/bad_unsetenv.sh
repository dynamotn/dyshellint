#!/usr/bin/env bash
# @file bad_unsetenv.sh
# @brief Environment variables read bare under set -u
# @description A variable nothing in the run sets may be missing, and `set -u`
#   then stops the script on its first bare read.
set -euo pipefail

: "${LOG_DIR:=/var/log}"
readonly CONFIG_PATH="/etc/app.conf"

#######################################
# @description Read settings from the environment
# @noargs
#######################################
function _settings {
  printf '%s\n' "${EDITOR}" "${#XDG_CONFIG_HOME}"
  printf '%s\n' "${PAGER:-less}" "${TERM-}" "${LOG_DIR}" "${CONFIG_PATH}"
  printf '%s\n' "${HOME}" "${BASH_SOURCE[0]}" "${EDITOR}" "${BATS_TEST_TMPDIR}"
  widget::draw "${WIDGET_STYLE_DIM}"
}
