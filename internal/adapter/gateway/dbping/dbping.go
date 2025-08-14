package dbping

import (
	"context"
	"database/sql"
)

type DBPing struct {
	DB *sql.DB
}

func (DBPing) Name() string { return "db" }

func (d DBPing) Ping(ctx context.Context) error {
	return d.DB.PingContext(ctx)
}
