#!/usr/bin/env bash
# @file bad_download.sh
# @brief Downloads run, installed or unpacked with nothing checking them
# @description A truncated file or a compromised mirror decides what runs.
set -euo pipefail

#######################################
# @description Install a tool from the network
# @arg $1 string URL of the installer
# @arg $2 string URL of the archive
#######################################
function _install {
  local url="$1" archive_url="$2" installer=./install.sh archive=./tool.tar.gz
  curl --fail -sSL --max-time 60 "${url}" -o "${installer}"
  bash "${installer}"
  curl -fsSLo "${archive}" --max-time 60 "${archive_url}"
  tar -xzf "${archive}" -C ./tool
  wget -q --timeout=60 "${archive_url}"
  unzip ./tool.zip
  curl --fail -sSL --max-time 60 "${archive_url}" | tar -xz -C ./tool
  curl --fail -sSL --max-time 60 --output ./tool/bin/tool "${url}"
  chmod 755 ./tool/bin/tool
  dybatpho::curl_download "${url}" ./data.json
  jq . ./data.json
}

_install "$@"
