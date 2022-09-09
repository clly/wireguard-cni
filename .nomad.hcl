plugin "docker" {
  config {
    allow_caps       = ["audit_write", "chown", "dac_override", "fowner", "fsetid", "kill", "mknod", "net_bind_service", "setfcap", "setgid", "setpcap", "setuid", "sys_chroot", "net_admin","sys_module","net_raw"]
    allow_privileged = true
  }
}

bind_addr = "0.0.0.0"

advertise {
  rpc = "{{ GetPrivateInterfaces | include \"network\" \"192.168.56.0/24\" | attr \"address\" }}"
  serf = "{{ GetPrivateInterfaces | include \"network\" \"192.168.56.0/24\" | attr \"address\" }}"
}

client {
    cni_config_dir = "/opt/cni/config"
    network_interface = "{{ GetPrivateInterfaces | include \"network\" \"192.168.56.0/24\" | attr \"name\" }}"
    host_network "public" {
      cidr = "10.0.2.15/24"
    }
}

