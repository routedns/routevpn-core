# RouteVPN Core

A self-hosted AmneziaWG VPN management platform with a reseller system built in Go. Admins create reseller accounts, top-up BDT balances, and define VPN packages. Resellers purchase peers for their clients, generating WireGuard configs, QR codes, and managing subscriptions — all through a REST API and web dashboard.

## Features

- **AmneziaWG support** — full obfuscation parameters (Jc, Jmin, Jmax, S1, S2, H1–H4)
- **Reseller system** — admins create resellers, top-up BDT balance; resellers purchase peers
- **Automatic IP allocation** — peers get the next available IP from the configured subnet
- **WireGuard key generation** — Curve25519 keypairs and preshared keys generated server-side
- **Config & QR download** — `.conf` file download and QR code PNG generation per peer
- **Expiry worker** — background job expires peers and removes them from the WireGuard interface
- **JWT authentication** — role-based access control (admin / reseller)
- **REST API** — 23 endpoints with consistent JSON envelope, pagination, and filtering
- **Web dashboard** — server-rendered HTML templates for admin and reseller panels
- **Valkey (Redis) caching** — session and rate-limit support via Valkey

## Tech Stack

| Component | Technology |
|-----------|------------|
| Language | Go 1.26 |
| Web Framework | [Fiber v3](https://gofiber.io/) |
| Database | PostgreSQL 16 |
| Cache | [Valkey 8](https://valkey.io/) (Redis-compatible) |
| VPN | AmneziaWG (WireGuard-compatible) |
| Auth | JWT (HS256) via [golang-jwt/jwt](https://github.com/golang-jwt/jwt) |
| Templates | Go HTML templates via [gofiber/template](https://github.com/gofiber/template) |
| Containerisation | Docker + Docker Compose |

## Prerequisites

- [Docker](https://docs.docker.com/get-docker/) & [Docker Compose](https://docs.docker.com/compose/install/)
- (Optional) [Go 1.26+](https://go.dev/dl/) for local development without Docker

## Quick Start

### 1. Clone the repository

```bash
git clone https://github.com/imzami/routevpn-core.git
cd routevpn-core
```

### 2. Configure environment

Copy the example environment file and edit as needed:

```bash
cp .env.example .env    # or edit the existing .env
```

### 3. Start the stack

```bash
docker compose up -d
```

This starts three containers:

| Container | Description | Port |
|-----------|-------------|------|
| `routevpn` | Go application | `3000` |
| `routevpn-db` | PostgreSQL 16 | — (internal) |
| `routevpn-cache` | Valkey 8 | — (internal) |

### 4. Verify

```bash
curl http://localhost:3000/api/v1/health
# {"success":true,"data":{"status":"ok"}}
```

### 5. Login

The default admin account is created on first boot from the `ADMIN_EMAIL` and `ADMIN_PASSWORD` env vars:

```bash
curl -X POST http://localhost:3000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "admin@routevpn.com", "password": "admin123"}'
```

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `3000` | HTTP server port |
| `DATABASE_URL` | — | PostgreSQL connection string |
| `VALKEY_URL` | — | Valkey (Redis) host:port |
| `JWT_SECRET` | — | Secret for signing JWT tokens. **Change in production.** |
| `ADMIN_EMAIL` | — | Default admin email (created on first boot) |
| `ADMIN_PASSWORD` | — | Default admin password |
| `WG_INTERFACE` | `awg0` | AmneziaWG network interface name |
| `WG_ENDPOINT` | — | Public VPN endpoint (`host:port`) for client configs |
| `WG_LISTEN_PORT` | `51820` | WireGuard listen port |
| `WG_DNS` | `1.1.1.1, 8.8.8.8` | DNS servers for client configs |
| `WG_SUBNET` | `10.66.66.0/24` | Subnet for peer IP allocation |
| `SERVER_PRIVATE_KEY` | — | WireGuard server private key |
| `SERVER_PUBLIC_KEY` | — | WireGuard server public key |
| `AWG_JC` | `4` | AmneziaWG Junk packet Count |
| `AWG_JMIN` | `40` | AmneziaWG Junk packet Min size |
| `AWG_JMAX` | `70` | AmneziaWG Junk packet Max size |
| `AWG_S1` | `0` | AmneziaWG Init packet junk Size |
| `AWG_S2` | `0` | AmneziaWG Response packet junk Size |
| `AWG_H1` | `1` | AmneziaWG Init Header |
| `AWG_H2` | `2` | AmneziaWG Response Header |
| `AWG_H3` | `3` | AmneziaWG Under-load Header |
| `AWG_H4` | `4` | AmneziaWG Cookie Reply Header |
| `WG_DRY_RUN` | `false` | Skip real `wg` commands (for dev/testing) |

## Project Structure

```
routevpn-core/
├── cmd/server/main.go          # Application entrypoint
├── internal/
│   ├── config/config.go        # Environment configuration loader
│   ├── database/
│   │   ├── postgres.go         # PostgreSQL connection & pool
│   │   ├── migrations.go       # Auto-migrations on boot
│   │   └── valkey.go           # Valkey (Redis) client
│   ├── handlers/
│   │   ├── admin.go            # Admin route handlers (web)
│   │   ├── auth.go             # Auth route handlers (web)
│   │   └── reseller.go         # Reseller route handlers (web)
│   ├── middleware/auth.go      # JWT auth middleware
│   ├── models/models.go        # Database models & structs
│   ├── services/
│   │   ├── auth.go             # Authentication & JWT service
│   │   ├── container.go        # Dependency injection container
│   │   └── peer.go             # Peer CRUD & IP allocation
│   ├── wg/
│   │   ├── wg.go               # WireGuard CLI wrapper (add/remove peer)
│   │   └── config.go           # Client config & QR generation
│   └── worker/expiry.go        # Background peer expiry worker
├── api/
│   ├── routes.go               # REST API route definitions
│   ├── handlers.go             # REST API handlers
│   ├── middleware.go           # API-specific middleware
│   ├── auth.go                 # API auth handlers
│   └── response.go            # JSON response helpers
├── migrations/                 # SQL migration files
├── templates/                  # Go HTML templates (admin & reseller dashboards)
├── static/                     # Static assets (CSS, JS)
├── Dockerfile                  # Multi-stage Docker build
├── docker-compose.yml          # Full stack: app + PostgreSQL + Valkey
├── API.md                      # Detailed REST API documentation
└── .env                        # Environment variables
```

## API

The REST API is available at `/api/v1`. All endpoints return a consistent JSON envelope:

```json
{
  "success": true,
  "data": { },
  "meta": { "page": 1, "per_page": 25, "total": 42 }
}
```

### Endpoints Overview

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| `GET` | `/api/v1/health` | — | Health check |
| `POST` | `/api/v1/auth/login` | — | Login, returns JWT |
| `GET` | `/api/v1/auth/me` | Bearer | Current user profile |
| `GET` | `/api/v1/admin/stats` | Admin | Dashboard statistics |
| `GET` | `/api/v1/admin/resellers` | Admin | List resellers |
| `POST` | `/api/v1/admin/resellers` | Admin | Create reseller |
| `GET` | `/api/v1/admin/resellers/:id` | Admin | Reseller detail |
| `POST` | `/api/v1/admin/resellers/:id/balance` | Admin | Top-up reseller balance |
| `GET` | `/api/v1/admin/packages` | Admin | List packages |
| `POST` | `/api/v1/admin/packages` | Admin | Create package |
| `PUT` | `/api/v1/admin/packages/:id` | Admin | Update package |
| `DELETE` | `/api/v1/admin/packages/:id` | Admin | Delete package |
| `GET` | `/api/v1/admin/peers` | Admin | List all peers |
| `GET` | `/api/v1/admin/peers/:id` | Admin | Peer detail |
| `DELETE` | `/api/v1/admin/peers/:id` | Admin | Remove peer |
| `GET` | `/api/v1/admin/transactions` | Admin | List all transactions |
| `GET` | `/api/v1/reseller/stats` | Reseller | Reseller dashboard stats |
| `GET` | `/api/v1/reseller/packages` | Reseller | Available packages |
| `GET` | `/api/v1/reseller/peers` | Reseller | List own peers |
| `POST` | `/api/v1/reseller/peers` | Reseller | Create peer |
| `GET` | `/api/v1/reseller/peers/:id/config` | Reseller | Download AmneziaWG config |
| `GET` | `/api/v1/reseller/peers/:id/qr` | Reseller | Download QR code PNG |
| `GET` | `/api/v1/reseller/transactions` | Reseller | Transaction history |

> 📄 See **[API.md](API.md)** for full documentation with request/response examples.

## Development

### Run locally without Docker

```bash
# Start PostgreSQL and Valkey
docker compose up -d postgres valkey

# Update .env with localhost connection strings
# DATABASE_URL=postgres://routevpn:routevpn@localhost:5432/routevpn?sslmode=disable
# VALKEY_URL=localhost:6379

# Run the app
go run ./cmd/server
```

### Rebuild after code changes

```bash
docker compose down
docker compose build --no-cache
docker compose up -d
```

### View logs

```bash
docker logs -f routevpn
```

### Reset database

```bash
docker compose down -v   # removes volumes (all data)
docker compose up -d
```

## Production Notes

1. **Change `JWT_SECRET`** to a long random string
2. **Change `ADMIN_PASSWORD`** to a strong password
3. **Set `WG_DRY_RUN=false`** (or remove it) to enable real WireGuard commands
4. **Set `WG_ENDPOINT`** to your server's public IP/domain
5. **Generate WireGuard keys** for production:
   ```bash
   wg genkey | tee privatekey | wg pubkey > publickey
   ```
6. Run the container with `NET_ADMIN` capability and `net.ipv4.ip_forward=1` (already configured in `docker-compose.yml`)
7. Ensure the AmneziaWG interface (`awg0`) is created on the host before starting the app

## License

MIT