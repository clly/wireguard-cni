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
            driver = "raw_exec"


            config {
                command = "/home/connor/p/wireguard-cni/bin/cmd/cluster-manager"
                args = ["--cidr-prefix","172.16.0.0/12"]
            }
        }
    }
}

