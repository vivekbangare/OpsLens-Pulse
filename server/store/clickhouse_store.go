package store

import (
	"database/sql"
	"sync/atomic"

	_ "github.com/ClickHouse/clickhouse-go/v2"
)

type ClickHouseStore struct {
	db     *sql.DB
	health atomic.Bool
}

func NewClickHouseStore(db *sql.DB) *ClickHouseStore {
	return &ClickHouseStore{
		db: db,
	}
}

func (c *ClickHouseStore) DB() *sql.DB {
	return c.db
}

////////////////////////////////////////////////////////////
// Close
////////////////////////////////////////////////////////////

func (c *ClickHouseStore) Close() error {
	return c.db.Close()
}

////////////////////////////////////////////////////////////
// Interface Check
////////////////////////////////////////////////////////////

var _ Store = (*ClickHouseStore)(nil)
