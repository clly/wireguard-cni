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
        task "node-manager" {
            driver = "docker"

            user = "root"
            config {
                network_mode = "host"
                cap_add = ["net_admin","sys_module","net_raw"] // net_admin, net_raw for iptables, sys_module for loading wireguard if necessary.
                image = "clly/wireguard-cni:v0.0.2"
                args = ["node-manager"]
            }


            template {
                data = <<EOL
CLUSTER_MANAGER_ADDR={{ range nomadService "wireguard-cluster-manager" }}http://{{ .Address }}:{{ .Port }} {{ end }}
                EOL
                env = true
                destination = "local/env"
            }
        }
    }
}

