# RouteVPN Core — REST API Documentation

**Base URL:** `http://localhost:3000/api/v1`  
**Content-Type:** `application/json`  
**Authentication:** Bearer JWT (`Authorization: Bearer <token>`)

---

## Response Envelope

All responses follow a consistent JSON envelope:

```json
{
  "success": true,
  "data": { ... },
  "error": null,
  "meta": null
}
```

### On Error

```json
{
  "success": false,
  "error": {
    "code": 401,
    "message": "invalid credentials"
  }
}
```

### With Pagination

```json
{
  "success": true,
  "data": [ ... ],
  "meta": {
    "page": 1,
    "per_page": 25,
    "total": 42
  }
}
```

---

## Table of Contents

| # | Endpoint | Method | Auth | Section |
|---|----------|--------|------|---------|
| 1 | `/api/v1/health` | GET | — | [Health](#1-health-check) |
| 2 | `/api/v1/auth/login` | POST | — | [Auth](#2-login) |
| 3 | `/api/v1/auth/me` | GET | Bearer | [Auth](#3-get-current-user) |
| 4 | `/api/v1/admin/stats` | GET | Admin | [Admin](#4-admin-dashboard-stats) |
| 5 | `/api/v1/admin/resellers` | GET | Admin | [Admin](#5-list-resellers) |
| 6 | `/api/v1/admin/resellers` | POST | Admin | [Admin](#6-create-reseller) |
| 7 | `/api/v1/admin/resellers/:id` | GET | Admin | [Admin](#7-get-reseller-detail) |
| 8 | `/api/v1/admin/resellers/:id/balance` | POST | Admin | [Admin](#8-add-reseller-balance) |
| 9 | `/api/v1/admin/packages` | GET | Admin | [Admin](#9-list-packages-admin) |
| 10 | `/api/v1/admin/packages` | POST | Admin | [Admin](#10-create-package) |
| 11 | `/api/v1/admin/packages/:id` | PUT | Admin | [Admin](#11-update-package) |
| 12 | `/api/v1/admin/packages/:id` | DELETE | Admin | [Admin](#12-delete-package) |
| 13 | `/api/v1/admin/peers` | GET | Admin | [Admin](#13-list-all-peers) |
| 14 | `/api/v1/admin/peers/:id` | GET | Admin | [Admin](#14-get-peer-detail) |
| 15 | `/api/v1/admin/peers/:id` | DELETE | Admin | [Admin](#15-remove-peer) |
| 16 | `/api/v1/admin/transactions` | GET | Admin | [Admin](#16-list-all-transactions) |
| 17 | `/api/v1/reseller/stats` | GET | Reseller | [Reseller](#17-reseller-dashboard-stats) |
| 18 | `/api/v1/reseller/packages` | GET | Reseller | [Reseller](#18-list-packages-reseller) |
| 19 | `/api/v1/reseller/peers` | GET | Reseller | [Reseller](#19-list-own-peers) |
| 20 | `/api/v1/reseller/peers` | POST | Reseller | [Reseller](#20-create-peer) |
| 21 | `/api/v1/reseller/peers/:id/config` | GET | Reseller | [Reseller](#21-download-peer-config) |
| 22 | `/api/v1/reseller/peers/:id/qr` | GET | Reseller | [Reseller](#22-get-peer-qr-code) |
| 23 | `/api/v1/reseller/transactions` | GET | Reseller | [Reseller](#23-list-own-transactions) |

---

## Public Endpoints

### 1. Health Check

Check API and database availability.

- **URL:** `/api/v1/health`
- **Method:** `GET`
- **Auth:** None

**Request:**

```bash
curl http://localhost:3000/api/v1/health
```

**Response `200 OK`:**

```json
{
  "success": true,
  "data": {
    "status": "ok"
  }
}
```

**Response `503 Service Unavailable`:**

```json
{
  "success": false,
  "error": {
    "code": 503,
    "message": "database unreachable"
  }
}
```

---

### 2. Login

Authenticate with email and password. Returns a Bearer JWT valid for 24 hours.

- **URL:** `/api/v1/auth/login`
- **Method:** `POST`
- **Auth:** None

**Request Body:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `email` | string | ✅ | User email address |
| `password` | string | ✅ | User password |

**Request:**

```bash
curl -X POST http://localhost:3000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "admin@routevpn.local", "password": "admin123"}'
```

**Response `200 OK`:**

```json
{
  "success": true,
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxLCJlbWFpbCI6ImFkbWluQHJvdXRldnBuLmxvY2FsIiwicm9sZSI6ImFkbWluIiwiZXhwIjoxNzc2NDUwNzk0LCJpYXQiOjE3NzYzNjQzOTR9.uILFS1MXG5U5jBBARhDJv_WVOP015m-tdIBJ4fML3ag",
    "expires_at": "2026-04-17T18:33:14Z",
    "role": "admin",
    "email": "admin@routevpn.local"
  }
}
```

**Response `400 Bad Request`:**

```json
{
  "success": false,
  "error": { "code": 400, "message": "email and password are required" }
}
```

**Response `401 Unauthorized`:**

```json
{
  "success": false,
  "error": { "code": 401, "message": "invalid credentials" }
}
```

---

### 3. Get Current User

Returns the profile of the currently authenticated user.

- **URL:** `/api/v1/auth/me`
- **Method:** `GET`
- **Auth:** Bearer (any role)

**Headers:**

| Header | Value |
|--------|-------|
| `Authorization` | `Bearer <token>` |

**Request:**

```bash
curl -H "Authorization: Bearer <token>" \
  http://localhost:3000/api/v1/auth/me
```

**Response `200 OK`:**

```json
{
  "success": true,
  "data": {
    "id": 1,
    "email": "admin@routevpn.local",
    "role": "admin",
    "balance": "0.00",
    "created_at": "2026-04-16 17:50:01.065986+00"
  }
}
```

**Response `401 Unauthorized`:**

```json
{
  "success": false,
  "error": { "code": 401, "message": "missing authorization header" }
}
```

---

## Admin Endpoints

> All admin endpoints require `Authorization: Bearer <token>` where the token belongs to a user with `role: "admin"`.

### 4. Admin Dashboard Stats

Returns aggregated platform statistics.

- **URL:** `/api/v1/admin/stats`
- **Method:** `GET`
- **Auth:** Admin

**Request:**

```bash
curl -H "Authorization: Bearer <admin_token>" \
  http://localhost:3000/api/v1/admin/stats
```

**Response `200 OK`:**

```json
{
  "success": true,
  "data": {
    "total_resellers": 3,
    "total_peers": 1,
    "active_peers": 1,
    "expired_peers": 0,
    "total_revenue_bdt": "200.00",
    "total_credits_bdt": "1000.00"
  }
}
```

---

### 5. List Resellers

Returns a paginated list of all resellers.

- **URL:** `/api/v1/admin/resellers`
- **Method:** `GET`
- **Auth:** Admin

**Query Parameters:**

| Param | Type | Default | Description |
|-------|------|---------|-------------|
| `page` | int | `1` | Page number |
| `per_page` | int | `25` | Items per page (max 100) |

**Request:**

```bash
curl -H "Authorization: Bearer <admin_token>" \
  "http://localhost:3000/api/v1/admin/resellers?page=1&per_page=10"
```

**Response `200 OK`:**

```json
{
  "success": true,
  "data": [
    {
      "id": 4,
      "email": "dealer@test.com",
      "role": "reseller",
      "balance": "0",
      "created_by": 1,
      "created_at": "2026-04-16T18:33:29.325305Z"
    },
    {
      "id": 3,
      "email": "reseller1@test.com",
      "role": "reseller",
      "balance": "5000",
      "created_by": 1,
      "created_at": "2026-04-16T18:28:22.341091Z"
    },
    {
      "id": 2,
      "email": "hi@imzami.com",
      "role": "reseller",
      "balance": "6000",
      "created_by": 1,
      "created_at": "2026-04-16T17:53:55.813229Z"
    }
  ],
  "meta": {
    "page": 1,
    "per_page": 25,
    "total": 3
  }
}
```

---

### 6. Create Reseller

Create a new reseller account.

- **URL:** `/api/v1/admin/resellers`
- **Method:** `POST`
- **Auth:** Admin

**Request Body:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `email` | string | ✅ | Reseller email (must be unique) |
| `password` | string | ✅ | Password (min 6 characters) |

**Request:**

```bash
curl -X POST -H "Authorization: Bearer <admin_token>" \
  -H "Content-Type: application/json" \
  -d '{"email": "new_reseller@example.com", "password": "secret123"}' \
  http://localhost:3000/api/v1/admin/resellers
```

**Response `201 Created`:**

```json
{
  "success": true,
  "data": {
    "id": 3,
    "email": "reseller1@test.com",
    "role": "reseller",
    "balance": "0",
    "created_by": 1,
    "created_at": "2026-04-16T18:28:22.341091Z"
  }
}
```

**Response `400 Bad Request` (duplicate email):**

```json
{
  "success": false,
  "error": {
    "code": 400,
    "message": "failed to create reseller: pq: duplicate key value violates unique constraint \"users_email_key\" (23505)"
  }
}
```

**Response `400 Bad Request` (validation):**

```json
{
  "success": false,
  "error": { "code": 400, "message": "email required, password min 6 characters" }
}
```

---

### 7. Get Reseller Detail

Returns a single reseller with peer count and total spend.

- **URL:** `/api/v1/admin/resellers/:id`
- **Method:** `GET`
- **Auth:** Admin

**Path Parameters:**

| Param | Type | Description |
|-------|------|-------------|
| `id` | int | Reseller user ID |

**Request:**

```bash
curl -H "Authorization: Bearer <admin_token>" \
  http://localhost:3000/api/v1/admin/resellers/3
```

**Response `200 OK`:**

```json
{
  "success": true,
  "data": {
    "id": 3,
    "email": "reseller1@test.com",
    "role": "reseller",
    "balance": "4800",
    "created_by": 1,
    "created_at": "2026-04-16T18:28:22.341091Z",
    "peer_count": 1,
    "total_spent_bdt": "200.00"
  }
}
```

**Response `404 Not Found`:**

```json
{
  "success": false,
  "error": { "code": 404, "message": "reseller not found" }
}
```

---

### 8. Add Reseller Balance

Top-up a reseller's BDT balance and record a credit transaction.

- **URL:** `/api/v1/admin/resellers/:id/balance`
- **Method:** `POST`
- **Auth:** Admin

**Path Parameters:**

| Param | Type | Description |
|-------|------|-------------|
| `id` | int | Reseller user ID |

**Request Body:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `amount` | string | ✅ | Positive BDT amount (e.g. `"5000.00"`) |
| `note` | string | ❌ | Optional note (default: `"Admin balance top-up"`) |

**Request:**

```bash
curl -X POST -H "Authorization: Bearer <admin_token>" \
  -H "Content-Type: application/json" \
  -d '{"amount": "5000.00", "note": "Initial top-up"}' \
  http://localhost:3000/api/v1/admin/resellers/3/balance
```

**Response `200 OK`:**

```json
{
  "success": true,
  "data": {
    "id": 3,
    "email": "reseller1@test.com",
    "role": "reseller",
    "balance": "5000",
    "created_by": 1,
    "created_at": "2026-04-16T18:28:22.341091Z"
  }
}
```

**Response `400 Bad Request`:**

```json
{
  "success": false,
  "error": { "code": 400, "message": "amount must be a positive number" }
}
```

---

### 9. List Packages (Admin)

Returns all VPN packages ordered by duration.

- **URL:** `/api/v1/admin/packages`
- **Method:** `GET`
- **Auth:** Admin

**Request:**

```bash
curl -H "Authorization: Bearer <admin_token>" \
  http://localhost:3000/api/v1/admin/packages
```

**Response `200 OK`:**

```json
{
  "success": true,
  "data": [
    {
      "id": 5,
      "name": "30 Days",
      "price_bdt": "200",
      "duration_days": 30,
      "created_at": "2026-04-16T18:33:45.025446Z"
    }
  ]
}
```

---

### 10. Create Package

Create a new VPN subscription package.

- **URL:** `/api/v1/admin/packages`
- **Method:** `POST`
- **Auth:** Admin

**Request Body:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | ✅ | Package display name |
| `price_bdt` | string | ✅ | Price in BDT (e.g. `"200.00"`) |
| `duration_days` | int | ✅ | Duration in days (≥ 1) |

**Request:**

```bash
curl -X POST -H "Authorization: Bearer <admin_token>" \
  -H "Content-Type: application/json" \
  -d '{"name": "30 Days", "price_bdt": "200.00", "duration_days": 30}' \
  http://localhost:3000/api/v1/admin/packages
```

**Response `201 Created`:**

```json
{
  "success": true,
  "data": {
    "id": 5,
    "name": "30 Days",
    "price_bdt": "200",
    "duration_days": 30,
    "created_at": "2026-04-16T18:33:45.025446Z"
  }
}
```

**Response `400 Bad Request`:**

```json
{
  "success": false,
  "error": { "code": 400, "message": "name required, duration_days must be >= 1" }
}
```

---

### 11. Update Package

Update an existing package by ID.

- **URL:** `/api/v1/admin/packages/:id`
- **Method:** `PUT`
- **Auth:** Admin

**Path Parameters:**

| Param | Type | Description |
|-------|------|-------------|
| `id` | int | Package ID |

**Request Body:** Same as [Create Package](#10-create-package).

**Request:**

```bash
curl -X PUT -H "Authorization: Bearer <admin_token>" \
  -H "Content-Type: application/json" \
  -d '{"name": "1 Month Premium", "price_bdt": "250.00", "duration_days": 30}' \
  http://localhost:3000/api/v1/admin/packages/5
```

**Response `200 OK`:**

```json
{
  "success": true,
  "data": {
    "id": 5,
    "name": "1 Month Premium",
    "price_bdt": "250",
    "duration_days": 30,
    "created_at": "2026-04-16T18:33:45.025446Z"
  }
}
```

**Response `404 Not Found`:**

```json
{
  "success": false,
  "error": { "code": 404, "message": "package not found" }
}
```

---

### 12. Delete Package

Delete a package by ID. Blocked if any active peers use it.

- **URL:** `/api/v1/admin/packages/:id`
- **Method:** `DELETE`
- **Auth:** Admin

**Path Parameters:**

| Param | Type | Description |
|-------|------|-------------|
| `id` | int | Package ID |

**Request:**

```bash
curl -X DELETE -H "Authorization: Bearer <admin_token>" \
  http://localhost:3000/api/v1/admin/packages/5
```

**Response `200 OK`:**

```json
{
  "success": true,
  "data": { "deleted": true }
}
```

**Response `400 Bad Request` (has active peers):**

```json
{
  "success": false,
  "error": { "code": 400, "message": "cannot delete: 3 active peers use this package" }
}
```

---

### 13. List All Peers

Returns a paginated, filterable list of all VPN peers across all resellers.

- **URL:** `/api/v1/admin/peers`
- **Method:** `GET`
- **Auth:** Admin

**Query Parameters:**

| Param | Type | Default | Description |
|-------|------|---------|-------------|
| `page` | int | `1` | Page number |
| `per_page` | int | `25` | Items per page (max 100) |
| `q` | string | — | Search by WhatsApp number or public key |
| `status` | string | — | Filter by `active` or `expired` |
| `reseller_id` | int | — | Filter by reseller user ID |

**Request:**

```bash
curl -H "Authorization: Bearer <admin_token>" \
  "http://localhost:3000/api/v1/admin/peers?status=active&page=1&per_page=10"
```

**Response `200 OK`:**

```json
{
  "success": true,
  "data": [
    {
      "id": 7,
      "user_id": 3,
      "whatsapp_number": "+8801712345678",
      "public_key": "lggeiJiw6JVjeLTT/aLX+tD13B/ZMYUbFbbh9z73yXo=",
      "assigned_ip": "10.66.66.2/32",
      "expiry_date": "2026-05-16T18:44:29.912087Z",
      "status": "active",
      "package_id": 5,
      "created_at": "2026-04-16T18:44:29.897626Z",
      "reseller_email": "reseller1@test.com",
      "package_name": "30 Days"
    }
  ],
  "meta": {
    "page": 1,
    "per_page": 10,
    "total": 1
  }
}
```

---

### 14. Get Peer Detail

Returns a single peer with reseller email and package name.

- **URL:** `/api/v1/admin/peers/:id`
- **Method:** `GET`
- **Auth:** Admin

**Path Parameters:**

| Param | Type | Description |
|-------|------|-------------|
| `id` | int | Peer ID |

**Request:**

```bash
curl -H "Authorization: Bearer <admin_token>" \
  http://localhost:3000/api/v1/admin/peers/7
```

**Response `200 OK`:**

```json
{
  "success": true,
  "data": {
    "id": 7,
    "user_id": 3,
    "whatsapp_number": "+8801712345678",
    "public_key": "lggeiJiw6JVjeLTT/aLX+tD13B/ZMYUbFbbh9z73yXo=",
    "assigned_ip": "10.66.66.2/32",
    "expiry_date": "2026-05-16T18:44:29.912087Z",
    "status": "active",
    "package_id": 5,
    "created_at": "2026-04-16T18:44:29.897626Z",
    "reseller_email": "reseller1@test.com",
    "package_name": "30 Days"
  }
}
```

**Response `404 Not Found`:**

```json
{
  "success": false,
  "error": { "code": 404, "message": "peer not found" }
}
```

---

### 15. Remove Peer

Removes a peer from the WireGuard interface and sets its status to `expired`.

- **URL:** `/api/v1/admin/peers/:id`
- **Method:** `DELETE`
- **Auth:** Admin

**Path Parameters:**

| Param | Type | Description |
|-------|------|-------------|
| `id` | int | Peer ID |

**Request:**

```bash
curl -X DELETE -H "Authorization: Bearer <admin_token>" \
  http://localhost:3000/api/v1/admin/peers/7
```

**Response `200 OK`:**

```json
{
  "success": true,
  "data": {
    "removed": true,
    "peer_id": 7
  }
}
```

---

### 16. List All Transactions

Returns a paginated, filterable list of all transactions across all resellers.

- **URL:** `/api/v1/admin/transactions`
- **Method:** `GET`
- **Auth:** Admin

**Query Parameters:**

| Param | Type | Default | Description |
|-------|------|---------|-------------|
| `page` | int | `1` | Page number |
| `per_page` | int | `25` | Items per page (max 100) |
| `reseller_id` | int | — | Filter by reseller user ID |
| `type` | string | — | Filter by `credit` or `debit` |

**Request:**

```bash
curl -H "Authorization: Bearer <admin_token>" \
  "http://localhost:3000/api/v1/admin/transactions?type=debit&page=1"
```

**Response `200 OK`:**

```json
{
  "success": true,
  "data": [
    {
      "id": 10,
      "reseller_id": 3,
      "amount": "200",
      "type": "debit",
      "note": "Peer creation: 30 Days for +8801712345678",
      "created_at": "2026-04-16T18:44:29.897626Z",
      "reseller_email": "reseller1@test.com"
    }
  ],
  "meta": {
    "page": 1,
    "per_page": 25,
    "total": 1
  }
}
```

---

## Reseller Endpoints

> All reseller endpoints require `Authorization: Bearer <token>` where the token belongs to a user with `role: "reseller"`. Resellers can only access their own data.

### 17. Reseller Dashboard Stats

Returns the current reseller's balance, peer counts, and total spend.

- **URL:** `/api/v1/reseller/stats`
- **Method:** `GET`
- **Auth:** Reseller

**Request:**

```bash
curl -H "Authorization: Bearer <reseller_token>" \
  http://localhost:3000/api/v1/reseller/stats
```

**Response `200 OK`:**

```json
{
  "success": true,
  "data": {
    "balance_bdt": "4800.00",
    "total_peers": 1,
    "active_peers": 1,
    "total_spent_bdt": "200.00"
  }
}
```

---

### 18. List Packages (Reseller)

Returns all available VPN packages the reseller can purchase.

- **URL:** `/api/v1/reseller/packages`
- **Method:** `GET`
- **Auth:** Reseller

**Request:**

```bash
curl -H "Authorization: Bearer <reseller_token>" \
  http://localhost:3000/api/v1/reseller/packages
```

**Response `200 OK`:**

```json
{
  "success": true,
  "data": [
    {
      "id": 5,
      "name": "30 Days",
      "price_bdt": "200",
      "duration_days": 30,
      "created_at": "2026-04-16T18:33:45.025446Z"
    }
  ]
}
```

---

### 19. List Own Peers

Returns the reseller's own peers with pagination and search.

- **URL:** `/api/v1/reseller/peers`
- **Method:** `GET`
- **Auth:** Reseller

**Query Parameters:**

| Param | Type | Default | Description |
|-------|------|---------|-------------|
| `page` | int | `1` | Page number |
| `per_page` | int | `25` | Items per page (max 100) |
| `q` | string | — | Search by WhatsApp number or public key |
| `status` | string | — | Filter by `active` or `expired` |

**Request:**

```bash
curl -H "Authorization: Bearer <reseller_token>" \
  "http://localhost:3000/api/v1/reseller/peers?status=active"
```

**Response `200 OK`:**

```json
{
  "success": true,
  "data": [
    {
      "id": 7,
      "user_id": 3,
      "whatsapp_number": "+8801712345678",
      "public_key": "lggeiJiw6JVjeLTT/aLX+tD13B/ZMYUbFbbh9z73yXo=",
      "assigned_ip": "10.66.66.2/32",
      "expiry_date": "2026-05-16T18:44:29.912087Z",
      "status": "active",
      "package_id": 5,
      "created_at": "2026-04-16T18:44:29.897626Z",
      "package_name": "30 Days"
    }
  ],
  "meta": {
    "page": 1,
    "per_page": 25,
    "total": 1
  }
}
```

---

### 20. Create Peer

Create a new VPN peer. Deducts the package price from the reseller's BDT balance, generates WireGuard keys, allocates an IP from the subnet, and adds the peer to the WireGuard interface.

- **URL:** `/api/v1/reseller/peers`
- **Method:** `POST`
- **Auth:** Reseller

**Request Body:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `whatsapp_number` | string | ✅ | Client's WhatsApp number |
| `package_id` | int | ✅ | ID of the package to assign |

**Request:**

```bash
curl -X POST -H "Authorization: Bearer <reseller_token>" \
  -H "Content-Type: application/json" \
  -d '{"whatsapp_number": "+8801712345678", "package_id": 5}' \
  http://localhost:3000/api/v1/reseller/peers
```

**Response `201 Created`:**

```json
{
  "success": true,
  "data": {
    "id": 7,
    "user_id": 3,
    "whatsapp_number": "+8801712345678",
    "public_key": "lggeiJiw6JVjeLTT/aLX+tD13B/ZMYUbFbbh9z73yXo=",
    "assigned_ip": "10.66.66.2/32",
    "expiry_date": "2026-05-16T18:44:29.912087342Z",
    "status": "active",
    "package_id": 5,
    "created_at": "2026-04-16T18:44:29.897626Z"
  }
}
```

**Response `400 Bad Request` (insufficient balance):**

```json
{
  "success": false,
  "error": {
    "code": 400,
    "message": "insufficient balance: have ৳0.00, need ৳200.00"
  }
}
```

**Response `400 Bad Request` (validation):**

```json
{
  "success": false,
  "error": { "code": 400, "message": "whatsapp_number and package_id are required" }
}
```

---

### 21. Download Peer Config

Download the AmneziaWG client configuration for a peer. Supports text file download or JSON format.

- **URL:** `/api/v1/reseller/peers/:id/config`
- **Method:** `GET`
- **Auth:** Reseller (only own peers)

**Path Parameters:**

| Param | Type | Description |
|-------|------|-------------|
| `id` | int | Peer ID |

**Query Parameters:**

| Param | Type | Default | Description |
|-------|------|---------|-------------|
| `format` | string | `text` | `text` returns `.conf` file download, `json` returns config string in JSON |

**Request (JSON format):**

```bash
curl -H "Authorization: Bearer <reseller_token>" \
  "http://localhost:3000/api/v1/reseller/peers/7/config?format=json"
```

**Response `200 OK` (format=json):**

```json
{
  "success": true,
  "data": {
    "peer_id": 7,
    "config": "[Interface]\nPrivateKey = iJE+N/wr9D7MiNEkMzI5cOt5ScnyrAAZOm6DfchCkV8=\nAddress = 10.66.66.2/32\nDNS = 1.1.1.1, 8.8.8.8\nJc = 4\nJmin = 40\nJmax = 70\nS1 = 0\nS2 = 0\nH1 = 1\nH2 = 2\nH3 = 3\nH4 = 4\n\n[Peer]\nPublicKey = TmWRFn8OGDMV/lppAVLfzufhQY3+G06Y5JH2G2e1XxQ=\nPresharedKey = VFKxD7LqMD+gwqF3Cw3wiV40zIp+KlbXkmWSYNyf83g=\nEndpoint = vpn.example.com:51820\nAllowedIPs = 0.0.0.0/0, ::/0\nPersistentKeepalive = 25\n"
  }
}
```

**Request (text file download):**

```bash
curl -H "Authorization: Bearer <reseller_token>" \
  "http://localhost:3000/api/v1/reseller/peers/7/config" \
  -o awg-client.conf
```

**Response `200 OK` (format=text):**

Returns a plain text `.conf` file with header `Content-Disposition: attachment; filename=awg-+8801712345678.conf`:

```ini
[Interface]
PrivateKey = iJE+N/wr9D7MiNEkMzI5cOt5ScnyrAAZOm6DfchCkV8=
Address = 10.66.66.2/32
DNS = 1.1.1.1, 8.8.8.8
Jc = 4
Jmin = 40
Jmax = 70
S1 = 0
S2 = 0
H1 = 1
H2 = 2
H3 = 3
H4 = 4

[Peer]
PublicKey = TmWRFn8OGDMV/lppAVLfzufhQY3+G06Y5JH2G2e1XxQ=
PresharedKey = VFKxD7LqMD+gwqF3Cw3wiV40zIp+KlbXkmWSYNyf83g=
Endpoint = vpn.example.com:51820
AllowedIPs = 0.0.0.0/0, ::/0
PersistentKeepalive = 25
```

**Response `404 Not Found`:**

```json
{
  "success": false,
  "error": { "code": 404, "message": "peer not found" }
}
```

---

### 22. Get Peer QR Code

Returns a QR code PNG image of the AmneziaWG client configuration. Supports raw PNG or base64-encoded JSON.

- **URL:** `/api/v1/reseller/peers/:id/qr`
- **Method:** `GET`
- **Auth:** Reseller (only own peers)

**Path Parameters:**

| Param | Type | Description |
|-------|------|-------------|
| `id` | int | Peer ID |

**Query Parameters:**

| Param | Type | Default | Description |
|-------|------|---------|-------------|
| `format` | string | `png` | `png` returns raw image, `base64` returns base64 string in JSON |

**Request (base64 format):**

```bash
curl -H "Authorization: Bearer <reseller_token>" \
  "http://localhost:3000/api/v1/reseller/peers/7/qr?format=base64"
```

**Response `200 OK` (format=base64):**

```json
{
  "success": true,
  "data": {
    "peer_id": 7,
    "qr_png": "iVBORw0KGgoAAAANSUhEUgAAAgAAAAIA..."
  }
}
```

**Request (raw PNG image):**

```bash
curl -H "Authorization: Bearer <reseller_token>" \
  "http://localhost:3000/api/v1/reseller/peers/7/qr" \
  -o qr-code.png
```

**Response `200 OK` (format=png):**

Returns a `512×512` PNG image with `Content-Type: image/png`.

---

### 23. List Own Transactions

Returns the reseller's own transaction history (credits and debits).

- **URL:** `/api/v1/reseller/transactions`
- **Method:** `GET`
- **Auth:** Reseller

**Query Parameters:**

| Param | Type | Default | Description |
|-------|------|---------|-------------|
| `page` | int | `1` | Page number |
| `per_page` | int | `25` | Items per page (max 100) |
| `type` | string | — | Filter by `credit` or `debit` |

**Request:**

```bash
curl -H "Authorization: Bearer <reseller_token>" \
  "http://localhost:3000/api/v1/reseller/transactions?page=1"
```

**Response `200 OK`:**

```json
{
  "success": true,
  "data": [
    {
      "id": 10,
      "reseller_id": 3,
      "amount": "200",
      "type": "debit",
      "note": "Peer creation: 30 Days for +8801712345678",
      "created_at": "2026-04-16T18:44:29.897626Z"
    },
    {
      "id": 6,
      "reseller_id": 3,
      "amount": "5000",
      "type": "credit",
      "note": "Initial top-up",
      "created_at": "2026-04-16T18:28:48.763379Z"
    }
  ],
  "meta": {
    "page": 1,
    "per_page": 25,
    "total": 2
  }
}
```

---

## Common Error Responses

These errors can be returned by any authenticated endpoint:

### 401 Unauthorized

```json
{
  "success": false,
  "error": { "code": 401, "message": "missing authorization header" }
}
```

```json
{
  "success": false,
  "error": { "code": 401, "message": "invalid or expired token" }
}
```

```json
{
  "success": false,
  "error": { "code": 401, "message": "invalid authorization format, use: Bearer <token>" }
}
```

### 403 Forbidden

Returned when an authenticated user tries to access an endpoint outside their role.

```json
{
  "success": false,
  "error": { "code": 403, "message": "insufficient permissions" }
}
```

### 500 Internal Server Error

```json
{
  "success": false,
  "error": { "code": 500, "message": "query failed" }
}
```

---

## Authentication Flow

```
1. POST /api/v1/auth/login  →  get Bearer token
2. Use token in all subsequent requests:
   Authorization: Bearer eyJhbGciOiJIUzI1...
3. Token expires after 24 hours — re-login to get a new one.
```

## Environment Notes

| Variable | Description |
|----------|-------------|
| `WG_DRY_RUN=true` | Skips real WireGuard `wg set` commands (for dev/testing). Remove or set to `false` in production. |
| `JWT_SECRET` | Secret key for signing JWT tokens. **Change from default in production.** |
