#!/usr/bin/env bash
# @file bad_pathsafety.sh
# @brief PATH values that let a planted command win
# @description Current directory, empty elements and /tmp.
set -euo pipefail

PATH=".:${PATH}"
PATH=":${PATH}"
PATH="${PATH}:"
PATH="/tmp/tools:${PATH}"
PATH="${PATH}::/usr/bin"
PATH="${HOME}/.local/bin${PATH:+:${PATH}}"
PATH=/usr/local/bin:/usr/bin:/bin
export PATH
