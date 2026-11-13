#!/usr/bin/env bash
# @file bad_daemonize.sh
# @brief A service detached from the script that started it
# @description Nothing restarts or stops these processes.
set -euo pipefail

nohup ./server.sh --port 8080 &> server.log &
./worker.sh &
disown
setsid ./watcher.sh &
