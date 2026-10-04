# @file dollarzero.sh
# @brief A library that finds its own directory through $0
# @description When sourced, `$0` names the script that sourced it.

DOLLARZERO_DIR="$(dirname "$0")"
DOLLARZERO_SELF="$(dirname "${BASH_SOURCE[0]}")"

#######################################
# @description Print where the library lives
# @noargs
#######################################
function dollarzero::where {
  printf '%s %s %s\n' "${DOLLARZERO_DIR}" "${DOLLARZERO_SELF}" "${0%/*}"
  if [[ "${BASH_SOURCE[0]}" == "$0" ]]; then
    printf 'run directly\n'
  fi
}
