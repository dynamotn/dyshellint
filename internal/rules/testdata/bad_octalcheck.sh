#!/usr/bin/env bash
# @file bad_octalcheck.sh
# @brief Numbers checked for digits and then read as octal
# @description A digits-only check lets `08` and `010` through, and arithmetic
#   reads a leading zero as octal.
set -euo pipefail

minute="$(date +%M)"
[[ "${minute}" =~ ^[0-9]+$ ]] || exit 1
if ((minute >= 30)); then
  printf 'second half\n'
fi

#######################################
# @description Wait the number of seconds it is given, read as octal
# @arg $1 number Seconds to wait
#######################################
function _wait_octal {
  local seconds="$1" retries="${RETRIES:-3}"
  [[ "${seconds}" =~ ^[0-9]+$ ]] || return 1
  case "${retries}" in
    '' | *[!0-9]*) return 1 ;;
    *) ;;
  esac
  sleep "$((seconds * 2))"
  ((retries > 0)) || return 1
  ((seconds > 1)) || return 1
}

#######################################
# @description Wait the number of seconds it is given, read in base ten
# @arg $1 number Seconds to wait
# @arg $2 number Factor, without a leading zero
#######################################
function _wait_decimal {
  local seconds="$1" scale="$2" retries="${RETRIES:-3}"
  [[ "${seconds}" =~ ^[0-9]{1,18}$ ]] || return 1
  [[ "${scale}" =~ ^(0|[1-9][0-9]*)$ ]] || return 1
  case "${retries}" in
    '' | *[!0-9]*) return 1 ;;
    *) ;;
  esac
  retries=$((10#${retries}))
  sleep "$((10#${seconds} * scale))"
  ((retries > 0)) || return 1
}

#######################################
# @description Read a number in base ten into the variable it names
# @arg $1 string Name of the variable
#######################################
function _decimal_into {
  printf -v "$1" '%s' "$((10#${!1}))"
}

#######################################
# @description Read counts from the input, one normalised by a helper
# @noargs
#######################################
function _read_counts {
  local count total
  read -r count total < "${COUNTS_FILE:?}"
  [[ "${count}" =~ ^[0-9]+$ ]] || return 1
  [[ "${total}" =~ ^[0-9]+$ ]] || return 1
  _decimal_into count
  ((count > 0 && total > 0)) || return 1
}

#######################################
# @description Step a counter used before its check, then normalised in place
# @arg $1 number Start of the count
#######################################
function _step {
  local step="$1"
  printf '%s\n' "$((step + 1))"
  [[ "${step}" =~ ^[0-9]{1,18}$ ]] || return 1
  ((step = 10#${step}))
  ((step > 0)) || return 1
}
