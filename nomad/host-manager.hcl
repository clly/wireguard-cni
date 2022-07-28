job "node-manager" {
    type = "system"
    datacenters = ["dc1"]

    group "host-manager" {
        count = 1

        network {
            port "http" {
                static = "5242"
            }
            mode = "host"
        }
        task "cluster-manager" {
            driver = "docker"

            user = "root"
            config {
                network_mode = "host"
                cap_add = ["net_admin","sys_module"]
                image = "wireguard-cni:local-1658981062"
                args = ["node-manager"]
            }
        }
    }
}

