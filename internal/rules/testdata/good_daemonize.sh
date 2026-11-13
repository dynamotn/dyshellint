#!/usr/bin/env bash
# @file good_daemonize.sh
# @brief Services handed to systemd
# @description systemd restarts, stops and logs them.
set -euo pipefail

systemctl --user start app-server.service
setsid --wait ./job.sh
setsid -w ./job.sh
