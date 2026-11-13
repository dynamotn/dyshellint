#!/usr/bin/env bash
# @file bad_funcsize.sh
# @brief One function that does every step
# @description 51 lines of code are one too many; blank lines and comments do not count.
set -euo pipefail

#######################################
# @description Print fifty-one steps
# @noargs
#######################################
function _long {
  printf 'step 1\n'
  printf 'step 2\n'
  printf 'step 3\n'
  printf 'step 4\n'
  printf 'step 5\n'
  printf 'step 6\n'
  printf 'step 7\n'
  printf 'step 8\n'
  printf 'step 9\n'
  printf 'step 10\n'
  printf 'step 11\n'
  printf 'step 12\n'
  printf 'step 13\n'
  printf 'step 14\n'
  printf 'step 15\n'
  printf 'step 16\n'
  printf 'step 17\n'
  printf 'step 18\n'
  printf 'step 19\n'
  printf 'step 20\n'
  printf 'step 21\n'
  printf 'step 22\n'
  printf 'step 23\n'
  printf 'step 24\n'
  printf 'step 25\n'
  printf 'step 26\n'
  printf 'step 27\n'
  printf 'step 28\n'
  printf 'step 29\n'
  printf 'step 30\n'
  printf 'step 31\n'
  printf 'step 32\n'
  printf 'step 33\n'
  printf 'step 34\n'
  printf 'step 35\n'
  printf 'step 36\n'
  printf 'step 37\n'
  printf 'step 38\n'
  printf 'step 39\n'
  printf 'step 40\n'
  printf 'step 41\n'
  printf 'step 42\n'
  printf 'step 43\n'
  printf 'step 44\n'
  printf 'step 45\n'
  printf 'step 46\n'
  printf 'step 47\n'
  printf 'step 48\n'
  printf 'step 49\n'
  printf 'step 50\n'
  printf 'step 51\n'
}

#######################################
# @description Print fifty steps, with comments and blank lines between
# @noargs
#######################################
function _short {
  printf 'step 1\n'
  printf 'step 2\n'
  printf 'step 3\n'
  printf 'step 4\n'
  printf 'step 5\n'
  printf 'step 6\n'
  printf 'step 7\n'
  printf 'step 8\n'
  printf 'step 9\n'
  printf 'step 10\n'

  # The next ten steps
  printf 'step 11\n'
  printf 'step 12\n'
  printf 'step 13\n'
  printf 'step 14\n'
  printf 'step 15\n'
  printf 'step 16\n'
  printf 'step 17\n'
  printf 'step 18\n'
  printf 'step 19\n'
  printf 'step 20\n'

  # The next ten steps
  printf 'step 21\n'
  printf 'step 22\n'
  printf 'step 23\n'
  printf 'step 24\n'
  printf 'step 25\n'
  printf 'step 26\n'
  printf 'step 27\n'
  printf 'step 28\n'
  printf 'step 29\n'
  printf 'step 30\n'

  # The next ten steps
  printf 'step 31\n'
  printf 'step 32\n'
  printf 'step 33\n'
  printf 'step 34\n'
  printf 'step 35\n'
  printf 'step 36\n'
  printf 'step 37\n'
  printf 'step 38\n'
  printf 'step 39\n'
  printf 'step 40\n'

  # The next ten steps
  printf 'step 41\n'
  printf 'step 42\n'
  printf 'step 43\n'
  printf 'step 44\n'
  printf 'step 45\n'
  printf 'step 46\n'
  printf 'step 47\n'
  printf 'step 48\n'
  printf 'step 49\n'
  printf 'step 50\n'
}

_long && _short
