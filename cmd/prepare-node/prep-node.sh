#!/bin/sh
#
# Prepare node for ansible
set -eux
uname=$(uname)
user=$user

if [ "${uname}" = "Linux" ]; then
    sudo groupadd -g 2000 $user
    sudo useradd -m -g $user -G sudo -r $user
elif [ "${uname}" = "Darwin" ]; then
    home=/Users/$user
    sudo dscl . -create /Groups/$user
    sudo dscl . -create /Groups/$user gid 2000
    sudo dscl . -create /Users/$user
    sudo dscl . -create /Users/$user UserShell /bin/bash
    sudo dscl . -create /Users/$user RealName "$user"
    sudo dscl . -create /Users/$user UniqueID 2000
    sudo dscl . -create /Users/$user PrimaryGroupID 2000
    sudo dscl . -create /Users/$user NFSHomeDirectory $home
fi

sudo -u $user -H ./prep-ansible.sh
sudo -u $user -H ./prep-confsync.sh
