#!/bin/sh

set -e

RESOURCE_GROUP="${RESOURCE_GROUP:-isucon6-final}"
LOCATION="${LOCATION:-japaneast}"
SSH_PUBLIC_KEY="${SSH_PUBLIC_KEY:-`cat ~/.ssh/id_rsa.pub`}"
PREFIX="${PREFIX:-prefix}"
VM_COUNT="${VM_COUNT:-1}"

# azure login
# azure config mode arm
# azure group create ${RESOURCE_GROUP} ${LOCATION}
azure group deployment create -f deploy.json -p "{\"sshPublicKey\":{\"value\":\"${SSH_PUBLIC_KEY}\"},\"prefix\":{\"value\":\"${PREFIX}\"},\"vmCount\":{\"value\":${VM_COUNT}}}" -g ${RESOURCE_GROUP}
