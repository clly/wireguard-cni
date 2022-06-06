#!/usr/bin/env bash

set -euo pipefail
set -x
[[ $UID == 0 ]] || { echo "You must be root to run this."; exit 1; }
NAME=$1
ADDRESS=$2
mkdir -p $NAME
wg genkey > $NAME/$NAME.key
wg pubkey <$NAME/$NAME.key > $NAME/$NAME.pub
set -x
ip netns add $NAME || true
ip link del dev $NAME 2>/dev/null || true
ip -link add dev $NAME type wireguard
ip link set $NAME netns $NAME


ip netns exec $NAME wg set $NAME private-key $NAME/$NAME.key listen-port 0
ip -n $NAME address add "${ADDRESS}/24" dev $NAME
ip -n $NAME link set $NAME up
ip -n $NAME link set lo up
#ip -n $NAME route add default dev ns1
listen_port=$(ip netns exec $NAME wg show $NAME listen-port)
cat <<<EOF > /etc/wireguard/$NAME.conf
[Interface]
PrivateKey = $(cat $NAME/$NAME.key)
Address = 10.8.0.1/24
ListenPort = $listen_port
EOF
