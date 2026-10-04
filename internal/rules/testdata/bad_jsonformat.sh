#!/usr/bin/env bash
# @file bad_jsonformat.sh
# @brief JSON documents built by splicing raw values into a format
# @description A quote or a newline in a value spliced between JSON quotes
#   breaks the document.
set -euo pipefail

#######################################
# @description Print a JSON object from raw values
# @arg $1 string Name to report
# @arg $2 string Count to report
#######################################
function _report {
  local name="$1" count="$2"
  printf '{"name":"%s","count":%d}\n' "${name}" "${count}"
  echo "{\"name\":\"${name}\"}"
  printf '{"count":%d}\n' "${count}"
  printf '%s: "%s"\n' "label" "${name}"
}

#######################################
# @description Print a JSON object from escaped values
# @arg $1 string Name to report
#######################################
function _report_escaped {
  local name
  name="$(_json_escape "$1")"
  printf '{"name":"%s"}\n' "${name}"
}

#######################################
# @description Escape a value for a JSON string
# @arg $1 string Value to escape
# @stdout The escaped value
#######################################
function _json_escape {
  local value="${1//\\/\\\\}"
  printf '%s' "${value//\"/\\\"}"
}
