#!/usr/bin/env bash
# @file good_trace.sh
# @brief Trace prompts that name the line
# @description Every trace line says file, line and function.
set -euo pipefail

PS4='+ ${BASH_SOURCE[0]##*/}:${LINENO}:${FUNCNAME[0]:-main}: '
export PS4="${PS4}"'> '
printf '%s\n' "${PS4}"
