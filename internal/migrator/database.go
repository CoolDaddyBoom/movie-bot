package migrator

import (
	"database/sql"

	_ "github.com/lib/pq"
)

func Connect(dsn string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}

func EnsureSchemaMigrationsTable(db *sql.DB) error {
	query := `
    CREATE TABLE IF NOT EXISTS schema_migrations (
        version BIGINT PRIMARY KEY,
        dirty BOOLEAN NOT NULL DEFAULT FALSE
    )`
	_, err := db.Exec(query)
	return err
}
