package database

import (
	"github.com/ozgur-yalcin/mfa/internal/models"
)

func (db *Database) ListAccounts(issuer string, user string) (accounts []models.Account, err error) {
	db.db.Where(&models.Account{Issuer: issuer, User: user}).Find(&accounts)
	return
}

func (db *Database) AddAccount(account *models.Account) error {
	return db.db.Create(account).Error
}

func (db *Database) DelAccount(issuer string, user string) error {
	return db.db.Where(&models.Account{Issuer: issuer, User: user}).Delete(&models.Account{}).Error
}

func (db *Database) GetAccount(issuer string, user string) (account models.Account) {
	db.db.Model(models.Account{Issuer: issuer, User: user}).First(&account)
	return
}

func (db *Database) SetAccount(account models.Account) error {
	return db.db.Save(&account).Error
}
