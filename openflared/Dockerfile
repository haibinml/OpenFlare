FROM golang:1.25-alpine AS builder

WORKDIR /build

COPY openflared/go.mod openflared/go.sum ./
COPY openflare_server/go.mod openflare_server/go.sum ../openflare_server/
RUN go mod download

COPY openflare_server/ ../openflare_server/
COPY openflared/ .
RUN CGO_ENABLED=0 GOOS=linux go build -o flared ./cmd/flared

FROM fatedier/frpc:v0.69.0

WORKDIR /app

COPY --from=builder /build/flared .

ENV OPENFLARE_DATA_DIR=/app/data
ENV OPENFLARE_FRPC_PATH=/usr/bin/frpc

ENTRYPOINT ["/app/flared"]
CMD []
