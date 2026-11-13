#!/usr/bin/env bash
# @file bad_lineendings.sh
# @brief A file saved with a byte order mark
# @description The kernel does not see the shebang behind it.
set -euo pipefail

printf 'hello\n'
