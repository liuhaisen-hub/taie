package data

import (
	"github.com/glebarez/sqlite"
	"github.com/google/wire"
	"gorm.io/gorm"
)

var ProviderSet = wire.NewSet(NewData, NewModelRepo)

// DefaultDSN is used when no DSN is provided.
const DefaultDSN = "taie.db"

// Data wraps the GORM connection to the SQLite database. The glebarez/sqlite
// dialect is a pure-Go driver backed by modernc.org/sqlite (no cgo).
type Data struct {
	db *gorm.DB
}

// NewData opens (and creates if needed) the SQLite database at dsn and
// returns the connection wrapper. An empty dsn falls back to DefaultDSN.
func NewData(dsn string) (*Data, error) {
	if dsn == "" {
		dsn = DefaultDSN
	}

	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	return &Data{db: db}, nil
}

// DB returns the underlying GORM handle.
func (d *Data) DB() *gorm.DB {
	return d.db
}
