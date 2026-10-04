#!/usr/bin/env bash
# @file bad_heredoc_dash.sh
# @brief A here document indented with tabs
# @description <<- strips tabs only.
set -euo pipefail

if true; then
	cat <<- EOF
		indented
	EOF
fi
