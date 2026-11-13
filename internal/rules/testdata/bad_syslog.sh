#!/usr/bin/env bash
# @file bad_syslog.sh
# @brief Logs written into /var/log by hand
# @description Nothing rotates these files, and they need root.
set -euo pipefail

#######################################
# @description Run a backup and log it
# @noargs
#######################################
function _backup {
  printf 'Backup started\n' >> /var/log/backup.log
  rsync -a /srv/ /backup/ &>> "/var/log/backup.log"
  printf 'Backup done\n' | tee -a /var/log/backup.log
}

_backup
