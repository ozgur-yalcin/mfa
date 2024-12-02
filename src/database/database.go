package database

import (
	"errors"

	"github.com/ozgur-yalcin/mfa/src/backend"
	"github.com/ozgur-yalcin/mfa/src/config"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Database struct {
	db      *gorm.DB
	backend backend.Backend
}

func (db *Database) Open() (err error) {
	var conn *gorm.DB
	switch db.backend.Engine() {
	case "sqlite":
		conn, err = gorm.Open(sqlite.Open(db.backend.Params()), &gorm.Config{})
	case "postgresql":
		conn, err = gorm.Open(postgres.Open(db.backend.Params()), &gorm.Config{})
	default:
		return errors.New("not supported database engine")
	}
	if err != nil {
		return err
	}
	db.db = conn
	return
}

func (db *Database) Close() (err error) {
	client, err := db.db.DB()
	if err != nil {
		return err
	}
	return client.Close()
}

func LoadDatabase() (*Database, error) {
	return &Database{backend: config.Default()}, nil
}

func (db *Database) Engine() string {
	return db.backend.Engine()
}

func (db *Database) AutoMigrate(dst ...any) (err error) {
	return db.db.AutoMigrate(dst...)
}
