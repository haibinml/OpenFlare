FROM golang:1.25-alpine AS builder

WORKDIR /build

COPY openflare_relay/go.mod openflare_relay/go.sum ./
# Note: In a real monorepo build, openflare and openflare_server need to be available. 
# For now, we assume this is built from the root of the project where both are present.
COPY openflare_server /openflare_server
COPY openflare_relay /openflare_relay

WORKDIR /openflare_relay
RUN go build -o openflare-relay ./cmd/relay

FROM fatedier/frps:v0.69.0

COPY --from=builder /openflare_relay/openflare-relay /usr/local/bin/openflare-relay

VOLUME ["/var/lib/openflare-relay"]

ENTRYPOINT ["/usr/local/bin/openflare-relay", "-config", "/etc/openflare-relay/relay.json"]
CMD []
