#!/bin/sh
#
# Prepare node for ansible
set -eux
uname=$(uname)
user="barista"

if [ "${uname}" = "Linux" ]; then
    getent group $user || sudo groupadd -g 2000 $user
    id barista || sudo useradd -m -g $user -G sudo -r $user
elif [ "${uname}" = "Darwin" ]; then
    home=/Users/$user
    if ! dscl . -read /Groups/barista; then
        sudo dscl . -create /Groups/$user
        sudo dscl . -create /Groups/$user gid 2000
    fi

    if ! id barista; then
        sudo dscl . -create /Users/$user
        sudo dscl . -create /Users/$user UserShell /bin/bash
        sudo dscl . -create /Users/$user RealName "$user"
        sudo dscl . -create /Users/$user UniqueID 2000
        sudo dscl . -create /Users/$user PrimaryGroupID 2000
        sudo dscl . -create /Users/$user NFSHomeDirectory $home
        sudo dscl . -append /Groups/admin GroupMembership $user
        sudo mkdir -p /Users/$user
    fi
    sudo chown -R barista:2000 /Users/$user
fi

echo "MANUAL: Run visudo and add:"
echo "barista         ALL = (ALL) NOPASSWD: ALL"
echo ""
read xy

sudo -u $user -H ./prep-authorized-keys.sh
sudo -u $user -H ./prep-kconfsync.sh
