package db

import (
	"database/sql"
	"os"
	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func MySqlHandler() (*sql.DB, error) {
	dataSource := os.Getenv("MYSQL_USERNAME") + ":" +
		os.Getenv("MYSQL_PASSWORD") + "@tcp(" +
		os.Getenv("MYSQL_HOST") + ":" +
		os.Getenv("MYSQL_PORT") + ")/" +
		os.Getenv("MYSQL_DATABASE") + "?multiStatements=true"

	db, err := sql.Open("mysql", dataSource)
	if err != nil {
		return nil, err
	}

	maxIdleConns, err := strconv.Atoi(os.Getenv("MYSQL_MAX_IDLE_CONNS"))
	if err != nil {
		return nil, err
	}

	maxOpenConns, err := strconv.Atoi(os.Getenv("MYSQL_MAX_OPEN_CONNS"))
	if err != nil {
		return nil, err
	}

	db.SetMaxIdleConns(maxIdleConns)
	db.SetMaxOpenConns(maxOpenConns)
	db.SetConnMaxLifetime(20 * time.Minute)

	return db, nil
}
