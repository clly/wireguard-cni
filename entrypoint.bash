#!/usr/bin/env bash

set -euo pipefail

bin=/opt/$1

if [[ -f $bin ]]; then
    echo "Starting $bin" >&2
    echo "whoami: $(whoami)"
    exec $bin
fi

echo "failed to start ${bin}"
exit 1
