#!/usr/bin/env bash

N=$1
ADDR=$2
ip netns add vnet$N
ip link add veth$N type veth peer name vpeer$N
ip link set veth$N netns vnet$N
ip -n vnet$N addr add $ADDR/24 dev veth$N
ip -n vnet$N link set veth$N up
ip -n vnet$N link set lo up
ip -n vnet$N addr show


