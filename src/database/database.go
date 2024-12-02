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
	client  *gorm.DB
	backend backend.Backend
}

func (db *Database) Open() (err error) {
	var client *gorm.DB
	switch db.backend.Engine() {
	case "sqlite":
		client, err = gorm.Open(sqlite.Open(db.backend.Params()), &gorm.Config{})
	case "postgresql":
		client, err = gorm.Open(postgres.Open(db.backend.Params()), &gorm.Config{})
	default:
		return errors.New("not supported database engine")
	}
	if err != nil {
		return err
	}
	db.client = client
	return
}

func (db *Database) Close() (err error) {
	client, err := db.client.DB()
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
	return db.client.AutoMigrate(dst...)
}
