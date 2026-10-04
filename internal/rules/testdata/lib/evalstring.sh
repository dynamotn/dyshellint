# @file evalstring.sh
# @brief Values spliced into code that another parser reads
# @description Each call below hands a string to something that parses it as
#   shell code; a variable spliced into it runs whatever it holds.

#######################################
# @description Run a sign command over a path, every way
# @arg $1 string Path to sign
#######################################
function evalstring::sign {
  local path safe
  dybatpho::expect_args path -- "$@"
  printf -v safe '%q' "${path}"
  dybatpho::dry_run "gpg --sign ${path}"
  bash -c "gpg --sign ${path}"
  dybatpho::dry_run "gpg --sign ${safe}"
  bash -c "gpg --sign ${path@Q}"
  dybatpho::dry_run gpg --sign "${path}"
  bash -c 'gpg --sign "$1"' _ "${path}"
}
