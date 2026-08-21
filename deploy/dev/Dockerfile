FROM golang:1.26-bookworm AS builder

WORKDIR /src
COPY go.mod ./
COPY go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/nodara-core ./cmd/nodara-core \
    && CGO_ENABLED=0 go build -o /out/nodara-agent ./cmd/nodara-agent

FROM debian:bookworm-slim

RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates openssl \
    && rm -rf /var/lib/apt/lists/*

COPY --from=builder /out/nodara-core /usr/local/bin/nodara-core
COPY --from=builder /out/nodara-agent /usr/local/bin/nodara-agent
COPY --from=builder /src/db/migrations /app/db/migrations
COPY scripts/dev-certs.sh /usr/local/bin/dev-certs.sh
RUN chmod 0755 /usr/local/bin/dev-certs.sh

WORKDIR /app
