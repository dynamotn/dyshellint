#!/usr/bin/env bash
# @file good_crlf.sh
# @brief A protocol line ending is not a redraw
# @description `\r\n` ends a line of an HTTP answer.
set -euo pipefail

printf 'HTTP/1.1 200 OK\r\n\r\n'
