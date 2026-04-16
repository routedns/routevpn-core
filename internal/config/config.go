package config

import (
	"fmt"
	"os"
)

type Config struct {
	Port              string
	DatabaseURL       string
	ValkeyURL         string
	JWTSecret         string
	DefaultAdminEmail string
	DefaultAdminPass  string
	WGInterface       string
	WGEndpoint        string
	WGListenPort      string
	WGDNS             string
	WGSubnet          string
	// AmneziaWG parameters
	AWGJc            string
	AWGJmin          string
	AWGJmax          string
	AWGS1            string
	AWGS2            string
	AWGH1            string
	AWGH2            string
	AWGH3            string
	AWGH4            string
	ServerPrivateKey string
	ServerPublicKey  string
	CORSOrigins      string
}

func Load() *Config {
	return &Config{
		Port:              getEnv("PORT", "3000"),
		DatabaseURL:       getEnv("DATABASE_URL", "postgres://routevpn:routevpn@localhost:5432/routevpn?sslmode=disable"),
		ValkeyURL:         getEnv("VALKEY_URL", "localhost:6379"),
		JWTSecret:         getEnv("JWT_SECRET", "change-me-in-production"),
		DefaultAdminEmail: getEnv("ADMIN_EMAIL", "admin@routevpn.local"),
		DefaultAdminPass:  getEnv("ADMIN_PASSWORD", "admin123"),
		WGInterface:       getEnv("WG_INTERFACE", "awg0"),
		WGEndpoint:        getEnv("WG_ENDPOINT", "vpn.example.com:51820"),
		WGListenPort:      getEnv("WG_LISTEN_PORT", "51820"),
		WGDNS:             getEnv("WG_DNS", "1.1.1.1, 8.8.8.8"),
		WGSubnet:          getEnv("WG_SUBNET", "10.66.66.0/24"),
		AWGJc:             getEnv("AWG_JC", "4"),
		AWGJmin:           getEnv("AWG_JMIN", "40"),
		AWGJmax:           getEnv("AWG_JMAX", "70"),
		AWGS1:             getEnv("AWG_S1", "0"),
		AWGS2:             getEnv("AWG_S2", "0"),
		AWGH1:             getEnv("AWG_H1", "1"),
		AWGH2:             getEnv("AWG_H2", "2"),
		AWGH3:             getEnv("AWG_H3", "3"),
		AWGH4:             getEnv("AWG_H4", "4"),
		ServerPrivateKey:  getEnv("SERVER_PRIVATE_KEY", ""),
		ServerPublicKey:   getEnv("SERVER_PUBLIC_KEY", ""),
		CORSOrigins:       getEnv("CORS_ORIGINS", "*"),
	}
}

func (c *Config) Validate() error {
	if c.JWTSecret == "change-me-in-production" {
		return fmt.Errorf("JWT_SECRET must be set to a secure value")
	}
	if len(c.JWTSecret) < 32 {
		return fmt.Errorf("JWT_SECRET must be at least 32 characters")
	}
	if c.ServerPrivateKey == "" || c.ServerPublicKey == "" {
		return fmt.Errorf("SERVER_PRIVATE_KEY and SERVER_PUBLIC_KEY must be set")
	}
	if c.DefaultAdminPass == "admin123" {
		return fmt.Errorf("ADMIN_PASSWORD must be changed from default")
	}
	return nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
