#!/bin/sh
# init script for isucon6-qualifier

set -ex
curl https://github.com/{Songmu,motemen,tatsuru,edvakf,catatsuy,walf443,st-cyrill,myfinder,aereal,tarao,yuuki}.keys >> /home/isucon/.ssh/authorized_keys
