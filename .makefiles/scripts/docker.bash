#!/usr/bin/env bash

set -euo pipefail

tag=local-$(date +%s)

docker build -t wireguard-cni:$tag .

echo "Finished building $tag"
