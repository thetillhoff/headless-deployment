- bind volume on docker for netboot files (like unpacked debian netboot image)
- docker image
  - username as env var
  - check if netboot data exists:
    - if yes, skip download
    - if no, download, unpack, delete archive
  - needs access to port 51
  - contains start.sh
    - sets dhcp range in subnet,
    sets dhcp-boot ip address
  - ENTRYPOINT /bin/bash -c
  - CMD /start.sh
  -> docker run --name headless-installer --rm -v $PWD/netinstall:/netinstall headless-installer

- build-docker image
  ? one for go -> if, then as multistage dockerfile for main docker image
  - one for ipxe -> as multistage dockerfile for main docker image


- go executable
  - (re-)installs _every_ machine that wants to netboot during uptime
    - requires a preseed.tmpl file
    - sets a predefined username
    - sets a random password, stores it in a file, hashes it and uses it in the preseed template
    - ssh needs to be installed & enabled & the user must be allowed to ssh into the machine.
  - take care of race conditions when reading/writing to /hosts file

