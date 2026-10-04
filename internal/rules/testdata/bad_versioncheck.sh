#!/usr/bin/env bash
# @file bad_versioncheck.sh
# @brief Bash version checks that refuse newer versions
# @description Compound and text comparisons.
set -euo pipefail

((BASH_VERSINFO[0] >= 5 && BASH_VERSINFO[1] >= 2)) || exit 1
[[ "${BASH_VERSION}" > "5.2" ]] || exit 1
((BASH_VERSINFO[0] > 5 || (BASH_VERSINFO[0] == 5 && BASH_VERSINFO[1] >= 2))) || exit 1
