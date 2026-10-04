#!/usr/bin/env bash
# @file bad_interactive.sh
# @brief Prompts that hang when nobody is there
# @description In CI, cron or a pipe, `read` waits forever.
set -euo pipefail

#######################################
# @description Ask for confirmation without looking for a terminal
# @noargs
#######################################
function _proceed {
  local answer
  read -r -p "Continue? " answer
  [[ "${answer}" == y ]]
}

#######################################
# @description Ask only when someone can answer
# @noargs
#######################################
function _proceed_safely {
  local answer
  [[ -t 0 ]] || return 1
  read -r -p "Continue? " answer
  [[ "${answer}" == y ]]
}

#######################################
# @description Read a file and a timed answer, which never hang
# @arg $1 string File
#######################################
function _read_input {
  local line answer
  while read -r line || [[ -n "${line}" ]]; do
    printf '%s\n' "${line}"
  done < "$1"
  read -r -t 5 answer || true
  printf '%s\n' "${answer}"
}

#######################################
# @description Ask a question, which is this function's whole job
# @noargs
#######################################
function _prompt {
  local answer
  read -r answer
  read -rsn1 answer
  printf '%s\n' "${answer}"
}

#######################################
# @description Read with options kept in an array
# @noargs
#######################################
function _proceed_with_options {
  local answer
  local -a options=(-r -t 5)
  read "${options[@]}" answer || true
  printf '%s\n' "${answer}"
}
