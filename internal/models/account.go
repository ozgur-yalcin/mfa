package models

import (
	"errors"

	"github.com/ozgur-yalcin/mfa/otp"
)

type Account struct {
	ID      uint   `json:"id" gorm:"primaryKey"`
	Issuer  string `json:"issuer" binding:"required"`
	User    string `json:"user"`
	Secret  string `json:"secret" binding:"required"`
	Mode    string `json:"mode"`
	Hash    string `json:"hash"`
	Digits  int    `json:"digits"`
	Period  int64  `json:"period"`
	Counter int64  `json:"counter"`
}

func (a Account) OTP() (code string, err error) {
	if a.Mode == "hotp" {
		hotp := otp.NewHOTP(a.Hash, a.Digits, a.Counter)
		code, err = hotp.GeneratePassCode(a.Secret)
	} else if a.Mode == "totp" {
		totp := otp.NewTOTP(a.Hash, a.Digits, a.Period)
		code, err = totp.GeneratePassCode(a.Secret)
	} else {
		return code, errors.New("mode should be hotp or totp")
	}
	return
}
