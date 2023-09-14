package utils

import (
	"encoding/hex"
	mRand "math/rand"

	whirl "github.com/balacode/zr-whirl"
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
	b := make([]byte, 22)
	for i := range b {
		// #nosec
		b[i] = SaltCharacters[mRand.Intn(len(SaltCharacters))]
	}
	return "$2a$06$" + string(b)
}

// GeneratePassword generates a random password for a user to change
func GeneratePassword() string {
	b := make([]byte, 12)
	for i := range b {
		// #nosec
		b[i] = PasswordCharacters[mRand.Intn(len(PasswordCharacters))]
	}
	return string(b)
}
