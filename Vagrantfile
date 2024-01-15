# -*- mode: ruby -*-
# vi: set ft=ruby :

# All Vagrant configuration is done below. The "2" in Vagrant.configure
# configures the configuration version (we support older styles for
# backwards compatibility). Please don't change it unless you know what
# you're doing.
Vagrant.configure("2") do |config|
  # The most common configuration options are documented and commented below.
  # For a complete reference, please see the online documentation at
  # https://docs.vagrantup.com.

  # Every Vagrant development environment requires a box. You can search for
  # boxes at https://vagrantcloud.com/search.
  #config.env.enable # enable .env support plugin (it will let us easily enable cloud_init support)

  # URL used as a source for the vm.box defined above
  config.vm.box_url = "https://cloud-images.ubuntu.com/jammy/current/jammy-server-cloudimg-amd64-vagrant.box"

  config.vm.define "server" do |server|
    server.vm.box = "wg-server"
    server.vm.hostname = "wg-server"
    server.vm.network "private_network", ip: "192.168.56.11"
    server.vm.network "forwarded_port", guest: 4646, host: 14646, host_ip: "127.0.0.1"
    server.vm.network "forwarded_port", guest: 9002, host: 19002, host_ip: "127.0.0.1"
  end

  config.vm.define "peer" do |peer|
    peer.vm.box = "wg-peer"
    peer.vm.hostname = "wg-peer"
    peer.vm.network "private_network", ip: "192.168.56.10"
    peer.vm.network "forwarded_port", guest: 9002, host: 29002, host_ip: "127.0.0.1"
  end
  # Disable automatic box update checking. If you disable this, then
  # boxes will only be checked for updates when the user runs
  # `vagrant box outdated`. This is not recommended.
  # config.vm.box_check_update = false

  # Create a forwarded port mapping which allows access to a specific port
  # within the machine from a port on the host machine. In the example below,
  # accessing "localhost:8080" will access port 80 on the guest machine.
  # NOTE: This will enable public access to the opened port
  # config.vm.network "forwarded_port", guest: 80, host: 8080

  # Create a forwarded port mapping which allows access to a specific port
  # within the machine from a port on the host machine and only allow access
  # via 127.0.0.1 to disable public access


  # Create a private network, which allows host-only access to the machine
  # using a specific IP.


  # Create a public network, which generally matched to bridged network.
  # Bridged networks make the machine appear as another physical device on
  # your network.
  # config.vm.network "public_network"

  # Share an additional folder to the guest VM. The first argument is
  # the path on the host to the actual folder. The second argument is
  # the path on the guest to mount the folder. And the optional third
  # argument is a set of non-required options.
  # config.vm.synced_folder "../data", "/vagrant_data"

  # Provider-specific configuration so you can fine-tune various
  # backing providers for Vagrant. These expose provider-specific options.
  # Example for VirtualBox:
  #
  # config.vm.provider "virtualbox" do |vb|
  #   # Display the VirtualBox GUI when booting the machine
  #   vb.gui = true
  #
  #   # Customize the amount of memory on the VM:
  #   vb.memory = "1024"
  # end
  #
  # View the documentation for the provider you are using for more
  # information on available options.

#  config.vm.cloud_init do |cloud_init|
#    cloud_init.content_type = "text/cloud-config"
#    cloud_init.path = "../digitalocean/config/cloud-init"
#  end
  #config.vm.cloud_init do |cloud_init|
  #  cloud_init.content_type = "text/cloud-config"
  #  cloud_init.path = "terraform/user-data.yml"
  #end

  #config.vm.network "forwarded_port", guest: 4646, host: 4646
  # Enable provisioning with a shell script. Additional provisioners such as
  # Puppet, Chef, Ansible, Salt, and Docker are also available. Please see the
  # documentation for more information about their specific syntax and use.
  config.vm.provision "shell", inline: <<-SHELL
    apt-get update && apt-get install -y git wireguard-tools docker.io jq containernetworking-plugins make && apt-get upgrade -y
    cd /tmp &&  wget https://go.dev/dl/go1.21.6.linux-amd64.tar.gz && tar -C /usr/local -xzf go1.21.6.linux-amd64.tar.gz
    cd /vagrant
    PATH=/usr/local/go/bin:$PATH make extra-deps build
    ls -l /opt /opt/cni
    mkdir -p /opt/cni/config
    ln -snf /usr/lib/cni /opt/cni/bin
    cp bin/cmd/cni /opt/cni/bin/wireguard
    cp .wgnet.conflist /opt/cni/config/wgnet.conflist
    cp -f .bin/nomad /usr/local/bin
    mkdir -p /etc/nomad.d
    cp .nomad.hcl /etc/nomad.d/nomad.hcl
    cp config/nomad-systemd.service /etc/systemd/system/nomad.service

    mkdir -p /opt/nomad
    chmod +x /usr/local/bin/nomad
    if [[ $(hostnamectl --static) == "wg-peer" ]]; then
      cp config/client.hcl /etc/nomad.d/
      cat config/nomad-systemd.service | sed "s/-dev//" > /etc/systemd/system/nomad.service
      echo 192.168.56.11 wg-server | sudo tee -a /etc/hosts
    else
      echo 192.168.56.10 wg-peer | sudo tee -a /etc/hosts
    fi
    systemctl daemon-reload && systemctl start nomad
  SHELL
end
