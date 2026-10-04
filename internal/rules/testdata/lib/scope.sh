# @file scope.sh
# @brief Functions whose locals a caller's variable names can reach
# @description Each public function writes to a variable its caller names. A
#   plain local is what the caller gets when its variable has the same name.

#######################################
# @description Fill a caller's variable through a nameref, with plain locals
# @arg $1 string Name of the variable to fill
# @arg $2 string Value to store
#######################################
function scope::fill_into {
  local -n ref="$1"
  local value="$2"
  ref="${value}"
}

#######################################
# @description Fill a caller's variable with printf -v, with a plain local
# @arg $1 string Name of the variable to fill
#######################################
function scope::stamp_into {
  local stamp="now"
  printf -v "$1" '%s' "${stamp}"
}

#######################################
# @description Read a line into a caller's variable, with a plain local
# @arg $1 string Name of the variable to fill
#######################################
function scope::line_into {
  local source="/dev/null"
  read -r "$1" < "${source}" || true
}

#######################################
# @description Fill a caller's variable, with every local prefixed
# @arg $1 string Name of the variable to fill
#######################################
function scope::fill_safely {
  local -n __scope_fill_safely_ref="$1"
  local __scope_fill_safely_value="ok"
  __scope_fill_safely_ref="${__scope_fill_safely_value}"
}

#######################################
# @description Bind a nameref to a fixed global, which no caller names
# @noargs
#######################################
function scope::fill_fixed {
  local -n target=SCOPE_RESULT
  local value="fixed"
  target="${value}"
}

#######################################
# @description A private helper, which only the library itself calls
# @arg $1 string Name of the variable to fill
#######################################
function __scope_fill {
  local -n ref="$1"
  local value="private"
  ref="${value}"
}

#######################################
# @description Set one of its own settings by name, which a case arm pins down
# @arg $1 string Setting name
# @arg $2 string Value
#######################################
function scope::set_setting {
  local name="$1" min="" max=""
  case "${name}" in
    min | max) printf -v "${name}" '%s' "$2" ;;
    *) return 1 ;;
  esac
  printf '%s %s\n' "${min}" "${max}"
}

#######################################
# @description Hand the caller's variable name on to a private writer
# @arg $1 string Name of the variable to fill
#######################################
function scope::fill_through {
  local name="$1" label="through"
  __scope_fill "${name}"
  printf '%s\n' "${label}"
}

#######################################
# @description Hand only a value to the private writer, under its own name
# @arg $1 string Value to print
#######################################
function scope::fill_value {
  local value="$1" result
  __scope_fill result
  printf '%s %s\n' "${value}" "${result}"
}

#######################################
# @description Hand the caller's name on before declaring any local
# @arg $1 string Name of the variable to fill
#######################################
function scope::fill_first {
  __scope_fill "$1"
  local after="late"
  printf '%s\n' "${after}"
}
