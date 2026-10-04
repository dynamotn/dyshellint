# @file callercode.sh
# @brief Functions that run code their caller passed
# @description The code a caller passes runs inside the function, where it sees
#   every plain local declared before it and can change it.

#######################################
# @description Run a command and report its status, with plain locals
# @arg $@ string Command to run
# @exitcode 0 Always
#######################################
function callercode::run {
  local status=0
  "$@" || status=$?
  local done_at="now"
  printf '%s %s\n' "${status}" "${done_at}"
}

#######################################
# @description Hand the command on to the runner, with a plain local
# @arg $@ string Command to run
#######################################
function callercode::twice {
  local round
  for round in 1 2; do
    callercode::run "$@"
  done
}

#######################################
# @description Run a handler named by a variable, with every local prefixed
# @arg $1 string Handler to call
#######################################
function callercode::notify {
  local __callercode_notify_handler="$1"
  "${__callercode_notify_handler}" "ready"
}

#######################################
# @description Evaluate code the caller wrote, with a plain local
# @arg $1 string Code to evaluate
#######################################
function callercode::evaluate {
  local code="$1"
  eval "${code}"
}

#######################################
# @description Print a value, which runs nothing the caller passed
# @arg $1 string Value to print
#######################################
function callercode::show {
  local value="$1"
  printf '%s\n' "${value}"
}

#######################################
# @description Run a command line it builds itself, which no caller passed
# @noargs
#######################################
function callercode::own_command {
  local -a command=(printf '%s\n' ok)
  "${command[@]}"
}
