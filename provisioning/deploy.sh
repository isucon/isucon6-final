#!/bin/sh

set -e

RESOURCE_GROUP="${RESOURCE_GROUP:-i6f}"
LOCATION="${LOCATION:-japaneast}"
SSH_PUBLIC_KEY="${SSH_PUBLIC_KEY:-`cat ~/.ssh/id_rsa.pub`}"

# azure login
# azure config mode arm
# azure account list
# azure account set <subscriptionNameOrId>
azure group create ${RESOURCE_GROUP} ${LOCATION}
azure group deployment create -f deploy.json -p "{\"sshPublicKey\":{\"value\":\"${SSH_PUBLIC_KEY}\"}}" -g ${RESOURCE_GROUP}
