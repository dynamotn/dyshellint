# @file headerdoc.sh
# @brief The file header is not the doc of the first function
# @description A blank line ends the header, so the function below it has none.
# @set counter

function headerdoc::undocumented {
  printf -v counter '%s' 1
}
