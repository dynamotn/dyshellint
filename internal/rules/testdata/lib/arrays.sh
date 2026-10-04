# @file arrays.sh
# @brief Arrays that may still be empty where they are expanded
# @description Bash 4.3 stops a `set -u` script on `"${a[@]}"` of an empty
#   array, and dybatpho supports Bash 4.3.

#######################################
# @description Collect the arguments that look like options, then print them
# @arg $@ string Arguments to filter
#######################################
function arrays::options {
  local -a options=()
  local argument
  for argument in "$@"; do
    [[ "${argument}" == -* ]] && options+=("${argument}")
  done
  dybatpho::info "options: ${options[*]}"
  printf '%s\n' "${options[@]}"
  printf '%s\n' ${options[@]+"${options[@]}"}
  printf '%s\n' "${#options[@]}"
}

#######################################
# @description Copy the arguments, which may be none, then print them
# @arg $@ string Arguments to copy
#######################################
function arrays::copy {
  local -a copied=()
  copied=("$@")
  printf '%s\n' "${copied[@]}"
}

#######################################
# @description Start from a fixed element, so the array is never empty
# @noargs
#######################################
function arrays::seeded {
  local -a flags=()
  flags=(--quiet)
  printf '%s\n' "${flags[@]}"
  local -a modes=(fast)
  printf '%s\n' "${modes[@]}"
}

#######################################
# @description Fall back to a default when the copy came out empty
# @arg $@ string Arguments to copy
#######################################
function arrays::defaulted {
  local -a targets=()
  targets=("$@")
  if ((${#targets[@]} == 0)); then
    targets=(.)
  fi
  printf '%s\n' "${targets[@]}"
}
