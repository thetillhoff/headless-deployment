#!/bin/bash

## download netinst image if not exists
if test -f "/netboot/pxelinux.0"; then
  echo "Netboot files already exist."
else
  echo "Downloading netboot files..."
  wget -q http://ftp.debian.org/debian/dists/buster/main/installer-amd64/current/images/netboot/netboot.tar.gz -O /netboot/debian-netboot.tar.gz
  tar -xvzf /netboot/debian-netboot.tar.gz -C /netboot/
  rm /netboot/debian-netboot.tar.gz
  echo "Downloading netboot files finsished."
fi

## configure dnsmasq for local ip
sed -i "s%dhcp-boot=tag:ipxe,http://xxx.xxx.xxx.xxx/default%dhcp-boot=tag:ipxe,http://$(cut -d' ' -f1 <<<$(hostname -I))/default?mac=\${net0/mac:hexhyp}%" /etc/dnsmasq.conf
## configure dnsmasq's subnet
sed -i "s%dhcp-range=192.168.1.0,proxy%dhcp-range=$SUBNET,proxy%" /etc/dnsmasq.conf
