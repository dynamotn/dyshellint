# @file runnable.sh
# @brief A library that carries the executable bit
# @description Everything here is correct except the file mode: a library is
#   sourced, never run.

#######################################
# @description Print the name of this library
# @noargs
# @stdout The name
#######################################
function runnable::name {
  printf '%s\n' "runnable"
}
