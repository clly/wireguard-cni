FROM golang:1.21 as build

WORKDIR /build
COPY . ./
RUN make build

FROM ubuntu:jammy

WORKDIR /opt
COPY --from=build /build/bin/cmd/ ./
COPY entrypoint.bash /entrypoint.bash


RUN apt-get update && apt-get install -y \
    wireguard-tools jq iproute2 iptables \
 && rm -rf /var/lib/apt/lists/*

USER nobody
ENTRYPOINT ["/entrypoint.bash"]

