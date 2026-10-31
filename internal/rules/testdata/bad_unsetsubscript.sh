#!/usr/bin/env bash
# @file bad_unsetsubscript.sh
# @brief Array elements removed with a subscript expanded twice
# @description `unset` evaluates the subscript it is given, so a key that the
#   shell already expanded is expanded again.
set -euo pipefail

declare -A TIMERS=([build]=1)
declare -a QUEUE=(a b)
printf '%s\n' "${TIMERS[@]}" "${QUEUE[@]}"

#######################################
# @description Forget a timer and a queue entry, expanding the key twice
# @arg $1 string Timer name
# @arg $2 number Queue index
#######################################
function _forget_twice {
  local name="$1" index="$2"
  unset "TIMERS[${name}]"
  unset -v "QUEUE[${index}]"
  unset "TIMERS[$(printf '%s' "${name}")]"
}

#######################################
# @description Forget a timer and a queue entry, expanding the key once
# @arg $1 string Timer name
# @arg $2 number Queue index
#######################################
function _forget_once {
  local name="$1" index="$2"
  unset -v 'TIMERS[$name]'
  unset -v 'QUEUE[index]' 'QUEUE[-1]'
  unset -v "${name}"
  unset -v name
}
