#!/usr/bin/env bash
# @file bad_text.sh
# @brief Formatting the guide rejects
# @description A tab, a trailing space and a line over the limit.
set -euo pipefail

#######################################
# @description Print a long line
# @noargs
#######################################
function _print {
	printf "%s\\n" "a" 
  printf "%s\\n" "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
}

_print
