# @file libcd.sh
# @brief A library that changes the caller's directory
# @description The working directory belongs to the script that sourced it.

#######################################
# @description Build in a directory and stay there
# @arg $1 string Directory
#######################################
function libcd::build {
  cd "$1" || return 1
  make
}

#######################################
# @description Build in a subshell, which leaves the caller where it was
# @arg $1 string Directory
#######################################
function libcd::build_isolated {
  (cd "$1" && make)
  local top
  top="$(cd "$1" && pwd)"
  printf '%s\n' "${top}"
}

#######################################
# @description Build and go back
# @arg $1 string Directory
#######################################
function libcd::build_and_return {
  local before="${PWD}"
  cd "$1" || return 1
  make
  cd "${before}" || return 1
}
