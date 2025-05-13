package wallet

import (
	"log/slog"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

const ttl = 24 * time.Hour

type Repository struct {
	db     *sqlx.DB
	cache  *redis.Client
	logger *slog.Logger
}

func New(db *sqlx.DB, cache *redis.Client, logger *slog.Logger) *Repository {
	return &Repository{
		db:     db,
		cache:  cache,
		logger: logger,
	}
}
