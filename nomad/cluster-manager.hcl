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
        service {
            name = "wireguard-cluster-manager"
            provider = "nomad"
            port = "http"
            task = "cluster-manager"
        }
        task "cluster-manager" {
            driver = "docker"

            config {
                network_mode = "host"
                image = "clly/wireguard-cni:v0.0.4"
                args = ["cluster-manager","-cidr-prefix","172.16.0.0/12"]
            }
        }
    }
}

