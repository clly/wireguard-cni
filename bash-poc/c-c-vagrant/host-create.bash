#!/usr/bin/bash

set -euo pipefail

N=$1
cp $N-wg0.conf /etc/wireguard/wg0.conf
sysctl -w net.ipv4.ip_forward=1
wg-quick up wg0
