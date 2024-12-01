package database

import (
	"log"
	"sync"

	"github.com/ozgur-yalcin/mfa/internal/backend"
	"github.com/ozgur-yalcin/mfa/internal/conf"
	"gorm.io/driver/mysql"
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
	case "mysql":
		params := db.backend.Params()
		conn, err = gorm.Open(mysql.Open(params), &gorm.Config{})
	case "postgresql":
		params := db.backend.Params()
		conn, err = gorm.Open(postgres.Open(params), &gorm.Config{})
	default:
		log.Fatalf("not supported database engine: %s", db.backend.Engine())
	}
	if err != nil {
		log.Fatalf("failed to connect database:%s", err.Error())
	}
	db.db = conn
	return nil
}

func (db *Database) Close() error {
	client, err := db.db.DB()
	if err != nil {
		log.Fatalln(err)
	}
	return client.Close()
}

func LoadDatabase() (*Database, error) {
	return &Database{backend: conf.DefaultConfig().DatabaseBackend}, nil
}

func (db *Database) Engine() string {
	return db.backend.Engine()
}

func (db *Database) AutoMigrate(dst ...interface{}) error {
	var err error
	if db.Engine() == "mysql" {
		err = db.db.Set("gorm:table_options", "ENGINE=InnoDB CHARSET=utf8mb4").AutoMigrate(dst...)
	} else {
		err = db.db.AutoMigrate(dst...)
	}
	return err
}
