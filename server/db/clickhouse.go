package db

import (
	"database/sql"
	"fmt"
	"time"

	"opslense-pulse/server/config"

	_ "github.com/ClickHouse/clickhouse-go/v2"
)

func NewClickHouse(cfg config.ClickHouseSection) (*sql.DB, error) {

	if cfg.Port == 0 {
		cfg.Port = 9000
	}

	dsn := fmt.Sprintf(
		"tcp://%s:%d/%s?username=%s&password=%s",
		cfg.Host,
		cfg.Port,
		cfg.Database,
		cfg.User,
		cfg.Password,
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

		dbConn, err = sql.Open("clickhouse", dsn)
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
