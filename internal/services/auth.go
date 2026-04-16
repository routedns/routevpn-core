package services

import (
	"fmt"

	"github.com/imzami/routevpn-core/internal/config"
	"github.com/imzami/routevpn-core/internal/models"
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
)

func EnsureDefaultAdmin(db *sqlx.DB, cfg *config.Config) error {
	var count int
	if err := db.Get(&count, "SELECT COUNT(*) FROM users WHERE role='admin'"); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(cfg.DefaultAdminPass), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	_, err = db.Exec("INSERT INTO users (email, password, role) VALUES ($1, $2, 'admin')",
		cfg.DefaultAdminEmail, string(hash))
	return err
}

func AuthenticateUser(db *sqlx.DB, email, password string) (*models.User, error) {
	var u models.User
	if err := db.Get(&u, "SELECT * FROM users WHERE email=$1", email); err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password)); err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}
	return &u, nil
}
