#!/usr/bin/env bash
# @file bad_directive.sh
# @brief ShellCheck directives with and without a reason
# @description A directive should say why.
set -euo pipefail

# shellcheck disable=SC2034
unused_a=1
# shellcheck disable=SC2034 # read by a sourced template
unused_b=1
# Read by the template sourced after this file.
# shellcheck disable=SC2034
unused_c=1
