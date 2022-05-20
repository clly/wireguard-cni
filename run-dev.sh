#!/usr/bin/env bash

set -euo pipefail

PACKAGED_CNI_PATH=/opt/cni/bin/
NETCONFPATH=$PWD/config

cnipath=$(mktemp -d)
trap "rm -rf $cnipath" EXIT

rsync -a $PACKAGED_CNI_PATH $cnipath/
go build .
cp wireguard-cni $cnipath
(
cd scripts
sudo id
sudo NETCONFPATH=$NETCONFPATH CNI_PATH=$cnipath ./priv-net-run.sh ip addr && ip link
)
