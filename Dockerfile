FROM golang:1.26-alpine AS builder
RUN apk add --no-cache gcc musl-dev
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /routevpn ./cmd/server

FROM alpine:3.23.3
RUN apk add --no-cache ca-certificates wireguard-tools iptables
RUN adduser -D -h /app appuser
WORKDIR /app
COPY --from=builder /routevpn .
COPY templates/ ./templates/
COPY static/ ./static/
RUN chown -R appuser:appuser /app
EXPOSE 3000
CMD ["./routevpn"]
