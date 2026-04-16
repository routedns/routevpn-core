package services

import (
	"github.com/imzami/routevpn-core/internal/config"
	"github.com/imzami/routevpn-core/internal/database"
	"github.com/jmoiron/sqlx"
)

type Container struct {
	DB    *sqlx.DB
	Cache *database.Valkey
	Cfg   *config.Config
}
