package models

import (
	"time"

	"github.com/shopspring/decimal"
)

type User struct {
	ID        int64           `json:"id" db:"id"`
	Email     string          `json:"email" db:"email"`
	Password  string          `json:"-" db:"password"`
	Role      string          `json:"role" db:"role"`
	Balance   decimal.Decimal `json:"balance" db:"balance"`
	CreatedBy *int64          `json:"created_by" db:"created_by"`
	CreatedAt time.Time       `json:"created_at" db:"created_at"`
}

type VPNPeer struct {
	ID              int64     `json:"id" db:"id"`
	UserID          int64     `json:"user_id" db:"user_id"`
	WhatsappNumber  string    `json:"whatsapp_number" db:"whatsapp_number"`
	PublicKey       string    `json:"public_key" db:"public_key"`
	PrivateKey      string    `json:"-" db:"private_key"`
	PresharedKey    string    `json:"-" db:"preshared_key"`
	AssignedIP      string    `json:"assigned_ip" db:"assigned_ip"`
	ExpiryDate      time.Time `json:"expiry_date" db:"expiry_date"`
	Status          string    `json:"status" db:"status"`
	PackageID       int64     `json:"package_id" db:"package_id"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	// Joined fields
	ResellerEmail   string          `json:"reseller_email,omitempty" db:"reseller_email"`
	PackageName     string          `json:"package_name,omitempty" db:"package_name"`
}

type Package struct {
	ID           int64           `json:"id" db:"id"`
	Name         string          `json:"name" db:"name"`
	PriceBDT     decimal.Decimal `json:"price_bdt" db:"price_bdt"`
	DurationDays int             `json:"duration_days" db:"duration_days"`
	CreatedAt    time.Time       `json:"created_at" db:"created_at"`
}

type Transaction struct {
	ID         int64           `json:"id" db:"id"`
	ResellerID int64           `json:"reseller_id" db:"reseller_id"`
	Amount     decimal.Decimal `json:"amount" db:"amount"`
	Type       string          `json:"type" db:"type"`
	Note       string          `json:"note" db:"note"`
	CreatedAt  time.Time       `json:"created_at" db:"created_at"`
	// Joined
	ResellerEmail string `json:"reseller_email,omitempty" db:"reseller_email"`
}
