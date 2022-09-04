# Analysis

Wireguard interfaces are namespace aware. After creation, they keep the 
socket in the originating namespace (host namespace in this case) and then 
forward traffic to the interface in the container namespace.

Listening UDP Ports
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


With multiple listen ports we can have each container on a different UDP 
port and then configuration Peers for each `address:port` combination for each
wireguard interface. This will slightly tighten the security boundary 
because there is no longer a bridge to pass traffic in clear text. With the 
current firewall rules we will still forward traffic if a user knows the 
Wireguard network space and where a wireguard interface lives. 

We can not use wg-quick with this model because it creates, configures and 
brings up the interface at the same time. Once the interface is live it 
cannot be moved from one namespace to another.

## Testing 
* open two terminals
* `vagrant ssh peer` and `vagrant ssh server`
* `./host-create.sh <peer|server>`
* `./container-create.sh <peer|server>`

Then ping or curl or whatever between them. Using something like echo-server would make it easier to validate