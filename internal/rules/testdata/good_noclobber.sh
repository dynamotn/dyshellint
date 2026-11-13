#!/usr/bin/env bash
# @file good_noclobber.sh
# @brief A guessable path written under noclobber
# @description `set -o noclobber` at the top level holds in every function, so
#   `>` refuses a path planted in advance.
set -euo pipefail
set -o noclobber

#######################################
# @description Write the lock file of this process
# @noargs
#######################################
function _lock {
  printf '%s\n' "$$" > "/var/lock/app.$$"
}

_lock
