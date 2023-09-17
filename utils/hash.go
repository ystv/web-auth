package utils

import (
	"crypto/rand"
	"encoding/hex"
	whirl "github.com/balacode/zr-whirl"
	"log"
	"math/big"
)

const (
	SaltCharacters     = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789/."
	PasswordCharacters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@*()&"
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

// GenerateSalt generates a salt for the password to be salted against
func GenerateSalt() string {
	lenSalt := big.NewInt(int64(len(SaltCharacters)))

	b := make([]byte, 22)
	for i := range b {
		randInt, err := rand.Int(rand.Reader, lenSalt)
		if err != nil {
			log.Println("Error generating random salt:", err)
		}
		b[i] = SaltCharacters[randInt.Int64()]
	}
	return "$2a$06$" + string(b)
}

// GeneratePassword generates a random password for a user to change
func GeneratePassword() string {
	lenPass := big.NewInt(int64(len(PasswordCharacters)))

	b := make([]byte, 12)
	for i := range b {
		randInt, err := rand.Int(rand.Reader, lenPass)
		if err != nil {
			log.Println("Error generating random password:", err)
		}
		b[i] = PasswordCharacters[randInt.Int64()]
	}
	return string(b)
}
