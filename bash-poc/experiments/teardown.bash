#!/usr/bin/env bash

for i in $(ls -d --color=never *); do 
    if [[ -d $i ]]; then
        ip -n $i link set $i down
        ip link del $i
        ip -n $i link del $i
        for ns in $(ip netns list|cut -f1 -d ' '); do
            echo $ns
            ip netns delete $ns
        done
    fi
done
