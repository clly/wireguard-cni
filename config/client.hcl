data_dir = "/var/run/nomad"
client {
    enabled = true
    server_join {
        retry_join = [ "192.168.56.11" ]
    }
}
