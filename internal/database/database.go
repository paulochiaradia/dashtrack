package database

import (
	"database/sql"
	"log"

	_ "github.com/jackc/pgx/v5/stdlib"
	migrate "github.com/rubenv/sql-migrate"
)

// NewDatabase creates and returns a new database connection pool.
func NewDatabase(dsn string) *sql.DB {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatalf("could not connect to the database: %v", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("could not ping the database: %v", err)
	}

	// Migrations are handled by docker-entrypoint.sh using golang-migrate CLI
	// runMigrations(db)

	return db
}

func runMigrations(db *sql.DB) {
	migrations := &migrate.FileMigrationSource{
		Dir: "migrations",
	}

	n, err := migrate.Exec(db, "postgres", migrations, migrate.Up)
	if err != nil {
		log.Fatalf("could not run migrations: %v", err)
	}

	if n > 0 {
		log.Printf("Applied %d migrations!\n", n)
	} else {
		log.Println("No new migrations to apply.")
	}
}
