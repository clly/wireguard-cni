#!/usr/bin/bash

set -euof pipefail

N=$1


netnspath=/var/run/netns/container
function catch() {
	NETCONFPATH=./cni CNI_PATH=/usr/lib/cni ./exec-plugins.sh del container $netnspath
	ip netns del container
}

trap 'catch' ERR


NETCONFPATH=./cni CNI_PATH=/usr/lib/cni ./exec-plugins.sh del container $netnspath
ip netns del container
#ip link add container0 type wireguard
#ip link set container0 netns container
