plugin "docker" {
  config {
    allow_caps       = ["audit_write", "chown", "dac_override", "fowner", "fsetid", "kill", "mknod", "net_bind_service", "setfcap", "setgid", "setpcap", "setuid", "sys_chroot", "net_admin","sys_module","net_raw"]
    allow_privileged = true
  }
}

client {
    cni_config_dir = "/opt/cni/config"
}