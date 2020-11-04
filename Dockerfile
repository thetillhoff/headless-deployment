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
RUN apt-get install -y wget dnsmasq nginx
## retrieve files from previous stages
COPY --from=ipxe-builder /home/ipxe.efi /netboot/
COPY --from=ipxe-builder /home/undionly.kpxe /netboot/
COPY --from=go-builder /home/main /home/
## retrieve static files
COPY ./dnsmasq.conf /home/
COPY ./nginx.conf /etc/nginx/sites-enabled/default
COPY ./start.sh /home/
COPY ./templates/* /home/
## final configuration
RUN chmod +x /home/start.sh
EXPOSE 53/udp
EXPOSE 69/udp
EXPOSE 69/tcp
## startup command
CMD /home/start.sh && service nginx start && dnsmasq -C /etc/dnsmasq.conf && cd /home/; /home/main
