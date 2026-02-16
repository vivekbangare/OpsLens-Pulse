package db

import (
	"database/sql"
	"time"

	_ "github.com/ClickHouse/clickhouse-go/v2"
)

func NewClickHouse(dsn string) (*sql.DB, error) {

	db, err := sql.Open("clickhouse", dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(15 * time.Minute)

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
