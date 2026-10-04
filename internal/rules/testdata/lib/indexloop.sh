# @file indexloop.sh
# @brief Index loops over arrays that may have gaps
# @description An array a caller passed by name can be sparse, and counting up
#   to its length then reads indexes that do not exist.

#######################################
# @description Join a caller's array by counting its indexes
# @arg $1 string Name of the array
#######################################
function indexloop::join {
  local -n __indexloop_join_ref="$1"
  local __indexloop_join_i
  for ((__indexloop_join_i = 0; __indexloop_join_i < ${#__indexloop_join_ref[@]}; __indexloop_join_i++)); do
    printf '%s\n' "${__indexloop_join_ref[__indexloop_join_i]}"
  done
}

#######################################
# @description Count through an array the function built itself
# @noargs
#######################################
function indexloop::own {
  local -a __indexloop_own_items=(a b c)
  local __indexloop_own_i
  for ((__indexloop_own_i = 0; __indexloop_own_i < ${#__indexloop_own_items[@]}; __indexloop_own_i++)); do
    printf '%s\n' "${__indexloop_own_items[__indexloop_own_i]}"
  done
}
