job "cluster-manager" {
    type = "service"

    datacenters = ["dc1"]

    group "cluster-manager" {
        count = 1

        network {
            port "http" {
                static = "8080"
            }
            mode = "host"
        }
        task "cluster-manager" {
            driver = "docker"


            config {
                network_mode = "host"
                image = "wireguard-cni:local-1659155194"
                args = ["cluster-manager"]
            }
        }
    }
}

