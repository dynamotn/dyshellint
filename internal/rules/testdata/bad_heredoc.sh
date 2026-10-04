#!/usr/bin/env bash
# @file bad_heredoc.sh
# @brief Here documents written the hard way
# @description Tabs with <<- and escapes by hand.
set -euo pipefail

#######################################
# @description Print help, twice
# @noargs
#######################################
function _help {
  cat << EOF
Run: \$HOME/bin/tool, cost: \$5
EOF
  cat << 'EOF'
Run: $HOME/bin/tool
EOF
  cat << EOF
Home is ${HOME}, cost: \$5
EOF
}

_help

