#!/usr/bin/env bash
# @file bad_secrets_masked.sh
# @brief A registered secret is masked by dybatpho, and only by dybatpho
# @description `echo` prints a registered value as it is, and a value nobody
#   registered is printed by every logger.
# shellcheck source=/dev/null
. ./lib/dybatpho/init.sh
dybatpho::register_common_handlers

#######################################
# @description Log a request failure
# @noargs
#######################################
function _report {
  dybatpho::secret_register "${SESSION_TOKEN}"
  dybatpho::error "Request failed for ${SESSION_TOKEN}"
  echo "Request failed for ${SESSION_TOKEN}"
  dybatpho::warn "Retrying with ${API_TOKEN}"
}

_report
