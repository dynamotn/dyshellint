# @file arithinput.sh
# @brief A private helper trusts what its public caller checked
# @description Only what the helper reads itself counts as input there.

#######################################
# @description Scale a number its public caller already checked
# @arg $1 number Number to scale
#######################################
function __arithinput_scale {
  local line
  printf '%s\n' "$(($1 * 2))"
  read -r line < /dev/null || true
  printf '%s\n' "$((line + 1))"
}
