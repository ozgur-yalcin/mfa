package initialize

import (
	"github.com/ozgur-yalcin/mfa/src/database"
	"github.com/ozgur-yalcin/mfa/src/models"
)

func DB() (err error) {
	db, err := database.LoadDatabase()
	if err != nil {
		return err
	}
	if err := db.Open(); err != nil {
		return err
	}
	defer db.Close()
	err = db.AutoMigrate(&models.Account{})
	if err != nil {
		return err
	}
	return
}
