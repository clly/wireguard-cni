job "sleep" {
    type = "service"
    datacenters = ["dc1"]

    group "sleep" {
        count = 4

        network {
            mode = "cni/wgnet"
            port "http" {
                to = "8080"
            }
        }
        service {
            provider = "nomad"
            port = "http"
            address_mode = "alloc"
            task = "sleep"
        }
        task "sleep" {
            driver = "docker"
            resources {
                memory = 10
            }
            config {
                image = "clly/debug"
                cap_add = ["net_raw"] // for pings
                args = ["sleep","3600"]
            }
        }
    }
}

