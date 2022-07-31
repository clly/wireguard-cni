job "sleep" {
    type = "service"
    datacenters = ["dc1"]

    group "sleep" {
        count = 1

        network {
            mode = "cni/wgnet"
        }
        task "sleep" {
            driver = "docker"

            config {
                image = "clly/debug"
                cap_add = ["net_raw"] // for pings
                args = ["sleep", "3600"]
            }
        }
    }
}

