# @file archive.sh
# @brief Unpack archives
# @description A library with a sibling test folder, but no test of its own.

#######################################
# @description Unpack one archive
# @arg $1 string Path of the archive
# @exitcode 0 If the archive is unpacked
#######################################
function archive::unpack {
  local path
  dybatpho::expect_args path -- "$@"
  dybatpho::dry_run tar -xf "${path}"
}
