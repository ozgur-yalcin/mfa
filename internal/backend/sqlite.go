package backend

import (
	"fmt"
)

type Sqlite struct {
	engine string
	name   string
}

func (db Sqlite) Engine() string {
	return db.engine
}

func (db Sqlite) Params() string {
	return fmt.Sprintf("file:%s?_journal=WAL&_vacuum=incremental", db.name)
}

func NewSqlite(filePath string) *Sqlite {
	return &Sqlite{engine: "sqlite", name: filePath}
}
