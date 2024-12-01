package backend

import "fmt"

type Postgresql struct {
	engine   string
	name     string
	user     string
	password string
	host     string
	port     int
	sslMode  string
}

func (db Postgresql) Engine() string {
	return db.engine
}

func (db Postgresql) Params() string {
	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=Europe/Istanbul", db.host, db.user, db.password, db.name, db.port, db.sslMode)
}
