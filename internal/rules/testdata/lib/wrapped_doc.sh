# @file wrapped_doc.sh
# @brief A library whose shdoc text starts on the line below its tag
# @description
#   shdoc also accepts a bare tag line, with the text following underneath.
#   The header is complete even though nothing sits next to `@description`.

#######################################
# @description
#   Print the name of this library, documented in the same wrapped style.
# @noargs
# @stdout The name
#######################################
function wrapped_doc::name {
  printf '%s\n' "wrapped_doc"
}
