#!/usr/bin/env bash

set -euf pipefail

cp ../../bin/cmd/cni /usr/lib/cni/wireguard
./container-teardown.bash container || true
./container-create.bash container
