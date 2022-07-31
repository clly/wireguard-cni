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
                image = "clly/wireguard-cni:v0.0.1"
                args = ["cluster-manager"]
            }
        }
    }
}

