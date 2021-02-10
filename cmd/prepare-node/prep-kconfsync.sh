#!/bin/sh
#
# Prepare node for confsync
set -eux

repo="host-$(hostname -s)"
srcdir="$(dirname $0)"
cat /dev/zero | ssh-keygen -q -N ""
git config --global user.email "${repo}@kernel.cafe"
git config --global user.name "${repo}"

tf=$(mktemp)
echo "*/5 * * * * $HOME/${repo}/sync.sh" > $tf
crontab $tf
rm -f $tf

echo "MANUAL STEP: Add key to kernelcafe-hostbot:"
cat $HOME/.ssh/id_rsa.pub
read nada

echo "MANUAL STEP: Create ${repo} repository on GitHub"
read nadab

cd $HOME
git checkout git@github.com:KernelCafe/${repo}.git
cd "${repo}"
cp "${srcdir}/../kconfsync/kconfsync.sh" sync.sh
cp "${srcdir}/../kconfsync/gitignore" .gitignore
chmod 755 sync.sh
./sync.sh
