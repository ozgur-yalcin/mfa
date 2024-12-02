package otp

import (
	"crypto/hmac"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base32"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"strings"
)

type HOTP struct {
	hash    string
	digits  int
	counter int64
}

func NewHOTP(hash string, digits int, counter int64) *HOTP {
	return &HOTP{
		hash:    hash,
		digits:  digits,
		counter: counter,
	}
}

func (t *HOTP) GeneratePassCode(key string) (code string, err error) {
	key = strings.Join(strings.Fields(key), "")
	key = strings.ToUpper(key)
	secret, err := base32.StdEncoding.DecodeString(key)
	if err != nil {
		return "", errors.New("base32 decoding failed: secret key is invalid")
	}
	var sum []byte
	switch t.hash {
	case "SHA1":
		mac := hmac.New(sha1.New, secret)
		mac.Write(counterToBytes(t.counter))
		sum = mac.Sum(nil)
	case "SHA256":
		mac := hmac.New(sha256.New, secret)
		mac.Write(counterToBytes(t.counter))
		sum = mac.Sum(nil)
	case "SHA512":
		mac := hmac.New(sha512.New, secret)
		mac.Write(counterToBytes(t.counter))
		sum = mac.Sum(nil)
	default:
		return "", errors.New("invalid hash algorithm. valid hash algorithms include values SHA1, SHA256, or SHA512")
	}
	offset := sum[len(sum)-1] & 0xf
	binaryCode := binary.BigEndian.Uint32(sum[offset:])
	verificationCode := int64(binaryCode) & 0x7FFFFFFF
	truncatedCode := verificationCode % int64(math.Pow10(t.digits))
	code = fmt.Sprintf(fmt.Sprintf("%%0%dd", t.digits), truncatedCode)
	return code, err
}
