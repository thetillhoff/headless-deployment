FROM tillhoff/debian AS ipxe-builder
## install required software
RUN apt-get install -y git binutils gcc liblzma-dev make mtools genisoimage perl syslinux
## set workdir
WORKDIR /home
## get ipxe src
RUN git clone --depth 1 -b v1.20.1 git://git.ipxe.org/ipxe.git
## startup command
RUN cd ipxe/src; \
  make bin-x86_64-efi/ipxe.efi; \
  mv bin-x86_64-efi/ipxe.efi /home/; \
  make bin/undionly.kpxe; \
  mv bin/undionly.kpxe /home/;


FROM tillhoff/debian AS go-builder
## install required software
RUN apt-get install -y wget git
## install required software
RUN wget -q https://dl.google.com/go/go1.14.2.linux-amd64.tar.gz -O /go.tar.gz && \
    tar -zxf /go.tar.gz -C / && \
    rm /go.tar.gz
## go get dependencies
RUN /go/bin/go get gopkg.in/yaml.v3
## copy src
COPY main.go /home/
## startup command
RUN cd /home/; /go/bin/go build -o main


FROM tillhoff/debian
## install required software
# whois contains mkpasswd, which is used for password generation
RUN apt-get install -y wget dnsmasq nginx whois pwgen
## retrieve files from previous stages
COPY --from=ipxe-builder /home/ipxe.efi /tftp/
COPY --from=ipxe-builder /home/undionly.kpxe /tftp/
COPY --from=go-builder /home/main /home/
## retrieve static files
COPY ./dnsmasq.conf /etc/dnsmasq.conf
COPY ./nginx.conf /etc/nginx/sites-enabled/default
COPY ./start.sh /home/
COPY ./templates/* /home/
## final configuration
RUN chmod +x /home/start.sh
## startup command
CMD /home/start.sh && service nginx start && dnsmasq -C /etc/dnsmasq.conf -u root --log-facility=/var/log/dnsmasq.log && cd /home/; /home/main
