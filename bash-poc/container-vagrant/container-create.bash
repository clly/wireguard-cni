#!/usr/bin/bash

set -euof pipefail

N=$1


netnspath=/var/run/netns/container
function catch() {
	NETCONFPATH=./cni CNI_PATH=/usr/lib/cni ./exec-plugins.sh del container $netnspath
	ip netns del container
}

trap 'catch' ERR

ip netns add container || true

NETCONFPATH=./cni CNI_PATH=/usr/lib/cni ./exec-plugins.sh add container $netnspath
set -x
#ip link add container0 type wireguard
#ip link set container0 netns container
cp $N-container0.conf /etc/wireguard/container0.conf
ip netns exec container wg-quick up container0
