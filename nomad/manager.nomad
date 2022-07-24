job "cni-manager" {
    type = "service"

    group "cluster-manager" {
        count = 1

        task "cluster-manager" {
            driver = "exec"

            config {
                command = "./bin/cmd/cluster-manager --cidr-prefix="172.16.0.0/12"
            }
        }
    }
}

