package database

import (
	"github.com/jmoiron/sqlx"
)

const migrationSQL = `
CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    role VARCHAR(20) NOT NULL DEFAULT 'reseller' CHECK (role IN ('admin', 'reseller')),
    balance NUMERIC(12,2) NOT NULL DEFAULT 0.00,
    created_by BIGINT REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS packages (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    price_bdt NUMERIC(12,2) NOT NULL,
    duration_days INTEGER NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS vpn_peers (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id),
    whatsapp_number VARCHAR(20) NOT NULL,
    public_key VARCHAR(64) NOT NULL UNIQUE,
    private_key VARCHAR(64) NOT NULL,
    preshared_key VARCHAR(64) NOT NULL,
    assigned_ip VARCHAR(18) NOT NULL UNIQUE,
    expiry_date TIMESTAMPTZ NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'expired')),
    package_id BIGINT NOT NULL REFERENCES packages(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_vpn_peers_status ON vpn_peers(status);
CREATE INDEX IF NOT EXISTS idx_vpn_peers_expiry ON vpn_peers(expiry_date);
CREATE INDEX IF NOT EXISTS idx_vpn_peers_whatsapp ON vpn_peers(whatsapp_number);
CREATE INDEX IF NOT EXISTS idx_vpn_peers_user_id ON vpn_peers(user_id);

CREATE TABLE IF NOT EXISTS transactions (
    id BIGSERIAL PRIMARY KEY,
    reseller_id BIGINT NOT NULL REFERENCES users(id),
    amount NUMERIC(12,2) NOT NULL,
    type VARCHAR(10) NOT NULL CHECK (type IN ('credit', 'debit')),
    note TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_transactions_reseller ON transactions(reseller_id);
`

func RunMigrations(db *sqlx.DB) error {
	_, err := db.Exec(migrationSQL)
	return err
}
