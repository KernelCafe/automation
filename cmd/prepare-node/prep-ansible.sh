#!/bin/sh
#
# Prepare node for ansible
# 
# To be run as the `barista` user
set -eux

srcdir="$(dirname $0)"

mkdir -p $HOME/.ssh
cp "${srcdir}/barista.pub" $HOME/.ssh/authorized_keys
chmod 700 $HOME/.ssh
chmod 400 $HOME/.ssh/*
