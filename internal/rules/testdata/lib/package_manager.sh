# @file package_manager.sh
# @brief Install packages with the system package manager
# @description Every function here is namespaced after the file, so a caller
#   can tell where it comes from without looking it up.

#######################################
# @description Install one package
# @arg $1 string Name of the package
# @exitcode 0 If the package is installed
#######################################
function package_manager::install {
  local package
  dybatpho::expect_args package -- "$@"

  local -a options=()
  options+=(--noconfirm)
  dybatpho::dry_run pacman -S "${options[@]}" "${package}"
}

#######################################
# @description Print the options every call shares
# @noargs
# @stdout One option per line
#######################################
function __package_manager_common_options {
  printf '%s\n' "--noconfirm"
}
