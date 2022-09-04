#!/usr/bin/bash

set -euof pipefail

N=$1

case $N in
	server)
		ADDRESS=10.0.10.3
		ROUTES="10.0.0.0/24 "
		;;
	peer)
		ADDRESS=10.0.0.3
		;;
	*)
		echo "unsupported"
		exit 1
		;;
esac

netnspath=/var/run/netns/container
#function catch() {
#	NETCONFPATH=./cni CNI_PATH=/usr/lib/cni ./exec-plugins.sh del container $netnspath
#	ip netns del container
#}

function catch() {
	ip netns exec container ip link del container0
}

trap 'catch' ERR

ip netns add container || true

#NETCONFPATH=./cni CNI_PATH=/usr/lib/cni ./exec-plugins.sh add container $netnspath
set -x
ip link add container0 type wireguard
ip link set container0 netns container
cp $N-container0.conf /etc/wireguard/container0.conf
ip netns exec container wg setconf container0 /etc/wireguard/container0.conf
ip netns exec container ip -4 address add "${ADDRESS}" dev container0
ip netns exec container ip link set mtu 1420 up dev container0
#ip netns exec container wg set container0 fwmark 51820
ip netns exec container ip -4 route add 0.0.0.0/0 dev container0
#ip netns exec container ip -4 rule add not fwmark 51820 table 51820
#ip netns exec container ip -4 rule add table main suppress_prefixlength 0
##sysctl -q net.ipv4.conf.all.src_valid_mark=1
ip netns exec container ip link set lo up
#ip link up container0
#ip netns exec container wg-quick up container0
#] ip link add container0 type wireguard
#] wg setconf container0 /dev/fd/63
#] ip -4 address add 10.0.10.3 dev container0
#] ip link set mtu 1420 up dev container0
#] wg set container0 fwmark 51820
#] ip -4 route add 0.0.0.0/0 dev container0 table 51820
#] ip -4 rule add not fwmark 51820 table 51820
#] ip -4 rule add table main suppress_prefixlength 0
#] sysctl -q net.ipv4.conf.all.src_valid_mark=1
#] iptables-restore -n
#] ip link set lo up

