package utils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math/big"

	whirl "github.com/balacode/zr-whirl"
)

type Type int

const (
	GeneratePassword Type = iota
	GenerateSalt
)

const (
	saltCharacters     = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789/."
	passwordCharacters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@*()&"
)

// HashPass hashes a password using a Whirlpool hash.
// Passwords are presumed to be hashed.
func HashPass(password string) string {
	iter := 1000
	var next string
	for i := 0; i < iter; i++ {
		next += password
		tmp := whirl.HashOfBytes([]byte(next), []byte(""))
		next = hex.EncodeToString(tmp)
	}
	return next
}

// GenerateRandom generates a random string for either password or salt
func GenerateRandom(t Type) (string, error) {
	switch t {
	case GeneratePassword:
		lenPass := big.NewInt(int64(len(passwordCharacters)))
		b, err := rangeLoop(lenPass, 12)
		if err != nil {
			return "", fmt.Errorf("error generating random: %w", err)
		}
		return b, nil
	case GenerateSalt:
		lenSalt := big.NewInt(int64(len(saltCharacters)))
		b, err := rangeLoop(lenSalt, 22)
		if err != nil {
			return "", fmt.Errorf("error generating random: %w", err)
		}
		return "$2a$06$" + b, nil
	default:
		return "", fmt.Errorf("invalid type: %d", t)
	}
}

// rangeLoop creates the random string that will be used
func rangeLoop(len *big.Int, size int) (string, error) {
	b := make([]byte, size)
	for i := range b {
		randInt, err := rand.Int(rand.Reader, len)
		if err != nil {
			return "", err
		}
		b[i] = passwordCharacters[randInt.Int64()]
	}
	return string(b), nil
}
