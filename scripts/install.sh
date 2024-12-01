#!/usr/bin/env bash

# exit codes
# 0 - exited without problems
# 1 - OS not supported by this script

set -e

#detect the platform
OS="$(uname)"
case $OS in
  Linux)
    OS='linux'
    download_link="https://github.com/ozgur-yalcin/mfa/releases/latest/download/mfa_Linux_x86_64.tar.gz"
    mfa_archive="mfa_Linux_x86_64.tar.gz"
    ;;
  Darwin)
    OS='osx'
    download_link="https://github.com/ozgur-yalcin/mfa/releases/latest/download/mfa_Darwin_x86_64.tar.gz"
    mfa_archive="mfa_Darwin_x86_64.tar.gz"
    binTgtDir=/usr/local/bin
    ;;
  *)
    echo 'OS not supported'
    exit 1
    ;;
esac

OS_type="$(uname -m)"
case "$OS_type" in
  x86_64|amd64)
    OS_type='amd64'
    ;;
  *)
    echo 'OS type not supported'
    exit 1
    ;;
esac


printf "Downloading package, please wait\n"
curl -LO "$download_link"

printf "Extracting archive...\n"
decompressed_dir="/tmp/mfa"
if [ ! -d "$decompressed_dir" ]
then
  mkdir "$decompressed_dir"
fi

tar -xzf "$mfa_archive" --directory "$decompressed_dir"

printf "Successfully extracted archive\n"

cd "$decompressed_dir"

printf "Starting package install...\n"

case "$OS" in
  'linux')
    cp mfa /usr/bin/mfa.new
    chmod 755 /usr/bin/mfa.new
    chown root:root /usr/bin/mfa.new
    mv /usr/bin/mfa.new /usr/bin/mfa
    ;;
  'osx')
    mkdir -m 0555 -p ${binTgtDir}
    cp mfa ${binTgtDir}/mfa.new
    mv ${binTgtDir}/mfa.new ${binTgtDir}/mfa
    chmod a=x ${binTgtDir}/mfa
    ;;
  *)
    echo 'OS not supported'
    exit 2
esac

printf "Successfully installed\n"

cd ..

if [ -d "$decompressed_dir" ]
then
  rm -r "$decompressed_dir"
fi