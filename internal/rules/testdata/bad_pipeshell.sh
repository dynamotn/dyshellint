#!/usr/bin/env bash
# @file bad_pipeshell.sh
# @brief Downloads run straight from the network
# @description Whatever the server sends runs, truncated or not.
set -euo pipefail

#######################################
# @description Install a tool, every way
# @noargs
#######################################
function _install {
  curl -fsSL https://example.com/install.sh | bash
  wget -qO- https://example.com/install.sh | sh -s -- --yes
  bash <(curl -fsSL https://example.com/install.sh)
  curl -fsSL https://example.com/install.sh | source /dev/stdin
  curl -fsSL https://example.com/data.json | jq .
  curl -fsSL https://example.com/install.sh -o install.sh
  bash install.sh
}
