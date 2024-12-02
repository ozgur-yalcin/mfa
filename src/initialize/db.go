package initialize

import (
	"log"

	"github.com/ozgur-yalcin/mfa/src/database"
	"github.com/ozgur-yalcin/mfa/src/models"
)

func DB() {
	db, err := database.LoadDatabase()
	if err != nil {
		log.Fatal(err)
	}
	if err := db.Open(); err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = db.AutoMigrate(&models.Account{})
	if err != nil {
		log.Fatal(err)
	}
}
