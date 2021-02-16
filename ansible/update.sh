#!/bin/sh
git fetch; git pull; ansible-playbook -i hosts playbooks/users.yaml

