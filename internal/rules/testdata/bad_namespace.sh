# @file bad_namespace.sh
# @brief A namespace declaration the guide rejects
# @description The declared namespace is not a name a function can carry, so it
#   is reported and the name of the file stands in for it.
# dyshellint namespace=Pkg-1

#######################################
# @description Named after the broken declaration
# @noargs
#######################################
function Pkg-1::install {
  printf 'unusable namespace\n'
}
