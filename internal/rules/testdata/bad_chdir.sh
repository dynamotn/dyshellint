#!/usr/bin/env bash
# @file bad_chdir.sh
# @brief The directory stack, and a way back through OLDPWD
# @description pushd and popd print the stack; `cd -` prints the directory.
set -euo pipefail

#######################################
# @description Build in a directory with the stack
# @arg $1 string Directory
#######################################
function _build {
  local dir
  dir="${1:?}"
  pushd "${dir}" > /dev/null
  make
  popd > /dev/null
}

#######################################
# @description Build in a directory and come back through OLDPWD
# @arg $1 string Directory
#######################################
function _build_back {
  local dir
  dir="${1:?}"
  cd -P -- "${dir}"
  make
  cd - > /dev/null
}

#######################################
# @description Go back to where the caller came from
# @noargs
#######################################
function _back {
  popd > /dev/null
}

#######################################
# @description Build in a subshell, or point the tool at the directory
# @arg $1 string Directory
#######################################
function _build_good {
  local dir
  dir="${1:?}"
  (cd -P -- "${dir}" && make)
  make -C "${dir}"
  git -C "${dir}" status
}

_build "$@" && _build_back "$@" && _back && _build_good "$@"
