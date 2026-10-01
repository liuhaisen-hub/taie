package data

import (
	"github.com/glebarez/sqlite"
	"github.com/google/wire"
	"gorm.io/gorm"
)

var ProviderSet = wire.NewSet(NewData,
	NewModelRepo,
	NewAgentRepo,
	NewSessionRepo,
	NewToolsRepo,
	NewSystemRepo,
	NewCheckPointRepo,
	NewChatSessionRepo,
	NewTokenUseRepo,
	NewUseRepo,
	NewPermissionRepo,
)

// DefaultDSN is used when no DSN is provided.
const DefaultDSN = "taie.db"

// Data wraps the GORM connection to the SQLite database. The glebarez/sqlite
// dialect is a pure-Go driver backed by modernc.org/sqlite (no cgo).
type Data struct {
	db *gorm.DB
}

// NewData opens (and creates if needed) the SQLite database at dsn and
// returns the connection wrapper. An empty dsn falls back to DefaultDSN.
func NewData(dsn string) (*Data, func(), error) {
	if dsn == "" {
		dsn = DefaultDSN
	}

	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, nil, err
	}
	sql, err := db.DB()
	if err != nil {
		return nil, nil, err
	}
	sql.SetMaxOpenConns(1)
	sql.SetMaxIdleConns(1)
	cleanup := func() {
		sql.Close()
	}
	return &Data{db: db}, cleanup, nil
}
