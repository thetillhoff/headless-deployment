# headless-deployment

## what this is about
To pack all required tools for automated deployment of a cluster infrastructure, this repository was created. It includes a pxe server in dns-proxy mode
## how to use

## prerequisites
- install docker
- install go-task (https://taskfile.dev)
- run ```task init```
- ensure the 'to-be-pxe-booted' host's BIOS is set to boot from network first. Debian will add itself to bootmenu too during installation. Afterwards recheck the bootmenu and make sure network is still number one and debian listed second. Disable every other bootdevice. This may be done automatically in the future (#TODO).


## settings (via environment variables)
SUBNET=
USER=

## timings
To get a feeling for when a problem occured, here are some measured times for a bare-metal-deployment of a single machine:
- from pressing the boot button until accessible: ~ 7-15min
