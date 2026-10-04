# @file libstate.sh
# @brief A library that changes the caller's shell state
# @description Options, IFS and umask belong to the script that sourced it.

#######################################
# @description Change the shell state and leave it changed
# @noargs
#######################################
function libstate::careless {
  set +e
  shopt -s nullglob
  IFS=,
  umask 077
  set -- a b
}

#######################################
# @description Change the shell state for this call only
# @noargs
#######################################
function libstate::careful {
  local -
  local IFS=, a b
  set +e
  IFS=: read -r a b <<< "x:y"
  (umask 077 && : > /dev/null)
  shopt -q nullglob || true
  printf '%s %s\n' "${a}" "${b}"
}

#######################################
# @description Turn strict mode off for a step and back on after it
# @noargs
#######################################
function libstate::lenient {
  local old_ifs="${IFS}"
  set +e
  IFS=,
  false
  IFS="${old_ifs}"
  set -e
}
