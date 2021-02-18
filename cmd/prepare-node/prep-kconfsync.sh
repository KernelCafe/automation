#!/bin/sh
#
# Prepare node for confsync
set -eux

if [ -f /proc/sys/kernel/hostname ]; then
  hostname=$(cat /proc/sys/kernel/hostname)
else
  hostname=$(hostname -s)
fi

repo="host-$hostname"
srcdir="$(dirname $0)"

if [ ! -f "${HOME}/.ssh/id_rsa.pub" ]; then
  cat /dev/zero | ssh-keygen -q -N ""
fi

git config --global user.email "${repo}@kernel.cafe"
git config --global user.name "${repo}"

rsync --version

tf=$(mktemp)
echo "*/15 * * * * $HOME/${repo}/sync.sh" > $tf
crontab $tf
rm -f $tf

echo "MANUAL STEP: Add key to kernelcafe-hostbot:"
cat $HOME/.ssh/id_rsa.pub
echo "MANUAL STEP: Create ${repo} repository on GitHub"
read nadab

if [ ! -d "${HOME}/${repo}/.git" ]; then
  git clone git@github.com:KernelCafe/${repo}.git "${HOME}/${repo}"
fi

cp "${srcdir}/../kconfsync/kconfsync.sh" "${HOME}/${repo}/sync.sh"
cp "${srcdir}/../kconfsync/gitignore" "${HOME}/${repo}/.gitignore"
chmod 755 "${HOME}/${repo}/sync.sh"
cd "${HOME}/${repo}"
./sync.sh
