#!/usr/bin/bash

set -euo pipefail

N=$1
cp $N-wg0.conf /etc/wireguard/wg0.conf
wg-quick up wg0
