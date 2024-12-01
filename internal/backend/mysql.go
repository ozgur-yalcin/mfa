package backend

import "fmt"

type Mysql struct {
	engine   string
	name     string
	user     string
	password string
	host     string
	port     int
	sslMode  string
}

func (db Mysql) Engine() string {
	return db.engine
}

func (db Mysql) Params() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local&tls=%s", db.user, db.password, db.host, db.port, db.name, db.sslMode)
}
