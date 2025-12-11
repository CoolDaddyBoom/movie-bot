package migrator

import (
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type Migration struct {
	Version int
	Name    string
	UpSQL   string
	DownSQL string
}

func readMigrations() ([]Migration, error) {
	entries, err := os.ReadDir("migrations")
	if err != nil {
		return nil, err
	}

	migrationsMap := make(map[int]*Migration)

	for _, entry := range entries {
		name := entry.Name()

		// Парсити: 001_create_movies.up.sql
		parts := strings.Split(name, "_")
		if len(parts) < 2 {
			continue
		}

		version, err := strconv.Atoi(parts[0])
		if err != nil {
			continue
		}

		isUp := strings.HasSuffix(name, ".up.sql")
		isDown := strings.HasSuffix(name, ".down.sql")

		if !isUp && !isDown {
			continue
		}

		sql, err := os.ReadFile(filepath.Join("migrations", name))
		if err != nil {
			return nil, err
		}

		if migrationsMap[version] == nil {
			migrationsMap[version] = &Migration{
				Version: version,
				Name:    strings.TrimSuffix(strings.TrimSuffix(name, ".up.sql"), ".down.sql"),
			}
		}

		if isUp {
			migrationsMap[version].UpSQL = string(sql)
		} else {
			migrationsMap[version].DownSQL = string(sql)
		}
	}
	// Перетворити map в slice
	var migrations []Migration
	for _, m := range migrationsMap {
		migrations = append(migrations, *m)
	}

	// Сортувати за версією
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	return migrations, nil
}
