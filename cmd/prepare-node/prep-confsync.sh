#!/bin/sh
#
# Prepare node for ansible
set -eux

repo="host-$(hostname -s)"
cat /dev/zero | ssh-keygen -q -N ""
git config --global user.email "${repo}@kernel.cafe"
git config --global user.name "${repo}"

tf=$(mktemp)
repo="host-$(hostname -s)"
echo "*/5 * * * * $HOME/${repo}/sync.sh" > $tf
crontab $tf
rm -f $tf

echo "MANUAL STEP: Add key to kernelcafe-hostbot:"
cat $HOME/.ssh/id_rsa.pub
read nada

echo "MANUAL STEP: Create ${repo} repository on GitHub"
cd $HOME
git checkout git@github.com:KernelCafe/${repo}.git
cd "${repo}"
./sync.sh
