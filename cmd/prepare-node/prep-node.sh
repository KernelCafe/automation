#!/bin/sh
#
# Prepare node for ansible
set -eux
sudo groupadd -g 2010 barista
sudo useradd -m -g barista -G sudo -r barista
sudo -u barista -H ./prepare-ansible.sh
sudo -u barista -H ./prepare-confsync.sh
