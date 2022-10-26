# Analysis

Wireguard interfaces are namespace aware. After creation, they keep the
socket in the originating namespace (host namespace in this case) and then
forward traffic to the interface in the container namespace.

Wireguard UDP Listen Ports in the host namespace.
```commandline
UNCONN   0        0                   0.0.0.0:51820           0.0.0.0:*
UNCONN   0        0                   0.0.0.0:51821           0.0.0.0:*
UNCONN   0        0                      [::]:51820              [::]:*
UNCONN   0        0                      [::]:51821              [::]:*
```

```commandline
root@wg-server:/vagrant/bash-poc/c-c-vagrant# wg
interface: wg0
public key: v2XPnF7Yk7KEugBRVyCdD6mssm2Tsc4edDUApwzFoBk=
private key: (hidden)
listening port: 51820

peer: X7qQfHNDtiUS5+qg2H5T25IwhcWdVbyoFjGEDTFJ1WU=
endpoint: 192.168.56.11:51821
allowed ips: 10.0.10.3/32
latest handshake: 56 seconds ago
transfer: 532 B received, 4.75 KiB sent
persistent keepalive: every 10 seconds
```


With multiple listen ports we can have each container on a different UDP port. Those containers ultimately communicate
with each other through those UDP listen ports. Traffic that is received at the wireguard listen address in the host
namespace goes to the wireguard interface in the container namespace. We give each container Peer configuration that
points their Peers at the `address:port` combincation in the host namespace. This will slightly tighten the security
boundary because there is no longer a bridge to pass traffic in clear text. With the current firewall rules we will
still forward traffic if a user knows the Wireguard network space and where a wireguard interface lives.

We can not use wg-quick with this model because it creates, configures and brings up the interface in the container
namespace. Once the interface is live it cannot be moved from one namespace to another so we would not be able to bring
the interface up in the host namespace and move it to the container namespace.

## Convergence notes

### When are endpoints required
While being lazy during setup I've discovered configuring actual dialing and routing does _not seem_ to be required.
After one party performs any request against any other party in the mesh that is properly configured as a peer. The peer
can then respond back over the wireguard tunnel. Only one side _requires_ an endpoint which feels like it allows a level
of control and protection around which clients can start a connection. Unsure if those ever expire.

This was confirmed with tcpdump to ensure that packets were not passing through the host accidentally. If the route
didn't exist would the host pass traffic? If the peer exists and the route doesn't it appears that traffic won't pass?

### Hosts can still pass traffic but require a route

When a subnet is configured in AllowedIPs and a route is configured and IP forwarding is configured, the peer will still
pass traffic between nodes. As long as we configure routes properly on Hosts first new network namespaces should be able
to communicate between other network namespaces quickly.

This was confirmed using traceroute.

### iptables changes

A minor iptables change _was_ required to ensure that we don't NAT traffic intended for containers. If a container
starts the request then the host can pass traffic back to the container. Why didn't we find this out earlier? Unknown.
Maybe I started the request from inside the container so things just sort of worked or traffic could more easily get
passed into the container because of the connection from the bridge to the host-local IP address for the wireguard dial.

### When will a container pass traffic using a host.
When fully configured a container will _not_ attempt to pass traffic through it's host unless it does not have an
AllowedIPs for the address set on the interface.

## Testing
* open two terminals
* `vagrant ssh peer` and `vagrant ssh server`
* `./host-create.sh <peer|server>`
* `./container-create.sh <peer|server>`

Then ping or curl or whatever between them. Using something like echo-server would make it easier to validate