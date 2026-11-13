#!/usr/bin/env bash
# @file good_commentedcode_unicode.sh
# @brief Comments written in another language are prose, not code
# @description Words with diacritics count as ordinary words.
set -euo pipefail

# "các tùy chọn lệnh hoặc tên đường dẫn"
# ($1 được giả định là chứa một giá trị ở đây)
grep -li -- Hugo /dev/null "${1:-}"

# $* là sai hầu hết mọi lúc
#
#   chứa dấu cách và loại bỏ các chuỗi trống
printf '%s\n' "$@"

# Đây là trong hầu hết các trường hợp những gì bạn muốn dùng
printf '%s\n' "finished"
