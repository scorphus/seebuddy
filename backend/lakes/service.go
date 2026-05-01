// Package lakes is the only Encore service in the app. It owns the readings
// database, runs a single cron job that polls every registered adapter, and
// exposes the public read API.
package lakes

import (
	"encore.dev/storage/sqldb"
	"github.com/jackc/pgx/v5/pgxpool"
)

var db = sqldb.Driver[*pgxpool.Pool](sqldb.NewDatabase("lakes", sqldb.DatabaseConfig{
	Migrations: "./migrations",
}))

var queries = New(db)
