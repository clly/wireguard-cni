#!/usr/bin/bash

set -euof pipefail

#N=$1


netnspath=/var/run/netns/container

ip netns del container
#ip link add container0 type wireguard
#ip link set container0 netns container
