#!/usr/bin/env bash

# exit codes
# 0 - exited without problems
# 1 - OS not supported by this script

set -e

#detect the platform
OS="$(uname)"
case $OS in
  Linux)
    sudo rm -rf /usr/bin/mfa
    ;;
  Darwin)
    rm -rf /usr/local/bin/mfa
    ;;
  *)
    echo 'OS not supported'
    exit 1
    ;;
esac

printf "Successfully uninstalled mfa\n"