package config

import (
	"path"

	"github.com/ozgur-yalcin/mfa/src/backend"
)

const (
	sqliteFileName = "mfa.db"
)

func Default() backend.Backend {
	return backend.NewSqlite(path.Join(".", sqliteFileName))
}
