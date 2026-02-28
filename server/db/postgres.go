package db

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"

	"opslense-pulse/server/config"
)

func NewPostgres(cfg config.PostgresSection) (*sql.DB, error) {

	if cfg.Port == 0 {
		cfg.Port = 5432
	}

	if cfg.SSLMode == "" {
		cfg.SSLMode = "disable"
	}

	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host,
		cfg.Port,
		cfg.User,
		cfg.Password,
		cfg.Database,
		cfg.SSLMode,
	)

	var dbConn *sql.DB
	var err error

	retries := cfg.ConnectRetries
	if retries <= 0 {
		retries = 3
	}

	retryDelay := time.Duration(cfg.RetryDelaySec) * time.Second
	if retryDelay <= 0 {
		retryDelay = 2 * time.Second
	}

	for i := 1; i <= retries; i++ {

		dbConn, err = sql.Open("postgres", dsn)
		if err != nil {
			return nil, err
		}

		err = dbConn.Ping()
		if err == nil {
			break
		}

		if i == retries {
			return nil, err
		}

		time.Sleep(retryDelay)
	}

	// ---------- Pool Settings ----------
	if cfg.MaxOpenConns > 0 {
		dbConn.SetMaxOpenConns(cfg.MaxOpenConns)
	}

	if cfg.MaxIdleConns > 0 {
		dbConn.SetMaxIdleConns(cfg.MaxIdleConns)
	}

	if cfg.ConnMaxLifetimeMin > 0 {
		dbConn.SetConnMaxLifetime(
			time.Duration(cfg.ConnMaxLifetimeMin) * time.Minute,
		)
	}

	return dbConn, nil
}
