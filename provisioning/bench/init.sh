#!/bin/sh
# init script for isucon6-qualifier

set -ex

export DEBIAN_FRONTEND=noninteractive
apt update
apt install -y ansible git aptitude
apt remove -y snapd

mkdir -p -m 700 /root/.ssh
wget -O /root/.ssh/id_rsa https://gist.githubusercontent.com/tatsuru/eab77dd56a7bf81e05c34ae9b0938bb4/raw/caefb7a51a8f2be0ea4207e135824c488ae7b045/id_rsa
chmod 600 /root/.ssh/id_rsa
ssh-keyscan -t rsa github.com >> /root/.ssh/known_hosts
export HOME=/root
git config --global user.name "isucon"
git config --global user.email "isucon@isucon.net"

curl -fsSL https://mackerel.io/assets/files/scripts/setup-apt.sh | sh
sudo apt install mackerel-agent
sudo mackerel-agent init -apikey="6iMqSamBFuKarU8aTXBY3s3sit8qrSxxW8EMc8oxRkSC"
sudo sh -c 'echo >> /etc/default/mackerel-agent'
sudo sh -c "echo "'OTHER_OPTS=\"-role=Portal:worker\" >> /etc/default/mackerel-agent'
sudo /etc/init.d/mackerel-agent start

git clone git@github.com:Songmu/isucon6-qualifier /tmp/isucon6-qualifier
cd /tmp/isucon6-qualifier/provisioning/bench
PYTHONUNBUFFERED=1 ANSIBLE_FORCE_COLOR=true ansible-playbook -i localhost, ansible/*.yml --connection=local
cd /tmp && rm -rf /tmp/isucon6-qualifier
curl https://github.com/{Songmu,motemen,tatsuru,edvakf,catatsuy,walf443,st-cyrill,myfinder,aereal,tarao,yuuki}.keys >> /home/isucon/.ssh/authorized_keys

