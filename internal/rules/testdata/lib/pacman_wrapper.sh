# @file pacman_wrapper.sh
# @brief Install packages with pacman
# @description The functions here read better under `pkg` than under the name
#   of the file, so the header declares the namespace they belong to.
# @namespace pkg

#######################################
# @description Install one package
# @arg $1 string Name of the package
# @exitcode 0 If the package is installed
#######################################
function pkg::install {
  local package
  dybatpho::expect_args package -- "$@"

  dybatpho::dry_run pacman -S --noconfirm "${package}"
}

#######################################
# @description Print the options every call shares
# @noargs
# @stdout One option per line
#######################################
function __pkg_common_options {
  printf '%s\n' "--noconfirm"
}
