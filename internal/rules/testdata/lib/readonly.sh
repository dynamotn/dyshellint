# @file readonly.sh
# @brief A library that declares constants without a source guard
# @description Sourcing it twice stops on the second `readonly`.

readonly READONLY_VERSION="1.0.0"
declare -r READONLY_NAME="readonly"
declare -g READONLY_MUTABLE="ok"
