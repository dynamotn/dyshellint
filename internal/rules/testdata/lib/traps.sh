# @file traps.sh
# @brief Libraries that touch the caller's signal handlers
# @description A library runs inside someone else's script, whose cleanup is
#   already on EXIT, INT and TERM.

#######################################
# @description Install a cleanup that replaces the caller's
# @noargs
#######################################
function traps::careless {
  trap 'rm -f /tmp/x' EXIT
  trap - INT TERM
  trap 'printf debug' DEBUG
}

#######################################
# @description Install a handler for the call and put the caller's back
# @noargs
#######################################
function traps::careful {
  local saved
  saved="$(trap -p INT)"
  trap 'printf stop' INT
  trap - INT
  eval "${saved}"
}

#######################################
# @description Run a job in the background with handlers of its own
# @noargs
#######################################
function traps::background {
  (
    trap 'exit 143' TERM
    sleep 1
  ) &
}
