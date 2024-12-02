package otp

import (
	"encoding/binary"
	"time"
)

type TOTP struct {
	hash   string
	digits int
	period int64
}

func NewTOTP(hash string, digits int, period int64) *TOTP {
	return &TOTP{
		hash:   hash,
		digits: digits,
		period: period,
	}
}

func (t *TOTP) counter() int64 {
	return time.Now().UTC().Unix() / t.period
}

func (t *TOTP) GeneratePassCode(key string) (string, error) {
	return NewHOTP(t.hash, t.digits, t.counter()).GeneratePassCode(key)
}

func counterToBytes(counter int64) []byte {
	bytes := make([]byte, 8)
	binary.BigEndian.PutUint64(bytes, uint64(counter))
	return bytes
}
