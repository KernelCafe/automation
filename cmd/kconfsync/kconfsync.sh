#!/bin/sh
# check that rsync is installed
rsync --version || exit 1

if [ -f /proc/sys/kernel/hostname ]; then
  hostname=$(cat /proc/sys/kernel/hostname)
else
  hostname=$(hostname -s)
fi

dirs="/etc /usr/local/etc /root /usr/pkg/etc /opt/homebrew/etc"
cd $HOME/host-$hostname || exit 1

for dir in $dirs; do
    if [ ! -d "${dir}" ]; then
        continue
    fi
    if [ ! -r "${dir}" ]; then
        continue
    fi

    rsync -vaRm --delete-excluded $dir \
        --exclude "securetty" \
	--exclude "kernel.d/" \
        --exclude "wpa_supplicant/" \
	--exclude "*.cache" \
	--exclude "*.seed" \
	--exclude "*pgp*" \
	--exclude "*private*" \
	--exclude ".ssh/" \
	--exclude "crypttab" \
	--exclude acme/ \
	--exclude audisp-remote.conf \
	--exclude gnupg/ \
	--exclude libaudit.conf \
	--exclude master.passwd \
	--exclude radiusd.conf \
	--exclude useradd \
	--exclude zos-remote.conf \
        --exclude "*.db" \
        --exclude "*.log" \
        --exclude "*.pem" \
        --exclude "*.pid" \
        --exclude "*.png" \
        --exclude "*.sample" \
        --exclude "*.swp" \
        --exclude "*cert*" \
        --exclude "*chatscript*" \
        --exclude "*key" \
        --exclude "*keys" \
        --exclude "*polkit*" \
        --exclude "*pwd*" \
        --exclude "*shadow*" \
        --exclude "*snmp*" \
        --exclude "*sudoers*" \
        --exclude "*users*" \
        --exclude "*~" \
        --exclude "gcp-cups*json" \
        --exclude "key.*" \
        --exclude bluetooth/ \
        --exclude cups/ \
        --exclude devfs.rules \
        --exclude errors/ \
        --exclude fonts/ \
        --exclude nsmb.conf \
        --exclude ntp/ \
        --exclude opieaccess \
        --exclude ppp/ \
        --exclude rc.d/ \
        --exclude security/ \
        --exclude snmpd.conf \
        --exclude ssl/ \
        --exclude ssmtp \
        --exclude syncthing \
        --exclude tcsd.conf \
        --exclude templates/ \
        --exclude wireguard/ \
        --exclude xdg/ \
	.
done

git add .
git commit -am "$(git status -s | xargs)"
git push

