Wireugard-CNI Codebase Documentation
===

This directory contains some documentation about the wireguard-cni codebase,
aimed at readers who are interested in making code contributions.

If you're looking for information on _using_ wireguard-cni, please refer
to the main README instead.

* cmd - CNI, cluster-manager, and node-manager runtimes
* gen - generated code
* pkg/server - wireguard and ipam server implementations
* pkg/wireguard - creating, syncing, and destroying wireguard interfaces
* wgcni - proto definitions

## Creating new APIs
All APIs are defined in proto files and generated using buf. The dependencies can be installed using `make deps`. Modify
the definitions in wgcni and then generate the new code using `make proto`. Then implement the changes in the
pkg/server. You can validate that the implementation still passes tests using `make test`

## Testing
* Unit tests can be executed using `make test`

## Developing with Vagrant
* Install [Virtualbox](https://www.virtualbox.org/)
* Install dependencies with `make extra-deps`
* Build binaries with `make`
* ssh to vagrant machine
** `vagrant ssh server`
* Execute binaries directly 
** `/vagrant/bin/cmd/cluster-manager`
** `/vagrant/bin/cmd/node-manager`
** `cd /vagrant/bash-poc/wg-cni && sudo ./container-create.bash server`
** Use logs and other output to examine outputs

## Developing with Vagrant and Nomad

* Install [Virtualbox](https://www.virtualbox.org/)
* Install dependencies with `make extra-deps`
* Build binaries with `make`
* Build local docker container
* Set nomad configuration with the local docker tag
* start nomad jobs with `nomad run cluster-manager.hcl`, `nomad run host-manager.hcl`, `nomad run sleep.hcl`
