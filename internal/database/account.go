package database

import (
	"github.com/ozgur-yalcin/mfa/internal/models"
)

func (db *Database) ListAccounts(accountName string, userName string) (accounts []models.Account, err error) {
	db.db.Where(&models.Account{AccountName: accountName, Username: userName}).Find(&accounts)
	return
}

func (db *Database) CreateAccount(account *models.Account) error {
	return db.db.Create(account).Error
}

func (db *Database) RemoveAccount(accountName string, userName string) error {
	return db.db.Where(&models.Account{AccountName: accountName, Username: userName}).Delete(&models.Account{}).Error
}

func (db *Database) GetAccount(accountName string, userName string) (account models.Account) {
	db.db.Model(models.Account{AccountName: accountName, Username: userName}).First(&account)
	return
}

func (db *Database) SaveAccount(account models.Account) error {
	return db.db.Save(&account).Error
}
