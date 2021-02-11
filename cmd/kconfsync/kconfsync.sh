#!/bin/sh
dirs="/etc /usr/local/etc /root /usr/pkg/etc /opt/homebrew/etc"
cd $HOME/host-$(hostname -s)
for dir in $dirs; do
    if [ ! -d "${dir}" ]; then
        continue
    fi

    rsync -vaRm --delete-excluded $dir \
        --exclude "*.db" \
        --exclude "*.log" \
        --exclude "*.pem" \
        --exclude "*.pid" \
        --exclude "*.png" \
        --exclude "*.sample" \
        --exclude "*.swp" \
        --exclude "key.*" \
        --exclude "*cert*" \
        --exclude "*chatscript*" \
        --exclude "*key" \
        --exclude "*keys" \
        --exclude "*passw*" \
        --exclude "*polkit*" \
        --exclude "*pwd*" \
        --exclude "*shadow*" \
        --exclude "*snmp*" \
        --exclude "*sudoers*" \
        --exclude "*users*" \
        --exclude "*~" \
        --exclude "gcp-cups*json" \
        --exclude "passw*" \
        --exclude bluetooth/ \
        --exclude nsmb.conf \
        --exclude devfs.rules \
        --exclude opieaccess \
        --exclude snmpd.conf \
        --exclude tcsd.conf \
        --exclude cups/ \
        --exclude errors/ \
        --exclude fonts/ \
        --exclude ntp/ \
        --exclude ppp/ \
        --exclude rc.d/ \
        --exclude security/ \
        --exclude ssl/ \
        --exclude ssmtp \
        --exclude syncthing \
        --exclude templates/ \
        --exclude wireguard/ \
        --exclude xdg/ \
        .
done

git add .
git diff
git commit -am "$(git status -s | xargs)"
git push

# self-update
curl --version || exit 0
tf=$(mktemp)
curl https://raw.githubusercontent.com/KernelCafe/automation/main/cmd/kconfsync/kconfsync.sh > "${tf}"
chmod 755 "${tf}"
"${tf}" && cp "${tf}" sync.sh