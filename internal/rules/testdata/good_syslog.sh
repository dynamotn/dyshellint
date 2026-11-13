#!/usr/bin/env bash
# @file good_syslog.sh
# @brief Logs sent to the supervisor
# @description STDERR goes to journald, and logger reaches syslog.
set -euo pipefail

#######################################
# @description Run a backup and log it
# @noargs
#######################################
function _backup {
  printf 'Backup started\n' >&2
  rsync -a /srv/ /backup/
  logger -t backup "Backup done"
  grep -c error /var/log/syslog > /dev/null || true
  wc -l < /var/log/backup.log
}

_backup
