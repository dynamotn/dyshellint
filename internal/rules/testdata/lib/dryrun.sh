# @file dryrun.sh
# @brief A library whose public function decides the dry run
# @description Its private helper changes files for a caller that checked
#   DRY_RUN already.

#######################################
# @description Publish a file unless DRY_RUN says not to
# @arg $1 string File
#######################################
function dryrun::publish {
  [[ "${DRY_RUN:-false}" == true ]] && return 0
  __dryrun_move "$1"
}

#######################################
# @description Move a file into place
# @arg $1 string File
#######################################
function __dryrun_move {
  mv -- "$1" "$1.done"
}
