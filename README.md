# Wireguard CNI

## Sequence Diagram

```mermaid
sequenceDiagram
    participant Cluster Manager
    participant Node Manager
    participant CNI  
    Node Manager->>Cluster Manager: Allocate Subnet
    Node Manager->>Cluster Manager: Register WG Public Key
    Node Manager->>Cluster Manager: Request Peers
    Cluster Manager->>Node Manager: Return All other cluster-peers (other node managers _or_ external clients dialing into the wireguard network)
    loop Manage Routes
        Node Manager->>Node Manager: Setup Routes for Cluster Peers and CNI Peers 
    end
    CNI->>Node Manager: Allocate IP Address
    CNI->>Node Manager: Register Public Key
    CNI->>Node Manager: Request Peers
    Node Manager->>CNI: Return Node-Manager Peer (include other allocs if CNI <-> CNI communication is enabled)
```


# Proof of Concept

Initial proof of concepts exist in the bash-poc. Read README's there for more instructions. All POC's should be started
by running `vagrant up`

# Simple CNI plugin

This is an example of a sample chained plugin. It includes solutions for some
of the more subtle cases that can be experienced with multi-version chained
plugins.

To use it, just add your code to the cmdAdd and cmdDel plugins.
