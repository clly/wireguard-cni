job "countdash" {
  datacenters = ["dc1"]

  group "api" {
    network {
      mode = "cni/wgnet"
    }

    service {
        provider = "nomad"
        address_mode = "alloc"
        name = "count-api"

        task = "web"
        port = "9001"
    }

    task "web" {
      driver = "docker"

      config {
        image = "hashicorpdev/counter-api:v3"
      }
    }
    task "debug" {
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

  group "dashboard" {
    network {
      mode = "bridge"

      port "http" {
        static = 9002
        to     = 9002
        host_network = "public"
      }
    }

    service {
      name = "count-dashboard"
      port = "http"
      provider = "nomad"
      address_mode = "alloc"
      task = "dashboard"

    }

    task "dashboard" {
      driver = "docker"

      template {
          data = <<EOL
COUNTING_SERVICE_URL={{ range nomadService "count-api" }}http://{{ .Address }}:{{ .Port }} {{ end }}
          EOL
          env = true
          destination = "local/env"
          change_mode = "restart"
      }

      config {
        image = "hashicorpdev/counter-dashboard:v3"
      }
    }

    task "debug" {
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
