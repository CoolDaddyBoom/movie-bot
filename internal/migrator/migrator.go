package migrator

import (
	"database/sql"
	"fmt"
	"log/slog"
)

type Migrator struct {
	db     *sql.DB
	logger *slog.Logger
}

func New(dsn string, logger *slog.Logger) (*Migrator, error) {
	db, err := Connect(dsn)
	if err != nil {
		return nil, err
	}

	if err := EnsureSchemaMigrationsTable(db); err != nil {
		return nil, err
	}

	return &Migrator{
		db:     db,
		logger: logger,
	}, nil
}

func (m *Migrator) Close() error {
	return m.db.Close()
}

func (m *Migrator) Up() error {
	m.logger.Info("Starting migrations")

	migrations, err := readMigrations()
	if err != nil {
		m.logger.Error("Failed to read migrations", "error", err)
		return fmt.Errorf("failed to read migrations: %w", err)
	}

	applied, err := m.getAppliedVersions()
	if err != nil {
		m.logger.Error("Failed to get applied versions", "error", err)
		return fmt.Errorf("failed to get applied versions: %w", err)
	}

	for _, migration := range migrations {
		if contains(applied, migration.Version) {
			m.logger.Debug("Migration already applied", "version", migration.Version)
			continue
		}

		m.logger.Info("Applying migration", "version", migration.Version, "name", migration.Name)
		if err := m.applyMigration(migration); err != nil {
			return fmt.Errorf("failed to apply migration %d: %w", migration.Version, err)
		}
		m.logger.Info("Migration applied successfully", "version", migration.Version)
	}

	m.logger.Info("All migrations completed")
	return nil
}

func (m *Migrator) Down() error {
	m.logger.Info("Starting rollback")

	migrations, err := readMigrations()
	if err != nil {
		m.logger.Error("Failed to read migrations", "error", err)
		return fmt.Errorf("failed to read migrations: %w", err)
	}

	applied, err := m.getAppliedVersions()
	if err != nil {
		m.logger.Error("Failed to get applied versions", "error", err)
		return fmt.Errorf("failed to get applied versions: %w", err)
	}

	if len(applied) == 0 {
		m.logger.Info("No migrations to rollback")
		return nil
	}

	lastVersion := applied[len(applied)-1]

	var target Migration
	found := false
	for _, mig := range migrations {
		if mig.Version == lastVersion {
			target = mig
			found = true
			break
		}
	}

	if !found {
		m.logger.Error("Migration file not found", "version", lastVersion)
		return fmt.Errorf("migration file for version %d not found", lastVersion)
	}

	m.logger.Info("Rolling back migration", "version", target.Version, "name", target.Name)

	tx, err := m.db.Begin()
	if err != nil {
		m.logger.Error("Failed to begin transaction", "error", err)
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	_, err = tx.Exec(target.DownSQL)
	if err != nil {
		m.logger.Error("Failed to execute down SQL", "error", err)
		return fmt.Errorf("failed to run down SQL: %w", err)
	}

	_, err = tx.Exec("DELETE FROM schema_migrations WHERE version = $1", target.Version)
	if err != nil {
		m.logger.Error("Failed to delete migration record", "error", err)
		return err
	}

	if err = tx.Commit(); err != nil {
		m.logger.Error("Failed to commit transaction", "error", err)
		return err
	}

	m.logger.Info("Migration rolled back successfully", "version", target.Version)
	return nil
}

func (m *Migrator) Status() error {
	m.logger.Info("Checking migration status")

	migrations, err := readMigrations()
	if err != nil {
		m.logger.Error("Failed to read migrations", "error", err)
		return fmt.Errorf("failed to read migrations: %w", err)
	}

	applied, err := m.getAppliedVersions()
	if err != nil {
		m.logger.Error("Failed to get applied versions", "error", err)
		return fmt.Errorf("failed to get applied versions: %w", err)
	}

	dirty, err := m.getDirtyVersions()
	if err != nil {
		m.logger.Error("Failed to get dirty versions", "error", err)
		return fmt.Errorf("failed to get dirty versions: %w", err)
	}

	// Мапи для швидкого доступу
	appliedMap := make(map[int]bool)
	for _, v := range applied {
		appliedMap[v] = true
	}

	dirtyMap := make(map[int]bool)
	for _, v := range dirty {
		dirtyMap[v] = true
	}

	// Логування статусу кожної міграції
	for _, mig := range migrations {
		if dirtyMap[mig.Version] {
			m.logger.Warn("Migration dirty",
				"version", mig.Version,
				"name", mig.Name,
				"status", "dirty")
		} else if appliedMap[mig.Version] {
			m.logger.Info("Migration applied",
				"version", mig.Version,
				"name", mig.Name,
				"status", "applied")
		} else {
			m.logger.Info("Migration pending",
				"version", mig.Version,
				"name", mig.Name,
				"status", "pending")
		}

		delete(appliedMap, mig.Version)
		delete(dirtyMap, mig.Version)
	}

	// Застарілі міграції (файли видалені)
	for lost := range appliedMap {
		m.logger.Warn("Migration file missing",
			"version", lost,
			"status", "applied but file missing")
	}

	for lost := range dirtyMap {
		m.logger.Error("Dirty migration file missing",
			"version", lost,
			"status", "dirty but file missing")
	}

	// Підсумок
	totalMigrations := len(migrations)
	appliedCount := len(applied)
	pendingCount := totalMigrations - appliedCount
	dirtyCount := len(dirty)

	m.logger.Info("Migration summary",
		"total", totalMigrations,
		"applied", appliedCount,
		"pending", pendingCount,
		"dirty", dirtyCount)

	if dirtyCount > 0 {
		m.logger.Warn("Some migrations are dirty",
			"dirty_count", dirtyCount,
			"action", "Fix manually or use: migrator force <version>")
	}

	return nil
}

func (m *Migrator) getAppliedVersions() ([]int, error) {
	rows, err := m.db.Query("SELECT version FROM schema_migrations WHERE dirty = false ORDER BY version")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var versions []int
	for rows.Next() {
		var v int
		if err := rows.Scan(&v); err != nil {
			return nil, err
		}
		versions = append(versions, v)
	}
	return versions, nil
}

func (m *Migrator) getDirtyVersions() ([]int, error) {
	rows, err := m.db.Query("SELECT version FROM schema_migrations WHERE dirty = true ORDER BY version")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var versions []int
	for rows.Next() {
		var v int
		if err := rows.Scan(&v); err != nil {
			return nil, err
		}
		versions = append(versions, v)
	}
	return versions, nil
}

func (m *Migrator) applyMigration(migration Migration) error {
	tx, err := m.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	_, err = tx.Exec("INSERT INTO schema_migrations (version, dirty) VALUES ($1, true)", migration.Version)
	if err != nil {
		return err
	}

	_, err = tx.Exec(migration.UpSQL)
	if err != nil {
		return err
	}

	_, err = tx.Exec("UPDATE schema_migrations SET dirty = false WHERE version = $1", migration.Version)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func contains(slice []int, val int) bool {
	for _, v := range slice {
		if v == val {
			return true
		}
	}
	return false
}
