FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /routevpn ./cmd/server

# Build amneziawg-tools (awg)
FROM alpine:3.23 AS awg-builder
RUN apk add --no-cache git build-base linux-headers bash
RUN git clone --depth 1 https://github.com/amnezia-vpn/amneziawg-tools.git /awg-tools
WORKDIR /awg-tools/src
RUN make

# Build amneziawg-go (userspace - no kernel module needed)
FROM golang:1.26-alpine AS awg-go-builder
RUN apk add --no-cache git
RUN git clone --depth 1 https://github.com/amnezia-vpn/amneziawg-go.git /awg-go
WORKDIR /awg-go
RUN go build -o /amneziawg-go

# Final image
FROM alpine:3.23
RUN apk add --no-cache ca-certificates iptables iptables-legacy curl tzdata bash iproute2

# Copy amneziawg tools
COPY --from=awg-builder /awg-tools/src/wg /usr/bin/awg
COPY --from=awg-builder /awg-tools/src/wg-quick/linux.bash /usr/bin/awg-quick

# Copy amneziawg-go userspace daemon
COPY --from=awg-go-builder /amneziawg-go /usr/bin/amneziawg-go

RUN chmod +x /usr/bin/awg /usr/bin/awg-quick /usr/bin/amneziawg-go \
    && ln -s /usr/bin/amneziawg-go /usr/bin/wireguard-go

WORKDIR /app
COPY --from=builder /routevpn .
COPY templates/ ./templates/
COPY static/ ./static/

EXPOSE 3000
CMD ["./routevpn"]
