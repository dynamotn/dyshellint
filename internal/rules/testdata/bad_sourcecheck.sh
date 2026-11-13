#!/usr/bin/env bash
# @file bad_sourcecheck.sh
# @brief An entrypoint that sources libraries it never found
# @description A failed `.` prints an error and the script runs on.
set -euo pipefail

LIB_DIR="${XDG_DATA_HOME:-${HOME}/.local/share}/tool/lib"
OWN_DIR="$(dirname "${BASH_SOURCE[0]}")"
CHECKED_DIR="${LIB_DIR}"
# shellcheck source=/dev/null
. "${LIB_DIR}/common.sh"
[[ -r "${CHECKED_DIR}/extra.sh" ]] || exit 1
# shellcheck source=/dev/null
. "${CHECKED_DIR}/extra.sh"
# shellcheck source=/dev/null
. "${LIB_DIR}/more.sh" || exit 1
# shellcheck source=/dev/null
. "${OWN_DIR}/lib/common.sh"
ZERO_DIR="$(cd "$(dirname "${0}")" && pwd)"
# shellcheck source=/dev/null
. "${ZERO_DIR}/lib/util.sh"
SCRIPT_DIR="$(dirname "${BASH_SOURCE[0]}")"
NEXT_DIR="${SCRIPT_DIR}/../lib"
# shellcheck source=/dev/null
. "${NEXT_DIR}/util.sh"
