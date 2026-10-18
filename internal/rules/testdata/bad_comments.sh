#!/usr/bin/env bash
set -euo pipefail

# TODO(someone): explain why this exists
function _undocumented {
  printf '%s\n' "$1"
}

#######################################
# @description Take an argument without saying so
#######################################
function _partly_documented {
  printf '%s\n' "$1"
}

# TODO @alice: rewrite the loop
# TODO: @bob check the format
# TODO: document the flags with @arg tags
# TODO @arg tags for the flags
# TODO: ask bob@example.com about the format
