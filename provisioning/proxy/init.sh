#!/bin/sh
# init script for isucon6-qualifier

set -ex

export DEBIAN_FRONTEND=noninteractive
apt update
apt install -y ansible git aptitude
apt remove -y snapd

mkdir -p -m 700 /root/.ssh
wget -O /root/.ssh/id_rsa https://gist.githubusercontent.com/catatsuy/6fcfb32e59c2356c4b525fa4fface701/raw/811e3a6f94e52dcdd198224a56c537b1cd3afb98/id_rsa
chmod 600 /root/.ssh/id_rsa
ssh-keyscan -t rsa github.com >> /root/.ssh/known_hosts
export HOME=/root
git config --global user.name "isucon"
git config --global user.email "isucon@isucon.net"

git clone git@github.com:catatsuy/isucon6-final.git /tmp/isucon6-final
cd /tmp/isucon6-final/provisioning/proxy
PYTHONUNBUFFERED=1 ANSIBLE_FORCE_COLOR=true ansible-playbook -i localhost, ansible/*.yml --connection=local
curl https://github.com/{Songmu,motemen,tatsuru,edvakf,catatsuy,walf443,st-cyrill,myfinder,aereal,tarao,yuuki}.keys >> /home/isucon/.ssh/authorized_keys
cd /tmp && rm -rf /tmp/isucon6-final
/sbin/shutdown -r now
