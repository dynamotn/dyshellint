#!/usr/bin/env bash
# @file bad_secrets.sh
# @brief Credentials put where other users or a log can read them
# @description Arguments of curl show up in `ps`, and printed values end up in
#   the terminal and in logs.
set -euo pipefail

#######################################
# @description Call an API with a token, every way
# @noargs
#######################################
function _call {
  curl -H "Authorization: Bearer ${API_TOKEN}" https://example.com
  curl -H "Authorization: Bearer ${auth}" https://example.com
  curl "https://hooks.example.com/${SLACK_WEBHOOK}"
  echo "using ${DB_PASSWORD}"
  printf '%s' "${API_TOKEN}" | curl -H @- https://example.com
  printf '%s\n' "${DB_PASSWORD}" > "${HOME}/.pgpass"
  printf '%s\n' "${#API_TOKEN}"
  echo "token file: ${TOKEN_FILE}"
  local masked
  masked="$(printf '%s' "${API_TOKEN}" | cut -c1-4)"
  echo "${masked}"
}
