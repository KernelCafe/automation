#!/bin/sh
w="${HOME}/welcome"
pb="${HOME}/ansible/run"

mkdir -p $pb
cd $HOME/automation/ansible

git fetch && git pull
cp ansible.cfg $HOME/ansible
cwd=$(pwd)


cd $w && git fetch && git pull
cd $cwd
cd ..

killall ansible
rm -f $pb/*.yaml
cp $HOME/automation/ansible/playbooks/* $pb
go run cmd/generate-ansible/main.go --usermap $w/auth/users.yaml --groupmap $w/auth/groups.yaml --nodemap $w/nodes/nodes.yaml --out $pb
egrep -o -- "- name:.*"  $w/nodes/nodes.yaml | cut -d " " -f3 > $HOME/ansible/hosts
ls $pb/* | xargs -n1 -P24 ansible-playbook -i $HOME/ansible/hosts

