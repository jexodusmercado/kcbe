package store

import (
	"database/sql"
	"fmt"
	"io/fs"
	"kabancount/internal/config"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

func Open() (*sql.DB, error) {
	cfg := config.Get()
	dbString := cfg.GetDatabaseDSN()
	db, err := sql.Open("pgx", dbString)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %w", err)
	}

	fmt.Println("Database connection established")
	return db, nil
}

func MigrateFS(db *sql.DB, migrationFS fs.FS, dir string) error {
	goose.SetBaseFS(migrationFS)
	defer func() {
		goose.SetBaseFS(nil)
	}()
	return Migrate(db, dir)
}

func Migrate(db *sql.DB, dir string) error {
	err := goose.SetDialect("postgres")
	if err != nil {
		return fmt.Errorf("migrate: %w", err)
	}

	err = goose.Up(db, dir)
	if err != nil {
		return fmt.Errorf("goose up: %w", err)
	}

	return nil
}
