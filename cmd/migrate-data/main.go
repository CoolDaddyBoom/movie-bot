package main

/*import (
	"context"
	"log"

	"whattowatchbot/storage"
	"whattowatchbot/storage/postgres"
	"whattowatchbot/storage/sqlite"
)

func main() {
	log.Println("starting data migration...")

	ctx := context.Background()

	log.Println("connecting to sqlite database...")

	sqliteDB, err := sqlite.New("movies.db")
	if err != nil {
		log.Fatalf("failed to connect to sqlite database: %v", err)
	}
	defer sqliteDB.Close()
	log.Println("✅ SQLite connected")

	log.Println("connecting to postgres database...")
	postgresDB, err := postgres.New("host=localhost port=5432 user=moviebot password=211973 dbname=moviebot sslmode=disable")
	if err != nil {
		log.Fatalf("failed to connect to postgres database: %v", err)
	}
	defer postgresDB.Close()

	if err := postgresDB.Init(ctx); err != nil {
		log.Fatalf("❌ Failed to initialize PostgreSQL: %v", err)
	}
	log.Println("✅ PostgreSQL connected and initialized")

	log.Println("📥 Fetching movies from SQLite...")
	movies, err := sqliteDB.ListAll(ctx)
	if err != nil {
		log.Fatalf("❌ Failed to fetch movies: %v", err)
	}
	log.Printf("📊 Found %d movies in SQLite", len(movies))

	// Мігрувати фільми
	log.Println("🔄 Migrating movies to PostgreSQL...")
	migrated, skipped := migrateMovies(ctx, postgresDB, movies)

	// Статистика
	log.Println("📈 Migration completed:")
	log.Printf("   ✅ Migrated: %d movies", migrated)
	log.Printf("   ⏭️  Skipped: %d movies (duplicates)", skipped)
	log.Printf("   📊 Total: %d movies", len(movies))
}

func migrateMovies(ctx context.Context, db storage.Storage, movies []*storage.Movie) (migrated, skipped int) {
	for i, movie := range movies {
		log.Printf("[%d/%d] Processing: %s (chat: %d)",
			i+1, len(movies), movie.Title, movie.ChatID)
		// Перевірити чи існує
		exists, err := db.IsExists(ctx, movie)
		if err != nil {
			log.Printf("   ⚠️  Error checking existence: %v", err)
			continue // Пропускаємо при помилці
		}

		if exists {
			log.Printf("   ⏭️  Already exists, skipping")
			skipped++
			continue
		}

		// Зберегти в PostgreSQL
		if err := db.Save(ctx, movie); err != nil {
			log.Printf("   ❌ Failed to save: %v", err)
			continue
		}

		log.Printf("   ✅ Migrated successfully")
		migrated++
	}

	return migrated, skipped
}
*/
