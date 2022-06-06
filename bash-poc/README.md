# Proof of Concepts

These are all bash based proof of concepts.

* experiments - These are all single host based container experiments. I could never route out of the container and
  needed a bridge and I didn't know how to do that manually so I switched to using CNI
* static-vagrant - This runs in a pair of hosts. I used Vagrant. Basically making sure I got wireguard working between hosts
* container-vagrant - This is the meat of the poc. This validates that I can create containers and then route between the containers and hosts over wireguard
