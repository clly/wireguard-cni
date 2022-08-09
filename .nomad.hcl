plugin "docker" {
  config {
    allow_caps       = ["audit_write", "chown", "dac_override", "fowner", "fsetid", "kill", "mknod", "net_bind_service", "setfcap", "setgid", "setpcap", "setuid", "sys_chroot", "net_admin","sys_module","net_raw"]
    allow_privileged = true
  }
}

bind_addr = "0.0.0.0"

client {
    cni_config_dir = "/opt/cni/config"
    network_interface = "{{ GetDefaultInterfaces | limit 1 | attr \"name\" }}"
}
