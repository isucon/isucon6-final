#!/bin/sh

set -e

export DEBIAN_FRONTEND=noninteractive
apt-get update
apt-get install -y ansible git

tar zxf ansible.tar.gz
(
  cd ansible
  PYTHONUNBUFFERED=1 ANSIBLE_FORCE_COLOR=true ansible-playbook -i development -c local playbook/setup.yml
)
rm -rf ansible
