package conf

import (
	"github.com/ozgur-yalcin/mfa/internal/backend"
	"github.com/ozgur-yalcin/mfa/internal/fs"
)

type Config struct {
	DatabaseBackend backend.Backend
}

const (
	sqliteFileName = "mfa.db"
)

func DefaultConfig() *Config {
	path := fs.MakeFilenamePath(sqliteFileName)
	//fmt.Println(path)
	return &Config{DatabaseBackend: backend.NewSqlite(path)}
}
