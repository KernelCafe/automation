#!/bin/sh
w="${HOME}/welcome"
pb="${HOME}/ansible/playbooks"
mkdir -p $pb

git fetch && git pull
cp ansible.cfg $HOME/ansible

cwd=$(pwd)


cd $w && git fetch && git pull
cd $cwd
cd ..

killall ansible
go run cmd/generate-ansible/main.go --usermap $w/auth/users.yaml --groupmap $w/auth/groups.yaml --nodemap $w/nodes/nodes.yaml --out $pb
cat $w/nodes/nodes.yaml | grep "^- name:" | cut -d" " -f3 > $HOME/ansible/hosts
ls $pb | xargs -n1 -J8 ansible-playbook -i $HOME/ansible/hosts

