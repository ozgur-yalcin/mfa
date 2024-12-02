package database

import (
	"github.com/ozgur-yalcin/mfa/src/models"
)

func (db *Database) ListAccounts(issuer string, user string) (accounts []models.Account, err error) {
	db.client.Where(&models.Account{Issuer: issuer, User: user}).Find(&accounts)
	return
}

func (db *Database) AddAccount(account *models.Account) (err error) {
	return db.client.Create(account).Error
}

func (db *Database) DelAccount(issuer string, user string) (err error) {
	return db.client.Where(&models.Account{Issuer: issuer, User: user}).Delete(&models.Account{}).Error
}

func (db *Database) GetAccount(issuer string, user string) (account models.Account) {
	db.client.Model(models.Account{Issuer: issuer, User: user}).First(&account)
	return
}

func (db *Database) SetAccount(account models.Account) (err error) {
	return db.client.Save(&account).Error
}
