#!/bin/sh
#
# Prepare node for confsync
set -eux

repo="host-$(hostname -s)"
srcdir="$(dirname $0)"

if [ ! -f "${HOME}/.ssh/id_rsa.pub" ]; then
  cat /dev/zero | ssh-keygen -q -N ""
fi

git config --global user.email "${repo}@kernel.cafe"
git config --global user.name "${repo}"

tf=$(mktemp)
echo "*/5 * * * * $HOME/${repo}/sync.sh" > $tf
crontab $tf
rm -f $tf

echo "MANUAL STEP: Add key to kernelcafe-hostbot:"
cat $HOME/.ssh/id_rsa.pub
echo "MANUAL STEP: Create ${repo} repository on GitHub"
read nadab

git clone git@github.com:KernelCafe/${repo}.git "${HOME}/${repo}"
cp "${srcdir}/../kconfsync/kconfsync.sh" "${HOME}/${repo}/sync.sh"
cp "${srcdir}/../kconfsync/gitignore" "${HOME}/${repo}/.gitignore"
chmod 755 "${HOME}/${repo}/sync.sh"
cd "${HOME}/${repo}"
./sync.sh
