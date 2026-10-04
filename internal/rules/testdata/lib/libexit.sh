# @file libexit.sh
# @brief A library that ends the script that sourced it
# @description `exit` in a sourced function ends the caller too.

#######################################
# @description Give up on bad input by exiting
# @arg $1 string Value to check
#######################################
function libexit::check {
  [[ -n "$1" ]] || exit 1
  (exit 0)
  return 0
}

#######################################
# @description Stop the script, which is this function's job
# @arg $1 string Message
#######################################
function libexit::die {
  printf '%s\n' "$1" >&2
  exit 1
}

#######################################
# @description Clean up when the script is interrupted
# @noargs
#######################################
function libexit::on_term {
  exit 143
}

#######################################
# @description Install the handler
# @noargs
#######################################
function libexit::install {
  trap 'libexit::on_term' TERM
}
