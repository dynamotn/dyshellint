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

#######################################
# @description Draw a box at the place its caller computed, and at one read in
# @arg $1 string Rectangle, `x y width height`
#######################################
function __arithinput_box {
  local rect="$1" x y width height left right
  read -r x y width height <<< "${rect}"
  printf '%s\n' "$((x + width)) $((y + height))"
  read -r left right <<< "$(cat -- "${BOX_FILE:?}")"
  printf '%s\n' "$((left + right))"
}
