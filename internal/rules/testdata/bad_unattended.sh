#!/usr/bin/env bash
# @file bad_unattended.sh
# @brief Scripts that load a login profile
# @description A profile is written for a terminal, not for cron or CI.
set -euo pipefail

# shellcheck source=/dev/null
. ~/.bashrc
# shellcheck source=/dev/null
source "${HOME}/.profile"
# shellcheck source=/dev/null
. "$HOME/.bash_profile"
# shellcheck source=/dev/null
source /etc/profile
# shellcheck source=/dev/null
. /etc/bash.bashrc
