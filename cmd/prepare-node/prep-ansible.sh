#!/bin/sh
#
# Prepare node for ansible
set -eux

srcdir="$(dirname $0)"

mkdir -p $HOME/.ssh
cp "${srcdir}/barista.pub" $HOME/.ssh/authorized_keys

