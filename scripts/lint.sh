#!/usr/bin/env bash
# @file lint.sh
# @brief Check every shell script against the Bash coding style guide
# @description Builds `dyshellint` from `cmd/dyshellint` and runs it over the
#   repository. `dyshellint` drives ShellCheck and shfmt itself, so this script
#   is the single entrypoint for the whole check.
DYBATPHO_PATH="${DYBATPHO_DIR:-${HOME}/Dotfiles/scripts/lib/dybatpho}"
if [[ ! -r "${DYBATPHO_PATH}/init.sh" ]]; then
  printf 'dybatpho not found at %s; set DYBATPHO_DIR to its checkout\n' \
    "${DYBATPHO_PATH}" >&2
  exit 1
fi
# shellcheck source=/dev/null
. "${DYBATPHO_PATH}/init.sh" --modules release safety cli
dybatpho::register_common_handlers

readonly REPO_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
readonly LINTER="${REPO_DIR}/bin/dyshellint"

#######################################
# @description Build the linter, so that a check never runs a stale binary
# @noargs
# @set LINTER
# @exitcode 0 If the linter is built
#######################################
function _build_linter {
  dybatpho::require go
  dybatpho::info "Building ${LINTER}"
  dybatpho::dry_run go -C "${REPO_DIR}" build -o "${LINTER}" ./cmd/dyshellint
}

#######################################
# @description Rewrite every script with shfmt, using the options of the guide
# @noargs
# @exitcode 0 If every file is formatted
#######################################
function _format {
  dybatpho::require shfmt
  dybatpho::info "Formatting shell scripts"
  dybatpho::dry_run shfmt --language-dialect bash --indent 2 --case-indent \
    --binary-next-line --space-redirects --write "${REPO_DIR}/scripts"
}

# The options below are declared by `_spec_main` and filled in by
# `dybatpho::generate_from_spec`, which ShellCheck cannot see through.
# shellcheck disable=SC2154
#######################################
# @description Run the linter over the paths the caller asked for
# @noargs
# @exitcode 0 If no error is reported
# @exitcode 1 If any error is reported
#######################################
function _main {
  _build_linter
  if [[ "${FIX}" == "true" ]]; then
    _format
  fi

  local -a paths=()
  paths=("${LINT_ARGS[@]}")
  if ((${#paths[@]} == 0)); then
    paths=("${REPO_DIR}/scripts")
  fi

  local -a options=(--format "${FORMAT}")
  if [[ "${STRICT}" == "true" ]]; then
    options+=(--warnings-as-errors)
  fi

  dybatpho::info "Linting ${paths[*]}"
  "${LINTER}" "${options[@]}" "${paths[@]}"
}

#######################################
# @description Spec of lint.sh
# @noargs
#######################################
function _spec_main {
  dybatpho::opts::setup "Check shell scripts against the Bash coding style guide" \
    LINT_ARGS action:"_main"
  dybatpho::opts::flag "Reformat the scripts with shfmt before linting" FIX \
    --fix on:true off:false init:="false"
  dybatpho::opts::flag "Fail on warnings too" STRICT \
    --strict on:true off:false init:="false"
  dybatpho::opts::param "Output format, text or json" FORMAT \
    --format init:="text" choices:text,json
  dybatpho::opts::param "Log level" LOG_LEVEL --log-level init:="info" \
    validate:"dybatpho::validate_log_level \$OPTARG"
  dybatpho::opts::disp "Show help" --help action:"dybatpho::generate_help _spec_main"
}

dybatpho::generate_from_spec _spec_main "$@"
