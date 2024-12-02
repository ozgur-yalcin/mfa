package database

import (
	"errors"
	"sync"

	"github.com/ozgur-yalcin/mfa/src/backend"
	"github.com/ozgur-yalcin/mfa/src/config"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Database struct {
	db      *gorm.DB
	dbLock  sync.Mutex
	backend backend.Backend
}

func (db *Database) Open() error {
	var conn *gorm.DB
	var err error
	switch db.backend.Engine() {
	case "sqlite":
		params := db.backend.Params()
		conn, err = gorm.Open(sqlite.Open(params), &gorm.Config{})
	case "postgresql":
		params := db.backend.Params()
		conn, err = gorm.Open(postgres.Open(params), &gorm.Config{})
	default:
		return errors.New("not supported database engine")
	}
	if err != nil {
		return err
	}
	db.db = conn
	return nil
}

func (db *Database) Close() error {
	client, err := db.db.DB()
	if err != nil {
		return err
	}
	return client.Close()
}

func LoadDatabase() (*Database, error) {
	return &Database{backend: config.DefaultConfig().DatabaseBackend}, nil
}

func (db *Database) Engine() string {
	return db.backend.Engine()
}

func (db *Database) AutoMigrate(dst ...interface{}) error {
	return db.db.AutoMigrate(dst...)
}
