#!/usr/bin/env bash

set -eof pipefail



sudo ./create.bash ns1 10.0.0.1
sudo ./create.bash ns2 10.0.0.2
sudo ./create.bash ns3 10.0.0.3


for i in $(ls -d --color=never ns*/*.pub); do
	sudo ip netns exec ns1 wg set peer $i
	sudo ip netns exec ns2 wg set peer $i
	sudo ip netns exec ns3 wg set peer $i
done
#for-each-peer ip netns exec wg set  peer $(cat $i/$i.pub)
#set -x
#for i in $(ls -d --color=never *); do
#    if [[ -d $i ]]; then
#        ip netns exec $i wg set $i peer $(cat $i/$i.pub)
#    fi
#done


