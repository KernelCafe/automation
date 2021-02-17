#!/bin/sh
#
# Prepare node for ansible
# 
# To be run as the `barista` user
set -eux

srcdir="$(dirname $0)"

mkdir -p $HOME/.ssh
if [ ! -f "$HOME/.ssh/authorized_keys" ]; then
  cp "${srcdir}/barista.pub" $HOME/.ssh/authorized_keys
fi
chmod 700 $HOME/.ssh
chmod 400 $HOME/.ssh/*
