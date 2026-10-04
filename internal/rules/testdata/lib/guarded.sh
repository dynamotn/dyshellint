# @file guarded.sh
# @brief A library that returns early when it is already loaded
# @description The guard keeps a second `source` from reaching the constants.

[[ -n "${GUARDED_LOADED-}" ]] && return 0
GUARDED_LOADED=1
readonly GUARDED_VERSION="1.0.0"
