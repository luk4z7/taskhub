package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func MigrationRun(ctx context.Context, db *sql.DB) error {
	try := func() error {
		driver, err := mysql.WithInstance(db, &mysql.Config{})
		if err != nil {
			return err
		}

		m, err := migrate.NewWithDatabaseInstance(
			"file:taskmanager/db/migrations",
			"mysql", driver)
		if err != nil {
			return err
		}

		if err := m.Up(); err != migrate.ErrNoChange {
			return err
		}

		return nil
	}

	interval, err := time.ParseDuration("2s")
	if err != nil {
		return err
	}

	return ticker(ctx, interval, try)
}

func ticker(ctx context.Context, d time.Duration, f func() error) error {
	ticker := time.NewTicker(d)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			err := f()
			if err == nil {
				return nil
			}

			fmt.Println("try connection to database", err)
		}
	}
}
