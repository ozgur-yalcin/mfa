package initialize

import (
	"log"

	"github.com/ozgur-yalcin/mfa/internal/database"
	"github.com/ozgur-yalcin/mfa/internal/models"
)

func DB() {
	db, err := database.LoadDatabase()
	if err != nil {
		log.Fatalf("failed to initialze database:%s", err.Error())
	}
	if err := db.Open(); err != nil {
		log.Fatalf("failed to connect database:%s", err.Error())
	}
	defer db.Close()

	err = db.AutoMigrate(&models.Account{})
	if err != nil {
		log.Fatalf("failed migrate database: %s", err.Error())
	}
}
