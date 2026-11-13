#!/usr/bin/env bash
# @file bad_ps4.sh
# @brief Trace prompts that hide the line
# @description A trace line that does not say where it comes from.
set -euo pipefail

PS4='+ '
export PS4='+ ${FUNCNAME[0]:-main}: '
printf '%s\n' "${PS4}"
